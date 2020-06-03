package zeitgeber

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFailStop(t *testing.T) {
	// serverNo := 10
	// byzantineNo := 2
	// cfg := make_config(t, serverNo, byzantineNo, false)
}

func TestRandomPick(t *testing.T) {
	byzIDs := RandomPick(10, 3)
	require.Equal(t, 3, len(byzIDs))
}

func TestFindIntSlice(t *testing.T) {
	intSlice := []int{1, 2, 3}
	t1 := FindIntSlice(intSlice, 1)
	require.True(t, t1)
	t2 := FindIntSlice(intSlice, 4)
	require.False(t, t2)
	t3 := FindIntSlice(intSlice, 3)
	require.True(t, t3)
	emptySlice := []int{}
	t4 := FindIntSlice(emptySlice, 1)
	require.False(t, t4)
}
