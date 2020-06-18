test:
	go test ./... -v -timeout 5s

build:
	go build github.com/shyimo/kubeobserver/cmd/kubeobserver

format:
	go fmt github.com/shyimo/kubeobserver/cmd/kubeobserver