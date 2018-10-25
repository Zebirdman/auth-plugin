package main

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/go-plugins-helpers/authorization"
)

type mountGuard struct {
}

func main() {

	plugin, err := newPlugin()
	handler := authorization.NewHandler(plugin)

	err = handler.ServeUnix("mountGuard", 0)
	if err != nil {
		log.Fatal(err)
	}
}

func newPlugin() (*mountGuard, error) {
	return &mountGuard{}, nil
}

func (p *mountGuard) AuthZReq(req authorization.Request) authorization.Response {

	fmt.Println("\n[INFO] Request recieved update: ")
	/* test */

	log.WithFields(log.Fields{
		"User": req.User,
	}).Info("User field")

	log.WithFields(log.Fields{
		"Method": req.RequestMethod,
	}).Info("Method field")

	log.WithFields(log.Fields{
		"URI": req.RequestURI,
	}).Info("URI field")

	log.WithFields(log.Fields{
		"Body": req.RequestBody,
	}).Info("Body field")

	return authorization.Response{Allow: true}
}

func (p *mountGuard) AuthZRes(req authorization.Request) authorization.Response {

	return authorization.Response{Allow: true}
}
