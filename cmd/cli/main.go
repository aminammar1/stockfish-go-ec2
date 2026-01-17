package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {
	baseURL := flag.String("base", "http://localhost:8080", "base URL of service")
	cmd := flag.String("cmd", "", "command: health|analyze (leave empty for interactive)")
	fen := flag.String("fen", "", "FEN position")
	pgn := flag.String("pgn", "", "PGN game")
	uci := flag.String("uci", "", "UCI move list (space-separated)")
	san := flag.String("san", "", "SAN move list (space-separated)")
	flag.Parse()

	if strings.TrimSpace(*cmd) == "" {
		runInteractive(*baseURL)
		return
	}

	switch *cmd {
	case "health":
		req, _ := http.NewRequest(http.MethodGet, *baseURL+"/api/v1/health", nil)
		do(req)
	case "analyze":
		payload := map[string]string{"fen": *fen, "pgn": *pgn, "uci": *uci, "san": *san}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest(http.MethodPost, *baseURL+"/api/v1/analyze", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		doWithSpinner(req, "Analyzing")
	default:
		fmt.Fprintln(os.Stderr, "unknown cmd")
		os.Exit(1)
	}
}

func runInteractive(baseURL string) {
	scanner := bufio.NewScanner(os.Stdin)
	printBanner()

	for {
		fmt.Println("\nChoose action:")
		fmt.Println("  [1]  Health check")
		fmt.Println("  [2]  Analyze position")
		fmt.Println("  [3]  Exit")
		fmt.Print("Select [1-3]: ")
		if !scanner.Scan() {
			return
		}
		choice := strings.TrimSpace(scanner.Text())

		switch choice {
		case "1":
			req, _ := http.NewRequest(http.MethodGet, baseURL+"/api/v1/health", nil)
			do(req)
		case "2":
			payload := map[string]string{}
			fmt.Println("\nSelect input type:")
			fmt.Println("  [1]  FEN")
			fmt.Println("  [2]  PGN")
			fmt.Println("  [3]  UCI moves")
			fmt.Println("  [4]  SAN moves")
			fmt.Print("Select [1-4]: ")
			kind := readLine(scanner)

			switch kind {
			case "1":
				fmt.Print("FEN: ")
				payload["fen"] = readLine(scanner)
			case "2":
				fmt.Println("PGN (single line): ")
				payload["pgn"] = readLine(scanner)
			case "3":
				fmt.Print("UCI moves (space-separated): ")
				payload["uci"] = readLine(scanner)
			case "4":
				fmt.Print("SAN moves (space-separated): ")
				payload["san"] = readLine(scanner)
			default:
				fmt.Println("Invalid selection")
				continue
			}

			body, _ := json.Marshal(payload)
			req, _ := http.NewRequest(http.MethodPost, baseURL+"/api/v1/analyze", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			doWithSpinner(req, "Analyzing")
		case "3":
			return
		default:
			fmt.Println("Invalid selection")
		}
	}
}

func do(req *http.Request) {
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	fmt.Println(string(data))
}

func doWithSpinner(req *http.Request, label string) {
	done := make(chan struct{})
	go func() {
		frames := []string{"-", "\\", "|", "/"}
		idx := 0
		for {
			select {
			case <-done:
				fmt.Print("\r                    \r")
				return
			default:
				fmt.Printf("\r[%s] %s...", frames[idx%len(frames)], label)
				idx++
				time.Sleep(120 * time.Millisecond)
			}
		}
	}()

	resp, err := http.DefaultClient.Do(req)
	close(done)
	if err != nil {
		fmt.Println()
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	fmt.Println()
	printAnalysisResult(data)
}

func printAnalysisResult(data []byte) {
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		fmt.Println(string(data))
		return
	}

	// Check for error
	if errMsg, ok := result["error"]; ok {
		red := "\033[31m"
		reset := "\033[0m"
		fmt.Printf("%sError: %v%s\n", red, errMsg, reset)
		return
	}

	green := "\033[32m"
	yellow := "\033[33m"
	cyan := "\033[36m"
	reset := "\033[0m"

	bestUCI := getString(result, "bestMoveUci")
	bestSAN := getString(result, "bestMoveSan")
	evalBar := getInt(result, "evalBar")

	fmt.Println()
	fmt.Printf("%sBest Move:%s  %s (%s)\n", cyan, reset, bestSAN, bestUCI)
	fmt.Println()
	fmt.Printf("%sEval Bar:%s\n", cyan, reset)
	printEvalBar(evalBar, green, yellow, reset)
	fmt.Println()
}

func printEvalBar(bar int, _, _, reset string) {
	white := "\033[47m"
	black := "\033[40m"

	whiteLen := bar / 2
	blackLen := 50 - whiteLen

	fmt.Printf("  White %3d%% ", bar)
	fmt.Print(white)
	fmt.Print(strings.Repeat(" ", whiteLen))
	fmt.Print(reset)
	fmt.Print(black)
	fmt.Print(strings.Repeat(" ", blackLen))
	fmt.Print(reset)
	fmt.Printf(" %3d%% Black\n", 100-bar)
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func getInt(m map[string]interface{}, key string) int {
	if v, ok := m[key]; ok {
		switch n := v.(type) {
		case float64:
			return int(n)
		case int:
			return n
		}
	}
	return 50
}

func readLine(scanner *bufio.Scanner) string {
	if !scanner.Scan() {
		return ""
	}
	return strings.TrimSpace(scanner.Text())
}

func printBanner() {
	cyan := "\033[36m"
	reset := "\033[0m"
	banner := `
  ____  _             _   __ _      _           _____ ____ ____
 / ___|| |_ ___   ___| | / _(_)___| |__       | ____/ ___|___ \
 \___ \| __/ _ \ / __| |/ _| | / __| '_ \ _____| _|| |     __) |
  ___) | || (_) | (__| |  _| | \__ \ | | |_____|___| |___ / __/
 |____/ \__\___/ \___|_|_| |_| |___/_| |_|     |_____\____|_____|
                       STOCKFISH-EC2
`
	fmt.Println(cyan + banner + reset)
}
