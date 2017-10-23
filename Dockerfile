# Start from a Debian image with the latest version of Go installed
# and a workspace (GOPATH) configured at /go.
FROM golang:1.8

# Copy the local package files to the container's workspace.
WORKDIR /go/src/github.com/golang/openbaton/go-docker-driver

COPY . .
RUN curl -fsSL -o /usr/local/bin/dep https://github.com/golang/dep/releases/download/v0.3.2/dep-linux-amd64 && chmod +x /usr/local/bin/dep
RUN dep ensure -v

WORKDIR /go/src/github.com/golang/openbaton/go-docker-driver/main
RUN go-wrapper download   # "go get -d -v ./..."
#RUN go-wrapper install    # "go install -v ./..."

# Run the outyet command by default when the container starts.
ENTRYPOINT ["go", "run", "docker-driver.go"]
