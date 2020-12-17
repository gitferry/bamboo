package message

import (
	"encoding/gob"
	"fmt"
	"time"

	"github.com/gitferry/bamboo/config"
	"github.com/gitferry/bamboo/db"
	"github.com/gitferry/bamboo/identity"
)

func init() {
	gob.Register(Transaction{})
	gob.Register(TransactionReply{})
	gob.Register(Query{})
	gob.Register(QueryReply{})
	gob.Register(Read{})
	gob.Register(ReadReply{})
	gob.Register(Register{})
	gob.Register(config.Config{})
}

/***************************
 * Client-Replica Messages *
 ***************************/

// Transaction is client reqeust with http response channel
type Transaction struct {
	Command    db.Command
	Properties map[string]string
	Timestamp  time.Time
	NodeID     identity.NodeID // forward by node
	ID         string
	C          chan TransactionReply // reply channel created by request receiver
}

// TransactionReply replies to current client session
func (r *Transaction) Reply(reply TransactionReply) {
	r.C <- reply
}

func (r Transaction) String() string {
	return fmt.Sprintf("Transaction {cmd=%v nid=%v}", r.Command, r.NodeID)
}

// TransactionReply includes all info that might replies to back the client for the coresponding reqeust
type TransactionReply struct {
	Command    db.Command
	Value      db.Value
	Properties map[string]string
	Delay      time.Duration
	Err        error
}

func NewReply(delay time.Duration) TransactionReply {
	return TransactionReply{
		Delay: delay,
	}
}

func (r TransactionReply) String() string {
	return fmt.Sprintf("TransactionReply {cmd=%v value=%x prop=%v}", r.Command, r.Value, r.Properties)
}

// Read can be used as a special request that directly read the value of key without go through replication protocol in Replica
type Read struct {
	CommandID int
	Key       db.Key
}

func (r Read) String() string {
	return fmt.Sprintf("Read {cid=%d, key=%d}", r.CommandID, r.Key)
}

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
