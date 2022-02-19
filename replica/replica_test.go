package replica

import (
	"log"
	"testing"
)

func TestPickRandomNodes(t *testing.T) {
	n, d, index := 4, 1, 2
	test := 100
	for i := 0; i < test; i++ {
		pick := pickRandomNodes(n, d, index)
		log.Print(pick)
	}
}
