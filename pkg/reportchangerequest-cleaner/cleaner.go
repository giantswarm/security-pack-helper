package reportchangerequestcleaner

import (
	"context"
	"strings"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/prometheus/client_golang/prometheus"
	"go.etcd.io/etcd/clientv3"
	// "github.com/prometheus/client_golang/prometheus"
)

// Full path looks like /giantswarm.io/kyverno.io/reportchangerequests/.
const ReportChangeRequestPrefix = "kyverno.io/reportchangerequests/"

const (
	metricNamespace        = "security_pack_helper"
	metricSubsystem        = "interventions"
	metricInterventionType = "delete_reportchangerequests"
)

var (
	interventionCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricNamespace,
			Subsystem: metricSubsystem,
			Name:      "intervention_count",
			Help:      "The number of times the helper has needed to intervene in the cluster.",
		},
		[]string{
			"intevention_type",
		},
	)

	// = prometheus.NewDesc(
	// 	prometheus.BuildFQName(metricNamespace, metricSubsystem, "count"),
	// 	"The number of times the helper has needed to intervene in the cluster.",
	// 	[]string{
	// 		"intervention_type",
	// 	},
	// 	nil,
	// )
)

type Config struct {
	Logger micrologger.Logger
	// PromDesc         *prometheus.Desc
	EtcdClientConfig *clientv3.Config
	EtcdPrefix       string

	RCRLimit int
}

type RCRCleaner struct {
	logger micrologger.Logger
	// interventionMetric *prometheus.CounterVec
	etcdClientConfig   *clientv3.Config
	etcdResourcePrefix string // Note: this prefix is modified from the one passed in via config.

	rcrLimit int
}

func NewRCRCleaner(config Config) (*RCRCleaner, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.RCRLimit <= 0 {
		return nil, microerror.Maskf(invalidConfigError, "%T.RCRLimit must be greater than 0", config)
	}

	if config.EtcdClientConfig == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.EtcdClientConfig must not be empty", config)
	}
	if !strings.HasSuffix(config.EtcdPrefix, "/") || !strings.HasPrefix(config.EtcdPrefix, "/") {
		return nil, microerror.Maskf(invalidConfigError, "%T.EtcdPrefix has to start and end with a '/'", config)
	}

	// Create a prometheus counter to track how many times we've deleted resources.
	prometheus.MustRegister(interventionCount)

	// We hardcode the resource type for this behavior.
	// We allow otional configured prefixes, but we will enforce deletion of the correct resources.
	resourcePrefix := config.EtcdPrefix + ReportChangeRequestPrefix

	return &RCRCleaner{
		logger:             config.Logger,
		etcdClientConfig:   config.EtcdClientConfig,
		etcdResourcePrefix: resourcePrefix,
		rcrLimit:           config.RCRLimit,
	}, nil
}

// Retrieves the list of ReportChangeRequests in the cluster and deletes them if the list exceeds the configured threshold.
func (r *RCRCleaner) CheckAndDelete(ctx context.Context) error {
	cli, err := clientv3.New(*r.etcdClientConfig)
	if err != nil {
		return microerror.Mask(err)
	}

	defer cli.Close()

	resp, err := r.countResources(ctx, cli)
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.Debugf(ctx, "found %d resources matching %s", resp.Count, r.etcdResourcePrefix)

	if resp.Count > int64(r.rcrLimit) {
		r.logger.Debugf(ctx, "deleting resources")

		interventionCount.WithLabelValues(metricInterventionType).Inc()

		resp, err := r.deleteResources(ctx, cli)
		if err != nil {
			return microerror.Mask(err)
		}

		r.logger.Debugf(ctx, "deleted %d resources", resp.Deleted)
	} else {
		r.logger.Debugf(ctx, "resources are below threshold")
	}

	return nil
}

// Deletes the resources matching the configured prefix from etcd.
func (r *RCRCleaner) deleteResources(ctx context.Context, cli *clientv3.Client) (*clientv3.DeleteResponse, error) {
	resp, err := cli.Delete(ctx, r.etcdResourcePrefix, clientv3.WithPrefix())
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return resp, nil
}

// Counts the resources matching the configured prefix stored in etcd.
func (r *RCRCleaner) countResources(ctx context.Context, cli *clientv3.Client) (*clientv3.GetResponse, error) {
	resp, err := cli.Get(context.Background(), r.etcdResourcePrefix, clientv3.WithPrefix(), clientv3.WithCountOnly())
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return resp, nil
}
