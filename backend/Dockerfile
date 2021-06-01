FROM golang:1.16.4-alpine3.13 as builder

WORKDIR /go/src/github.com/htchan/WebHistory

COPY ./src .
COPY ./go.mod .
RUN sed -i 's/.\/src\//.\//g' ./go.mod

RUN apk add gcc musl-dev libc-dev
RUN go mod download

RUN go build -v ./main/backend.go


FROM golang:1.16.3-alpine3.13

WORKDIR /usr/src/app

COPY --from=builder /go/src/github.com/htchan/BookSpider/main .

CMD ['./main']
