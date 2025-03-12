package play

import (
	"errors"
	"log/slog"
	"strconv"
	"time"

	"github.com/vs-ude/btml/internal/model"
	"github.com/vs-ude/btml/internal/peer"
)

type Step interface {
	Setup(string) error
	Run(*peer.Me, *model.Model) error
}

type Train struct {
}

func (t *Train) Setup(_ string) error {
	return nil
}

func (t *Train) Run(_ *peer.Me, mod *model.Model) error {
	// log.Default().Println("Training model")
	mod.Train()
	return nil
}

type Eval struct {
}

func (t *Eval) Setup(_ string) error {
	return nil
}

func (t *Eval) Run(_ *peer.Me, mod *model.Model) error {
	// log.Default().Println("Evaluating model")
	mod.Eval()
	return nil
}

type Wait struct {
	T time.Duration
}

func (w *Wait) Setup(in string) error {
	w.T, _ = time.ParseDuration(in)
	return nil
}

func (w *Wait) Run(_ *peer.Me, _ *model.Model) error {
	slog.Debug("Waiting", "duration", w.T)
	time.Sleep(w.T)
	return nil
}

type IncreaseData struct {
	Ratio float32
}

func (c *IncreaseData) Setup(in string) error {
	f64, _ := strconv.ParseFloat(in, 32)
	c.Ratio = float32(f64)
	return nil
}

func (c *IncreaseData) Run(_ *peer.Me, mod *model.Model) error {
	return errors.New("not implemented")
}
