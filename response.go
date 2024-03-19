package x

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/hashicorp/go-retryablehttp"
	"gitlab.com/tozd/go/errors"
)

var (
	ErrResponseClosed               = errors.Base("response already closed")
	ErrResponseReadBeyondEnd        = errors.Base("read beyond the expected end of the response body")
	ErrResponseBadStatus            = errors.Base("bad response status")
	ErrResponseMissingContentLength = errors.Base("missing Content-Length header in response")
	ErrResponseLengthMismatch       = errors.Base("content after retry has different length than before")
)

// RetryableResponse reads the response body until it is completely read.
//
// If reading fails before full contents have been read
// (based on the Content-Length header), it transparently retries the request
// using Range request header and continues reading the new response body.
//
// It embeds the current response (so you can access response headers, etc.)
// but the current response can change when the request is retried.
type RetryableResponse struct {
	client *retryablehttp.Client
	req    *retryablehttp.Request
	count  int64
	size   int64
	lock   sync.Mutex
	*http.Response
}

// Read implements io.Reader for RetryableResponse.
//
// Use this to read the response
// body and not RetryableResponse.Response.Body.Read.
func (d *RetryableResponse) Read(p []byte) (int, error) {
	d.lock.Lock()
	resp := d.Response
	d.lock.Unlock()

	if resp == nil {
		return 0, errors.WithStack(ErrResponseClosed)
	}

	n, err := resp.Body.Read(p)
	count := atomic.AddInt64(&d.count, int64(n))

	size := d.Size()
	if count == size {
		// We read everything, just return as-is.
		if err == io.EOF {
			// See: https://github.com/golang/go/issues/39155
			return n, io.EOF
		}
		return n, errors.WithStack(err)
	} else if count > size {
		if err != nil {
			errE := errors.WrapWith(err, ErrResponseReadBeyondEnd)
			errors.Details(errE)["count"] = count
			errors.Details(errE)["size"] = size
			return n, errE
		}
		return n, errors.WithDetails(
			ErrResponseReadBeyondEnd,
			"count", count,
			"size", size,
		)
	} else if contextErr := d.req.Context().Err(); contextErr != nil {
		// Do not retry on context.Canceled or context.DeadlineExceeded.
		if contextErr == io.EOF { //nolint:errorlint
			// See: https://github.com/golang/go/issues/39155
			return n, io.EOF
		}
		return n, errors.WithStack(contextErr)
	} else if err != nil {
		// We have not read everything, but we got an error. We retry.
		errStart := d.start()
		if errStart != nil {
			return n, errStart
		}
		if n > 0 {
			return n, nil
		}
		return d.Read(p)
	}

	// Something else, just return as-is.
	if err == io.EOF {
		// See: https://github.com/golang/go/issues/39155
		return n, io.EOF
	}
	return n, errors.WithStack(err)
}

// Count implements counter interface for RetryableResponse.
//
// It returns the number of bytes read until now.
func (d *RetryableResponse) Count() int64 {
	return atomic.LoadInt64(&d.count)
}

// Size returns the expected number of bytes to read.
func (d *RetryableResponse) Size() int64 {
	return atomic.LoadInt64(&d.size)
}

// Close implements io.Closer interface for RetryableResponse.
//
// It closes the underlying response body.
func (d *RetryableResponse) Close() error {
	d.lock.Lock()
	resp := d.Response
	d.Response = nil
	d.lock.Unlock()

	if resp != nil {
		return errors.WithStack(resp.Body.Close())
	}

	return nil
}

func (d *RetryableResponse) start() errors.E {
	d.Close()

	count := d.Count()
	if count > 0 {
		d.req.Header.Set("Range", fmt.Sprintf("bytes=%d-", count))
	} else {
		d.req.Header.Del("Range")
	}
	resp, err := d.client.Do(d.req) //nolint:bodyclose
	if err != nil {
		return errors.WithStack(err)
	}
	if (count > 0 && resp.StatusCode != http.StatusPartialContent) || (count <= 0 && resp.StatusCode != http.StatusOK) {
		body, _ := io.ReadAll(resp.Body)
		return errors.WithDetails(
			ErrResponseBadStatus,
			"status", resp.Status,
			"body", strings.TrimSpace(string(body)),
		)
	}
	length := resp.ContentLength
	if length == -1 {
		// Check GCP header. GCP omits Content-Length header when response is Content-Encoding compressed.
		l, err := strconv.ParseInt(resp.Header.Get("X-Goog-Stored-Content-Length"), 10, 64)
		if err == nil {
			length = l
		}
	}
	if length == -1 {
		return errors.WithStack(ErrResponseMissingContentLength)
	}

	size := d.Size()
	if count > 0 {
		if count+length != size {
			return errors.WithDetails(
				ErrResponseLengthMismatch,
				"new", count+length,
				"old", size,
			)
		}
	} else {
		atomic.StoreInt64(&d.size, length)
	}

	d.lock.Lock()
	d.Response = resp
	d.lock.Unlock()

	return nil
}

// NewRetryableResponse returns a RetryableResponse given the client and request to do (and potentially retry).
func NewRetryableResponse(client *retryablehttp.Client, req *retryablehttp.Request) (*RetryableResponse, errors.E) {
	r := &RetryableResponse{
		client:   client,
		req:      req,
		count:    0,
		size:     0,
		lock:     sync.Mutex{},
		Response: nil,
	}
	err := r.start()
	if err != nil {
		return nil, err
	}
	return r, nil
}
