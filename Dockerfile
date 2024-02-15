#
# MIT License
#
# Copyright (c) His Majesty the King in Right of Canada, as represented by the Minister responsible for Statistics Canada, 2024

# Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"),
# to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
# and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

# The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
# WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

ARG IMAGE_REPOSITORY=docker.io/docker
ARG TARGETARCH
ARG TARGETOS=linux

FROM ${IMAGE_REPOSITORY}/golang:1.22 as builder
ARG TARGETOS
ARG TARGETARCH

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

# Copy the go source
COPY api/ api/
COPY internal/ internal/
COPY cmd/main.go cmd/main.go

# Build
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -a -o nodecidrallocator cmd/main.go

# Using scratch base to host binary with minimal impact/attach surface area
FROM scratch
WORKDIR /
COPY --from=builder /workspace/nodecidrallocator .
USER 65532:65532

ENTRYPOINT ["/nodecidrallocator"]
