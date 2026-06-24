package hlive

import (
	"context"
	"net/http"
	"time"

	"log/slog"

	"github.com/gorilla/websocket"
	"github.com/vmihailenco/msgpack/v5"
)

func NewPageServer(pf func() *Page) *PageServer {
	return NewPageServerWithSessionStore(pf, NewPageSessionStore())
}

func NewPageServerWithSessionStore(pf func() *Page, sess *PageSessionStore) *PageServer {
	return &PageServer{
		pageFunc: pf,
		Sessions: sess,
		logger:   slog.New(slog.DiscardHandler),
	}
}

type PageServer struct {
	Sessions *PageSessionStore
	Upgrader websocket.Upgrader

	pageFunc func() *Page
	logger   *slog.Logger
}

func (s *PageServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// WebSocket?
	sessID := r.URL.Query().Get("hlive")

	if sessID == "" {
		s.pageFunc().ServeHTTP(w, r)

		return
	}

	var sess *PageSession
	// New
	if sessID == "1" {
		sess = s.Sessions.New()
		sess.muSess.Lock()
		sess.page = s.pageFunc()
		sess.connectedAt = time.Now()
		sess.lastActive = sess.connectedAt
		sess.ctxInitial, sess.ctxInitialCancel = context.WithCancel(r.Context())
		sess.ctxPage, sess.ctxPageCancel = context.WithCancel(sess.ctxInitial)
		sess.muSess.Unlock()
	} else { // Reconnect
		// TODO: need to rethink reconnect and double check my assumptions
		//sess = s.Sessions.Get(sessID)
		//
		//if sess != nil && sess.IsConnected() {
		//	LoggerDev.Error().Str("sessID", sessID).
		//		Msg("ws connect: is connected: connection blocked as an active connection exists")
		//
		//	w.WriteHeader(http.StatusNotFound)
		//
		//	return
		//}

		//if sess != nil {
		//
		//	//sess.GetInitialContextCancel()()
		//
		//	//sess.muSess.Lock()
		//
		//	//sess.ctxInitial, sess.ctxInitialCancel = context.WithCancel(r.Context())
		//	//sess.ctxPage, sess.ctxPageCancel = context.WithCancel(sess.ctxInitial)
		//
		//	//sess.muSess.Unlock()
		//}
	}

	// LoggerDev.Debug().Str("sessID", sessID).Bool("success", sess != nil).Msg("ws connect")

	if sess == nil {
		w.WriteHeader(http.StatusNotFound)

		return
	}

	hhash := r.URL.Query().Get("hhash")

	s.logger = sess.GetPage().logger
	s.logger.Debug("ws start", "sessionID", sessID, "hash", hhash)

	if sess.GetPage().cache != nil && hhash != "" && sessID == "1" {
		val, hit := sess.GetPage().cache.Get(hhash)

		b, ok := val.([]byte)
		if hit && ok {
			s.logger.Debug("cache get", "hit", hit, "hhash", hhash, "size", len(b)/1024)
			newTree := G()
			if err := msgpack.Unmarshal(b, newTree); err != nil {
				s.logger.Error("ServeHTTP: msgpack.Unmarshal", "error", err)
			} else {
				sess.GetPage().domBrowser = newTree
			}
		}
	}

	sess.muSess.Lock()

	var err error
	sess.wsConn, err = s.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		sess.muSess.Unlock()
		s.logger.Error("ws upgrade", "error", err)
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	sess.connected = true
	sess.lastActive = time.Now()

	sess.muSess.Unlock()

	go sess.writePump()
	go sess.readPump()

	if err := sess.GetPage().ServeWS(sess.GetContextPage(), sess.GetID(), sess.Send, sess.Receive); err != nil {
		sess.GetPage().logger.Error("ws serve", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
	}

	// This needs to say open to keep the context active
	<-sess.done
}
