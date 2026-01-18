ARG GO_VERSION=1.25.5
ARG ALPINE_VERSION=3.23

FROM golang:${GO_VERSION}-alpine${ALPINE_VERSION} AS builder
WORKDIR /app
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download
COPY . ./

# Generate swagger docs during build 
RUN go install github.com/swaggo/swag/cmd/swag@v1.16.2 \
  && /go/bin/swag init -g cmd/server/main.go -o docs

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/bin/stockfish-ec2-service ./cmd/server

FROM alpine:${ALPINE_VERSION}
WORKDIR /app
RUN adduser -D -g '' appuser
COPY --from=builder /app/bin/stockfish-ec2-service /app/stockfish-ec2-service
USER appuser
EXPOSE 8080
ENTRYPOINT ["/app/stockfish-ec2-service"]
