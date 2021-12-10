package blockchain

// add one qc
//func TestUpdateHighQC1(t *testing.T) {
//	bc := NewBlockchain(10)
//	qc1 := &QC{
//		View:    1,
//		BlockID: utils.IdentifierFixture(),
//	}
//	bc.UpdateHighQC(qc1)
//	highQC := bc.GetHighQC()
//	require.Equal(t, qc1, highQC)
//}
//
//// add two qcs, the second is higher
//func TestUpdateHighQC2(t *testing.T) {
//	bc := NewBlockchain(10)
//	qc1 := &QC{
//		View:    1,
//		BlockID: utils.IdentifierFixture(),
//	}
//	bc.UpdateHighQC(qc1)
//	qc2 := &QC{
//		View:    2,
//		BlockID: utils.IdentifierFixture(),
//	}
//	bc.UpdateHighQC(qc2)
//	highQC := bc.GetHighQC()
//	require.Equal(t, qc2, highQC)
//}
//
//// add two qcs, the first is higher
//func TestUpdateHighQC3(t *testing.T) {
//	bc := NewBlockchain(10)
//	qc1 := &QC{
//		View:    2,
//		BlockID: utils.IdentifierFixture(),
//	}
//	bc.UpdateHighQC(qc1)
//	qc2 := &QC{
//		View:    1,
//		BlockID: utils.IdentifierFixture(),
//	}
//	bc.UpdateHighQC(qc2)
//	highQC := bc.GetHighQC()
//	require.Equal(t, qc1, highQC)
//}
//
//func TestAddBlock(t *testing.T) {
//	bc := NewBlockchain(10)
//	qc := &QC{
//		View:    1,
//		BlockID: utils.IdentifierFixture(),
//	}
//	b := &Block{
//		View:   2,
//		QC:     qc,
//		PrevID: qc.BlockID,
//	}
//	bc.AddBlock(b)
//	highQC := bc.GetHighQC()
//	require.Equal(t, qc, highQC)
//}
//
//// add two blocks with parent-child relationship
//func TestParentBlock1(t *testing.T) {
//	bc := NewBlockchain(10)
//	qc1 := &QC{
//		View:    1,
//		BlockID: utils.IdentifierFixture(),
//	}
//	b1 := BuildProposal(2, qc1, nil, "1")
//	bc.AddBlock(b1)
//	qc2 := &QC{
//		View:    2,
//		BlockID: b1.ID,
//	}
//	b2 := BuildProposal(3, qc2, nil, "1")
//	bc.AddBlock(b2)
//	parent, err := bc.GetParentBlock(b2.ID)
//	require.NoError(t, err)
//	require.Equal(t, b1, parent)
//}
//
//// add two blocks without parent-child relationship
//func TestParentBlock2(t *testing.T) {
//	bc := NewBlockchain(10)
//	qc1 := &QC{
//		View:    1,
//		BlockID: utils.IdentifierFixture(),
//	}
//	b1 := BuildProposal(2, qc1, nil, "1")
//	bc.AddBlock(b1)
//	qc2 := &QC{
//		View:    2,
//		BlockID: utils.IdentifierFixture(),
//	}
//	b2 := BuildProposal(3, qc2, nil, "1")
//	bc.AddBlock(b2)
//	parent, err := bc.GetParentBlock(b2.ID)
//	require.Error(t, err)
//	require.Nil(t, parent)
//}
//
//// add three blocks with grandparent-parent-child relationship
//func TestGrandParentBlock1(t *testing.T) {
//	bc := NewBlockchain(10)
//	qc1 := &QC{
//		View:    1,
//		BlockID: utils.IdentifierFixture(),
//	}
//	b1 := BuildProposal(2, qc1, nil, "1")
//	bc.AddBlock(b1)
//	qc2 := &QC{
//		View:    2,
//		BlockID: b1.ID,
//	}
//	b2 := BuildProposal(3, qc2, nil, "1")
//	bc.AddBlock(b2)
//	qc3 := &QC{
//		View:    3,
//		BlockID: b2.ID,
//	}
//	b3 := BuildProposal(4, qc3, nil, "1")
//	bc.AddBlock(b3)
//	grandParent, err := bc.GetGrandParentBlock(b3.ID)
//	require.NoError(t, err)
//	require.Equal(t, b1, grandParent)
//}
//
//// add three blocks without grandparent-parent-child relationship
//func TestGrandParentBlock2(t *testing.T) {
//	bc := NewBlockchain(10)
//	qc1 := &QC{
//		View:    1,
//		BlockID: utils.IdentifierFixture(),
//	}
//	b1 := BuildProposal(2, qc1, nil, "1")
//	bc.AddBlock(b1)
//	qc2 := &QC{
//		View:    2,
//		BlockID: b1.ID,
//	}
//	b2 := BuildProposal(3, qc2, nil, "1")
//	bc.AddBlock(b2)
//	qc3 := &QC{
//		View:    3,
//		BlockID: utils.IdentifierFixture(),
//	}
//	b3 := BuildProposal(4, qc3, nil, "1")
//	bc.AddBlock(b3)
//	grandParent, err := bc.GetGrandParentBlock(b3.ID)
//	require.Error(t, err)
//	require.Nil(t, grandParent)
//}
//
//// add one block and commit the block
//func TestCommitBlock1(t *testing.T) {
//	bc := NewBlockchain(10)
//	qc1 := &QC{
//		View:    0,
//		BlockID: utils.IdentifierFixture(),
//	}
//	b1 := BuildProposal(1, qc1, nil, "1")
//	bc.AddBlock(b1)
//	blocks, err := bc.CommitBlock(b1.ID)
//	require.NoError(t, err)
//	require.Equal(t, 1, len(blocks))
//	require.Equal(t, b1, blocks[0])
//	exists := bc.forrest.HasVertex(blocks[0].ID)
//	require.True(t, exists)
//}
//
//// add two blocks and commit the blocks
//func TestCommitBlock2(t *testing.T) {
//	bc := NewBlockchain(10)
//	qc1 := &QC{
//		View:    0,
//		BlockID: utils.IdentifierFixture(),
//	}
//	b1 := BuildProposal(1, qc1, nil, "1")
//	bc.AddBlock(b1)
//	qc2 := &QC{
//		View:    1,
//		BlockID: b1.ID,
//	}
//	b2 := BuildProposal(2, qc2, nil, "1")
//	bc.AddBlock(b2)
//	blocks, err := bc.CommitBlock(b2.ID)
//	require.NoError(t, err)
//	require.Equal(t, 2, len(blocks))
//	require.Equal(t, b2, blocks[0])
//	require.Equal(t, b1, blocks[1])
//	exists := bc.forrest.HasVertex(b1.ID)
//	require.False(t, exists)
//	exists = bc.forrest.HasVertex(b2.ID)
//	require.True(t, exists)
//}
//
//// add three blocks with a fork
//func TestCommitBlock3(t *testing.T) {
//	bc := NewBlockchain(10)
//	qc1 := &QC{
//		View:    0,
//		BlockID: utils.IdentifierFixture(),
//	}
//	b1 := BuildProposal(1, qc1, nil, "1")
//	bc.AddBlock(b1)
//	qc2 := &QC{
//		View:    1,
//		BlockID: b1.ID,
//	}
//	b2 := BuildProposal(2, qc2, nil, "1")
//	bc.AddBlock(b2)
//	qc3 := &QC{
//		View:    0,
//		BlockID: utils.IdentifierFixture(),
//	}
//	b3 := BuildProposal(1, qc3, nil, "1")
//	bc.AddBlock(b3)
//	blocks, err := bc.CommitBlock(b2.ID)
//	require.NoError(t, err)
//	require.Equal(t, 2, len(blocks))
//	require.Equal(t, b2, blocks[0])
//	require.Equal(t, b1, blocks[1])
//}
