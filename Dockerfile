FROM golang:1.23.2 AS builder

WORKDIR /app
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
        GOPROXY="https://goproxy.cn,direct"

COPY go.mod go.sum ./

RUN go mod download


COPY . .

RUN go build -o keepalived_expoter .
FROM alpine:latest
ENV TZ=Asia/Shanghai
RUN ln -sf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone
COPY --from=builder /app/keepalived_expoter /keepalived_expoter

ENV PORT=2112

EXPOSE 2112


CMD ["/keepalived_expoter"]

