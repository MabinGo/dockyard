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

func AddUserToOrganizationTest(t *testing.T, oumJSON *OrganizationUserMapJSON, userName, password string) {
	b, err := json.Marshal(oumJSON)
	if err != nil {
		t.Error("marshal error.")
	}

	req, err := http.NewRequest("POST", "127.0.0.1:8080\\organizationusermap", strings.NewReader(string(b)))
	if err != nil {
		t.Error(err)
	}
	req.SetBasicAuth(userName, password)

	ctx := &macaron.Context{Req: macaron.Request{req}}
	if rt, b := AddUserToOrganization(ctx, &logs.BeeLogger{}); rt != http.StatusOK {
		t.Error(string(b))
	}
}

func Test_AddUserToOrganization(t *testing.T) {
	openDB(t)

	//1.create user for org admin
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

	//2.create organization
	org := &dao.Organization{
		Name: "huawei",
		//Email           string
		//Comment         string
		//URL             string
		//Location        string
		MemberPrivilege: dao.WRITE,
	}
	if err := org.Save(); err != nil {
		t.Error(err)
	}

	//3. set user as org admin
	oum := &dao.OrganizationUserMap{
		User: user,
		Role: dao.ORGADMIN,
		Org:  org,
	}
	if err := oum.Save(); err != nil {
		t.Error(err)
	}

	//4. create user for org member
	u1 := &dao.User{
		Name:     "wangqilin",
		Email:    "wangqilin@huawei.com",
		Password: "wangqilin",
		RealName: "wangqilin",
		Comment:  "commnet",
		Role:     dao.SYSMEMBER,
	}
	if err := u1.Save(); err != nil {
		t.Error(err)
	}

	//5. add user to org
	oumJSON := &OrganizationUserMapJSON{
		UserName: u1.Name,
		Role:     dao.ORGMEMBER,
		OrgName:  org.Name,
	}
	AddUserToOrganizationTest(t, oumJSON, user.Name, "liugenping")

	//6. delete add user, oum,org and user
	if err := u1.Delete(); err != nil {
		t.Error(err)
	}
	if err := user.Delete(); err != nil {
		t.Error(err)
	}
	if err := org.Delete(); err != nil {
		t.Error(err)
	}
}

func UpdateOrganizationUserMapTest(t *testing.T, oumJSON *OrganizationUserMapJSON, userName, password string) {
	b, err := json.Marshal(oumJSON)
	if err != nil {
		t.Error("marshal error.")
	}

	req, err := http.NewRequest("PUT", "127.0.0.1:8080\\updateorganizationusermap", strings.NewReader(string(b)))
	if err != nil {
		t.Error(err)
	}
	req.SetBasicAuth(userName, password)

	ctx := &macaron.Context{Req: macaron.Request{req}}
	if rt, b := UpdateOrganizationUserMap(ctx, &logs.BeeLogger{}); rt != http.StatusOK {
		t.Error(string(b))
	}
}

func Test_UpdateOrganizationUserMap(t *testing.T) {
	openDB(t)

	//1.create user for org admin
	user := &dao.User{
		Name:     "liugenpingl",
		Email:    "liugenping@huawei.com",
		Password: "liugenping",
		RealName: "liugenping",
		Comment:  "commnet",
		Role:     dao.SYSMEMBER,
	}
	if err := user.Save(); err != nil {
		t.Error(err)
	}

	//2.create organization
	org := &dao.Organization{
		Name: "huawei.com",
		//Email           string
		//Comment         string
		//URL             string
		//Location        string
		MemberPrivilege: dao.WRITE,
	}
	if err := org.Save(); err != nil {
		t.Error(err)
	}

	//3. set user as org admin
	oum := &dao.OrganizationUserMap{
		User: user,
		Role: dao.ORGADMIN,
		Org:  org,
	}
	if err := oum.Save(); err != nil {
		t.Error(err)
	}

	//4. create user for org member
	u1 := &dao.User{
		Name:     "wangqilinw",
		Email:    "wangqilin@huawei.com",
		Password: "wangqilin",
		RealName: "wangqilin",
		Comment:  "commnet",
		Role:     dao.SYSMEMBER,
	}
	if err := u1.Save(); err != nil {
		t.Error(err)
	}

	//5. add user to org
	oumJSON := &OrganizationUserMapJSON{
		UserName: u1.Name,
		Role:     dao.ORGMEMBER,
		OrgName:  org.Name,
	}
	AddUserToOrganizationTest(t, oumJSON, user.Name, "liugenping")

	//6. update
	oumJSON.Role = dao.ORGADMIN
	UpdateOrganizationUserMapTest(t, oumJSON, user.Name, "liugenping")

	//7. query update result
	oum1 := &dao.OrganizationUserMap{
		User: u1,
		Org:  org,
	}
	if exist, err := oum1.Get(); err != nil {
		t.Error(err)
	} else if !exist {
		t.Error("OrganizationUserMap is not exitst")
	} else {
		if oum1.Role != dao.ORGADMIN {
			t.Error("update OrganizationUserMap info failed")
		}
	}

	//8. delete add user, oum,org and user
	if err := u1.Delete(); err != nil {
		t.Error(err)
	}
	if err := user.Delete(); err != nil {
		t.Error(err)
	}
	if err := org.Delete(); err != nil {
		t.Error(err)
	}
}
