package authz

import (
	"crypto"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"testing"

	"github.com/docker/distribution/registry/auth/token"
	"github.com/docker/libtrust"

	"github.com/containerops/dockyard/auth/authn"
	"github.com/containerops/dockyard/auth/dao"
	"github.com/containerops/dockyard/utils/db"
	"github.com/containerops/dockyard/utils/setting"
)

var (
	service        string
	rootCertBundle string

	publicKeys map[string]libtrust.PublicKey
	rootCerts  *x509.CertPool
)

func openDB() error {
	if err := db.RegisterDriver("mysql"); err != nil {
		return err
	} else {
		db.Drv.RegisterModel(new(dao.Organization), new(dao.User),
			new(dao.OrganizationUserMap), new(dao.RepositoryEx),
			new(dao.Team), new(dao.TeamRepositoryMap), new(dao.TeamUserMap))
		if err := db.Drv.InitDB("mysql", "root", "root", "127.0.0.1:3306", "dockyard", 0); err != nil {
			return err
		}
	}
	return nil
}
func Test_SetupEnv(t *testing.T) {
	//1. set val
	setting.Expiration = 50
	setting.Issuer = "registry-token-issuer"
	setting.Authn = "authn_db"
	setting.PrivateKey = "../../cert/containerops/containerops.key"
	rootCertBundle = "../../cert/containerops/containerops.crt"
	service = "dockyard.com"

	//2. parse certs
	var err error
	if publicKeys, rootCerts, err = ParseCertAndPublicKey(rootCertBundle); err != nil {
		t.Fatal(err)
	}

	//3. open db
	if err := openDB(); err != nil {
		t.Fatal(err)
	}

	//4. create fk and root user
	if err := dao.InitDAO(); err != nil {
		t.Fatal(err)
	}

	//5. start authn
	authn.Authn, err = authn.NewAuthenticator()
	if err != nil {
		t.Fatal(fmt.Errorf("New Authenticator error: %s\n", err.Error()))
	}
	//6. start authz
	setting.NameSpace = NameSpace_All
	Authz, err = NewAuthorizer()
	if err != nil {
		t.Fatal(fmt.Errorf("New Authorizer error: %s\n", err.Error()))
	}
}

//test case 2
func Test_UserNamespaceDelete1(t *testing.T) {

	user := &dao.User{
		Name:     "liugenping",
		Email:    "liugenping@huawei.com",
		Password: "liugenping",
		Status:   0,
		Role:     dao.SYSMEMBER,
	}
	if err := user.Save(); err != nil {
		t.Fatal(err)
	}
	user.Password = "liugenping"
	defer func() {
		if err := user.Delete(); err != nil {
			t.Fatal(err)
		}
	}()

	repo := &dao.RepositoryEx{
		//Id       int
		Name:     "busybox",
		IsPublic: false,
		//Comment  string
		IsOrgRep: false,
		//Org:      org,
		User: user,
	}
	if err := repo.Save(); err != nil {
		t.Error(err)
	}

	expectedAccess := []*token.ResourceActions{
		&token.ResourceActions{
			Type:    "repository",
			Name:    "liugenping/busybox",
			Actions: []string{"*"},
		},
	}

	scope := "repository:liugenping/busybox:*"
	if a, err := testDeleteAuthorize(user, scope); err != nil {
		t.Fatal(err)
	} else {
		if !eqActionSlice(a, expectedAccess) {
			t.Log(*a[0])
			t.Log(*expectedAccess[0])
			t.Fatal("access is not eq expected")
		}
	}

	if exist, err := repo.Get(); err != nil {
		t.Fatal(err)
	} else if exist {
		t.Fatal("not delete repo")
	}

}

//input:
//      user: dao.SYSMEMBER, dao.ORGADMIN
//      orgnaization:  MemberPrivilege is w
//      orgnaizationusermap:  XXX
//      repository :private
//output:
//      error: nil
//      token: xxxxx
//      access: null
func Test_OrgNamespaceDelete1(t *testing.T) {
	user := &dao.User{
		Name:     "liugenping",
		Email:    "liugenping@huawei.com",
		Password: "liugenping",
		Status:   0,
		Role:     dao.SYSMEMBER,
	}
	if err := user.Save(); err != nil {
		t.Fatal(err)
	}
	user.Password = "liugenping"
	defer func() {
		if err := user.Delete(); err != nil {
			t.Fatal(err)
		}
	}()

	org := &dao.Organization{
		Name:            "huawei",
		MemberPrivilege: dao.WRITE,
	}
	if err := org.Save(); err != nil {
		t.Error(err)
	}
	defer func() {
		if err := org.Delete(); err != nil {
			t.Fatal(err)
		}
	}()

	orgUserMap := &dao.OrganizationUserMap{
		//ID      int           `orm:"auto"`
		User: user,
		Role: dao.ORGADMIN,
		Org:  org,
	}
	if err := orgUserMap.Save(); err != nil {
		t.Error(err)
	}

	repo := &dao.RepositoryEx{
		//Id       int
		Name:     "busybox",
		IsPublic: false,
		//Comment  string
		IsOrgRep: true,
		Org:      org,
		//User     *User
	}
	if err := repo.Save(); err != nil {
		t.Error(err)
	}

	expectedAccess := []*token.ResourceActions{
		&token.ResourceActions{
			Type:    "repository",
			Name:    "huawei/busybox",
			Actions: []string{"*"},
		},
	}

	scope := "repository:huawei/busybox:*"

	if a, err := testDeleteAuthorize(user, scope); err != nil {
		t.Fatal(err)
	} else {
		if !eqActionSlice(a, expectedAccess) {
			t.Log("token:", a[0])
			t.Log("expected:", expectedAccess[0])
			t.Fatal("access is not eq expected")
		}
	}

	if exist, err := repo.Get(); err != nil {
		t.Fatal(err)
	} else if exist {
		t.Fatal("not delete repo")
	}
}

//***************begin user namespace test***********************************
//test case 1
func Test_Login(t *testing.T) {
	user := &dao.User{
		Name:     "liugenping",
		Email:    "liugenping@huawei.com",
		Password: "liugenping",
		Status:   0,
		Role:     dao.SYSMEMBER,
	}

	if err := user.Save(); err != nil {
		t.Fatal(err)
	}
	user.Password = "liugenping"

	scope := ""
	if _, err := testAuthorize(user, scope); err != nil {
		t.Fatal(err)
	}

	if err := user.Delete(); err != nil {
		t.Fatal(err)
	}
}

//test case 2
func Test_UserNamespacePush(t *testing.T) {

	user := &dao.User{
		Name:     "liugenping",
		Email:    "liugenping@huawei.com",
		Password: "liugenping",
		Status:   0,
		Role:     dao.SYSMEMBER,
	}
	if err := user.Save(); err != nil {
		t.Fatal(err)
	}
	user.Password = "liugenping"
	defer func() {
		if err := user.Delete(); err != nil {
			t.Fatal(err)
		}
	}()

	expectedAccess := []*token.ResourceActions{
		&token.ResourceActions{
			Type:    "repository",
			Name:    "liugenping/busybox",
			Actions: []string{"push", "pull"},
		},
	}

	scope := "repository:liugenping/busybox:pull,push"
	if a, err := testAuthorize(user, scope); err != nil {
		t.Fatal(err)
	} else {
		if !eqActionSlice(a, expectedAccess) {
			t.Log(*a[0])
			t.Log(*expectedAccess[0])
			t.Fatal("access is not eq expected")
		}
	}

}

//test case 3
func Test_UserNamespacePull(t *testing.T) {
	user := &dao.User{
		Name:     "liugenping",
		Email:    "liugenping@huawei.com",
		Password: "liugenping",
		Status:   0,
		Role:     dao.SYSMEMBER,
	}
	if err := user.Save(); err != nil {
		t.Fatal(err)
	}
	user.Password = "liugenping"
	defer func() {
		if err := user.Delete(); err != nil {
			t.Fatal(err)
		}
	}()

	expectedAccess := []*token.ResourceActions{
		&token.ResourceActions{
			Type:    "repository",
			Name:    "liugenping/busybox",
			Actions: []string{"push", "pull"},
		},
	}

	scope := "repository:liugenping/busybox:push"
	if a, err := testAuthorize(user, scope); err != nil {
		t.Fatal(err)
	} else {
		if !eqActionSlice(a, expectedAccess) {
			t.Fatal("access is not eq expected")
		}
	}

	scope = "repository:liugenping/busybox:pull"
	if a, err := testAuthorize(user, scope); err != nil {
		t.Fatal(err)
	} else {
		if !eqActionSlice(a, expectedAccess) {
			t.Fatal("access is not eq expected")
		}
	}
}

//***************end user namespace test*************************************

//***************begin organization  namespace test**************************
//user not be created and organizaiton not be created
func Test_OrgNamespacePush1(t *testing.T) {
	user := &dao.User{
		Name:     "liugenping",
		Email:    "liugenping@huawei.com",
		Password: "liugenping",
		Status:   0,
		Role:     dao.SYSMEMBER,
	}

	scope := "repository:huawei/busybox:push,pull"
	if _, err := testAuthorize(user, scope); err == nil {
		t.Fatal("not pass for inexist user")
	} else {
		t.Log(err)
	}

}

//user = dao.SYSMEMBER(doesn't have create repo Privilege)
//&& not create orgnaization
func Test_OrgNamespacePush2(t *testing.T) {
	user := &dao.User{
		Name:     "liugenping",
		Email:    "liugenping@huawei.com",
		Password: "liugenping",
		Status:   0,
		Role:     dao.SYSMEMBER,
	}
	if err := user.Save(); err != nil {
		t.Fatal(err)
	}
	user.Password = "liugenping"
	defer func() {
		if err := user.Delete(); err != nil {
			t.Fatal(err)
		}
	}()

	scope := "repository:huawei/busybox:push,pull"
	if _, err := testAuthorize(user, scope); err == nil {
		t.Fatal("not pass for inexist org")
	} else {
		t.Log(err)
	}
}

//user = dao.SYSMEMBER(doesn't have create repo Privilege)
//create orgnaization
//not create repo
func Test_OrgNamespacePush3(t *testing.T) {
	user := &dao.User{
		Name:     "liugenping",
		Email:    "liugenping@huawei.com",
		Password: "liugenping",
		Status:   0,
		Role:     dao.SYSMEMBER,
	}
	if err := user.Save(); err != nil {
		t.Fatal(err)
	}
	user.Password = "liugenping"
	defer func() {
		if err := user.Delete(); err != nil {
			t.Fatal(err)
		}
	}()

	org := &dao.Organization{
		Name:            "huawei",
		MemberPrivilege: dao.WRITE,
	}

	if err := org.Save(); err != nil {
		t.Error(err)
	}
	defer func() {
		if err := org.Delete(); err != nil {
			t.Fatal(err)
		}
	}()

	scope := "repository:huawei/busybox:push,pull"
	if _, err := testAuthorize(user, scope); err == nil {
		t.Fatal("not pass for inexit repo")
	} else {
		t.Log(err)
	}
}

//input:
//      user: dao.SYSMEMBER, is not in organization
//      orgnaization:  MemberPrivilege is w
//      orgnaizationusermap:  null
//      repository :public
//output:
//      error: nil
//      token: xxxxx
//      access: pull
func Test_OrgNamespacePush4(t *testing.T) {
	user := &dao.User{
		Name:     "liugenping",
		Email:    "liugenping@huawei.com",
		Password: "liugenping",
		Status:   0,
		Role:     dao.SYSMEMBER,
	}
	if err := user.Save(); err != nil {
		t.Fatal(err)
	}
	user.Password = "liugenping"
	defer func() {
		if err := user.Delete(); err != nil {
			t.Fatal(err)
		}
	}()

	org := &dao.Organization{
		Name:            "huawei",
		MemberPrivilege: dao.WRITE,
	}
	if err := org.Save(); err != nil {
		t.Error(err)
	}
	defer func() {
		if err := org.Delete(); err != nil {
			t.Fatal(err)
		}
	}()

	repo := &dao.RepositoryEx{
		//Id       int
		Name:     "busybox",
		IsPublic: true,
		//Comment  string
		IsOrgRep: true,
		Org:      org,
		//User     *User
	}
	if err := repo.Save(); err != nil {
		t.Error(err)
	}
	defer func() {
		if err := repo.Delete(); err != nil {
			t.Fatal(err)
		}
	}()

	expectedAccess := []*token.ResourceActions{
		&token.ResourceActions{
			Type:    "repository",
			Name:    "huawei/busybox",
			Actions: []string{"pull"},
		},
	}

	scope := "repository:huawei/busybox:push,pull"
	if a, err := testAuthorize(user, scope); err != nil {
		t.Fatal(err)
	} else {
		if !eqActionSlice(a, expectedAccess) {
			t.Log("token:", a[0])
			t.Log("expected:", expectedAccess[0])
			t.Fatal("access is not eq expected")
		}
	}
}

//input:
//      user: dao.SYSMEMBER
//      orgnaization:  MemberPrivilege is w
//      orgnaizationusermap:  null
//      repository :private
//output:
//      error: nil
//      token: xxxxx
//      access: null
func Test_OrgNamespacePush5(t *testing.T) {
	user := &dao.User{
		Name:     "liugenping",
		Email:    "liugenping@huawei.com",
		Password: "liugenping",
		Status:   0,
		Role:     dao.SYSMEMBER,
	}
	if err := user.Save(); err != nil {
		t.Fatal(err)
	}
	user.Password = "liugenping"
	defer func() {
		if err := user.Delete(); err != nil {
			t.Fatal(err)
		}
	}()

	org := &dao.Organization{
		Name:            "huawei",
		MemberPrivilege: dao.WRITE,
	}
	if err := org.Save(); err != nil {
		t.Error(err)
	}
	defer func() {
		if err := org.Delete(); err != nil {
			t.Fatal(err)
		}
	}()

	repo := &dao.RepositoryEx{
		//Id       int
		Name:     "busybox",
		IsPublic: false,
		//Comment  string
		IsOrgRep: true,
		Org:      org,
		//User     *User
	}
	if err := repo.Save(); err != nil {
		t.Error(err)
	}
	defer func() {
		if err := repo.Delete(); err != nil {
			t.Fatal(err)
		}
	}()

	expectedAccess := []*token.ResourceActions{
		&token.ResourceActions{
			Type:    "repository",
			Name:    "huawei/busybox",
			Actions: []string{},
		},
	}

	scope := "repository:huawei/busybox:push,pull"
	if a, err := testAuthorize(user, scope); err != nil {
		t.Fatal(err)
	} else {
		if !eqActionSlice(a, expectedAccess) {
			t.Log("token:", a[0])
			t.Log("expected:", expectedAccess[0])
			t.Fatal("access is not eq expected")
		}
	}

}

//input:
//      user: dao.SYSMEMBER, dao.ORGADMIN
//      orgnaization:  MemberPrivilege is w
//      orgnaizationusermap:  user in  orgnaization
//      repository :private
//output:
//      error: nil
//      token: xxxxx
//      access: push, pull
func Test_OrgNamespacePush6(t *testing.T) {
	user := &dao.User{
		Name:     "liugenping",
		Email:    "liugenping@huawei.com",
		Password: "liugenping",
		Status:   0,
		Role:     dao.SYSMEMBER,
	}
	if err := user.Save(); err != nil {
		t.Fatal(err)
	}
	user.Password = "liugenping"
	defer func() {
		if err := user.Delete(); err != nil {
			t.Fatal(err)
		}
	}()

	org := &dao.Organization{
		Name:            "huawei",
		MemberPrivilege: dao.WRITE,
	}
	if err := org.Save(); err != nil {
		t.Error(err)
	}
	defer func() {
		if err := org.Delete(); err != nil {
			t.Fatal(err)
		}
	}()

	orgUserMap := &dao.OrganizationUserMap{
		//ID      int           `orm:"auto"`
		User: user,
		Role: dao.ORGADMIN,
		Org:  org,
	}
	if err := orgUserMap.Save(); err != nil {
		t.Fatal(err)
	}

	repo := &dao.RepositoryEx{
		//Id       int
		Name:     "busybox",
		IsPublic: false,
		//Comment  string
		IsOrgRep: true,
		Org:      org,
		//User     *User
	}
	if err := repo.Save(); err != nil {
		t.Error(err)
	}
	defer func() {
		if err := repo.Delete(); err != nil {
			t.Fatal(err)
		}
	}()

	expectedAccess := []*token.ResourceActions{
		&token.ResourceActions{
			Type:    "repository",
			Name:    "huawei/busybox",
			Actions: []string{"push", "pull"},
		},
	}
	scope := "repository:huawei/busybox:push,pull"
	if a, err := testAuthorize(user, scope); err != nil {
		t.Fatal(err)
	} else {
		if !eqActionSlice(a, expectedAccess) {
			t.Log("token:", a[0])
			t.Log("expected:", expectedAccess[0])
			t.Fatal("access is not eq expected")
		}
	}
}

//input:
//      user: dao.SYSMEMBER, dao.ORGADMIN
//      orgnaization:  MemberPrivilege is w
//      orgnaizationusermap:  user in  orgnaization
//      repository :nil
//output:
//      error: nil
//      token: xxxxx
//      access: push, pull
func Test_OrgNamespacePush7(t *testing.T) {
	user := &dao.User{
		Name:     "liugenping",
		Email:    "liugenping@huawei.com",
		Password: "liugenping",
		Status:   0,
		Role:     dao.SYSMEMBER,
	}
	if err := user.Save(); err != nil {
		t.Fatal(err)
	}
	user.Password = "liugenping"
	defer func() {
		if err := user.Delete(); err != nil {
			t.Fatal(err)
		}
	}()

	org := &dao.Organization{
		Name:            "huawei",
		MemberPrivilege: dao.WRITE,
	}
	if err := org.Save(); err != nil {
		t.Error(err)
	}
	defer func() {
		if err := org.Delete(); err != nil {
			t.Fatal(err)
		}
	}()

	orgUserMap := &dao.OrganizationUserMap{
		//ID      int           `orm:"auto"`
		User: user,
		Role: dao.ORGADMIN,
		Org:  org,
	}
	if err := orgUserMap.Save(); err != nil {
		t.Fatal(err)
	}

	expectedAccess := []*token.ResourceActions{
		&token.ResourceActions{
			Type:    "repository",
			Name:    "huawei/busybox",
			Actions: []string{"push", "pull"},
		},
		&token.ResourceActions{
			Type:    "repository",
			Name:    "huawei/ubuntu",
			Actions: []string{"push", "pull"},
		},
	}
	scope := "repository:huawei/busybox:push,pull repository:huawei/ubuntu:push,pull"
	if a, err := testAuthorize(user, scope); err != nil {
		t.Fatal(err)
	} else {
		if !eqActionSlice(a, expectedAccess) {
			t.Log("token:", a[0])
			t.Log("expected:", expectedAccess[0])
			t.Log("token:", a[1])
			t.Log("expected:", expectedAccess[1])
			t.Fatal("access is not eq expected")
		}
	}
}

//input:
//      user: dao.SYSMEMBER, dao.ORGMEMBER
//      orgnaization:  MemberPrivilege is NONE
//      orgnaizationusermap:  user in  orgnaization
//      repository :org/busybox
//output:
//      error: nil
//      token: xxxxx
//      access: push, pull
func Test_OrgNamespacePush8(t *testing.T) {
	user := &dao.User{
		Name:     "liugenping",
		Email:    "liugenping@huawei.com",
		Password: "liugenping",
		Status:   0,
		Role:     dao.SYSMEMBER,
	}
	if err := user.Save(); err != nil {
		t.Fatal(err)
	}
	user.Password = "liugenping"
	defer func() {
		if err := user.Delete(); err != nil {
			t.Fatal(err)
		}
	}()

	org := &dao.Organization{
		Name:            "huawei",
		MemberPrivilege: dao.NONE,
	}
	if err := org.Save(); err != nil {
		t.Error(err)
	}
	defer func() {
		if err := org.Delete(); err != nil {
			t.Fatal(err)
		}
	}()

	orgUserMap := &dao.OrganizationUserMap{
		//ID      int           `orm:"auto"`
		User: user,
		Role: dao.ORGMEMBER,
		Org:  org,
	}
	if err := orgUserMap.Save(); err != nil {
		t.Fatal(err)
	}

	repo := &dao.RepositoryEx{
		//Id       int
		Name:     "busybox",
		IsPublic: false,
		//Comment  string
		IsOrgRep: true,
		Org:      org,
		//User     *User
	}
	if err := repo.Save(); err != nil {
		t.Error(err)
	}
	defer func() {
		if err := repo.Delete(); err != nil {
			t.Fatal(err)
		}
	}()

	expectedAccess := []*token.ResourceActions{
		&token.ResourceActions{
			Type:    "repository",
			Name:    "huawei/busybox",
			Actions: []string{},
		},
	}
	scope := "repository:huawei/busybox:push,pull"
	if a, err := testAuthorize(user, scope); err != nil {
		t.Fatal(err)
	} else {
		if !eqActionSlice(a, expectedAccess) {
			t.Log("token:", a[0])
			t.Log("expected:", expectedAccess[0])
			t.Fatal("access is not eq expected")
		}
	}
}

//input:
//      user: dao.SYSMEMBER, dao.ORGMEMBER
//      orgnaization:  MemberPrivilege is NONE
//      orgnaizationusermap:  user in  orgnaization
//      repository :org/busybox
//output:
//      error: nil
//      token: xxxxx
//      access: push, pull
func Test_OrgNamespacePush9(t *testing.T) {
	user := &dao.User{
		Name:     "liugenping",
		Email:    "liugenping@huawei.com",
		Password: "liugenping",
		Status:   0,
		Role:     dao.SYSMEMBER,
	}
	if err := user.Save(); err != nil {
		t.Fatal(err)
	}
	user.Password = "liugenping"
	defer func() {
		if err := user.Delete(); err != nil {
			t.Fatal(err)
		}
	}()

	org := &dao.Organization{
		Name:            "huawei",
		MemberPrivilege: dao.NONE,
	}
	if err := org.Save(); err != nil {
		t.Error(err)
	}
	defer func() {
		if err := org.Delete(); err != nil {
			t.Fatal(err)
		}
	}()

	orgUserMap := &dao.OrganizationUserMap{
		//ID      int           `orm:"auto"`
		User: user,
		Role: dao.ORGMEMBER,
		Org:  org,
	}
	if err := orgUserMap.Save(); err != nil {
		t.Fatal(err)
	}

	repo := &dao.RepositoryEx{
		//Id       int
		Name:     "busybox",
		IsPublic: false,
		//Comment  string
		IsOrgRep: true,
		Org:      org,
		//User     *User
	}
	if err := repo.Save(); err != nil {
		t.Error(err)
	}
	defer func() {
		if err := repo.Delete(); err != nil {
			t.Fatal(err)
		}
	}()

	team := &dao.Team{
		//Id        int           `orm:"auto"`
		Name: "teamr",
		//Comment   string        `orm:"size(100);null"`
		Org: org,
	}
	if err := team.Save(); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := team.Delete(); err != nil {
			t.Fatal(err)
		}
	}()

	tum := &dao.TeamUserMap{
		Team: team,
		User: user,
		Role: dao.TEAMMEMBER,
	}

	if err := tum.Save(); err != nil {
		t.Fatal(err)
	}

	trm := &dao.TeamRepositoryMap{
		//Id      int           `orm:"auto"`
		Team:   team,
		Repo:   repo,
		Permit: dao.WRITE, //team's access permit for repository， 1:write,2:read.
	}
	if err := trm.Save(); err != nil {
		t.Fatal(err)
	}

	expectedAccess := []*token.ResourceActions{
		&token.ResourceActions{
			Type:    "repository",
			Name:    "huawei/busybox",
			Actions: []string{"push", "pull"},
		},
	}
	scope := "repository:huawei/busybox:push,pull"
	if a, err := testAuthorize(user, scope); err != nil {
		t.Fatal(err)
	} else {
		if !eqActionSlice(a, expectedAccess) {
			t.Log("token:", a[0])
			t.Log("expected:", expectedAccess[0])
			t.Fatal("access is not eq expected")
		}
	}
}

//***************begin organization  namespace test**************************
func testAuthorize(user *dao.User, scope string) ([]*token.ResourceActions, error) {
	/*
		resActions, err := getResourceActions(scope)
		if err != nil {
			return fmt.Errorf("parse scope error")
		}
	*/
	rt, jsonToken := Authz.GetAuthorize(user.Name, user.Password, service, scope)
	if rt != http.StatusOK {
		return nil, fmt.Errorf(string(jsonToken))
	}
	return verifyToken(jsonToken)
}

func testDeleteAuthorize(user *dao.User, scope string) ([]*token.ResourceActions, error) {
	rt, jsonToken := Authz.DeleteAuthorize(user.Name, user.Password, service, scope, "true")
	if rt != http.StatusOK {
		return nil, fmt.Errorf(string(jsonToken))
	}
	return verifyToken(jsonToken)
}

func verifyToken(jsonToken []byte) ([]*token.ResourceActions, error) {
	rawToken := map[string]string{}
	if err := json.Unmarshal(jsonToken, &rawToken); err != nil {
		return nil, err
	}

	tk, err := token.NewToken(rawToken["token"])
	if err != nil {
		return nil, err
	}

	verifyOpts := token.VerifyOptions{
		TrustedIssuers:    []string{setting.Issuer},
		AcceptedAudiences: []string{service},
		Roots:             rootCerts,
		TrustedKeys:       publicKeys,
	}

	if err = tk.Verify(verifyOpts); err != nil {
		return nil, err
	}

	return tk.Claims.Access, nil
}

func ParseCertAndPublicKey(rootCertBundle string) (map[string]libtrust.PublicKey, *x509.CertPool, error) {
	fp, err := os.Open(rootCertBundle)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to open token auth root certificate bundle file %q: %s", rootCertBundle, err)
	}
	defer fp.Close()

	rawCertBundle, err := ioutil.ReadAll(fp)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to read token auth root certificate bundle file %q: %s", rootCertBundle, err)
	}

	var rootCerts []*x509.Certificate
	pemBlock, rawCertBundle := pem.Decode(rawCertBundle)
	for pemBlock != nil {
		cert, err := x509.ParseCertificate(pemBlock.Bytes)
		if err != nil {
			return nil, nil, fmt.Errorf("unable to parse token auth root certificate: %s", err)
		}

		rootCerts = append(rootCerts, cert)

		pemBlock, rawCertBundle = pem.Decode(rawCertBundle)
	}

	if len(rootCerts) == 0 {
		return nil, nil, fmt.Errorf("token auth requires at least one token signing root certificate")
	}

	rootPool := x509.NewCertPool()
	trustedKeys := make(map[string]libtrust.PublicKey, len(rootCerts))
	for _, rootCert := range rootCerts {
		rootPool.AddCert(rootCert)
		pubKey, err := libtrust.FromCryptoPublicKey(crypto.PublicKey(rootCert.PublicKey))
		if err != nil {
			return nil, nil, fmt.Errorf("unable to get public key from token auth root certificate: %s", err)
		}
		trustedKeys[pubKey.KeyID()] = pubKey
	}

	return trustedKeys, rootPool, nil
}

func eqActionStruct(access *token.ResourceActions, expectedAccess *token.ResourceActions) bool {
	if access.Name != expectedAccess.Name || access.Type != expectedAccess.Type {
		return false
	}

	if len(access.Actions) != len(expectedAccess.Actions) {
		return false
	}

	i := len(access.Actions)
	for _, a := range access.Actions {
		for _, b := range expectedAccess.Actions {
			if a == b {
				i--
			}
		}
	}
	if i > 0 {
		return false
	} else {
		return true
	}
}

func eqActionSlice(access []*token.ResourceActions, expectedAccess []*token.ResourceActions) bool {
	if len(access) != len(expectedAccess) {
		return false
	}
	i := len(access)
	for _, a := range access {
		for _, b := range expectedAccess {
			if eqActionStruct(a, b) {
				i--
			}
		}
	}
	if i > 0 {
		return false
	} else {
		return true
	}
}
