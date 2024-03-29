package service

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/Duvewo/testexercise/constraints"
	"github.com/Duvewo/testexercise/handler"
	"github.com/Duvewo/testexercise/storage"
	"github.com/gorilla/websocket"
)

type BattleService struct {
	*constraints.Battle
	websocket.Upgrader
	handler.Handler
}

func (srv *BattleService) Handle(w http.ResponseWriter, r *http.Request) {
	ws, err := srv.Upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Fatalf("battle: failed to upgrade %v", err)

		return
	}

	q := r.URL.Query()

	user, err := srv.Handler.Users.ByUsername(context.Background(), storage.User{Username: q.Get("username")})

	if err != nil {
		log.Printf("battle: failed to find account %v\n", err)
		http.Error(w, "Internal error", http.StatusConflict)

		return
	}

	// При разрыве соединения у мага всем текущим участникам приходит ивент об этом с указанием имени мага.
	ws.SetCloseHandler(func(code int, text string) error {
		log.Println(code, text)
		for _, wizard := range srv.Battle.All() {
			wizard.Client.WriteJSON(constraints.ResponseForm{
				"type": "info",
				"data": fmt.Sprintf("Wizard %s left", user.Username),
			})
		}

		srv.Battle.Delete(constraints.Wizard{
			Username: user.Username,
		})
		return ws.Close()

	})

	switch q.Get("type") {
	case "join":
		for _, wizard := range srv.Battle.All() {
			// Как только маг присоединяется, всем остальным участникам приходит оповещение об этом с именем мага.
			wizard.Client.WriteJSON(constraints.ResponseForm{
				"username": q.Get("username"),
				"type":     "new",
			})
		}

		// При установке соединения магу приходит ивент с информацией о текущих участниках битвы.
		ws.WriteJSON(constraints.ResponseForm{
			"wizards": srv.Battle.All(),
		})

		srv.Battle.Add(
			constraints.Wizard{
				Username:     user.Username,
				HealthPoints: user.HealthPoints,
				Client:       ws,
			},
		)

		// Маг может отправить ивент-фаербол в другого указанного в ивенте мага
	case "attack":
		target := q.Get("target")
		attacker := q.Get("username")

		if exists, _ := srv.Battle.FindByUsername(constraints.Wizard{
			Username: attacker,
		}); !exists {
			ws.WriteJSON(constraints.ResponseForm{
				"type": "error",
				"data": "You are not in game",
			})
			ws.Close()
		}

		// Если маг умер, он больше не может кинуть фаербол.
		if user.HealthPoints <= 0 {
			ws.WriteJSON(constraints.ResponseForm{
				"type": "info",
				"data": "You can't throw a fireball because you are dead",
			})

			srv.Battle.Delete(
				constraints.Wizard{
					Username: user.Username,
				},
			)

			// Если маг умер, его вебсокет соединение обрывается.
			ws.Close()

			return
		}

		exists, wizard := srv.Battle.FindByUsername(constraints.Wizard{
			Username: target,
		})

		if !exists {
			ws.WriteJSON(constraints.ResponseForm{
				"type": "error",
				"data": "This wizard is offline",
			})

			return
		}

		targetUsr, err := srv.Handler.Users.ByUsername(context.Background(), storage.User{Username: target})

		if err != nil {
			log.Printf("attack: failed to find account %v\n", err)
			http.Error(w, "Internal error", http.StatusConflict)

			return
		}

		// Каждый фаербол отнимает у цели 10 hp.
		targetUsr.HealthPoints -= 10

		if err := srv.Handler.Users.SetHealth(context.Background(), targetUsr); err != nil {
			log.Printf("attack: failed to set health %v\n", err)
			http.Error(w, "Internal error", http.StatusConflict)
			return
		}

		ws.WriteJSON(constraints.ResponseForm{
			"type": "info",
			"data": fmt.Sprintf("You have successfully hit %s", targetUsr.Username),
		})

		exists, wizard = srv.Battle.FindByUsername(constraints.Wizard{Username: targetUsr.Username})

		if !exists {
			//todo: probably create separate function Exists()
			return
		}

		// Если у мага уменьшилось здоровье, ему приходит ивент с актуальным кол-вом здоровья.
		wizard.Client.WriteJSON(
			constraints.ResponseForm{
				"type": "info",
				"data": fmt.Sprintf("Wizard %s have hit you: %v", attacker, targetUsr.HealthPoints),
			})

		// Если здоровье мага опустилось до 0 и ниже, он умер.
		if targetUsr.HealthPoints <= 0 {
			wizard.Client.WriteJSON(constraints.ResponseForm{
				"type": "info",
				"data": "You are dead",
			})

			srv.Battle.Delete(constraints.Wizard{
				Username: targetUsr.Username,
			})

			// Если маг умер, его вебсокет соединение обрывается.
			wizard.Client.Close()
		}

		return
	}
}
