build:
	go build cmd/grafana-interacter.go

install:
	go install cmd/grafana-interacter.go

lint:
	golangci-lint run --fix ./...
