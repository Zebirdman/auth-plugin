#!/bin/bash
export PATH=$PATH:/usr/local/go/bin
PLUGIN_NAME="mount-guard"
SOURCE_DIR="/home/zebirdman/go/src/auth-plugin/main.go"
BUILD_SUCCESS=0
POLICY_DIR="/home/zebirdman/go/src/auth-plugin/policies/"

monitor_source_files() {
  inotifywait -e create,delete,modify "${SOURCE_DIR}"
}

update_build() {
  CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -o "${PLUGIN_NAME}" *.go
  BUILD_SUCCESS=$?
}

update_container() {
  systemctl daemon-reload
  systemctl restart docker
  docker stop "${PLUGIN_NAME}"
  docker container rm "${PLUGIN_NAME}"
  docker rmi "${PLUGIN_NAME}"
  docker build -t "${PLUGIN_NAME}" .
  docker run -d --restart=always --name "${PLUGIN_NAME}" \
    -v /run/docker/plugins/:/run/docker/plugins/ \
    -v "${POLICY_DIR}:/policies" \
    "${PLUGIN_NAME}"
}

turn_auth_off() {
  local auth_regex='s/ --authorization-plugin=[^ ]*/ /g'
  sed -i -e "${auth_regex}" "/etc/systemd/system/docker.service"
}

turn_auth_on() {
  local auth_regex='s|fd://|fd:// --authorization-plugin=mount-guard|g'
  sed -i -e "${auth_regex}" "/etc/systemd/system/docker.service"
  systemctl daemon-reload
  systemctl restart docker
}

rebuild_plugin() {
  
  update_build
  if [ $BUILD_SUCCESS = 0 ]
  then
    turn_auth_off
    update_container
    turn_auth_on
  else
    echo "Build unsuccesful"
  fi
}

while monitor_source_files; do
  rebuild_plugin
done
