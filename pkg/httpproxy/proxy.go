package httpproxy

import (
	"context"
	"net"
	"net/http"
)

// HTTPProxy is a http.RoundTripper that proxies requests over a custom net.Conn
type HTTPProxy struct {
	dialer net.Conn
}

// NewHTTPProxy give a new HTTPProxy
func NewHTTPProxy(dialer net.Conn) *HTTPProxy {
	return &HTTPProxy{
		dialer: dialer,
	}
}

func (h *HTTPProxy) RoundTrip(req *http.Request) (*http.Response, error) {
	// reset some data http.Client doesn't like
	req.RequestURI = ""
	req.RemoteAddr = ""

	netClient := &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return h.dialer, nil
			},
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // annoying for a proxy to follow redirects...
		},
	}
	defer netClient.CloseIdleConnections()

	return netClient.Do(req)
}
