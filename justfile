install:
    curl -L https://install.meilisearch.com | sh

build:
    go get ./...
    go build -o build/zeno

clean:
    rm -rf build data.ms meilisearch meili_data zeno.db

deploy:
    fly deploy

dev:
    go run .

docker-run:
    docker run -it -p 8080:8080 --name zeno zeno -meili /data.ms -dsn file:/zeno.db?mode=rwc

docker-clean:
    docker rm -f zeno

docker-build:
    docker build . -t zeno

docker-dev: docker-build docker-clean docker-run
