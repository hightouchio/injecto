FROM golang:1.14 AS builder
ARG version

WORKDIR /app
COPY ./ ./

RUN GOOS=linux CGO_ENABLED=0 go build \
  -mod vendor \
  -ldflags "-X main.version=$version" \
  -o ./injecto

FROM scratch
COPY --from=builder /app/injecto /bin/injecto
ENTRYPOINT ["/bin/injecto"]
