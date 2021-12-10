package blockchain

// add only one vote
//func TestQuorum_SuperMajority1(t *testing.T) {
//	quorum := NewQuorum(4)
//	blockID := utils.IdentifierFixture()
//	v1 := MakeVote(1, "1", blockID)
//	quorum.Add(v1)
//	require.False(t, quorum.SuperMajority(blockID))
//}
//
//// add only two votes
//func TestQuorum_SuperMajority2(t *testing.T) {
//	quorum := NewQuorum(4)
//	blockID := utils.IdentifierFixture()
//	v1 := MakeVote(1, "1", blockID)
//	quorum.Add(v1)
//	v2 := MakeVote(1, "2", blockID)
//	quorum.Add(v2)
//	require.False(t, quorum.SuperMajority(blockID))
//}
//
//// add three votes for the same block from different nodes
//func TestQuorum_SuperMajority3(t *testing.T) {
//	quorum := NewQuorum(4)
//	blockID := utils.IdentifierFixture()
//	v1 := MakeVote(1, "1", blockID)
//	quorum.Add(v1)
//	v2 := MakeVote(1, "2", blockID)
//	quorum.Add(v2)
//	v3 := MakeVote(1, "3", blockID)
//	quorum.Add(v3)
//	require.True(t, quorum.SuperMajority(blockID))
//}
//
//// add three votes, two for the same block from different nodes
//// one for another block
//func TestQuorum_SuperMajority4(t *testing.T) {
//	quorum := NewQuorum(4)
//	blockID := utils.IdentifierFixture()
//	v1 := MakeVote(1, "1", blockID)
//	quorum.Add(v1)
//	v2 := MakeVote(1, "2", blockID)
//	quorum.Add(v2)
//	v3 := MakeVote(1, "3", utils.IdentifierFixture())
//	quorum.Add(v3)
//	require.False(t, quorum.SuperMajority(blockID))
//}
//
//// add three votes, two for the same block from different nodes
//// one from the same voter
//func TestQuorum_SuperMajority5(t *testing.T) {
//	quorum := NewQuorum(4)
//	blockID := utils.IdentifierFixture()
//	v1 := MakeVote(1, "1", blockID)
//	quorum.Add(v1)
//	v2 := MakeVote(1, "2", blockID)
//	quorum.Add(v2)
//	v3 := MakeVote(1, "2", blockID)
//	quorum.Add(v3)
//	require.False(t, quorum.SuperMajority(blockID))
//}
