FROM golang:alpine AS build
RUN mkdir -p /tmp/build/
WORKDIR /tmp/build
COPY . .
RUN go mod tidy

ARG CONFIGPATH=/etc/rule-test/config.yaml
ARG SWAGGERGPATH=/etc/rule-test/test.swagger.json

RUN go build -o /usr/bin/rule-test "./services/rule-test/cmd/main.go"
RUN mkdir -p /etc/rule-test

COPY ./services/rule-test/cmd/run/config.yaml ${CONFIGPATH}
COPY ./pkg/genproto/rule-test/api/test.swagger.json ${SWAGGERGPATH}

FROM alpine:latest
COPY --from=build /usr/bin/rule-test /usr/bin/rule-test
COPY --from=build /etc/rule-test /etc/rule-test
CMD /usr/bin/rule-test --config /etc/rule-test/config.yaml --swagger /etc/rule-test/test.swagger.json
