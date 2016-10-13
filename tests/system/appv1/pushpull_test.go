package appv1

import (
	"fmt"
	"testing"

	"github.com/containerops/dockyard/tests/api"
)

func TestPushPullSingleAPPWithoutTag(t *testing.T) {
	app := api.AppV1App{
		OS:   "linux",
		Arch: "arm",
		App:  fileName,
	}

	repo := api.AppV1Repo{
		URI:        api.DockyardURI,
		Namespace:  api.UserName,
		Repository: "pullpushapp",
	}

	code, err := pushPullAppTest(app, repo)
	if err != nil {
		if code == 0 {
			t.Log(err)
		} else {
			t.Fatal(err)
		}
	}
}

func TestPushPullSingleAPPWithTag(t *testing.T) {
	app := api.AppV1App{
		OS:   "linux",
		Arch: "arm",
		App:  fileName,
		Tag:  "latest",
	}

	repo := api.AppV1Repo{
		URI:        api.DockyardURI,
		Namespace:  api.UserName,
		Repository: "pullpushapp",
	}
	code, err := pushPullAppTest(app, repo)
	if err != nil {
		if code == 0 {
			t.Log(err)
		} else {
			t.Fatal(err)
		}
	}
}

func TestPushPullMutileAPPWithoutTag(t *testing.T) {
	app := api.AppV1App{
		OS:   "linux",
		Arch: "arm",
		App:  fileName,
	}
	repositoryGroups := []string{"webapp", "testapp", "searchapp", "pullpushapp"}
	for _, repository := range repositoryGroups {
		repo := api.AppV1Repo{
			URI:        api.DockyardURI,
			Namespace:  api.UserName,
			Repository: repository,
		}
		code, err := pushPullAppTest(app, repo)
		if err != nil {
			if code == 0 {
				t.Log(err)
			} else {
				t.Fatal(err)
			}
			break
		}
	}
}

func TestPushPullMutileAPPWithTag(t *testing.T) {
	app := api.AppV1App{
		OS:   "linux",
		Arch: "arm",
		App:  fileName,
		Tag:  "latest",
	}
	repositoryGroups := []string{"webapp", "testapp", "searchapp", "pullpushapp"}
	for _, repository := range repositoryGroups {
		repo := api.AppV1Repo{
			URI:        api.DockyardURI,
			Namespace:  api.UserName,
			Repository: repository,
		}
		code, err := pushPullAppTest(app, repo)
		if err != nil {
			if code == 0 {
				t.Log(err)
			} else {
				t.Fatal(err)
			}
			break
		}
	}
}

func TestPushPullAPPWithDifferentTag(t *testing.T) {
	repo := api.AppV1Repo{
		URI:        api.DockyardURI,
		Namespace:  api.UserName,
		Repository: "pullpushapp",
	}

	tagGroups := []string{"latest", "1.0", "2.0", "3.0"}
	for _, tagName := range tagGroups {
		app := api.AppV1App{
			OS:   "linux",
			Arch: "arm",
			App:  fileName,
			Tag:  tagName,
		}
		code, err := pushPullAppTest(app, repo)
		if err != nil {
			if code == 0 {
				t.Log(err)
			} else {
				t.Fatal(err)
			}
			break
		}
	}
}

func TestPushExistedAPP(t *testing.T) {
	app := api.AppV1App{
		OS:   "linux",
		Arch: "arm",
		App:  fileName,
		Tag:  "latest",
	}

	repo := api.AppV1Repo{
		URI:        api.DockyardURI,
		Namespace:  api.UserName,
		Repository: "pullpushapp",
	}
	code, err := pushPullAppTest(app, repo)
	if err != nil {
		if code == 0 {
			t.Log(err)
		} else {
			t.Fatal(err)
		}
		return
	}
	code, err = pushPullAppTest(app, repo)
	if err != nil {
		if code == 0 {
			t.Log(err)
		} else {
			t.Fatal(err)
		}
	}
}

/*
func TestPullAPPMeta(t *testing.T) {
	repo := api.AppV1Repo{
		URI:        api.DockyardURI,
		Namespace:  api.UserName,
		Repository: "pullpushapp",
	}

	_, code, err := repo.GetMeta(api.Token)
	if code != 200 || err != nil {
		if code == 0 {
			t.Logf("System (not dockyard) error :%s", err)
		} else {
			t.Fatalf(errStr, code, err, "get meta")
		}
		return
	}

	_, code, err = repo.GetMetaSign(api.Token)
	if (code != 200 && code != 500) || err != nil {
		if code == 0 {
			t.Log(fmt.Sprintf("System (not dockyard) error :%s", err))
		} else {
			t.Fatalf(errStr, code, err, "get meta sign")
		}
	}
}

func TestPullAPPMetaWithNamespaceEqualNull(t *testing.T) {
	repo := api.AppV1Repo{
		URI:        api.DockyardURI,
		Namespace:  "",
		Repository: "pullpushapp",
	}

	_, code, err := repo.GetMeta(api.Token)
	if err == nil && code == 200 {
		t.Fatal(errStr, code, err, "get meta")
		return
	}

	_, code, err = repo.GetMetaSign(api.Token)
	if err == nil && code == 200 {
		t.Fatal(errStr, code, err, "get meta sign")
	}
}

func TestPullAPPMetaWithRepositoryEqualNull(t *testing.T) {
	repo := api.AppV1Repo{
		URI:        api.DockyardURI,
		Namespace:  api.UserName,
		Repository: "",
	}

	_, code, err := repo.GetMeta(api.Token)
	if err == nil && code == 200 {
		t.Fatal(errStr, code, err, "get meta")
		return
	}

	_, code, err = repo.GetMetaSign(api.Token)
	if err == nil && code == 200 {
		t.Fatal(errStr, code, err, "get meta sign")
	}
}

func TestPullAPPMetaWithNamespaceEmpty(t *testing.T) {
	repo := api.AppV1Repo{
		URI:        api.DockyardURI,
		Namespace:  " ",
		Repository: "pullpushapp",
	}

	_, code, err := repo.GetMeta(api.Token)
	if err == nil && code == 200 {
		t.Fatal(errStr, code, err, "get meta")
		return
	}

	_, code, err = repo.GetMetaSign(api.Token)
	if err == nil && code == 200 {
		t.Fatal(errStr, code, err, "get meta sign")
	}
}

func TestPullAPPMetaWithRepositoryEqualEmpty(t *testing.T) {
	repo := api.AppV1Repo{
		URI:        api.DockyardURI,
		Namespace:  api.UserName,
		Repository: " ",
	}

	_, code, err := repo.GetMeta(api.Token)
	if err == nil && code == 200 {
		t.Fatal(errStr, code, err, "get meta")
		return
	}

	_, code, err = repo.GetMetaSign(api.Token)
	if err == nil && code == 200 {
		t.Fatal(errStr, code, err, "get meta sign")
	}
}
*/
func TestPushPull500MTarWithTag(t *testing.T) {
	app := api.AppV1App{
		OS:   "linux",
		Arch: "arm",
		App:  fileName,
		Tag:  "v500",
	}

	repo := api.AppV1Repo{
		URI:        api.DockyardURI,
		Namespace:  api.UserName,
		Repository: "pullpushapp",
	}

	appFile := testDir + "/" + maxSizeFile
	manifestFile := testDir + "/" + maxSizeManifest
	pullFileName := testDir + "/" + pullFile

	_, code, err := PushApp(repo, app, api.Token, appFile, manifestFile)
	if err != nil || code != 202 {
		if code == 0 {
			t.Logf("System (not dockyard) error :%s", err)
			return
		}
		t.Fatalf(errStr, code, err, "push app")
		return
	}

	pushSha512, code, err := PushApp(repo, app, api.Token, appFile, manifestFile)
	if err != nil || code != 202 {
		if code == 0 {
			t.Logf("System (not dockyard) error :%s", err)
			return
		}
		t.Fatalf(errStr, code, err, "push app")
		return
	}

	pullSha512, code, err := PullApp(repo, app, api.Token, pullFileName)
	if err != nil || code != 200 {
		if code == 0 {
			t.Logf("System (not dockyard) error :%s", err)
			return
		}
		t.Fatalf(errStr, code, err, "pull app")
		return
	}
	if diffPullPushSha512(pushSha512, pullSha512) {
		t.Fatalf("pull and push sha512 was inconsistent push:%v, pull: %v", pushSha512, pullSha512)
		return
	}
}

func pushPullAppTest(app api.AppV1App, repo api.AppV1Repo) (int, error) {
	appFile := testDir + "/" + fileName
	manifestFile := testDir + "/" + manifestName
	pullFileName := testDir + "/" + pullFile

	_, code, err := PushApp(repo, app, api.Token, appFile, manifestFile)
	if err != nil || code != 202 {
		fmt.Println("pushSha512 first")
		if code == 0 {
			return code, fmt.Errorf("System (not dockyard) error :%s", err)
		}
		return code, fmt.Errorf(errStr, code, err, "push app")
	}

	pushSha512, code, err := PushApp(repo, app, api.Token, appFile, manifestFile)
	if err != nil || code != 202 {
		fmt.Println("pushSha512")
		if code == 0 {
			return code, fmt.Errorf("System (not dockyard) error :%s", err)
		}
		return code, fmt.Errorf(errStr, code, err, "push app")
	}

	pullSha512, code, err := PullApp(repo, app, api.Token, pullFileName)
	if err != nil || code != 200 {
		fmt.Println("pullSha512")
		if code == 0 {
			return code, fmt.Errorf("System (not dockyard) error :%s", err)
		}
		return code, fmt.Errorf(errStr, code, err, "pull app")
	}
	if diffPullPushSha512(pushSha512, pullSha512) {
		return code, fmt.Errorf("pull and push sha512 was inconsistent push:%v, pull: %v", pushSha512, pullSha512)
	}
	return code, nil
}
