package oauth

import (
	"net/http"
	"net/url"
)

const sessionIDHeader = "X-OMA-Session-ID"

type sessionHeaderTransport struct {
	base      http.RoundTripper
	origin    *url.URL
	sessionID string
}

func WithSessionHeader(client *http.Client, mcpURL *url.URL, sessionID string) *http.Client {
	if sessionID == "" {
		return client
	}
	if client == nil {
		client = http.DefaultClient
	}
	cloned := *client
	base := client.Transport
	if base == nil {
		base = http.DefaultTransport
	}
	cloned.Transport = &sessionHeaderTransport{base: base, origin: mcpURL.Clone(), sessionID: sessionID}
	return &cloned
}

func (t *sessionHeaderTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	cloned := req.Clone(req.Context())
	cloned.Header.Del(sessionIDHeader)
	if sameURLOrigin(cloned.URL, t.origin) {
		cloned.Header.Set(sessionIDHeader, t.sessionID)
	}
	return t.base.RoundTrip(cloned)
}
