package pumlseqtrace

import (
	"github.com/Kuniwak/puml-tools/pumlseqast"
)

func CountAtMost(d *pumlseqast.Diagram, opt Options, limit int) int {
	n := 0
	if limit <= 0 {
		return 0
	}
	Walk(d, opt, func([]pumlseqast.MessageKey) bool {
		n++
		return n < limit
	})
	return n
}
