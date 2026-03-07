FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git

WORKDIR /src
COPY go.mod ./
COPY go.sum* ./
RUN go mod download 2>/dev/null || true
COPY . .

RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /bin/gangway .

FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /bin/gangway /usr/local/bin/gangway

ENTRYPOINT ["gangway"]
