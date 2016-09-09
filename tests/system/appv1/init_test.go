package appv1

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha512"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path"
	"testing"
	"time"
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

var randomBytes = make([]byte, 512<<20)

func TestInit(t *testing.T) {
	// increase the random bytes to the required maximum
	for i := range randomBytes {
		randomBytes[i] = byte(rand.Intn(2 << 8))
	}
}
func PushAPPInit() {
	namespaceGroups := []string{"test", "user", "root", "delete"}
	repositoryGroups := []string{"webapp", "testapp", "searchapp", "delapp"}
	for _, namespace := range namespaceGroups {
		for _, repository := range repositoryGroups {
			operation := "linux"
			arch := "arm"
			tagLast := "latest"
			file := "webapp-v0-linux-arm.tar.gz"
			manifest := "webapp-v0-linux-arm.txt"
			TarGz("init_test.go", file)
			fd, err := os.Create(manifest)
			defer fd.Close()
			if err != nil {
				fmt.Errorf("%v\n", err.Error())
			}

			_, err = fd.WriteString("this is a test to check a list and search of fuction")
			if err != nil {
				fmt.Errorf("%v\n", err.Error())
			}
			defer removeFile(file, manifest)
			if err := pushAPPWithTag(DockyardURL, namespace, repository, operation, arch, file, manifest, tagLast); err != nil {
				fmt.Errorf("%v\n", err.Error())
			}
		}
	}
}
func DeleteInitPushAPP() {
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
		}
	}
}
func removeFile(args ...string) {
	for _, k := range args {
		os.RemoveAll(k)
	}

}
func pushAPP(DockyardURL, namespace, repository, os, arch, file, manifest string) error {
	rawurl := DockyardURL + namespace + "/" + repository
	resp, err := postAPI(rawurl)
	if err != nil {
		return err
	}
	if resp.StatusCode != 202 {
		return fmt.Errorf("POST request failed: %s", resp.Status)
	}
	uuid := resp.Header.Get("App-Upload-Uuid")
	rawurl = DockyardURL + namespace + "/" + repository + "/" + os + "/" + arch + "/" + file
	resp, err = putAPI(rawurl, uuid, file)
	if err != nil {
		return err
	}
	if resp.StatusCode != 201 {
		return fmt.Errorf("PUT app failed: %s", resp.Status)
	}
	rawurl = DockyardURL + namespace + "/" + repository + "/" + os + "/" + arch + "/" + file + "/manifests"
	resp, err = putAPI(rawurl, uuid, manifest)
	if err != nil {
		return err
	}
	if resp.StatusCode != 201 {
		return fmt.Errorf("PUT manifest failed: %s", resp.Status)
	}
	rawurl = DockyardURL + namespace + "/" + repository + "/" + os + "/" + arch + "/" + file + "/done"
	resp, err = patchAPI(rawurl, uuid)
	if err != nil {
		return err
	}
	if resp.StatusCode != 202 {
		return fmt.Errorf("PATCH request failed: %s", resp.Status)
	}
	return nil
}
func pushAPPWithTag(DockyardURL, namespace, repository, os, arch, filePath, manifest, tag string) error {
	rawurl := DockyardURL + namespace + "/" + repository
	file := path.Base(filePath)
	resp, err := postAPI(rawurl)
	if err != nil {
		return err
	}
	if resp.StatusCode != 202 {
		return fmt.Errorf("POST request failed: %s", resp.Status)
	}
	uuid := resp.Header.Get("App-Upload-Uuid")
	rawurl = DockyardURL + namespace + "/" + repository + "/" + os + "/" + arch + "/" + file + "/" + tag
	resp, err = putAPI(rawurl, uuid, filePath)
	if err != nil {
		return err
	}
	if resp.StatusCode != 201 {
		return fmt.Errorf("PUT app failed: %s", resp.Status)
	}
	rawurl = DockyardURL + namespace + "/" + repository + "/" + os + "/" + arch + "/" + file + "/manifests" + "/" + tag
	resp, err = putAPI(rawurl, uuid, manifest)
	if err != nil {
		return err
	}
	if resp.StatusCode != 201 {
		return fmt.Errorf("PUT manifest failed: %s", resp.Status)
	}
	rawurl = DockyardURL + namespace + "/" + repository + "/" + os + "/" + arch + "/" + file + "/done" + "/" + tag
	resp, err = patchAPI(rawurl, uuid)
	if err != nil {
		return err
	}
	if resp.StatusCode != 202 {
		return fmt.Errorf("PATCH request failed: %s", resp.Status)
	}
	return nil
}

func pullAPP(DockyardURL, namespace, repository, os, arch, file, filePath, manifestPath string) error {
	rawurl := DockyardURL + namespace + "/" + repository + "/" + os + "/" + arch + "/" + file
	if err := pullAPI(rawurl, filePath); err != nil {
		return err
	}
	rawurl = DockyardURL + namespace + "/" + repository + "/" + os + "/" + arch + "/" + file + "/manifests"
	if err := pullAPI(rawurl, manifestPath); err != nil {
		return err
	}
	return nil

}
func pullAPPWithTag(DockyardURL, namespace, repository, os, arch, file, filePath, manifestPath, tag string) error {
	fname := path.Base(file)
	rawurl := DockyardURL + namespace + "/" + repository + "/" + os + "/" + arch + "/" + fname + "/" + tag
	if err := pullAPI(rawurl, filePath); err != nil {
		return err
	}
	rawurl = DockyardURL + namespace + "/" + repository + "/" + os + "/" + arch + "/" + fname + "/manifests" + "/" + tag
	if err := pullAPI(rawurl, manifestPath); err != nil {
		return err
	}
	return nil
}
func searchGlobal(DockyardURLSearch, query string) ([]byte, error) {
	rawurl := DockyardURLSearch + query
	msg, err := searchAPI(rawurl)
	if err != nil {
		return msg, err
	}
	return msg, nil
}

func searchScope(DockyardURL, namespace, repository, query string) ([]byte, error) {
	rawurl := DockyardURL + namespace + "/" + repository + "/search?key=" + query
	msg, err := searchAPI(rawurl)
	if err != nil {
		return msg, err
	}
	return msg, nil
}

func listScope(DockyardURL, namespace, repository string) ([]byte, error) {
	url := DockyardURL + namespace + "/" + repository + "/list"
	msg, err := searchAPI(url)
	if err != nil {
		return msg, err
	}
	return msg, nil
}

func deleteAPP(DockyardURL, namespace, repository, os, arch, file string) error {
	rawurl := DockyardURL + namespace + "/" + repository + "/" + os + "/" + arch + "/" + file
	if err := deleteAPI(rawurl); err != nil {
		return err
	}
	return nil
}

func deleteAPPWithTag(DockyardURL, namespace, repository, os, arch, file, tag string) error {
	rawurl := DockyardURL + namespace + "/" + repository + "/" + os + "/" + arch + "/" + file + "/" + tag
	if err := deleteAPI(rawurl); err != nil {
		return err
	}
	return nil
}

func postAPI(rawurl string) (*http.Response, error) {
	header := make(map[string]string)
	resp, err := sendHttpRequest("POST", rawurl, nil, header)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func putAPI(rawurl, uuid, file string) (*http.Response, error) {
	fd, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer fd.Close()
	sha512h := sha512.New()
	if _, err := io.Copy(sha512h, fd); err != nil {
		return nil, err
	}
	digest := fmt.Sprintf("%s:%x", "sha512", sha512h.Sum(nil))
	header := map[string]string{
		"App-Upload-UUID": uuid,
		"Digest":          digest,
	}
	if _, err := fd.Seek(0, 0); err != nil {
		return nil, err
	}
	resp, err := sendHttpRequest("PUT", rawurl, fd, header)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func patchAPI(rawurl, uuid string) (*http.Response, error) {
	header := map[string]string{
		"App-Upload-UUID": uuid,
	}
	resp, err := sendHttpRequest("PATCH", rawurl, nil, header)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func pullAPI(rawurl, path string) error {
	header := map[string]string{}
	resp, err := sendHttpRequest("GET", rawurl, nil, header)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0640)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(file, resp.Body)
	time.Sleep(100 * time.Millisecond)
	if err != nil {
		return err
	}
	return nil
}

func searchAPI(rawurl string) ([]byte, error) {
	header := make(map[string]string)
	resp, err := sendHttpRequest("GET", rawurl, nil, header)
	if err != nil {
		return []byte{}, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, err
	}
	return body, nil
}

func deleteAPI(rawurl string) error {
	header := map[string]string{}
	resp, err := sendHttpRequest("DELETE", rawurl, nil, header)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("HttpRespose : %v\n", resp.StatusCode)
	}
	return nil
}

func sendHttpRequest(methord, rawurl string, body io.Reader, header map[string]string) (*http.Response, error) {
	url, err := url.Parse(rawurl)
	if err != nil {
		return &http.Response{}, err
	}

	var client *http.Client
	switch url.Scheme {
	case "":
		fallthrough
	case "https":
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client = &http.Client{Transport: tr}
	case "http":
		client = &http.Client{}
	default:
		return &http.Response{}, fmt.Errorf("bad url schema: %v", url.Scheme)
	}

	req, err := http.NewRequest(methord, url.String(), body)
	if err != nil {
		return &http.Response{}, err
	}
	req.URL.RawQuery = req.URL.Query().Encode()
	req.Header.Set("Host", url.Host)
	for k, v := range header {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return &http.Response{}, err
	}

	return resp, nil

}

// Gzip and tar from source directory or file to destination file
// you need check file exist before you call this function
func TarGz(srcDirPath string, destFilePath string) error {
	fw, err := os.Create(destFilePath)
	if err != nil {
		return err
	}
	// Gzip writer
	gw := gzip.NewWriter(fw)
	defer gw.Close()
	// Tar writer
	tw := tar.NewWriter(gw)
	defer tw.Close()
	// Check if it's a file or a directory
	f, err := os.Open(srcDirPath)
	if err != nil {
		return err
	}
	fi, err := f.Stat()
	if err != nil {
		return err
	}
	if fi.IsDir() {
		// handle source directory
		tarGzDir(srcDirPath, path.Base(srcDirPath), tw)
	} else {
		// handle file directly
		tarGzFile(srcDirPath, fi.Name(), tw, fi)
	}
	return nil
}

// Deal with directories
// if find files, handle them with tarGzFile
// Every recurrence append the base path to the recPath
// recPath is the path inside of tar.gz
func tarGzDir(srcDirPath string, recPath string, tw *tar.Writer) error {
	// Open source diretory
	dir, err := os.Open(srcDirPath)
	if err != nil {
		return err
	}
	// Get file info slice
	fis, err := dir.Readdir(0)
	if err != nil {
		return err
	}
	for _, fi := range fis {
		// Append path
		curPath := srcDirPath + "/" + fi.Name()
		// Check it is directory or file
		if fi.IsDir() {
			// Directory
			// (Directory won't add unitl all subfiles are added)
			//fmt.Printf("Adding path...%s\n", curPath)
			tarGzDir(curPath, recPath+"/"+fi.Name(), tw)
		} else {
			// File
			//fmt.Printf("Adding file...%s\n", curPath)
		}
		tarGzFile(curPath, recPath+"/"+fi.Name(), tw, fi)
	}
	return nil
}

// Deal with files
func tarGzFile(srcFile string, recPath string, tw *tar.Writer, fi os.FileInfo) error {
	if fi.IsDir() {
		// Create tar header
		hdr := new(tar.Header)
		// if last charactrer of header name is '/' it also can be directory
		// but if you don't set Typeflag, error will occur when you untargz
		hdr.Name = recPath + "/"
		hdr.Typeflag = tar.TypeDir
		hdr.Size = 0
		// hdr.Mode = 0755 | c_ISDIR
		hdr.Mode = int64(fi.Mode())
		hdr.ModTime = fi.ModTime()
		// Write hander
		err := tw.WriteHeader(hdr)
		if err != nil {
			return err
		}
	} else {
		// File reader
		fr, err := os.Open(srcFile)
		if err != nil {
			return err
		}
		defer fr.Close()
		// Create tar header
		hdr := new(tar.Header)
		hdr.Name = recPath
		hdr.Size = fi.Size()
		hdr.Mode = int64(fi.Mode())
		hdr.ModTime = fi.ModTime()
		// Write hander
		err = tw.WriteHeader(hdr)
		if err != nil {
			return err
		}
		// Write file data
		_, err = io.Copy(tw, fr)
		if err != nil {
			return err
		}
	}
	return nil
}

// Ungzip and untar from source file to destination directory
// you need check file exist before you call this function
func UnTarGz(srcFilePath string, destDirPath string) error {
	//fmt.Println("UnTarGzing " + srcFilePath + "...")
	// Create destination directory
	os.Mkdir(destDirPath, os.ModePerm)
	fr, err := os.Open(srcFilePath)
	if err != nil {
		return err
	}
	defer fr.Close()
	// Gzip reader
	gr, err := gzip.NewReader(fr)
	// Tar reader
	tr := tar.NewReader(gr)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			// End of tar archive
			break
		}
		//handleFatal(err)
		//fmt.Println("UnTarGzing file..." + hdr.Name)
		// Check if it is diretory or file
		if hdr.Typeflag != tar.TypeDir {
			// Get files from archive
			// Create diretory before create file
			os.MkdirAll(destDirPath+"/"+path.Dir(hdr.Name), os.ModePerm)
			// Write data to file
			fw, _ := os.Create(destDirPath + "/" + hdr.Name)
			if err != nil {
				return err
			}
			_, err = io.Copy(fw, tr)
			if err != nil {
				return err
			}
		}
	}
	//fmt.Println("Well done!")
	return nil
}

func randomContents(length int64) []byte {
	return randomBytes[:length]
}
