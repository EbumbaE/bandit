FROM golang:alpine AS build
RUN mkdir -p /tmp/build/
WORKDIR /tmp/build
COPY . .
RUN go mod tidy
RUN go build -o /usr/bin/rule-admin "./services/rule-admin/cmd/main.go"
ARG CONFIGPATH=./services/rule-admin/cmd/run/config.yaml
RUN mkdir -p /etc/rule-admin
COPY ${CONFIGPATH} /etc/rule-admin/config.yaml

FROM alpine:latest
COPY --from=build /usr/bin/rule-admin /usr/bin/rule-admin
COPY --from=build /etc/rule-admin /etc/rule-admin
CMD /usr/bin/rule-admin --config /etc/rule-admin/config.yaml
