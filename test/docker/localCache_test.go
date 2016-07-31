package docker

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/containerops/dockyard/utils/setting"
)

func TestCacheInit(t *testing.T) {
	repoBase := "busybox:latest"
	repoDest := Domains + "/" + UserName + "/" + repoBase

	if err := exec.Command(DockerBinary, "inspect", repoBase).Run(); err != nil {
		cmd := exec.Command(DockerBinary, "pull", repoBase)
		if out, err := ParseCmdCtx(cmd); err != nil {
			t.Fatalf("Pull testing preparation is failed: [Info]%v, [Error]%v", out, err)
		}
	}

	cmd := exec.Command(DockerBinary, "tag", "-f", repoBase, repoDest)
	if out, err := ParseCmdCtx(cmd); err != nil {
		t.Fatalf("Tag %v failed: [Info]%v, [Error]%v", repoBase, out, err)
	}
	cmd = exec.Command(DockerBinary, "push", repoDest)
	if out, err := ParseCmdCtx(cmd); err != nil {
		t.Fatalf("Push testing preparation is failed: [Info]%v, [Error]%v", out, err)
	}
}

func TestCache(t *testing.T) {
	repoName := "busybox"
	repoTag := "latest"

	url := fmt.Sprintf(setting.ListenMode+"://%v/v2/%v/%v/manifests/%v", Domains, UserName, repoName, repoTag)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Error(err)
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Do(req)
	if err != nil {
		t.Error(err)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	t.Log(resp)

	var mnf map[string]interface{}
	if err := json.Unmarshal(body, &mnf); err != nil {
		t.Fatalf("Get manifest failed: [Error]%v", err)
	}
	var layerdesc = []string{"", "fsLayers", "layers"}
	var tarsumdesc = []string{"", "blobSum", "digest"}
	schemaVersion := int64(mnf["schemaVersion"].(float64))
	section := layerdesc[schemaVersion]
	item := tarsumdesc[schemaVersion]
	for k := len(mnf[section].([]interface{})) - 1; k >= 0; k-- {
		blobsum := mnf[section].([]interface{})[k].(map[string]interface{})[item].(string)
		tarsum := strings.Split(blobsum, ":")[1]
		path := fmt.Sprintf("../data/tarsum/" + tarsum + "/layer")

		if !setting.Cachable {
			if _, err = os.Stat(path); err == nil {
				t.Fatalf("localCache is not deleted")
			}
		} else {
			if _, err = os.Stat(path); err != nil {
				t.Fatalf("localCache is not existed")
			}
		}

	}
}
