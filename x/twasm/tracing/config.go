package tracing

import (
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"
)

// Module init related flags
const (
	flagTgradeTracingEnabled = "twasm.open-tracing"
)

var tracerEnabled bool

// AddModuleInitFlags implements servertypes.ModuleInitFlags interface.
func AddModuleInitFlags(startCmd *cobra.Command) {
	startCmd.Flags().Bool(flagTgradeTracingEnabled, false, "Enable opentracing agent")
}

// ReadTracerConfig reads the tracer flag
func ReadTracerConfig(opts servertypes.AppOptions) error {
	if v := opts.Get(flagTgradeTracingEnabled); v != nil {
		var err error
		if tracerEnabled, err = cast.ToBoolE(v); err != nil {
			return err
		}
	}
	return nil
}

func hasTracerFlagSet(cmd *cobra.Command) bool {
	ok, err := cmd.Flags().GetBool(flagTgradeTracingEnabled)
	return err == nil && ok
}
