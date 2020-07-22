package zeitgeber

import (
	"flag"
	"net/http"

	"github.com/gitferry/zeitgeber/config"
	"github.com/gitferry/zeitgeber/log"
)

func Init() {
	flag.Parse()
	log.Setup()
	config.Configuration.Load()
	http.DefaultTransport.(*http.Transport).MaxIdleConnsPerHost = 1000
}
