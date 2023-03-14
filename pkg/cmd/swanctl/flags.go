package swanctl

import (
	"github.com/spf13/pflag"
)

type addFlagFunc func(sf *swanctlFlags, fs *pflag.FlagSet)
type swanctlFlags struct {
	IKE     string
	Child   string
	Pretty  bool
	Raw     bool
	Timeout string
}

func addRawAndPretty(sf *swanctlFlags, fs *pflag.FlagSet) {
	fs.BoolVar(&sf.Pretty, "pretty", false, "Dump raw response message in pretty print")
	fs.BoolVar(&sf.Raw, "raw", false, "Dump raw response message")
}

func (sf *swanctlFlags) build(flags ...string) []string {
	if sf.Pretty {
		flags = append(flags, "--pretty")
	}

	if sf.Raw {
		flags = append(flags, "--raw")
	}

	if sf.Timeout != "" {
		flags = append(flags, "--timeout", sf.Timeout)
	}

	if sf.IKE != "" {
		flags = append(flags, "--ike", sf.IKE)
	}

	if sf.Child != "" {
		flags = append(flags, "--child", sf.Child)
	}

	return flags
}

func addIKE(sf *swanctlFlags, fs *pflag.FlagSet) {
	fs.StringVar(&sf.IKE, "ike", "", "IKE_SA name")
}

func addChild(sf *swanctlFlags, fs *pflag.FlagSet) {
	fs.StringVar(&sf.Child, "child", "", "CHILD_SA name")
}

func addTimeout(sf *swanctlFlags, fs *pflag.FlagSet) {
	fs.StringVar(&sf.Timeout, "timeout", "", "timeout in seconds before detaching")
}
