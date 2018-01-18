package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/golang/glog"
	"github.com/nuagenetworks/vspk-go/vspk"
	"github.com/satori/go.uuid"

	agent "github.com/OpenPlatformSDN/nuage-cni/agent/client"
	"github.com/OpenPlatformSDN/nuage-oci-plugin/config"
	"github.com/OpenPlatformSDN/nuage-oci-plugin/runc"
)

////
//// Definitions for runc container metadata -- expressed as Annotations -- that is relevant for Nuage. We define those (so far):
////

//     "annotations": {
//                "io.nuage.Enterprise": "Enterprise Name",
//                "io.nuage.Domain": "Domain Name",
//                "io.nuage.Zone": "Zone Name"
//                "io.nuage.Subnet": "Subnet Name",
//        }

const (
	keyEnterprise = "io.nuage.Enterprise"
	keyDomain     = "io.nuage.Domain"
	keyZone       = "io.nuage.Zone"
	keySubnet     = "io.nuage.Subnet"
)

// vspk.Container.OrchestrationID for runc containers.
const (
	orchestrationID = "runc"
)

// At which glog error level we do logging by default
const errorLogLevel = "INFO"

var (
	Config         = new(config.Config)
	UseNetPolicies = false
)

////////
////////
////////

func myflags() {

	flag.CommandLine.StringVar(&Config.ConfigFile, "config",
		"./nuage-oci-plugin-config.yaml", "configuration file for Nuage OCI plugin. If this file is specified, all remaining arguments will be ignored")

	// Agent Server flags
	flag.CommandLine.StringVar(&Config.AgentServerConfig.ServerPort, "serverport",
		"7443", "Nuage OCI agent server port")

	flag.CommandLine.StringVar(&Config.AgentServerConfig.CaFile, "cafile",
		"/opt/nuage/etc/ca.crt", "Nuage OCI agent server CA certificate")

	// Set the values for log_dir and logtostderr.  Because this happens before flag.Parse(), cli arguments will override these.
	// Also set the DefValue parameter so -help shows the new defaults.
	// XXX - Make sure "glog" package is imported at this point, otherwise this will panic
	flag.CommandLine.Lookup("log_dir").DefValue = fmt.Sprintf("/var/log/%s", path.Base(os.Args[0]))
	flag.CommandLine.Lookup("logtostderr").DefValue = "false"
	flag.CommandLine.Lookup("stderrthreshold").DefValue = errorLogLevel

	// XXX - skip os.Args[1] i.e. the subcommand
	flag.CommandLine.Parse(os.Args[2:])

	// Set log_dir -- either to given value or to the default + create the directory
	if mylogdir := flag.CommandLine.Lookup("log_dir").Value.String(); mylogdir != "" {
		os.MkdirAll(mylogdir, os.ModePerm)
	} else { // set it to default log_dir value
		flag.CommandLine.Lookup("log_dir").Value.Set(flag.CommandLine.Lookup("log_dir").DefValue)
		os.MkdirAll(flag.CommandLine.Lookup("log_dir").DefValue, os.ModePerm)
	}
}

// Silly little wrapper around os.Exit(). Needed since os.Exit() does not honor defer calls and glog.Fatalf() looks ugly _and_ does not flush the logs.
func osExit(context string, err error) {
	glog.Errorf("%s: %s", context, err)
	glog.Flush()
	os.Exit(255)
}

func main() {

	if os.Args[1] != "prestart" && os.Args[1] != "poststart" {
		osExit("Not a valid command", errors.New(os.Args[1]))
	}

	myflags()

	// Flush logs upon exit
	defer glog.Flush()

	glog.Infof("===> Starting %s...", path.Base(os.Args[0]))

	if err := config.LoadConfig(Config); err != nil {
		osExit("Cannot read configuration file", err)
	}

	if err := agent.InitClient(Config.AgentServerConfig); err != nil {
		osExit("Cannot initialize Nuage CNI Agent client", err)
	}

	// Get container state
	cstate, err := runc.ReadState()

	if err != nil {
		osExit("Error reading container state", err)
	}

	switch os.Args[1] {
	case "prestart":
		////
		//// Running as poststat hook -- top-side of split-activation
		////

		//// Create a vspk.Container and submit it to the agent server. This is a placeholder
		//// XXX -- Since this is a placeholder, below we abuse some of the fields in "vspk.Container" structure by passing (yet) unverified container metadata to the agent server in those fields

		container := new(vspk.Container)

		// Container.Name :  Use runc container ID as "Name". For cri-o containers "ID" is actually an UUID (64 characters)
		container.Name = cstate.ID

		// Container.UUID : If it's cri-o (256 bits / 64 characters) use it. If not, generate a new one.
		if len(cstate.ID) == 64 {
			container.UUID = cstate.ID
		} else {
			// Generate a random VSD UUID from two UUID v4 (RFC 4122)
			container.UUID = strings.Replace(uuid.Must(uuid.NewV4()).String(), "-", "", -1) + strings.Replace(uuid.Must(uuid.NewV4()).String(), "-", "", -1)
		}

		// Container.OrchestrationID : Pre-defined string
		container.OrchestrationID = orchestrationID

		// container.EnterpriseName : Pick it from metadata, if present (null string otherwise). It is up to the Agent server to valide it
		container.EnterpriseName = cstate.Annotations[keyEnterprise]

		// XXX -- container.DomainIDs : Use this field to encode the Domain name from container metadata (if present)
		// var myd []interface{}
		// container.DomainIDs = append(myd, cstate.Annotations[keyDomain])
		container.DomainIDs = []interface{}{cstate.Annotations[keyDomain]}

		// XXX -- container.ZoneIDs : Use this field to encode the Zone name from container metadata (if present)
		container.ZoneIDs = []interface{}{cstate.Annotations[keyZone]}

		// XXX -- container.SubnetIDs : Use this field to encode the Subnet name from container metadata (if present)
		container.SubnetIDs = []interface{}{cstate.Annotations[keySubnet]}

		// JSON pretty-print it
		jsonc, _ := json.MarshalIndent(container, "", "\t")
		glog.Infof("Container: %s", string(jsonc))

		// Get (local) hostname as reported by the kernel, ignore errors
		localhn, _ := os.Hostname()
		// Submit it to the Agent server
		if err := agent.ContainerPUT(localhn, container); err != nil {
			osExit("Error submitting container to the Agent server", err)
		}

		glog.Infof("Successfully submitted to the Agent sever container with Name: %s", container.Name)

	case "poststart":
		////
		//// Running as runc poststart hook -- bottom-side of split-activation
		////
	}

}
