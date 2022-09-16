# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY ./security-pack-helper /security-pack-helper
USER 65532:65532

ENTRYPOINT ["/security-pack-helper"]
