# This is how we want to name the binary output.
BINARY = childtest1

# These are the values we want to pass for VERSION and BUILD.
VERSION = 1.0
BUILD = `git rev-parse HEAD`

# Setup the -ldflags option for go build here, interpolate the variable values.
LDFLAGS = -ldflags "-X main.Version=${VERSION} -X main.Build=${BUILD} -X main.Identity=${BINARY}"

# Default target.
.DEFAULT_GOAL: $(BINARY)

# Builds the project.
# GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -o "${BINARY}" ./...
$(BINARY):
	go build ${LDFLAGS} -o "${BINARY}" ./...

# Installs our project: copies binaries.
# GOOS=linux GOARCH=amd64 go install ${LDFLAGS} -o "${BINARY}" ./...
install:
	go install ${LDFLAGS} -o "${BINARY}" ./...

# Cleans our project: deletes binaries.
clean:
	rm -f "${BINARY}"

.PHONY: clean install
