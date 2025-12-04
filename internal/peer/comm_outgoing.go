package peer

import (
	"log/slog"
	"net"
	"sync"
	"time"

	"github.com/quic-go/quic-go"
	"github.com/vs-ude/btml/internal/structs"
	"google.golang.org/protobuf/proto"
)

func (me *Me) Outgoing() {
	defer me.Wg.Done()

	wg := &sync.WaitGroup{}
	for {
		select {
		case <-me.Ctx.Done():
			return
		case data := <-me.data.outgoingChan:
			wg.Wait() // We wait here so the application can be stopped at any time
			me.pds.Store(*data)
			if me.telemetry != nil {
				me.telemetry.RecordOnline(data.GetAge())
			}
			bytes, err := marshalUpdate(data, me.config.Name)
			if err != nil {
				slog.Warn("Failed marshaling model update", "error", err)
				continue
			}
			for _, peer := range me.peerset.GetUnchoked() {
				if distribute, _ := me.pds.Decide(peer, data); !distribute {
					continue
				}
				slog.Debug("Sending data to peer", "target", peer.Name, "age", data.GetAge())
				wg.Add(1)
				go peer.Send(bytes, data.GetAge(), wg, me.Ctx, me.dialPeer)
			}
		}
	}
}

func (me *Me) LaggingPeersLoop() {
	defer me.Wg.Done()

	wg := &sync.WaitGroup{}
	var data *structs.Weights
	var err error
	timer := time.NewTimer(time.Second)
	wait := time.Duration(time.Second * 5)
	for {
		select {
		case <-me.Ctx.Done():
			return
		case <-timer.C:
			wg.Wait() // We wait here so the application can be stopped at any time
			for _, peer := range me.peerset.GetUnchoked() {
				if data, err = me.pds.Retrieve(peer.LastSentUpdateAge); err != nil {
					slog.Debug("Did not get data for peer", "peer", peer.Name, "error", err)
					continue
				}
				bytes, err := marshalUpdate(data, me.config.Name)
				if err != nil {
					slog.Warn("Failed marshaling model update", "error", err)
					continue
				}
				slog.Debug("Sending data to lagging peer", "target", peer.Name, "age", data.GetAge())
				wg.Add(1)
				go peer.Send(bytes, data.GetAge(), wg, me.Ctx, me.dialPeer)

			}
			timer.Reset(wait)
		}
	}
}

func marshalUpdate(data *structs.Weights, source string) ([]byte, error) {
	// Create and marshal the model update
	update := &ModelUpdate{
		Source:  source,
		Weights: data.Get(),
		Age:     int64(data.GetAge()),
	}

	return proto.Marshal(update)
}

func (me *Me) dialPeer(addr net.Addr) (*quic.Conn, error) {
	return me.server.Dial(me.Ctx, addr, me.tlsConfig, me.quicConfig)
}
