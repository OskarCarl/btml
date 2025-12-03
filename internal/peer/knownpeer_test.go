package peer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vs-ude/btml/internal/structs"
	"github.com/vs-ude/btml/internal/telemetry"
	"github.com/vs-ude/btml/internal/trust"
)

func TestKnownPeerScore(t *testing.T) {
	// Create a mock peer
	mockPeer := &structs.Peer{Name: "peer1"}
	telemetryClient := &telemetry.Client{}
	kp := NewKnownPeer(mockPeer, telemetryClient)

	// Test initial score
	assert.Equal(t, trust.Score(0), kp.GetScore(), "Initial score should be 0")

	// Update score and verify
	kp.UpdateScore(2)
	assert.Equal(t, 2, int(kp.GetScore()), "Score should be updated to 2")

	kp.UpdateScore(-7)
	assert.Equal(t, 0, int(kp.GetScore()), "Score should be updated to 0")
}
