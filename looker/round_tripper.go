package looker

import (
	"net/http"
)

type dummyRoundTripper struct {
	Base    http.RoundTripper
	Headers map[string]string
}

// look into custom round tripper to add header
func (d dummyRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("x-looker-appid", "go-sdk")
	for k, v := range d.Headers {
		req.Header.Set(k, v)
	}

	return d.Base.RoundTrip(req)
}

// rt := dummyRoundTripper{
// 	Base: &http.Transport{
// 		TLSClientConfig: &tls.Config{
// 			InsecureSkipVerify: !data.Get("verify_ssl").(bool),
// 		},
// 	},
// 	Headers: map[string]string{
// 		"Accept": "application/json",
// 	},
// }
