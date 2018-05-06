package proxy

import (
	"log"
	"net/http"
	"net/url"

	"github.com/gocolly/colly"
	"github.com/sycamoreone/orc/control"
)

// Tor @todo
type Tor struct {
	Server  string
	Control string
}

// GetProxy fetches a proxy and returns
func (t *Tor) GetProxy(_ *http.Request) (*url.URL, error) {
	c, err := control.Dial(t.Control)
	if err != nil {
		log.Println(err)
	}

	err = c.Auth("")
	if err != nil {
		log.Println(err)
	}

	err = c.Signal(control.SignalNewNym)
	if err != nil {
		log.Println(err)
	}

	u, err := url.Parse(t.Server)
	if err != nil {
		log.Printf("PROXY: %v", err)
		return nil, err
	}

	return u, nil
}

// TorProxySwitcher creates a proxy switcher func which fetches
// ProxyURLs on every request.
func TorProxySwitcher(server string, control string) (colly.ProxyFunc, error) {
	s := &Tor{Server: server, Control: control}
	return s.GetProxy, nil
}
