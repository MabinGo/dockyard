package docker

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"

	"github.com/containerops/dockyard/utils/setting"
)

func TestCoverInit(t *testing.T) {
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

func TestCoverInfo(t *testing.T) {
	repoName := "busybox"
	tag := "latest"
	repoBase := repoName + ":" + tag
	repoDest := Domains + "/" + UserName + "/" + repoBase
	path := "./dockerfile1"

	url := fmt.Sprintf("%v://%v/v2/%v/%v/manifests/%v", setting.ListenMode, Domains, UserName, repoName, tag)
	var outpre []byte
	var outcur []byte
	var erro error

	outpre, erro = exec.Command("sudo", "curl", "-X", "GET", url).Output()
	if erro != nil {
		t.Fatalf("Get manifest failed: [Info]%v, [Error]%v", outpre, erro)
	}

	cmd := exec.Command(DockerBinary, "rmi", repoDest)
	if out, err := ParseCmdCtx(cmd); err != nil {
		t.Fatalf("Remove testing preparation is failed: [Info]%v, [Error]%v", out, err)
	}
	cmd = exec.Command(DockerBinary, "build", "-t", repoDest, path)
	if out, err := ParseCmdCtx(cmd); err != nil {
		t.Fatalf("Build %v failed: [Info]%v, [Error]%v", repoDest, out, err)
	}

	cmd = exec.Command(DockerBinary, "push", repoDest)
	if out, err := ParseCmdCtx(cmd); err != nil {
		t.Fatalf("Push testing preparation is failed: [Info]%v, [Error]%v", out, err)
	}
	url = fmt.Sprintf("%v://%v/v2/%v/%v/manifests/%v", setting.ListenMode, Domains, UserName, repoName, tag)

	outcur, erro = exec.Command("sudo", "curl", "-X", "GET", url).Output()
	if erro != nil {
		t.Fatalf("Get manifest failed: [Info]%v, [Error]%v", outpre, erro)
	}

	if strings.Compare(string(outpre), string(outcur)) == 0 {
		t.Fatalf("Image info cover failed: [Info]%v", outcur)
	}

}
