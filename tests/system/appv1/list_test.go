package appv1

import (
	"encoding/json"
	"testing"
)

func TestListScope(t *testing.T) {
	PushAPPInit()
	namespace := "test"
	repository := "webapp"
	defer DeleteInitPushAPP()
	msg, err := listScope(DockyardURL, namespace, repository)
	if err != nil {
		t.Fatal(err)
	}
	results := []Body{}
	if err := json.Unmarshal(msg, &results); err != nil {
		t.Fatal(err)
	}
	for _, v := range results {
		if v.Namespace == namespace && v.Repository == repository {
		} else {
			t.Fatalf("The ListScope fuction is failed! : [error]%v\n", v)
		}
	}
	t.Log("It is testing success to the ListScope fuction.")
}
