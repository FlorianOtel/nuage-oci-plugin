package runc

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/golang/glog"
	specs "github.com/opencontainers/runtime-spec/specs-go"
)

func StillTBD() error {
	// Get container state -- assumes we're running as a hook
	hookdata, err := readState()
	if err != nil {
		glog.Errorf("Reading container state error: %s", err)
		return err
	}

	// Just JSON pretty-print it
	jsonhd, _ := json.MarshalIndent(hookdata, "", "\t")

	glog.Infof("Container state as hook data: %s", string(jsonhd))
	return nil
}

// Decodes stdin as container state
func readState() (cstate specs.State, err error) {
	// Read container state from stdin
	b, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return cstate, fmt.Errorf("Reading container state from stdin failed: %v", err)
	}

	// Umarshal the hook state
	if err := json.Unmarshal(b, &cstate); err != nil {
		return cstate, fmt.Errorf("Unmarshal stdin as specs.State failed: %v", err)
	}

	glog.Infof("Container State: %#v", cstate)
	return cstate, nil

}
