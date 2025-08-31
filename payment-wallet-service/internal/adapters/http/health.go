package http

import (
	"net/http"
	"sync/atomic"
)

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	if atomic.LoadInt32(&_healthy) == 1 {
		s.JSONResponse(w, r, map[string]string{"status": "ok"})
		return
	}
	w.WriteHeader(http.StatusServiceUnavailable)
}
