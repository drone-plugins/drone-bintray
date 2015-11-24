# Docker image for Drone Bintray plugin
#
#     CGO_ENABLED=0 go build -a -tags netgo
#     docker build --rm=true -t plugins/drone-bintray .

FROM gliderlabs/alpine:3.1
RUN apk-install ca-certificates
RUN apk-install curl
ADD drone-bintray /bin/
ENTRYPOINT ["/bin/drone-bintray"]
