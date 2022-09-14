package cmd

import (
	"context"
	"io"
	"time"

	"github.com/giantswarm/k8sclient/v7/pkg/k8sclient"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	kyvernov1alpha2 "github.com/kyverno/kyverno/api/kyverno/v1alpha2"
	versioned "github.com/kyverno/kyverno/pkg/client/clientset/versioned"
	api "github.com/kyverno/kyverno/pkg/client/clientset/versioned/typed/kyverno/v1alpha2"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	cleaner "github.com/giantswarm/security-pack-helper/pkg/reportchangerequest-cleaner"
)

type runner struct {
	flag   *flag
	logger micrologger.Logger
	stdout io.Writer
	stderr io.Writer
}

func (r *runner) Run(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	err := r.flag.Validate(cmd)
	if err != nil {
		return err
	}

	err = r.run(ctx, cmd, args)
	if err != nil {
		return err
	}

	return nil
}

func (r *runner) run(ctx context.Context, cmd *cobra.Command, args []string) error {
	var err error

	r.logger.Debugf(ctx, "Interval: %d", r.flag.Interval)
	r.logger.Debugf(ctx, "RCR Limit: %d", r.flag.RCRLimit)

	var restConfig *rest.Config
	{
		restConfig, err = rest.InClusterConfig()
		if err != nil {
			return microerror.Mask(err)
		}
	}

	var k8sClient kubernetes.Interface
	{

		c := k8sclient.ClientsConfig{
			Logger: r.logger,
			SchemeBuilder: k8sclient.SchemeBuilder{
				//	v1alpha2.AddToScheme,
				kyvernov1alpha2.AddToScheme,
			},
			RestConfig: restConfig,
		}

		clients, err := k8sclient.NewClients(c)
		if err != nil {
			return microerror.Mask(err)
		}

		k8sClient = clients.K8sClient()
	}

	var kyvernoClient api.KyvernoV1alpha2Interface
	{
		kyverno, err := versioned.NewForConfig(restConfig)
		if err != nil {
			return microerror.Mask(err)
		}
		kyvernoClient = kyverno.KyvernoV1alpha2()
	}

	rcrCleaner, err := cleaner.NewRCRCleaner(cleaner.Config{
		Logger:    r.logger,
		K8sClient: k8sClient,
		KClient:   kyvernoClient,
		RCRLimit:  r.flag.RCRLimit,
	})
	if err != nil {
		return microerror.Mask(err)
	}

	for {
		ok, err := rcrCleaner.Check(ctx)
		if err != nil {
			r.logger.Errorf(ctx, err, "Error listing ReportChangeRequests")
			// Next loop.
			continue
		}
		if ok {
			r.logger.Debugf(ctx, "ReportChangeRequests are below threshold")
		} else {
			r.logger.Debugf(ctx, "Deleting ReportChangeRequests")
			err = rcrCleaner.DeleteRCRs(ctx)
			if err != nil {
				r.logger.Errorf(ctx, err, "Error deleting ReportChangeRequests")
				// Next loop.
				continue
			}

			r.logger.Debugf(ctx, "ReportChangeRequests deleted")
		}

		time.Sleep(time.Second * time.Duration(r.flag.Interval))
	}
}
