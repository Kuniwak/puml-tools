package pumlseqtrace

import (
	"math/rand/v2"

	"github.com/Kuniwak/puml-tools/pumlseqast"
)

func Sampled(rate float64, r *rand.Rand, inner func(trace []pumlseqast.MessageKey) bool) func(trace []pumlseqast.MessageKey) bool {
	return func(trace []pumlseqast.MessageKey) bool {
		if r.Float64() >= rate {
			return true
		}
		return inner(trace)
	}
}
