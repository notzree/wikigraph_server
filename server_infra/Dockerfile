FROM golang:1.20-alpine

RUN apk add --no-cache make protobuf

WORKDIR /wikigraph_server

COPY . .

RUN go mod download

RUN make build

EXPOSE 3001


CMD [ "make", "run"]