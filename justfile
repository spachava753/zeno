build:
    go get ./...
    go build -o build/zeno

clean:
    rm -rf build data.ms

deploy:
    fly deploy