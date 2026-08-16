FROM golang:1.22-alpine AS builder

WORKDIR /src
COPY go.mod ./
RUN go mod download || true

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/k8s-pod-autoscale .

FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /app/k8s-pod-autoscale /k8s-pod-autoscale
USER 65532:65532

ENTRYPOINT ["/k8s-pod-autoscale"]
