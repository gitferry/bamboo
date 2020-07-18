package mempool

import (
	"github.com/gitferry/zeitgeber/message"
)

type MemPool struct {
	txns []message.Request
}

func (m *MemPool) GetPayload() []message.Request {
	var payload []message.Request
	return payload
}

func (m *MemPool) StoreTxn(request message.Request) {
	m.txns = append(m.txns, request)
}
