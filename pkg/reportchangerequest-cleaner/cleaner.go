package reportchangerequestcleaner

import (
	"context"
	"strings"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"go.etcd.io/etcd/clientv3"
	// api "github.com/kyverno/kyverno/pkg/client/clientset/versioned/typed/kyverno/v1alpha2"
	// dclient "github.com/kyverno/kyverno/pkg/dclient"
)

// Full path is /giantswarm.io/kyverno.io/reportchangerequests/
const ReportChangeRequestPrefix = "kyverno.io/reportchangerequests/"

type Config struct {
	Logger micrologger.Logger
	// KyvernoClient    api.KyvernoV1alpha2Interface
	// KyvernoDClient   dclient.Interface
	EtcdClientConfig *clientv3.Config
	EtcdPrefix       string

	RCRLimit int
	// RCRNamespace string
}

type RCRCleaner struct {
	logger micrologger.Logger
	// kyvernoClient    api.KyvernoV1alpha2Interface
	// kyvernoDClient   dclient.Interface
	etcdClientConfig *clientv3.Config
	etcdPrefix       string

	rcrLimit int
	// rcrNamespace string
}

func NewRCRCleaner(config Config) (*RCRCleaner, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	// if config.KyvernoClient == nil {
	// 	return nil, microerror.Maskf(invalidConfigError, "%T.KClient must not be empty", config)
	// }

	if config.RCRLimit <= 0 {
		return nil, microerror.Maskf(invalidConfigError, "%T.RCRLimit must be greater than 0", config)
	}

	// if config.RCRNamespace == "" {
	// 	return nil, microerror.Maskf(invalidConfigError, "%T.RCRNamespace must not be empty", config)
	// }

	if config.EtcdClientConfig == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.EtcdClientConfig must not be empty", config)
	}
	if !strings.HasSuffix(config.EtcdPrefix, "/") || !strings.HasPrefix(config.EtcdPrefix, "/") {
		return nil, microerror.Maskf(invalidConfigError, "%T.EtcdPrefix has to start and end with a '/'", config)
	}

	return &RCRCleaner{
		logger: config.Logger,
		// kyvernoClient:    config.KyvernoClient,
		// kyvernoDClient:   config.KyvernoDClient,
		etcdClientConfig: config.EtcdClientConfig,
		etcdPrefix:       config.EtcdPrefix,
		rcrLimit:         config.RCRLimit,
		// rcrNamespace:     config.RCRNamespace,
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

	r.logger.Debugf(ctx, "found %d resources matching %s", resp.Count, r.etcdPrefix)

	if resp.Count > int64(r.rcrLimit) {
		r.logger.Debugf(ctx, "Deleting resources")
		//
		r.logger.Debugf(ctx, "resources deleted")
	} else {
		r.logger.Debugf(ctx, "resources are below threshold")
	}

	return nil
}

// Deletes the resources matching the configured prefix from etcd.
func (r *RCRCleaner) deleteResources(ctx context.Context, cli *clientv3.Client) error {

	resp, err := cli.Delete(ctx, r.etcdPrefix, clientv3.WithPrefix())
	if err != nil {
		return microerror.Mask(err)
	}

	r.logger.Debugf(ctx, "deleted %d resources", resp.Deleted)

	return nil
}

// Counts the resources stored in etcd
func (r *RCRCleaner) countResources(ctx context.Context, cli *clientv3.Client) (*clientv3.GetResponse, error) {
	resp, err := cli.Get(context.Background(), "/", clientv3.WithPrefix(), clientv3.WithCountOnly())
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return resp, nil
}

// Lists the resources stored in etcd
// func (r *RCRCleaner) listResources(ctx context.Context, cli *clientv3.Client) (*clientv3.GetResponse, error) {
// 	resp, err := cli.Get(context.Background(), "/", clientv3.WithPrefix(), clientv3.WithKeysOnly())
// 	if err != nil {
// 		return nil, microerror.Mask(err)
// 	}

// 	return resp, nil
// }
