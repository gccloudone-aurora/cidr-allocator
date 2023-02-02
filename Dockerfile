# Build the NodeCIDRAllocation controller binary
FROM artifactory.cloud.statcan.ca/docker/golang:1.19 as builder
ARG TARGETOS
ARG TARGETARCH

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

# Copy the go source
COPY main.go main.go
COPY api/ api/
COPY controllers/ controllers/
COPY util/ util/

# Build
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} go build -a -o cnp-nodecidrallocator main.go

# Using scratch base to host binary with minimal impact/attach surface area
FROM scratch
WORKDIR /
COPY --from=builder /workspace/cnp-nodecidrallocator .
USER 65532:65532

ENTRYPOINT ["/cnp-nodecidrallocator"]
