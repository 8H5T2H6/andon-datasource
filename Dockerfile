FROM golang:1.13-buster as build

WORKDIR /go/src/datasource
ADD . .

RUN go mod download
RUN go build -o /go/main

FROM gcr.io/distroless/base-debian10
WORKDIR /go/
COPY --from=build /go/main .
COPY *.env ./

EXPOSE 8080

CMD ["./main"]
