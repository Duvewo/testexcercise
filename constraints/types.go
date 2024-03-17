package constraints

import "github.com/gorilla/websocket"

const DefaultHealthPoints = 100

type ResponseForm map[string]any

type Battle struct {
	Wizards []Wizard `json:"wizards"`
}

type Wizard struct {
	Username     string          `json:"username"`
	HealthPoints int             `json:"hp"`
	Client       *websocket.Conn `json:"-"`
}
