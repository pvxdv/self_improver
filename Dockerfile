FROM golang:1.24.4 as builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .

ARG BUILD_ENV=local
RUN if [ "$BUILD_ENV" != "prod" ]; then \
      go install github.com/swaggo/swag/v2/cmd/swag@latest && \
      swag init --generalInfo ./cmd/app/main.go --output ./docs; \
    else \
      mkdir -p /src/docs; \
    fi

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server ./cmd/app

FROM golang:1.24.4 as goose-builder
WORKDIR /go
ENV CGO_ENABLED=0
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

FROM gcr.io/distroless/static-debian12
#FROM alpine
WORKDIR /app

COPY --from=builder /app/server .
COPY --from=builder /src/docs /app/docs

COPY --from=goose-builder /go/bin/goose /usr/local/bin/goose
COPY migrations /app/migrations

CMD ["/app/server"]