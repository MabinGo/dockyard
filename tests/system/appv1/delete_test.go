package appv1

import (
	"fmt"
	"testing"

	"github.com/containerops/dockyard/tests/api"
)

func TestDeleteAPPWithoutTag(t *testing.T) {
	app := api.AppV1App{
		OS:   "linux",
		Arch: "arm",
		App:  fileName,
	}

	repo := api.AppV1Repo{
		URI:        api.DockyardURI,
		Namespace:  api.UserName,
		Repository: "delapp",
	}

	code, err := deleteAppTest(app, repo)
	if err != nil {
		if code == 0 {
			t.Log(err)
		} else {
			t.Fatal(err)
		}
	}
}

func TestDeleteAPPWithTag(t *testing.T) {
	app := api.AppV1App{
		OS:   "linux",
		Arch: "arm",
		App:  fileName,
		Tag:  "latest",
	}

	repo := api.AppV1Repo{
		URI:        api.DockyardURI,
		Namespace:  api.UserName,
		Repository: "delapp",
	}
	code, err := deleteAppTest(app, repo)
	if err != nil {
		if code == 0 {
			t.Log(err)
		} else {
			t.Fatal(err)
		}
	}
}

func TestDeleteMutileAPPWithoutTag(t *testing.T) {
	app := api.AppV1App{
		OS:   "linux",
		Arch: "arm",
		App:  fileName,
	}
	repositoryGroups := []string{"webapp", "testapp", "searchapp", "delapp"}
	for _, repository := range repositoryGroups {
		repo := api.AppV1Repo{
			URI:        api.DockyardURI,
			Namespace:  api.UserName,
			Repository: repository,
		}
		code, err := deleteAppTest(app, repo)
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

func TestDeleteMutileAPPWithTag(t *testing.T) {
	app := api.AppV1App{
		OS:   "linux",
		Arch: "arm",
		App:  fileName,
		Tag:  "latest",
	}
	repositoryGroups := []string{"webapp", "testapp", "searchapp", "delapp"}
	for _, repository := range repositoryGroups {
		repo := api.AppV1Repo{
			URI:        api.DockyardURI,
			Namespace:  api.UserName,
			Repository: repository,
		}
		code, err := deleteAppTest(app, repo)
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

func deleteAppTest(app api.AppV1App, repo api.AppV1Repo) (int, error) {
	appFile := testDir + "/" + fileName
	manifestFile := testDir + "/" + manifestName
	pullFileName := testDir + "/" + pullFile

	_, code, err := PushApp(repo, app, api.Token, appFile, manifestFile)
	if err != nil || code != 202 {
		if code == 0 {
			return code, fmt.Errorf("System (not dockyard) error :%s", err)
		}
		return code, fmt.Errorf(errStr, code, err, "push app")
	}

	code, err = repo.Delete(app, api.Token)
	if err != nil || code != 200 {
		if code == 0 {
			return code, fmt.Errorf("System (not dockyard) error :%s", err)
		}
		return code, fmt.Errorf(errStr, code, err, "delete app")
	}

	_, code, err = PullApp(repo, app, api.Token, pullFileName)
	if code == 0 {
		return code, fmt.Errorf("System (not dockyard) error :%s", err)
	}
	if code == 200 {
		return code, fmt.Errorf(errStr, code, err, "pull app")
	}
	return code, nil
}
