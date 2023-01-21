build:
    cargo build ./...

clean:
    rm -rf build data.ms meilisearch meili_data zeno.db

deploy:
    fly deploy

dev:
    cargo run .

docker-run: docker-clean
    docker run -it -p 8080:8080 -v $(pwd)/static:/static:ro --name zeno zeno -meili /data.ms -dsn "file:/zeno.db?mode=rwc"

docker-clean:
    docker rm -f zeno

docker-build:
    docker build . -t zeno

docker-dev: docker-build docker-clean docker-run
