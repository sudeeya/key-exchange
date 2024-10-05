package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
	"go.uber.org/zap/buffer"
)

const logTemplate = `
==============================================================================
~~~ REQUEST ~~~
%s  %s  %s
HOST: %s
BODY:
%s

------------------------------------------------------------------------------
~~~ RESPONSE ~~~
STATUS       : %d
RECEIVED AT  : %s
TIME DURATION: %v
BODY         :
%s

==============================================================================
`

func WithLogging(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			reqBody, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			r.Body = io.NopCloser(bytes.NewBuffer(reqBody))

			reqJSON, err := formatWithIndent(reqBody)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			rw := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			var respBody buffer.Buffer
			rw.Tee(&respBody)

			next.ServeHTTP(rw, r)

			respJSON, _ := formatWithIndent(respBody.Bytes())

			logger.Sugar().Infof(logTemplate,
				r.Method, r.URL.Path, r.Proto, r.Host, string(reqJSON),
				rw.Status(), start.Format(time.RFC3339Nano), time.Since(start), string(respJSON),
			)
		})
	}
}

func formatWithIndent(raw []byte) ([]byte, error) {
	var data map[string]any
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, err
	}

	result, err := json.MarshalIndent(data, "", "   ")
	if err != nil {
		return nil, err
	}

	return result, nil
}
