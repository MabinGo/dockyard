package newconcurrent

import (
	//	"bytes"
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"strconv"
	//	"strings"
	"testing"

	"github.com/containerops/dockyard/tests/api"
	"github.com/containerops/dockyard/tests/system/appv1"
)

var Concurfreq int = 10
var Concurdata = make(chan bool, 100)
var size = 128
var randomBytes = make([]byte, size<<20)

var (
	ops        = "linux"
	arch       = "arm"
	tag        = "latest"
	repository = "concurrent"
)

var app = api.AppV1App{
	OS:   ops,
	Arch: arch,
	App:  file,
	Tag:  tag,
}

var repo = api.AppV1Repo{
	URI:        api.DockyardURI,
	Namespace:  api.UserName,
	Repository: repository,
}

func TestConcurInit(t *testing.T) {
	err := api.SetConfig("../testsuite.conf")
	if err != nil {
		panic("fail to load config file")
	}
	// increase the random bytes to the required maximum
	for i := range randomBytes {
		randomBytes[i] = byte(rand.Intn(2 << 8))
	}

	if api.AuthEnable == true {
		if token, err := api.GetAuthorize(api.AuthKey, api.AuthValue); err != nil {
			panic("get token error")
		} else {
			api.Token = "Bearer " + token
		}
	}
}

//Test push single tag of repository many times
func TestConcurPushSingleAPPWithTag(t *testing.T) {
	if err := createfile(file); err != nil {
		t.Fatal(err)
	}
	defer removeFile(file, manifest, "testdata", "_tmp/tmp", "_tmp", "tmp")

	for i := 0; i < Concurfreq; i++ {
		go func() {
			priapp := api.AppV1App{
				OS:   ops,
				Arch: arch,
				App:  file,
				Tag:  tag,
			}

			prirepo := api.AppV1Repo{
				URI:        api.DockyardURI,
				Namespace:  api.UserName,
				Repository: repository,
			}
			_, code, err := appv1.PushApp(prirepo, priapp, api.Token, file, manifest)
			if err != nil || code != 202 {
				if err != nil {
					//	fmt.Println("push app error: ", err.Error())
				}
				//fmt.Println("push app code=", code)
				Concurdata <- false
			} else {
				Concurdata <- true
			}
		}()
	}

	success := 0
	cnt := 0
	for {
		if v, ok := <-Concurdata; ok {
			cnt++
			if v == true {
				success++
			}
			if cnt == Concurfreq {
				if success != 1 {
					fmt.Println("success cnt=", success)
					t.Fatal(fmt.Errorf("Concur push failed"))
				}
				break
			}
		}
	}
	//close(Concurdata)

}

//Test push multiple tag of repository
func TestConcurPushAPPWithMutiTag(t *testing.T) {
	if err := createfile(file); err != nil {
		t.Fatal(err)
	}
	defer removeFile(file, manifest, "testdata", "_tmp/tmp", "_tmp", "tmp")

	for i := 0; i < Concurfreq; i++ {
		go func(i int) {
			priapp := api.AppV1App{
				OS:   ops,
				Arch: arch,
				App:  file,
				Tag:  tag,
			}

			prirepo := api.AppV1Repo{
				URI:        api.DockyardURI,
				Namespace:  api.UserName,
				Repository: repository,
			}
			priapp.Tag = strconv.Itoa(i)
			_, code, err := appv1.PushApp(prirepo, priapp, api.Token, file, manifest)
			if err != nil || code != 202 {
				Concurdata <- false
			} else {
				Concurdata <- true
			}
		}(i)
	}

	success := 0
	cnt := 0
	for {
		if v, ok := <-Concurdata; ok {
			cnt++
			if v == true {
				success++
			}
			if cnt == Concurfreq {
				if success != 1 {
					fmt.Println("success cnt=", success)
					t.Fatal(fmt.Errorf("Concur push failed"))
				}
				break
			}
		}
	}
	//close(Concurdata)

}

//Test push multiple repository
func TestConcurPushMutiAPP(t *testing.T) {
	for i := 0; i < Concurfreq; i++ {
		mutifile := strconv.Itoa(i) + file
		if err := createfile(mutifile); err != nil {
			t.Fatal(err)
		}
	}
	defer removeFile(file, manifest, "testdata", "_tmp/tmp", "_tmp", "tmp")

	for i := 0; i < Concurfreq; i++ {
		go func(i int) {
			priapp := api.AppV1App{
				OS:   ops,
				Arch: arch,
				App:  file,
				Tag:  tag,
			}

			prirepo := api.AppV1Repo{
				URI:        api.DockyardURI,
				Namespace:  api.UserName,
				Repository: repository,
			}
			prirepo.Repository = repository + strconv.Itoa(i)
			mutifile := strconv.Itoa(i) + file
			//			fmt.Println("repo.Repository=", repo.Repository)
			_, code, err := appv1.PushApp(prirepo, priapp, api.Token, mutifile, manifest)
			if err != nil || code != 202 {
				fmt.Println("repo.Repository=", prirepo.Repository, "code=", code)
				if err != nil {
					fmt.Println("error=", err)
				}
				Concurdata <- false
			} else {
				Concurdata <- true
			}
		}(i)
	}

	success := 0
	cnt := 0
	for {
		if v, ok := <-Concurdata; ok {
			cnt++
			if v == true {
				success++
			}
			if cnt == Concurfreq {
				if success != Concurfreq {
					fmt.Println("success cnt=", success)
					t.Fatal(fmt.Errorf("Concur push failed"))
				}
				break
			}
		}
	}

	for i := 0; i < Concurfreq; i++ {
		mutifile := strconv.Itoa(i) + file
		removeFile(mutifile)
	}
	//close(Concurdata)

}

//Test pull single tag of repository many times
func TestConcurPullSingleAPPWithTag(t *testing.T) {
	if err := createfile(file); err != nil {
		t.Fatal(err)
	}
	defer removeFile(file, manifest, "testdata", "_tmp/tmp", "_tmp", "tmp")

	_, code, err := appv1.PushApp(repo, app, api.Token, file, manifest)
	if err != nil {
		fmt.Println("push app error: ", err.Error())
		return
	} else if code != 202 {
		fmt.Println("push app error")
		return
	}
	for i := 0; i < Concurfreq; i++ {
		go func() {
			priapp := api.AppV1App{
				OS:   ops,
				Arch: arch,
				App:  file,
				Tag:  tag,
			}

			prirepo := api.AppV1Repo{
				URI:        api.DockyardURI,
				Namespace:  api.UserName,
				Repository: repository,
			}
			_, code, err = appv1.PullApp(prirepo, priapp, api.Token, file)
			if err != nil || code != 200 {
				Concurdata <- false
			} else {
				Concurdata <- true
			}
		}()
	}

	success := 0
	cnt := 0
	for {
		if v, ok := <-Concurdata; ok {
			cnt++
			if v == true {
				success++
			}
			if cnt == Concurfreq {
				if success != Concurfreq {
					fmt.Println("success cnt=", success)
					t.Fatal(fmt.Errorf("Concur pull failed"))
				}
				break
			}
		}
	}
	//close(Concurdata)

}

//Test pull multiple tag of repository
func TestConcurPullAPPWithMutiTag(t *testing.T) {
	if err := createfile(file); err != nil {
		t.Fatal(err)
	}
	defer removeFile(file, manifest, "testdata", "_tmp/tmp", "_tmp", "tmp")

	for i := 0; i < Concurfreq; i++ {
		app.Tag = strconv.Itoa(i)
		_, code, err := appv1.PushApp(repo, app, api.Token, file, manifest)
		if err != nil {
			fmt.Println("push app error: ", err.Error())
			return
		} else if code != 202 {
			fmt.Println("push app error")
			return
		}
	}
	for i := 0; i < Concurfreq; i++ {
		go func(i int) {
			priapp := api.AppV1App{
				OS:   ops,
				Arch: arch,
				App:  file,
				Tag:  tag,
			}

			prirepo := api.AppV1Repo{
				URI:        api.DockyardURI,
				Namespace:  api.UserName,
				Repository: repository,
			}
			priapp.Tag = strconv.Itoa(i)
			_, code, err := appv1.PullApp(prirepo, priapp, api.Token, file)
			if err != nil || code != 200 {
				Concurdata <- false
			} else {
				Concurdata <- true
			}
		}(i)
	}

	success := 0
	cnt := 0
	for {
		if v, ok := <-Concurdata; ok {
			cnt++
			if v == true {
				success++
			}
			if cnt == Concurfreq {
				if success != Concurfreq {
					fmt.Println("success cnt=", success)
					t.Fatal(fmt.Errorf("Concur pull failed"))
				}
				break
			}
		}
	}
	app.Tag = tag
	//close(Concurdata)
	/*
		if err := UnTarGz(filePath, "_tmp"); err != nil {
			t.Fatal(err)
		}
		buf, err := ioutil.ReadFile("_tmp/tmp")
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(buf, contents) {
			t.Fatal("not equal")
		}*/
}

//Test pull multiple repository
func TestConcurPullMutiAPP(t *testing.T) {
	if err := createfile(file); err != nil {
		t.Fatal(err)
	}
	defer removeFile(file, pullfile, manifest, "testdata", "_tmp/tmp", "_tmp", "tmp")

	for i := 0; i < Concurfreq; i++ {
		repo.Repository = repository + strconv.Itoa(i)
		_, code, err := appv1.PushApp(repo, app, api.Token, file, manifest)
		if err != nil {
			fmt.Println("push app error: ", err.Error())
			return
		} else if code != 202 {
			fmt.Println("push app error")
			return
		}
	}

	for i := 0; i < Concurfreq; i++ {
		go func(i int) {
			priapp := api.AppV1App{
				OS:   ops,
				Arch: arch,
				App:  file,
				Tag:  tag,
			}

			prirepo := api.AppV1Repo{
				URI:        api.DockyardURI,
				Namespace:  api.UserName,
				Repository: repository,
			}
			prirepo.Repository = repository + strconv.Itoa(i)
			_, code, err := appv1.PullApp(prirepo, priapp, api.Token, pullfile)
			if err != nil || code != 200 {
				Concurdata <- false
			} else {
				Concurdata <- true
			}
		}(i)
	}

	success := 0
	cnt := 0
	for {
		if v, ok := <-Concurdata; ok {
			cnt++
			if v == true {
				success++
			}
			if cnt == Concurfreq {
				if success != Concurfreq {
					fmt.Println("success cnt=", success)
					t.Fatal(fmt.Errorf("Concur pull failed"))
				}
				break
			}
		}
	}
	repo.Repository = repository
	//close(Concurdata)
	/*
		if err := UnTarGz(filePath, "_tmp"); err != nil {
			t.Fatal(err)
		}
		buf, err := ioutil.ReadFile("_tmp/tmp")
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(buf, contents) {
			t.Fatal("not equal")
		}*/
}

//Test delete single tag of app
func TestConcurDeleteAPPWithTag(t *testing.T) {
	if err := createfile(file); err != nil {
		t.Fatal(err)
	}
	defer removeFile(file, manifest, "testdata", "_tmp/tmp", "_tmp", "tmp")

	_, code, err := appv1.PushApp(repo, app, api.Token, file, manifest)
	if err != nil {
		fmt.Println("push app error: ", err.Error())
		return
	} else if code != 202 {
		fmt.Println("push app error")
		return
	}

	for i := 0; i < Concurfreq; i++ {
		go func() {
			priapp := api.AppV1App{
				OS:   ops,
				Arch: arch,
				App:  file,
				Tag:  tag,
			}

			prirepo := api.AppV1Repo{
				URI:        api.DockyardURI,
				Namespace:  api.UserName,
				Repository: repository,
			}
			code, err = prirepo.Delete(priapp, api.Token)
			if err != nil || code != 200 {
				Concurdata <- false
			} else {
				Concurdata <- true
			}
		}()
	}

	success := 0
	cnt := 0
	for {
		if v, ok := <-Concurdata; ok {
			cnt++
			if v == true {
				success++
			}
			if cnt == Concurfreq {
				if success == 0 {
					fmt.Println("success cnt=", success)
					t.Fatal(fmt.Errorf("Concur delete failed"))
				}
				break
			}
		}
	}
	//close(Concurdata)

}

//Test delete mutiple tag of app

func TestConcurDeleteAPPWithMutiTag(t *testing.T) {
	if err := createfile(file); err != nil {
		t.Fatal(err)
	}
	defer removeFile(file, manifest, "testdata", "_tmp/tmp", "_tmp", "tmp")

	for i := 0; i < Concurfreq; i++ {
		app.Tag = strconv.Itoa(i)
		_, code, err := appv1.PushApp(repo, app, api.Token, file, manifest)
		if err != nil {
			fmt.Println("push app error: ", err.Error())
			return
		} else if code != 202 {
			fmt.Println("push app error")
			return
		}
	}

	for i := 0; i < Concurfreq; i++ {
		go func(i int) {
			priapp := api.AppV1App{
				OS:   ops,
				Arch: arch,
				App:  file,
				Tag:  tag,
			}

			prirepo := api.AppV1Repo{
				URI:        api.DockyardURI,
				Namespace:  api.UserName,
				Repository: repository,
			}
			priapp.Tag = strconv.Itoa(i)
			code, err := prirepo.Delete(priapp, api.Token)
			if err != nil || code != 200 {
				Concurdata <- false
			} else {
				Concurdata <- true
			}
		}(i)
	}

	success := 0
	cnt := 0
	for {
		if v, ok := <-Concurdata; ok {
			cnt++
			if v == true {
				success++
			}
			if cnt == Concurfreq {
				if success == 0 {
					fmt.Println("success cnt=", success)
					t.Fatal(fmt.Errorf("Concur delete muti tag failed"))
				}
				break
			}
		}
	}
	app.Tag = tag
	//close(Concurdata)

}

//Test delete mutiple app
func TestConcurDeleteMutiAPP(t *testing.T) {
	if err := createfile(file); err != nil {
		t.Fatal(err)
	}
	defer removeFile(file, manifest, "testdata", "_tmp/tmp", "_tmp", "tmp")

	for i := 0; i < Concurfreq; i++ {
		repo.Repository = repository + strconv.Itoa(i)
		_, code, err := appv1.PushApp(repo, app, api.Token, file, manifest)
		if err != nil {
			fmt.Println("push app error: ", err.Error())
			return
		} else if code != 202 {
			fmt.Println("push app error")
			return
		}
	}

	for i := 0; i < Concurfreq; i++ {
		go func(i int) {
			priapp := api.AppV1App{
				OS:   ops,
				Arch: arch,
				App:  file,
				Tag:  tag,
			}

			prirepo := api.AppV1Repo{
				URI:        api.DockyardURI,
				Namespace:  api.UserName,
				Repository: repository,
			}
			prirepo.Repository = repository + strconv.Itoa(i)
			code, err := prirepo.Delete(priapp, api.Token)
			if err != nil || code != 200 {
				if err != nil {
					fmt.Println("failed to delete app: ", err.Error())
				}
				fmt.Println("code=", code)
				Concurdata <- false
			} else {
				Concurdata <- true
			}
		}(i)
	}

	success := 0
	cnt := 0
	for {
		if v, ok := <-Concurdata; ok {
			cnt++
			if v == true {
				success++
			}
			if cnt == Concurfreq {
				if success != Concurfreq {
					fmt.Println("success cnt=", success)
					t.Fatal(fmt.Errorf("Concur delete muti app failed"))
				}
				break
			}
		}
	}
	repo.Repository = repository
	//close(Concurdata)
}

//Test push pull single tag of app
func TestConcurPushPullAPPWithTag(t *testing.T) {
	if err := createfile(file); err != nil {
		t.Fatal(err)
	}
	defer removeFile(file, pullfile, manifest, "testdata", "_tmp/tmp", "_tmp", "tmp")

	_, code, err := appv1.PushApp(repo, app, api.Token, file, manifest)
	if err != nil {
		fmt.Println("push app error: ", err.Error())
		return
	} else if code != 202 {
		fmt.Println("push app error")
		return
	}

	go func() {
		priapp := api.AppV1App{
			OS:   ops,
			Arch: arch,
			App:  file,
			Tag:  tag,
		}

		prirepo := api.AppV1Repo{
			URI:        api.DockyardURI,
			Namespace:  api.UserName,
			Repository: repository,
		}
		_, code, err := appv1.PushApp(prirepo, priapp, api.Token, file, manifest)
		if err != nil || code != 202 {
			fmt.Println("code=", code)
			if err != nil {
				fmt.Println("push error: ", err)
			}
			Concurdata <- false
		} else {
			Concurdata <- true
		}
	}()

	go func() {
		priapp := api.AppV1App{
			OS:   ops,
			Arch: arch,
			App:  file,
			Tag:  tag,
		}

		prirepo := api.AppV1Repo{
			URI:        api.DockyardURI,
			Namespace:  api.UserName,
			Repository: repository,
		}
		_, code, err := appv1.PullApp(prirepo, priapp, api.Token, pullfile)
		if err != nil || code != 200 {
			fmt.Println("code=", code)
			if err != nil {
				fmt.Println("push error: ", err)
			}
			Concurdata <- false
		} else {
			Concurdata <- true
		}
	}()

	success := 0
	cnt := 0
	for {
		if v, ok := <-Concurdata; ok {
			cnt++
			if v == true {
				success++
			}
			if cnt == 2 {
				if success != 1 {
					fmt.Println("success cnt=", success)
					t.Fatal(fmt.Errorf("Concur delete muti app failed"))
				}
				break
			}
		}
	}
	//close(Concurdata)

}

//Test push delete single tag of app
func TestConcurPushDeleteAPPWithTag(t *testing.T) {
	if err := createfile(file); err != nil {
		t.Fatal(err)
	}
	defer removeFile(file, manifest, "testdata", "_tmp/tmp", "_tmp", "tmp")
	_, code, err := appv1.PushApp(repo, app, api.Token, file, manifest)
	if err != nil {
		fmt.Println("push app error: ", err.Error())
		return
	} else if code != 202 {
		fmt.Println("push app error")
		return
	}

	go func() {
		priapp := api.AppV1App{
			OS:   ops,
			Arch: arch,
			App:  file,
			Tag:  tag,
		}

		prirepo := api.AppV1Repo{
			URI:        api.DockyardURI,
			Namespace:  api.UserName,
			Repository: repository,
		}
		_, code, err := appv1.PushApp(prirepo, priapp, api.Token, file, manifest)
		if err != nil || code != 202 {
			Concurdata <- false
		} else {
			Concurdata <- true
		}
	}()
	go func() {
		priapp := api.AppV1App{
			OS:   ops,
			Arch: arch,
			App:  file,
			Tag:  tag,
		}

		prirepo := api.AppV1Repo{
			URI:        api.DockyardURI,
			Namespace:  api.UserName,
			Repository: repository,
		}
		code, err := prirepo.Delete(priapp, api.Token)
		if err != nil || code != 200 {
			Concurdata <- false
		} else {
			Concurdata <- true
		}
	}()

	success := 0
	cnt := 0
	for {
		if v, ok := <-Concurdata; ok {
			cnt++
			if v == true {
				success++
			}
			if cnt == 2 {
				if success != 1 {
					fmt.Println("success cnt=", success)
					t.Fatal(fmt.Errorf("Concur delete muti app failed"))
				}
				break
			}
		}
	}
	//close(Concurdata)

}

//Test concur pull delete single tag of app
func TestConcurPullDeleteAPPWithTag(t *testing.T) {
	if err := createfile(file); err != nil {
		t.Fatal(err)
	}
	defer removeFile(file, manifest, "testdata", "_tmp/tmp", "_tmp", "tmp")

	_, code, err := appv1.PushApp(repo, app, api.Token, file, manifest)
	if err != nil {
		fmt.Println("push app error: ", err.Error())
		return
	} else if code != 202 {
		fmt.Println("push app error")
		return
	}

	go func() {
		priapp := api.AppV1App{
			OS:   ops,
			Arch: arch,
			App:  file,
			Tag:  tag,
		}

		prirepo := api.AppV1Repo{
			URI:        api.DockyardURI,
			Namespace:  api.UserName,
			Repository: repository,
		}
		code, err := prirepo.Delete(priapp, api.Token)
		if err != nil || code != 200 {
			Concurdata <- false
		} else {
			Concurdata <- true
		}
	}()
	go func() {
		priapp := api.AppV1App{
			OS:   ops,
			Arch: arch,
			App:  file,
			Tag:  tag,
		}

		prirepo := api.AppV1Repo{
			URI:        api.DockyardURI,
			Namespace:  api.UserName,
			Repository: repository,
		}
		_, code, err := appv1.PullApp(prirepo, priapp, api.Token, file)
		if err != nil || code != 200 {
			Concurdata <- false
		} else {
			Concurdata <- true
		}
	}()

	success := 0
	cnt := 0
	for {
		if v, ok := <-Concurdata; ok {
			cnt++
			if v == true {
				success++
			}
			if cnt == 2 {
				if success != 1 {
					fmt.Println("success cnt=", success)
					t.Fatal(fmt.Errorf("Concur delete muti app failed"))
				}
				break
			}
		}
	}
	//close(Concurdata)

}

//Test push search single tag of app
func TestConcurPushSearchAPPWithTag(t *testing.T) {
	if err := createfile(file); err != nil {
		t.Fatal(err)
	}
	defer removeFile(file, manifest, "testdata", "_tmp/tmp", "_tmp", "tmp")

	_, code, err := appv1.PushApp(repo, app, api.Token, file, manifest)
	if err != nil {
		fmt.Println("push app error: ", err.Error())
		return
	} else if code != 202 {
		fmt.Println("push app error")
		return
	}

	go func() {
		priapp := api.AppV1App{
			OS:   ops,
			Arch: arch,
			App:  file,
			Tag:  tag,
		}

		prirepo := api.AppV1Repo{
			URI:        api.DockyardURI,
			Namespace:  api.UserName,
			Repository: repository,
		}
		_, code, err := appv1.PushApp(prirepo, priapp, api.Token, file, manifest)
		if err != nil || code != 202 {
			if err != nil {
				fmt.Println("push app error: ", err.Error())
			}
			fmt.Println("push app code=", code)
			Concurdata <- false
		} else {
			Concurdata <- true
		}
	}()
	go func() {
		query := []string{file, "latest"}
		prirepo := api.AppV1Repo{
			URI:        api.DockyardURI,
			Namespace:  api.UserName,
			Repository: repository,
		}
		_, code, err := prirepo.SearchGlobal(query, api.Token)
		if err != nil || code != 200 {
			if err != nil {
				fmt.Println("search app error: ", err.Error())
			}
			fmt.Println("search app code=", code)
			Concurdata <- false
		} else {
			Concurdata <- true
		}
	}()

	success := 0
	cnt := 0
	for {
		if v, ok := <-Concurdata; ok {
			cnt++
			if v == true {
				success++
			}
			if cnt == 2 {
				if success != 2 {
					fmt.Println("success cnt=", success)
					t.Fatal(fmt.Errorf("Concur delete muti app failed"))
				}
				break
			}
		}
	}
	//close(Concurdata)
}

//Test pull search single tag of app
func TestConcurPullSearchAPPWithTag(t *testing.T) {
	if err := createfile(file); err != nil {
		t.Fatal(err)
	}
	defer removeFile(file, manifest, "testdata", "_tmp/tmp", "_tmp", "tmp")

	_, code, err := appv1.PushApp(repo, app, api.Token, file, manifest)
	if err != nil {
		fmt.Println("push app error: ", err.Error())
		return
	} else if code != 202 {
		fmt.Println("push app error")
		return
	}
	go func() {
		priapp := api.AppV1App{
			OS:   ops,
			Arch: arch,
			App:  file,
			Tag:  tag,
		}

		prirepo := api.AppV1Repo{
			URI:        api.DockyardURI,
			Namespace:  api.UserName,
			Repository: repository,
		}
		_, code, err := appv1.PullApp(prirepo, priapp, api.Token, file)
		if err != nil || code != 200 {
			Concurdata <- false
		} else {
			Concurdata <- true
		}
	}()
	go func() {
		query := []string{file, "latest"}
		prirepo := api.AppV1Repo{
			URI:        api.DockyardURI,
			Namespace:  api.UserName,
			Repository: repository,
		}
		_, code, err := prirepo.SearchGlobal(query, api.Token)
		if err != nil || code != 200 {
			Concurdata <- false
		} else {
			Concurdata <- true
		}
	}()

	success := 0
	cnt := 0
	for {
		if v, ok := <-Concurdata; ok {
			cnt++
			if v == true {
				success++
			}
			if cnt == 2 {
				if success != 2 {
					fmt.Println("success cnt=", success)
					t.Fatal(fmt.Errorf("Concur delete muti app failed"))
				}
				break
			}
		}
	}
	//close(Concurdata)
}

func removeFile(args ...string) {
	for _, k := range args {
		os.RemoveAll(k)
	}

}

func randomContents(length int64) []byte {
	return randomBytes[:length]
}

//var pullfile = "pullwebapp-v1-linux-arm.tar.gz"
//var file = "webapp-v1-linux-arm.tar.gz"
var pullfile = "pullwebapp-v1-linux-arm.txt"
var file = "pushwebapp-v1-linux-arm.txt"
var manifest = "mnfwebapp-v1-linux-arm.txt"

func createfile(file string) error {
	contents := randomContents(1024 * 1024 * 10)
	if err := ioutil.WriteFile(file, contents, 0666); err != nil {
		return err
	}
	//	if err := TarGz("tmp", file); err != nil {
	//		return err
	//	}
	fd, err := os.Create(manifest)
	defer fd.Close()
	if err != nil {
		return err
	}
	_, err = fd.WriteString("mnfwebapp-v1-linux-arm.txt")
	if err != nil {
		return err
	}
	return nil
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
	os.Mkdir(destDirPath, 0750)
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
			os.MkdirAll(destDirPath+"/"+path.Dir(hdr.Name), 0750)
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

//Test push pull mutiple app
func TestConcurPushPullMutiAPP(t *testing.T) {
	if err := createfile(file); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < Concurfreq; i++ {
		mutifile := strconv.Itoa(i) + file
		if err := createfile(mutifile); err != nil {
			t.Fatal(err)
		}
	}
	defer removeFile(file, pullfile, manifest, "testdata", "_tmp/tmp", "_tmp", "tmp")

	twiceConcurfreq := Concurfreq * 2
	for i := Concurfreq; i < twiceConcurfreq; i++ {
		repo.Repository = repository + strconv.Itoa(i)
		_, code, err := appv1.PushApp(repo, app, api.Token, file, manifest)
		if err != nil {
			fmt.Println("push app error: ", err.Error())
			return
		} else if code != 202 {
			fmt.Println("push app error")
			return
		}
	}

	for i := 0; i < Concurfreq; i++ {
		go func(i int) {
			priapp := api.AppV1App{
				OS:   ops,
				Arch: arch,
				App:  file,
				Tag:  tag,
			}

			prirepo := api.AppV1Repo{
				URI:        api.DockyardURI,
				Namespace:  api.UserName,
				Repository: repository,
			}
			prirepo.Repository = repository + strconv.Itoa(i)
			mutifile := strconv.Itoa(i) + file
			_, code, err := appv1.PushApp(prirepo, priapp, api.Token, mutifile, manifest)
			if err != nil || code != 202 {
				fmt.Println("push code=", code)
				if err != nil {
					fmt.Println("push error: ", err)
				}
				Concurdata <- false
			} else {
				Concurdata <- true
			}
		}(i)
	}

	for i := Concurfreq; i < twiceConcurfreq; i++ {
		go func(i int) {
			priapp := api.AppV1App{
				OS:   ops,
				Arch: arch,
				App:  file,
				Tag:  tag,
			}

			prirepo := api.AppV1Repo{
				URI:        api.DockyardURI,
				Namespace:  api.UserName,
				Repository: repository,
			}
			prirepo.Repository = repository + strconv.Itoa(i)
			mutipullfile := strconv.Itoa(i) + pullfile
			_, code, err := appv1.PullApp(prirepo, priapp, api.Token, mutipullfile)
			if err != nil || code != 200 {
				fmt.Println("pull code=", code)
				if err != nil {
					fmt.Println("pull error: ", err)
				}
				Concurdata <- false
			} else {
				Concurdata <- true
			}
		}(i)
	}

	success := 0
	cnt := 0
	for {
		if v, ok := <-Concurdata; ok {
			cnt++
			if v == true {
				success++
			}
			if cnt == twiceConcurfreq {
				if success != twiceConcurfreq {
					fmt.Println("success cnt=", success)
					t.Fatal(fmt.Errorf("Concur push pull muti app failed"))
				}
				break
			}
		}
	}
	for i := 0; i < Concurfreq; i++ {
		mutifile := strconv.Itoa(i) + file
		removeFile(mutifile)
	}
	for i := Concurfreq; i < twiceConcurfreq; i++ {
		mutipullfile := strconv.Itoa(i) + pullfile
		removeFile(mutipullfile)
	}
	//close(Concurdata)
}
