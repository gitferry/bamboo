package mempool

import "github.com/gitferry/zeitgeber"

type MemPool struct {
	txns []zeitgeber.Request
}

func (m *MemPool) GetPayload() []zeitgeber.Request {
	var payload []zeitgeber.Request
	return payload
}

func (m *MemPool) StoreTxn(request zeitgeber.Request) {
	m.txns = append(m.txns, request)
}
