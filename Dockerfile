FROM golang:alpine3.23 AS builder

WORKDIR /src

# Cache dependencies first to speed up rebuilds.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/app-server ./main.go

FROM alpine:3.20

RUN apk add --no-cache ca-certificates \
	&& addgroup -S app \
	&& adduser -S -G app app

WORKDIR /app

COPY --from=builder /out/app-server /app/app-server
COPY static /app/static
COPY blogs.json /app/blogs.json

RUN chown -R app:app /app

USER app

EXPOSE 8080

ENTRYPOINT ["/app/app-server"]