package cmd

import (
	"context"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/pkg/transport"

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

	tlsInfo := transport.TLSInfo{
		TrustedCAFile: r.flag.CACert,
		CertFile:      r.flag.CertPath,
		KeyFile:       r.flag.KeyPath,
	}

	tlsConfig, err := tlsInfo.ClientConfig()
	if err != nil {
		return microerror.Mask(err)
	}

	rcrCleaner, err := cleaner.NewRCRCleaner(cleaner.Config{
		Logger:     r.logger,
		RCRLimit:   r.flag.RCRLimit,
		EtcdPrefix: r.flag.EtcdPrefix,

		EtcdClientConfig: &clientv3.Config{
			Endpoints:   r.flag.EtcdEndpoints,
			DialTimeout: time.Second * time.Duration(r.flag.DialTimeout),
			TLS:         tlsConfig,
		},
	})
	if err != nil {
		return microerror.Mask(err)
	}

	metrics := http.NewServeMux()
	metrics.Handle("/metrics", promhttp.Handler())
	go serveMetrics(r.flag, metrics)

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

func serveMetrics(config *flag, handler http.Handler) {
	server := &http.Server{
		Addr:    config.MetricsAddress,
		Handler: handler,
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM)
	go func() {
		<-sig
		err := server.Shutdown(context.Background())
		if err != nil {
			panic(microerror.JSON(err))
		}
	}()

	err := server.ListenAndServe()
	if err != nil {
		if err != http.ErrServerClosed {
			panic(microerror.JSON(err))
		}
	}
}
