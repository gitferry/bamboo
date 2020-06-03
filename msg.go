package zeitgeber

type WishArgs struct {
	NodeID       int
	VerifiedView int
	WishView     int
}

type WishReply struct {
	VerifiedView int
	Success      bool
}

type ViewSyncArgs struct {
	NodeID       int
	VerifiedView int
}

type ViewSyncReply struct {
	VerifiedView int
	Success      bool
}
