FROM golang:1.26.0-alpine3.23 as build

ARG version

RUN if [ -z "$version" ]; then \
			echo "version is not set"; \
			exit 1; \
    fi

RUN apk update && \
		apk add git

WORKDIR /go/src

COPY go.mod go.sum ./
RUN go mod download && \
		go mod verify

COPY . .

RUN go build  \
    -o ../bin/livestream-snapshot-tool \
    -ldflags "-X github.com/matthiasharzer/livestream-snapshot-tool/cmd/version.version=$version"  \
    ./main.go

FROM alpine:3.23

COPY --from=build /go/bin/livestream-snapshot-tool /usr/local/bin/livestream-snapshot-tool

WORKDIR /var/lib/livestream-snapshot-tool

EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/livestream-snapshot-tool"]

