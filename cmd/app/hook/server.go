package hook

import (
	"net/http"
	"sync"
)

type Server struct {
	c  http.Client
	wg sync.WaitGroup
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// eventType, eventGUID, payload, ok, resp := github.validate
}
