package zeitgeber

type MemPool interface {
	GetPayload() []byte
	StoreTxn(request Request)
}

type mempool struct {
	txns []Request
}

func (m *mempool) GetPayload() []byte {
	return []byte{}
}

func (m *mempool) StoreTxn(request Request) {
	m.txns = append(m.txns, request)
}
