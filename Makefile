test:
	go test ./... -v

build:
	go build github.com/PayU/kubeobserver/cmd/kubeobserver

format:
	go fmt github.com/PayU/kubeobserver/cmd/kubeobserver

docker-build-and-push:
	echo "$(DOCKER_PASSWORD)" | docker login -u "$(DOCKER_USENAME)" --password-stdin
	docker build -t $(DOCKER_IMAGE) .
	docker push $(DOCKER_IMAGE)