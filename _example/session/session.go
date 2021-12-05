package main

import (
	"context"
	"errors"
	"log"
	"net/http"

	l "github.com/SamHennessy/hlive"
	"github.com/rs/xid"
)

func main() {
	s := newService()

	http.HandleFunc("/", sessionMiddleware(home(s).ServeHTTP))

	log.Println("INFO: listing :3000")

	if err := http.ListenAndServe(":3000", nil); err != nil {
		log.Println("ERRO: http listen and serve: ", err)
	}
}

func home(s *service) *l.PageServer {
	f := func() *l.Page {
		page := l.NewPage()
		page.DOM.Title.Add("HTTP Session Example")
		page.DOM.Head.Add(l.T("link", l.Attrs{"rel": "stylesheet", "href": "https://cdn.simplecss.org/simple.min.css"}))

		page.DOM.Body.Add(
			l.T("header",
				l.T("h1", "HTTP Session"),
				l.T("p", "You can use middleware to implement a persistent session."),
			),
			l.T("main",
				l.T("p", "Enter a message, then open another tab to see it there."),
				l.T("h2", "Your Message"),
				newMessage(s),
				l.T("p", "This example uses a cookie and server memory to persist between page reloads but not server reloads. Changes are not synced between tabs in real-time."),
				l.T("p", "Be careful when testing in Firefox, as it will keep the current form value on refresh."),
			),
		)

		return page
	}

	return l.NewPageServer(f)
}

type ctxKey string

const sessionKey ctxKey = "session"

func sessionMiddleware(h http.HandlerFunc) http.HandlerFunc {
	cookieName := "hlive_session"

	return func(w http.ResponseWriter, r *http.Request) {
		var sessionID string

		cook, err := r.Cookie(cookieName)

		switch {
		case errors.Is(err, http.ErrNoCookie):
			sessionID = xid.New().String()

			http.SetCookie(w,
				&http.Cookie{Name: cookieName, Value: sessionID, Path: "/", SameSite: http.SameSiteStrictMode})
		case err != nil:
			log.Println("ERROR: get cookie: ", err.Error())
		default:
			sessionID = cook.Value
		}

		r = r.WithContext(context.WithValue(r.Context(), sessionKey, sessionID))

		h(w, r)
	}
}

func getSessionID(ctx context.Context) string {
	val, _ := ctx.Value(sessionKey).(string)

	return val
}

func newService() *service {
	return &service{userMessage: map[string]string{}}
}

type service struct {
	userMessage map[string]string
}

func (s *service) SetMessage(userID, message string) {
	s.userMessage[userID] = message
}

func (s *service) GetMessage(userID string) string {
	return s.userMessage[userID]
}

func newMessage(service *service) *message {
	c := &message{
		Component: l.C("textarea"),
		service:   service,
	}

	c.Add(l.On("input", func(ctx context.Context, e l.Event) {
		c.service.SetMessage(getSessionID(ctx), e.Value)
	}))

	return c
}

type message struct {
	*l.Component

	Message string

	service *service
}

func (c *message) Mount(ctx context.Context) {
	c.Message = c.service.GetMessage(getSessionID(ctx))
}

func (c *message) GetNodes() *l.NodeGroup {
	return l.Group(c.Message)
}
