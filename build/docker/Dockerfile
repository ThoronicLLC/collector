# syntax=docker/dockerfile:1

# Accept the Go version for the image to be set as a build argument.
# Default to Go 1.21
ARG GO_VERSION=1.21

# Build stage: Build the binary
FROM golang:${GO_VERSION}-alpine as golang

# Install git.
# Git is required for fetching the dependencies.
RUN apk update && apk add --no-cache git make zip bash

# Make empty directory for permission copy
RUN mkdir /empty

# Set the working directory outside $GOPATH to enable the support for modules.
WORKDIR /src

# Copy Go Module config
COPY go.mod .
COPY go.sum .

# Download Go Modules
RUN go mod download
RUN go mod verify

# Import the code from the context.
COPY . .

# Build app
RUN make linux

# Set permissions on app
RUN chmod +x /src/bin/linux/collector

# Final stage: Bare minimum run container.
FROM gcr.io/distroless/static as final

# Copy binary
COPY --chown=nonroot:nonroot --from=golang /src/bin/linux/collector /collector

# Create some useful empty directories
COPY --chown=nonroot:nonroot --from=golang /empty /output
COPY --chown=nonroot:nonroot --from=golang /empty /secrets

# Copy temp directory
COPY --chown=nonroot:nonroot --from=golang /empty /tmp

# Copy etc directory
COPY --chown=nonroot:nonroot --from=golang /empty /etc

# Copy collector config directory
COPY --chown=nonroot:nonroot --from=golang /empty /etc/collector

# Image cache COPY bugfix
LABEL maintainer="Coleton Pierson <coleton.pierson@thoronic.com>"

# Perform any further action as an unprivileged user.
USER nonroot:nonroot

CMD ["/collector", "serve"]