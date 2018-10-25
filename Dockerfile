
FROM alpine

RUN mkdir -p /run/docker/plugins/
RUN mkdir -p /bin

VOLUME /run/docker/plugins/
ADD ./mount-guard  /bin/mount-guard

ENTRYPOINT ["/bin/mount-guard"]