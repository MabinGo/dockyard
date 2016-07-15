package dao

import (
	"testing"
)

func Test_User(t *testing.T) {
	openDB(t)
	u := &User{
		Name:     "rootroot",
		Email:    "liugenping@huawei.com",
		Password: "root",
		//RealName string `orm:"size(100)";null`
		//Comment  string `orm:"size(100);null"`
		Status: 0,
		Role:   1,
	}
	if err := u.Save(); err != nil {
		t.Error(err)
	}

	u.Email = "user@gmail.com"
	u.RealName = "real name"
	u.Comment = "Comment update"
	if err := u.Update("Email", "RealName", "Comment", "Updated"); err != nil {
		t.Error(err)
	}

	u1 := &User{
		Name: "rootroot",
	}
	if b, err := u1.Get(); err != nil {
		t.Error(err)
	} else if !b {
		t.Error("Get User error: not found user")
	} else {
		if u1.Name != u.Name || u1.Email != u.Email || u1.Role != u.Role || u1.Status != u.Status {
			t.Error("Get User error: save is not same with get")
		}
	}
	if err := u.Deactive(); err != nil {
		t.Error(err)
	}
	if _, err := u1.Get(); err != nil {
		t.Fatal(err)
	} else {
		if u1.Status != INACTIVE {
			t.Fatal("Deactive failed")
		}
	}
	if err := u.Active(); err != nil {
		t.Error(err)
	}
	if _, err := u1.Get(); err != nil {
		t.Fatal(err)
	} else {
		if u1.Status != ACTIVE {
			t.Fatal("Active failed")
		}
	}

	if err := u.Delete(); err != nil {
		t.Error(err)
	}
}
