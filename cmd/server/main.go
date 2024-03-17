package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/Duvewo/testexercise/constraints"
	"github.com/Duvewo/testexercise/handler"
	"github.com/Duvewo/testexercise/service"
	"github.com/Duvewo/testexercise/storage"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var (
	FLAG_DBADDR  = flag.String("db-addr", os.Getenv("DB_ADDR"), "Database connectivity string")
	FLAG_SRVADDR = flag.String("srv-addr", os.Getenv("SRV_ADDR"), "Server address")
)

func main() {
	flag.Parse()
	router := mux.NewRouter()

	db, err := storage.InitDB(context.Background(), *FLAG_DBADDR)

	if err != nil {
		log.Fatalf("failed to init db: %v", err)
	}

	if err := db.Ping(context.Background()); err != nil {
		log.Fatalf("failed to ping db: %v", err)
	}

	h := handler.Handler{Pool: db, Users: &storage.Users{Pool: db}}
	upgrader := websocket.Upgrader{}

	auth := &service.AuthService{Handler: h}
	battle := &service.BattleService{
		Handler:  h,
		Upgrader: upgrader,
		Battle:   constraints.Battle{},
	}

	router.HandleFunc("/wizard", auth.Handle).Methods(http.MethodPost)
	router.HandleFunc("/battle", battle.Handle)

	log.Fatalf("failed to init server: %s", http.ListenAndServe(*FLAG_SRVADDR, router))
}
