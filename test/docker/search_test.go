package docker

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"

	"github.com/containerops/dockyard/utils/setting"
)

func TestPush(t *testing.T) {
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
		t.Fatalf("Pull testing preparation is failed: [Info]%v, [Error]%v", out, err)
	}

}

func TestGetRepository(t *testing.T) {
	repoName := "busybox"
	repoDest := UserName + "/" + repoName
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
