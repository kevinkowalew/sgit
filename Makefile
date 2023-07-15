mockgen:
	mockgen -source=internal/intefaces/interfaces.go -destination=internal/intefaces/mocks/mocks.go
build:
	go build -o ~/go/bin/sgit main.go
