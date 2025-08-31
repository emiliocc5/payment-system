package http

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
)

func (s *Server) JSONResponse(w http.ResponseWriter, _ *http.Request, result any) {
	body, err := json.Marshal(result)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		s.logger.Error("cannot marshal response", slog.Any("error", err))
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(prettyJSON(body))
}

func (s *Server) JSONResponseCode(w http.ResponseWriter, _ *http.Request, result interface{}, responseCode int) {
	body, err := json.Marshal(result)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		s.logger.Error("JSON marshal failed", slog.Any("error", err))
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(responseCode)
	_, _ = w.Write(prettyJSON(body))
}

func prettyJSON(b []byte) []byte {
	var out bytes.Buffer
	_ = json.Indent(&out, b, "", " ")
	return out.Bytes()
}

func (s *Server) ErrorResponse(w http.ResponseWriter, _ *http.Request, error string, code int) {
	data := struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}{
		Code:    code,
		Message: error,
	}

	body, err := json.Marshal(data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		s.logger.Error("JSON marshal failed", slog.Any("error", err))
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	_, _ = w.Write(prettyJSON(body))
}
