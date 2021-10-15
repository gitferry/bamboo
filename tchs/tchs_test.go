package tchs

import (
	"testing"
)

// one chain
func TestTchs_CommitRule1(t *testing.T) {
	//bc := blockchain.NewBlockchain(4)
	//hs := Newtchs(bc, "default")
	//qc1 := &blockchain.QC{
	//	View:    1,
	//	BlockID: utils.IdentifierFixture(),
	//}
	//b1 := blockchain.MakeBlock(2, qc1, nil, "1")
	//bc.AddBlock(b1)
	//err := hs.UpdateStateByQC(b1.QC)
	//require.NoError(t, err)
	//canCommit, blockID, err := hs.CommitRule(b1.QC)
	//require.Error(t, err)
	//require.False(t, canCommit)
	//require.Nil(t, blockID)
}

// two chain
func TestTchs_CommitRule2(t *testing.T) {
	//bc := blockchain.NewBlockchain(4)
	//hs := Newtchs(bc, "default")
	//qc1 := &blockchain.QC{
	//	View:    1,
	//	BlockID: utils.IdentifierFixture(),
	//}
	//b1 := blockchain.MakeBlock(2, qc1, nil, "1")
	//bc.AddBlock(b1)
	//_ = hs.UpdateStateByQC(b1.QC)
	//qc2 := &blockchain.QC{
	//	View:    2,
	//	BlockID: b1.Hash,
	//}
	//b2 := blockchain.MakeBlock(3, qc2, nil, "1")
	//bc.AddBlock(b2)
	//err := hs.UpdateStateByQC(b2.QC)
	//require.NoError(t, err)
	//canCommit, blockID, err := hs.CommitRule(b2.QC)
	//require.Error(t, err)
	//require.False(t, canCommit)
	//require.Nil(t, blockID)
}

// three chain
func TestTchs_CommitRule3(t *testing.T) {
	//bc := blockchain.NewBlockchain(4)
	//hs := Newtchs(bc, "default")
	//qc1 := &blockchain.QC{
	//	View:    1,
	//	BlockID: utils.IdentifierFixture(),
	//}
	//b1 := blockchain.MakeBlock(2, qc1, nil, "1")
	//bc.AddBlock(b1)
	//qc2 := &blockchain.QC{
	//	View:    2,
	//	BlockID: b1.Hash,
	//}
	//b2 := blockchain.MakeBlock(3, qc2, nil, "1")
	//bc.AddBlock(b2)
	//qc3 := &blockchain.QC{
	//	View:    3,
	//	BlockID: b2.Hash,
	//}
	//b3 := blockchain.MakeBlock(4, qc3, nil, "1")
	//bc.AddBlock(b3)
	//err := hs.UpdateStateByQC(b3.QC)
	//require.NoError(t, err)
	//canCommit, committedBlock, err := hs.CommitRule(b3.QC)
	//require.NoError(t, err)
	//require.True(t, canCommit)
	//require.Equal(t, b1, committedBlock)
}

// func TestTchs_ForkingForkchoice(t *testing.T) {
// 	bc := blockchain.NewBlockchain(4)
// 	hs := Newtchs(bc, "forking")
// 	qc1 := &blockchain.QC{
// 		View:    1,
// 		BlockID: utils.IdentifierFixture(),
// 	}
// 	b1 := blockchain.MakeBlock(2, qc1, nil, "1")
// 	bc.AddBlock(b1)
// 	qc2 := &blockchain.QC{
// 		View:    2,
// 		BlockID: b1.Hash,
// 	}
// 	b2 := blockchain.MakeBlock(3, qc2, nil, "1")
// 	bc.AddBlock(b2)
// 	qc3 := &blockchain.QC{
// 		View:    3,
// 		BlockID: b2.Hash,
// 	}
// 	b3 := blockchain.MakeBlock(4, qc3, nil, "1")
// 	bc.AddBlock(b3)
// 	_ = hs.UpdateStateByQC(b3.QC)
// 	qc4 := &blockchain.QC{
// 		View:    4,
// 		BlockID: b3.Hash,
// 	}
// 	b4 := blockchain.MakeBlock(5, qc4, nil, "1")
// 	bc.AddBlock(b4)
// 	err := hs.UpdateStateByQC(b4.QC)
// 	require.NoError(t, err)
// 	canCommit, committedBlock, err := hs.CommitRule(b4.QC)
// 	require.NoError(t, err)
// 	require.True(t, canCommit)
// 	require.Equal(t, b1, committedBlock)
// }
