package main

//import "encoding/json"
import (
	"bytes"
	"encoding/json"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-plugins-helpers/authorization"
)

type mountGuard struct{}

type configWrapper struct {
	*container.Config
	HostConfig *container.HostConfig
}

func main() {

	plugin, err := newPlugin()
	handler := authorization.NewHandler(plugin)

	err = handler.ServeUnix("mount-guard", 0)
	if err != nil {
		log.Fatal(err)
	}
}

func newPlugin() (*mountGuard, error) {
	return &mountGuard{}, nil
}

func (p *mountGuard) AuthZReq(req authorization.Request) authorization.Response {

	if req.RequestBody != nil {

		log.Info("hello")

		log.WithFields(log.Fields{"Method": req.RequestMethod}).Info("Method field")
		log.WithFields(log.Fields{"URI": req.RequestURI}).Info("URI field")

		body := &configWrapper{}
		json.NewDecoder(bytes.NewReader(req.RequestBody)).Decode(body)

		log.WithFields(log.Fields{"Image": body.Image}).Info("Struct-contents")
		log.WithFields(log.Fields{"User": body.User}).Info("Struct-contents")
		log.WithFields(log.Fields{"Tty": body.Tty}).Info("Struct-contents")

		for _, element := range body.HostConfig.Binds {
			log.WithFields(log.Fields{"Host-bind-point": element}).Info("Binds")
		}
	}

	return authorization.Response{Allow: true}
}

func (p *mountGuard) AuthZRes(req authorization.Request) authorization.Response {

	return authorization.Response{Allow: true}
}
