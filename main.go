/*
Drone plugin to upload one or more packages to Bintray.
See DOCS.md for usage.

Author: David Tootill November 2015 (GitHub tooda02)
*/
package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/drone/drone-go/drone"
	"github.com/drone/drone-go/plugin"
)

const bintray_endpoint = "https://api.bintray.com/%s/%s/%s/%s/"

type Bintray struct {
	Username  string     `json:"username"`
	ApiKey    string     `json:"api_key"`
	Branch    string     `json:"branch"`
	Debug     bool       `json:"debug"`
	Insecure  bool       `json:"insecure"`
	Artifacts []Artifact `json:"artifacts"`
}

type Artifact struct {
	File       string      `json:"file"`
	Type       string      `json:"type"`
	Owner      string      `json:"owner"`
	Repository string      `json:"repository"`
	Artifact   string      `json:"package"`
	Target     string      `json:"target"`
	Distr      string      `json:"distr,omitempty"`
	Component  string      `json:"component,omitempty"`
	Arch       []string    `json:"arch,omitempty"`
	Publish    bool        `json:"publish,omitempty"`
	Override   bool        `json:"override,omitempty"`
	Versioni   interface{} `json:"version"`
	Version    string
}

var (
	version   string = "1.0"
	buildDate string // Filled from build
	bintray   Bintray
)

func main() {
	fmt.Printf("\nDrone Bintray plugin version %s %s\n", version, buildDate)
	var workspace = drone.Workspace{}

	plugin.Param("workspace", &workspace)
	plugin.Param("vargs", &bintray)
	if err := plugin.Parse(); err != nil {
		fmt.Printf("ERROR Can't parse yaml config: %s", err.Error())
		os.Exit(1)
	}

	if bintray.Debug {
		saveApikey := bintray.ApiKey
		bintray.ApiKey = "******"
		fmt.Printf("DEBUG plugin input:\n%#v\n%#v\n", workspace, bintray)
		bintray.ApiKey = saveApikey
	}
	if len(bintray.Branch) == 0 || bintray.Branch == "master" {
		fmt.Printf("\nPublishing %d artifacts to Bintray for user %s\n", len(bintray.Artifacts), bintray.Username)
	} else {
		fmt.Printf("\nPublishing %d artifacts on branch %s to Bintray for user %s\n", len(bintray.Artifacts), bintray.Branch, bintray.Username)
	}
	for i, artifact := range bintray.Artifacts {
		artifact.Version = fmt.Sprintf("%v", artifact.Versioni)
		fmt.Printf("\nUploading file %d %s to %s\n", i+1,
			artifact.File, artifact.getEndpoint())
		artifact.Write(workspace.Path)
	}
}

// Upload a package to Bintray
func (this *Artifact) Write(path string) {
	if len(this.File) == 0 ||
		len(this.Owner) == 0 ||
		len(this.Repository) == 0 ||
		len(this.Artifact) == 0 ||
		(len(this.Version) == 0 && this.Type != "Maven") ||
		len(this.Target) == 0 {
		fmt.Printf("Bintray Plugin: Missing argument(s)\n\n")

		if len(this.Artifact) == 0 {
			fmt.Printf("\tpackage not defined in yaml config")
		}

		if len(this.File) == 0 {
			fmt.Printf("\tpackage %s: file not defined in yaml config", this.Artifact)
		}

		if len(this.Owner) == 0 {
			fmt.Printf("\tpackage %s: owner not defined in yaml config", this.Artifact)
		}

		if len(this.Repository) == 0 {
			fmt.Printf("\tpackage %s: repository not defined in yaml config", this.Artifact)
		}

		if len(this.Version) == 0 && this.Type != "Maven" {
			fmt.Printf("\tpackage %s: version not defined in yaml config", this.Artifact)
		}

		if len(this.Target) == 0 {
			fmt.Printf("\tpackage %s: target not defined in yaml config", this.Artifact)
		}

		os.Exit(1)
	}

	switch this.Type {
	case "Debian":
		this.debUpload(path)
	default:
		this.upload(path)
	}
}

// Upload a Debian package
func (this *Artifact) debUpload(path string) {
	if len(this.Distr) == 0 || len(this.Component) == 0 || len(this.Arch) == 0 {
		fmt.Printf("Bintray Plugin: Missing argument(s)\n\n")

		if len(this.Distr) == 0 {
			fmt.Printf("\tDebian package %s: distr not defined in yaml config", this.Artifact)
		}

		if len(this.Component) == 0 {
			fmt.Printf("\tDebian package %s: component not defined in yaml config", this.Artifact)
		}

		if len(this.Arch) == 0 {
			fmt.Printf("\tDebian package %s: arch not defined in yaml config", this.Artifact)
		}

		os.Exit(1)
	}

	this.execCommand("curl", bintray.ApiKey,
		"-H", fmt.Sprintf("X-Bintray-Debian-Distribution: %s", this.Distr),
		"-H", fmt.Sprintf("X-Bintray-Debian-Component: %s", this.Component),
		"-H", fmt.Sprintf("X-Bintray-Debian-Architecture: %s", strings.Join(this.Arch, ",")),
		"-H", fmt.Sprintf("X-Bintray-Override: %d", boolToInt(this.Override)),
		"-H", fmt.Sprintf("X-Bintray-Publish: %d", boolToInt(this.Publish)),
		"-T", fmt.Sprintf("%s/%s", path, this.File),
		fmt.Sprintf("-u%s:%s", bintray.Username, bintray.ApiKey),
		this.getEndpoint())
}

// Upload a non-Debian package
func (this *Artifact) upload(path string) {
	this.execCommand("curl", bintray.ApiKey,
		"-H", fmt.Sprintf("X-Bintray-Override: %d", boolToInt(this.Override)),
		"-H", fmt.Sprintf("X-Bintray-Publish: %d", boolToInt(this.Publish)),
		"-T", fmt.Sprintf("%s/%s", path, this.File),
		fmt.Sprintf("-u%s:%s", bintray.Username, bintray.ApiKey),
		this.getEndpoint())
}

// Get the Bintray endpoint corresponding to this package and version
func (this *Artifact) getEndpoint() string {
	contentType := "content"
	if this.Type == "Maven" {
		contentType = "maven"
	}
	endpoint := fmt.Sprintf(bintray_endpoint, contentType, this.Owner, this.Repository, this.Artifact)
	if len(this.Version) > 0 && this.Type != "Maven" {
		endpoint = fmt.Sprintf("%s%s/", endpoint, this.Version)
	}
	if len(bintray.Branch) == 0 || bintray.Branch == "master" {
		return endpoint + this.Target
	}
	return fmt.Sprintf("%stest/%s/%s", endpoint, bintray.Branch, this.Target)
}

// Run a command and test result
func (this *Artifact) execCommand(c string, api_key string, args ...string) {
	if bintray.Insecure {
		args = append(args, "-k")
	}
	cmd := exec.Command(c, args...)
	output, err := cmd.CombinedOutput()
	outputString := string(output)
	message := ""
	rxMessage := regexp.MustCompile(`"message":"([^"]*)"`)
	if match := rxMessage.FindStringSubmatch(outputString); match != nil {
		message = match[1]
		fmt.Printf("Result: %s\n", message)
	}
	if err == nil && message != "success" {
		if this.Override || !strings.Contains(message, "already exists") {
			err = fmt.Errorf("Bintray upload failed")
		}
	}
	if err != nil {
		fmt.Printf("%s\n", strings.Replace(strings.Join(cmd.Args, " "), api_key, "******", -1))
		fmt.Printf("%s\n", outputString)
		fmt.Printf("ERROR %s\n", err.Error())
		os.Exit(1)
	} else if bintray.Debug {
		fmt.Printf("DEBUG command executed:\n%s\n", strings.Replace(strings.Join(cmd.Args, " "), api_key, "******", -1))
		fmt.Printf("%s\n", outputString)
	}
	return
}

func boolToInt(val bool) int {
	if val {
		return 1
	} else {
		return 0
	}
}
