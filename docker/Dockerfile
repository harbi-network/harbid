# -- multistage docker build: stage #1: build stage
FROM golang:1.19-alpine AS build

RUN mkdir -p /go/src/github.com/harbi-network/harbid

WORKDIR /go/src/github.com/harbi-network/harbid

RUN apk add --no-cache curl git openssh binutils gcc musl-dev

COPY go.mod .
COPY go.sum .


# Cache harbid dependencies
RUN go mod download

COPY . .

RUN go build $FLAGS -o harbid .

# --- multistage docker build: stage #2: runtime image
FROM alpine
WORKDIR /app

RUN apk add --no-cache ca-certificates tini

COPY --from=build /go/src/github.com/harbi-network/harbid/harbid /app/
COPY --from=build /go/src/github.com/harbi-network/harbid/infrastructure/config/sample-harbid.conf /app/

USER nobody
ENTRYPOINT [ "/sbin/tini", "--" ]
