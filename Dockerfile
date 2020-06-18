FROM golang AS builder
LABEL maintainer="PayU SRE Clan"

RUN apt-get update && \
    apt-get install -y --no-install-recommends build-essential && \
    apt-get clean && \
    mkdir -p "$GOPATH/src/github.com/shyimo/kubeobserver"

ADD . "$GOPATH/src/github.com/shyimo/kubeobserver"

RUN make build

FROM scratch
COPY --from=builder /kubeobserver /bin/kubeobserver

ENTRYPOINT ["./bin/kubeobserver"]