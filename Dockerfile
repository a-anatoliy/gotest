# syntax=docker/dockerfile:1
# Multi-stage, multi-arch build (linux/amd64 and linux/arm64 via docker buildx).
FROM --platform=$BUILDPLATFORM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath -ldflags="-s -w" -o /out/gotest .

# Templates and static files are embedded in the binary, so the runtime
# image only needs the executable.
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/gotest /gotest
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/gotest"]
