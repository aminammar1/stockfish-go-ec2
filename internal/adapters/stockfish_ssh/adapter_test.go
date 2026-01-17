package stockfish_ssh

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/aminammar1/stockfish-go-ec2/internal/config"
	"github.com/aminammar1/stockfish-go-ec2/internal/ports"
)

func TestHealth(t *testing.T) {
	cfg := config.Config{
		SSHHost:       os.Getenv("SSH_HOST"),
		SSHUser:       os.Getenv("SSH_USER"),
		SSHPrivateKey: os.Getenv("SSH_PRIVATE_KEY"),
		SSHPort:       22,
		SSHTimeout:    10 * time.Second,
		StockfishPath: os.Getenv("STOCKFISH_PATH"),
	}

	if cfg.SSHHost == "" {
		t.Skip("SSH_HOST not set")
	}

	adapter := NewAdapter(cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	err := adapter.Health(ctx)
	if err != nil {
		t.Errorf("Health check failed: %v", err)
	}
}

func TestBuildPositionCommand_FEN(t *testing.T) {
	tests := []struct {
		name    string
		fen     string
		wantErr bool
	}{
		{"starting position", "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1", false},
		{"after e4", "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1", false},
		{"invalid fen", "invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := ports.AnalyzeRequest{FEN: tt.fen}
			_, _, err := buildPositionCommand(req)
			if (err != nil) != tt.wantErr {
				t.Errorf("buildPositionCommand() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestComputeEvalBar(t *testing.T) {
	tests := []struct {
		name    string
		cp      *int
		mate    *int
		wantMin int
		wantMax int
	}{
		{"equal position", intPtr(0), nil, 48, 52},
		{"white advantage +100cp", intPtr(100), nil, 55, 65},
		{"black advantage -100cp", intPtr(-100), nil, 35, 45},
		{"mate for white", nil, intPtr(5), 100, 100},
		{"mate for black", nil, intPtr(-5), 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bar := computeEvalBar(tt.cp, tt.mate)
			if bar == nil {
				t.Error("computeEvalBar() returned nil")
				return
			}
			if *bar < tt.wantMin || *bar > tt.wantMax {
				t.Errorf("computeEvalBar() = %d, want between %d and %d", *bar, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func intPtr(i int) *int {
	return &i
}
