package handle

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Alexandrfield/Gomarket/internal/common"
	"github.com/Alexandrfield/Gomarket/internal/server"
)

type (
	responseData struct {
		status int
		size   int
	}
	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size // захватываем размер
	if err != nil {
		err = fmt.Errorf("loggingResponseWriter err:%w", err)
	}
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode // захватываем код статуса
}

func WithLogging(logger common.Logger, config *server.Config, h http.HandlerFunc) http.HandlerFunc {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		uri := r.RequestURI
		method := r.Method

		responseData := &responseData{
			status: 0,
			size:   0,
		}

		var lw loggingResponseWriter
		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			logger.Debugf("try use gzip")
			gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
			if err != nil {
				_, _ = io.WriteString(w, err.Error())
				logger.Debugf("gzip.NewWriterLevel error:%w", err)
			}
			w.Header().Set("Content-Encoding", "gzip")
			defer func() {
				_ = gz.Close()
			}()
			lw = loggingResponseWriter{
				ResponseWriter: gzipWriter{ResponseWriter: w, Writer: gz}, // встраиваем оригинальный http.ResponseWriter
				responseData:   responseData,
			}
		} else {
			logger.Debugf("not use gzip")
			lw = loggingResponseWriter{
				ResponseWriter: w, // встраиваем оригинальный http.ResponseWriter
				responseData:   responseData,
			}
		}
		// data := make([]byte, 10000)
		// n, _ := r.Body.Read(data)
		// data = data[:n]
		// msgSign := r.Header.Get("HashSHA256")
		// if msgSign != "" && len(data) > 0 {
		// 	sig, _ := b64.StdEncoding.DecodeString(msgSign)
		// 	if !common.CheckHash(data, sig, config.SignKey) {
		// 		lw.WriteHeader(http.StatusBadRequest)
		// 	}
		// }
		// r.Body = io.NopCloser(bytes.NewBuffer(data))

		h.ServeHTTP(&lw, r)
		duration := time.Since(start)
		logger.Infof("uri:%s; method:%s; status:%d; size:%d; duration:%s;",
			uri, method, responseData.status, responseData.size, duration)
	}
	return http.HandlerFunc(logFn)
}
