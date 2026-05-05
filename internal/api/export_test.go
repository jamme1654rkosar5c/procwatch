// export_test.go exposes internal helpers for white-box testing.
package api

import "net/http"

// ServeHTTPForTest allows tests to call the mux directly without
// starting a real TCP listener.
func (s *Server) ServeHTTPForTest(w http.ResponseWriter, r *http.Request) {
	s.httpServer.Handler.ServeHTTP(w, r)
}
