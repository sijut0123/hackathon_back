FROM golang:1.18 as build
WORKDIR /go/src/app
ENV HOSTNAME "0.0.0.0"
COPY . .
EXPOSE 8050
RUN go mod download
RUN go build -o /go/bin/app
CMD ["/go/bin/app"]





