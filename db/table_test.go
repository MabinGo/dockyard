package db_test

import (
	"testing"

	"github.com/containerops/dockyard/db"
	"github.com/containerops/dockyard/models"
	"github.com/containerops/dockyard/setting"
)

func Test_InitDB(t *testing.T) {
	if err := setting.SetConfig("../conf/containerops.conf"); err != nil {
		t.Fatalf("Read config failed: %v\n", err.Error())
	}

	t.Log(setting.DatabaseDriver)
	if err := db.InitDB(setting.DatabaseDriver, setting.DatabaseUser, setting.DatabasePasswd, setting.DatabaseURI, "dockyard"); err != nil {
		t.Fatal(err)
	}
	if err := db.Instance.RegisterModel(new(models.AppV1), new(models.ArtifactV1)); err != nil {
		t.Fatal(err)
	}
	if err := db.Instance.RegisterModel(new(models.DockerV2), new(models.DockerImageV2), new(models.DockerTagV2)); err != nil {
		t.Fatal(err)
	}
	if err := new(models.AppV1).AddUniqueIndex(); err != nil {
		t.Fatal(err)
	}
	if err := new(models.ArtifactV1).AddUniqueIndex(); err != nil {
		t.Fatal(err)
	}
	if err := new(models.DockerV2).AddUniqueIndex(); err != nil {
		t.Fatal(err)
	}
	if err := new(models.DockerImageV2).AddUniqueIndex(); err != nil {
		t.Fatal(err)
	}
	if err := new(models.DockerTagV2).AddUniqueIndex(); err != nil {
		t.Fatal(err)
	}
}

func Test_Create(t *testing.T) {
	appV1 := &models.AppV1{
		//Id          int64      `json:"id" gorm:"primary_key"`
		Namespace:   "huawei",
		Repository:  "dockyard",
		Description: "description",
		//Manifests   string     `json:"manifests" sql:"null;type:text"`
		//Keys        string     `json:"keys" sql:"null;type:text"`
		//Size        int64      `json:"size" sql:"default:0"`
		//Locked      bool       `json:"locked" sql:"default:false"`
	}

	if err := db.Instance.Create(appV1); err != nil {
		t.Fatal(err)
	}

	artifactV1 := &models.ArtifactV1{
		//Id        int64      `json:"id" gorm:"primary_key"`
		AppV1: appV1.Id,
		OS:    "linux",
		Arch:  "amd64",
		App:   "dockyard 1.0",
		//OSS       string     `json:"oss" sql:"null;type:text"`
		//Manifest  string     `json:"manifest" sql:"null;type:text"`
		//Path      string     `json:"arch" sql:"null;type:text"`
		//Size      int64      `json:"size" sql:"default:0"`
		//Locked    bool       `json:"locked" sql:"default:false"`
	}

	if err := db.Instance.Create(artifactV1); err != nil {
		t.Fatal(err)
	}

	repo := &models.DockerV2{
		//Id        int64      `json:"id" gorm:"primary_key"`
		Namespace:     "user",
		Repository:    "webapp",
		SchemaVersion: "V2",
		//Manifests     string     `json:"manifests" sql:"null;type:text"`
		//Agent         string     `json:"agent" sql:"null;type:text"`
		//Description   string     `json:"description" sql:"null;type:text"`
		//Size          int64      `json:"size" sql:"default:0"`
	}

	if err := db.Instance.Create(repo); err != nil {
		t.Fatal(err)
	}

	image := &models.DockerImageV2{
		//Id        int64      `json:"id" gorm:"primary_key"`
		//ImageId         string     `json:"imageid" sql:"unique;type:varchar(255)"`
		BlobSum: "a3e95jisfjsdfjsdjfk",
		//V1Compatibility string     `json:"v1compatibility" sql:"null;type:text"`
		//Path            string     `json:"path" sql:"null;type:text"`
	}

	if err := db.Instance.Create(image); err != nil {
		t.Fatal(err)
	}

	tag := &models.DockerTagV2{
		//Id        int64      `json:"id" gorm:"primary_key"`
		DockerV2: repo.Id,
		Tag:      "latest",
		//ImageId   string     `json:"imageid" sql:"not null;type:varchar(255)"`
		//Manifest  string     `json:"manifest" sql:"null;type:text"`
		//Schema    int64      `json:"schema" sql:""`
	}

	if err := db.Instance.Create(tag); err != nil {
		t.Fatal(err)
	}
}

func Test_Count(t *testing.T) {
	appV1 := &models.AppV1{
		Namespace:  "huawei",
		Repository: "dockyard",
	}

	if _, err := db.Instance.Count(appV1); err != nil {
		t.Fatal(err)
	} else if appV1.Id <= 0 {
		t.Fatal("query error")
	}
	t.Log(appV1)

	appV1 = &models.AppV1{
		Namespace:  "huawei",
		Repository: "nonExsit",
	}

	if count, err := db.Instance.Count(appV1); err != nil {
		t.Fatal(err)
	} else if count != 0 {
		t.Fatal("query error")
	}
	t.Log(appV1)
}

func Test_Save(t *testing.T) {
	appV1 := &models.AppV1{
		Namespace:  "huawei",
		Repository: "dockyard",
	}

	if _, err := db.Instance.Count(appV1); err != nil {
		t.Fatal(err)
	} else if appV1.Id <= 0 {
		t.Fatal("query error")
	}

	a := &models.AppV1{Id: appV1.Id,
		Namespace:  "huawei",
		Repository: "dockyard",
		Size:       100,
	}

	if err := db.Instance.Save(a); err != nil {
		t.Fatal(err)
	}
}

func Test_QueryM(t *testing.T) {
	appV1 := &models.AppV1{
		//Id          int64      `json:"id" gorm:"primary_key"`
		Namespace:   "huawei",
		Repository:  "build",
		Description: "description", //     `json:"description" sql:"null;type:text"`
		//Manifests   string     `json:"manifests" sql:"null;type:text"`
		//Keys        string     `json:"keys" sql:"null;type:text"`
		//Size        int64      `json:"size" sql:"default:0"`
		//Locked      bool       `json:"locked" sql:"default:false"`
	}

	if err := db.Instance.Create(appV1); err != nil {
		t.Fatal(err)
	}

	//get huawei
	results1 := []models.AppV1{}
	if err := db.Instance.QueryM(&models.AppV1{Namespace: "huawei"}, &results1); err != nil {
		t.Fatal(err)
	} else {
		t.Log(results1)
	}

	//get all
	results2 := []models.AppV1{}
	if err := db.Instance.QueryM(&models.AppV1{}, &results2); err != nil {
		t.Fatal(err)
	} else {
		t.Log(results2)
	}

}

func Test_QueryF(t *testing.T) {

	//one condition
	results1 := []models.AppV1{}
	if err := db.Instance.QueryF(&models.AppV1{Namespace: "hua"}, &results1); err != nil {
		t.Fatal(err)
	} else {
		t.Log(results1)
	}

	//two conditions
	results2 := []models.AppV1{}
	if err := db.Instance.QueryF(&models.AppV1{Namespace: "hua", Repository: "d"}, &results2); err != nil {
		t.Fatal(err)
	} else {
		t.Log(results2)
	}

	//include int64  and bool conditions
	results3 := []models.AppV1{}
	if err := db.Instance.QueryF(&models.AppV1{Id: 1, Namespace: "hua", Repository: "d", Locked: 1}, &results3); err != nil {
		t.Fatal(err)
	} else {
		t.Log(results3)
	}

	results4 := []models.AppV1{}
	if err := db.Instance.QueryF(&models.AppV1{Id: 1, Namespace: "hua", Repository: "d", Locked: 0}, &results4); err != nil {
		t.Fatal(err)
	} else {
		t.Log(results4)
	}
}

func Test_SoftDelete(t *testing.T) {

	appV1 := &models.AppV1{
		Namespace:  "huawei",
		Repository: "dockyard",
	}

	if _, err := db.Instance.Count(appV1); err != nil {
		t.Fatal(err)
	} else if appV1.Id <= 0 {
		t.Fatal("query error")
	}
	if err := db.Instance.DeleteS(&models.AppV1{Id: appV1.Id}); err != nil {
		t.Fatal(err)
	}

	artifactV1 := &models.ArtifactV1{
		//Id        int64      `json:"id" gorm:"primary_key"`
		AppV1: appV1.Id,
		OS:    "linux",
		Arch:  "amd64",
		App:   "dockyard 1.0",
	}

	if _, err := db.Instance.Count(artifactV1); err != nil {
		t.Fatal(err)
	} else if appV1.Id <= 0 {
		t.Fatal("query error")
	}

	if err := db.Instance.DeleteS(&models.ArtifactV1{Id: artifactV1.Id}); err != nil {
		t.Fatal(err)
	}
}

func Test_Clean(t *testing.T) {
	db.Instance.Delete(&models.ArtifactV1{})
	db.Instance.Delete(&models.AppV1{})
	db.Instance.Delete(&models.DockerImageV2{})
	db.Instance.Delete(&models.DockerTagV2{})
	db.Instance.Delete(&models.DockerV2{})
}
