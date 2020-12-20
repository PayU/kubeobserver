FROM golang:1.15

# Set necessary environmet variables needed for our image
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# Copy everything from the current directory to the PWD (Present Working Directory) inside the container
COPY ./tests/test-web-server/ .

RUN apt-get update && apt-get install stress

RUN go build main.go

# This container exposes port 8080 to the outside world
EXPOSE 8888

# Run the executable
ENTRYPOINT [ "./main" ]