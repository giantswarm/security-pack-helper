package cmd

import (
	goflag "flag"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	intervalFlag = "interval"
	rcrLimitFlag = "rcr-limit"
)

type flag struct {
	Interval int
	RCRLimit int
}

func (f *flag) Init(cmd *cobra.Command) {
	// Add command line flags for glog.
	pflag.CommandLine.AddGoFlagSet(goflag.CommandLine)

	cmd.Flags().IntVar(&f.Interval, intervalFlag, 60, `Interval in seconds to wait between tests. Defaults to 60.`)
	cmd.Flags().IntVar(&f.RCRLimit, rcrLimitFlag, 2000, `If the number of ReportChangeRequests in the cluster exceeds this number, they will be deleted. Defaults to 2000.`)

}

func (f *flag) Validate(cmd *cobra.Command) error {
	var err error

	if f.Interval <= 0 {
		return fmt.Errorf("--%s should be greater than 0", intervalFlag)
	}

	if f.RCRLimit <= 0 {
		return fmt.Errorf("--%s should be greater than 0", rcrLimitFlag)
	}

	return err
}
