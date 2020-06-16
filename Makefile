run:
	go test ./... -v

build:
	go build github.com/shyimo/kubeobserver/cmd/kubeobserver

format:
	go fmt github.com/shyimo/kubeobserver/cmd/kubeobserver