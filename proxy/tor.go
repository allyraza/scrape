package proxy

import (
	"log"
	"net/http"
	"net/url"

	"github.com/gocolly/colly"
	"github.com/sycamoreone/orc/control"
)

// Tor holds tor address and a conn to tor control
type Tor struct {
	Address string
	Conn    *control.Conn
}

// GetProxy fetches a proxy and returns
func (t *Tor) GetProxy(_ *http.Request) (*url.URL, error) {
	err := t.Conn.Auth("")
	if err != nil {
		log.Println(err)
	}

	err = t.Conn.Signal(control.SignalNewNym)
	if err != nil {
		log.Println(err)
	}

	u, err := url.Parse(t.Address)
	if err != nil {
		log.Printf("PROXY: %v", err)
		return nil, err
	}

	return u, nil
}

// TorProxySwitcher creates a proxy switcher func which fetches
// ProxyURLs on every request.
func TorProxySwitcher(address string, controlAddress string) (colly.ProxyFunc, error) {
	c, err := control.Dial(controlAddress)
	if err != nil {
		log.Println(err)
	}

	s := &Tor{Address: address, Conn: c}
	return s.GetProxy, nil
}
