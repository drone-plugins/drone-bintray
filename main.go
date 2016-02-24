/*
Drone plugin to upload one or more packages to Bintray.
See DOCS.md for usage.

Author: David Tootill November 2015 (GitHub tooda02)
*/
package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"os"
	"path"
	"strings"

	"github.com/drone/drone-go/drone"
	"github.com/drone/drone-go/plugin"
)

const defaultHost = "https://api.bintray.com"

type Bintray struct {
	Username  string     `json:"username"`
	APIKey    string     `json:"api_key"`
	Branch    string     `json:"branch"`
	Host      string     `json:"host"`
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
	buildCommit string
	bintray   Bintray
)

type MessageText struct {
	Message string `json:"message"`
}

func main() {
	fmt.Printf("Drone Bintray Plugin built from %s\n", buildCommit)

	var workspace = drone.Workspace{}

	plugin.Param("workspace", &workspace)
	plugin.Param("vargs", &bintray)
	if err := plugin.Parse(); err != nil {
		fmt.Printf("ERROR Can't parse yaml config: %s", err.Error())
		os.Exit(1)
	}

	if bintray.Host == "" {
		bintray.Host = defaultHost
	}

	if bintray.Debug {
		saveApikey := bintray.APIKey
		bintray.APIKey = "******"
		fmt.Printf("DEBUG plugin input:\n%#v\n%#v\n", workspace, bintray)
		bintray.APIKey = saveApikey
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
		artifact.Upload(workspace.Path)
	}
}

// Upload a package to Bintray
func (this *Artifact) Upload(filename string) {
	// Build file upload request

	filepath := path.Join(filename, this.File)
	file, err := os.Open(filepath)
	if err != nil {
		fmt.Printf("Unable to open input file %s: %s\n", filepath, err.Error())
		os.Exit(1)
	}
	req, err := http.NewRequest("PUT", this.getEndpoint(), file)
	if err != nil {
		fmt.Printf("Unable to build REST request: %s\n", err.Error())
		os.Exit(1)
	}
	req.SetBasicAuth(bintray.Username, bintray.APIKey)
	req.Header.Add("X-Bintray-Override", boolToString(this.Override))
	req.Header.Add("X-Bintray-Publish", boolToString(this.Publish))
	if this.Type == "Debian" {
		this.addDebianHeaders(req)
	}

	// Set up an HTTP client with a bundled root CA certificate (borrowed from Ubuntu 14.04)
	// This is necessary because the required certificate is missing from the root image and
	// without it the upload fails with "x509: failed to load system roots and no roots provided"

	client := http.Client{}
	pool := x509.NewCertPool()
	if pemCerts, err := ioutil.ReadFile("/etc/ssl/certs/ca-certificates.crt"); err != nil {
		fmt.Printf("Unable to read ca-certificates.crt: %s\n", err.Error())
		os.Exit(1)
	} else {
		pool.AppendCertsFromPEM(pemCerts)
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:            pool,
				InsecureSkipVerify: bintray.Insecure,
			},
		}
	}
	if bintray.Debug {
		dumpRequest("DEBUG HTTP Request", req)
	}

	// Execute the upload request and format the response

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Upload request failed: %s\n", err.Error())
		dumpRequest("Failing request", req)
		os.Exit(1)
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Unable to read request response: %s\n", err.Error())
		dumpRequest("Failing request", req)
		os.Exit(1)
	}
	messageText := new(MessageText)
	if len(respBody) > 0 {
		json.Unmarshal(respBody, messageText)
	}
	if resp.StatusCode > 299 {
		errorText := fmt.Sprintf("Error %d", resp.StatusCode)
		httpErrorText := http.StatusText(resp.StatusCode)
		if len(httpErrorText) > 0 {
			errorText += " " + httpErrorText
		}
		if len(messageText.Message) > 0 {
			errorText += " - " + messageText.Message
		}
		fmt.Printf("%s\n", errorText)
		dumpRequest("Failing request", req)
		os.Exit(1)
	}

	if len(messageText.Message) == 0 && len(respBody) > 0 {
		messageText.Message = string(respBody)
	}
	fmt.Printf("Result: %s\n", messageText.Message)
	if messageText.Message != "success" {
		if this.Override || !strings.Contains(messageText.Message, "already exists") {
			dumpRequest("\nRequest was:", req)
			os.Exit(1)
		}
	}
}

// Add headers required to upload a Debian package
func (this *Artifact) addDebianHeaders(req *http.Request) {
	if len(this.Distr) == 0 || len(this.Component) == 0 || len(this.Arch) == 0 {
		fmt.Printf("ERROR Cannot process package %s - Missing Debian argument(s):\n", this.Artifact)

		if len(this.Distr) == 0 {
			fmt.Printf("    distr not defined in yaml config\n")
		}

		if len(this.Component) == 0 {
			fmt.Printf("    component not defined in yaml config\n")
		}

		if len(this.Arch) == 0 {
			fmt.Printf("    arch not defined in yaml config\n")
		}

		os.Exit(1)
	}

	req.Header.Add("X-Bintray-Debian-Distribution", this.Distr)
	req.Header.Add("X-Bintray-Debian-Component", this.Component)
	req.Header.Add("X-Bintray-Debian-Architecture", strings.Join(this.Arch, ","))
}

// Get the Bintray endpoint corresponding to this package and version
func (this *Artifact) getEndpoint() string {
	if len(this.File) == 0 ||
		len(this.Owner) == 0 ||
		len(this.Repository) == 0 ||
		len(this.Artifact) == 0 ||
		(len(this.Version) == 0 && this.Type != "Maven") ||
		len(this.Target) == 0 {
		fmt.Printf("ERROR Cannot process package %s - Missing argument(s):\n", this.Artifact)

		if len(this.Artifact) == 0 {
			fmt.Printf("    package not defined in yaml config\n")
		}

		if len(this.File) == 0 {
			fmt.Printf("    file not defined in yaml config\n")
		}

		if len(this.Owner) == 0 {
			fmt.Printf("    owner not defined in yaml config\n")
		}

		if len(this.Repository) == 0 {
			fmt.Printf("    repository not defined in yaml config\n")
		}

		if len(this.Version) == 0 && this.Type != "Maven" {
			fmt.Printf("    version not defined in yaml config\n")
		}

		if len(this.Target) == 0 {
			fmt.Printf("    target not defined in yaml config\n")
		}

		os.Exit(1)
	}

	contentType := "content"
	if this.Type == "Maven" {
		contentType = "maven"
	}

	endpoint := fmt.Sprintf("%s/%s/%s/%s/%s/", bintray.Host, contentType, this.Owner, this.Repository, this.Artifact)
	if len(this.Version) > 0 && this.Type != "Maven" {
		endpoint = fmt.Sprintf("%s%s/", endpoint, this.Version)
	}
	if len(bintray.Branch) == 0 || bintray.Branch == "master" {
		return endpoint + this.Target
	}
	return fmt.Sprintf("%stest/%s/%s", endpoint, bintray.Branch, this.Target)
}

// Dump an HTTP request to stdout, hiding the authorization tag
func dumpRequest(prefix string, req *http.Request) {
	fmt.Printf("%s:\n", prefix)
	if dumpedRequest, err := httputil.DumpRequestOut(req, false); err != nil {
		fmt.Printf("  %s\n", err.Error())
	} else {
		for _, line := range strings.Split(string(dumpedRequest), "\n") {
			line = strings.TrimSpace(line)
			if len(line) > 0 {
				if strings.HasPrefix(line, "Authorization:") {
					line = "Authorization: Basic xxxxxxxxxx"
				}
				fmt.Printf("    %s\n", line)
			}
		}
	}
}

// Convert a boolean value to string "1" or "0"
func boolToString(val bool) string {
	if val {
		return "1"
	} else {
		return "0"
	}
}
