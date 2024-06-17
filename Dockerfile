FROM golang:1.22-alpine3.20 as builder

WORKDIR /build
COPY . .

RUN go env -w GO111MODULE=on \
    && go env -w CGO_ENABLED=0 \
    && go env \
    && go mod tidy \
    && go build -gcflags=-trimpath=$GOPATH -asmflags=-trimpath=$GOPATH -ldflags "-w -s" -o /build/transfer main.go

FROM alpine:latest

ARG AUTHOR=1228022817@qq.com
LABEL org.opencontainers.image.authors=${AUTHOR}

WORKDIR /app
ENV TZ=Asia/Shanghai
RUN apk update --no-cache && apk add --no-cache tzdata

COPY --from=builder /build/transfer /usr/local/bin/


EXPOSE 80

CMD ["transfer"]
