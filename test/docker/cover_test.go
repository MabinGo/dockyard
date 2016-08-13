package docker

import (
	"encoding/base64"
	"fmt"
	"os/exec"
	"strings"
	"testing"

	"github.com/containerops/dockyard/utils/setting"
)

func TestCoverInit(t *testing.T) {
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
		t.Fatalf("Push testing preparation is failed: [Info]%v, [Error]%v", out, err)
	}
}

func TestCoverInfo(t *testing.T) {
	repoName := "busybox"
	tag := "latest"
	repoBase := repoName + ":" + tag
	repoDest := Domains + "/" + UserName + "/" + repoBase
	path := "./dockerfile1"

	var encodstr string
	var basecode string
	var authorization string
	if setting.Authmode == "token" {
		encodstr = user.Name + ":" + user.Password
		basecode = base64.StdEncoding.EncodeToString([]byte(encodstr))
		authorization = "Authorization: Basic " + basecode
	}
	url := fmt.Sprintf("%v://%v/v2/%v/%v/manifests/%v", setting.ListenMode, Domains, UserName, repoName, tag)
	var outpre []byte
	var outcur []byte
	var erro error
	if setting.Authmode == "token" {
		outpre, erro = exec.Command("sudo", "curl", "-H", authorization, "-X", "GET", url).Output()
		if erro != nil {
			t.Fatalf("Get manifest failed: [Info]%v, [Error]%v", outpre, erro)
		}
	} else {
		outpre, erro = exec.Command("sudo", "curl", "-X", "GET", url).Output()
		if erro != nil {
			t.Fatalf("Get manifest failed: [Info]%v, [Error]%v", outpre, erro)
		}
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
	if setting.Authmode == "token" {
		outcur, erro = exec.Command("sudo", "curl", "-H", authorization, "-X", "GET", url).Output()
		if erro != nil {
			t.Fatalf("Get manifest failed: [Info]%v, [Error]%v", outcur, erro)
		}
	} else {
		outcur, erro = exec.Command("sudo", "curl", "-X", "GET", url).Output()
		if erro != nil {
			t.Fatalf("Get manifest failed: [Info]%v, [Error]%v", outpre, erro)
		}
	}
	if strings.Compare(string(outpre), string(outcur)) == 0 {
		t.Fatalf("Image info cover failed: [Info]%v", outcur)
	}
	if setting.Authmode == "token" {
		if err := exec.Command(DockerBinary, "logout", Domains).Run(); err != nil {
			t.Fatalf("Docker logout failed:[Error]%v", err)
		}
	}
}
