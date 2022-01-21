package bamboo

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/gitferry/bamboo/config"
	"github.com/gitferry/bamboo/db"
	"github.com/gitferry/bamboo/identity"
	"github.com/gitferry/bamboo/log"
	"github.com/gitferry/bamboo/node"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"strconv"
	"sync"
)

// Client interface provides get and put for key value store
type Client interface {
	Get(db.Key) (string, error)
	Put(db.Key, db.Value) error
}

// AdminClient interface provides fault injection opeartion
type AdminClient interface {
	Consensus(db.Key) bool
	Crash(identity.NodeID, int)
	Drop(identity.NodeID, identity.NodeID, int)
	Partition(int, ...identity.NodeID)
}

// HTTPClient implements Client interface with REST API
type HTTPClient struct {
	HTTP map[identity.NodeID]string
	ID   identity.NodeID // client id use the same id as servers in local site
	N    int             // total number of nodes
	Zipf *rand.Zipf
	ids  []identity.NodeID

	CID int // command id
	*http.Client
}

// NewHTTPClient creates a new Client from config
func NewHTTPClient() *HTTPClient {
	c := &HTTPClient{
		N:      len(config.Configuration.Addrs),
		HTTP:   config.Configuration.HTTPAddrs,
		Client: &http.Client{},
	}
	// will not send request to Byzantine nodes
	bzn := config.GetConfig().ByzNo
	if config.GetConfig().Strategy == "silence" {
		for i := 1; i <= bzn; i++ {
			id := identity.NewNodeID(i)
			delete(c.HTTP, id)
		}
	}
	c.ids = make([]identity.NodeID, c.N)
	for i := 0; i < c.N; i++ {
		c.ids[i] = identity.NewNodeID(i + 1)
	}
	if config.GetConfig().Master != "0" {
		delete(c.HTTP, config.GetConfig().Master)
		masterIndex := config.GetConfig().Master.Node() - 1
		c.ids = append(c.ids[:masterIndex], c.ids[masterIndex+1:]...)
	}
	r := rand.New(rand.NewSource(1))
	c.Zipf = rand.NewZipf(r, config.Configuration.ZipfianS, config.Configuration.ZipfianV, uint64(len(c.ids)-1))
	return c
}

// Get gets value of given key (use REST)
// Default implementation of Client interface
func (c *HTTPClient) Get(key db.Key) (string, error) {
	c.CID++
	//v, _, err := c.RESTGet(key)
	//return v, err
	return "", nil
}

// Put puts new key value pair and return previous value (use REST)
// Default implementation of Client interface
func (c *HTTPClient) Put(key db.Key, value db.Value) error {
	c.CID++
	return c.RESTPut(key, value)
}

func (c *HTTPClient) GetURL(key db.Key) (identity.NodeID, string) {
	var replicaID identity.NodeID
	if config.Configuration.Zipf {
		v := c.Zipf.Uint64()
		replicaID = c.ids[v]
	} else {
		replicaID = c.ids[rand.Intn(len(c.ids))]
	}
	log.Debugf("send tx to %v", replicaID)
	return replicaID, c.HTTP[replicaID] + "/" + strconv.Itoa(int(key))
}

// rest accesses server's REST API with url = http://ip:port/key
// if value == nil, it's a read
func (c *HTTPClient) rest(url string, value db.Value) error {
	method := http.MethodGet
	var body io.Reader
	if value != nil {
		method = http.MethodPut
		body = bytes.NewBuffer(value)
	}
	//v, _ := ioutil.ReadAll(body)
	//log.Debugf("payload is %x", v)
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		log.Error(err)
		return err
	}
	req.Header.Set(node.HTTPClientID, string(c.ID))
	req.Header.Set(node.HTTPCommandID, strconv.Itoa(c.CID))
	req.Header.Set("Connection", "keep-alive")
	//log.Debugf("The payload is %x",)

	rep, err := c.Client.Do(req)
	if err != nil {
		log.Error(err)
		return err
	}
	defer rep.Body.Close()

	if rep.StatusCode == http.StatusOK {
		return nil
	}

	// http call failed
	dump, _ := httputil.DumpResponse(rep, true)
	log.Debugf("%q", dump)
	return errors.New(rep.Status)
}

// RESTGet issues a http call to node and return value and headers
//func (c *HTTPClient) RESTGet(key db.Key) (string, error) {
//	return c.rest(key, nil)
//}

// RESTPut puts new value as http.request body and return previous value
func (c *HTTPClient) RESTPut(key db.Key, value db.Value) error {
	//return c.AllPut(key, value)
	return c.SinglePut(key, value)
}

func (c *HTTPClient) json(id identity.NodeID, key db.Key, value db.Value) (db.Value, error) {
	url := c.HTTP[id]
	cmd := db.Command{
		Key:       key,
		Value:     value,
		ClientID:  c.ID,
		CommandID: c.CID,
	}
	data, err := json.Marshal(cmd)
	res, err := c.Client.Post(url, "json", bytes.NewBuffer(data))
	if err != nil {
		log.Error(err)
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusOK {
		b, _ := ioutil.ReadAll(res.Body)
		log.Debugf("key=%v value=%x", key, db.Value(b))
		return db.Value(b), nil
	}
	dump, _ := httputil.DumpResponse(res, true)
	log.Debugf("%q", dump)
	return nil, errors.New(res.Status)
}

// JSONGet posts get request in json format to server url
func (c *HTTPClient) JSONGet(key db.Key) (db.Value, error) {
	return c.json(c.ID, key, nil)
}

// JSONPut posts put request in json format to server url
func (c *HTTPClient) JSONPut(key db.Key, value db.Value) (db.Value, error) {
	return c.json(c.ID, key, value)
}

//// QuorumGet concurrently read values from majority nodes
//func (c *HTTPClient) QuorumGet(key db.Key) ([]string, []map[string]string) {
//	return c.MultiGet(c.N/2+1, key)
//}

// MultiGet concurrently read values from n nodes
//func (c *HTTPClient) MultiGet(n int, key db.Key) ([]string, []map[string]string) {
//	valueC := make(chan string)
//	metaC := make(chan map[string]string)
//	i := 0
//	for id := range c.HTTP {
//		go func(id identity.NodeID) {
//			v, meta, err := c.rest(key, nil)
//			if err != nil {
//				log.Error(err)
//				return
//			}
//			valueC <- v
//			metaC <- meta
//		}(id)
//		i++
//		if i >= n {
//			break
//		}
//	}
//
//	values := make([]string, 0)
//	metas := make([]map[string]string, 0)
//	for ; i > 0; i-- {
//		values = append(values, <-valueC)
//		metas = append(metas, <-metaC)
//	}
//	return values, metas
//}

// QuorumPut concurrently write values to majority of nodes
// TODO get headers
func (c *HTTPClient) AllPut(key db.Key, value db.Value) error {
	var wait sync.WaitGroup
	var err error
	for id, ip := range c.HTTP {
		wait.Add(1)
		go func(id int, ip string) {
			url := ip + "/" + strconv.Itoa(int(key)+id)
			err = c.rest(url, value)
			wait.Done()
		}(id.Node(), ip)
	}
	wait.Wait()
	return err
}

func (c *HTTPClient) SinglePut(key db.Key, value db.Value) error {
	var err error
	_, ip := c.GetURL(key)
	err = c.rest(ip, value)
	return err
}

// Consensus collects /history/key from every node and compare their values
func (c *HTTPClient) Consensus(k db.Key) bool {
	h := make(map[identity.NodeID][]db.Value)
	for id, url := range c.HTTP {
		h[id] = make([]db.Value, 0)
		r, err := c.Client.Get(url + "/history?key=" + strconv.Itoa(int(k)))
		if err != nil {
			log.Error(err)
			continue
		}
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Error(err)
			continue
		}
		holder := h[id]
		err = json.Unmarshal(b, &holder)
		if err != nil {
			log.Error(err)
			continue
		}
		h[id] = holder
		log.Debugf("node=%v key=%v h=%v", id, k, holder)
	}
	n := 0
	for _, v := range h {
		if len(v) > n {
			n = len(v)
		}
	}
	for i := 0; i < n; i++ {
		set := make(map[string]struct{})
		for id := range c.HTTP {
			if len(h[id]) > i {
				set[string(h[id][i])] = struct{}{}
			}
		}
		if len(set) > 1 {
			return false
		}
	}
	return true
}

// Crash stops the node for t seconds then recover
// node crash forever if t < 0
func (c *HTTPClient) Crash(id identity.NodeID, t int) {
	url := c.HTTP[id] + "/crash?t=" + strconv.Itoa(t)
	r, err := c.Client.Get(url)
	if err != nil {
		log.Error(err)
		return
	}
	r.Body.Close()
}

// Drop drops every message send for t seconds
func (c *HTTPClient) Drop(from, to identity.NodeID, t int) {
	url := c.HTTP[from] + "/drop?id=" + string(to) + "&t=" + strconv.Itoa(t)
	r, err := c.Client.Get(url)
	if err != nil {
		log.Error(err)
		return
	}
	r.Body.Close()
}
