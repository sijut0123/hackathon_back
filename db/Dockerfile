FROM golang:1.18 as build
WORKDIR /go/src/app
ENV HOSTNAME "0.0.0.0"
COPY . .
EXPOSE 8050
RUN echo 1
RUN go mod download
RUN echo 2
RUN go build -o /go/bin/app
RUN echo 3
CMD ["/go/bin/app"]
