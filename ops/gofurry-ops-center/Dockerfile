FROM node:24-alpine AS web-builder

WORKDIR /src/web
COPY web/package*.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

FROM golang:1.26-alpine AS go-builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=web-builder /src/internal/web/dist ./internal/web/dist
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/gofurry-ops-center ./cmd/center

FROM alpine:3.22

RUN adduser -D -H -s /sbin/nologin gofurry \
    && mkdir -p /etc/gofurry-ops-center \
    && chown -R gofurry:gofurry /etc/gofurry-ops-center

COPY --from=go-builder /out/gofurry-ops-center /usr/local/bin/gofurry-ops-center
USER gofurry
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/gofurry-ops-center"]
CMD ["serve", "--config", "/etc/gofurry-ops-center/config.yaml"]
