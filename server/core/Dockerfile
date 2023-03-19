FROM golang:1.19.3-alpine3.17 AS builder

WORKDIR /app

COPY . .

RUN echo "runner:x:1000:1000:runner,,,::/bin/false" > /user.txt

RUN cd ./cmd/httpserver/ &&\
    go mod tidy &&\
    CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -o /runner -gcflags=all="-l -B -C" -ldflags="-w -s" .

FROM scratch AS base

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /user.txt /etc/passwd
COPY --from=builder /runner /runner

USER runner

ENV PORT "9000"
ENV MODEL_MAX_TOKENS "1000"

EXPOSE ${PORT}

ENTRYPOINT ["./runner"]
