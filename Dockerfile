FROM golang:1.18-alpine

RUN apk add --no-cache make protobuf

WORKDIR /wikigraph_server

COPY go.mod go.sum ./

RUN go mod download

RUN make build

RUN make run

EXPOSE 3000

EXPOSE 8080

CMD [ "/wikigraph_server"]