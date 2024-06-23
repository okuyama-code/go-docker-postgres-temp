FROM golang:1.21.6

WORKDIR /usr/local/go/src/react_go_app/app

# アプリケーションのソースコードをコピー
COPY ./app .

# 依存関係のダウンロード
RUN go mod download

# アプリケーションのビルド
RUN go build -o main .

EXPOSE 8080

