package x_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/tozd/go/x"
)

const responseBody = "Hello, client\n"

func TestRetryableResponseSimple(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, responseBody) //nolint:errcheck
	}))
	defer ts.Close()

	ctx := t.Context()

	client := retryablehttp.NewClient()
	req, err := retryablehttp.NewRequestWithContext(ctx, http.MethodGet, ts.URL, nil)
	require.NoError(t, err)
	require.NotNil(t, req)

	res, errE := x.NewRetryableResponse(client, req)
	require.NoError(t, errE, "% -+#.1v", err)
	require.NotNil(t, res)
	defer res.Close() //nolint:errcheck

	assert.Equal(t, int64(14), res.ContentLength)
	assert.Equal(t, int64(14), res.Size())

	response, err := io.ReadAll(res)
	// It can be an error with details.
	require.NoError(t, err, "% -+#.1v", err)
	assert.Equal(t, responseBody, string(response))
}

func TestRetryableResponseRetry(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if reqRange := r.Header.Get("Range"); reqRange != "" {
			if assert.True(t, strings.HasPrefix(reqRange, "bytes=")) {
				reqRange = strings.TrimPrefix(reqRange, "bytes=")
				rs := strings.Split(reqRange, "-")
				if assert.Len(t, rs, 2) {
					end := rs[1]
					assert.Empty(t, end)
					start, err := strconv.Atoi(rs[0])
					if assert.NoError(t, err) {
						assert.Equal(t, 6, start)
						rest := responseBody[start:]
						w.Header().Set("Content-Length", strconv.Itoa(len(rest)))
						w.WriteHeader(http.StatusPartialContent)
						// Send the rest.
						fmt.Fprint(w, rest) //nolint:errcheck
					}
				}
			}
		} else {
			w.Header().Set("Content-Length", strconv.Itoa(len(responseBody)))
			w.WriteHeader(http.StatusOK)
			// Send only the first 6 bytes.
			fmt.Fprint(w, responseBody[0:6]) //nolint:errcheck
		}
	}))
	defer ts.Close()

	ctx := t.Context()

	client := retryablehttp.NewClient()
	req, err := retryablehttp.NewRequestWithContext(ctx, http.MethodGet, ts.URL, nil)
	require.NoError(t, err)
	require.NotNil(t, req)

	res, errE := x.NewRetryableResponse(client, req)
	require.NoError(t, err, "% -+#.1v", errE)
	require.NotNil(t, res)
	defer res.Close() //nolint:errcheck

	assert.Equal(t, int64(14), res.ContentLength)
	assert.Equal(t, int64(14), res.Size())

	data, err := io.ReadAll(res)
	// It can be an error with details.
	require.NoError(t, err, "% -+#.1v", err)
	assert.Equal(t, responseBody, string(data))
}

func TestRetryableResponseRetryWithContentRange(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if reqRange := r.Header.Get("Range"); reqRange != "" {
			if assert.True(t, strings.HasPrefix(reqRange, "bytes=")) {
				reqRange = strings.TrimPrefix(reqRange, "bytes=")
				rs := strings.Split(reqRange, "-")
				if assert.Len(t, rs, 2) {
					end := rs[1]
					assert.Empty(t, end)
					start, err := strconv.Atoi(rs[0])
					if assert.NoError(t, err) {
						assert.Equal(t, 6, start)
						rest := responseBody[start:]
						w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, len(responseBody)-1, len(responseBody)))
						w.WriteHeader(http.StatusPartialContent)
						// Send the rest.
						fmt.Fprint(w, rest) //nolint:errcheck
						if f, ok := w.(http.Flusher); ok {
							// Forcing flush to not have Content-Length header set by Go.
							f.Flush()
						}
					}
				}
			}
		} else {
			w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", 0, len(responseBody)-1, len(responseBody)))
			w.WriteHeader(http.StatusOK)
			// Send only the first 6 bytes.
			fmt.Fprint(w, responseBody[0:6]) //nolint:errcheck
			if f, ok := w.(http.Flusher); ok {
				// Forcing flush to not have Content-Length header set by Go.
				f.Flush()
			}
		}
	}))
	defer ts.Close()

	ctx := t.Context()

	client := retryablehttp.NewClient()
	req, err := retryablehttp.NewRequestWithContext(ctx, http.MethodGet, ts.URL, nil)
	require.NoError(t, err)
	require.NotNil(t, req)

	res, errE := x.NewRetryableResponse(client, req)
	require.NoError(t, errE, "% -+#.1v", errE)
	require.NotNil(t, res)
	defer res.Close() //nolint:errcheck

	assert.Equal(t, int64(-1), res.ContentLength)
	assert.Equal(t, int64(14), res.Size())

	data, err := io.ReadAll(res)
	// It can be an error with details.
	require.NoError(t, err, "% -+#.1v", err)
	assert.Equal(t, responseBody, string(data))
}

func TestRetryableClient(t *testing.T) {
	t.Parallel()

	client := retryablehttp.NewClient()
	baseClient := client.StandardClient()
	client2 := x.RetryableClient(baseClient)
	assert.Equal(t, client, client2)
}

func TestRetryableClientNil(t *testing.T) {
	t.Parallel()

	// A plain http.Client (not from retryablehttp) should return nil.
	result := x.RetryableClient(&http.Client{})
	assert.Nil(t, result)
}

func TestResponseSize(t *testing.T) {
	t.Parallel()

	t.Run("with content length", func(t *testing.T) {
		t.Parallel()

		resp := &http.Response{
			ContentLength: 42,
			Header:        http.Header{},
		}
		size, errE := x.ResponseSize(resp)
		require.NoError(t, errE, "% -+#.1v", errE)
		assert.Equal(t, int64(42), size)
	})

	t.Run("with x-goog header", func(t *testing.T) {
		t.Parallel()

		resp := &http.Response{
			ContentLength: -1,
			Header: http.Header{
				"X-Goog-Stored-Content-Length": []string{"100"},
			},
		}
		size, errE := x.ResponseSize(resp)
		require.NoError(t, errE, "% -+#.1v", errE)
		assert.Equal(t, int64(100), size)
	})

	t.Run("missing size", func(t *testing.T) {
		t.Parallel()

		resp := &http.Response{
			ContentLength: -1,
			Header:        http.Header{},
		}
		size, errE := x.ResponseSize(resp)
		require.Error(t, errE)
		assert.Equal(t, int64(0), size)
		assert.ErrorIs(t, errE, x.ErrResponseMissingSize)
	})
}

func TestNewRetryableResponseBadStatus(t *testing.T) {
	t.Parallel()

	// Use 403 Forbidden: retryablehttp passes it through as a response (not an error),
	// but our code treats non-200 as ErrResponseBadStatus.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer ts.Close()

	client := retryablehttp.NewClient()
	client.RetryMax = 0
	client.Logger = nil
	req, err := retryablehttp.NewRequestWithContext(t.Context(), http.MethodGet, ts.URL, nil)
	require.NoError(t, err)

	res, errE := x.NewRetryableResponse(client, req)
	require.Error(t, errE)
	assert.Nil(t, res)
	assert.ErrorIs(t, errE, x.ErrResponseBadStatus)
}

func TestNewRetryableResponseNetworkError(t *testing.T) {
	t.Parallel()

	// Start and immediately close a server so the port is unavailable.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))
	addr := ts.URL
	ts.Close()

	client := retryablehttp.NewClient()
	client.RetryMax = 0
	client.Logger = nil
	req, err := retryablehttp.NewRequestWithContext(t.Context(), http.MethodGet, addr, nil)
	require.NoError(t, err)

	res, errE := x.NewRetryableResponse(client, req)
	require.Error(t, errE)
	assert.Nil(t, res)
}

func TestRetryableResponseReadAfterClose(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, responseBody) //nolint:errcheck
	}))
	defer ts.Close()

	client := retryablehttp.NewClient()
	req, err := retryablehttp.NewRequestWithContext(t.Context(), http.MethodGet, ts.URL, nil)
	require.NoError(t, err)

	res, errE := x.NewRetryableResponse(client, req)
	require.NoError(t, errE, "% -+#.1v", errE)
	require.NotNil(t, res)

	require.NoError(t, res.Close())

	_, err = res.Read(make([]byte, 10))
	assert.ErrorIs(t, err, x.ErrResponseClosed)
}

func TestRetryableResponseContextCancel(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	started := make(chan struct{})

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "100")
		w.WriteHeader(http.StatusOK)
		// Send partial data and signal the test.
		_, _ = w.Write(make([]byte, 5))
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
		close(started)
		// Block until request context is cancelled.
		<-r.Context().Done()
	}))
	defer ts.Close()

	client := retryablehttp.NewClient()
	client.RetryMax = 0
	client.Logger = nil
	req, err := retryablehttp.NewRequestWithContext(ctx, http.MethodGet, ts.URL, nil)
	require.NoError(t, err)

	res, errE := x.NewRetryableResponse(client, req)
	require.NoError(t, errE, "% -+#.1v", errE)
	require.NotNil(t, res)
	defer res.Close() //nolint:errcheck

	// Read the partial data that was sent.
	buf := make([]byte, 5)
	_, err = io.ReadFull(res, buf)
	require.NoError(t, err)

	// Wait for the server to signal before cancelling.
	<-started
	cancel()

	// Next read should fail with context cancellation error.
	_, err = res.Read(buf)
	assert.Error(t, err)
}
