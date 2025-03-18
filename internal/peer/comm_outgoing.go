package peer

import (
	"log/slog"
	"net"
	"sync"
	"time"

	"github.com/quic-go/quic-go"
	"github.com/vs-ude/btml/internal/model"
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
			bytes, err := marshalUpdate(data, me.config.Name)
			if err != nil {
				slog.Warn("Failed marshaling model update", "error", err)
				continue
			}
			for _, peer := range me.peerset.Active {
				if distribute, _ := me.pds.Decide(peer, data); !distribute {
					continue
				}
				wg.Add(1)
				go peer.Send(bytes, data.GetAge(), wg, me.Ctx, me.dialPeer)
			}
		}
	}
}

func (me *Me) LaggingPeersLoop() {
	defer me.Wg.Done()

	wg := &sync.WaitGroup{}
	var data *model.Weights
	var err error
	timer := time.NewTimer(time.Second)
	wait := time.Duration(time.Second * 5)
	for {
		select {
		case <-me.Ctx.Done():
			return
		case <-timer.C:
			wg.Wait() // We wait here so the application can be stopped at any time
			for _, peer := range me.peerset.Active {
				if data, err = me.pds.Retrieve(peer.LastUpdatedAge); err != nil {
					continue
				}
				bytes, err := marshalUpdate(data, me.config.Name)
				if err != nil {
					slog.Warn("Failed marshaling model update", "error", err)
					continue
				}
				wg.Add(1)
				go peer.Send(bytes, data.GetAge(), wg, me.Ctx, me.dialPeer)

			}
			timer.Reset(wait)
		}
	}
}

func marshalUpdate(data *model.Weights, source string) ([]byte, error) {
	// Create and marshal the model update
	update := &ModelUpdate{
		Source:  source,
		Weights: data.Get(),
		Age:     int64(data.GetAge()),
	}

	return proto.Marshal(update)
}

func (me *Me) dialPeer(addr net.Addr) (quic.Connection, error) {
	return me.server.Dial(me.Ctx, addr, me.tlsConfig, me.quicConfig)
}
