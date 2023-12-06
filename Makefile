SHELL := /bin/bash

IMAGE_NAME := reg.cadoles.com/cadoles/hydra-werther

NFPM_VERSION ?= 2.20.0
NFPM_PACKAGERS ?= deb rpm

MKT_GITEA_RELEASE_ORG ?= Cadoles
MKT_GITEA_RELEASE_PROJECT ?= hydra-werther
MKT_GITEA_RELEASE_VERSION ?= $(MKT_PROJECT_VERSION)

define MKT_GITEA_RELEASE_BODY
## Docker usage

```
docker pull $(IMAGE_NAME):$(MKT_PROJECT_VERSION)
```
endef
export MKT_GITEA_RELEASE_BODY

build: build-bin build-image

build-bin: clean generate
	CGO_ENABLED=0 misc/script/build

test: scan

generate:
	go generate ./...

clean:
	rm -rf bin dist

dist:
	mkdir -p dist

package: clean build-bin $(foreach p,$(NFPM_PACKAGERS), package-$(p))

package-%: dist tools/nfpm/bin/nfpm
	PACKAGE_VERSION=$(MKT_PROJECT_VERSION) \
		tools/nfpm/bin/nfpm package \
			--config misc/packaging/nfpm.yml \
			--target ./dist \
			--packager $*

tools/nfpm/bin/nfpm:
	mkdir -p tools/nfpm/bin
	curl -L --output tools/nfpm/nfpm_$(NFPM_VERSION)_Linux_x86_64.tar.gz https://github.com/goreleaser/nfpm/releases/download/v$(NFPM_VERSION)/nfpm_$(NFPM_VERSION)_Linux_x86_64.tar.gz \
        && tar -xzf tools/nfpm/nfpm_$(NFPM_VERSION)_Linux_x86_64.tar.gz -C tools/nfpm/bin \
        && chmod +x tools/nfpm/bin/nfpm \
		&& rm -f tools/nfpm/nfpm_$(NFPM_VERSION)_Linux_x86_64.tar.gz

build-image:
	docker build \
		-t "${IMAGE_NAME}:latest" \
		.	

scan: build-image tools/trivy/bin/trivy
	mkdir -p .trivy
	tools/trivy/bin/trivy --cache-dir .trivy/.cache image --ignorefile .trivyignore.yaml $(TRIVY_ARGS) $(IMAGE_NAME):latest
	
tools/trivy/bin/trivy:
	mkdir -p tools/trivy/bin
	curl -sfL https://raw.githubusercontent.com/aquasecurity/trivy/main/contrib/install.sh | sh -s -- -b ./tools/trivy/bin v0.47.0

release: release-image release-gitea

release-gitea: .mktools package
	@[ ! -z "$(MKT_PROJECT_VERSION)" ] || ( echo "Just downloaded mktools. Please re-run command."; exit 1 )
	$(MAKE) MKT_GITEA_RELEASE_ATTACHMENTS="$$(find dist/* -type f -printf '%p ')" mkt-gitea-release

release-image: .mktools
	@[ ! -z "$(MKT_PROJECT_VERSION)" ] || ( echo "Just downloaded mktools. Please re-run command."; exit 1 )
	docker tag "${IMAGE_NAME}:latest" "${IMAGE_NAME}:$(MKT_PROJECT_VERSION)"
	docker tag "${IMAGE_NAME}:latest" "${IMAGE_NAME}:$(MKT_PROJECT_SHORT_VERSION)"
	docker tag "${IMAGE_NAME}:latest" "${IMAGE_NAME}:$(MKT_PROJECT_VERSION_CHANNEL)-latest"
	
	docker push "${IMAGE_NAME}:$(MKT_PROJECT_VERSION)"
	docker push "${IMAGE_NAME}:$(MKT_PROJECT_SHORT_VERSION)"
	docker push "${IMAGE_NAME}:$(MKT_PROJECT_VERSION_CHANNEL)-latest"

.mktools:
	rm -rf .mktools
	curl -q https://forge.cadoles.com/Cadoles/mktools/raw/branch/master/install.sh | TASKS="version gitea" $(SHELL)

-include .mktools/*.mk