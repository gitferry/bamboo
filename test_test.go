package zeitgeber

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestLeaderElection(t *testing.T) {
	serverNo := 7
	byzantineNo := 2
	cfg := make_config(t, serverNo, byzantineNo, false)
	view, isLeader := cfg.replicas[0].GetState()
	require.Equal(t, 1, view)
	require.True(t, isLeader)
}

func TestFailStop(t *testing.T) {
	serverNo := 4
	byzantineNo := 1
	cfg := make_config(t, serverNo, byzantineNo, false)
	// var view int
	// var isLeader bool
	// for _, r := range cfg.replicas {
	// 	view, isLeader = r.GetState()
	// 	require.Equal(t, 1, view)
	// }
	cfg.begin("Fail-stop testing with 1 byzantine nodes out of 4")
	time.Sleep(10 * time.Second)
	defer cfg.cleanup()
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
