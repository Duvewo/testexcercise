package main

import (
	"context"
	"encoding/json"
	"flag"
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

type ReqRespForm struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

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

	db.Ping(context.Background())

	upgrader := websocket.Upgrader{}
	r.HandleFunc("/wizard", func(w http.ResponseWriter, r *http.Request) {
		data, err := io.ReadAll(r.Body)

		if err != nil {
			return
		}

		var form AuthForm
		if err := json.Unmarshal(data, &form); err != nil {
			return
		}

		switch form.Type {
		case "register":
			if _, err := db.Exec(context.Background(), "INSERT INTO users (username, password) VALUES ($1, $2)", form.Username, form.Password); err != nil {
				log.Println(err)
			}
		case "login":
			db.QueryRow(context.Background(), "SELECT * FROM users WHERE username = $1 AND password = $2", form.Username, form.Password)
			log.Println("Auth est reussi comme " + form.Username)
		}

	}).Methods("POST")

	battle := Battle{}
	r.HandleFunc("/battle", func(w http.ResponseWriter, r *http.Request) {
		ws, err := upgrader.Upgrade(w, r, nil)

		if err != nil {
			return
		}

		var req ReqRespForm
		ws.ReadJSON(&req)

		switch req.Name {
		case "join":

			for _, wizard := range battle.Wizards {
				//Как только маг присоединяется, всем остальным участникам приходит оповещение об этом с именем мага

				wizard.Client.WriteJSON(ReqRespForm{Name: "new-wizard", Data: ""})
			}

			// wizards, err := json.Marshal(battle.Wizards)

			// if err != nil {
			// 	log.Fatalf("failed to marshal wizards: %s", err)
			// }

			ws.WriteJSON(ReqRespForm{Name: "info"})

			battle.Wizards = append(battle.Wizards, Wizard{
				// Username:     req.WName,
				HealthPoints: 100,
				Client:       ws,
			})

		case "attack":
			// log.Println("On veut frapper: " + req.WName)
		}

	})

	log.Fatalf("failed to init server: %s", http.ListenAndServe(*FLAG_SRVADDR, r))

}
