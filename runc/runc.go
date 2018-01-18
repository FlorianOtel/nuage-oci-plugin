package runc

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/golang/glog"
	specs "github.com/opencontainers/runtime-spec/specs-go"
)

// Decodes stdin as container state. Assumes we run as a OCI container (prestart) hook
func ReadState() (*specs.State, error) {
	cstate := new(specs.State)

	// Read container state from stdin
	b, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return nil, fmt.Errorf("Reading container state from stdin failed: %v", err)
	}

	// Umarshal the hook state
	if err := json.Unmarshal(b, cstate); err != nil {
		return nil, fmt.Errorf("Unmarshal stdin as specs.State failed: %v", err)
	}

	glog.Infof("Container State: %#v", *cstate)
	return cstate, nil

}
