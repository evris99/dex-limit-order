FROM golang:1.17-alpine3.15
ADD . /app
WORKDIR /app
RUN apk add build-base
RUN go build -o dex-limit-order .
CMD [ "/app/dex-limit-order" ]