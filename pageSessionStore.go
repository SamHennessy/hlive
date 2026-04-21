package hlive

import (
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rs/xid"
	"github.com/rs/zerolog"
)

func NewPageSessionStore() *PageSessionStore {
	pss := &PageSessionStore{
		DisconnectTimeout:     WebSocketDisconnectTimeoutDefault,
		SessionLimit:          PageSessionLimitDefault,
		GarbageCollectionTick: PageSessionGarbageCollectionTick,
		Done:                  make(chan bool),
		sessions:              &sync.Map{},
	}

	go pss.GarbageCollection()

	return pss
}

type PageSessionStore struct {
	sessions              *sync.Map
	DisconnectTimeout     time.Duration
	SessionLimit          uint32
	sessionCount          uint32
	GarbageCollectionTick time.Duration
	Done                  chan bool
}

// New PageSession.
func (pss *PageSessionStore) New() *PageSession {
	// Block until we have room for a new session
	pss.newWait()

	ps := &PageSession{
		id:      xid.New().String(),
		logger:  zerolog.Nop(),
		Send:    make(chan MessageWS),
		Receive: make(chan MessageWS),
		done:    make(chan bool),
	}

	pss.mapAdd(ps)

	return ps
}

// TODO: use sync.Cond
func (pss *PageSessionStore) newWait() {
	for atomic.LoadUint32(&pss.sessionCount) > pss.SessionLimit {
		runtime.Gosched()
	}
}

func (pss *PageSessionStore) Get(id string) *PageSession {
	return pss.mapGet(id)
}

func (pss *PageSessionStore) mapAdd(ps *PageSession) {
	pss.sessions.Store(ps.id, ps)
	atomic.AddUint32(&pss.sessionCount, 1)
}

func (pss *PageSessionStore) mapGet(id string) *PageSession {
	if ps, ok := pss.sessions.Load(id); ok {
		return ps.(*PageSession)
	}
	return nil
}

func (pss *PageSessionStore) mapDelete(id string) {
	if _, loaded := pss.sessions.LoadAndDelete(id); loaded {
		atomic.AddUint32(&pss.sessionCount, ^uint32(0))
	}
}

func (pss *PageSessionStore) GarbageCollection() {
	for {
		time.Sleep(pss.GarbageCollectionTick)

		select {
		case <-pss.Done:
			return
		default:
			now := time.Now()
			pss.sessions.Range(func(key, value any) bool {
				id := key.(string)
				sess := value.(*PageSession)

				if sess.IsConnected() {
					return true
				}

				// Keep until it exceeds the timeout
				sess.muSess.RLock()
				la := sess.lastActive
				sess.muSess.RUnlock()

				if now.Sub(la) > pss.DisconnectTimeout {
					if sess.page != nil {
						sess.page.Close(sess.ctxPage)
					}

					if sess.ctxInitialCancel != nil {
						sess.ctxInitialCancel()
					}

					close(sess.done)
					pss.mapDelete(id)
				}

				return true
			})
		}
	}
}

func (pss *PageSessionStore) Delete(id string) {
	ps := pss.mapGet(id)
	if ps == nil {
		return
	}

	if ps.GetPage() != nil {
		ps.GetPage().Close(ps.GetContextPage())
	}

	pss.mapDelete(id)
}

func (pss *PageSessionStore) GetSessionCount() int {
	return int(atomic.LoadUint32(&pss.sessionCount))
}
