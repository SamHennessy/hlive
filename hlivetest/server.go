package hlivetest

import (
	"net/http/httptest"

	l "github.com/SamHennessy/hlive"
)

type Server struct {
	PageServer       *l.PageServer
	PageSessionStore *l.PageSessionStore
	HTTPServer       *httptest.Server
}

func NewServer(pageFn func() *l.Page) *Server {
	s := &Server{}
	s.PageSessionStore = l.NewPageSessionStore()
	s.PageServer = l.NewPageServerWithSessionStore(addAck(pageFn), s.PageSessionStore)
	s.HTTPServer = httptest.NewServer(s.PageServer)

	return s
}

func addAck(pageFn func() *l.Page) func() *l.Page {
	return func() *l.Page {
		p := pageFn()

		p.DOM().HTML().Add(Ack())

		return p
	}
}
