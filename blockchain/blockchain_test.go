package blockchain

import (
	"testing"

	"github.com/gitferry/zeitgeber/utils"
	"github.com/stretchr/testify/require"
)

// add one qc
func TestUpdateHighQC1(t *testing.T) {
	bc := NewBlockchain(10)
	qc1 := &QC{
		View:    1,
		BlockID: utils.IdentifierFixture(),
	}
	err := bc.UpdateHighQC(qc1)
	require.NoError(t, err)
	highQC := bc.GetHighQC()
	require.Equal(t, qc1, highQC)
}

// add two qcs, the second is higher
func TestUpdateHighQC2(t *testing.T) {
	bc := NewBlockchain(10)
	qc1 := &QC{
		View:    1,
		BlockID: utils.IdentifierFixture(),
	}
	_ = bc.UpdateHighQC(qc1)
	qc2 := &QC{
		View:    2,
		BlockID: utils.IdentifierFixture(),
	}
	_ = bc.UpdateHighQC(qc2)
	highQC := bc.GetHighQC()
	require.Equal(t, qc2, highQC)
}

// add two qcs, the first is higher
func TestUpdateHighQC3(t *testing.T) {
	bc := NewBlockchain(10)
	qc1 := &QC{
		View:    2,
		BlockID: utils.IdentifierFixture(),
	}
	_ = bc.UpdateHighQC(qc1)
	qc2 := &QC{
		View:    1,
		BlockID: utils.IdentifierFixture(),
	}
	_ = bc.UpdateHighQC(qc2)
	highQC := bc.GetHighQC()
	require.Equal(t, qc1, highQC)
}

func TestAddBlock(t *testing.T) {
	bc := NewBlockchain(10)
	qc := &QC{
		View:    1,
		BlockID: utils.IdentifierFixture(),
	}
	b := &Block{
		View:   2,
		QC:     qc,
		PrevID: qc.BlockID,
	}
	bc.AddBlock(b)
	highQC := bc.GetHighQC()
	require.Equal(t, qc, highQC)
}
