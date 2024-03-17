package constraints

import (
	"sync"

	"github.com/gorilla/websocket"
)

const DefaultHealthPoints = 100

type ResponseForm map[string]any

type Battle struct {
	sync.Mutex
	Wizards []Wizard `json:"wizards"`
}

func (b *Battle) Add(wizard Wizard) {
	b.Lock()
	defer b.Unlock()
	b.Wizards = append(b.Wizards, wizard)
}
func (b *Battle) FindByUsername(wizard Wizard) (bool, Wizard) {
	b.Lock()
	defer b.Unlock()
	for _, w := range b.Wizards {
		if w.Username == wizard.Username {
			return true, w
		}
	}
	return false, Wizard{}
}

func (b *Battle) Delete(wizard Wizard) {
	b.Lock()
	defer b.Unlock()
	for i, w := range b.Wizards {
		if w.Username == wizard.Username {
			b.Wizards = append(b.Wizards[:i], b.Wizards[i+1:]...)
			break
		}
	}
}

func (b *Battle) All() []Wizard {
	b.Lock()
	defer b.Unlock()
	return b.Wizards
}

type Wizard struct {
	Username     string          `json:"username"`
	HealthPoints int             `json:"hp"`
	Client       *websocket.Conn `json:"-"`
}
