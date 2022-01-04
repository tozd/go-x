package x

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/hashicorp/go-retryablehttp"
	"gitlab.com/tozd/go/errors"
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
	*http.Response
}

// Read implements io.Reader for RetryableResponse.
//
// Use this to read the response
// body and not RetryableResponse.Response.Body.Read.
func (d *RetryableResponse) Read(p []byte) (int, error) {
	n, err := d.Response.Body.Read(p)
	d.count += int64(n)
	if d.count == d.size {
		// We read everything, just return as-is.
		return n, err
	} else if d.count > d.size {
		if err != nil {
			return n, errors.Wrapf(err, "read beyond the expected end of the response body (%d vs. %d)", d.count, d.size)
		}
		return n, errors.Errorf("read beyond the expected end of the response body (%d vs. %d)", d.count, d.size)
	} else if contextErr := d.req.Context().Err(); contextErr != nil {
		// Do not retry on context.Canceled or context.DeadlineExceeded.
		return n, contextErr
	} else if err != nil {
		// We have not read everything, but we got an error. We retry.
		errStart := d.start(d.count)
		if errStart != nil {
			return n, errStart
		}
		if n > 0 {
			return n, nil
		}
		return d.Read(p)
	} else {
		// Something else, just return as-is.
		return n, err
	}
}

// Count implements counter interface for RetryableResponse.
//
// It returns the number of bytes read until now.
func (d *RetryableResponse) Count() int64 {
	return d.size
}

// Size returns the expected number of bytes to read.
func (d *RetryableResponse) Size() int64 {
	return d.size
}

// Close implements io.Closer interface for RetryableResponse.
//
// It closes the underlying response body.
func (d *RetryableResponse) Close() error {
	if d.Response != nil {
		err := errors.WithStack(d.Response.Body.Close())
		d.Response = nil
		return err
	}
	return nil
}

func (d *RetryableResponse) start(from int64) errors.E {
	d.Close()
	if from > 0 {
		d.req.Header.Set("Range", fmt.Sprintf("bytes=%d-", from))
	} else {
		d.req.Header.Del("Range")
	}
	resp, err := d.client.Do(d.req) //nolint:bodyclose
	if err != nil {
		return errors.WithStack(err)
	}
	if (from > 0 && resp.StatusCode != http.StatusPartialContent) || (from <= 0 && resp.StatusCode != http.StatusOK) {
		body, _ := io.ReadAll(resp.Body)
		return errors.Errorf("bad response status (%s): %s", resp.Status, strings.TrimSpace(string(body)))
	}
	d.Response = resp
	lengthStr := resp.Header.Get("Content-Length")
	if lengthStr == "" {
		return errors.Errorf("missing Content-Length header in response")
	}
	length, err := strconv.ParseInt(lengthStr, 10, 64) //nolint:gomnd
	if err != nil {
		return errors.WithStack(err)
	}
	if from > 0 {
		if d.count+length != d.size {
			return errors.Errorf("content after retry has different length (%d) than before (%d)", d.count+length, d.size)
		}
	} else {
		d.size = length
	}
	return nil
}

// NewRetryableResponse returns a RetryableResponse given the client and request to do (and potentially retry).
func NewRetryableResponse(client *retryablehttp.Client, req *retryablehttp.Request) (*RetryableResponse, errors.E) {
	r := &RetryableResponse{
		client:   client,
		req:      req,
		count:    0,
		size:     0,
		Response: nil,
	}
	err := r.start(0)
	if err != nil {
		return nil, err
	}
	return r, nil
}
