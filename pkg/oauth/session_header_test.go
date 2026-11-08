package oauth

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSessionHeaderIsScopedToMCPOrigin(t *testing.T) {
	t.Parallel()
	other := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Empty(t, r.Header.Get(sessionIDHeader))
		w.WriteHeader(http.StatusOK)
	}))
	defer other.Close()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "session-one", r.Header.Get(sessionIDHeader))
		http.Redirect(w, r, other.URL, http.StatusFound)
	}))
	defer server.Close()
	origin, err := url.Parse(server.URL)
	require.NoError(t, err)
	base := server.Client()
	client := WithSessionHeader(base, origin, "session-one")
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, server.URL, nil)
	require.NoError(t, err)
	req.Header.Set(sessionIDHeader, "untrusted")
	resp, err := client.Do(req)
	require.NoError(t, err)
	resp.Body.Close()
	require.Equal(t, "untrusted", req.Header.Get(sessionIDHeader))
	require.NotSame(t, base, client)
	require.Same(t, base, WithSessionHeader(base, origin, ""))
}

func TestSessionHeaderConcurrentClients(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, r.Header.Get(sessionIDHeader))
	}))
	t.Cleanup(server.Close)
	origin, err := url.Parse(server.URL)
	require.NoError(t, err)
	base := server.Client()
	for _, id := range []string{"session-one", "session-two", ""} {
		t.Run(id, func(t *testing.T) {
			t.Parallel()
			client := WithSessionHeader(base, origin, id)
			for range 10 {
				req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, server.URL, nil)
				require.NoError(t, err)
				resp, err := client.Do(req)
				require.NoError(t, err)
				body, err := io.ReadAll(resp.Body)
				resp.Body.Close()
				require.NoError(t, err)
				require.Equal(t, id, string(body))
			}
		})
	}
}
