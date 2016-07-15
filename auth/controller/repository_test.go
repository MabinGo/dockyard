package controller

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/astaxie/beego/logs"
	"gopkg.in/macaron.v1"

	"github.com/containerops/dockyard/auth/dao"
	"github.com/containerops/dockyard/utils/setting"
)

func init() {
	setting.Authn = "authn_db"
}

func CreateRepositoryTest(t *testing.T, repoJSON *RepositoryJSON, userName, password string) {
	b, err := json.Marshal(repoJSON)
	if err != nil {
		t.Error("marshal error.")
	}

	req, err := http.NewRequest("POST", "127.0.0.1:8080\\repository", strings.NewReader(string(b)))
	if err != nil {
		t.Error(err)
	}
	req.SetBasicAuth(userName, password)
	ctx := &macaron.Context{Req: macaron.Request{req}}
	if rt, b := CreateRepository(ctx, &logs.BeeLogger{}); rt != http.StatusOK {
		t.Error(string(b))
	}
}

func Test_CreateRepository(t *testing.T) {
	openDB(t)

	//1. create user
	user := &dao.User{
		Name:     "liugenping",
		Email:    "liugenping@huawei.com",
		Password: "liugenping",
		RealName: "liugenping",
		Comment:  "commnet",
		Role:     dao.SYSMEMBER,
	}
	if err := user.Save(); err != nil {
		t.Error(err)
	}

	//2. create repo
	repoJSON := &RepositoryJSON{
		Name:     "ubuntu",
		IsPublic: true,
		Comment:  "comment",
		//OrgName:  "",
		UserName: "liugenping",
	}
	CreateRepositoryTest(t, repoJSON, user.Name, "liugenping")

	//3. query repo
	repo := &dao.RepositoryEx{
		Name:     repoJSON.Name,
		IsOrgRep: false,
		User:     &dao.User{Name: repoJSON.UserName},
	}

	if b, err := repo.Get(); err != nil {
		t.Error(err)
	} else if !b {
		t.Error("repo is not exist")
	} else {
		if repo.IsPublic != repoJSON.IsPublic ||
			repo.Comment != repoJSON.Comment {
			t.Error("repos's save is not same with get")
		}
	}

	//4. delete repo and user
	if err := user.Delete(); err != nil {
		t.Error(err)
	}
}

func UpdateRepositoryTest(t *testing.T, repoJSON *RepositoryJSON, userName, password string) {
	b, err := json.Marshal(repoJSON)
	if err != nil {
		t.Error("marshal error.")
	}

	req, err := http.NewRequest("PUT", "127.0.0.1:8080\\repository\\update", strings.NewReader(string(b)))
	if err != nil {
		t.Error(err)
	}
	req.SetBasicAuth(userName, password)
	ctx := &macaron.Context{Req: macaron.Request{req}}
	if rt, b := UpdateRepository(ctx, &logs.BeeLogger{}); rt != http.StatusOK {
		t.Error(string(b))
	}
}

func Test_UpdateRepository(t *testing.T) {

	//1. create user
	user := &dao.User{
		Name:     "test",
		Email:    "test@huawei.com",
		Password: "test",
		RealName: "test",
		Comment:  "commnet",
		Role:     dao.SYSMEMBER,
	}
	if err := user.Save(); err != nil {
		t.Error(err)
	}

	//2. create repo
	repoJSON := &RepositoryJSON{
		Name:     "ubuntu",
		IsPublic: true,
		Comment:  "comment",
		//OrgName:  "",
		UserName: user.Name,
	}
	CreateRepositoryTest(t, repoJSON, user.Name, "test")

	//3. update repo
	repoJSON.IsPublic = false
	repoJSON.Comment = "update_test"
	UpdateRepositoryTest(t, repoJSON, user.Name, "test")

	//4. query repo
	repo := &dao.RepositoryEx{
		Name:     repoJSON.Name,
		IsOrgRep: false,
		User:     &dao.User{Name: repoJSON.UserName},
	}

	if b, err := repo.Get(); err != nil {
		t.Error(err)
	} else if !b {
		t.Error("repo is not exist")
	} else {
		if repo.IsPublic != repoJSON.IsPublic ||
			repo.Comment != repoJSON.Comment {
			t.Error("update repository info failed")
		}
	}

	//4. delete repo and user
	if err := user.Delete(); err != nil {
		t.Error(err)
	}
}
