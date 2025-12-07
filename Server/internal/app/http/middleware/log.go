package middleware

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"strings"
)

type LogMux struct {
	h http.Handler
}

func NewLogMux(h http.Handler) http.Handler {
	return &LogMux{h: h}
}

type loggingResponseWriter struct {
	http.ResponseWriter
	status int
	buf    bytes.Buffer
}

func (lrw *loggingResponseWriter) WriteHeader(statusCode int) {
	lrw.status = statusCode
	lrw.ResponseWriter.WriteHeader(statusCode)
}

func (lrw *loggingResponseWriter) Write(b []byte) (int, error) {
	// Copy response body into buffer (for logging) and pass through.
	_, _ = lrw.buf.Write(b)
	return lrw.ResponseWriter.Write(b)
}

func (m *LogMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Логируем стартовую строку и заголовки всегда.
	dumpR, err := httputil.DumpRequest(r, false)
	if err != nil {
		fmt.Println("Failed to dump request:", err.Error())
	} else {
		fmt.Println("Request:", string(dumpR))
	}

	// Не трогаем WebSocket‑подключения, чтобы не ломать апгрейд.
	connHdr := strings.ToLower(r.Header.Get("Connection"))
	upgradeHdr := strings.ToLower(r.Header.Get("Upgrade"))
	if strings.Contains(connHdr, "upgrade") && upgradeHdr == "websocket" {
		m.h.ServeHTTP(w, r)
		return
	}

	// Считаем, что "не фото" — это не multipart и не image/*.
	ct := r.Header.Get("Content-Type")
	isBinaryUpload := strings.HasPrefix(ct, "multipart/form-data") ||
		strings.HasPrefix(ct, "image/")

	var bodyCopy []byte
	if !isBinaryUpload && r.Body != nil {
		// Прочитать тело, залогировать и восстановить для хендлеров.
		bodyCopy, err = io.ReadAll(r.Body)
		if err != nil {
			fmt.Println("Failed to read request body:", err.Error())
		} else if len(bodyCopy) > 0 {
			fmt.Println("Request body:", string(bodyCopy))
		}
		r.Body = io.NopCloser(bytes.NewReader(bodyCopy))
	}

	lrw := &loggingResponseWriter{
		ResponseWriter: w,
		status:         http.StatusOK,
	}

	m.h.ServeHTTP(lrw, r)

	if !isBinaryUpload {
		fmt.Printf("Response status: %d\n", lrw.status)
		if lrw.buf.Len() > 0 {
			fmt.Println("Response body:", lrw.buf.String())
		}
	}
}
