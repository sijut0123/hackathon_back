FROM golang:1.18 as build
WORKDIR /go/src/app
ENV HOSTNAME "0.0.0.0"
COPY ./main.go .
EXPOSE 8050
RUN echo 1
RUN go mod init app && go mod tidy
RUN echo 2
RUN go build -o /go/bin/app
CMD ["/go/bin/app"]
