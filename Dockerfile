FROM bufbuild/buf:latest AS buf

FROM scratch
ENTRYPOINT ["/fauxrpc"]
ARG TARGETPLATFORM
COPY --from=buf /usr/local/bin/buf /usr/local/bin/buf
COPY ${TARGETPLATFORM}/fauxrpc /fauxrpc
