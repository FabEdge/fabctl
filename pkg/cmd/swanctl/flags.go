package swanctl

import (
	"github.com/spf13/pflag"
)

type swanctlFlags struct {
	IKE    string
	Child  string
	Pretty bool
	Raw    bool
}

func (sf *swanctlFlags) addRawAndPretty(fs *pflag.FlagSet) {
	fs.BoolVar(&sf.Pretty, "pretty", false, "Dump raw response message in pretty print")
	fs.BoolVar(&sf.Raw, "raw", false, "Dump raw response message")
}

func (sf *swanctlFlags) addIKE(fs *pflag.FlagSet) {
	fs.StringVar(&sf.IKE, "ike", "", "IKE_SA name")
}

func (sf *swanctlFlags) addChild(fs *pflag.FlagSet) {
	fs.StringVar(&sf.Child, "child", "", "CHILD_SA name")
}

func (sf *swanctlFlags) build(flags ...string) []string {
	if sf.Pretty {
		flags = append(flags, "--pretty")
	}

	if sf.Raw {
		flags = append(flags, "--raw")
	}

	if sf.IKE != "" {
		flags = append(flags, "--ike", sf.IKE)
	}

	if sf.Child != "" {
		flags = append(flags, "--child", sf.Child)
	}

	return flags
}
