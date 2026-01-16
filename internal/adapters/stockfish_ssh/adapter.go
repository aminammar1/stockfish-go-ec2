package stockfish_ssh

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"regexp"
	"strings"

	"github.com/notnil/chess"
	"golang.org/x/crypto/ssh"

	"stockfish-ec2-service/internal/config"
	"stockfish-ec2-service/internal/ports"
)

type Adapter struct {
	cfg config.Config
}

func NewAdapter(cfg config.Config) *Adapter {
	return &Adapter{cfg: cfg}
}

func (a *Adapter) Health(ctx context.Context) error {
	client, err := a.dial(ctx)
	if err != nil {
		return err
	}
	defer client.Close()
	return nil
}

func (a *Adapter) Analyze(ctx context.Context, req ports.AnalyzeRequest) (ports.AnalyzeResult, error) {
	client, err := a.dial(ctx)
	if err != nil {
		return ports.AnalyzeResult{}, err
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return ports.AnalyzeResult{}, err
	}
	defer session.Close()

	var stdin bytes.Buffer
	var stdout bytes.Buffer
	session.Stdin = &stdin
	session.Stdout = &stdout
	session.Stderr = &stdout

	posCmd, err := buildPositionCommand(req)
	if err != nil {
		return ports.AnalyzeResult{}, err
	}

	depth := a.cfg.AnalysisDepth
	stdin.WriteString("uci\n")
	stdin.WriteString("isready\n")
	stdin.WriteString(posCmd + "\n")
	stdin.WriteString(fmt.Sprintf("go depth %d\n", depth))
	stdin.WriteString("quit\n")

	cmd := a.cfg.StockfishPath
	if err := session.Run(cmd); err != nil {
		return ports.AnalyzeResult{}, err
	}

	output := stdout.String()
	bestMove := parseBestMove(output)
	eval := parseEval(output)

	return ports.AnalyzeResult{BestMove: bestMove, Evaluation: eval, Raw: output}, nil
}

func (a *Adapter) dial(ctx context.Context) (*ssh.Client, error) {
	if a.cfg.SSHHost == "" || a.cfg.SSHUser == "" {
		return nil, errors.New("SSH_HOST and SSH_USER required")
	}

	auth, err := buildAuth(a.cfg)
	if err != nil {
		return nil, err
	}

	config := &ssh.ClientConfig{
		User:            a.cfg.SSHUser,
		Auth:            []ssh.AuthMethod{auth},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         a.cfg.SSHTimeout,
	}

	addr := fmt.Sprintf("%s:%d", a.cfg.SSHHost, a.cfg.SSHPort)
	dialer := net.Dialer{}
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, err
	}

	c, chans, reqs, err := ssh.NewClientConn(conn, addr, config)
	if err != nil {
		return nil, err
	}
	return ssh.NewClient(c, chans, reqs), nil
}

func buildAuth(cfg config.Config) (ssh.AuthMethod, error) {
	if cfg.SSHPrivateKey != "" {
		keyData := []byte(cfg.SSHPrivateKey)
		if _, err := os.Stat(cfg.SSHPrivateKey); err == nil {
			if data, readErr := os.ReadFile(cfg.SSHPrivateKey); readErr == nil {
				keyData = data
			}
		}
		signer, err := ssh.ParsePrivateKey(keyData)
		if err != nil {
			return nil, err
		}
		return ssh.PublicKeys(signer), nil
	}
	if cfg.SSHPassword != "" {
		return ssh.Password(cfg.SSHPassword), nil
	}
	return nil, errors.New("SSH_PASSWORD or SSH_PRIVATE_KEY required")
}

func buildPositionCommand(req ports.AnalyzeRequest) (string, error) {
	if req.FEN == "" && req.PGN == "" && req.UCIMoves == "" && req.SANMoves == "" {
		return "", errors.New("fen, pgn, uci or san required")
	}

	if req.PGN != "" {
		pgn, err := chess.PGN(strings.NewReader(req.PGN))
		if err != nil {
			return "", err
		}
		game := chess.NewGame(pgn)
		pos := game.Position()
		return "position fen " + pos.String(), nil
	}

	base := "position startpos"
	if req.FEN != "" {
		base = "position fen " + req.FEN
	}

	if strings.TrimSpace(req.SANMoves) != "" {
		uciMoves, err := sanToUciMoves(req.SANMoves, req.FEN)
		if err != nil {
			return "", err
		}
		return base + " moves " + uciMoves, nil
	}

	if strings.TrimSpace(req.UCIMoves) != "" {
		return base + " moves " + strings.TrimSpace(req.UCIMoves), nil
	}

	return base, nil
}

func sanToUciMoves(sanMoves, fen string) (string, error) {
	var game *chess.Game
	if fen != "" {
		fenOpt, err := chess.FEN(fen)
		if err != nil {
			return "", err
		}
		game = chess.NewGame(fenOpt)
	} else {
		game = chess.NewGame()
	}

	pos := game.Position()
	uci := chess.UCINotation{}
	alg := chess.AlgebraicNotation{}

	var moves []string
	for _, token := range strings.Fields(sanMoves) {
		moveToken := cleanSANToken(token)
		if moveToken == "" {
			continue
		}
		move, err := alg.Decode(pos, moveToken)
		if err != nil {
			return "", err
		}
		moves = append(moves, uci.Encode(pos, move))
		pos = pos.Update(move)
	}

	if len(moves) == 0 {
		return "", errors.New("no SAN moves parsed")
	}
	return strings.Join(moves, " "), nil
}

func cleanSANToken(token string) string {
	if token == "" {
		return ""
	}
	if token == "1-0" || token == "0-1" || token == "1/2-1/2" {
		return ""
	}
	if strings.HasSuffix(token, ".") {
		return ""
	}
	if strings.Contains(token, ".") {
		parts := strings.Split(token, ".")
		return parts[len(parts)-1]
	}
	return token
}

func parseBestMove(output string) string {
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "bestmove ") {
			parts := strings.Split(line, " ")
			if len(parts) >= 2 {
				return parts[1]
			}
		}
	}
	return ""
}

func parseEval(output string) string {
	re := regexp.MustCompile(`score (cp|mate) (-?\d+)`)
	matches := re.FindAllStringSubmatch(output, -1)
	if len(matches) == 0 {
		return ""
	}
	last := matches[len(matches)-1]
	if len(last) < 3 {
		return ""
	}
	return fmt.Sprintf("%s %s", last[1], last[2])
}
