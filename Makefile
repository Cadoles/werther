PACKAGE_VERSION ?= $(shell git describe --always | rev | cut -d '/' -f 1 | rev)
NFPM_PACKAGER ?= deb

build: clean generate
	CGO_ENABLED=0 misc/script/build

generate:
	go generate ./...

clean:
	rm -rf bin

package: dist
	PACKAGE_VERSION=$(PACKAGE_VERSION) \
	nfpm package \
		--config misc/packaging/nfpm.yml \
		--target ./dist \
		--packager $(NFPM_PACKAGER)

dist:
	mkdir -p dist

.PHONY: build