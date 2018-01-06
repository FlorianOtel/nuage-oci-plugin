package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"

	"github.com/golang/glog"

	agentclient "github.com/OpenPlatformSDN/nuage-oci-plugin/cni-agent-client"
	"github.com/OpenPlatformSDN/nuage-oci-plugin/config"
)

const errorLogLevel = 2

var (
	Config         *config.Config
	UseNetPolicies = false
)

////////
////////
////////

func Flags(conf *config.Config, flagSet *flag.FlagSet) {
	flagSet.StringVar(&conf.ConfigFile, "config",
		"./nuage-oci-plugin-config.yaml", "configuration file for Nuage OCI plugin. If this file is specified, all remaining arguments will be ignored")

	// Agent Server flags
	flagSet.StringVar(&conf.AgentServerConfig.ServerPort, "serverport",
		"7443", "Nuage OCI agent server port")

	flagSet.StringVar(&conf.AgentServerConfig.CaFile, "cafile",
		"/opt/nuage/etc/ca.crt", "Nuage OCI agent server CA certificate")

	// Set the values for log_dir and logtostderr.  Because this happens before flag.Parse(), cli arguments will override these.
	// Also set the DefValue parameter so -help shows the new defaults.
	// XXX - Make sure "glog" package is imported at this point, otherwise this will panic
	log_dir := flagSet.Lookup("log_dir")
	log_dir.Value.Set(fmt.Sprintf("/var/log/%s", path.Base(os.Args[0])))
	log_dir.DefValue = fmt.Sprintf("/var/log/%s", path.Base(os.Args[0]))
	logtostderr := flagSet.Lookup("logtostderr")
	logtostderr.Value.Set("false")
	logtostderr.DefValue = "false"
	stderrlogthreshold := flagSet.Lookup("stderrthreshold")
	stderrlogthreshold.Value.Set("2")
	stderrlogthreshold.DefValue = "2"
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {

	Config = new(config.Config)

	Flags(Config, flag.CommandLine)
	flag.Parse()

	if len(os.Args) == 1 { // With no arguments, print default usage
		flag.PrintDefaults()
		os.Exit(0)
	}
	// Flush the logs upon exit
	defer glog.Flush()

	glog.Infof("===> Starting %s...", path.Base(os.Args[0]))

	if err := config.LoadConfig(Config); err != nil {
		glog.Errorf("Cannot read configuration file: %s", err)
		os.Exit(255)
	}

	if err := agentclient.InitClient(Config); err != nil {
		glog.Errorf("Cannot read configuration file: %s", err)
		os.Exit(255)
	}

	select {}
}
