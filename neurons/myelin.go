package neurons

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"
)

// ref:
// https://gist.github.com/CJEnright/bc2d8b8dc0c1389a9feeddb110f822d7
// https://gist.github.com/the42/1956518
// https://gist.github.com/erikdubbelboer/7df2b2b9f34f9f839a84

var gzPool = sync.Pool{
	New: func() interface{} {
		w, _ := gzip.NewWriterLevel(io.Discard, gzip.BestCompression)
		return w
	},
}

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w *gzipResponseWriter) WriteHeader(status int) {
	w.Header().Del("Content-Length")
	w.ResponseWriter.WriteHeader(status)
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	if "" == w.Header().Get("Content-Type") {
		// If no content type, apply sniffing algorithm to un-gzipped body.
		w.Header().Set("Content-Type", http.DetectContentType(b))
	}
	return w.Writer.Write(b)
}

func Gzip(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			http.HandlerFunc(f).ServeHTTP(w, r) // equivalent to f(w, r)
			return
		}

		w.Header().Set("Content-Encoding", "gzip")

		gz := gzPool.Get().(*gzip.Writer)
		defer gzPool.Put(gz)

		gz.Reset(w)
		defer gz.Close()

		r.Header.Del("Accept-Encoding") // prevent double-gzipping from other handlers down the chain
		http.HandlerFunc(f).ServeHTTP(&gzipResponseWriter{ResponseWriter: w, Writer: gz}, r)
	}
}

// since this uses json.NewEncoder().Encode(), the resulting json will naturally have a trailing newline
func GzipJson(w http.ResponseWriter, data interface{}) error {
	gz := gzPool.Get().(*gzip.Writer)
	defer gzPool.Put(gz)
	gz.Reset(w)
	defer gz.Close()

	if err := json.NewEncoder(gz).Encode(data); err != nil {
		return err
	}
	return nil
}
