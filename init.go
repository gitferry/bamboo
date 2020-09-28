package bamboo

import (
	"flag"
	"net/http"

	"github.com/gitferry/bamboo/config"
	"github.com/gitferry/bamboo/log"
)

func Init() {
	flag.Parse()
	log.Setup()
	config.Configuration.Load()
	http.DefaultTransport.(*http.Transport).MaxIdleConnsPerHost = 1000
}
