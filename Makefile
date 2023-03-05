run:
	go build && ./goupbox

compile_all: compile_linux compile_win compile_macos

test_linux:
	GOOS=linux GOARCH=amd64 go test -v

compile_linux:
	GOOS=linux GOARCH=amd64 go build -v -o bin/goupbox-linux-amd64

test_win:
	GOOS=windows GOARCH=amd64 go test -v

compile_win:
	GOOS=windows GOARCH=amd64 go build -v -o bin/goupbox-win-amd64.exe

test_macos:
	GOOS=darwin GOARCH=amd64 go test -v

compile_macos:
	GOOS=darwin GOARCH=amd64 go build -o bin/goupbox-macos-amd64
