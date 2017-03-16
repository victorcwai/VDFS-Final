# Start from a Debian image with the latest version of Go installed
# and a workspace (GOPATH) configured at /go.
FROM golang

# Copy the local package files to the container's workspace.
ADD . /go/src/wai/volumeServer

# Build the outyet command inside the container.
# (You may fetch or manage dependencies here,
# either manually or with a tool like "godep".)
RUN go install wai/volumeServer

# Run the outyet command by default when the container starts.
ENTRYPOINT /go/bin/volumeServer

# Document that the service listens on port 2000.
EXPOSE 2001
