package appv1

import (
	"fmt"
	"os"

	"github.com/containerops/dockyard/tests/api"
	"github.com/containerops/dockyard/utils"
)

var (
	testDir         = "testdata"
	fileName        = "appv1-testfile"
	manifestName    = fileName + ".manifest"
	errStr          = "HTTPCODE: %d, ERROR: (%v) when %s"
	maxSizeFile     = "maxsizefile"
	maxSizeManifest = maxSizeFile + ".manifest"
	pullFile        = "pulldata"
)

func init() {
	err := api.SetConfig("../testsuite.conf")
	if err != nil {
		panic("fail to load config file")
	}

	if !api.DockyardHealth() {
		panic("should setup dockyard first")
	}

	if err := api.CheckDockyardAuth(); err != nil {
		panic(err)
	}

	if !utils.IsDirExist(testDir) {
		if err := os.Mkdir(testDir, 0666); err != nil {
			panic(fmt.Sprintf("Prepared testdata dir error:%v", err))
		}
	}

	if err := createFile(testDir+"/"+fileName, 0); err != nil {
		panic(fmt.Sprintf("Prepared testdata file error:%v", err))
	}
	if err := createFile(testDir+"/"+manifestName, 0); err != nil {
		panic(fmt.Sprintf("Prepared testdata file error:%v", err))
	}
	if err := createFile(testDir+"/"+maxSizeFile, 500); err != nil {
		panic(fmt.Sprintf("Prepared testdata file error:%v", err))
	}
	if err := createFile(testDir+"/"+maxSizeManifest, 0); err != nil {
		panic(fmt.Sprintf("Prepared testdata file error:%v", err))
	}

	if api.AuthEnable == true {
		if token, err := api.GetAuthorize(api.AuthKey, api.AuthValue); err != nil {
			panic("get token error")
		} else {
			api.Token = "Bearer " + token
		}
	}
}

func createFile(filePath string, size int) error {
	if !utils.IsFileExist(filePath) {
		f, err := os.Create(filePath)
		if err != nil {
			return err
		}
		defer f.Close()
		f.WriteString(fmt.Sprintf("this is appv1 test file name: %v", filePath))
		if size > 0 {
			byteArr := make([]byte, 1024)
			count := size * 1000
			for i := 0; i < count; i++ {
				f.Write(byteArr)
			}
		}
	}
	return nil
}

func PushApp(repo api.AppV1Repo, app api.AppV1App, token string, file, manifests string) (map[string]string, int, error) {
	uuid, code, err := repo.Post(token)
	if err != nil || code != 202 {
		return nil, code, err
	}
	fileSha512, code, err := repo.PutFile(app, token, uuid, file)
	if err != nil || code != 201 {
		repo.Patch(app, token, uuid, "error")
		return nil, code, fmt.Errorf("put file err: %v, code: %d", err, code)
	}

	manifestsSha512, code, err := repo.PutManifest(app, token, uuid, manifests)
	if err != nil || code != 201 {
		repo.Patch(app, token, uuid, "error")
		return nil, code, fmt.Errorf("put manifest name: %s.manifest, err: %v", fileName, err)
	}
	code, err = repo.Patch(app, token, uuid, "done")
	return map[string]string{"file": fileSha512, "manifests": manifestsSha512}, code, err
}

func PullApp(repo api.AppV1Repo, app api.AppV1App, token string, file string) (map[string]string, int, error) {
	fileSha512, code, err := repo.PullFile(app, token, file)
	if err != nil || code != 200 {
		return nil, code, err
	}
	manifestsSha512, code, err := repo.PullManifest(app, token, file)
	return map[string]string{"file": fileSha512, "manifests": manifestsSha512}, code, err
}

func diffPullPushSha512(pull, push map[string]string) bool {
	if pull == nil || push == nil {
		return true
	}
	for k, v := range pull {
		if val, ok := push[k]; !ok || v != val {
			return true
		}
	}
	return false
}
