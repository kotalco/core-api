FROM golang:1.19-alpine AS builder

RUN apk add --no-cache curl tar
ENV K8S_VERSION=1.25.0
# download kubebuilder tools required by envtest
RUN curl -sSLo envtest-bins.tar.gz "https://storage.googleapis.com/kubebuilder-tools/kubebuilder-tools-${K8S_VERSION}-$(go env GOOS)-$(go env GOARCH).tar.gz"
RUN tar -vxzf envtest-bins.tar.gz -C /usr/local/

WORKDIR /api

COPY . .

ARG EC_PUBLIC_KEY
ARG SENDGRID_API_KEY

RUN CGO_ENABLED=0 go build -ldflags="-X 'github.com/kotalco/cloud-api/pkg/config.SendgridAPIKey=${SENDGRID_API_KEY}' -X 'github.com/kotalco/cloud-api/pkg/config.ECCPublicKey=${EC_PUBLIC_KEY}'" -v -o server

FROM alpine

# Add new user 'kotal'
RUN adduser -D kotal
USER kotal

# required by api server to determine config/crds path
ENV GOPATH=/go
COPY --from=builder /go/pkg/mod/github.com/kotalco /go/pkg/mod/github.com/kotalco
# tools (etcd, apiserver, and kubectl) required by envtest
COPY --from=builder /usr/local/kubebuilder /usr/local/kubebuilder
COPY --from=builder /api/server /home/kotal/api/server

WORKDIR /home/kotal
EXPOSE 8080
ENV ETCD_UNSUPPORTED_ARCH=arm64
ENTRYPOINT [ "./api/server" ]
