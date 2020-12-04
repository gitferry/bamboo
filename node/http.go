package node

import (
	"io"
	"net/http"
	"net/url"
	"sync"

	"github.com/gitferry/bamboo/config"
	"github.com/gitferry/bamboo/log"
	"github.com/gitferry/bamboo/message"
)

// http request header names
const (
	HTTPClientID  = "Id"
	HTTPCommandID = "Cid"
)

var ppFree = sync.Pool{
	New: func() interface{} {
		return make(chan message.TransactionReply, 1)
	},
}

// serve serves the http REST API request from clients
func (n *node) http() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", n.handleRoot)
	mux.HandleFunc("/query", n.handleQuery)

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
	//var err error

	req.C = ppFree.Get().(chan message.TransactionReply)
	req.NodeID = n.id // TODO does this work when forward twice
	req.ID = r.RequestURI
	n.TxChan <- req

	reply := <-req.C

	if reply.Err != nil {
		http.Error(w, reply.Err.Error(), http.StatusInternalServerError)
		return
	}
	//_, err = io.WriteString(w, string(reply.Value))
	//if err != nil {
	//	log.Error(err)
	//}
}
