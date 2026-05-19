# syntax=docker/dockerfile:1.7

FROM golang:1.26-alpine AS build
WORKDIR /src

# Cache modules separately from sources.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG VERSION=dev
ENV CGO_ENABLED=0 GOFLAGS=-trimpath
RUN go build -ldflags="-s -w" -o /out/deep-search-mcp ./cmd/deep-search-mcp

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/deep-search-mcp /usr/local/bin/deep-search-mcp

EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/usr/local/bin/deep-search-mcp", "--transport=http", "--listen=:8080"]
