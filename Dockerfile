FROM golang:1.26-alpine AS builder

ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_DATE=unknown

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build \
    -ldflags "-s -w -X main.Version=${VERSION} -X main.Commit=${COMMIT} -X main.BuildDate=${BUILD_DATE}" \
    -o /webhook ./cmd/webhook

FROM gcr.io/distroless/static:nonroot
COPY --from=builder /webhook /webhook
USER nonroot:nonroot
ENTRYPOINT ["/webhook"]
