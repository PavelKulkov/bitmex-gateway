FROM golang:1.17.2-alpine

ENV APP_NAME provider

COPY . /go/src/${APP_NAME}
WORKDIR /go/src/${APP_NAME}

RUN go build -o ${APP_NAME}

CMD ./${APP_NAME}