FROM golang:alpine AS build
RUN mkdir -p /tmp/build/
WORKDIR /tmp/build
COPY . .
RUN go mod tidy

ARG CONFIGPATH=/etc/rule-analytic/config.yaml

RUN go build -o /usr/bin/rule-analytic "./services/rule-analytic/cmd/main.go"
RUN mkdir -p /etc/rule-analytic

COPY ./services/rule-analytic/cmd/run/config.yaml ${CONFIGPATH}

FROM alpine:latest
COPY --from=build /usr/bin/rule-analytic /usr/bin/rule-analytic
COPY --from=build /etc/rule-analytic /etc/rule-analytic
CMD /usr/bin/rule-analytic --config /etc/rule-analytic/config.yaml
