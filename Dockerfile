FROM gcr.io/distroless/static
ADD k8s-generic-validator /
ENTRYPOINT ["/k8s-generic-validator"]
