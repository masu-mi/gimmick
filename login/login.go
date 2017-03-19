package login

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/satori/go.uuid"
	"golang.org/x/oauth2"
)

type (
	session struct {
		startAt time.Time
		from    string
	}
	// User is user
	User struct {
		*oauth2.Token
		session string
	}
	// Service is login service
	Service struct {
		Config *oauth2.Config

		mSess    sync.RWMutex
		sessions map[string]*session

		mUser sync.RWMutex
		users map[string]User

		UserCookie    string
		SessionCookie string

		Secure bool
		Domain string
		Path   string
		MaxAge int
	}
)

// NewService creates login service
func NewService(a *oauth2.Config) *Service {
	return &Service{
		Config:   a,
		mSess:    sync.RWMutex{},
		sessions: map[string]*session{},
		mUser:    sync.RWMutex{},
		users:    map[string]User{},

		UserCookie:    "_I",
		SessionCookie: "_s",

		Secure: false,
		Domain: "localhost",
		Path:   "/",
		MaxAge: 600,
	}
}

// GetUser gets user infomation
func (s *Service) GetUser(r *http.Request) (*User, error) {
	if c, err := r.Cookie(s.UserCookie); err != nil {
		return nil, err
	} else if u, ok := s.users[c.Value]; !ok {
		return nil, errors.New("no user " + c.Value)
	} else if sess, err := r.Cookie(s.SessionCookie); err != nil {
		return nil, err
	} else if sess.Value != u.session {
		return nil, errors.New("sesssion timeouot")
	} else {
		return &u, nil
	}
}

// RegisterUser registers user infomation with session
func (s *Service) RegisterUser(session, id string, t *oauth2.Token) {
	s.mSess.Lock()
	delete(s.sessions, session)
	s.mSess.Unlock()

	s.mUser.Lock()
	s.users[id] = User{session: session, Token: t}
	s.mUser.Unlock()
}

func (s *Service) registerSession(path string) string {
	state := newState()
	s.mSess.Lock()
	s.sessions[state] = &session{
		startAt: time.Now(),
		from:    path,
	}
	s.mSess.Unlock()
	return state
}
func (s *Service) getSession(state string) (*session, error) {
	s.mSess.RLock()
	sess, ok := s.sessions[state]
	s.mSess.RUnlock()
	if !ok {
		return nil, errors.New("no session in sessions")
	} else if time.Now().After(sess.startAt.Add(1 * time.Minute)) {
		return nil, errors.New("time out")
	}
	return sess, nil
}

// AddService add login service to mux
func AddService(m *http.ServeMux, path string, s *Service) func(http.HandlerFunc) http.HandlerFunc {
	m.HandleFunc(path, s.startLoginHandler)
	m.HandleFunc(path+"/callback", s.callbackHandler)
	return s.authenticator(path)
}

func (s *Service) authenticator(prefix string) func(http.HandlerFunc) http.HandlerFunc {
	return func(h http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if _, err := s.GetUser(r); err != nil {
				s.setCookie(w,
					s.SessionCookie, s.registerSession(r.URL.Path),
				)
				http.Redirect(w, r, prefix, http.StatusTemporaryRedirect)
				return
			}
			h(w, r)
		}
	}
}
func (s *Service) startLoginHandler(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(s.SessionCookie); err != nil {
		w.WriteHeader(http.StatusBadRequest)
	} else if _, err := s.getSession(c.Value); err != nil {
		w.WriteHeader(http.StatusBadRequest)
	} else {
		http.Redirect(w, r, s.Config.AuthCodeURL(c.Value), http.StatusTemporaryRedirect)
	}
}
func (s *Service) callbackHandler(w http.ResponseWriter, r *http.Request) {
	state := r.URL.Query().Get("state")
	sess, err := s.getSession(state)
	if err != nil {
		fmt.Println(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	code := r.URL.Query().Get("code")
	token, err := s.Config.Exchange(context.TODO(), code)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	id := newState()
	s.RegisterUser(state, id, token)
	s.setCookie(w, s.UserCookie, id)

	path := sess.from
	if path == "" {
		path = "/"
	}
	http.Redirect(w, r, path, http.StatusTemporaryRedirect)
}
func (s *Service) setCookie(w http.ResponseWriter, k, v string) {
	http.SetCookie(w, &http.Cookie{
		Name: k, Value: v,
		Secure: s.Secure,
		Domain: s.Domain,
		Path:   s.Path,
		MaxAge: s.MaxAge,
	})
}
func newState() string {
	return uuid.NewV4().String()
}
