FROM golang AS builder
LABEL maintainer="PayU SRE Clan"

RUN apt-get update && \
    apt-get install -y --no-install-recommends build-essential && \
    apt-get clean && \
    mkdir -p "$GOPATH/src/github.com/shyimo/kubeobserver"

COPY . "$GOPATH/src/github.com/shyimo/kubeobserver"

RUN cd "$GOPATH/src/github.com/shyimo/kubeobserver" && \
    make build

RUN cp $GOPATH/src/github.com/shyimo/kubeobserver/kubeobserver .

FROM scratch
COPY --from=builder /kubeobserver /kubeobserver

ENTRYPOINT ["./kubeobserver"]