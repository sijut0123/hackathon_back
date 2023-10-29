FROM golang:1.18 as build
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download && go build main.go  # 依存関係を取得
COPY main.go .
EXPOSE 8000
RUN go build main.go
CMD ["./main"]