package appv1

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	//"path"
	"testing"
)

func TestPushPullTarWithoutTag(t *testing.T) {
	namespace := "user"
	repository := "webapp"
	operation := "linux"
	arch := "arm"
	file := "webapp-v1-linux-arm.tar.gz"
	manifest := "webapp-v1-linux-arm.txt"
	filePath := "./tmp.tar.gz"
	manifestPath := "./webapp-v1-linux-arm.txt"
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

	_, err = fd.WriteString("this is webapp-v1-linux-arm")
	if err != nil {
		t.Fatal(err)
	}
	defer removeFile(file, filePath, manifest, manifestPath, "tmp", "_tmp")
	if err := pushAPP(DockyardURL, namespace, repository, operation, arch, file, manifest); err != nil {
		t.Fatal(err)
	}

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

func TestPushPullTarWithTag(t *testing.T) {
	namespace := "user"
	repository := "webapp"
	operation := "linux"
	arch := "arm"
	tagLast := "latest"
	file := "webapp-v1-linux-arm.tar.gz"
	manifest := "webapp-v1-linux-arm.txt"
	filePath := "./tmp.tar.gz"
	manifestPath := "./webapp-v1-linux-arm.txt"
	contents := randomContentsfile(1024 * 1024)
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
	_, err = fd.WriteString("this is webapp-v1-linux-arm")
	if err != nil {
		t.Fatal(err)
	}
	defer removeFile(file, filePath, manifest, manifestPath, "tmp", "_tmp/tmp", "_tmp")
	if err := pushAPPWithTag(DockyardURL, namespace, repository, operation, arch, file, manifest, tagLast); err != nil {
		t.Fatal(err)
	}

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
func TestPushPullJarWithoutTag(t *testing.T) {
	namespace := "user"
	repository := "webapp"
	operation := "linux"
	arch := "arm"
	file := "webapp-v1-linux-arm.jar"
	manifest := "webapp-v1-linux-arm.txt"
	filePath := "./example.jar.tmp"
	manifestPath := "./webapp-v1-linux-arm.txt"
	fd, err := os.Create(manifest)
	defer fd.Close()
	if err != nil {
		t.Fatal(err)
	}
	contents := randomContentsfile(1024 * 1024)
	if err := ioutil.WriteFile(file, contents, 0666); err != nil {
		t.Fatal(err)
	}
	_, err = fd.WriteString("this is webapp-v1-linux-arm")
	if err != nil {
		t.Fatal(err)
	}
	defer removeFile(filePath, manifestPath)
	if err := pushAPP(DockyardURL, namespace, repository, operation, arch, file, manifest); err != nil {
		t.Fatal(err)
	}

	if err := pullAPP(DockyardURL, namespace, repository, operation, arch, file, filePath, manifestPath); err != nil {
		t.Fatal(err)
	}
}
func TestPushPullJarWithTag(t *testing.T) {
	namespace := "user"
	repository := "webapp"
	operation := "linux"
	arch := "arm"
	tagLast := "latest"
	file := "webapp-v1-linux-arm.jar"
	manifest := "webapp-v1-linux-arm.txt"
	filePath := "./tmp.jar"
	manifestPath := "./webapp-v1-linux-arm.txt"
	fd, err := os.Create(manifest)
	defer fd.Close()
	if err != nil {
		t.Fatal(err)
	}
	contents := randomContentsfile(1024 * 1024)
	if err := ioutil.WriteFile(file, contents, 0666); err != nil {
		t.Fatal(err)
	}
	_, err = fd.WriteString("this is webapp-v1-linux-arm")
	if err != nil {
		t.Fatal(err)
	}
	defer removeFile(file, filePath, manifest, manifestPath)
	if err := pushAPPWithTag(DockyardURL, namespace, repository, operation, arch, file, manifest, tagLast); err != nil {
		t.Fatal(err)
	}

	if err := pullAPPWithTag(DockyardURL, namespace, repository, operation, arch, file, filePath, manifestPath, tagLast); err != nil {
		t.Fatal(err)
	}
}
func TestPushPullWarWithoutTag(t *testing.T) {
	namespace := "user"
	repository := "webapp"
	operation := "linux"
	arch := "arm"
	file := "webapp-v1-linux-arm.war"
	manifest := "webapp-v1-linux-arm.txt"
	filePath := "./example.war.tmp"
	manifestPath := "./webapp-v1-linux-arm.txt"
	fd, err := os.Create(manifest)
	defer fd.Close()
	if err != nil {
		t.Fatal(err)
	}
	contents := randomContentsfile(1024 * 1024)
	if err := ioutil.WriteFile(file, contents, 0666); err != nil {
		t.Fatal(err)
	}
	_, err = fd.WriteString("this is webapp-v1-linux-arm")
	if err != nil {
		t.Fatal(err)
	}
	defer removeFile(filePath, manifestPath)
	if err := pushAPP(DockyardURL, namespace, repository, operation, arch, file, manifest); err != nil {
		t.Fatal(err)
	}

	if err := pullAPP(DockyardURL, namespace, repository, operation, arch, file, filePath, manifestPath); err != nil {
		t.Fatal(err)
	}
}
func TestPushPullWarWithTag(t *testing.T) {
	namespace := "user"
	repository := "webapp"
	operation := "linux"
	arch := "arm"
	tagLast := "latest"
	file := "webapp-v1-linux-arm.war"
	manifest := "webapp-v1-linux-arm.txt"
	filePath := "./tmp.war"
	manifestPath := "./webapp-v1-linux-arm.txt"
	fd, err := os.Create(manifest)
	defer fd.Close()
	if err != nil {
		t.Fatal(err)
	}
	contents := randomContentsfile(1024 * 1024)
	if err := ioutil.WriteFile(file, contents, 0666); err != nil {
		t.Fatal(err)
	}
	_, err = fd.WriteString("this is webapp-v1-linux-arm")
	if err != nil {
		t.Fatal(err)
	}
	defer removeFile(file, filePath, manifest, manifestPath)
	if err := pushAPPWithTag(DockyardURL, namespace, repository, operation, arch, file, manifest, tagLast); err != nil {
		t.Fatal(err)
	}

	if err := pullAPPWithTag(DockyardURL, namespace, repository, operation, arch, file, filePath, manifestPath, tagLast); err != nil {
		t.Fatal(err)
	}

}
func TestPushPullExeWithoutTag(t *testing.T) {
	namespace := "user"
	repository := "webapp"
	operation := "linux"
	arch := "arm"
	file := "webapp-v1-linux-arm.exe"
	manifest := "webapp-v1-linux-arm.txt"
	filePath := "./example.exe"
	manifestPath := "./webapp-v1-linux-arm.txt"
	fd, err := os.Create(manifest)
	defer fd.Close()
	if err != nil {
		t.Fatal(err)
	}
	contents := randomContentsfile(1024 * 1024)
	if err := ioutil.WriteFile(file, contents, 0666); err != nil {
		t.Fatal(err)
	}
	_, err = fd.WriteString("this is webapp-v1-linux-arm")
	if err != nil {
		t.Fatal(err)
	}
	defer removeFile(filePath, manifestPath)

	if err := pushAPP(DockyardURL, namespace, repository, operation, arch, file, manifest); err != nil {
		t.Fatal(err)
	}

	if err := pullAPP(DockyardURL, namespace, repository, operation, arch, file, filePath, manifestPath); err != nil {
		t.Fatal(err)
	}

}
func TestPushPullExeWithTag(t *testing.T) {
	namespace := "user"
	repository := "webapp"
	operation := "linux"
	arch := "arm"
	tagLast := "latest"
	file := "webapp-v1-linux-arm.exe"
	manifest := "webapp-v1-linux-arm.txt"
	filePath := "./tmp.exe"
	manifestPath := "./webapp-v1-linux-arm.txt"
	fd, err := os.Create(manifest)
	defer fd.Close()
	if err != nil {
		t.Fatal(err)
	}
	contents := randomContentsfile(1024 * 1024)
	if err := ioutil.WriteFile(file, contents, 0666); err != nil {
		t.Fatal(err)
	}
	_, err = fd.WriteString("this is webapp-v1-linux-arm")
	if err != nil {
		t.Fatal(err)
	}
	defer removeFile(file, filePath, manifest, manifestPath)
	if err := pushAPPWithTag(DockyardURL, namespace, repository, operation, arch, file, manifest, tagLast); err != nil {
		t.Fatal(err)
	}

	if err := pullAPPWithTag(DockyardURL, namespace, repository, operation, arch, file, filePath, manifestPath, tagLast); err != nil {
		t.Fatal(err)
	}
}

func TestPushPull500MTarWithTag(t *testing.T) {
	namespace := "user"
	repository := "webapp"
	operation := "linux"
	arch := "arm"
	tagLast := "latest"
	file := "webapp-v1-linux-arm.tar.gz"
	manifest := "webapp-v1-linux-arm.txt"
	filePath := "./tmp.tar.gz"
	manifestPath := "./webapp-v1-linux-arm.txt"
	contents := randomContentsfile(500 * 1024 * 1024)
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

	_, err = fd.WriteString("this is webapp-v1-linux-arm")
	if err != nil {
		t.Fatal(err)
	}
	defer removeFile(file, filePath, manifest, manifestPath, "tmp", "_tmp/tmp", "_tmp")
	if err := pushAPPWithTag(DockyardURL, namespace, repository, operation, arch, file, manifest, tagLast); err != nil {
		t.Fatal(err)
	}

	if err := pullAPPWithTag(DockyardURL, namespace, repository, operation, arch, file, filePath, manifestPath, tagLast); err != nil {
		t.Fatal(err)
	}
	if err := UnTarGz(filePath, "_tmp"); err != nil {
		t.Fatal(err)
	}
}

func TestPushPull100MTarWithTag(t *testing.T) {
	namespace := "user"
	repository := "webapp"
	operation := "linux"
	arch := "arm"
	tagLast := "latest"
	file := "tgzUpload/webapp-v1-linux-arm.tar.gz"
	manifest := "webapp-v1-linux-arm.txt"
	//filePath := "./tmp.tar.gz"
	manifestPath := "./webapp-v1-linux-arm.txt"
	contents := randomContentsfile(100 * 1024 * 1024)
	os.Mkdir("fileUpload", os.ModePerm)
	os.Mkdir("tgzUpload", os.ModePerm)
	os.Mkdir("download", os.ModePerm)
	if err := ioutil.WriteFile("fileUpload/tmp", contents, 0666); err != nil {
		t.Fatal(err)
	}
	if err := TarGz("fileUpload/tmp", file); err != nil {
		t.Fatal(err)
	}
	fd, err := os.Create(manifest)
	defer fd.Close()
	if err != nil {
		t.Fatal(err)
	}

	_, err = fd.WriteString("this is webapp-v1-linux-arm")
	if err != nil {
		t.Fatal(err)
	}
	defer removeFile(manifest, manifestPath, "fileUpload", "tgzUpload", "download", "untar")
	if err := pushAPPWithTag(DockyardURL, namespace, repository, operation, arch, file, manifest, tagLast); err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 20; i++ {
		filePath := fmt.Sprintf("download/downloadtest-%d.tar.gz", i)
		if err := pullAPPWithTag(DockyardURL, namespace, repository, operation, arch, file, filePath, manifestPath, tagLast); err != nil {
			t.Fatal(err)
		}
		if err := UnTarGz(filePath, "untar"); err != nil {
			t.Fatal(err)
		}
	}

}

func randomContentsfile(length int64) []byte {
	return randomBytes[:length]
}
