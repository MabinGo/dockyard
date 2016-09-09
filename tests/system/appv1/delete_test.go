package appv1

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func TestDeleteAPPWithoutTag(t *testing.T) {
	namespace := "delete"
	repository := "delapp"
	operation := "linux"
	arch := "arm"
	file := "delapp-v1-linux-arm.tar.gz"
	manifest := "delapp-v1-linux-arm.txt"

	TarGz("init_test.go", file)
	fd, err := os.Create(manifest)
	defer fd.Close()
	if err != nil {
		t.Fatal(err)
	}

	_, err = fd.WriteString("this is a test to check a deletable fuction")
	if err != nil {
		t.Fatal(err)
	}

	defer removeFile(file, manifest)
	if err := pushAPP(DockyardURL, namespace, repository, operation, arch, file, manifest); err != nil {
		t.Fatal(err)
	}

	if err := deleteAPP(DockyardURL, namespace, repository, operation, arch, file); err != nil {
		t.Fatal(err)
	}

	filePath := "./delapp-v1-tmp.tar.gz"
	manifestPath := "./delapp-v1-tmp.txt"
	defer removeFile(filePath, manifestPath)
	if err := pullAPP(DockyardURL, namespace, repository, operation, arch, file, filePath, manifestPath); err != nil {
		t.Fatal(err)
	}
	buf, err := ioutil.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(buf), "Not found app") {
		t.Fatalf("The delete app fuction is falied!")
	}

}
func TestDeleteAPPWithTag(t *testing.T) {
	namespace := "delete"
	repository := "delapp"
	operation := "linux"
	arch := "arm"
	tagLast := "latest"
	file := "delapp-v1-linux-arm.tar.gz"
	manifest := "delapp-v1-linux-arm.txt"

	TarGz("init_test.go", file)
	fd, err := os.Create(manifest)
	defer fd.Close()
	if err != nil {
		t.Fatal(err)
	}

	_, err = fd.WriteString("this is a test to check a deletable fuction")
	if err != nil {
		t.Fatal(err)
	}
	defer removeFile(file, manifest)

	if err := pushAPPWithTag(DockyardURL, namespace, repository, operation, arch, file, manifest, tagLast); err != nil {
		t.Fatal(err)
	}

	if err := deleteAPPWithTag(DockyardURL, namespace, repository, operation, arch, file, tagLast); err != nil {
		t.Fatal(err)
	}

	filePath := "./delapp-v1-tmp.tar.gz"
	manifestPath := "./delapp-v1-tmp.txt"
	defer removeFile(filePath, manifestPath)
	if err := pullAPPWithTag(DockyardURL, namespace, repository, operation, arch, file, filePath, manifestPath, tagLast); err != nil {
		t.Fatal(err)
	}

	buf, err := ioutil.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(buf), "Not found app") {
		t.Fatalf("The delete app fuction is falied!")
	}
}
func TestDeleteMutileAPPWithoutTag(t *testing.T) {
	PushAPPInit()
	namespaceGroups := []string{"test", "user", "root", "delete"}
	repositoryGroups := []string{"webapp", "testapp", "searchapp", "delapp"}
	for _, namespace := range namespaceGroups {
		for _, repository := range repositoryGroups {
			operation := "linux"
			arch := "arm"
			file := "webapp-v0-linux-arm.tar.gz"
			if err := deleteAPP(DockyardURL, namespace, repository, operation, arch, file); err != nil {
				fmt.Errorf("%v\n", err.Error())
			}
			filePath := "./delapp-v0-tmp.tar.gz"
			manifestPath := "./delapp-v0-tmp.txt"
			defer removeFile(file, filePath, manifestPath)
			if err := pullAPP(DockyardURL, namespace, repository, operation, arch, file, filePath, manifestPath); err != nil {
				t.Fatal(err)
			}

			buf, err := ioutil.ReadFile(manifestPath)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(buf), "Not found app") {
				t.Fatalf("The delete app fuction is falied!")
			}
		}
	}

}
func TestDeleteMutileAPPWithTag(t *testing.T) {
	PushAPPInit()
	namespaceGroups := []string{"test", "user", "root", "delete"}
	repositoryGroups := []string{"webapp", "testapp", "searchapp", "delapp"}
	for _, namespace := range namespaceGroups {
		for _, repository := range repositoryGroups {
			operation := "linux"
			arch := "arm"
			tagLast := "latest"
			file := "webapp-v0-linux-arm.tar.gz"
			if err := deleteAPPWithTag(DockyardURL, namespace, repository, operation, arch, file, tagLast); err != nil {
				fmt.Errorf("%v\n", err.Error())
			}
			filePath := "./delapp-v0-tmp.tar.gz"
			manifestPath := "./delapp-v0-tmp.txt"
			defer removeFile(file, filePath, manifestPath)
			if err := pullAPPWithTag(DockyardURL, namespace, repository, operation, arch, file, filePath, manifestPath, tagLast); err != nil {
				t.Fatal(err)
			}

			buf, err := ioutil.ReadFile(manifestPath)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(buf), "Not found app") {
				t.Fatalf("The delete app fuction is falied!")
			}
		}
	}

}
