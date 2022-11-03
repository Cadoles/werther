FROM alpine:3.16

RUN apk add --no-cache make git curl jq bash openssl go zip

RUN curl -k https://forge.cadoles.com/Cadoles/Jenkins/raw/branch/master/resources/com/cadoles/common/add-letsencrypt-ca.sh | bash

RUN wget https://github.com/goreleaser/nfpm/releases/download/v2.20.0/nfpm_2.20.0_Linux_x86_64.tar.gz \
    && tar -xzf nfpm_2.20.0_Linux_x86_64.tar.gz -C /usr/local/bin \
    && chmod +x /usr/local/bin/nfpm