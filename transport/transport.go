package transport

import (
	"bytes"
	"encoding/gob"
	"errors"
	"flag"
	"github.com/gitferry/bamboo/log"
	"net"
	"net/url"
	"strings"
	"sync"
	"time"
)

var Scheme = flag.String("transport", "tcp", "transport scheme (tcp, udp, chan), default tcp")

// Transport = transport + pipe + client + server
type Transport interface {
	// Scheme returns tranport scheme
	Scheme() string

	// Send sends message into t.send chan
	Send(interface{})

	// Recv waits for message from t.recv chan
	Recv() interface{}

	// Dial connects to remote server non-blocking once connected
	Dial() error

	// Listen waits for connections, non-blocking once listener starts
	Listen()

	SendBitsCount() int64

	RecvBitsCount() int64

	// Close closes send channel and stops listener
	Close()
}

// NewTransport creates new transport object with url
func NewTransport(addr string) Transport {
	if !strings.Contains(addr, "://") {
		addr = *Scheme + "://" + addr
	}
	uri, err := url.Parse(addr)
	if err != nil {
		log.Fatalf("error parsing address %s : %s\n", addr, err)
	}

	transport := &transport{
		uri:   uri,
		send:  make(chan interface{}, 1024),
		recv:  make(chan interface{}, 1024),
		close: make(chan struct{}),
	}

	switch uri.Scheme {
	case "chan":
		t := new(channel)
		t.transport = transport
		return t
	case "tcp":
		t := new(tcp)
		t.transport = transport
		return t
	case "udp":
		t := new(udp)
		t.transport = transport
		return t
	default:
		log.Fatalf("unknown scheme %s", uri.Scheme)
	}
	return nil
}

type transport struct {
	uri              *url.URL
	send             chan interface{}
	recv             chan interface{}
	startSendingTime time.Time
	startRecvTime    time.Time
	totalSentBits    int64
	totalRecvBits    int64
	close            chan struct{}
}

func (t *transport) Send(m interface{}) {
	t.send <- m
}

func (t *transport) Recv() interface{} {
	return <-t.recv
}

func (t *transport) Close() {
	close(t.send)
	close(t.close)
}

func (t *transport) Scheme() string {
	return t.uri.Scheme
}

func (t *transport) Dial() error {
	conn, err := net.Dial(t.Scheme(), t.uri.Host)
	if err != nil {
		return err
	}
	t.startSendingTime = time.Now()

	go func(conn net.Conn) {
		// w := bufio.NewWriter(conn)
		// codec := NewCodec(config.Codec, conn)
		encoder := gob.NewEncoder(conn)
		defer conn.Close()
		for m := range t.send {
			err := encoder.Encode(&m)
			if err != nil {
				log.Error(err)
			}
			var buf bytes.Buffer
			enc := gob.NewEncoder(&buf)
			enc.Encode(&m)
			t.totalSentBits += int64(buf.Len()) * 8
		}
	}(conn)

	return nil
}

func (t *transport) SendBitsCount() int64 {
	rate := int64(float64(t.totalSentBits) / time.Now().Sub(t.startSendingTime).Seconds())
	t.totalSentBits = 0
	t.startSendingTime = time.Now()
	return rate
}

func (t *transport) RecvBitsCount() int64 {
	rate := int64(float64(t.totalRecvBits) / time.Now().Sub(t.startRecvTime).Seconds())
	t.totalRecvBits = 0
	t.startRecvTime = time.Now()
	return rate
}

/******************************
/*     TCP communication      *
/******************************/
type tcp struct {
	*transport
}

func (t *tcp) Listen() {
	log.Debug("start listening ", t.uri.Port())
	listener, err := net.Listen("tcp", ":"+t.uri.Port())
	if err != nil {
		log.Fatal("TCP Listener error: ", err)
	}
	t.startRecvTime = time.Now()

	go func(listener net.Listener) {
		defer listener.Close()
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Error("TCP Accept error: ", err)
				continue
			}

			go func(conn net.Conn) {
				// codec := NewCodec(config.Codec, conn)
				decoder := gob.NewDecoder(conn)
				defer conn.Close()
				//r := bufio.NewReader(conn)
				for {
					select {
					case <-t.close:
						return
					default:
						var m interface{}
						err := decoder.Decode(&m)
						if err != nil {
							log.Error(err)
							continue
						}
						t.recv <- m
						var buf bytes.Buffer
						enc := gob.NewEncoder(&buf)
						enc.Encode(&m)
						t.totalRecvBits += int64(buf.Len()) * 8
					}
				}
			}(conn)
		}
	}(listener)
}

/******************************
/*     UDP communication      *
/******************************/
type udp struct {
	*transport
}

func (u *udp) Dial() error {
	addr, err := net.ResolveUDPAddr("udp", u.uri.Host)
	if err != nil {
		log.Fatal("UDP resolve address error: ", err)
	}
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return err
	}

	u.startSendingTime = time.Now()

	go func(conn *net.UDPConn) {
		// packet := make([]byte, 1500)
		// w := bytes.NewBuffer(packet)
		w := new(bytes.Buffer)
		for m := range u.send {
			gob.NewEncoder(w).Encode(&m)
			_, err := conn.Write(w.Bytes())
			if err != nil {
				log.Error(err)
			}
			u.totalSentBits += int64(w.Len()) * 8
			w.Reset()
		}
	}(conn)

	return nil
}

func (u *udp) Listen() {
	addr, err := net.ResolveUDPAddr("udp", ":"+u.uri.Port())
	if err != nil {
		log.Fatal("UDP resolve address error: ", err)
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatal("UDP Listener error: ", err)
	}
	u.startRecvTime = time.Now()
	go func(conn *net.UDPConn) {
		packet := make([]byte, 1500)
		defer conn.Close()
		for {
			select {
			case <-u.close:
				return
			default:
				_, err := conn.Read(packet)
				if err != nil {
					log.Error(err)
					continue
				}
				r := bytes.NewReader(packet)
				u.totalRecvBits += int64(r.Len()) * 8
				var m interface{}
				gob.NewDecoder(r).Decode(&m)
				u.recv <- m
			}
		}
	}(conn)
}

/*******************************
/* Intra-process communication *
/*******************************/

var chans = make(map[string]chan interface{})
var chansLock sync.RWMutex

type channel struct {
	*transport
}

func (c *channel) Scheme() string {
	return "chan"
}

func (c *channel) Dial() error {
	chansLock.RLock()
	defer chansLock.RUnlock()
	conn, ok := chans[c.uri.Host]
	if !ok {
		return errors.New("server not ready")
	}
	c.startSendingTime = time.Now()
	go func(conn chan<- interface{}) {
		for m := range c.send {
			conn <- m
		}
	}(conn)
	return nil
}

func (c *channel) Listen() {
	chansLock.Lock()
	defer chansLock.Unlock()
	chans[c.uri.Host] = make(chan interface{}, 1024)
	c.startRecvTime = time.Now()
	go func(conn <-chan interface{}) {
		for {
			select {
			case <-c.close:
				return
			case m := <-conn:
				c.recv <- m
			}
		}
	}(chans[c.uri.Host])
}
