# Docker image for Drone Bintray plugin
#
#     CGO_ENABLED=0 go build -a -tags netgo
#     docker build --rm=true -t plugins/drone-bintray .

FROM alpine:3.3
RUN apk update && apk add ca-certificates
ADD drone-bintray /bin/
ENTRYPOINT ["/bin/drone-bintray"]
