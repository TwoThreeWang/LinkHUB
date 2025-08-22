FROM golang:alpine AS builder

LABEL stage=gobuilder

ENV CGO_ENABLED 0

RUN apk update --no-cache && apk add --no-cache tzdata

WORKDIR /build

# 先复制依赖文件，利用缓存
COPY go.mod go.sum ./
RUN go mod download

# 再复制源码，避免每次重新下载依赖
COPY . .
RUN go mod tidy && go build -p $(nproc) -ldflags="-s -w" -o /app/main main.go


FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /usr/share/zoneinfo/Asia/Shanghai /usr/share/zoneinfo/Asia/Shanghai
ENV TZ Asia/Shanghai

WORKDIR /
COPY --from=builder /app/main /main
CMD ["/main"]