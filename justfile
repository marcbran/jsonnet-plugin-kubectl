
test:
    @go test ./...
    @jpoet test .

build: test
    @jpoet pkg build

push: build
    @jpoet pkg push
