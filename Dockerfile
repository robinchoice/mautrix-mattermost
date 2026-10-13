FROM golang:1.27-alpine AS builder

RUN apk add --no-cache build-base

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG COMMIT=unknown
RUN go build -tags goolm -ldflags "-s -w -X main.Commit=$COMMIT" -o /mautrix-mattermost ./cmd/mautrix-mattermost

# Checks only: docker build --target test .
FROM builder AS test
RUN test -z "$(gofmt -l .)" && go vet -tags goolm ./... && go test -race -tags goolm ./...

FROM alpine:3.23

ENV UID=1337 \
    GID=1337

RUN apk add --no-cache su-exec ca-certificates bash yq-go

COPY --from=builder /mautrix-mattermost /usr/bin/mautrix-mattermost
COPY docker-run.sh /docker-run.sh
VOLUME /data
WORKDIR /data

CMD ["/docker-run.sh"]
