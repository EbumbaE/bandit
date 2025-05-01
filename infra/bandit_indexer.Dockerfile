FROM golang:alpine AS build
RUN mkdir -p /tmp/build/
WORKDIR /tmp/build
COPY . .
RUN go mod tidy

ARG CONFIGPATH=/etc/bandit-indexer/config.yaml
ARG SWAGGERGPATH=/etc/bandit-indexer/indexer.swagger.json

RUN go build -o /usr/bin/bandit-indexer "./services/bandit-indexer/cmd/main.go"
RUN mkdir -p /etc/bandit-indexer

COPY ./services/bandit-indexer/cmd/run/config.yaml ${CONFIGPATH}
COPY ./pkg/genproto/bandit-indexer/api/indexer.swagger.json ${SWAGGERGPATH}

FROM alpine:latest
COPY --from=build /usr/bin/bandit-indexer /usr/bin/bandit-indexer
COPY --from=build /etc/bandit-indexer /etc/bandit-indexer
CMD /usr/bin/bandit-indexer --config /etc/bandit-indexer/config.yaml --swagger /etc/bandit-indexer/indexer.swagger.json
