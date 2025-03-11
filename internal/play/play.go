package play

import (
	"encoding/json"

	"github.com/vs-ude/btml/internal/model"
	"github.com/vs-ude/btml/internal/peer"
)

type Play struct {
	me    *peer.Me
	mod   model.Model
	steps []Step
}

func NewPlay(me *peer.Me, mod model.Model) *Play {
	return &Play{me, mod, []Step{}}
}

func (p *Play) AddStep(step Step) {
	p.steps = append(p.steps, step)
}

func (p *Play) Run() error {
	for _, step := range p.steps {
		if err := step.Run(p.me, p.mod); err != nil {
			return err
		}
	}
	return nil
}

func (p *Play) MarshalJSON() ([]byte, error) {
	// TODO This needs to differentiate between step types
	return json.Marshal(p.steps)
}

func (p *Play) UnmarshalJSON(data []byte) error {
	// TODO This needs to differentiate between step types
	return json.Unmarshal(data, &p.steps)
}
