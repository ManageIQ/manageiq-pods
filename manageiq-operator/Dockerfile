# Build the manager binary
FROM docker.io/library/golang:1.23 as builder

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

# Build
RUN CGO_ENABLED=0 go build -a -o manager cmd/main.go

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
