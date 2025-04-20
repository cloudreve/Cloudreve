package request

import (
	"io"
	"net/http"
	"strconv"
)

var contentLengthHeaders = []string{
	"Content-Length",
	"X-Expected-Entity-Length", // DavFS on MacOS
}

// BlackHole 将客户端发来的数据放入黑洞
func BlackHole(r io.Reader) {
	io.Copy(io.Discard, r)
}

// SniffContentLength tries to get the content length from the request. It also returns
// a reader that will limit to the sniffed content length.
func SniffContentLength(r *http.Request) (LimitReaderCloser, int64, error) {
	for _, header := range contentLengthHeaders {
		if length := r.Header.Get(header); length != "" {
			res, err := strconv.ParseInt(length, 10, 64)
			if err != nil {
				return nil, 0, err
			}

			return newLimitReaderCloser(r.Body, res), res, nil
		}
	}
	return newLimitReaderCloser(r.Body, 0), 0, nil
}

type LimitReaderCloser interface {
	io.Reader
	io.Closer
	Count() int64
}

type limitReaderCloser struct {
	io.Reader
	io.Closer
	read int64
}

func newLimitReaderCloser(r io.ReadCloser, limit int64) LimitReaderCloser {
	return &limitReaderCloser{
		Reader: io.LimitReader(r, limit),
		Closer: r,
	}
}

func (l *limitReaderCloser) Read(p []byte) (n int, err error) {
	n, err = l.Reader.Read(p)
	l.read += int64(n)
	return n, err
}

func (l *limitReaderCloser) Count() int64 {
	return l.read
}
