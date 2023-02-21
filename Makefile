run:
	go build && ./goupbox

compile_all: compile_linux compile_win compile_macos

compile_linux:
	GOOS=linux GOARCH=amd64 go build -o bin/goupbox-linux-amd64

compile_win:
	GOOS=windows GOARCH=amd64 go build -o bin/goupbox-win-amd64.exe

compile_macos:
	GOOS=darwin GOARCH=amd64 go build -o bin/goupbox-macos-amd64
