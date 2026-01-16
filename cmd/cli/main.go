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
		do(req)
	default:
		fmt.Fprintln(os.Stderr, "unknown cmd")
		os.Exit(1)
	}
}

func runInteractive(baseURL string) {
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Println("\nChoose action:")
		fmt.Println("  1) Health check")
		fmt.Println("  2) Analyze position")
		fmt.Println("  3) Exit")
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
			fmt.Print("FEN (optional): ")
			payload["fen"] = readLine(scanner)
			fmt.Print("PGN (optional): ")
			payload["pgn"] = readLine(scanner)
			fmt.Print("UCI moves (optional, space-separated): ")
			payload["uci"] = readLine(scanner)
			fmt.Print("SAN moves (optional, space-separated): ")
			payload["san"] = readLine(scanner)

			body, _ := json.Marshal(payload)
			req, _ := http.NewRequest(http.MethodPost, baseURL+"/api/v1/analyze", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			do(req)
		case "3":
			return
		default:
			fmt.Println("Invalid selection")
		}
	}
}

func readLine(scanner *bufio.Scanner) string {
	if !scanner.Scan() {
		return ""
	}
	return strings.TrimSpace(scanner.Text())
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
