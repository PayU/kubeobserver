test:
	go test ./... -coverprofile c.out -v
	go tool cover -func=c.out
	go tool cover -html=c.out

build:
	go build github.com/PayU/kubeobserver/cmd/kubeobserver

format:
	go fmt github.com/PayU/kubeobserver/cmd/kubeobserver

docker-build-and-push:
	echo "$(DOCKER_PASSWORD)" | docker login -u "$(DOCKER_USENAME)" --password-stdin
	docker build -t $(DOCKER_IMAGE):$(VERSION) .
	docker push $(DOCKER_IMAGE):$(VERSION)
	docker tag $(DOCKER_IMAGE):$(VERSION) $(DOCKER_IMAGE):latest
	docker push $(DOCKER_IMAGE):latest