.PHONY: compile compile-linux compile-darwin

compile-linux:
	@mkdir -p dist
	env GOOS=linux GOARCH=386 go build -o ./dist/ssm-template-Linux-x86_64 main.go

compile-darwin:
	@mkdir -p dist
	env GOOS=darwin GOARCH=386 go build -o ./dist/ssm-template-Darwin-x86_64 main.go

compile: compile-linux compile-darwin
