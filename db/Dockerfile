FROM golang:1.18 as build
WORKDIR /db
COPY ./main.go ./go.mod ./go.sum ./.env /db/
CMD ["go", "run", "main.go"]
EXPOSE 8050
ENV HOSTNAME "0.0.0.0"