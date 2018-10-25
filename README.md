auth_plugin

Notes:

1) simply need to run container with /run/docker/plugin dir mounted to host
2) the .socket file is what we pass to Docker as the name of the plugin minus the .socket part
3) we can use logrus to log to stdout or stderr, can simple view logs in real time using docker logs -f
win!
