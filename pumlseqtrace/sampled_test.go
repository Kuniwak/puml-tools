package pumlseqtrace

import (
	"math/rand/v2"
	"testing"

	"github.com/Kuniwak/puml-tools/pumlseqast"
)

func TestSampled(t *testing.T) {
	traces := func(rate float64, seed uint64) int {
		n := 0
		yield := Sampled(rate, rand.New(rand.NewPCG(seed, seed)), func([]pumlseqast.MessageKey) bool {
			n++
			return true
		})
		for range 1000 {
			yield(nil)
		}
		return n
	}

	t.Run("a rate of 1 keeps every trace", func(t *testing.T) {
		if got := traces(1, 1); got != 1000 {
			t.Errorf("Sampled(1, _, _) kept %d of 1000 traces, want 1000", got)
		}
	})

	t.Run("a rate below 1 thins out the traces", func(t *testing.T) {
		got := traces(0.1, 1)
		if got == 0 || got == 1000 {
			t.Errorf("Sampled(0.1, _, _) kept %d of 1000 traces, want some but not all", got)
		}
	})

	t.Run("the same seed keeps the same traces", func(t *testing.T) {
		if first, second := traces(0.1, 1), traces(0.1, 1); first != second {
			t.Errorf("Sampled(0.1, _, _) kept %d traces and then %d, want the same", first, second)
		}
	})
}
