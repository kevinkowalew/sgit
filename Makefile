build:
	go build -o ~/go/bin/sgit main.go
lint:
	 golangci-lint run
