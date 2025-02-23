FROM us-central1-docker.pkg.dev/regions-spy/containers/backend-builder:latest AS builder

WORKDIR /app

COPY ./go.mod ./go.sum ./
RUN --mount=type=secret,id=netrc,dst=/root/.netrc go mod download

COPY .. .

RUN --mount=type=secret,id=netrc,dst=/root/.netrc go mod download
RUN ls -l
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o auth ./cmd/main.go

FROM alpine:3.21 AS runner

WORKDIR /app
COPY --from=builder /app/auth .

EXPOSE 8000

ENTRYPOINT ["./auth"]
