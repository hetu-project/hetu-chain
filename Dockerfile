FROM golang:1.23.4-bullseye AS build-env

WORKDIR /go/src/github.com/hhub/hhub

RUN apt-get update -y
RUN apt-get install git -y

COPY . .

RUN make build

FROM golang:1.23.4-bullseye

RUN apt-get update -y
RUN apt-get install ca-certificates jq -y

WORKDIR /root

COPY --from=build-env /go/src/github.com/hhub/hhub/build/hhubd /usr/bin/hhubd

EXPOSE 26656 26657 1317 9090 8545 8546

CMD ["hhubd"]
