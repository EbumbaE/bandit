FROM golang:alpine AS build
RUN mkdir -p /tmp/build/
WORKDIR /tmp/build
COPY . .
RUN go mod tidy

ARG CONFIGPATH=/etc/rule-diller/config.yaml
ARG SWAGGERGPATH=/etc/rule-diller/diller.swagger.json

RUN go build -o /usr/bin/rule-diller "./services/rule-diller/cmd/main.go"
RUN mkdir -p /etc/rule-diller

COPY ./services/rule-diller/cmd/run/config.yaml ${CONFIGPATH}
COPY ./pkg/genproto/rule-diller/api/diller.swagger.json ${SWAGGERGPATH}

FROM alpine:latest
COPY --from=build /usr/bin/rule-diller /usr/bin/rule-diller
COPY --from=build /etc/rule-diller /etc/rule-diller
CMD /usr/bin/rule-diller --config /etc/rule-diller/config.yaml --swagger /etc/rule-diller/diller.swagger.json
