package appv1

import (
	"fmt"
	"strings"
	"testing"
)

func TestSearchNoExisNamespace(t *testing.T) {
	namespace := "noexist"
	repository := "webapp"
	query := "linux"
	msg, err := searchScope(DockyardURL, namespace, repository, query)
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(string(msg), "Not found repository") && strings.Contains(string(msg), "noexist/webapp") {
		t.Log("It is testing success to the fuction.")
	} else {
		t.Fatalf("The  fuction is failed! : [error]%v\n", msg)
	}
}
func TestSearchNoExisRepository(t *testing.T) {
	namespace := "user"
	repository := "noexist"
	query := "linux"
	msg, err := searchScope(DockyardURL, namespace, repository, query)
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(string(msg), "Not found repository") && strings.Contains(string(msg), "user/noexist") {
		t.Log("It is testing success to the  fuction.")
	} else {
		t.Fatalf("The fuction is failed! : [error]%v\n", msg)
	}
}
func TestSearchNoExistQuery(t *testing.T) {
	namespace := "user"
	repository := "webapp"
	query := "noexist"

	msg, err := searchScope(DockyardURL, namespace, repository, query)
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(string(msg), "") {
		t.Log("It is a testing success to the fuction.")
	} else {
		t.Fatalf("The fuction is failed! : [error]%v\n", msg)
	}

}
func TestPushNoExistAPP(t *testing.T) {
	namespace := "user"
	repository := "webapp"
	operation := "linux"
	arch := "arm"
	file := "noexist.tar.gz"
	manifest := "webapp-v1-linux-arm.txt"
	err := pushAPP(DockyardURL, namespace, repository, operation, arch, file, manifest)
	if err == nil {
		t.Fatal(err)
	}
	str := fmt.Sprintf("%v", err)
	if strings.Contains(str, "") {
		t.Log("It is a testing success to the function!")
	} else {
		t.Fatalf("The fuction is failed: [error]%v\n", err)
	}
}

func TestPullNoExistAPP(t *testing.T) {
	namespace := "user"
	repository := "webapp"
	operation := "linux"
	arch := "arm"
	file := "onexist.tar.gz"
	filePath := "./webapp-v1-linux-arm.tar.gz"
	manifestPath := "./webapp-v1-linux-arm.txt"
	err := pullAPP(DockyardURL, namespace, repository, operation, arch, file, filePath, manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	str := fmt.Sprintf("%v", err)
	if strings.Contains(str, "") {
		t.Log("It is a testing success to the function!")
	} else {
		t.Fatalf("The fuction is failed: [error]%v\n", err)
	}
}

func TestDeleteNoExistAPP(t *testing.T) {
	namespace := "user"
	repository := "webapp"
	operation := "linux"
	arch := "arm"
	file := "onexist.tar.gz"
	err := deleteAPP(DockyardURL, namespace, repository, operation, arch, file)
	str := fmt.Sprintf("%v", err)
	if strings.Contains(str, "404") {
		t.Log("It is a testing success to the function!")
	} else {
		t.Fatalf("It is a testing failed: [error]%v\n", err)
	}
}
func TestDeleteNoExistNamespace(t *testing.T) {
	namespace := "onexist"
	repository := "webapp"
	operation := "linux"
	arch := "arm"
	file := "webapp-v1-linux-arm.tar.gz"
	err := deleteAPP(DockyardURL, namespace, repository, operation, arch, file)
	str := fmt.Sprintf("%v", err)
	if strings.Contains(str, "404") {
		t.Log("It is a testing success to the function!")
	} else {
		t.Fatalf("It is a testing failed: [error]%v\n", err)
	}
}

func TestDeleteNoExistRepository(t *testing.T) {
	namespace := "user"
	repository := "noexist"
	operation := "linux"
	arch := "arm"
	file := "webapp-v1-linux-arm.tar.gz"
	err := deleteAPP(DockyardURL, namespace, repository, operation, arch, file)
	str := fmt.Sprintf("%v", err)
	if strings.Contains(str, "404") {
		t.Log("It is a testing success to the function!")
	} else {
		t.Fatalf("It is a testing failed: [error]%v\n", err)
	}
}
