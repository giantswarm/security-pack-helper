package reportchangerequestcleaner

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/kyverno/kyverno/api/kyverno/v1alpha2"
	api "github.com/kyverno/kyverno/pkg/client/clientset/versioned/typed/kyverno/v1alpha2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Config struct {
	Logger  micrologger.Logger
	KClient api.KyvernoV1alpha2Interface

	RCRLimit     int
	RCRNamespace string
}

type RCRCleaner struct {
	logger  micrologger.Logger
	kClient api.KyvernoV1alpha2Interface

	rcrLimit     int
	rcrNamespace string
}

func NewRCRCleaner(config Config) (*RCRCleaner, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.KClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.KClient must not be empty", config)
	}

	if config.RCRLimit <= 0 {
		return nil, microerror.Maskf(invalidConfigError, "%T.RCRLimit must be greater than 0", config)
	}

	if config.RCRNamespace == "" {
		return nil, microerror.Maskf(invalidConfigError, "%T.RCRNamespace must not be empty", config)
	}

	return &RCRCleaner{
		logger:   config.Logger,
		kClient:  config.KClient,
		rcrLimit: config.RCRLimit,
	}, nil
}

// Retrieves the list of ReportChangeRequests in the cluster and deletes them if the list exceeds the configured threshold.
func (r *RCRCleaner) CheckAndDelete(ctx context.Context) error {
	rcrs, err := r.listRCRs(ctx)
	if err != nil {
		return microerror.Mask(err)
	}

	rcrCount := len(rcrs.Items)
	r.logger.Debugf(ctx, "found %d ReportChangeRequests", rcrCount)

	if rcrCount > r.rcrLimit {
		r.logger.Debugf(ctx, "Deleting ReportChangeRequests")

		err = r.deleteRCRs(ctx, rcrs)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.Debugf(ctx, "ReportChangeRequests deleted")
	} else {
		r.logger.Debugf(ctx, "ReportChangeRequests are below threshold")
	}

	return nil
}

// Deletes a list of ReportChangeRequests. Does not error if individual deletions fail.
func (r *RCRCleaner) deleteRCRs(ctx context.Context, rcrs *v1alpha2.ReportChangeRequestList) error {
	r.logger.Debugf(ctx, "deletion namespace: %s", r.rcrNamespace)

	for _, rcr := range rcrs.Items {
		r.logger.Debugf(ctx, "deleting: %s", rcr.GetName())
		err := r.kClient.ReportChangeRequests(r.rcrNamespace).Delete(
			ctx,
			// apitypes.NamespacedName{Namespace: r.rcrNamespace, Name: rcr.Name}.String(),
			rcr.GetName(),
			metav1.DeleteOptions{})
		if err != nil {
			r.logger.Errorf(ctx, err, "error deleting ReportChangeRequest")
			// Continue anyway -- we expect to have lots of these as RCRs are deleted after our initial list call.
		}
	}

	return nil
}

// Lists the ReportChangeRequests in the cluster in the configured watched namespace.
func (r *RCRCleaner) listRCRs(ctx context.Context) (*v1alpha2.ReportChangeRequestList, error) {
	rcrs, err := r.kClient.ReportChangeRequests(r.rcrNamespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, microerror.Mask(err)
	}
	return rcrs, nil
}
