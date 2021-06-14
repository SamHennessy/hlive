package hlive

import (
	"context"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rs/xid"
)

type PageSession struct {
	ID         string
	LastActive time.Time
	Page       *Page
}

// TODO:
// 		test limit wait,
// 		test garbage collection,
// 		does the garbage collection exit?

func NewPageSessionStore() *PageSessionStore {
	pss := &PageSessionStore{
		sessions:              map[string]*PageSession{},
		DisconnectTimeout:     WebSocketDisconnectTimeoutDefault,
		SessionLimit:          PageSessionLimitDefault,
		GarbageCollectionTick: PageSessionGarbageCollectionTick,
	}

	go pss.GarbageCollection()

	return pss
}

type PageSessionStore struct {
	sessions              map[string]*PageSession
	lock                  sync.RWMutex
	DisconnectTimeout     time.Duration
	SessionLimit          int
	sessionCount          uint32
	GarbageCollectionTick time.Duration
}

// New PageSession
func (pss *PageSessionStore) New() *PageSession {
	// Block
	pss.newWait()

	ps := &PageSession{ID: xid.New().String()}
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
	atomic.StoreUint32(&pss.sessionCount, pss.sessionCount+1)
	pss.sessions[ps.ID] = ps

	pss.lock.Unlock()
}

func (pss *PageSessionStore) mapGet(id string) *PageSession {
	pss.lock.RLock()

	ps := pss.sessions[id]

	pss.lock.RUnlock()

	return ps
}

func (pss *PageSessionStore) mapDelete(id string) {
	pss.lock.Lock()

	if _, exists := pss.sessions[id]; exists {
		delete(pss.sessions, id)
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

		for id, sess := range pss.sessions {
			if sess.Page == nil {
				pss.mapDelete(id)

				continue
			}

			if sess.Page.IsConnected() {
				continue
			}
			// Allow to stay until the exceed the timeout
			if now.Sub(sess.LastActive) > pss.DisconnectTimeout {
				sess.Page.Close(context.Background())
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
	s := &PageServer{
		pageFunc: pf,
		Sessions: NewPageSessionStore(),
	}

	return s
}

type PageServer struct {
	pageFunc func() *Page
	Sessions *PageSessionStore
}

func (s *PageServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// WebSocket?
	if sessID := r.URL.Query().Get("ws"); sessID != "" {
		var sess *PageSession
		// TODO: need to rethink reconnect
		if sessID != "1" {
			sess = s.Sessions.Get(sessID)
		}
		// New or not found
		if sess == nil {
			sess = s.Sessions.New()
			sess.Page = s.pageFunc()
			sess.LastActive = time.Now()
		}

		sess.Page.ServerWS(w, r.WithContext(context.WithValue(r.Context(), CtxPageSess, sess)))

		return
	}

	s.pageFunc().ServeHTTP(w, r)
}

func PageSessID(ctx context.Context) string {
	v := PageSess(ctx)

	if v == nil {
		return ""
	}

	return v.ID
}

func PageSess(ctx context.Context) *PageSession {
	v, _ := ctx.Value(CtxPageSess).(*PageSession)

	return v
}
