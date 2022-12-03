FROM golang:1.19 as builder
COPY . /build
WORKDIR /build
RUN go mod download
# Build
RUN CGO_ENABLED=0 GOOS=linux go build -a -o controller .

FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /build/controller .

COPY --from=builder --chown=nonroot:nonroot /build/controller /

USER nonroot:nonroot

ENTRYPOINT ["/controller"]