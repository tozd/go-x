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

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, responseBody)
	}))
	defer ts.Close()

	ctx := context.Background()

	client := retryablehttp.NewClient()
	req, err := retryablehttp.NewRequestWithContext(ctx, http.MethodGet, ts.URL, nil)
	require.NoError(t, err)
	require.NotNil(t, req)

	res, err := x.NewRetryableResponse(client, req)
	require.NoError(t, err, "% -+#.1v", err)
	require.NotNil(t, res)
	defer res.Close()

	assert.Equal(t, int64(14), res.ContentLength)
	assert.Equal(t, int64(14), res.Size())

	response, err := io.ReadAll(res)
	require.NoError(t, err)
	assert.Equal(t, responseBody, string(response))
}

func TestRetryableResponseRetry(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if reqRange := r.Header.Get("Range"); reqRange != "" {
			require.True(t, strings.HasPrefix(reqRange, "bytes="))
			reqRange = strings.TrimPrefix(reqRange, "bytes=")
			rs := strings.Split(reqRange, "-")
			require.Equal(t, 2, len(rs))
			end := rs[1]
			require.Equal(t, "", end)
			start, err := strconv.Atoi(rs[0])
			require.NoError(t, err)
			require.Equal(t, 6, start)
			rest := responseBody[start:]
			w.Header().Set("Content-Length", strconv.Itoa(len(rest)))
			w.WriteHeader(http.StatusPartialContent)
			// Send the rest.
			fmt.Fprint(w, rest)
		} else {
			w.Header().Set("Content-Length", strconv.Itoa(len(responseBody)))
			w.WriteHeader(http.StatusOK)
			// Send only the first 6 bytes.
			fmt.Fprint(w, responseBody[0:6])
		}
	}))
	defer ts.Close()

	ctx := context.Background()

	client := retryablehttp.NewClient()
	req, err := retryablehttp.NewRequestWithContext(ctx, http.MethodGet, ts.URL, nil)
	require.NoError(t, err)
	require.NotNil(t, req)

	res, err := x.NewRetryableResponse(client, req)
	require.NoError(t, err, "% -+#.1v", err)
	require.NotNil(t, res)
	defer res.Close()

	assert.Equal(t, int64(14), res.ContentLength)
	assert.Equal(t, int64(14), res.Size())

	data, err := io.ReadAll(res)
	require.NoError(t, err)
	assert.Equal(t, responseBody, string(data))
}
