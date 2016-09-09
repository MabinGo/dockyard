package appv1

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestSearchGlobal(t *testing.T) {
	PushAPPInit()
	query := "webapp+v0"
	msg, err := searchGlobal(DockyardURLSearch, query)
	if err != nil {
		t.Fatal(err)
	}
	results := []Body{}
	if err := json.Unmarshal(msg, &results); err != nil {
		t.Fatal(err)
	}
	for _, v := range results {
		if strings.Contains(v.URL, "webapp") && strings.Contains(v.URL, "v0") {
		} else {
			t.Fatalf("The searchGlobal fuction is failed! : [error]%v\n", v)
		}
	}
	t.Log("It is testing success to the searchGlobal fuction.")
}

func TestSearchScope(t *testing.T) {
	namespace := "test"
	repository := "searchapp"
	query := "webapp"

	msg, err := searchScope(DockyardURL, namespace, repository, query)
	if err != nil {
		t.Fatal(err)
	}
	results := []Body{}
	if err := json.Unmarshal(msg, &results); err != nil {
		t.Fatal(err)
	}
	for _, v := range results {
		if v.Namespace == namespace && v.Repository == repository && strings.Contains(v.URL, "webapp") {
		} else {
			t.Fatalf("The searchScope fuction is failed! : [error]%v\n", v)
		}
	}
	t.Log("It is testing success to the searchScope fuction.")
}

func TestSearchScopeTwoParse(t *testing.T) {
	namespace := "test"
	repository := "searchapp"
	query := "webapp+v0"

	msg, err := searchScope(DockyardURL, namespace, repository, query)
	if err != nil {
		t.Fatal(err)
	}
	results := []Body{}
	if err := json.Unmarshal(msg, &results); err != nil {
		t.Fatal(err)
	}
	for _, v := range results {
		if v.Namespace == namespace && v.Repository == repository && strings.Contains(v.URL, "webapp") &&
			strings.Contains(v.URL, "v0") {
		} else {
			t.Fatalf("The searchScopeTwoParse fuction is failed! : [error]%v\n", v)
		}
	}
	t.Log("It is testing success to the searchScopeTwoParse fuction.")
}

func TestSearchScopeThreeParse(t *testing.T) {
	namespace := "test"
	repository := "searchapp"
	query := "webapp+v0+linux"

	msg, err := searchScope(DockyardURL, namespace, repository, query)
	if err != nil {
		t.Fatal(err)
	}
	results := []Body{}
	if err := json.Unmarshal(msg, &results); err != nil {
		t.Fatal(err)
	}
	for _, v := range results {
		if v.Namespace == namespace && v.Repository == repository && strings.Contains(v.URL, "webapp") &&
			strings.Contains(v.URL, "v0") && strings.Contains(v.URL, "linux") {
		} else {
			t.Fatalf("The searchScopeThreeParse fuction is failed! : [error]%v\n", v)
		}
	}
	t.Log("It is testing success to the searchScopeThreeParse fuction.")
}

func TestSearchScopeFourParse(t *testing.T) {
	namespace := "test"
	repository := "searchapp"
	query := "webapp+v0+linux+latest"

	msg, err := searchScope(DockyardURL, namespace, repository, query)
	if err != nil {
		t.Fatal(err)
	}
	results := []Body{}
	if err := json.Unmarshal(msg, &results); err != nil {
		t.Fatal(err)
	}
	for _, v := range results {
		if v.Namespace == namespace && v.Repository == repository && strings.Contains(v.URL, "webapp") &&
			strings.Contains(v.URL, "v0") && strings.Contains(v.URL, "linux") && strings.Contains(v.URL, "latest") {
		} else {
			t.Fatalf("The searchScopeFourParse fuction is failed! : [error]%v\n", v)
		}
	}
	t.Log("It is testing success to the searchScopeFourParse fuction.")
}

func TestSearchScopeMutileParse(t *testing.T) {
	namespace := "test"
	repository := "searchapp"
	query := "webapp+v0+linux+latest+amd"

	msg, err := searchScope(DockyardURL, namespace, repository, query)
	if err != nil {
		t.Fatal(err)
	}
	results := []Body{}
	if err := json.Unmarshal(msg, &results); err != nil {
		t.Fatal(err)
	}
	for _, v := range results {
		if v.Namespace == namespace && v.Repository == repository && strings.Contains(v.URL, "webapp") &&
			strings.Contains(v.URL, "v0") && strings.Contains(v.URL, "linux") &&
			strings.Contains(v.URL, "latest") && !strings.Contains(v.URL, "amd") {
		} else {
			t.Fatalf("The searchScopeMutileParse fuction is failed! : [error]%v\n", v)
		}
	}
	t.Log("It is testing success to the searchScopeMutileParse fuction.")
}
func TestSearchScopeRandomParse(t *testing.T) {
	namespace := "test"
	repository := "searchapp"
	query := "linux+v0+webapp"

	msg, err := searchScope(DockyardURL, namespace, repository, query)
	if err != nil {
		t.Fatal(err)
	}
	results := []Body{}
	if err := json.Unmarshal(msg, &results); err != nil {
		t.Fatal(err)
	}
	for _, v := range results {
		if v.Namespace == namespace && v.Repository == repository && strings.Contains(v.URL, "webapp") &&
			strings.Contains(v.URL, "v0") && strings.Contains(v.URL, "linux") {
		} else {
			t.Fatalf("The searchScopeRandomParse fuction is failed! : [error]%v\n", v)
		}
	}
	t.Log("It is testing success to the searchScopeRandomParse fuction.")
}

func TestSearchScopeNoExistAPP(t *testing.T) {
	namespace := "test"
	repository := "searchapp"
	query := "webapp+v1"
	defer DeleteInitPushAPP()
	msg, err := searchScope(DockyardURL, namespace, repository, query)
	if err != nil {
		t.Fatal(err)
	}
	results := []Body{}
	if err := json.Unmarshal(msg, &results); err != nil {
		t.Fatal(err)
	}
	for _, v := range results {
		if v.Namespace == namespace && v.Repository == repository && strings.Contains(v.URL, "webapp") &&
			strings.Contains(v.URL, "v1") {
			t.Fatalf("The searchScopeNoExistAPP fuction is failed! : [error]%v\n", v)
		}
	}
	t.Log("It is testing success to the searchScopeNoExistAPP fuction.")
}
