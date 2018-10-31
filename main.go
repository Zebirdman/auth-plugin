package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-plugins-helpers/authorization"
	"github.com/fsnotify/fsnotify"
)

const debug = true
const policyDir = "/policies/"

/* empty struct representing the plugin, we implement the
*  plugins required functions below
**/
type mountGuard struct {
	Policies []policy
}

/* wrapper for holding data associated with container creation
* see https://docs.docker.com/engine/api/v1.37/#operation/ContainerCreate
**/
type configWrapper struct {
	*container.Config
	HostConfig *container.HostConfig
}

/* policy struct matches the format of the JSON policies themselves
*  for easy unmarhsaling
**/
type policy struct {
	User          string   `json:"user"`
	AllowedMounts []string `json:"allowedMounts"`
}

/* main, initialize plugin and serve on unix socket
**/
func main() {
	plugin, err := newPlugin()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	go func() {
		for {
			select {

			case event := <-watcher.Events:

				log.WithFields(log.Fields{
					"Event": event,
				}).Info("Policy watch event")

				files, err := ioutil.ReadDir(policyDir)
				handleErr(err, "File read")
				plugin.extractAllPolicies(files)

			case err := <-watcher.Errors:
				log.WithFields(log.Fields{
					"Error": err,
				}).Error("Policy watch error")
			}
		}
	}()

	err = watcher.Add(policyDir)
	if err != nil {
		log.Fatal(err)
	}

	log.WithFields(log.Fields{
		"Dir": policyDir,
	}).Info("Policy watcher set up")

	handler := authorization.NewHandler(plugin)
	err = handler.ServeUnix("mount-guard", 0)
	handleErr(err, "Serve unix")

}

func newPlugin() (*mountGuard, error) {
	p := &mountGuard{}
	files, err := ioutil.ReadDir(policyDir)
	handleErr(err, "File read")
	p.extractAllPolicies(files)
	return p, nil
}

/* extracts a JSON file into a policy structure returns pointer
*  @param name - filename to be opened
*  @return - pointer to the new policy struct
**/
func extractPolicy(name string) policy {
	// open file and set to close on func exit
	jsonFile, err := os.Open(name)
	handleErr(err, "Policy read")
	defer jsonFile.Close()
	// read json into byte array
	byteValue, err := ioutil.ReadAll(jsonFile)
	handleErr(err, "IO read")
	usePol := policy{}
	err = json.Unmarshal(byteValue, &usePol)
	handleErr(err, "json unmarshall")
	return usePol
}

/* checks requested bind mounts vs allowed, returns requested mount
*  if denied and error. Otherwise returns empty string and nil
*  @param r - array of requested mount points
*  @param a - array of allowed mount points
*  @return - mount point string if error and error type
**/
func checkBindPoints(r []string, a []string) (string, error) {
	var valid bool
	for _, requested := range r {
		valid = false
		for _, allowed := range a {
			if strings.HasPrefix(requested, allowed) {
				valid = true
			}
		}
		if !valid {
			return requested, errors.New("Illegal bind")
		}
	}
	return "", nil
}

/* extract all policies from JSON files array and return policy array
*  @param files - FileInfo array for all json policies
*  @return - policy pointer array
**/
func (plugin *mountGuard) extractAllPolicies(files []os.FileInfo) {
	var path strings.Builder
	var policies []policy
	// iterate through all files in directory
	for _, f := range files {
		// make sure file is json
		if strings.HasSuffix(f.Name(), ".json") {

			log.WithFields(log.Fields{
				"Name": f.Name(),
			}).Info("JSON policy found")

			// build full string path to file and append
			path.WriteString(policyDir)
			path.WriteString(f.Name())
			policies = append(policies, extractPolicy(path.String()))
			path.Reset()
		}
	}
	plugin.Policies = policies
}

/* find which policy user matches request user and return policy, if
*  match isnt found a default policy is returned that allows no mounts
*  @param p - policy array to be checked
*  @param user - user name to be matched
*  @return - matching policy as pointer and error if applicable
**/
func matchPolicy(p []policy, user string) (policy, error) {
	defPolicy := policy{"", []string{}}

	for _, pol := range p {
		if pol.User == user {
			return pol, nil
		}
	}
	return defPolicy, errors.New("No match found, using default policy")
}

/* required function for plugin auth flow, implemented over plugin struct so docker
*  can call it over socket
*  @param req - the docker authorization request object see API for details
*  @return - auth.Response object to allow or disallow API calls to daemon
 */
func (plugin *mountGuard) AuthZReq(req authorization.Request) authorization.Response {

	if req.RequestBody != nil {
		// extract request body into config structure
		body := &configWrapper{}
		json.NewDecoder(bytes.NewReader(req.RequestBody)).Decode(body)

		policyMatch, err := matchPolicy(plugin.Policies, body.User)
		handleErr(err, "Matching policy")

		// log current policy and allowed mounts
		log.WithFields(log.Fields{
			"User": policyMatch.User,
		}).Info("Policy User")

		for _, element := range policyMatch.AllowedMounts {
			log.WithFields(log.Fields{
				"Mount": element,
			}).Info("Allowed Mounts")
		}

		mnt, err := checkBindPoints(body.HostConfig.Binds, policyMatch.AllowedMounts)
		if err != nil {
			log.WithFields(log.Fields{
				"Mount": mnt,
				"User":  policyMatch.User,
			}).Warn("Illegal mount request")
			return authorization.Response{
				Allow: false,
				Msg:   fmt.Sprintf("Illegal mount request made (User: '%s' Mount: '%s'", policyMatch.User, mnt),
			}
		}
	}
	return authorization.Response{Allow: true}
}

/* always allow requests from the server */
func (plugin *mountGuard) AuthZRes(req authorization.Request) authorization.Response {

	return authorization.Response{Allow: true}
}

/* basic output for response parameters
*  nothing special here - TODO make pretty or possibly delete
**/
func logResponse(req authorization.Request) {
	if debug {
		body := &configWrapper{}
		json.NewDecoder(bytes.NewReader(req.RequestBody)).Decode(body)

		log.WithFields(log.Fields{
			"URI": req.RequestURI,
		}).Info("URI field")

		log.WithFields(log.Fields{
			"Image": body.Image,
		}).Info("Struct-contents")

		log.WithFields(log.Fields{
			"User": body.User,
		}).Info("Struct-contents")
	}
}

/* simple error handler wrapper for neater err handling and logging with levels
*  @param err - our error type to be processed
*  @param name - string value we use to format the output message
**/
func handleErr(err error, name string) {
	var fail strings.Builder
	var succ strings.Builder
	fail.WriteString(name)
	succ.WriteString(name)
	fail.WriteString(" failure")
	succ.WriteString(" success")
	if err != nil {
		log.WithFields(log.Fields{"Error": err}).Warn(fail.String())
		return
	}
	log.Info(succ.String())
}
