package types

import (
	"strings"
)

type Args struct {
	raw  []string
	args map[string][]string
}

func NewArgs(args_ []string) *Args {
	args := &Args{
		raw:  args_,
		args: make(map[string][]string),
	}
	args.parse()

	return args
}

// parse assumes each line is in this format: --name=value
func (args *Args) parse() {
	for _, line := range args.raw {
		index := strings.IndexByte(line, '=')
		if index == -1 {
			continue
		}

		// remove "--" from name
		name, value := line[2:index], line[index+1:]
		args.args[name] = append(args.args[name], value)
	}
}

func (args Args) GetValue(name string) string {
	values := args.args[name]
	if values == nil {
		return ""
	}

	return values[0]
}
