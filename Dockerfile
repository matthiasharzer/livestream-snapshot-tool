FROM golang:1.26.2-alpine3.23 AS build

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
    -o ../bin/livestream-snapshotting-tool \
    -ldflags "-X github.com/matthiasharzer/livestream-snapshotting-tool/cmd/version.version=$version"  \
    ./main.go

FROM alpine:3.23

RUN apk update && \
		apk add --no-cache ffmpeg yt-dlp

COPY --from=build /go/bin/livestream-snapshotting-tool /usr/local/bin/livestream-snapshotting-tool

WORKDIR /var/lib/livestream-snapshotting-tool

EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/livestream-snapshotting-tool"]

