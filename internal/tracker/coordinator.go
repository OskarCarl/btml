package tracker

import (
	"encoding/json"
	"net/http"

	"github.com/vs-ude/btfl/internal/structs"
)

// initPeer gives a requesting peer all information it needs to initialize itself correctly to join the swarm
func (t *tracker) initPeer(w http.ResponseWriter, r *http.Request) {
	buf, _ := json.Marshal(structs.WhoAmI{Id: 42})
	w.Write(buf)
}
