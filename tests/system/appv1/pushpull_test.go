package appv1

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"
)

func TestPushPullSingleAPPWithoutTag(t *testing.T) {
	namespace := "user"
	repository := "webapp"
	operation := "linux"
	arch := "arm"
	file := "webapp-v1-linux-arm.tar.gz"
	manifest := "webapp-v1-linux-arm.txt"
	contents := randomContents(1024 * 1024)
	if err := ioutil.WriteFile("tmp", contents, 0666); err != nil {
		t.Fatal(err)
	}
	if err := TarGz("tmp", file); err != nil {
		t.Fatal(err)
	}
	fd, err := os.Create(manifest)
	defer fd.Close()
	if err != nil {
		t.Fatal(err)
	}
	_, err = fd.WriteString("webapp-v1-linux-arm.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer removeFile(file, manifest, "_tmp/tmp", "tmp")
	if err := pushAPP(DockyardURL, namespace, repository, operation, arch, file, manifest); err != nil {
		t.Fatal(err)
	}
	filePath := "./webapp-v1-linux-arm.tar.gz"
	manifestPath := "./webapp-v1-linux-arm.txt"
	defer removeFile(filePath, manifestPath)
	if err := pullAPP(DockyardURL, namespace, repository, operation, arch, file, filePath, manifestPath); err != nil {
		t.Fatal(err)
	}
	if err := UnTarGz(filePath, "_tmp"); err != nil {
		t.Fatal(err)
	}
	buf, err := ioutil.ReadFile("_tmp/tmp")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(buf, contents) {
		t.Fatal("not equal")
	}

}
func TestPushPullMutileAPPWithoutTag(t *testing.T) {
	namespace := "user"
	repository := "webapp"
	operation := "linux"
	arch := "arm"
	file := "webapp-v1-linux-arm.tar.gz"
	manifest := "webapp-v1-linux-arm.txt"
	file1 := "webapp-v2-linux-arm.tar.gz"
	manifest1 := "webapp-v2-linux-arm.txt"
	contents := randomContents(1024 * 1024)
	if err := ioutil.WriteFile("tmp", contents, 0666); err != nil {
		t.Fatal(err)
	}
	if err := TarGz("tmp", file); err != nil {
		t.Fatal(err)
	}
	fd, err := os.Create(manifest)
	defer fd.Close()
	if err != nil {
		t.Fatal(err)
	}
	_, err = fd.WriteString("webapp-v1-linux-arm.txt")
	if err != nil {
		t.Fatal(err)
	}
	contents1 := randomContents(1024 * 1024)
	if err := ioutil.WriteFile("tmp1", contents1, 0666); err != nil {
		t.Fatal(err)
	}
	if err := TarGz("tmp1", file1); err != nil {
		t.Fatal(err)
	}
	fd, err = os.Create(manifest1)
	defer fd.Close()
	if err != nil {
		t.Fatal(err)
	}
	_, err = fd.WriteString("webapp-v2-linux-arm.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer removeFile(file, file1, manifest, manifest1)
	if err := pushAPP(DockyardURL, namespace, repository, operation, arch, file, manifest); err != nil {
		t.Fatal(err)
	}
	if err := pushAPP(DockyardURL, namespace, repository, operation, arch, file1, manifest1); err != nil {
		t.Fatal(err)
	}
	filePath := "./webapp-v1-linux-arm.tar.gz"
	manifestPath := "./webapp-v1-linux-arm.txt"
	defer removeFile(filePath, manifestPath, "_tmp/tmp", "_tmp1/tmp", "_tmp", "_tmp1")
	if err := pullAPP(DockyardURL, namespace, repository, operation, arch, file, filePath, manifestPath); err != nil {
		t.Fatal(err)
	}
	if err := UnTarGz(filePath, "_tmp"); err != nil {
		t.Fatal(err)
	}
	buf, err := ioutil.ReadFile("_tmp/tmp")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(buf, contents) {
		t.Fatal("not equal")
	}
	filePath1 := "./webapp-v2-linux-arm.tar.gz"
	manifestPath1 := "./webapp-v2-linux-arm.txt"
	defer removeFile(filePath1, manifest1)

	if err := pullAPP(DockyardURL, namespace, repository, operation, arch, file1, filePath1, manifestPath1); err != nil {
		t.Fatal(err)
	}
	if err := UnTarGz(filePath, "_tmp1"); err != nil {
		t.Fatal(err)
	}
	buf, err = ioutil.ReadFile("_tmp1/tmp")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(buf, contents) {
		t.Fatal("not equal")
	}
}
func TestPushPullSingleAPPWithTag(t *testing.T) {

	namespace := "user"
	repository := "webapp"
	operation := "linux"
	arch := "arm"
	tagLast := "latest"
	file := "webapp-v1-linux-arm.tar.gz"
	manifest := "webapp-v1-linux-arm.txt"
	contents := randomContents(1024 * 1024)
	if err := ioutil.WriteFile("tmp", contents, 0666); err != nil {
		t.Fatal(err)
	}
	if err := TarGz("tmp", file); err != nil {
		t.Fatal(err)
	}
	fd, err := os.Create(manifest)
	defer fd.Close()
	if err != nil {
		t.Fatal(err)
	}
	_, err = fd.WriteString("webapp-v1-linux-arm.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer removeFile(file, manifest, "_tmp/tmp", "_tmp", "tmp")
	if err := pushAPPWithTag(DockyardURL, namespace, repository, operation, arch, file, manifest, tagLast); err != nil {
		t.Fatal(err)
	}
	filePath := "./webapp-v1-linux-arm.tar.gz"
	manifestPath := "./webapp-v1-linux-arm.txt"
	defer removeFile(filePath, manifestPath)
	if err := pullAPPWithTag(DockyardURL, namespace, repository, operation, arch, file, filePath, manifestPath, tagLast); err != nil {
		t.Fatal(err)
	}
	if err := UnTarGz(filePath, "_tmp"); err != nil {
		t.Fatal(err)
	}
	buf, err := ioutil.ReadFile("_tmp/tmp")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(buf, contents) {
		t.Fatal("not equal")
	}
}
func TestPushPullMutileAPPWithTag(t *testing.T) {
	namespace := "user"
	repository := "webapp"
	operation := "linux"
	arch := "arm"
	tagLast := "latest"
	file := "webapp-v1-linux-arm.tar.gz"
	manifest := "webapp-v1-linux-arm.txt"
	file1 := "webapp-v2-linux-arm.tar.gz"
	manifest1 := "webapp-v2-linux-arm.txt"
	contents := randomContents(1024 * 1024)
	if err := ioutil.WriteFile("tmp", contents, 0666); err != nil {
		t.Fatal(err)
	}
	if err := TarGz("tmp", file); err != nil {
		t.Fatal(err)
	}
	fd, err := os.Create(manifest)
	defer fd.Close()
	if err != nil {
		t.Fatal(err)
	}
	_, err = fd.WriteString("webapp-v1-linux-arm.txt")
	if err != nil {
		t.Fatal(err)
	}
	contents1 := randomContents(1024 * 1024)
	if err := ioutil.WriteFile("tmp1", contents1, 0666); err != nil {
		t.Fatal(err)
	}
	if err := TarGz("tmp1", file1); err != nil {
		t.Fatal(err)
	}
	fd, err = os.Create(manifest1)
	defer fd.Close()
	if err != nil {
		t.Fatal(err)
	}
	_, err = fd.WriteString("webapp-v2-linux-arm.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer removeFile(file, file1, manifest, manifest1, "_tmp/tmp", "_tmp1/tmp", "_tmp", "_tmp1")
	if err := pushAPPWithTag(DockyardURL, namespace, repository, operation, arch, file, manifest, tagLast); err != nil {
		t.Fatal(err)
	}
	if err := pushAPPWithTag(DockyardURL, namespace, repository, operation, arch, file1, manifest1, tagLast); err != nil {
		t.Fatal(err)
	}
	filePath := "./webapp-v1-linux-arm.tar.gz"
	manifestPath := "./webapp-v1-linux-arm.txt"
	defer removeFile(filePath, manifestPath)
	if err := pullAPPWithTag(DockyardURL, namespace, repository, operation, arch, file, filePath, manifestPath, tagLast); err != nil {
		t.Fatal(err)
	}
	if err := UnTarGz(filePath, "_tmp"); err != nil {
		t.Fatal(err)
	}
	buf, err := ioutil.ReadFile("_tmp/tmp")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(buf, contents) {
		t.Fatal("not equal")
	}
	filePath1 := "./webapp-v2-linux-arm.tar.gz"
	manifestPath1 := "./webapp-v2-linux-arm.txt"
	defer removeFile(filePath1, manifestPath1)
	if err := pullAPPWithTag(DockyardURL, namespace, repository, operation, arch, file1, filePath1, manifestPath1, tagLast); err != nil {
		t.Fatal(err)
	}
	if err := UnTarGz(filePath, "_tmp1"); err != nil {
		t.Fatal(err)
	}
	buf, err = ioutil.ReadFile("_tmp1/tmp")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(buf, contents) {
		t.Fatal("not equal")
	}
}
func TestPushPullDifferentAPPWithDifferentTag(t *testing.T) {
	namespace := "user"
	repository := "webapp"
	operation := "linux"
	arch := "arm"
	file := "webapp-v1-linux-arm.tar.gz"
	manifest := "webapp-v1-linux-arm.txt"
	file1 := "webapp-v2-linux-arm.tar.gz"
	manifest1 := "webapp-v2-linux-arm.txt"
	tagGroups := []string{"latest", "1.0", "2.0"}

	contents := randomContents(1024 * 1024)
	if err := ioutil.WriteFile("tmp", contents, 0666); err != nil {
		t.Fatal(err)
	}
	if err := TarGz("tmp", file); err != nil {
		t.Fatal(err)
	}
	fd, err := os.Create(manifest)
	defer fd.Close()
	if err != nil {
		t.Fatal(err)
	}
	_, err = fd.WriteString("webapp-v1-linux-arm.txt")
	if err != nil {
		t.Fatal(err)
	}

	for _, tagLast := range tagGroups {
		defer removeFile(file, manifest)
		if err := pushAPPWithTag(DockyardURL, namespace, repository, operation, arch, file, manifest, tagLast); err != nil {
			t.Fatal(err)
		}
	}
	for _, tagLast := range tagGroups {
		filePath := "./" + tagLast + "webapp-v1-linux-arm.tar.gz"
		manifestPath := "./" + tagLast + "webapp-v1-linux-arm.txt"
		defer removeFile(filePath, manifestPath, "_tmp/tmp", "_tmp")
		if err := pullAPPWithTag(DockyardURL, namespace, repository, operation, arch, file, filePath, manifestPath, tagLast); err != nil {
			t.Fatal(err)
			if err := UnTarGz(filePath, "_tmp"); err != nil {
				t.Fatal(err)
			}
			buf, err := ioutil.ReadFile("_tmp/tmp")
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(buf, contents) {
				t.Fatal("not equal")
			}

		}
	}
	contents1 := randomContents(1024 * 1024)
	if err := ioutil.WriteFile("tmp1", contents1, 0666); err != nil {
		t.Fatal(err)
	}
	if err := TarGz("tmp1", file1); err != nil {
		t.Fatal(err)
	}
	fd, err = os.Create(manifest1)
	defer fd.Close()
	if err != nil {
		t.Fatal(err)
	}
	_, err = fd.WriteString("webapp-v2-linux-arm.txt")
	if err != nil {
		t.Fatal(err)
	}
	for _, tagLast := range tagGroups {
		defer removeFile(file, file1, manifest, manifest1, "tmp", "tmp1")
		if err := pushAPPWithTag(DockyardURL, namespace, repository, operation, arch, file1, manifest1, tagLast); err != nil {
			t.Fatal(err)
		}
	}
	for _, tagLast := range tagGroups {
		filePath1 := "./" + tagLast + "webapp-v2-linux-arm.tar.gz"
		manifestPath1 := "./" + tagLast + "webapp-v2-linux-arm.txt"
		defer removeFile(filePath1, manifestPath1, "_tmp1/tmp", "_tmp1")
		if err := pullAPPWithTag(DockyardURL, namespace, repository, operation, arch, file1, filePath1, manifestPath1, tagLast); err != nil {
			t.Fatal(err)
			if err := UnTarGz(filePath1, "_tmp"); err != nil {
				t.Fatal(err)
			}
			buf, err := ioutil.ReadFile("_tmp/tmp")
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(buf, contents) {
				t.Fatal("not equal")
			}
		}
	}

}
