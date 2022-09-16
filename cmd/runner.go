package cmd

import (
	"context"
	"io"
	"time"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	versioned "github.com/kyverno/kyverno/pkg/client/clientset/versioned"
	api "github.com/kyverno/kyverno/pkg/client/clientset/versioned/typed/kyverno/v1alpha2"
	"github.com/spf13/cobra"
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
	r.logger.Debugf(ctx, "RCR Namespace: %s", r.flag.RCRNamespace)

	var restConfig *rest.Config
	{
		restConfig, err = rest.InClusterConfig()
		if err != nil {
			return microerror.Mask(err)
		}
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
		Logger:       r.logger,
		KClient:      kyvernoClient,
		RCRLimit:     r.flag.RCRLimit,
		RCRNamespace: r.flag.RCRNamespace,
	})
	if err != nil {
		return microerror.Mask(err)
	}

	for {
		err := rcrCleaner.CheckAndDelete(ctx)
		if err != nil {
			r.logger.Errorf(ctx, err, "Error checking ReportChangeRequests")
			// Next loop.
			continue
		}

		time.Sleep(time.Second * time.Duration(r.flag.Interval))
	}
}
