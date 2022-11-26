install:
    curl -L https://install.meilisearch.com | sh

build:
    go get ./...
    go build -o build/zeno

clean:
    rm -rf build data.ms meilisearch

deploy:
    fly deploy

dev:
    go run .