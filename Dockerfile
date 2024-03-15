FROM golang:1.22.1 AS builder

WORKDIR /src

COPY . .

RUN go mod download && go mod verify

RUN go build -o srv cmd/server/main.go

FROM alpine:latest AS runner

COPY --from=builder /src/srv .

EXPOSE 5050

ENTRYPOINT [ "./srv" ]
