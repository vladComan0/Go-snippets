FROM golang:1.21.2-alpine AS builder

RUN /sbin/apk update && \
	/sbin/apk --no-cache add ca-certificates git tzdata && \
	/usr/sbin/update-ca-certificates

RUN adduser -D -g '' web
WORKDIR /home/web

COPY go.mod     ./
COPY go.sum     ./
COPY cmd        ./cmd/
COPY internal   ./internal/
COPY tls        ./tls/
COPY ui         ./ui/
COPY *.go       ./

ARG VERSION

RUN CGO_ENABLED=0 go build -a -tags netgo,osusergo \
    -ldflags "-extldflags '-static' -s -w" \
    -ldflags "-X main.version=$VERSION" -o web ./cmd/web

FROM busybox:musl

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /home/web/web /home/web

USER web
WORKDIR /home
EXPOSE 8080

ENTRYPOINT ["./web"]
CMD ["-addr=:8080", "-dsn=snippet_user:pass1234@db:3307/snippetbox?parseTime=true", "-debug=false"]
