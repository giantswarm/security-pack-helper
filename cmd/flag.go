package cmd

import (
	goflag "flag"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	intervalFlag = "interval"
	rcrLimitFlag = "rcr-limit"

	// ETCD configuration
	caCertFlag        = "ca-cert-path"
	certPathFlag      = "cert-path"
	keyPathFlag       = "key-path"
	etcdEndpointsFlag = "etcd-endpoints"
	etcdPrefixFlag    = "etcd-prefix"
	dialTimeoutFlag   = "etcd-dial-timeout"
)

type flag struct {
	Interval int
	RCRLimit int

	// ETCD configuration
	CACert        string
	CertPath      string
	KeyPath       string
	EtcdEndpoints []string
	EtcdPrefix    string
	DialTimeout   int
}

func (f *flag) Init(cmd *cobra.Command) {
	// Add command line flags for glog.
	pflag.CommandLine.AddGoFlagSet(goflag.CommandLine)

	cmd.Flags().IntVar(&f.Interval, intervalFlag, 60, `Interval in seconds to wait between tests. Defaults to 60.`)
	cmd.Flags().IntVar(&f.RCRLimit, rcrLimitFlag, 2000, `If the number of ReportChangeRequests in the cluster exceeds this number, they will be deleted. Defaults to 2000.`)
	// cmd.Flags().StringVar(&f.RCRNamespace, rcrNamespaceFlag, "kyverno", `The namespace where ReportChangeRequests will be watched. Defaults to "kyverno".`)

	// ETCD configuration flags
	cmd.Flags().StringVar(&f.CACert, caCertFlag, "/certs/server-ca.pem", "The path to the root CA certificate to use for etcd connections.")
	cmd.Flags().StringVar(&f.CertPath, certPathFlag, "/certs/server-crt.pem", "Path to the server cert to use for etcd connections.")
	cmd.Flags().StringVar(&f.KeyPath, keyPathFlag, "/certs/server-key.pem", "Path to the server key file to use for etcd connections.")
	cmd.Flags().StringArrayVar(&f.EtcdEndpoints, etcdEndpointsFlag, []string{"https://127.0.0.1:2379"}, "Array of etcd endpoints to connect to.")
	cmd.Flags().StringVar(&f.EtcdPrefix, etcdPrefixFlag, "/giantswarm.io/", "Prefix under which target etcd resources are stored.")
	cmd.Flags().IntVar(&f.DialTimeout, dialTimeoutFlag, 10, "Timeout duration for etcd connections, in seconds.")
}

func (f *flag) Validate(cmd *cobra.Command) error {
	var err error

	if f.Interval <= 0 {
		return fmt.Errorf("--%s should be greater than 0", intervalFlag)
	}

	if f.RCRLimit <= 0 {
		return fmt.Errorf("--%s should be greater than 0", rcrLimitFlag)
	}

	if f.CACert == "" {
		return fmt.Errorf("--%s must not be empty", caCertFlag)
	}

	if f.CertPath == "" {
		return fmt.Errorf("--%s must not be empty", certPathFlag)
	}

	if f.KeyPath == "" {
		return fmt.Errorf("--%s must not be empty", keyPathFlag)
	}

	if len(f.EtcdEndpoints) == 0 {
		return fmt.Errorf("--%s must not be empty", etcdEndpointsFlag)
	}

	if f.EtcdPrefix == "" {
		return fmt.Errorf("--%s must not be empty", etcdPrefixFlag)
	}

	if !strings.HasSuffix(f.EtcdPrefix, "/") || !strings.HasPrefix(f.EtcdPrefix, "/") {
		return fmt.Errorf("--%s has to start and end with a '/'", etcdPrefixFlag)
	}

	return err
}
