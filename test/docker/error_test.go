package docker

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"testing"

	"github.com/containerops/dockyard/utils/setting"
)

var digest string

func TestErrorInit(t *testing.T) {
	repoBase := "busybox:latest"
	repoDest := Domains + "/" + UserName + "/" + repoBase

	if setting.Authmode == "token" {
		cmd := exec.Command(DockerBinary, "login", "-u", user.Name, "-p", user.Password, "-e", user.Email, Domains)
		if err := cmd.Run(); err != nil {
			t.Fatalf("Docker login faild: [Error]%v", err)
		}
	}
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
	} else {
		sha := strings.Split(out, "sha256:")[1]
		digest = "sha256:" + strings.Split(sha, " ")[0]
	}

}

func TestDeleteError(t *testing.T) {
	repoName := "busybox"
	tag := "latest"

	var encodstr string
	var basecode string
	var authorization string
	if setting.Authmode == "token" {
		encodstr = user.Name + ":" + user.Password
		basecode = base64.StdEncoding.EncodeToString([]byte(encodstr))
		authorization = "Authorization: Basic " + basecode
	}
	url := fmt.Sprintf("%v://%v/v2/%v/%v/manifests/%v", setting.ListenMode, Domains, UserName, repoName, tag)
	var output []byte
	var out []byte
	var err error
	if setting.Authmode == "token" {
		output, err = exec.Command("sudo", "curl", "-H", authorization, "-X", "GET", url).Output()
	} else {
		output, err = exec.Command("sudo", "curl", "-X", "GET", url).Output()
	}
	if err != nil {
		t.Fatalf("Get manifest failed: [Info]%v, [Error]%v", output, err)
	}
	url = fmt.Sprintf("%v://%v/v2/%v/%v/manifests/%v", setting.ListenMode, Domains, UserName, repoName, "no")
	if setting.Authmode == "token" {
		out, err = exec.Command("sudo", "curl", "-H", authorization, "-X", "DELETE", url).Output()
	} else {
		out, err = exec.Command("sudo", "curl", "-X", "DELETE", url).Output()
	}
	if err != nil {
		t.Fatalf("Delete tag failed: [Info]%v, [Error]%v", out, err)
	}
	if strings.Contains(string(out), "DIGEST_INVALID") == false {
		t.Fatalf("Delete manifest failed: [Info]%v", string(out))
	}
	url = fmt.Sprintf("%v://%v/v2/%v/%v/manifests/%v", setting.ListenMode, Domains, UserName, repoName, digest)
	if setting.Authmode == "token" {
		out, err = exec.Command("sudo", "curl", "-H", authorization, "-X", "DELETE", url).Output()
	} else {
		out, err = exec.Command("sudo", "curl", "-X", "DELETE", url).Output()
	}
	if err != nil {
		t.Fatalf("Delete tag failed: [Info]%v, [Error]%v", out, err)
	}
	if strings.Contains(string(out), "errors") == true {
		t.Fatalf("Delete manifest failed: [Info]%v", string(out))
	}

	url = fmt.Sprintf("%v://%v/v2/%v/%v/manifests/%v", setting.ListenMode, Domains, UserName, repoName, digest)
	if setting.Authmode != "token" {
		out, err = exec.Command("sudo", "curl", "-X", "DELETE", url).Output()
		if err != nil {
			t.Fatalf("Delete tag failed: [Info]%v, [Error]%v", out, err)
		}
		if strings.Contains(string(out), "MANIFEST_UNKNOWN") == false {
			t.Fatalf("Delete manifest failed: [Info]%v", string(out))
		}
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
		if setting.Authmode == "token" {
			out, err = exec.Command("sudo", "curl", "-H", authorization, "-X", "DELETE", url).Output()
		} else {
			out, err = exec.Command("sudo", "curl", "-X", "DELETE", url).Output()
		}

		if err != nil {
			t.Fatalf("Delete tag failed: [Info]%v, [Error]%v", out, err)
		}
		if strings.Contains(string(out), "errors") == true {
			t.Fatalf("Delete blob failed: [Info]%v", out)
		}

		url = fmt.Sprintf("%v://%v/v2/%v/%v/blobs/test", setting.ListenMode, Domains, UserName, repoName)
		if setting.Authmode == "token" {
			out, err = exec.Command("sudo", "curl", "-H", authorization, "-X", "DELETE", url).Output()
		} else {
			out, err = exec.Command("sudo", "curl", "-X", "DELETE", url).Output()
		}
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
		if setting.Authmode == "token" {
			out, err = exec.Command("sudo", "curl", "-H", authorization, "-X", "DELETE", url).Output()
		} else {
			out, err = exec.Command("sudo", "curl", "-X", "DELETE", url).Output()
		}
		if err != nil {
			t.Fatalf("Delete tag failed: [Info]%v, [Error]%v", out, err)
		}
		if strings.Contains(string(out), sha) == true {
			t.Fatalf("Delete blob failed: [Info]%v", out)
		}

		if setting.Authmode == "token" {
			out, err = exec.Command("sudo", "curl", "-H", authorization, "-X", "DELETE", url).Output()
		} else {
			out, err = exec.Command("sudo", "curl", "-X", "DELETE", url).Output()
		}
		if err != nil {
			t.Fatalf("Delete tag failed: [Info]%v, [Error]%v", out, err)
		}
		if strings.Contains(string(out), "BLOB_UNKNOWN") == false {
			t.Fatalf("Delete blob failed: [Info]%v", out)
		}
	}

	if setting.Authmode != "token" {
		url = fmt.Sprintf("%v://%v/v2/%v/%v/manifests/%v", setting.ListenMode, Domains, UserName, repoName, tag)
		out, err = exec.Command("sudo", "curl", "-H", authorization, "-X", "GET", url).Output()
		if err != nil {
			t.Fatalf("Delete tag failed: [Info]%v, [Error]%v", out, err)
		}
		if strings.Contains(string(out), "MANIFEST_UNKNOWN") == false {
			t.Fatalf("Delete manifest failed: [Info]%v", out)
		}
	}
}
func TestTaglistError(t *testing.T) {
	if setting.Authmode != "token" {
		repoName := "busybox"
		url := fmt.Sprintf("%v://%v/v2/name/%v/tags/list", setting.ListenMode, Domains, repoName)
		out, err := exec.Command("sudo", "curl", "-X", "GET", url).Output()
		if err != nil {
			t.Fatalf("Get taglist failed: [Info]%v, [Error]%v", out, err)
		}
		if !strings.Contains(string(out), "TAG_INVALID") && !strings.Contains(string(out), "NAME_UNKNOWN") {
			t.Fatalf("Get taglist failed: [Info]%v", string(out))
		}
	} else {
		if err := exec.Command(DockerBinary, "logout", Domains).Run(); err != nil {
			t.Fatalf("Docker logout failed:[Error]%v", err)
		}
	}
}
