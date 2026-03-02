FROM gcr.io/distroless/static-debian13:nonroot
LABEL maintainer="metal-stack authors <info@metal-stack.io>"
COPY bin/os-installer /os-installer
ENTRYPOINT ["/os-installer"]
