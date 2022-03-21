# syntax=docker/dockerfile:1

# NOTE: to avoid error https://github.com/moby/moby/issues/37965 on OCI Linux
# use a RUN true between followed COPYs

# Accept the Go version for the image to be set as a build argument.
# Default to Go 1.18
ARG GO_VERSION=1.18

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
RUN true
COPY go.sum .
RUN true

# Download Go Modules
RUN go mod download
RUN go mod verify

# Import the code from the context.
COPY . .
RUN true

# Build app
RUN make linux

# Set permissions on app
RUN chmod +x /src/bin/linux/collector

# Final stage: Bare minimum run container.
FROM gcr.io/distroless/static as final

# Copy binary
COPY --chown=nonroot:nonroot --from=golang /src/bin/linux/collector /collector
RUN true

# Create some useful empty directories
COPY --chown=nonroot:nonroot --from=golang /empty /output
RUN true
COPY --chown=nonroot:nonroot --from=golang /empty /secrets
RUN true

# Copy temp directory
COPY --chown=nonroot:nonroot --from=golang /empty /tmp
RUN true

# Copy config directory
COPY --chown=nonroot:nonroot --from=golang /empty /etc
RUN true

# Copy collector config directory
COPY --chown=nonroot:nonroot --from=golang /empty /etc/collector
RUN true

# Image cache COPY bugfix
LABEL maintainer="Coleton Pierson <coleton.pierson@thoronic.com>"

# Perform any further action as an unprivileged user.
USER nonroot:nonroot

CMD ["/collector", "serve", "--config", "/etc/collector"]