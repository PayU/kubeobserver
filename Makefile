test:
	go test ./... -v

build:
	go build github.com/PayU/kubeobserver/cmd/kubeobserver

format:
	go fmt github.com/PayU/kubeobserver/cmd/kubeobserver