package proxy

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"

	"github.com/gocolly/colly"
)

var (
	// TorURL is base url used to retrieve an ip
	TorURL = "http://proxy01.ams.local/ip"
)

// Tor @todo
type Tor struct {
	Endpoint string
}

// GetProxy fetches a proxy and returns
func (t *Tor) GetProxy(_ *http.Request) (*url.URL, error) {
	r, err := http.Get(t.Endpoint)
	if err != nil {
		log.Printf("PROXY: %v", err)
		return nil, err
	}

	schema := map[string]string{}
	json.NewDecoder(r.Body).Decode(&schema)

	u, err := url.Parse(schema["ip"])
	if err != nil {
		log.Printf("PROXY: %v", err)
		return nil, err
	}

	u.Scheme = "http"

	return u, nil
}

// TorProxySwitcher creates a proxy switcher func which fetches
// ProxyURLs on every request.
func TorProxySwitcher(endpoint string) (colly.ProxyFunc, error) {
	s := &Tor{Endpoint: endpoint}
	return s.GetProxy, nil
}
