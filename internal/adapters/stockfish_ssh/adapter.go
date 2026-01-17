package stockfish_ssh

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

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

	stdinPipe, err := session.StdinPipe()
	if err != nil {
		return ports.AnalyzeResult{}, err
	}

	var stdout bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stdout

	posCmd, pos, err := buildPositionCommand(req)
	if err != nil {
		return ports.AnalyzeResult{}, err
	}

	cmd := a.cfg.StockfishPath
	if err := session.Start(cmd); err != nil {
		return ports.AnalyzeResult{}, err
	}

	depth := a.cfg.AnalysisDepth
	commands := []string{
		"uci",
		"isready",
		"ucinewgame",
		"isready",
		posCmd,
		fmt.Sprintf("go depth %d", depth),
	}

	// Send commands and wait for bestmove
	go func() {
		for _, c := range commands {
			fmt.Fprintln(stdinPipe, c)
		}
		// Wait for bestmove to appear in output before quitting
		for i := 0; i < 300; i++ { // 30 seconds max
			if strings.Contains(stdout.String(), "bestmove ") {
				break
			}
			select {
			case <-ctx.Done():
				return
			case <-time.After(100 * time.Millisecond):
			}
		}
		fmt.Fprintln(stdinPipe, "quit")
		stdinPipe.Close()
	}()

	if err := session.Wait(); err != nil {
		// Ignore exit errors when we sent quit
		if !strings.Contains(stdout.String(), "bestmove ") {
			return ports.AnalyzeResult{}, err
		}
	}

	output := stdout.String()
	bestMove := parseBestMove(output)
	info := parseEngineInfo(output)

	var bestMoveSAN string
	if pos != nil && bestMove != "" {
		bestMoveSAN = uciToSAN(bestMove, pos)
	}

	result := ports.AnalyzeResult{
		BestMoveUCI: bestMove,
		BestMoveSAN: bestMoveSAN,
		Depth:       info.Depth,
		Nodes:       info.Nodes,
		NPS:         info.NPS,
		PV:          info.PV,
	}
	if a.cfg.IncludeRaw {
		result.Raw = output
	}
	if pos != nil {
		result.PositionFEN = pos.String()
	}

	// Stockfish reports score from side-to-move perspective
	isWhiteToMove := pos == nil || pos.Turn() == chess.White

	if info.EvalCp != nil {
		cp := *info.EvalCp
		if !isWhiteToMove {
			cp = -cp // Flip for White's perspective
		}
		result.EvaluationCp = &cp
	}
	if info.EvalMate != nil {
		mate := *info.EvalMate
		if !isWhiteToMove {
			mate = -mate // Flip for White's perspective
		}
		result.EvaluationMate = &mate
	}
	if bar := computeEvalBar(result.EvaluationCp, result.EvaluationMate); bar != nil {
		result.EvalBar = bar
	}

	return result, nil
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

func buildPositionCommand(req ports.AnalyzeRequest) (string, *chess.Position, error) {
	if req.FEN == "" && req.PGN == "" && req.UCIMoves == "" && req.SANMoves == "" {
		return "", nil, errors.New("fen, pgn, uci or san required")
	}

	var pos *chess.Position
	var baseCmd string

	if req.PGN != "" {
		pgnText := sanitizePGN(req.PGN)
		pgn, err := chess.PGN(strings.NewReader(pgnText))
		if err != nil {
			pgn, err = chess.PGN(strings.NewReader(req.PGN))
			if err != nil {
				return "", nil, err
			}
		}
		game := chess.NewGame(pgn)
		pos = game.Position()
		return "position fen " + pos.String(), pos, nil
	}

	baseCmd = "position startpos"
	if req.FEN != "" {
		fenOpt, err := chess.FEN(req.FEN)
		if err != nil {
			return "", nil, err
		}
		game := chess.NewGame(fenOpt)
		pos = game.Position()
		baseCmd = "position fen " + req.FEN
	} else {
		game := chess.NewGame()
		pos = game.Position()
	}

	if strings.TrimSpace(req.SANMoves) != "" {
		uciMoves, err := sanToUciMoves(req.SANMoves, req.FEN)
		if err != nil {
			return "", nil, err
		}
		if pos != nil {
			pos, err = applyUCIMoves(pos, uciMoves)
			if err != nil {
				return "", nil, err
			}
		}
		return baseCmd + " moves " + uciMoves, pos, nil
	}

	if strings.TrimSpace(req.UCIMoves) != "" {
		uciMoves := strings.TrimSpace(req.UCIMoves)
		if pos != nil {
			updatedPos, applyErr := applyUCIMoves(pos, uciMoves)
			if applyErr != nil {
				return "", nil, applyErr
			}
			pos = updatedPos
		}
		return baseCmd + " moves " + uciMoves, pos, nil
	}

	return baseCmd, pos, nil
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

func applyUCIMoves(pos *chess.Position, uciMoves string) (*chess.Position, error) {
	uci := chess.UCINotation{}
	current := pos
	for _, token := range strings.Fields(uciMoves) {
		move, err := uci.Decode(current, token)
		if err != nil {
			return nil, err
		}
		current = current.Update(move)
	}
	return current, nil
}

func uciToSAN(uciMove string, pos *chess.Position) string {
	if uciMove == "" || pos == nil {
		return ""
	}
	uci := chess.UCINotation{}
	alg := chess.AlgebraicNotation{}
	move, err := uci.Decode(pos, uciMove)
	if err != nil {
		return ""
	}
	return alg.Encode(pos, move)
}

type engineInfo struct {
	Depth    int
	Nodes    int
	NPS      int
	EvalCp   *int
	EvalMate *int
	PV       string
}

func parseEngineInfo(output string) engineInfo {
	var info engineInfo
	lines := strings.Split(output, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if strings.HasPrefix(line, "info ") {
			if strings.Contains(line, " score ") {
				info = parseInfoLine(line)
				if info.PV != "" || info.EvalCp != nil || info.EvalMate != nil {
					return info
				}
			}
		}
	}
	return info
}

func parseInfoLine(line string) engineInfo {
	var info engineInfo
	fields := strings.Fields(line)
	for i := 0; i < len(fields); i++ {
		switch fields[i] {
		case "depth":
			if i+1 < len(fields) {
				if v, err := strconv.Atoi(fields[i+1]); err == nil {
					info.Depth = v
				}
			}
		case "nodes":
			if i+1 < len(fields) {
				if v, err := strconv.Atoi(fields[i+1]); err == nil {
					info.Nodes = v
				}
			}
		case "nps":
			if i+1 < len(fields) {
				if v, err := strconv.Atoi(fields[i+1]); err == nil {
					info.NPS = v
				}
			}
		case "score":
			if i+2 < len(fields) {
				scoreType := fields[i+1]
				if v, err := strconv.Atoi(fields[i+2]); err == nil {
					switch scoreType {
					case "cp":
						info.EvalCp = &v
					case "mate":
						info.EvalMate = &v
					}
				}
			}
		case "pv":
			if i+1 < len(fields) {
				info.PV = strings.Join(fields[i+1:], " ")
				return info
			}
		}
	}
	return info
}

func sanitizePGN(pgn string) string {
	text := pgn
	text = regexp.MustCompile(`\[[^\]]*\]`).ReplaceAllString(text, " ")
	text = regexp.MustCompile(`\{[^}]*\}`).ReplaceAllString(text, " ")
	text = regexp.MustCompile(`\([^)]*\)`).ReplaceAllString(text, " ")
	text = regexp.MustCompile(`\$\d+`).ReplaceAllString(text, " ")
	text = strings.ReplaceAll(text, "\r", "\n")
	text = strings.TrimSpace(text)
	return text
}

func computeEvalBar(cp *int, mate *int) *int {
	if mate != nil {
		if *mate > 0 {
			v := 100
			return &v
		}
		if *mate < 0 {
			v := 0
			return &v
		}
		v := 50
		return &v
	}
	if cp == nil {
		return nil
	}
	value := 50 + 50*math.Tanh(float64(*cp)/400.0)
	bar := int(math.Round(value))
	if bar < 0 {
		bar = 0
	}
	if bar > 100 {
		bar = 100
	}
	return &bar
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
