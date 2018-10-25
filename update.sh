#!/bin/bash
export PATH=$PATH:/usr/local/go/bin
PLUGIN_NAME="mount-guard"
SOURCE_DIR="/home/zebirdman/go/src/auth-plugin/"

monitor_source_files() {
  inotifywait "${SOURCE_DIR}"
}

update_build() {
  go build -o "${PLUGIN_NAME}" *.go
}

update_container() {
  systemctl daemon-reload
  systemctl restart docker
  docker stop "${PLUGIN_NAME}"
  docker container rm "${PLUGIN_NAME}"
  docker rmi "${PLUGIN_NAME}"
  docker build -t "${PLUGIN_NAME}" .
  docker run -d --restart=always --name "${PLUGIN_NAME}" \
    -v /run/docker/plugins/:/run/docker/plugins/ "${PLUGIN_NAME}"
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
  turn_auth_off
  update_build
  update_container
  turn_auth_on
}

while monitor_source_files; do
  rebuild_plugin
done
