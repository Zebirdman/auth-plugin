
FROM alpine

RUN mkdir -p /run/docker/plugins/ && \
    mkdir -p /bin && \
    mkdir -p /plugin-config &&\ 
    mkdir -p /policies

VOLUME /run/docker/plugins/
ADD ./mount-guard  /bin/mount-guard

ENTRYPOINT ["/bin/mount-guard"]