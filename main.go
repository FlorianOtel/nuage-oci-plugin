package main

import (
	"flag"
	"fmt"
	"os"
	"path"

	"github.com/golang/glog"

	agentclient "github.com/OpenPlatformSDN/nuage-oci-plugin/cni-agent-client"
	"github.com/OpenPlatformSDN/nuage-oci-plugin/config"
	"github.com/OpenPlatformSDN/nuage-oci-plugin/runc"
)

const errorLogLevel = "INFO"

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
	flagSet.Lookup("log_dir").DefValue = fmt.Sprintf("/var/log/%s", path.Base(os.Args[0]))
	flagSet.Lookup("logtostderr").DefValue = "false"
	flagSet.Lookup("stderrthreshold").DefValue = errorLogLevel

	flag.Parse()

	// Set log_dir -- either to given value or to the default + create the directory
	if mylogdir := flag.CommandLine.Lookup("log_dir").Value.String(); mylogdir != "" {
		os.MkdirAll(mylogdir, os.ModePerm)
	} else { // set it to default log_dir value
		flag.CommandLine.Lookup("log_dir").Value.Set(flag.CommandLine.Lookup("log_dir").DefValue)
		os.MkdirAll(flag.CommandLine.Lookup("log_dir").DefValue, os.ModePerm)
	}
}

func main() {

	Config = new(config.Config)

	Flags(Config, flag.CommandLine)

	// Flush logs upon exit
	defer glog.Flush()

	glog.Infof("===> Starting %s...", path.Base(os.Args[0]))

	if err := config.LoadConfig(Config); err != nil {
		glog.Errorf("Cannot read configuration file: %s", err)
		glog.Flush() // os.Exit() does not honor defer calls
		os.Exit(255)
	}

	if err := agentclient.InitClient(Config); err != nil {
		glog.Errorf("Cannot read configuration file: %s", err)
		glog.Flush() // os.Exit() does not honor defer calls
		os.Exit(255)
	}

	if err := runc.StillTBD(); err != nil {
		glog.Errorf("runc plugin error: %s", err)
		glog.Flush() // os.Exit() does not honor defer calls
		os.Exit(255)
	}

}
