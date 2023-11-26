# Stage 1
FROM golang:1.21.1-alpine3.18 AS builder

RUN apk update \
  && apk -U upgrade \
  && apk --no-cache add ca-certificates tzdata gcc bash musl-dev \
  && update-ca-certificates --fresh \
  && rm -rf /var/cache/apk/*

ENV CGO_ENABLED=1
ENV GO111MODULE=on
ENV GOOS=linux
ENV GOARCH=amd64

# Set the Current Working Directory inside the container
WORKDIR /src

# We want to populate the module cache based on the go.{mod,sum} files.
COPY cmd/go.mod .
COPY cmd/go.sum .

RUN go mod download

COPY cmd/netpol_generator/ ./

RUN CGO_ENABLED=$CGO_ENABLED GOOS=$GOOS GOARCH=$GOARCH go build -ldflags='-s' -C netpol_generator -o ./

# Stage 2
FROM alpine:3.18 as runner

RUN apk update \
  && apk -U upgrade \
  && apk --no-cache add ca-certificates tzdata \
  && update-ca-certificates --fresh \
  && rm -rf /var/cache/apk/*

RUN addgroup runtime && adduser -S runtime -u 1000 -G runtime

WORKDIR /app

COPY --chown=runtime:runtime --from=builder /src/netpol_generator /app/netpol_generator

RUN chmod +x netpol_generator

# This container exposes port 9090 to the outside world
EXPOSE 9090

# Run the binary program produced by `go install`
ENTRYPOINT ["./netpol_generator"]