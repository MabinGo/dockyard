package docker

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"testing"

	"github.com/containerops/dockyard/utils/setting"
	"github.com/containerops/dockyard/utils/signature"
)

func TestErrorInit(t *testing.T) {
	repoBase := "busybox:latest"
	repoDest := Domains + "/" + UserName + "/" + repoBase

	if err := exec.Command("sudo", DockerBinary, "inspect", repoBase).Run(); err != nil {
		cmd := exec.Command(DockerBinary, "pull", repoBase)
		if out, err := ParseCmdCtx(cmd); err != nil {
			t.Fatalf("Pull testing preparation is failed: [Info]%v, [Error]%v", out, err)
		}
	}
	cmd := exec.Command("sudo", DockerBinary, "tag", "-f", repoBase, repoDest)
	if out, err := ParseCmdCtx(cmd); err != nil {
		t.Fatalf("Tag %v failed: [Info]%v, [Error]%v", repoBase, out, err)
	}
	cmd = exec.Command("sudo", DockerBinary, "push", repoDest)
	if out, err := ParseCmdCtx(cmd); err != nil {
		t.Fatalf("Pull testing preparation is failed: [Info]%v, [Error]%v", out, err)
	}

}

func TestDeleteError(t *testing.T) {
	repoName := "busybox"
	tag := "latest"

	url := fmt.Sprintf("%v://%v/v2/%v/%v/manifests/%v", setting.ListenMode, Domains, UserName, repoName, tag)
	var digest string
	var output []byte
	var out []byte
	var err error

	output, err = exec.Command("sudo", "curl", "-X", "GET", url).Output()
	if err != nil {
		t.Fatalf("Get manifest failed: [Info]%v, [Error]%v", output, err)
	}
	digest, err = signature.DigestManifest(output)
	url = fmt.Sprintf("%v://%v/v2/%v/%v/manifests/%v", setting.ListenMode, Domains, UserName, repoName, "no")
	out, err = exec.Command("sudo", "curl", "-X", "DELETE", url).Output()
	if err != nil {
		t.Fatalf("Delete tag failed: [Info]%v, [Error]%v", out, err)
	}
	if strings.Contains(string(out), "DIGEST_INVALID") == false {
		t.Fatalf("Delete manifest failed: [Info]%v", string(out))
	}
	url = fmt.Sprintf("%v://%v/v2/%v/%v/manifests/%v", setting.ListenMode, Domains, UserName, repoName, digest)
	out, err = exec.Command("sudo", "curl", "-X", "DELETE", url).Output()
	if err != nil {
		t.Fatalf("Delete tag failed: [Info]%v, [Error]%v", out, err)
	}
	if strings.Contains(string(out), "errors") == true {
		t.Fatalf("Delete manifest failed: [Info]%v", string(out))
	}

	url = fmt.Sprintf("%v://%v/v2/%v/%v/manifests/%v", setting.ListenMode, Domains, UserName, repoName, digest)
	out, err = exec.Command("sudo", "curl", "-X", "DELETE", url).Output()
	if err != nil {
		t.Fatalf("Delete tag failed: [Info]%v, [Error]%v", out, err)
	}
	if strings.Contains(string(out), "MANIFEST_UNKNOWN") == false {
		t.Fatalf("Delete manifest failed: [Info]%v", string(out))
	}

	var mnf map[string]interface{}
	if err = json.Unmarshal(output, &mnf); err != nil {
		t.Fatalf("Get manifest failed: [Error]%v", err)
	}
	ret := false
	for k := range mnf {
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
		url := fmt.Sprintf("%v://%v/v2/%v/%v/blobs/%v", setting.ListenMode, Domains, UserName, repoName, sha)
		out, err = exec.Command("sudo", "curl", "-X", "DELETE", url).Output()

		if err != nil {
			t.Fatalf("Delete tag failed: [Info]%v, [Error]%v", out, err)
		}
		if strings.Contains(string(out), "errors") == true {
			t.Fatalf("Delete blob failed: [Info]%v", out)
		}

		url = fmt.Sprintf("%v://%v/v2/%v/%v/blobs/test", setting.ListenMode, Domains, UserName, repoName)
		out, err = exec.Command("sudo", "curl", "-X", "DELETE", url).Output()

		if err != nil {
			t.Fatalf("Delete tag failed: [Info]%v, [Error]%v", out, err)
		}
		if strings.Contains(string(out), "DIGEST_INVALID") == false {
			t.Fatalf("Delete blob failed: [Info]%v", out)
		}
	}
	schemaVersion = int64(mnf["schemaVersion"].(float64))
	if schemaVersion == 2 {
		sha := mnf["config"].(map[string]interface{})["digest"].(string)
		url := fmt.Sprintf("%v://%v/v2/%v/%v/blobs/%v", setting.ListenMode, Domains, UserName, repoName, sha)
		out, err = exec.Command("sudo", "curl", "-X", "DELETE", url).Output()

		if err != nil {
			t.Fatalf("Delete tag failed: [Info]%v, [Error]%v", out, err)
		}
		if strings.Contains(string(out), sha) == true {
			t.Fatalf("Delete blob failed: [Info]%v", out)
		}

		out, err = exec.Command("sudo", "curl", "-X", "DELETE", url).Output()

		if err != nil {
			t.Fatalf("Delete tag failed: [Info]%v, [Error]%v", out, err)
		}
		if strings.Contains(string(out), "BLOB_UNKNOWN") == false {
			t.Fatalf("Delete blob failed: [Info]%v", out)
		}
	}

	if setting.Authmode != "token" {
		url = fmt.Sprintf("%v://%v/v2/%v/%v/manifests/%v", setting.ListenMode, Domains, UserName, repoName, tag)
		out, err = exec.Command("sudo", "curl", "-X", "GET", url).Output()
		if err != nil {
			t.Fatalf("Delete tag failed: [Info]%v, [Error]%v", out, err)
		}
		if strings.Contains(string(out), "MANIFEST_UNKNOWN") == false {
			t.Fatalf("Delete manifest failed: [Info]%v", out)
		}
	}
}
func TestTaglistError(t *testing.T) {
	repoName := "busybox"
	url := fmt.Sprintf("%v://%v/v2/name/%v/tags/list", setting.ListenMode, Domains, repoName)
	out, err := exec.Command("sudo", "curl", "-X", "GET", url).Output()
	if err != nil {
		t.Fatalf("Get taglist failed: [Info]%v, [Error]%v", out, err)
	}
	if !strings.Contains(string(out), "TAG_INVALID") && !strings.Contains(string(out), "NAME_UNKNOWN") {
		t.Fatalf("Get taglist failed: [Info]%v", string(out))
	}
}
