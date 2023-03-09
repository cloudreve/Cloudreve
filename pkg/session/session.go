package session

// TODO: unit test

import (
	"encoding/base32"
	"net/http"
	"strings"
	"time"

	ginSession "github.com/gin-contrib/sessions"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/jinzhu/gorm"
)

type UserSession struct {
	ID        string `gorm:"unique_index"`
	Data      string `gorm:"text"`
	CreatedAt time.Time
	UpdatedAt time.Time
	ExpiresAt time.Time `gorm:"index"`
}

type SessionCacher interface {
	Setup()
	Get(sessionID string) *UserSession
	Create(val *UserSession) error
	Update(val *UserSession) error
	Delete(val *UserSession) error
}

type SqliteSessionCacher struct {
	db *gorm.DB
}

func (c *SqliteSessionCacher) sessionTable() *gorm.DB {
	return c.db.Table("user_sessions")
}

func (c *SqliteSessionCacher) Setup() {
	c.sessionTable().AutoMigrate(&UserSession{})
}

func (c *SqliteSessionCacher) Get(sessionID string) *UserSession {
	// after get session id, try get persisted session
	persistedSession := &UserSession{}
	record := c.sessionTable().
		Where("id = ? AND expires_at > ?", sessionID, time.Now()).
		Limit(1).
		Find(persistedSession)

	if record.Error != nil || record.RowsAffected == 0 {
		return nil
	}

	return persistedSession
}

func (c *SqliteSessionCacher) Create(val *UserSession) error {
	return c.sessionTable().Create(val).Error
}

func (c *SqliteSessionCacher) Update(val *UserSession) error {
	return c.sessionTable().Save(val).Error
}

func (c *SqliteSessionCacher) Delete(val *UserSession) error {
	return c.sessionTable().Delete(val).Error
}

type SessionStore struct {
	SessionOptions *sessions.Options
	Codecs         []securecookie.Codec
	cache          SessionCacher
}

func NewSessionStore(db *gorm.DB, keyPires ...[]byte) *SessionStore {
	store := SessionStore{
		SessionOptions: &sessions.Options{
			Path:     "/",
			MaxAge:   60 * 86400,
			HttpOnly: true,
		},
		Codecs: securecookie.CodecsFromPairs(keyPires...),
		cache:  &SqliteSessionCacher{db: db},
	}

	store.setup()
	store.MaxAge(store.SessionOptions.MaxAge)
	return &store
}

/*
	Session interface implementation
*/

func (s *SessionStore) New(r *http.Request, name string) (*sessions.Session, error) {
	session := sessions.NewSession(s, name)
	session.Options = s.SessionOptions
	session.IsNew = true

	s.MaxAge(s.SessionOptions.MaxAge)
	if persistedSession := s.tryGetPersistedSessionFromCookie(r, name); persistedSession != nil {
		// decode persisted session data to session.Values if found
		err := securecookie.DecodeMulti(session.Name(), persistedSession.Data, &session.Values, s.Codecs...)
		if err != nil {
			return session, err
		}

		session.ID = persistedSession.ID
		session.IsNew = false
		return session, nil
	}

	return session, nil
}

func (s *SessionStore) Save(r *http.Request, w http.ResponseWriter, session *sessions.Session) error {
	// try get current session from cookie
	persistedSession := s.tryGetPersistedSessionFromCookie(r, session.Name())

	// delete session if MaxAge < 0
	if session.Options.MaxAge < 0 {
		if persistedSession != nil {
			// delete persisted session
			if err := s.cache.Delete(persistedSession); err != nil {
				return err
			}
		}

		// delete cookie
		http.SetCookie(w, sessions.NewCookie(session.Name(), "", session.Options))
		return nil
	}

	// or cretae new session / update current session
	data, err := securecookie.EncodeMulti(session.Name(), session.Values, s.Codecs...)
	if err != nil {
		return err
	}

	if persistedSession == nil {
		session.ID = strings.TrimRight(
			base32.StdEncoding.EncodeToString(
				securecookie.GenerateRandomKey(32)), "=")

		// create new session
		persistedSession = &UserSession{
			ID:        session.ID,
			Data:      data,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			ExpiresAt: time.Now().Add(time.Duration(session.Options.MaxAge) * time.Second),
		}

		if err := s.cache.Create(persistedSession); err != nil {
			return err
		}
	} else {
		// update current session
		persistedSession.Data = data
		persistedSession.UpdatedAt = time.Now()
		persistedSession.ExpiresAt = time.Now().Add(time.Duration(session.Options.MaxAge) * time.Second)

		if err := s.cache.Update(persistedSession); err != nil {
			return err
		}
	}

	// set cookie
	encoded, err := securecookie.EncodeMulti(session.Name(), session.ID, s.Codecs...)
	if err != nil {
		return err
	}

	http.SetCookie(w, sessions.NewCookie(session.Name(), encoded, session.Options))
	return nil
}

func (s *SessionStore) Get(r *http.Request, name string) (*sessions.Session, error) {
	return sessions.GetRegistry(r).Get(s, name)
}

/*
	Handy tools
*/

func (st *SessionStore) setup() {
	st.cache.Setup()
}

func (s *SessionStore) MaxAge(age int) {
	s.SessionOptions.MaxAge = age

	// Set the maxAge for each securecookie instance.
	for _, codec := range s.Codecs {
		if sc, ok := codec.(*securecookie.SecureCookie); ok {
			sc.MaxAge(age)
		}
	}
}

func (s *SessionStore) tryGetPersistedSessionFromCookie(r *http.Request, name string) *UserSession {
	// get cookie from request
	cookie, err := r.Cookie(name)
	if err != nil {
		return nil
	}

	// decode cookie value to session id
	var sessionID string
	err = securecookie.DecodeMulti(name, cookie.Value, &sessionID, s.Codecs...)
	if err != nil {
		return nil
	}

	// after get session id, try get persisted session
	return s.cache.Get(sessionID)
}

/*
	Below is gin-session wrapper
*/

type Store interface {
	ginSession.Store
}

type store struct {
	*SessionStore
}

func (s *store) Options(options ginSession.Options) {
	s.SessionStore.SessionOptions = options.ToGorillaOptions()
}

func NewStore(db *gorm.DB, keyPires ...[]byte) Store {
	return &store{NewSessionStore(db, keyPires...)}
}
