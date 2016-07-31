package docker

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"strings"
	"testing"

	"github.com/containerops/dockyard/utils/setting"
	"github.com/containerops/dockyard/utils/signature"
)

var (
	repoName  string   = "busybox"
	repoTag   string   = "latest"
	repoTags  []string = []string{"1.0", "2.0", "3.0"}
	paths     []string = []string{"", "./dockerfile1", "./dockerfile2"}
	repoBase  string   = repoName + ":" + repoTag
	repoDests []string = make([]string, 3)
)

func TestDeleteInit(t *testing.T) {
	for i := 0; i < 3; i++ {
		repoDests[i] = UserName + "/" + repoName + ":" + repoTags[i]
	}

	if err := exec.Command("sudo", DockerBinary, "inspect", repoBase).Run(); err != nil {
		cmd := exec.Command(DockerBinary, "pull", repoBase)
		if out, err := ParseCmdCtx(cmd); err != nil {
			t.Fatalf("Pull testing preparation is failed: [Info]%v, [Error]%v", out, err)
		}
	}

	for i := 1; i < 3; i++ {
		cmd := exec.Command("sudo", DockerBinary, "build", "-t", repoDests[i], paths[i])
		if out, err := ParseCmdCtx(cmd); err != nil {
			t.Fatalf("Build %v failed: [Info]%v, [Error]%v", repoTags[i], out, err)
		}
	}

	cmd := exec.Command("sudo", DockerBinary, "tag", "-f", repoBase, Domains+"/"+repoDests[0])
	if out, err := ParseCmdCtx(cmd); err != nil {
		t.Fatalf("Tag %v failed: [Info]%v, [Error]%v", repoDests[0], out, err)
	}

	for i := 1; i < 3; i++ {
		cmd := exec.Command("sudo", DockerBinary, "tag", "-f", repoDests[i], Domains+"/"+repoDests[i])
		if out, err := ParseCmdCtx(cmd); err != nil {
			t.Fatalf("Tag %v failed: [Info]%v, [Error]%v", repoDests[i], out, err)
		}
	}
}

func TestDeleteSingleRepo(t *testing.T) {
	cmd := exec.Command("sudo", DockerBinary, "push", Domains+"/"+repoDests[0])
	if out, err := ParseCmdCtx(cmd); err != nil {
		t.Fatalf("Push testing preparation is failed: [Info]%v, [Error]%v", out, err)
	}

	deleteTag(t, repoTags[0])
}

func TestDeleteMutiRepo(t *testing.T) {
	for _, v := range repoDests {
		cmd := exec.Command("sudo", DockerBinary, "push", Domains+"/"+v)
		if out, err := ParseCmdCtx(cmd); err != nil {
			t.Fatalf("Push testing preparation is failed: [Info]%v, [Error]%v", out, err)
		}
	}

	deleteTag(t, repoTags[0])
	for j := 2; j < len(repoTags); j++ {
		deleteTag(t, repoTags[j])
	}

	cmd := exec.Command("sudo", DockerBinary, "pull", Domains+"/"+repoDests[1])
	if out, err := ParseCmdCtx(cmd); err != nil {
		t.Fatalf("Pull testing preparation is failed: [Info]%v, [Error]%v", out, err)
	}

	deleteTag(t, repoTags[1])
}

func deleteTag(t *testing.T, tag string) {
	repo := UserName + "/" + repoName
	url := fmt.Sprintf(setting.ListenMode+"://%v/v2/%v/%v/manifests/%v", Domains, UserName, repoName, tag)
	body := methodHttp(url, "GET", t)

	digest, err := signature.DigestManifest(body)
	url = fmt.Sprintf(setting.ListenMode+"://%v/v2/%v/%v/manifests/%v", Domains, UserName, repoName, digest)
	out := methodHttp(url, "DELETE", t)

	if strings.Contains(string(out), repo) == true {
		t.Fatalf("Delete manifest failed: [Info]%v", string(out))
	}

	var mnf map[string]interface{}
	if err := json.Unmarshal(body, &mnf); err != nil {
		t.Fatalf("Get manifest failed: [Error]%v", err)
	}
	ret := false
	for k, _ := range mnf {
		if k == "schemaVersion" {
			ret = true
			break
		}
	}
	if ret == false {
		t.Fatalf("Get manifest failed: [Error]%v", err)
	}
	var layerdesc = []string{"", "fsLayers", "layers"}
	var tarsumdesc = []string{"", "blobSum", "digest"}
	schemaVersion := int64(mnf["schemaVersion"].(float64))
	section := layerdesc[schemaVersion]
	item := tarsumdesc[schemaVersion]
	for k := len(mnf[section].([]interface{})) - 1; k >= 0; k-- {
		sha := mnf[section].([]interface{})[k].(map[string]interface{})[item].(string)
		url := fmt.Sprintf(setting.ListenMode+"://%v/v2/%v/%v/blobs/%v", Domains, UserName, repoName, sha)
		out := methodHttp(url, "DELETE", t)

		if strings.Contains(string(out), sha) == true {
			t.Fatalf("Delete blob failed: [Info]%v", string(out))
		}
	}
	schemaVersion = int64(mnf["schemaVersion"].(float64))
	if schemaVersion == 2 {
		sha := mnf["config"].(map[string]interface{})["digest"].(string)
		url := fmt.Sprintf(setting.ListenMode+"://%v/v2/%v/%v/blobs/%v", Domains, UserName, repoName, sha)
		out := methodHttp(url, "DELETE", t)

		if strings.Contains(string(out), sha) == true {
			t.Fatalf("Delete blob failed: [Info]%v", out)
		}
	}
}

func methodHttp(url, meth string, t *testing.T) []byte {
	req, err := http.NewRequest(meth, url, nil)
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
	out, _ := ioutil.ReadAll(resp.Body)
	return out
}
