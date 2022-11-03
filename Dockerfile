FROM ubuntu as downloader

RUN apt-get update && \
    apt-get install curl -y && \
    curl -OL https://github.com/meilisearch/meilisearch/releases/download/v0.29.1/meilisearch-linux-amd64 && \
    chmod u+x meilisearch-linux-amd64

FROM golang:1.19-bullseye as builder

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
COPY static ./static

RUN go mod download

COPY *.go ./

RUN go build -o /zeno

FROM gcr.io/distroless/cc-debian11

WORKDIR /

COPY --from=downloader /meilisearch-linux-amd64 /meilisearch
COPY --from=builder /zeno .

EXPOSE 8080

ENTRYPOINT ["/zeno"]
CMD ["--dbpath", "/meili_data/data"]
