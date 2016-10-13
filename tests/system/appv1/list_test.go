package appv1

import (
	"encoding/json"
	"testing"

	"github.com/containerops/dockyard/tests/api"
)

func TestListInit(t *testing.T) {
	app := api.AppV1App{
		OS:   "linux",
		Arch: "arm",
		App:  fileName,
		Tag:  "latest",
	}

	repo := api.AppV1Repo{
		URI:        api.DockyardURI,
		Namespace:  api.UserName,
		Repository: "listapp",
	}

	appFile := testDir + "/" + fileName
	manifestFile := testDir + "/" + manifestName

	_, code, err := PushApp(repo, app, api.Token, appFile, manifestFile)
	if err != nil || code != 202 {
		if code == 0 {
			t.Logf("System (not dockyard) error :%s", err)
			return
		}
		t.Fatalf(errStr, code, err, "push app")
		return
	}
}

func TestListScope(t *testing.T) {
	repo := api.AppV1Repo{
		URI:        api.DockyardURI,
		Namespace:  api.UserName,
		Repository: "listapp",
	}

	byteArr, code, err := repo.GetList(api.Token)
	if err != nil || code != 200 {
		if code == 0 {
			t.Logf("System (not dockyard) error :%s", err)
			return
		}
		t.Fatalf(errStr, code, err, "get list scope")
		return
	}
	results := []Body{}
	if err := json.Unmarshal(byteArr, &results); err != nil {
		t.Fatalf(errStr, code, err, "get list scope format json")
		return
	}
	for _, v := range results {
		if v.Namespace != repo.Namespace || v.Repository != repo.Repository {
			t.Fatalf(errStr, code, err, "get list scope result is wrong")
			return
		}
	}
}

func TestListScopeNotExistRepository(t *testing.T) {
	repo := api.AppV1Repo{
		URI:        api.DockyardURI,
		Namespace:  api.UserName,
		Repository: "noexistrfvrfvrfv",
	}

	_, code, err := repo.GetList(api.Token)
	if err == nil && code == 200 {
		t.Fatalf(errStr, code, err, "get list scope empty repository")
		return
	}
}

func TestListScopeEmptyRepository(t *testing.T) {
	repo := api.AppV1Repo{
		URI:        api.DockyardURI,
		Namespace:  api.UserName,
		Repository: " ",
	}

	_, code, err := repo.GetList(api.Token)
	if err == nil && code == 200 {
		t.Fatalf(errStr, code, err, "get list scope empty repository")
		return
	}
}

func TestListScopeEmptyNamespace(t *testing.T) {
	repo := api.AppV1Repo{
		URI:        api.DockyardURI,
		Namespace:  " ",
		Repository: "listapp",
	}

	_, code, err := repo.GetList(api.Token)
	if err == nil && code == 200 {
		t.Fatalf(errStr, code, err, "get list scope empty repository")
		return
	}
}

func TestListScopeIllegalRepository(t *testing.T) {
	repo := api.AppV1Repo{
		URI:        api.DockyardURI,
		Namespace:  api.UserName,
		Repository: "abc&^%$",
	}

	_, code, err := repo.GetList(api.Token)
	if err == nil && code == 200 {
		t.Fatalf(errStr, code, err, "get list scope empty repository")
		return
	}
}

func TestListScopeIllegalNamespace(t *testing.T) {
	repo := api.AppV1Repo{
		URI:        api.DockyardURI,
		Namespace:  "abc*^%$&",
		Repository: "listapp",
	}

	_, code, err := repo.GetList(api.Token)
	if err == nil && code == 200 {
		t.Fatalf(errStr, code, err, "get list scope empty repository")
		return
	}
}

func TestListScopeRepositoryLenMoreThan255(t *testing.T) {
	repo := api.AppV1Repo{
		URI: api.DockyardURI,
		Namespace: "abcdefghij1234567890" +
			"abcdefghij1234567890" + "abcdefghij1234567890" + "abcdefghij1234567890" +
			"abcdefghij1234567890" + "abcdefghij1234567890" + "abcdefghij1234567890" +
			"abcdefghij1234567890" + "abcdefghij1234567890" + "abcdefghij1234567890" +
			"abcdefghij1234567890" + "abcdefghij1234567890" + "abcdefghij123456",
		Repository: "listapp",
	}

	_, code, err := repo.GetList(api.Token)
	if err == nil && code == 200 {
		t.Fatalf(errStr, code, err, "get list scope repository len more than 255")
		return
	}
}

func TestListScopeChineseRepository(t *testing.T) {
	repo := api.AppV1Repo{
		URI:        api.DockyardURI,
		Namespace:  api.UserName,
		Repository: "中文输入",
	}

	_, code, err := repo.GetList(api.Token)
	if err == nil && code == 200 {
		t.Fatalf(errStr, code, err, "get list scope chinese repository")
		return
	}
}

func TestListScopeChineseNamespace(t *testing.T) {
	repo := api.AppV1Repo{
		URI:        api.DockyardURI,
		Namespace:  "中文输入",
		Repository: "listapp",
	}

	_, code, err := repo.GetList(api.Token)
	if err == nil && code == 200 {
		t.Fatalf(errStr, code, err, "get list scope chinese namespace")
		return
	}
}
