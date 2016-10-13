package appv1

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/containerops/dockyard/tests/api"
)

type Body struct {
	Namespace   string    `json:"namespace"`
	Repository  string    `json:"repository"`
	OS          string    `json:"os"`
	Arch        string    `json:"arch"`
	Name        string    `json:"name"`
	Tag         string    `json:"tag"`
	Description string    `json:"description"`
	URL         string    `json:"url"`
	Size        int64     `json:"size"`
	CreatedAt   time.Time `json:"created"`
	UpdatedAt   time.Time `json:"updated"`
}

func TestSearchInit(t *testing.T) {
	app := api.AppV1App{
		OS:   "linux",
		Arch: "arm",
		App:  fileName,
		Tag:  "latest",
	}

	repo := api.AppV1Repo{
		URI:        api.DockyardURI,
		Namespace:  api.UserName,
		Repository: "searchapp",
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

func TestSearchGlobal(t *testing.T) {
	repo := api.AppV1Repo{
		URI:        api.DockyardURI,
		Namespace:  api.UserName,
		Repository: "searchapp",
	}

	query := []string{fileName, "latest"}

	byteArr, code, err := repo.SearchGlobal(query, api.Token)
	if err != nil || code != 200 {
		if code == 0 {
			t.Logf("System (not dockyard) error :%s", err)
			return
		}
		t.Fatalf(errStr, code, err, "search global")
		return
	}
	results := []Body{}
	if err := json.Unmarshal(byteArr, &results); err != nil {
		t.Fatalf(errStr, code, err, "search global format json")
		return
	}
	for _, v := range results {
		for _, str := range query {
			if !strings.Contains(v.URL, str) {
				t.Fatalf(errStr, code, err, "search global result is wrong")
				return
			}
		}
	}
}

func TestSearchScope(t *testing.T) {
	repo := api.AppV1Repo{
		URI:        api.DockyardURI,
		Namespace:  api.UserName,
		Repository: "searchapp",
	}

	query := []string{fileName}

	byteArr, code, err := repo.SearchScoped(query, api.Token)
	if err != nil || code != 200 {
		if code == 0 {
			t.Logf("System (not dockyard) error :%s", err)
			return
		}
		t.Fatalf(errStr, code, err, "search scope")
		return
	}
	results := []Body{}
	if err := json.Unmarshal(byteArr, &results); err != nil {
		t.Fatalf(errStr, code, err, "search scope format json")
		return
	}
	for _, v := range results {
		if v.Namespace != repo.Namespace || v.Repository != repo.Repository {
			t.Fatalf(errStr, code, err, "search scope result is wrong")
			return
		}
		for _, str := range query {
			if !strings.Contains(v.URL, str) {
				t.Fatalf(errStr, code, err, "search scope result is wrong")
				return
			}
		}
	}
}

func TestSearchScopeTwoParse(t *testing.T) {
	repo := api.AppV1Repo{
		URI:        api.DockyardURI,
		Namespace:  api.UserName,
		Repository: "searchapp",
	}

	query := []string{fileName, "latest"}

	byteArr, code, err := repo.SearchScoped(query, api.Token)
	if err != nil || code != 200 {
		if code == 0 {
			t.Logf("System (not dockyard) error :%s", err)
			return
		}
		t.Fatalf(errStr, code, err, "search scope two parse")
		return
	}
	results := []Body{}
	if err := json.Unmarshal(byteArr, &results); err != nil {
		t.Fatalf(errStr, code, err, "search scope two parse format json")
		return
	}
	for _, v := range results {
		if v.Namespace != repo.Namespace || v.Repository != repo.Repository {
			t.Fatalf(errStr, code, err, "search scope two parse result is wrong")
			return
		}
		for _, str := range query {
			if !strings.Contains(v.URL, str) {
				t.Fatalf(errStr, code, err, "search scope two parse result is wrong")
				return
			}
		}
	}
}

func TestSearchScopeThreeParse(t *testing.T) {
	repo := api.AppV1Repo{
		URI:        api.DockyardURI,
		Namespace:  api.UserName,
		Repository: "searchapp",
	}

	query := []string{fileName, "latest", "linux"}

	byteArr, code, err := repo.SearchScoped(query, api.Token)
	if err != nil || code != 200 {
		if code == 0 {
			t.Logf("System (not dockyard) error :%s", err)
			return
		}
		t.Fatalf(errStr, code, err, "search scope three parse")
		return
	}
	results := []Body{}
	if err := json.Unmarshal(byteArr, &results); err != nil {
		t.Fatalf(errStr, code, err, "search scope three parse format json")
		return
	}
	for _, v := range results {
		if v.Namespace != repo.Namespace || v.Repository != repo.Repository {
			t.Fatalf(errStr, code, err, "search scope three parse result is wrong")
			return
		}
		for _, str := range query {
			if !strings.Contains(v.URL, str) {
				t.Fatalf(errStr, code, err, "search scope three parse result is wrong")
				return
			}
		}
	}
}

func TestSearchScopeFourParse(t *testing.T) {
	repo := api.AppV1Repo{
		URI:        api.DockyardURI,
		Namespace:  api.UserName,
		Repository: "searchapp",
	}

	query := []string{fileName, "latest", "linux", "arm"}

	byteArr, code, err := repo.SearchScoped(query, api.Token)
	if err != nil || code != 200 {
		if code == 0 {
			t.Logf("System (not dockyard) error :%s", err)
			return
		}
		t.Fatalf(errStr, code, err, "search scope four parse")
		return
	}
	results := []Body{}
	if err := json.Unmarshal(byteArr, &results); err != nil {
		t.Fatalf(errStr, code, err, "search scope four parse format json")
		return
	}
	for _, v := range results {
		if v.Namespace != repo.Namespace || v.Repository != repo.Repository {
			t.Fatalf(errStr, code, err, "search scope four parse result is wrong")
			return
		}
		for _, str := range query {
			if !strings.Contains(v.URL, str) {
				t.Fatalf(errStr, code, err, "search scope four parse result is wrong")
				return
			}
		}
	}
}

func TestSearchScopeMutileParse(t *testing.T) {
	repo := api.AppV1Repo{
		URI:        api.DockyardURI,
		Namespace:  api.UserName,
		Repository: "searchapp",
	}

	query := []string{fileName, "latest", "linux", "arm", "v1"}

	byteArr, code, err := repo.SearchScoped(query, api.Token)
	if err != nil || code != 200 {
		if code == 0 {
			t.Logf("System (not dockyard) error :%s", err)
			return
		}
		t.Fatalf(errStr, code, err, "search scope mutile parse")
		return
	}
	results := []Body{}
	if err := json.Unmarshal(byteArr, &results); err != nil {
		t.Fatalf(errStr, code, err, "search scope mutile parse format json")
		return
	}
	for _, v := range results {
		if v.Namespace != repo.Namespace || v.Repository != repo.Repository {
			t.Fatalf(errStr, code, err, "search scope mutile parse result is wrong")
			return
		}
		for _, str := range query {
			if !strings.Contains(v.URL, str) {
				t.Fatalf(errStr, code, err, "search scope mutile parse result is wrong")
				return
			}
		}
	}
}

func TestSearchScopeNoExistAPP(t *testing.T) {
	repo := api.AppV1Repo{
		URI:        api.DockyardURI,
		Namespace:  api.UserName,
		Repository: "searchapp",
	}

	query := []string{fileName, "latest", "linux", "arm", "avppv1rvoovtsevarchapvplinuvxarm"}

	byteArr, code, err := repo.SearchScoped(query, api.Token)
	if err != nil || code != 200 {
		if code == 0 {
			t.Logf("System (not dockyard) error :%s", err)
			return
		}
		t.Fatalf(errStr, code, err, "search scope not exist")
		return
	}
	results := []Body{}
	if err := json.Unmarshal(byteArr, &results); err != nil {
		t.Fatalf(errStr, code, err, "search scope not exist format json")
		return
	}
	for _, v := range results {
		has := true
		if v.Namespace != repo.Namespace || v.Repository != repo.Repository {
			has = false
		}
		if !has {
			break
		}
		for _, str := range query {
			if !strings.Contains(v.URL, str) {
				has = false
			}
		}
		if has {
			t.Fatalf(errStr, code, err, "search scope not exist result is wrong")
			return
		}
	}
}

func TestSearchScopeEmptyQuery(t *testing.T) {
	repo := api.AppV1Repo{
		URI:        api.DockyardURI,
		Namespace:  api.UserName,
		Repository: "searchapp",
	}
	query := []string{" "}

	byteArr, code, err := repo.SearchScoped(query, api.Token)
	if code == 0 {
		t.Logf("System (not dockyard) error :%s", err)
		return
	}
	if err != nil {
		t.Fatalf(errStr, code, err, "search scope empty parse")
		return
	}
	if strings.Index(string(byteArr), "errors") == -1 {
		t.Fatalf(errStr, code, err, "search scope empty parse")
		return
	}
}

func TestSearchScopeEmptyRepository(t *testing.T) {
	repo := api.AppV1Repo{
		URI:        api.DockyardURI,
		Namespace:  api.UserName,
		Repository: " ",
	}
	query := []string{fileName}

	_, code, err := repo.SearchScoped(query, api.Token)
	if err == nil && code == 200 {
		t.Fatalf(errStr, code, err, "search scope empty repository")
		return
	}
}

func TestSearchScopeEmptyNamespace(t *testing.T) {
	repo := api.AppV1Repo{
		URI:        api.DockyardURI,
		Namespace:  " ",
		Repository: "searchapp",
	}
	query := []string{fileName}

	_, code, err := repo.SearchScoped(query, api.Token)
	if err == nil && code == 200 {
		t.Fatalf(errStr, code, err, "search scope empty namespace")
		return
	}
}

func TestSearchScopeIllegalQuery(t *testing.T) {
	repo := api.AppV1Repo{
		URI:        api.DockyardURI,
		Namespace:  api.UserName,
		Repository: "searchapp",
	}
	query := []string{"abc*$%#"}

	_, code, err := repo.SearchScoped(query, api.Token)
	if err == nil && code == 200 {
		t.Fatalf(errStr, code, err, "search scope illegal qiery")
		return
	}
}

func TestSearchScopeIllegalRepository(t *testing.T) {
	repo := api.AppV1Repo{
		URI:        api.DockyardURI,
		Namespace:  api.UserName,
		Repository: "abc*&^%$",
	}
	query := []string{fileName}

	_, code, err := repo.SearchScoped(query, api.Token)
	if err == nil && code == 200 {
		t.Fatalf(errStr, code, err, "search scope illegal repository")
		return
	}
}

func TestSearchScopeIllegalNamespace(t *testing.T) {
	repo := api.AppV1Repo{
		URI:        api.DockyardURI,
		Namespace:  "bac*&^%$",
		Repository: "searchapp",
	}
	query := []string{fileName}

	_, code, err := repo.SearchScoped(query, api.Token)
	if err == nil && code == 200 {
		t.Fatalf(errStr, code, err, "search scope illegal namespace")
		return
	}
}

func TestSearchScopeQueryLenMoreThan255(t *testing.T) {
	repo := api.AppV1Repo{
		URI:        api.DockyardURI,
		Namespace:  api.UserName,
		Repository: "searchapp",
	}
	query := []string{"abcdefghij1234567890" +
		"abcdefghij1234567890" + "abcdefghij1234567890" + "abcdefghij1234567890" +
		"abcdefghij1234567890" + "abcdefghij1234567890" + "abcdefghij1234567890" +
		"abcdefghij1234567890" + "abcdefghij1234567890" + "abcdefghij1234567890" +
		"abcdefghij1234567890" + "abcdefghij1234567890" + "abcdefghij123456"}

	byteArr, code, err := repo.SearchScoped(query, api.Token)
	if code == 0 {
		t.Logf("System (not dockyard) error :%s", err)
		return
	}
	if err != nil {
		t.Fatalf(errStr, code, err, "search scope query len more than 255")
		return
	}
	if strings.Index(string(byteArr), "errors") == -1 {
		t.Fatalf(errStr, code, err, "search scope empty parse")
		return
	}
}

func TestSearchScopeRepositoryLenMoreThan255(t *testing.T) {
	repo := api.AppV1Repo{
		URI: api.DockyardURI,
		Namespace: "abcdefghij1234567890" +
			"abcdefghij1234567890" + "abcdefghij1234567890" + "abcdefghij1234567890" +
			"abcdefghij1234567890" + "abcdefghij1234567890" + "abcdefghij1234567890" +
			"abcdefghij1234567890" + "abcdefghij1234567890" + "abcdefghij1234567890" +
			"abcdefghij1234567890" + "abcdefghij1234567890" + "abcdefghij123456",
		Repository: "searchapp",
	}
	query := []string{fileName}

	_, code, err := repo.SearchScoped(query, api.Token)
	if err == nil && code == 200 {
		t.Fatalf(errStr, code, err, "search scope repository len more than 255")
		return
	}
}

func TestSearchScopeChineseQuery(t *testing.T) {
	repo := api.AppV1Repo{
		URI:        api.DockyardURI,
		Namespace:  api.UserName,
		Repository: "searchapp",
	}

	query := []string{"中文输入"}

	byteArr, code, err := repo.SearchScoped(query, api.Token)
	if code == 0 {
		t.Logf("System (not dockyard) error :%s", err)
		return
	}
	if err != nil {
		t.Fatalf(errStr, code, err, "search scope chinese parse")
		return
	}
	if strings.Index(string(byteArr), "errors") == -1 {
		t.Fatalf(errStr, code, err, "search scope chinese parse")
		return
	}
}

func TestSearchScopeChineseRepository(t *testing.T) {
	repo := api.AppV1Repo{
		URI:        api.DockyardURI,
		Namespace:  api.UserName,
		Repository: "中文输入",
	}

	query := []string{fileName}

	_, code, err := repo.SearchScoped(query, api.Token)
	if err == nil && code == 200 {
		t.Fatalf(errStr, code, err, "search scope chinese repository")
		return
	}
}

func TestSearchScopeChineseNamespace(t *testing.T) {
	repo := api.AppV1Repo{
		URI:        api.DockyardURI,
		Namespace:  "中文输入",
		Repository: "searchapp",
	}

	query := []string{fileName}

	_, code, err := repo.SearchScoped(query, api.Token)
	if err == nil && code == 200 {
		t.Fatalf(errStr, code, err, "search scope chinese namespace")
		return
	}
}
