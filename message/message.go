package message

import (
	"encoding/gob"
	"fmt"
	"github.com/gitferry/bamboo/crypto"
	"time"

	"github.com/gitferry/bamboo/db"
	"github.com/gitferry/bamboo/identity"
)

func init() {
	gob.Register(Transaction{})
	gob.Register(Query{})
	gob.Register(QueryReply{})
	gob.Register(Read{})
	gob.Register(ReadReply{})
	gob.Register(Register{})
}

/***************************
 * Client-Replica Messages *
 ***************************/

// Transaction is client reqeust with http response channel
type Transaction struct {
	Command   db.Command
	Timestamp time.Time
	NodeID    identity.NodeID // forward by node
	ID        string
}

func (r Transaction) String() string {
	return fmt.Sprintf("Transaction {cmd=%v nid=%v}", r.Command, r.NodeID)
}

// Read can be used as a special request that directly read the value of key without go through replication protocol in Replica
type Read struct {
	CommandID int
	Key       db.Key
}

func (r Read) String() string {
	return fmt.Sprintf("Read {cid=%d, key=%d}", r.CommandID, r.Key)
}

type MissingMBRequest struct {
	RequesterID   identity.NodeID
	ProposalID    crypto.Identifier
	MissingMBList []crypto.Identifier
}

//type Ack struct {
//SentTime time.Time
//AckTime  time.Time
//BackTime time.Time
//Receiver identity.NodeID
//Sig      crypto.Signature
//ID       crypto.Identifier
//Type     string
//}

// ReadReply cid and value of reading key
type ReadReply struct {
	CommandID int
	Value     db.Value
}

// Query can be used as a special request that directly read the value of key without go through replication protocol in Replica
type Query struct {
	C chan QueryReply
}

func (r *Query) Reply(reply QueryReply) {
	r.C <- reply
}

// QueryReply cid and value of reading key
type QueryReply struct {
	Info string
}

/**************************
 *     Config Related     *
 **************************/

// Register message type is used to register self (node or client) with master node
type Register struct {
	Client bool
	ID     identity.NodeID
	Addr   string
}
