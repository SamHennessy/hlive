package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	l "github.com/SamHennessy/hlive"
	"github.com/rs/xid"
	"github.com/rs/zerolog"
)

func main() {
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).With().Timestamp().Logger().Level(zerolog.InfoLevel)
	s := NewService()

	http.HandleFunc("/", sessionMiddleware(Home(logger, s).ServeHTTP))

	logger.Info().Str("addr", ":3000").Msg("listing")

	if err := http.ListenAndServe(":3000", nil); err != nil {
		logger.Err(err).Msg("http listen and serve")
	}
}

func Home(logger zerolog.Logger, s *Service) *l.PageServer {
	f := func() *l.Page {
		page := l.NewPage()
		page.Title.Add("HTTP Session Example")
		page.SetLogger(logger)
		page.Body.Add(newMessage(s))

		return page
	}

	return l.NewPageServer(f)
}

type CtxKey string

const SessionKey CtxKey = "session"

func sessionMiddleware(h http.HandlerFunc) http.HandlerFunc {
	cookieName := "hlive_session"

	return func(w http.ResponseWriter, r *http.Request) {
		var sessionID string

		cook, err := r.Cookie(cookieName)
		switch {
		case err == http.ErrNoCookie:
			sessionID = xid.New().String()
			http.SetCookie(w,
				&http.Cookie{Name: cookieName, Value: sessionID, Path: "/", SameSite: http.SameSiteStrictMode})
		case err != nil:
			fmt.Println("ERROR: get cookie: ", err.Error())
		default:
			sessionID = cook.Value
		}

		r = r.WithContext(context.WithValue(r.Context(), SessionKey, sessionID))

		h(w, r)
	}
}

func GetSessionID(ctx context.Context) string {
	val, _ := ctx.Value(SessionKey).(string)

	return val
}

func NewService() *Service {
	return &Service{userMessage: map[string]string{}}
}

type Service struct {
	userMessage map[string]string
}

func (s *Service) SetMessage(userID, message string) {
	s.userMessage[userID] = message
}

func (s *Service) GetMessage(userID string) string {
	return s.userMessage[userID]
}

func newMessage(service *Service) *message {
	c := &message{
		Component: l.NewComponent("span"),
		service:   service,
	}

	c.On(l.OnKeyUp(func(ctx context.Context, e l.Event) {
		c.service.SetMessage(GetSessionID(ctx), e.Value)
	}))

	return c
}

type message struct {
	*l.Component

	Message string

	service *Service
}

func (c *message) Mount(ctx context.Context) {
	c.Message = c.service.GetMessage(GetSessionID(ctx))
}

func (c *message) GetNodes() []interface{} {
	c.SetName("textarea")

	return l.Tree(c.Message)
}
