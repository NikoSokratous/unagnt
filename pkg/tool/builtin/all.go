package builtin

import (
	"github.com/NikoSokratous/unagnt/pkg/tool"
)

// All returns all built-in tools.
func All() []tool.Tool {
	return []tool.Tool{
		&HTTPRequest{},
		&Echo{},
		&Calc{},
	}
}
