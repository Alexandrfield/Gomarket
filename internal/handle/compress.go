package handle

import (
	"fmt"
	"io"
	"net/http"
)

type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w gzipWriter) Write(b []byte) (int, error) {
	n, err := w.Writer.Write(b)
	if err != nil {
		err = fmt.Errorf("gzipWriter write:%w", err)
	}
	return n, err
}
