package service

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/Duvewo/testexercise/handler"
	"github.com/Duvewo/testexercise/storage"
)

type AuthService struct {
	handler.Handler
}

type AuthForm struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Type     string `json:"type"`
}

func (srv *AuthService) Handle(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)

	if err != nil {
		log.Printf("wizard: failed to read body %v\n", err)
		http.Error(w, "Internal error", http.StatusConflict)
		return
	}

	var form AuthForm
	if err := json.Unmarshal(data, &form); err != nil {
		log.Printf("wizard: failed to unmarshal auth data %v\n", err)
		http.Error(w, "Internal error", http.StatusConflict)
		return
	}

	if form.Username == "" || form.Password == "" {
		http.Error(w, "Username or password are not provided", http.StatusBadRequest)
		return
	}

	switch form.Type {
	case "register":
		srv.Create(w, r, form)
	case "login":
		srv.Login(w, r, form)
	}

}

func (srv *AuthService) Create(w http.ResponseWriter, r *http.Request, form AuthForm) {
	exists, err := srv.Handler.Users.ExistsByUsername(
		context.Background(),
		storage.User{Username: form.Username},
	)

	if err != nil {
		log.Printf("register/exists: %v\n", err)
	}

	if exists {
		http.Error(w, "Account already exists", http.StatusForbidden)
		return
	}

	// У каждого мага 100 hp после регистрации.
	if err := srv.Handler.Users.Create(
		context.Background(),
		storage.User{
			Username:     form.Username,
			Password:     form.Password,
			HealthPoints: 100,
		},
	); err != nil {
		log.Printf("register: %v\n", err)
		http.Error(w, "Internal error", http.StatusConflict)
		return
	}
}

func (srv *AuthService) Login(w http.ResponseWriter, r *http.Request, form AuthForm) {
	exists, err := srv.Handler.Users.ExistsByUsernameAndPassword(
		context.Background(),
		storage.User{Username: form.Username, Password: form.Password},
	)

	if err != nil {
		log.Printf("login: failed to search: %v\n", err)
		http.Error(w, "Internal error", http.StatusConflict)
		return
	}

	if !exists {
		http.Error(w, "This account with this user/password combination doesn't exist!", http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)
}
