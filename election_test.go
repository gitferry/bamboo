package zeitgeber

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRotation_IsLeader(t *testing.T) {
	elect := NewRotation(4)
	leaderID := elect.FindLeaderFor(1)
	require.True(t, elect.IsLeader(leaderID, 1))

	leaderID = elect.FindLeaderFor(4)
	require.Equal(t, "3", leaderID)

	leaderID = elect.FindLeaderFor(3)
	require.True(t, elect.IsLeader(leaderID, 3))
}
