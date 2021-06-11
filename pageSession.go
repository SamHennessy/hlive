package hlive

import (
	"context"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rs/xid"
)

// TODO:
// 		test limit wait,
// 		test garbage collection,
// 		does the garbage collection exit?

func NewPageSessionStorage() *PageSessionStorage {
	pss := &PageSessionStorage{
		sessions:          map[string]*PageSession{},
		DisconnectTimeout: time.Second * 5,
		SessionLimit:      1000,
	}

	go pss.GarbageCollection()

	return pss
}

type PageSessionStorage struct {
	sessions          map[string]*PageSession
	lock              sync.RWMutex
	DisconnectTimeout time.Duration
	SessionLimit      int
	sessionCount      uint32
}

func (pss *PageSessionStorage) mapAdd(ps *PageSession) {
	pss.lock.Lock()
	atomic.StoreUint32(&pss.sessionCount, pss.sessionCount+1)
	pss.sessions[ps.ID] = ps

	pss.lock.Unlock()
}

func (pss *PageSessionStorage) mapGet(id string) *PageSession {
	pss.lock.RLock()

	ps := pss.sessions[id]

	pss.lock.RUnlock()

	return ps
}

func (pss *PageSessionStorage) mapDelete(id string) {
	pss.lock.Lock()

	if _, exists := pss.sessions[id]; exists {
		delete(pss.sessions, id)
		atomic.StoreUint32(&pss.sessionCount, pss.sessionCount-1)
	}

	pss.lock.Unlock()
}

func (pss *PageSessionStorage) GarbageCollection() {
	for {
		if pss == nil {
			return
		}

		time.Sleep(time.Second)
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

func (pss *PageSessionStorage) New() *PageSession {
	// Block
	pss.newWait()

	ps := &PageSession{ID: xid.New().String()}
	pss.mapAdd(ps)

	return ps
}

func (pss *PageSessionStorage) newWait() {
	for {
		if pss.sessionCount < uint32(pss.SessionLimit) {
			return
		}
	}
}

func (pss *PageSessionStorage) Get(id string) *PageSession {
	return pss.mapGet(id)
}

func (pss *PageSessionStorage) Delete(id string) {
	ps := pss.mapGet(id)
	if ps == nil {
		return
	}

	if ps.Page != nil {
		ps.Page.Close(context.Background())
	}

	pss.mapDelete(id)
}

func (pss *PageSessionStorage) GetSessionCount() int {
	return int(pss.sessionCount)
}

func NewPageServer(pf func() *Page) *PageServer {
	s := &PageServer{
		pageFunc: pf,
		Sessions: NewPageSessionStorage(),
	}

	return s
}

type PageServer struct {
	pageFunc func() *Page
	Sessions *PageSessionStorage
}

type PageSession struct {
	ID         string
	LastActive time.Time
	Page       *Page
}

func (s *PageServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// WebSocket?
	if sessID := r.URL.Query().Get("ws"); sessID != "" {
		var sess *PageSession
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
