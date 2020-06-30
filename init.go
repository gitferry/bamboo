package zeitgeber

import (
	"flag"
	"net/http"

	"github.com/gitferry/zeitgeber/log"
)

func Init() {
	flag.Parse()
	log.Setup()
	config.Load()
	http.DefaultTransport.(*http.Transport).MaxIdleConnsPerHost = 1000
}
