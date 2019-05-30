package server

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/go-redis/redis"
	"github.com/pkg/errors"
	"github.com/pquerna/otp/hotp"

	"github.com/cmars/peerless"
)

type Service struct {
	db *redis.Client
}

func NewService(db *redis.Client) *Service {
	return &Service{db: db}
}

func (s *Service) TokenHandler(w http.ResponseWriter, r *http.Request) {
	fail := func(code int, err error) {
		w.WriteHeader(code)
		fmt.Fprintf(w, "%v", err)
	}
	if r.Method != "POST" {
		fail(http.StatusBadRequest, errors.Errorf("invalid method %v", r.Method))
		return
	}
	auth, err := peerless.NewAuthorization()
	if err != nil {
		fail(http.StatusInternalServerError, err)
		return
	}
	contents, err := json.Marshal(&auth)
	if err != nil {
		fail(http.StatusInternalServerError, err)
	}
	fmt.Fprintf(w, "%s", base64.StdEncoding.EncodeToString(contents))
}

func (s *Service) IndexHandler(w http.ResponseWriter, r *http.Request) {
	fail := func(code int, err error) {
		w.WriteHeader(code)
		fmt.Fprintf(w, "%v", err)
	}
	ok, statusCode, err := s.Authorize(r)
	if err != nil {
		log.Printf("%+v", err)
		fail(statusCode, err)
		return
	}
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintf(w, "authorization required")
		return
	}
	fmt.Fprintf(w, "OK")
}

func (s *Service) Authorize(req *http.Request) (bool, int, error) {
	authHeader := req.Header.Get("Authorization")
	authHeaderFields := strings.SplitN(authHeader, " ", 2)
	if len(authHeaderFields) < 2 {
		return false, http.StatusBadRequest, errors.New("invalid authorization")
	}
	if authHeaderFields[0] != "Bearer" {
		return false, http.StatusBadRequest, errors.Errorf("unsupported authorization %q", authHeaderFields[0])
	}
	token := authHeaderFields[1]
	if token == "" {
		return false, http.StatusUnauthorized, nil
	}
	contents, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return false, http.StatusBadRequest, errors.Wrap(err, "failed to decode authorization")
	}
	var auth peerless.Authorization
	if err := json.Unmarshal(contents, &auth); err != nil {
		return false, http.StatusBadRequest, errors.Wrap(err, "failed to unmarshal authorization")
	}
	log.Printf("request from %s", auth.Secret)
	counter, err := s.db.Incr(auth.Secret).Result()
	if err != nil {
		log.Printf("failed to increment token counter")
		return false, http.StatusBadRequest, err
	}
	return hotp.Validate(auth.Code, uint64(counter-1), auth.Secret), http.StatusOK, nil
}

func Run() {
	db := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	_, err := db.Ping().Result()
	if err != nil {
		log.Fatalf("failed to connect to redis: %v", err)
	}
	s := NewService(db)
	http.HandleFunc("/token", s.TokenHandler)
	http.HandleFunc("/", s.IndexHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
