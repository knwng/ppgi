build:
	go build -o ./cmd/ppgi ./cmd/main.go

run:
	go run ./cmd/main.go

cross_build:
	GOOS=linux GOARCH=arm64 go build ./cmd/ppgi ./cmd/main.go
