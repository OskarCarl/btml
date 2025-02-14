package tracker

import (
	"crypto/rand"
	"encoding/json"
	"math/big"
	"net/http"

	"github.com/vs-ude/btfl/internal/structs"
)

// initPeer gives a requesting peer all information it needs to initialize itself correctly to join the swarm
func (t *Tracker) initPeer(w http.ResponseWriter, r *http.Request) {
	i, _ := rand.Int(rand.Reader, big.NewInt(100))
	who := structs.WhoAmI{
		Id:      int(i.Int64()),
		Dataset: "fMNIST",
	}
	buf, _ := json.Marshal(who)
	w.Write(buf)
}
