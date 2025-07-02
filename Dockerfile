FROM golang:1.24.4 as builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .

ARG BUILD_ENV=local
RUN if [ "$BUILD_ENV" = "local" ] || [ "$BUILD_ENV" = "dev" ]; then \
      go install github.com/swaggo/swag/v2/cmd/swag@latest && \
      swag init --generalInfo ./cmd/app/main.go --output ./docs; \
    fi

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/server ./cmd/app

FROM gcr.io/distroless/static-debian12

WORKDIR /app
COPY --from=builder /app/bin/server /app/server
COPY --from=builder /src/docs /app/docs
CMD ["/app/server"]