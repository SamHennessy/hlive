package hlivetest

import (
	"net/http/httptest"

	l "github.com/SamHennessy/hlive"
)

type Server struct {
	PageServer *l.PageServer
	HTTPServer *httptest.Server
}

func NewServer(pageFn func() *l.Page) *Server {
	s := &Server{}
	s.PageServer = l.NewPageServer(addAck(pageFn))
	s.HTTPServer = httptest.NewServer(s.PageServer)

	return s
}

func addAck(pageFn func() *l.Page) func() *l.Page {
	return func() *l.Page {
		p := pageFn()

		p.HTML.Add(Ack())

		return p
	}
}
