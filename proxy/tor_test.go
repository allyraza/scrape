package proxy

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"testing"
)

func Test_TorGetIP(t *testing.T) {
	schema := map[string]string{}

	res, err := http.Get("http://proxy01.ams.local/ip/")
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)

	json.Unmarshal(body, &schema)

	u, err := url.Parse(schema["ip"])
	u.Scheme = "http"

	fmt.Printf("%#v\n", u.String())
}
