package constraints

import "github.com/gorilla/websocket"

type ResponseForm map[string]any

type Battle struct {
	Wizards []Wizard `json:"wizards"`
}

type Wizard struct {
	Username     string          `json:"username"`
	HealthPoints int             `json:"hp"`
	Client       *websocket.Conn `json:"-"`
}
