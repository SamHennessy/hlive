package hlive

import (
	"context"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cornelk/hashmap"
	"github.com/gorilla/websocket"
	"github.com/rs/xid"
	"github.com/rs/zerolog"
	"github.com/vmihailenco/msgpack/v5"
)

type PageSession struct {
	ID             string
	ConnectedAt    time.Time
	LastActive     time.Time
	Page           *Page
	InitialContext context.Context //nolint:containedctx // we are a router
	Done           chan bool

	connected bool
	weConn    *websocket.Conn
	// Buffered channel of outbound messages.
	Send chan MessageWS
	// Buffered channel of inbound messages.
	Receive chan MessageWS

	logger zerolog.Logger
}

// TODO:
// 		test limit wait,
// 		test garbage collection,
// 		does the garbage collection exit?

func NewPageSessionStore() *PageSessionStore {
	pss := &PageSessionStore{
		DisconnectTimeout:     WebSocketDisconnectTimeoutDefault,
		SessionLimit:          PageSessionLimitDefault,
		GarbageCollectionTick: PageSessionGarbageCollectionTick,
	}

	go pss.GarbageCollection()

	return pss
}

type PageSessionStore struct {
	sessions              hashmap.HashMap
	lock                  sync.RWMutex
	DisconnectTimeout     time.Duration
	SessionLimit          int
	sessionCount          uint32
	GarbageCollectionTick time.Duration
}

// New PageSession.
func (pss *PageSessionStore) New() *PageSession {
	// Block
	pss.newWait()

	ps := &PageSession{
		ID:     xid.New().String(),
		logger: zerolog.Nop(),
		// TODO: Do we need buffer?
		Send:    make(chan MessageWS, 256),
		Receive: make(chan MessageWS),
		Done:    make(chan bool),
	}
	pss.mapAdd(ps)

	return ps
}

func (pss *PageSessionStore) newWait() {
	for {
		if pss.sessionCount < uint32(pss.SessionLimit) {
			return
		}
	}
}

func (pss *PageSessionStore) Get(id string) *PageSession {
	return pss.mapGet(id)
}

func (pss *PageSessionStore) mapAdd(ps *PageSession) {
	pss.lock.Lock()
	pss.sessions.Set(ps.ID, ps)
	// TODO: maybe do swap to avoid lock?
	atomic.StoreUint32(&pss.sessionCount, pss.sessionCount+1)

	pss.lock.Unlock()
}

func (pss *PageSessionStore) mapGet(id string) *PageSession {
	val, _ := pss.sessions.GetStringKey(id)
	ps, _ := val.(*PageSession)

	return ps
}

func (pss *PageSessionStore) mapDelete(id string) {
	pss.lock.Lock()

	if _, exists := pss.sessions.GetStringKey(id); exists {
		pss.sessions.Del(id)
		atomic.StoreUint32(&pss.sessionCount, pss.sessionCount-1)
	}

	pss.lock.Unlock()
}

func (pss *PageSessionStore) GarbageCollection() {
	for {
		if pss == nil {
			return
		}

		time.Sleep(pss.GarbageCollectionTick)
		now := time.Now()

		for keyVal := range pss.sessions.Iter() {
			id, _ := keyVal.Key.(string)
			sess, _ := keyVal.Value.(*PageSession)

			if sess.Page == nil {
				pss.mapDelete(id)

				continue
			}

			if sess.Page.IsConnected() {
				continue
			}
			// Keep until it exceeds the timeout
			if now.Sub(sess.LastActive) > pss.DisconnectTimeout {
				sess.Page.Close(sess.InitialContext)
				close(sess.Done)
				pss.mapDelete(id)
			}
		}
	}
}

func (pss *PageSessionStore) Delete(id string) {
	ps := pss.mapGet(id)
	if ps == nil {
		return
	}

	if ps.Page != nil {
		ps.Page.Close(context.Background())
	}

	pss.mapDelete(id)
}

func (pss *PageSessionStore) GetSessionCount() int {
	return int(pss.sessionCount)
}

func NewPageServer(pf func() *Page) *PageServer {
	return NewPageServerWithSessionStore(pf, NewPageSessionStore())
}

func NewPageServerWithSessionStore(pf func() *Page, sess *PageSessionStore) *PageServer {
	s := &PageServer{
		pageFunc: pf,
		Sessions: sess,
		logger:   zerolog.Nop(),
	}

	return s
}

type MessageWS struct {
	Message  []byte
	IsBinary bool
}

type PageServer struct {
	Sessions *PageSessionStore
	Upgrader websocket.Upgrader

	pageFunc func() *Page
	logger   zerolog.Logger
}

func (s *PageServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// WebSocket?
	if sessID := r.URL.Query().Get("hlive"); sessID != "" {
		var sess *PageSession

		if sessID != "1" {
			sess = s.Sessions.Get(sessID)

			if sess != nil && sess.Page.connected {
				sess.Page.logger.Warn().Str("id", sessID).
					Msg("connection blocked as an active connection exists")

				w.WriteHeader(http.StatusNotFound)

				return
			}
		}

		// New or not found
		if sessID == "1" {
			sess = s.Sessions.New()
			sess.Page = s.pageFunc()
			sess.ConnectedAt = time.Now()
			sess.LastActive = sess.ConnectedAt
			sess.InitialContext = r.Context()
		}

		if sess == nil {
			log.Println("HLive: WARN: session not found:", sessID)
			w.WriteHeader(http.StatusNotFound)

			return
		}

		hhash := r.URL.Query().Get("hhash")

		s.logger = sess.Page.logger
		s.logger.Debug().Str("sessionID", sessID).Str("hash", hhash).Msg("ws start")

		if sess.Page.cache != nil && hhash != "" {
			val, hit := sess.Page.cache.Get(hhash)

			b, ok := val.([]byte)
			if hit && ok {
				s.logger.Debug().Bool("hit", hit).Str("hhash", hhash).Int("size", len(b)/1024).
					Msg("cache get")
				newTree := G()
				if err := msgpack.Unmarshal(b, newTree); err != nil {
					s.logger.Err(err).Msg("ServeHTTP: msgpack.Unmarshal")
				} else {
					sess.Page.DOMBrowser = newTree
				}
			}
		}

		var err error

		sess.weConn, err = s.Upgrader.Upgrade(w, r, nil)
		if err != nil {
			s.logger.Err(err).Msg("ws upgrade")
			w.WriteHeader(http.StatusInternalServerError)

			return
		}

		sess.LastActive = time.Now()

		go sess.writePump()
		go sess.readPump()

		if err := sess.Page.ServeWS(sess.InitialContext, sess.ID, sess.Send, sess.Receive); err != nil {
			sess.Page.logger.Err(err).Msg("ws serve")
			w.WriteHeader(http.StatusInternalServerError)
		}

		// This needs to say open to keep the context active
		<-sess.Done

		return
	}

	s.pageFunc().ServeHTTP(w, r)
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (sess *PageSession) readPump() {
	defer func() {
		sess.connected = false
		sess.Page.connected = false

		if err := sess.weConn.Close(); err != nil {
			sess.logger.Err(err).Msg("ws conn close")
		} else {
			sess.logger.Trace().Msg("ws close")
		}
	}()

	// c.conn.SetReadLimit(maxMessageSize)
	if err := sess.weConn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		sess.logger.Err(err).Msg("read pump set read deadline")
	}

	sess.weConn.SetPongHandler(func(string) error {
		sess.logger.Trace().Msg("ws pong")

		if err := sess.weConn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
			sess.logger.Err(err).Msg("pong handler: set read deadline")
		}

		return nil
	})

	for {
		mt, message, err := sess.weConn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				sess.logger.Debug().Err(err).Msg("unexpected close error")
			}

			break
		}

		sess.LastActive = time.Now()

		sess.Receive <- MessageWS{
			Message:  message,
			IsBinary: mt == websocket.BinaryMessage,
		}
	}
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.

/*
TODO:

panic: concurrent write to websocket connection

goroutine 73 [running]:
github.com/gorilla/websocket.(*messageWriter).flushFrame(0x1400036f200, 0x1, {0x0?, 0x12bab4d28?, 0x104b2c5b8?})
	/Users/sam/go/pkg/mod/github.com/gorilla/websocket@v1.5.0/conn.go:617 +0x470
github.com/gorilla/websocket.(*messageWriter).Close(0x0?)
	/Users/sam/go/pkg/mod/github.com/gorilla/websocket@v1.5.0/conn.go:731 +0x4c
github.com/gorilla/websocket.(*Conn).beginMessage(0x1400021a580, 0x1400049eb10, 0x1)
	/Users/sam/go/pkg/mod/github.com/gorilla/websocket@v1.5.0/conn.go:480 +0x40
github.com/gorilla/websocket.(*Conn).NextWriter(0x1400021a580, 0x1)
	/Users/sam/go/pkg/mod/github.com/gorilla/websocket@v1.5.0/conn.go:520 +0x44
github.com/SamHennessy/hlive.(*PageSession).writePump(0x140002d5ad0)
	/Users/sam/go/pkg/mod/github.com/!sam!hennessy/hlive@v0.0.0-20220301072939-9601d863612f/pageSession.go:361 +0x16c
created by github.com/SamHennessy/hlive.(*PageServer).ServeHTTP
	/Users/sam/go/pkg/mod/github.com/!sam!hennessy/hlive@v0.0.0-20220301072939-9601d863612f/pageSession.go:238 +0x3cc

*/

func (sess *PageSession) writePump() {
	defer func() {
		err := recover()
		if err != nil {
			sess.logger.Error().Interface("returned", err).Msg("write pump panic recover")
		}
	}()

	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()

		if err := sess.weConn.Close(); err != nil {
			sess.logger.Trace().Err(err).Msg("write pump: close ws connection")
		}
	}()

	for {
		select {
		case message, ok := <-sess.Send:
			if err := sess.weConn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				sess.logger.Err(err).Msg("write pump: message set write deadline")
			}

			if !ok {
				// Send channel closed.
				if err := sess.weConn.WriteMessage(websocket.CloseMessage, []byte{}); err != nil {
					sess.logger.Err(err).Msg("write pump: write close message")
				}

				return
			}

			mt := websocket.TextMessage
			if message.IsBinary {
				mt = websocket.BinaryMessage
			}

			w, err := sess.weConn.NextWriter(mt)
			if err != nil {
				sess.logger.Err(err).Msg("write pump: create writer")

				return
			}

			if _, err := w.Write(message.Message); err != nil {
				sess.logger.Err(err).Msg("write pump: write first message")
			}

			// TODO: is this worth it? Do I even want to buffer?
			// Add queued messages to the current websocket message.
			// n := len(p.Send)
			// for i := 0; i < n; i++ {
			// 	if _, err := w.Write(newline); err != nil {
			// 		p.logger.Err(err).Msg("write pump: write queued message")
			// 	}
			//
			// 	if _, err := w.Write(<-p.Send); err != nil {
			// 		p.logger.Err(err).Msg("write pump: write message delimiter")
			// 	}
			// }

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			sess.logger.Trace().Msg("ws ping")

			if err := sess.weConn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				sess.logger.Err(err).Msg("write pump: ping tick: set write deadline")
			}

			if err := sess.weConn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
