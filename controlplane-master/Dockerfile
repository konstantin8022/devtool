FROM golang:1.13.5-alpine as build-env

ENV GO111MODULE=on
WORKDIR /build

# download mods
COPY go.mod .
COPY go.sum .
RUN go mod download

#build binary
COPY . .
RUN go build -v -o /controlplane

FROM alpine:latest
COPY --from=build-env /controlplane /controlplane

WORKDIR /
ENTRYPOINT [ "/controlplane", "-logtostderr" ]
