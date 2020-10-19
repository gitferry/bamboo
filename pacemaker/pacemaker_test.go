package pacemaker

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// receive only one tmo
func TestRemoteTmo1(t *testing.T) {
	pm := NewPacemaker(4)
	tmo1 := &TMO{
		View:   2,
		NodeID: "1",
	}
	isBuilt, tc := pm.ProcessRemoteTmo(tmo1)
	require.False(t, isBuilt)
	require.Nil(t, tc)
}

// receive only two tmo
func TestRemoteTmo2(t *testing.T) {
	pm := NewPacemaker(4)
	tmo1 := &TMO{
		View:   2,
		NodeID: "1",
	}
	isBuilt, tc := pm.ProcessRemoteTmo(tmo1)
	tmo2 := &TMO{
		View:   2,
		NodeID: "2",
	}
	isBuilt, tc = pm.ProcessRemoteTmo(tmo2)
	require.False(t, isBuilt)
	require.Nil(t, tc)
}

// receive only three tmo
func TestRemoteTmo3(t *testing.T) {
	pm := NewPacemaker(4)
	tmo1 := &TMO{
		View:   2,
		NodeID: "1",
	}
	isBuilt, tc := pm.ProcessRemoteTmo(tmo1)
	tmo2 := &TMO{
		View:   2,
		NodeID: "2",
	}
	isBuilt, tc = pm.ProcessRemoteTmo(tmo2)
	tmo3 := &TMO{
		View:   2,
		NodeID: "3",
	}
	isBuilt, tc = pm.ProcessRemoteTmo(tmo3)
	require.True(t, isBuilt)
	require.NotNil(t, tc)
}

// receive four tmo
func TestRemoteTmo4(t *testing.T) {
	pm := NewPacemaker(4)
	tmo1 := &TMO{
		View:   2,
		NodeID: "1",
	}
	isBuilt, tc := pm.ProcessRemoteTmo(tmo1)
	tmo2 := &TMO{
		View:   2,
		NodeID: "2",
	}
	isBuilt, tc = pm.ProcessRemoteTmo(tmo2)
	tmo3 := &TMO{
		View:   2,
		NodeID: "3",
	}
	isBuilt, tc = pm.ProcessRemoteTmo(tmo3)
	tmo4 := &TMO{
		View:   2,
		NodeID: "4",
	}
	isBuilt, tc = pm.ProcessRemoteTmo(tmo4)
	require.False(t, isBuilt)
	require.NotNil(t, tc)
}
