FROM golang:1.23 as builder
ARG CGO_ENABLED=0
WORKDIR /app

COPY vendor ./
COPY go.mod go.sum ./
COPY . .

RUN go build -o server ./cmd/server/main.go
RUN go build -o client ./cmd/client/main.go

FROM scratch

COPY --from=builder /app/server /server
COPY --from=builder /app/client /client

ENTRYPOINT ["/server"]
