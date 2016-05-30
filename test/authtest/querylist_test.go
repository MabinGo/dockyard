package authtest

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/huawei-openlab/newdb/orm"

	"github.com/containerops/dockyard/auth/controller"
	"github.com/containerops/dockyard/auth/dao"
	"github.com/containerops/dockyard/utils/setting"
)

func createqueryuser() (dao.User, error) {
	queryuser := dao.User{
		Name:     "queryuser",
		Email:    "queryuser@mail.com",
		Password: "queryuser",
		//RealName string `orm:"size(100);null"`
		//Comment  string `orm:"size(100);null"`
		Status: 0,
		Role:   2,
	}
	requestUrl := setting.ListenMode + "://" + Domains + "/" + "uam" + "/" + "user" + "/" + "signup"
	rst, err := json.Marshal(queryuser)
	if err != nil {
		return dao.User{}, err
	}
	body := bytes.NewReader(rst)
	if _, err := querylisthttp("POST", requestUrl, body, "", ""); err != nil {
		return dao.User{}, err
	}
	return queryuser, nil
}

func createqueryorg(name, password string) (dao.Organization, error) {
	queryorg := dao.Organization{
		Name:            "queryorg",
		MemberPrivilege: dao.WRITE,
	}
	requestUrl := setting.ListenMode + "://" + Domains + "/" + "uam" + "/" + "organization" + "/"
	rst, err := json.Marshal(queryorg)
	if err != nil {
		return dao.Organization{}, err
	}
	body := bytes.NewReader(rst)
	if _, err := querylisthttp("POST", requestUrl, body, name, password); err != nil {
		return dao.Organization{}, err
	}
	return queryorg, nil
}

func querylisthttp(method string, url string, body io.Reader, name, password string) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return &http.Response{}, err
	}
	req.SetBasicAuth(name, password)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	return client.Do(req)
}

func orglistres(name, password string) ([]dao.Organization, error) {
	requestUrl := setting.ListenMode + "://" + Domains + "/" + "uam" + "/" + "organization" + "/" + "list"
	rsp, err := querylisthttp("GET", requestUrl, nil, name, password)
	if err != nil {
		return []dao.Organization{}, err
	}

	defer rsp.Body.Close()
	if rspbody, err := ioutil.ReadAll(rsp.Body); err != nil {
		return []dao.Organization{}, err
	} else {
		rsporg := []dao.Organization{}
		if err := json.Unmarshal(rspbody, &rsporg); err != nil {
			return []dao.Organization{}, err
		} else {
			return rsporg, nil
		}
	}
}

func oumJSONlistres(name, password, orgname string) ([]controller.OrganizationUserMapJSON, error) {
	requestUrl := setting.ListenMode + "://" + Domains + "/" + "uam" + "/" + "organization" + "/" + orgname + "/" + "listuser"
	rsp, err := querylisthttp("GET", requestUrl, nil, name, password)
	if err != nil {
		return []controller.OrganizationUserMapJSON{}, err
	}

	defer rsp.Body.Close()
	if rspbody, err := ioutil.ReadAll(rsp.Body); err != nil {
		return []controller.OrganizationUserMapJSON{}, err
	} else {
		rspoumJSON := []controller.OrganizationUserMapJSON{}
		if err := json.Unmarshal(rspbody, &rspoumJSON); err != nil {
			return []controller.OrganizationUserMapJSON{}, err
		} else {
			return rspoumJSON, nil
		}
	}
}

func repolistres(name, password, namespace string) ([]dao.RepositoryEx, error) {
	requestUrl := setting.ListenMode + "://" + Domains + "/" + "uam" + "/" + "repository" + "/" + namespace + "/" + "list"
	rsp, err := querylisthttp("GET", requestUrl, nil, name, password)
	if err != nil {
		return []dao.RepositoryEx{}, err
	}

	defer rsp.Body.Close()
	if rspbody, err := ioutil.ReadAll(rsp.Body); err != nil {
		return []dao.RepositoryEx{}, err
	} else {
		rsprepo := []dao.RepositoryEx{}
		if err := json.Unmarshal(rspbody, &rsprepo); err != nil {
			return []dao.RepositoryEx{}, err
		} else {
			return rsprepo, nil
		}
	}
}

func teamJSONlistres(name, password, orgname string) ([]controller.TeamJSON, error) {
	requestUrl := setting.ListenMode + "://" + Domains + "/" + "uam" + "/" + "team" + "/" + orgname + "/" + "listteam"
	rsp, err := querylisthttp("GET", requestUrl, nil, name, password)
	if err != nil {
		return nil, err
	}

	defer rsp.Body.Close()
	if rspbody, err := ioutil.ReadAll(rsp.Body); err != nil {
		return nil, err
	} else {
		rspteamJSON := []controller.TeamJSON{}
		if err := json.Unmarshal(rspbody, &rspteamJSON); err != nil {
			return nil, err
		} else {
			return rspteamJSON, nil
		}
	}
}

func tumJSONlistres(name, password, orgname, teamname string) ([]controller.TeamUserMapJSON, error) {
	requestUrl := setting.ListenMode + "://" + Domains + "/" + "uam" + "/" + "team" + "/" + orgname + "/" + teamname + "/" + "listuser"
	rsp, err := querylisthttp("GET", requestUrl, nil, name, password)
	if err != nil {
		return []controller.TeamUserMapJSON{}, err
	}

	defer rsp.Body.Close()
	if rspbody, err := ioutil.ReadAll(rsp.Body); err != nil {
		return []controller.TeamUserMapJSON{}, err
	} else {
		rsptumJSON := []controller.TeamUserMapJSON{}
		if err := json.Unmarshal(rspbody, &rsptumJSON); err != nil {
			return []controller.TeamUserMapJSON{}, err
		} else {
			return rsptumJSON, nil
		}
	}
}

func trmJSONlistres(name, password, orgname, teamname string) ([]controller.TeamRepositoryMapJSON, error) {
	requestUrl := setting.ListenMode + "://" + Domains + "/" + "uam" + "/" + "team" + "/" + orgname + "/" + teamname + "/" + "listrepository"
	rsp, err := querylisthttp("GET", requestUrl, nil, name, password)
	if err != nil {
		return []controller.TeamRepositoryMapJSON{}, err
	}

	defer rsp.Body.Close()
	if rspbody, err := ioutil.ReadAll(rsp.Body); err != nil {
		return []controller.TeamRepositoryMapJSON{}, err
	} else {
		rsptrmJSON := []controller.TeamRepositoryMapJSON{}
		if err := json.Unmarshal(rspbody, &rsptrmJSON); err != nil {
			return []controller.TeamRepositoryMapJSON{}, err
		} else {
			return rsptrmJSON, nil
		}
	}
}

func frepolistres(name, password, fuzzyrepo string) ([]dao.RepositoryEx, error) {
	requestUrl := setting.ListenMode + "://" + Domains + "/" + "uam" + "/" + "repository" + "/" + fuzzyrepo + "/" + "fuzzylist"
	rsp, err := querylisthttp("GET", requestUrl, nil, name, password)
	if err != nil {
		return nil, err
	}

	defer rsp.Body.Close()
	if rspbody, err := ioutil.ReadAll(rsp.Body); err != nil {
		return nil, err
	} else {
		rsprepo := []dao.RepositoryEx{}
		if err := json.Unmarshal(rspbody, &rsprepo); err != nil {
			return nil, err
		} else {
			return rsprepo, nil
		}
	}
}

func cleartable(tablelist []string) error {
	for _, table := range tablelist {
		o := orm.NewOrm()
		SQL := "delete from " + table
		if _, err := o.Raw(SQL).Exec(); err != nil {
			return err
		}
	}
	return nil
}

func TestGetOrganizationList(t *testing.T) {
	//clear user and organization table
	if err := cleartable([]string{"user", "organization"}); err != nil {
		t.Error(err)
	}

	//user sign up
	queryuser, err := createqueryuser()
	if err != nil {
		t.Error(err)
	}

	//get empty organization list
	if rsplist, err := orglistres(queryuser.Name, queryuser.Password); err != nil {
		t.Error(err)
	} else if len(rsplist) != 0 {
		t.Error(fmt.Errorf("List is not empty"))
	}

	//create organizaiton
	if _, err := createqueryorg(queryuser.Name, queryuser.Password); err != nil {
		t.Error(err)
	}

	//get organization list
	if rsplist, err := orglistres(queryuser.Name, queryuser.Password); err != nil {
		t.Error(err)
	} else if len(rsplist) == 0 {
		t.Error(fmt.Errorf("List is empty"))
	}
}

func TestGetUserListFromOrganization(t *testing.T) {
	//clear user, organization and organization_user_map table
	if err := cleartable([]string{"user", "organization", "organization_user_map"}); err != nil {
		t.Error(err)
	}

	//user sign up
	queryuser, err := createqueryuser()
	if err != nil {
		t.Error(err)
	}

	//create organizaiton
	queryorg, err := createqueryorg(queryuser.Name, queryuser.Password)
	if err != nil {
		t.Error(err)
	}

	//get user list from organization
	if rsplist, err := oumJSONlistres(queryuser.Name, queryuser.Password, queryorg.Name); err != nil {
		t.Error(err)
	} else if len(rsplist) == 0 {
		t.Error(fmt.Errorf("List is empty"))
	}
}

func TestGetRepositoryListFromOrganization(t *testing.T) {
	//clear user, organization, repository_ex and organization_user_map table
	if err := cleartable([]string{"user", "organization", "organization_user_map", "repository_ex"}); err != nil {
		t.Error(err)
	}

	//user sign up
	queryuser, err := createqueryuser()
	if err != nil {
		t.Error(err)
	}

	//create organizaiton
	queryorg, err := createqueryorg(queryuser.Name, queryuser.Password)
	if err != nil {
		t.Error(err)
	}

	//get empty repo list from organization
	if rsplist, err := repolistres(queryuser.Name, queryuser.Password, queryorg.Name); err != nil {
		t.Error(err)
	} else if len(rsplist) != 0 {
		t.Error(fmt.Errorf("List is not empty"))
	}

	//create repository in organization
	queryorgrepo := controller.RepositoryJSON{
		Name:     "queryorgrepo",
		IsPublic: true,
		Comment:  "comment",
		OrgName:  queryorg.Name,
	}
	requestUrl := setting.ListenMode + "://" + Domains + "/" + "uam" + "/" + "repository" + "/"
	rst, err := json.Marshal(queryorgrepo)
	if err != nil {
		t.Fatal(err.Error())
	}
	body := bytes.NewReader(rst)
	if _, err := querylisthttp("POST", requestUrl, body, queryuser.Name, queryuser.Password); err != nil {
		t.Error(err)
	}

	//get repo list from organization
	if rsplist, err := repolistres(queryuser.Name, queryuser.Password, queryorg.Name); err != nil {
		t.Error(err)
	} else if len(rsplist) == 0 {
		t.Error(fmt.Errorf("List is empty"))
	}

	//get empty repo list from user
	if rsplist, err := repolistres(queryuser.Name, queryuser.Password, queryuser.Name); err != nil {
		t.Error(err)
	} else if len(rsplist) != 0 {
		t.Error(fmt.Errorf("List is not empty"))
	}

	//create repository in organization
	queryuserrepo := controller.RepositoryJSON{
		Name:     "queryuserrepo",
		IsPublic: true,
		Comment:  "comment",
		UserName: queryuser.Name,
	}
	requestUrl = setting.ListenMode + "://" + Domains + "/" + "uam" + "/" + "repository" + "/"
	rst, err = json.Marshal(queryuserrepo)
	if err != nil {
		t.Fatal(err.Error())
	}
	body = bytes.NewReader(rst)
	if _, err := querylisthttp("POST", requestUrl, body, queryuser.Name, queryuser.Password); err != nil {
		t.Error(err)
	}

	//get repo list from user
	if rsplist, err := repolistres(queryuser.Name, queryuser.Password, queryuser.Name); err != nil {
		t.Error(err)
	} else if len(rsplist) == 0 {
		t.Error(fmt.Errorf("List is empty"))
	}
}

func TestGetTeamListFromOrganization(t *testing.T) {
	//clear user ,organization and team table
	if err := cleartable([]string{"user", "organization", "team"}); err != nil {
		t.Error(err)
	}

	//user sign up
	queryuser, err := createqueryuser()
	if err != nil {
		t.Error(err)
	}

	//create organizaiton
	queryorg, err := createqueryorg(queryuser.Name, queryuser.Password)
	if err != nil {
		t.Error(err)
	}

	//get empty team list from organization
	if rsplist, err := teamJSONlistres(queryuser.Name, queryuser.Password, queryorg.Name); err != nil {
		t.Error(err)
	} else if len(rsplist) != 0 {
		t.Error(fmt.Errorf("List is not empty"))
	}

	//create team in organization
	queryteam := controller.TeamJSON{
		TeamName: "queryteam",
		Comment:  "comment",
		OrgName:  queryorg.Name,
	}
	requestUrl := setting.ListenMode + "://" + Domains + "/" + "uam" + "/" + "team" + "/"
	rst, err := json.Marshal(queryteam)
	if err != nil {
		t.Fatal(err.Error())
	}
	body := bytes.NewReader(rst)
	if _, err := querylisthttp("POST", requestUrl, body, queryuser.Name, queryuser.Password); err != nil {
		t.Error(err)
	}

	//get team list from organization
	if rsplist, err := teamJSONlistres(queryuser.Name, queryuser.Password, queryorg.Name); err != nil {
		t.Error(err)
	} else if len(rsplist) == 0 {
		t.Error(fmt.Errorf("List is empty"))
	}
}

func TestGetUserListFromTeam(t *testing.T) {
	//clear user, organization, team_user_map and team table
	if err := cleartable([]string{"user", "organization", "team_user_map", "team"}); err != nil {
		t.Error(err)
	}

	//user sign up
	queryuser, err := createqueryuser()
	if err != nil {
		t.Error(err)
	}

	//create organizaiton
	queryorg, err := createqueryorg(queryuser.Name, queryuser.Password)
	if err != nil {
		t.Error(err)
	}

	//create team in organization
	queryteam := controller.TeamJSON{
		TeamName: "queryteam",
		Comment:  "comment",
		OrgName:  queryorg.Name,
	}
	requestUrl := setting.ListenMode + "://" + Domains + "/" + "uam" + "/" + "team" + "/"
	rst, err := json.Marshal(queryteam)
	if err != nil {
		t.Fatal(err.Error())
	}
	body := bytes.NewReader(rst)
	if _, err := querylisthttp("POST", requestUrl, body, queryuser.Name, queryuser.Password); err != nil {
		t.Error(err)
	}

	//get user list from team
	if rsplist, err := tumJSONlistres(queryuser.Name, queryuser.Password, queryorg.Name, queryteam.TeamName); err != nil {
		t.Error(err)
	} else if len(rsplist) == 0 {
		t.Error(fmt.Errorf("List is empty"))
	}
}

func TestGetRepoListFromTeam(t *testing.T) {
	//clear user, organization, team_user_map and team table
	if err := cleartable([]string{"user", "organization", "team_repository_map", "team"}); err != nil {
		t.Error(err)
	}

	//user sign up
	queryuser, err := createqueryuser()
	if err != nil {
		t.Error(err)
	}

	//create organizaiton
	queryorg, err := createqueryorg(queryuser.Name, queryuser.Password)
	if err != nil {
		t.Error(err)
	}

	//create repository in organization
	queryrepo := controller.RepositoryJSON{
		Name:     "queryrepo",
		IsPublic: true,
		Comment:  "comment",
		OrgName:  queryorg.Name,
	}
	requestUrl := setting.ListenMode + "://" + Domains + "/" + "uam" + "/" + "repository" + "/"
	rst, err := json.Marshal(queryrepo)
	if err != nil {
		t.Fatal(err.Error())
	}
	body := bytes.NewReader(rst)
	if _, err := querylisthttp("POST", requestUrl, body, queryuser.Name, queryuser.Password); err != nil {
		t.Error(err)
	}

	//create team in organization
	queryteam := controller.TeamJSON{
		TeamName: "queryteam",
		Comment:  "comment",
		OrgName:  queryorg.Name,
	}
	requestUrl = setting.ListenMode + "://" + Domains + "/" + "uam" + "/" + "team" + "/"
	rst, err = json.Marshal(queryteam)
	if err != nil {
		t.Fatal(err.Error())
	}
	body = bytes.NewReader(rst)
	if _, err := querylisthttp("POST", requestUrl, body, queryuser.Name, queryuser.Password); err != nil {
		t.Error(err)
	}

	//get empty list from team
	if rsplist, err := trmJSONlistres(queryuser.Name, queryuser.Password, queryorg.Name, queryteam.TeamName); err != nil {
		t.Error(err)
	} else if len(rsplist) != 0 {
		t.Error(fmt.Errorf("List is not empty"))
	}

	//create repository in team
	trm := controller.TeamRepositoryMapJSON{
		OrgName:  queryorg.Name,
		RepoName: queryrepo.Name,
		TeamName: queryteam.TeamName,
		Permit:   dao.WRITE,
	}
	requestUrl = setting.ListenMode + "://" + Domains + "/" + "uam" + "/" + "team" + "/" + "addrepository" + "/"
	rst, err = json.Marshal(trm)
	if err != nil {
		t.Fatal(err.Error())
	}
	body = bytes.NewReader(rst)
	if _, err := querylisthttp("POST", requestUrl, body, queryuser.Name, queryuser.Password); err != nil {
		t.Error(err)
	}

	//get repository list from team
	if rsplist, err := trmJSONlistres(queryuser.Name, queryuser.Password, queryorg.Name, queryteam.TeamName); err != nil {
		t.Error(err)
	} else if len(rsplist) == 0 {
		t.Error(fmt.Errorf("List is empty"))
	}
}

func TestGetFuzzyRepoList(t *testing.T) {
	//clear user, organization, team_user_map and team table
	if err := cleartable([]string{"user", "organization", "repository_ex"}); err != nil {
		t.Error(err)
	}

	//user sign up
	queryuser, err := createqueryuser()
	if err != nil {
		t.Error(err)
	}

	//create organizaiton
	queryorg, err := createqueryorg(queryuser.Name, queryuser.Password)
	if err != nil {
		t.Error(err)
	}

	//create repository in organization
	queryrepo := controller.RepositoryJSON{
		Name:     "queryrepo",
		IsPublic: true,
		Comment:  "comment",
		OrgName:  queryorg.Name,
	}
	requestUrl := setting.ListenMode + "://" + Domains + "/" + "uam" + "/" + "repository" + "/"
	rst, err := json.Marshal(queryrepo)
	if err != nil {
		t.Fatal(err.Error())
	}
	body := bytes.NewReader(rst)
	if _, err := querylisthttp("POST", requestUrl, body, queryuser.Name, queryuser.Password); err != nil {
		t.Error(err)
	}

	//get fuzzy repository list
	if rsplist, err := frepolistres(queryuser.Name, queryuser.Password, "repo"); err != nil {
		t.Error(err)
	} else if len(rsplist) == 0 {
		t.Error(fmt.Errorf("List is empty"))
	}
}

func TestCreateAdmin(t *testing.T) {
	u := &dao.User{
		Name:     "root",
		Email:    "root@rootroot56789.com",
		Password: "root",
		Comment:  "administrator for dockyard system",
		Status:   0,
		Role:     dao.SYSADMIN,
	}
	if exist, err := u.Get(); err != nil {
		t.Error(err)
	} else if !exist {
		if err := u.Save(); err != nil {
			t.Error(err)
		}
	}
}
