# Build the manager binary
FROM registry.access.redhat.com/ubi9/ubi-minimal as builder

COPY go.mod /go.mod

RUN microdnf -y install gzip jq tar && \
    arch=$(uname -m | sed 's/x86_64/amd64/') && \
    goversion=$(sed -nr 's/^go\s+([0-9]+\.[0-9]+)(\.[0-9]+)?/go\1/p' /go.mod) && \
    file=$(curl -s 'https://go.dev/dl/?mode=json&include=all' | jq -r --arg GOVERSION "${goversion}" --arg ARCH "${arch}" 'first(.[] | select(.version | startswith($GOVERSION))).files[] | select(.arch == $ARCH and .os == "linux").filename') && \
    curl -s -f -L https://go.dev/dl/${file} | tar -C / -xzf -

ENV PATH=/go/bin:/root/.local/bin:/root/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY cmd/main.go cmd/main.go
COPY api/ api/
COPY internal/controller/ internal/controller/

# Build go executable
RUN CGO_ENABLED=0 go build -a -o manager cmd/main.go


# Build operator image
FROM registry.access.redhat.com/ubi9/ubi-minimal:latest

LABEL name="ManageIQ Operator" \
      summary="ManageIQ Operator manages the ManageIQ application on a Kubernetes cluster" \
      vendor="ManageIQ" \
      description="ManageIQ is a management and automation platform for virtual, private, and hybrid cloud infrastructures."

WORKDIR /
COPY --from=builder /workspace/manager .
USER 65532:65532

# install operator manifest
COPY BUILD /opt/manageiq/manifest

ENTRYPOINT ["/manager"]
