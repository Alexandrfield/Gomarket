package handle

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
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

func (han *ServiceHandler) WithLogging(h http.HandlerFunc) http.HandlerFunc {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		uri := r.RequestURI
		method := r.Method

		responseData := &responseData{
			status: 0,
			size:   0,
		}
		han.Logger.Debugf("------> Request: %s", r)
		tokenString := r.Header.Get("Authorization")
		if tokenString == "" {
			w.WriteHeader(http.StatusUnauthorized)
			han.Logger.Debugf("Missing authorization header")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		//tokenString = tokenString[len("Bearer "):]

		isValidToken, err := han.authServer.CheckToken(tokenString)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			han.Logger.Debugf("Error check Token. err:%w", err)
		}
		if !isValidToken {
			w.WriteHeader(http.StatusUnauthorized)
			han.Logger.Debugf("Token is not valid")
		}
		var lw loggingResponseWriter
		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			han.Logger.Debugf("try use gzip")
			gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
			if err != nil {
				_, _ = io.WriteString(w, err.Error())
				han.Logger.Debugf("gzip.NewWriterLevel error:%w", err)
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
			han.Logger.Debugf("not use gzip")
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
		han.Logger.Infof("uri:%s; method:%s; status:%d; size:%d; duration:%s;",
			uri, method, responseData.status, responseData.size, duration)
	}
	return http.HandlerFunc(logFn)
}
