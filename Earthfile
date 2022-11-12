VERSION 0.6
FROM golang:1.19-bullseye
WORKDIR /app

meilisearch:
    WORKDIR /
    RUN apt-get update
    RUN apt-get install curl -y
    RUN curl -OL https://github.com/meilisearch/meilisearch/releases/download/v0.29.1/meilisearch-linux-amd64
    RUN chmod u+x meilisearch-linux-amd64
    RUN chmod u+x meilisearch-linux-amd64
    SAVE ARTIFACT /meilisearch-linux-amd64 /meilisearch AS LOCAL build/meilisearch

build:
    COPY go.* ./
    COPY static ./static
    RUN go mod download
    COPY *.go ./
    RUN go build -o /zeno
    SAVE ARTIFACT /zeno AS LOCAL build/zeno

docker:
    FROM bitnami/minideb:bullseye
    WORKDIR /
    RUN install_packages poppler-utils ca-certificates
    COPY +meilisearch/meilisearch /meilisearch
    COPY +build/zeno .

prod:
    FROM +docker
    COPY ./static /static
    EXPOSE 8080
    ENTRYPOINT ["/zeno"]
    CMD ["--dbpath", "/meili_data/data"]
    SAVE IMAGE --push registry.fly.io/zeno:latest

dev:
    FROM +docker
    EXPOSE 8080
    ENTRYPOINT ["/zeno"]
    CMD ["--dbpath", "/meili_data/data"]
    SAVE IMAGE zeno-dev

clean:
    LOCALLY
    RUN rm -rf build

# docker run -it --rm --name zeno -p 8080:8080 -v /workspace/scrap:/meili_data -v $(pwd)/static:/static:ro zeno-dev
# docker run -it --rm --name zeno -p 8080:8080 -v /workspace/scrap:/meili_data registry.fly.io/zeno:latest
# docker run -it --rm --name zeno -p 8080:8080 -v /workspace/scrap:/meili_data -v $(pwd)/static:/static:ro --entrypoint /bin/bash zeno-dev
# docker run -it --rm --name zeno -p 8080:8080 -v /workspace/scrap:/meili_data --entrypoint /bin/bash registry.fly.io/zeno:latest