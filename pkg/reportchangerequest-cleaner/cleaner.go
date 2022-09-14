package reportchangerequestcleaner

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"k8s.io/client-go/kubernetes"
)

type Config struct {
	Logger    micrologger.Logger
	K8sClient kubernetes.Interface

	RCRLimit int
}

type RCRCleaner struct {
	logger    micrologger.Logger
	k8sClient kubernetes.Interface

	rcrLimit int
}

func NewRCRCleaner(config Config) (*RCRCleaner, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}

	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.K8sClient must not be empty", config)
	}

	if config.RCRLimit <= 0 {
		return nil, microerror.Maskf(invalidConfigError, "%T.RCRLimit must be greater than 0", config)
	}

	return &RCRCleaner{
		logger:    config.Logger,
		k8sClient: config.K8sClient,
		rcrLimit:  config.RCRLimit,
	}, nil
}

func (r *RCRCleaner) Check(ctx context.Context) bool {
	// Check RCRs are under configured limit
	return false
}

func (r *RCRCleaner) DeleteRCRs(ctx context.Context) error {
	// podList, err := r.k8sClient.CoreV1().Pods(r.namespace).List(ctx, v1.ListOptions{
	// 	LabelSelector: r.labelSelector,
	// 	FieldSelector: fmt.Sprintf("spec.nodeName=%s", r.nodeName),
	// })
	// if err != nil {
	// 	return microerror.Mask(err)
	// }

	// pod := podList.Items[0]

	// err = r.k8sClient.CoreV1().Pods(pod.Namespace).Delete(ctx, pod.Name, v1.DeleteOptions{})
	// if err != nil {
	// 	return microerror.Mask(err)
	// }

	return nil
}
