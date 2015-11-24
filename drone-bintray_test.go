package main

/*
Unit test for drone-bintray plugin.  This test uploads a single
artifact to Bintray.  Before running, update testData/testConfig.json
with suitable Bintray credentials and repo name.
*/

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/drone/drone-go/plugin"
)

func TestDroneBintray(t *testing.T) {
	configFilename := filepath.Join("testData", "testConfig.json")
	configData, err := ioutil.ReadFile(configFilename)
	if err != nil {
		fmt.Printf("Can't read %s: %s", configFilename, err.Error())
		os.Exit(1)
	}
	plugin.Stdin = plugin.NewParamSet(bytes.NewBuffer(configData))
	main()
}
