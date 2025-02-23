linux-build:
	env GOOS=linux GOARCH=arm go build -o dp -v cmd/display-board/main.go

build:
	go build -o dp -v cmd/display-board/main.go

run:
	go run cmd/display-board/main.go


