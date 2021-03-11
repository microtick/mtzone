FROM golang:1.14-alpine AS build-env

# Set up dependencies
ENV PACKAGES curl make git libc-dev bash gcc linux-headers eudev-dev python3

# Set working directory for the build
WORKDIR /go/src/github.com/microtick/mtzone

# Add source files
COPY . .

# Install minimum necessary dependencies, build Cosmos SDK, remove packages
RUN apk add --no-cache $PACKAGES && \
    make install

# Final image
FROM alpine:edge

# Install ca-certificates
RUN apk add --update ca-certificates
WORKDIR /root

# Copy over binaries from the build-env
COPY --from=build-env /go/bin/mtd /usr/bin/mtd
COPY --from=build-env /go/bin/mtcli /usr/bin/mtcli

# Run gaiad by default, omit entrypoint to ease using container with mtcli
CMD ["mtd"]
