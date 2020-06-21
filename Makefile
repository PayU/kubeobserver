test:
	go test ./... -v

build:
	go build github.com/PayU/kubeobserver/cmd/kubeobserver

format:
	go fmt github.com/PayU/kubeobserver/cmd/kubeobserver

dockerpush:
	echo "$(DOCKER_PASSWORD)" | docker login -u "$(DOCKER_USENAME)" --password-stdin
	docker tag $(DOCKER_IMAGE_NAME) zooz/$(DOCKER_IMAGE_NAME):latest
	docker push $(DOCKER_IMAGE_NAME):latest