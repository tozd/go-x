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

	ctx := context.Background()

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

	ctx := context.Background()

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

	ctx := context.Background()

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
