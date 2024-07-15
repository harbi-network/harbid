ARG HARPAD_IMAGE
ARG HARPADMINER_IMAGE

FROM ${HARPAD_IMAGE} as harbid
FROM ${HARPADMINER_IMAGE} as harbiminer

FROM golang:1.19-alpine

RUN mkdir -p /go/src/github.com/harbi-network/harbid

WORKDIR /go/src/github.com/harbi-network/harbid

RUN apk add bash build-base git

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

COPY --from=harbid /app/ /app/
COPY --from=harbiminer /app/ /app/
ENV PATH="/app:${PATH}"

WORKDIR /go/src/github.com/harbi-network/harbid/stability-tests

RUN go install ./...

ENTRYPOINT ["./run/run.sh"]
