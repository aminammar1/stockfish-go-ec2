# Stockfish-EC2-Service

A hexagonal Go/Gin API that connects to a remote EC2 Stockfish engine via SSH to provide health checks and chess analysis from FEN/PGN/UCI/SAN.

## Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                          stockfish-ec2-service                              │
│                                                                             │
│  ┌──────────┐    ┌─────────────────────────────┐    ┌───────────────────┐  │
│  │          │    │            CORE             │    │                   │  │
│  │ HTTP API │───>│   (Application + Domain)   │───>│ EC2 SSH Connector │  │
│  │          │    │                             │    │                   │  │
│  └──────────┘    │  - Health Check             │    └─────────┬─────────┘  │
│       ^         │  - Analyze (FEN/PGN/UCI/SAN)│              │            │
│       │          └─────────────────────────────┘              │            │
└───────│───────────────────────────────────────────────────────│────────────┘
        │                                                       │
        │                                                       v
   ┌────┴────┐                                         ┌────────────────┐
   │ Clients │                                         │  EC2 Instance  │
   │         │                                         │  (Stockfish)   │
   └─────────┘                                         └────────────────┘
```

## Folder Structure

```
stockfish-ec2-service/
├── cmd/
│   ├── server/
│   │   └── main.go
│   └── cli/
│       └── main.go
├── internal/
│   ├── adapters/
│   │   └── stockfish_ssh/
│   │       ├── adapter.go
│   │       └── adapter_test.go
│   ├── app/
│   │   └── service.go
│   ├── config/
│   │   └── config.go
│   ├── http/
│   │   └── handler.go
│   └── ports/
│       └── stockfish.go
├── tests/
│   └── performance/
│       └── load_test.js
├── .github/
│   └── workflows/
│       └── notify.yml
├── .env.example
├── Dockerfile
├── Jenkinsfile
├── Makefile
├── go.mod
├── go.sum
└── README.md
```

## Requirements

- Go 1.25.5+
- Docker (optional)
- EC2 instance with Stockfish installed
- SSH access to EC2 (RSA key)

## Environment Variables

Create a `.env` file based on `.env.example`:

| Variable | Description | Example |
|----------|-------------|---------|
| `SSH_HOST` | EC2 public DNS or IP | `ec2-xx-xx-xx-xx.compute.amazonaws.com` |
| `SSH_PORT` | SSH port | `22` |
| `SSH_USER` | SSH username | `ubuntu` |
| `SSH_PRIVATE_KEY` | Path to RSA private key | `/home/user/keys/stockfish.pem` |
| `STOCKFISH_PATH` | Stockfish binary path on EC2 | `/usr/local/bin/stockfish` |
| `ANALYSIS_DEPTH` | Default analysis depth | `20` |
| `SERVER_PORT` | HTTP server port | `8080` |

## Quick Start

### 1. Clone and Setup

```bash
git clone https://github.com/aminammar1/stockfish-go-ec2.git
cd stockfish-go-ec2
cp .env.example .env
```

### 2. Run Server

```bash
make run
```

### 3. Run Interactive CLI

```bash
make cli-run
```

## Makefile Commands

| Command | Description |
|---------|-------------|
| `make tidy` | Clean up go.mod and go.sum |
| `make build` | Build server binary |
| `make run` | Run the server |
| `make cli-build` | Build CLI binary |
| `make cli-run` | Run interactive CLI |
| `make swag` | Generate Swagger docs |

## API Endpoints

### Health Check

```bash
GET /api/v1/health
```

Response:
```json
{
  "status": "ok"
}
```

### Analyze Position

```bash
POST /api/v1/analyze
Content-Type: application/json
```

With FEN:
```json
{
  "fen": "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq - 0 1"
}
```

With UCI moves:
```json
{
  "uci": "e2e4 e7e5 g1f3"
}
```

With SAN moves:
```json
{
  "san": "e4 e5 Nf3"
}
```

With PGN:
```json
{
  "pgn": "1. e4 e5 2. Nf3 Nc6"
}
```

Response:
```json
{
  "bestMoveUci": "d7d5",
  "bestMoveSan": "d5",
  "evaluationCp": 32,
  "evalBar": 54,
  "depth": 20,
  "nodes": 1234567,
  "nps": 2345678,
  "pv": "d7d5 e4d5 d8d5",
  "positionFen": "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq - 0 1"
}
```

## Interactive CLI

```
  ____  _             _   __ _      _           _____ ____ ____
 / ___|| |_ ___   ___| | / _(_)___| |__       | ____/ ___|___ \
 \___ \| __/ _ \ / __| |/ _| | / __| '_ \ _____| _|| |     __) |
  ___) | || (_) | (__| |  _| | \__ \ | | |_____|___| |___ / __/
 |____/ \__\___/ \___|_|_| |_| |___/_| |_|     |_____\____|_____|
                       STOCKFISH-EC2

Choose action:
  [1]  Health check
  [2]  Analyze position
  [3]  Exit
```

Analysis output:
```
Best Move:  d5 (d7d5)

Eval Bar:
  White  54%                                                    46% Black
```

## Docker

### Build

```bash
docker build -t stockfish-ec2-service .
```

### Run

```bash
docker run -p 8080:8080 \
  -e SSH_HOST=your-ec2-host \
  -e SSH_USER=ubuntu \
  -e SSH_PRIVATE_KEY="$(cat /path/to/key.pem)" \
  -e STOCKFISH_PATH=/usr/local/bin/stockfish \
  stockfish-ec2-service
```

## CI/CD

### Jenkins Pipeline

The Jenkinsfile includes:
- Unit tests
- k6 performance tests
- Docker build and push to GHCR
- Deploy to Render

### Required Jenkins Credentials

| ID | Kind | Description |
|----|------|-------------|
| `ghcr-token` | Secret text | GitHub PAT with `write:packages` |
| `render-api-token` | Secret text | Render API key |
| `render-service-id` | Secret text | Render service ID |

### GitHub Actions

Notifications on fork and pull request events.

## EC2 Setup

Ensure your EC2 instance has:

1. Stockfish installed at the configured path
2. SSH access enabled (port 22)
3. Security group allows inbound SSH

Install Stockfish on EC2:
```bash
sudo apt update
sudo apt install stockfish -y
```

Or download latest:
```bash
wget https://github.com/official-stockfish/Stockfish/releases/latest/download/stockfish-ubuntu-x86-64.tar
tar -xf stockfish-ubuntu-x86-64.tar
sudo mv stockfish/stockfish-ubuntu-x86-64 /usr/local/bin/stockfish
```

## License

MIT
