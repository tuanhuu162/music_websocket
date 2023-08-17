FROM golang:latest as builder

WORKDIR /go/src/ex8
# Syntax: COPY ./source ./destination
COPY . .

RUN go get -d -v ./...

RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o main .
EXPOSE 8080
CMD ["./main"]
