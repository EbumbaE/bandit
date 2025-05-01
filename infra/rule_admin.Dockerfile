FROM golang:alpine AS build
RUN mkdir -p /tmp/build/
WORKDIR /tmp/build
COPY . .
RUN go mod tidy

ARG CONFIGPATH=/etc/rule-admin/config.yaml
ARG SWAGGERGPATH=/etc/rule-admin/admin.swagger.json

RUN go build -o /usr/bin/rule-admin "./services/rule-admin/cmd/main.go"
RUN mkdir -p /etc/rule-admin

COPY ./services/rule-admin/cmd/run/config.yaml ${CONFIGPATH}
COPY ./pkg/genproto/rule-admin/api/admin.swagger.json ${SWAGGERGPATH}

FROM alpine:latest
COPY --from=build /usr/bin/rule-admin /usr/bin/rule-admin
COPY --from=build /etc/rule-admin /etc/rule-admin
CMD /usr/bin/rule-admin --config /etc/rule-admin/config.yaml --swagger /etc/rule-admin/admin.swagger.json
