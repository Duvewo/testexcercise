package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/Duvewo/testexercise/storage"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var (
	FLAG_DBADDR  = flag.String("db-addr", os.Getenv("DB_ADDR"), "Database connectivity string")
	FLAG_SRVADDR = flag.String("srv-addr", os.Getenv("SRV_ADDR"), "Server address")
)

type ResponseForm map[string]any

type AuthForm struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Type     string `json:"type"`
}

type Battle struct {
	Wizards []Wizard `json:"wizards"`
}

type Wizard struct {
	Username     string          `json:"username"`
	HealthPoints int             `json:"hp"`
	Client       *websocket.Conn `json:"-"`
}

func main() {
	flag.Parse()
	r := mux.NewRouter()

	db, err := storage.InitDB(context.Background(), *FLAG_DBADDR)

	if err != nil {
		log.Fatalf("failed to init db: %w", err)
	}

	if err := db.Ping(context.Background()); err != nil {
		log.Fatalf("failed to ping db: %v", err)
	}

	users := storage.Users{Pool: db}

	upgrader := websocket.Upgrader{}
	r.HandleFunc("/wizard", func(w http.ResponseWriter, r *http.Request) {
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

		switch form.Type {
		case "register":
			if err := users.Create(context.Background(), storage.User{Username: form.Username, Password: form.Password}); err != nil {
				log.Printf("register: %v\n", err)
				http.Error(w, "Internal error", http.StatusConflict)
				return
			}
		case "login":
			exists, err := users.ExistsByUsernameAndPassword(context.Background(), storage.User{Username: form.Username, Password: form.Password})

			if err != nil {
				log.Printf("login: failed to search: %v\n", err)
				http.Error(w, "Internal error", http.StatusConflict)
				return
			}

			if !exists {
				http.Error(w, "This account doesn't exist!", http.StatusUnauthorized)
				return
			}

			w.WriteHeader(http.StatusOK)

		}

	}).Methods("POST")

	battle := Battle{}
	r.HandleFunc("/battle", func(w http.ResponseWriter, r *http.Request) {
		ws, err := upgrader.Upgrade(w, r, nil)

		if err != nil {
			log.Fatalf("battle: failed to upgrade %v", err)
			return
		}

		q := r.URL.Query()

		user, err := users.ByUsername(context.Background(), storage.User{Username: q.Get("username")})

		if err != nil {
			log.Printf("battle: failed to find account %v\n", err)
			http.Error(w, "Internal error", http.StatusConflict)
			return
		}

		ws.SetCloseHandler(func(code int, text string) error {
			log.Println("someone left")
			for _, wizard := range battle.Wizards {
				wizard.Client.WriteJSON(ResponseForm{
					"type": "info",
					"data": fmt.Sprintf("Wizard %s left", user.Username),
				})
			}

			return ws.Close()

		})

		switch q.Get("type") {
		case "join":

			for _, wizard := range battle.Wizards {
				// При установке соединения магу приходит ивент с информацией о текущих участниках битвы
				//Как только маг присоединяется, всем остальным участникам приходит оповещение об этом с именем мага
				ws.WriteJSON(ResponseForm{"username": wizard.Username, "health_points": wizard.HealthPoints})
				wizard.Client.WriteJSON(ResponseForm{"username": q.Get("username")})
			}

			battle.Wizards = append(battle.Wizards, Wizard{
				Username:     user.Username,
				HealthPoints: user.HealthPoints,
				Client:       ws,
			})

		case "attack":
			target := q.Get("target")
			attacker := q.Get("username")

			if user.HealthPoints <= 0 {
				ws.WriteJSON(ResponseForm{
					"type": "info",
					"data": "You can't throw a fireball because you are dead",
				})
				ws.Close()
				return
			}

			targetUsr, err := users.ByUsername(context.Background(), storage.User{Username: target})

			if err != nil {
				log.Printf("attack: failed to find account %v\n", err)
				http.Error(w, "Internal error", http.StatusConflict)
				return
			}

			targetUsr.HealthPoints -= 10

			if err := users.SetHealth(context.Background(), targetUsr); err != nil {
				log.Printf("attack: failed to set health %v\n", err)
				http.Error(w, "Internal error", http.StatusConflict)
				return
			}

			for _, wizard := range battle.Wizards {
				if wizard.Username == targetUsr.Username {
					wizard.Client.WriteJSON(
						ResponseForm{
							"type": "info",
							"data": fmt.Sprintf("Wizard %s vous avez frappé, votre PV est: %v", attacker, targetUsr.HealthPoints),
						})

					if targetUsr.HealthPoints <= 0 {
						wizard.Client.WriteJSON(ResponseForm{
							"type": "info",
							"data": "You are dead",
						})
						wizard.Client.Close()
					}

					return
				}
			}

		}

	})

	log.Fatalf("failed to init server: %s", http.ListenAndServe(*FLAG_SRVADDR, r))

}
