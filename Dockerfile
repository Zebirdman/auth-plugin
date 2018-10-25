
FROM ubuntu

RUN mkdir -p /run/docker/plugins/

VOLUME /run/docker/plugins/
ADD ./mountGuard  /bin/mountGuard

ENTRYPOINT ["/bin/mountGuard"]