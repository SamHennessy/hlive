package hlive

import (
	"sync/atomic"
	"time"

	"github.com/cornelk/hashmap"
	"github.com/rs/xid"
	"github.com/rs/zerolog"
)

func NewPageSessionStore() *PageSessionStore {
	pss := &PageSessionStore{
		DisconnectTimeout:     WebSocketDisconnectTimeoutDefault,
		SessionLimit:          PageSessionLimitDefault,
		GarbageCollectionTick: PageSessionGarbageCollectionTick,
		Done:                  make(chan bool),
	}

	go pss.GarbageCollection()

	return pss
}

type PageSessionStore struct {
	sessions              hashmap.HashMap
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

func (pss *PageSessionStore) newWait() {
	for {
		if atomic.LoadUint32(&pss.sessionCount) < pss.SessionLimit {
			return
		}
	}
}

func (pss *PageSessionStore) Get(id string) *PageSession {
	return pss.mapGet(id)
}

func (pss *PageSessionStore) mapAdd(ps *PageSession) {
	pss.sessions.Set(ps.id, ps)
	atomic.AddUint32(&pss.sessionCount, 1)
}

func (pss *PageSessionStore) mapGet(id string) *PageSession {
	val, _ := pss.sessions.GetStringKey(id)
	ps, _ := val.(*PageSession)

	return ps
}

func (pss *PageSessionStore) mapDelete(id string) {
	if _, exists := pss.sessions.GetStringKey(id); exists {
		pss.sessions.Del(id)
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
			for keyVal := range pss.sessions.Iter() {
				id, _ := keyVal.Key.(string)
				sess, _ := keyVal.Value.(*PageSession)

				if sess.IsConnected() {
					continue
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
			}
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
