FROM golang:1.19-bullseye as meilisearch
WORKDIR /
RUN apt-get update
RUN apt-get install curl -y
RUN curl -OL https://github.com/meilisearch/meilisearch/releases/download/v0.29.1/meilisearch-linux-amd64
RUN chmod u+x meilisearch-linux-amd64
RUN mv meilisearch-linux-amd64 meilisearch

FROM golang:1.19-bullseye as builder
WORKDIR /app
COPY go.* ./
COPY static ./static
RUN go mod download
COPY . ./
RUN go build -o /zeno

FROM bitnami/minideb:bullseye
WORKDIR /
RUN install_packages poppler-utils ca-certificates
COPY --from=meilisearch /meilisearch /meilisearch
COPY --from=builder /zeno .
COPY ./static /static
EXPOSE 8080
ENTRYPOINT ["/zeno"]
CMD ["--dbpath", "/meili_data/data"]
