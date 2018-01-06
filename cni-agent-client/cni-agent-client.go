package cniagentclient

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/OpenPlatformSDN/nuage-oci-plugin/config"
	"github.com/golang/glog"

	"github.com/OpenPlatformSDN/nuage-cni/agent/types"
)

var (
	Client     *http.Client
	ServerPort string
)

func InitClient(conf *config.Config) error {

	// Pick up Agent server port from startup configuration
	ServerPort = conf.AgentServerConfig.ServerPort

	certPool := x509.NewCertPool()

	if pemData, err := ioutil.ReadFile(conf.AgentServerConfig.CaFile); err != nil {
		err = fmt.Errorf("Error loading CNI agent server CA certificate data from: %s. Error: %s", conf.AgentServerConfig.CaFile, err)
		glog.Error(err)
		return err
	} else {
		certPool.AppendCertsFromPEM(pemData)
	}

	// configure a TLS client to use those certificates
	Client = new(http.Client)
	*Client = http.Client{
		Transport: &http.Transport{
			MaxIdleConns:    types.MAX_CONNS,
			IdleConnTimeout: types.MAX_IDLE,
			TLSClientConfig: &tls.Config{
				RootCAs: certPool,
				// InsecureSkipVerify: true, // In case we want to skip server verification
			},
		},
	}

	return nil
}
