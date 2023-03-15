FROM --platform=$BUILDPLATFORM golang:1.20-alpine AS build_base
ARG TARGETOS
ARG TARGETARCH
ARG BUILDPLATFORM
RUN apk add bash ca-certificates git gcc g++ libc-dev
WORKDIR /workspace
# We want to populate the module cache based on the go.{mod,sum} files.
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
# And compile the project
RUN echo "I am running on $BUILDPLATFORM, building for $TARGETOS $TARGETARCH" > /log
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -ldflags="-s -w" -o ist .
FROM alpine as runner
COPY --from=build_base /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build_base /workspace/ist /bin/ist
ENTRYPOINT ["/bin/ist"]
