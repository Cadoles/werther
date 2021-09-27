build: clean generate
	CGO_ENABLED=0 misc/script/build

generate:
	go generate ./...

clean:
	rm -rf bin

.PHONY: build