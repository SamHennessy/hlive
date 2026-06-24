package hlive

import (
	"context"
	"sync"
	"time"

	"log/slog"

	"github.com/gorilla/websocket"
)

type PageSession struct {
	// Buffered channel of outbound messages.
	Send chan MessageWS
	// Buffered channel of inbound messages.
	Receive chan MessageWS

	id               string
	connected        bool
	connectedAt      time.Time
	lastActive       time.Time
	page             *Page
	ctxInitial       context.Context //nolint:containedctx // we are a router and create new contexts from this one
	ctxPage          context.Context //nolint:containedctx // we are a router and create new contexts from this one
	ctxPageCancel    context.CancelFunc
	ctxInitialCancel context.CancelFunc
	done             chan bool
	wsConn           *websocket.Conn
	logger           *slog.Logger
	muSess           sync.RWMutex
	muWrite          sync.RWMutex
	muRead           sync.RWMutex
}

type MessageWS struct {
	Message  []byte
	IsBinary bool
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (sess *PageSession) readPump() {
	defer func() {
		sess.muSess.Lock()
		sess.connected = false
		sess.muSess.Unlock()

		sess.muWrite.Lock()
		if err := sess.wsConn.Close(); err != nil {
			sess.logger.Error("ws conn close", "error", err, "sess", sess.id)
		} else {
			sess.logger.Debug("ws conn close", "sess", sess.id)
		}
		sess.muWrite.Unlock()
	}()

	sess.muWrite.Lock()

	// c.conn.SetReadLimit(maxMessageSize)
	if err := sess.wsConn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		sess.logger.Error("read pump set read deadline", "error", err)
	}

	sess.wsConn.SetPongHandler(func(string) error {
		sess.logger.Log(context.Background(), LevelTrace, "ws pong")

		sess.muWrite.Lock()

		if err := sess.wsConn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
			sess.logger.Error("pong handler: set read deadline", "error", err)
		}

		sess.muWrite.Unlock()

		return nil
	})

	sess.muWrite.Unlock()

	for {
		select {
		case <-sess.GetContextInitial().Done():
			return
		default:
			sess.muRead.Lock()

			mt, message, err := sess.wsConn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					sess.logger.Debug("unexpected close error", "error", err)
				}

				return
			}

			sess.muRead.Unlock()

			sess.muSess.Lock()
			sess.lastActive = time.Now()
			sess.muSess.Unlock()

			sess.Receive <- MessageWS{
				Message:  message,
				IsBinary: mt == websocket.BinaryMessage,
			}
		}
	}
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (sess *PageSession) writePump() {
	ticker := time.NewTicker(pingPeriod)

	for {
		select {
		case <-sess.GetContextInitial().Done():
			ticker.Stop()

			return
		case message, ok := <-sess.Send:
			sess.muWrite.Lock()

			if err := sess.wsConn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				sess.logger.Error("write pump: message set write deadline", "error", err)
			}

			sess.muWrite.Unlock()

			if !ok {
				// Send channel closed.
				sess.muWrite.Lock()

				if err := sess.wsConn.WriteMessage(websocket.CloseMessage, []byte{}); err != nil {
					sess.logger.Error("write pump: write close message", "error", err)
				}

				sess.muWrite.Unlock()

				return
			}

			mt := websocket.TextMessage
			if message.IsBinary {
				mt = websocket.BinaryMessage
			}

			sess.muWrite.Lock()

			w, err := sess.wsConn.NextWriter(mt)
			if err != nil {
				sess.logger.Error("write pump: create writer", "error", err)

				sess.muWrite.Unlock()

				continue
			}

			if _, err := w.Write(message.Message); err != nil {
				sess.logger.Error("write pump: write first message", "error", err)
			}

			if err := w.Close(); err != nil {
				sess.logger.Error("write pump: close write", "error", err)
			}

			sess.muWrite.Unlock()

		case <-ticker.C:
			sess.logger.Log(context.Background(), LevelTrace, "ws ping")

			sess.muWrite.Lock()

			if err := sess.wsConn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				sess.logger.Error("write pump: ping tick: set write deadline", "error", err)
			}

			if err := sess.wsConn.WriteMessage(websocket.PingMessage, nil); err != nil {
				sess.logger.Error("write pump: ping tick: write write message", "error", err)
			}

			sess.muWrite.Unlock()
		}
	}
}

func (sess *PageSession) GetPage() *Page {
	sess.muSess.RLock()
	defer sess.muSess.RUnlock()

	return sess.page
}

func (sess *PageSession) SetPage(page *Page) {
	sess.muSess.Lock()
	sess.page = page
	sess.muSess.Unlock()
}

func (sess *PageSession) GetID() string {
	sess.muSess.RLock()
	defer sess.muSess.RUnlock()

	return sess.id
}

func (sess *PageSession) GetContextInitial() context.Context {
	sess.muSess.RLock()
	defer sess.muSess.RUnlock()

	return sess.ctxInitial
}

func (sess *PageSession) GetContextPage() context.Context {
	sess.muSess.RLock()
	defer sess.muSess.RUnlock()

	return sess.ctxPage
}

func (sess *PageSession) GetPageContextCancel() context.CancelFunc {
	sess.muSess.RLock()
	defer sess.muSess.RUnlock()

	return sess.ctxPageCancel
}

func (sess *PageSession) GetInitialContextCancel() context.CancelFunc {
	sess.muSess.RLock()
	defer sess.muSess.RUnlock()

	return sess.ctxInitialCancel
}

func (sess *PageSession) SetContextPage(ctx context.Context) {
	sess.muSess.Lock()
	sess.ctxPage = ctx
	sess.muSess.Unlock()
}

func (sess *PageSession) SetContextCancel(cancel context.CancelFunc) {
	sess.muSess.Lock()
	sess.ctxPageCancel = cancel
	sess.muSess.Unlock()
}

func (sess *PageSession) IsConnected() bool {
	sess.muSess.RLock()
	defer sess.muSess.RUnlock()

	return sess.connected
}

func (sess *PageSession) SetConnected(connected bool) {
	sess.muSess.Lock()
	sess.connected = connected
	sess.muSess.Unlock()
}
