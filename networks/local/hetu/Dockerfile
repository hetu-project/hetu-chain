FROM golang:1.23.4-bullseye as build-env

# Install minimum necessary dependencies
ENV PACKAGES curl make git libc-dev bash gcc
RUN apt-get update && apt-get upgrade -y && \
    apt-get install -y $PACKAGES

# Set working directory for the build
WORKDIR /go/src/github.com/hetu/hetu

# Add source files
COPY . .

# build Ethermint
RUN make build-linux

# Final image
FROM golang:1.23.4-bullseye as final

WORKDIR /

RUN apt-get update

# Copy over binaries from the build-env
COPY --from=build-env /go/src/github.com/hetu/hetu/build/hetud /
COPY --from=build-env /go/src/github.com/hetu/hetu/scripts/start-docker.sh /

EXPOSE 26656 26657 1317 9090 8545 8546

# Run hetud by default, omit entrypoint to ease using container with hetud
ENTRYPOINT ["/bin/bash", "-c"]