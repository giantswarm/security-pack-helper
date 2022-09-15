package reportchangerequestcleaner

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	api "github.com/kyverno/kyverno/pkg/client/clientset/versioned/typed/kyverno/v1alpha2"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func (r *RCRCleaner) Check(ctx context.Context) (bool, error) {
	// Check RCRs are under configured limit
	rcrs, err := r.kClient.ReportChangeRequests(r.rcrNamespace).List(ctx, v1.ListOptions{})
	if err != nil {
		return false, microerror.Mask(err)
	}

	rcrCount := len(rcrs.Items)

	r.logger.Debugf(ctx, "found %d ReportChangeRequests", rcrCount)

	if rcrCount > r.rcrLimit {
		// We are over the limit. Fail the check.
		return false, nil
	}

	return true, nil
}

func (r *RCRCleaner) DeleteRCRs(ctx context.Context) error {
	err := r.kClient.ReportChangeRequests(r.rcrNamespace).DeleteCollection(ctx, v1.DeleteOptions{}, v1.ListOptions{})
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
