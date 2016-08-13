package docker

import (
	"encoding/base64"
	"fmt"
	"os/exec"
	"strings"
	"testing"

	"github.com/containerops/dockyard/utils/setting"
)

func TestPush(t *testing.T) {
	repoBase := "busybox:latest"
	repoDest := Domains + "/" + UserName + "/" + repoBase

	if setting.Authmode == "token" {
		cmd := exec.Command(DockerBinary, "login", "-u", user.Name, "-p", user.Password, "-e", user.Email, Domains)
		if err := cmd.Run(); err != nil {
			t.Fatalf("Docker login faild: [Error]%v", err)
		}
	}
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
		t.Fatalf("Pull testing preparation is failed: [Info]%v, [Error]%v", out, err)
	}

}

func TestGetRepository(t *testing.T) {
	repoName := "busybox"
	repoDest := UserName + "/" + repoName
	if setting.Authmode == "token" {
		encodstr := user.Name + ":" + user.Password
		basecode := base64.StdEncoding.EncodeToString([]byte(encodstr))
		authorization := "Authorization: Basic " + basecode
		url := fmt.Sprintf("%v://%v/v2/_catalog", setting.ListenMode, Domains)
		out, err := exec.Command("sudo", "curl", "-H", authorization, "-X", "GET", url).Output()
		if err != nil {
			t.Fatalf("Get repository failed: [Info]%v, [Error]%v", out, err)
		}
		if strings.Contains(string(out), repoDest) != true {
			t.Fatalf("Get repository failed: [Info]%v, [Error]%v", string(out), err)
		}
		if err := exec.Command(DockerBinary, "logout", Domains).Run(); err != nil {
			t.Fatalf("Docker logout failed:[Error]%v", err)
		}
	} else {
		url := fmt.Sprintf("%v://%v/v2/_catalog", setting.ListenMode, Domains)
		out, err := exec.Command("sudo", "curl", "-X", "GET", url).Output()
		if err != nil {
			t.Fatalf("Get repository failed: [Info]%v, [Error]%v", out, err)
		}
		if strings.Contains(string(out), repoDest) != true {
			t.Fatalf("Get repository failed: [Info]%v, [Error]%v", string(out), err)
		}
		if err := exec.Command(DockerBinary, "logout", Domains).Run(); err != nil {
			t.Fatalf("Docker logout failed:[Error]%v", err)
		}
	}
}
