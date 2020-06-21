FROM golang AS builder
LABEL maintainer="PayU SRE Clan"

# Set necessary environmet variables needed for our image
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

RUN apt-get update && \
    apt-get install -y --no-install-recommends build-essential && \
    apt-get clean && \
    mkdir -p "$GOPATH/src/github.com/shyimo/kubeobserver"

COPY . "$GOPATH/src/github.com/shyimo/kubeobserver"

RUN cd "$GOPATH/src/github.com/shyimo/kubeobserver" && \
    make build

RUN cp $GOPATH/src/github.com/shyimo/kubeobserver/kubeobserver .

FROM debian
COPY --from=builder go/kubeobserver /kubeobserver

RUN apt-get update && \
    apt-get -y install curl

# Install aws-iam-authenticator in order kubectl will be able to work with Amazon EKS
RUN curl -o aws-iam-authenticator https://amazon-eks.s3.us-west-2.amazonaws.com/1.16.8/2020-04-16/bin/linux/amd64/aws-iam-authenticator && \
    chmod +x ./aws-iam-authenticator && \
    cp ./aws-iam-authenticator /usr/local/bin

ENTRYPOINT ["/kubeobserver"]