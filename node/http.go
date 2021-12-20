package node

import (
	"github.com/gitferry/bamboo/config"
	"github.com/gitferry/bamboo/log"
	"github.com/gitferry/bamboo/message"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"time"
)

// http request header names
const (
	HTTPClientID  = "Id"
	HTTPCommandID = "Cid"
)

// serve serves the http REST API request from clients
func (n *node) http() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", n.handleRoot)
	mux.HandleFunc("/query", n.handleQuery)
	mux.HandleFunc("/slow", n.handleSlow)
	mux.HandleFunc("/flaky", n.handleFlaky)
	mux.HandleFunc("/crash", n.handleCrash)

	// http string should be in form of ":8080"
	ip, err := url.Parse(config.Configuration.HTTPAddrs[n.id])
	if err != nil {
		log.Fatal("http url parse error: ", err)
	}
	port := ":" + ip.Port()
	n.server = &http.Server{
		Addr:    port,
		Handler: mux,
	}
	log.Info("http server starting on ", port)
	log.Fatal(n.server.ListenAndServe())
}

func (n *node) handleQuery(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var query message.Query
	query.C = make(chan message.QueryReply)
	n.TxChan <- query
	reply := <-query.C
	_, err := io.WriteString(w, reply.Info)
	if err != nil {
		log.Error(err)
	}
}

func (n *node) handleRoot(w http.ResponseWriter, r *http.Request) {
	var req message.Transaction
	defer r.Body.Close()

	v, _ := ioutil.ReadAll(r.Body)
	req.Command.Value = v
	req.NodeID = n.id
	req.Timestamp = time.Now()
	req.ID = r.RequestURI
	n.TxChan <- req

	//reply := <-req.C
	//
	//log.Debugf("[%v] tx %v delay is %v", n.id, req.Hash, strconv.Itoa(int(reply.Delay.Nanoseconds())))

	//if reply.Err != nil {
	//	http.Error(w, reply.Err.Error(), http.StatusInternalServerError)
	//	return
	//}
	//w.Header().Set(HTTPCommandID, strconv.Itoa(int(reply.Delay.Nanoseconds())))
	//_, err := io.WriteString(w, string(reply.Delay.Nanoseconds()))
	//if err != nil {
	//	log.Error(err)
	//}
}

func (n *node) handleCrash(w http.ResponseWriter, r *http.Request) {
	n.Socket.Crash(config.GetConfig().Crash)
}

func (n *node) handleSlow(w http.ResponseWriter, r *http.Request) {
	//t, err := strconv.Atoi(r.URL.Query().Get("t"))
	//if err != nil {
	//	log.Error(err)
	//	http.Error(w, "invalide time", http.StatusBadRequest)
	//	return
	//}
	//d, err := strconv.Atoi(r.URL.Query().Get("d"))
	//if err != nil {
	//	log.Error(err)
	//	http.Error(w, "invalide time", http.StatusBadRequest)
	//	return
	//}
	for id, _ := range config.GetConfig().HTTPAddrs {
		n.Socket.Slow(id, rand.Intn(config.GetConfig().Slow), 10)
	}
}

func (n *node) handleFlaky(w http.ResponseWriter, r *http.Request) {
	for id, _ := range config.GetConfig().HTTPAddrs {
		n.Socket.Flaky(id, 0.5, 10)
	}
}
