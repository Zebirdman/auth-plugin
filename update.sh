#!/bin/bash
export PATH=$PATH:/usr/local/go/bin
PLUGIN_NAME='mount-guard'
SOURCE_DIR="/media/sf_VM_share/go_projects/auth_plugin/"

monitor_source_files() {
  inotifywait "${SOURCE_DIR}"
}
update_build() {
  ~/Documents/local-update.sh
  cp -rf /media/sf_VM_share/go_projects/*  /home/ben/Documents/go_projects/
  cp ~/Documents/go_projects/auth_plugin/* ~/go/src/plugin/
  go build -o "${PLUGIN_name}" *.go
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
  local auth_regex=`s|fd://|fd:// --authorization-plugin=${PLUGIN_NAME}|g`
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
