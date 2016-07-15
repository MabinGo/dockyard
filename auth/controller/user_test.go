package controller

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/astaxie/beego/logs"
	"gopkg.in/macaron.v1"

	"github.com/containerops/dockyard/auth/dao"
	"github.com/containerops/dockyard/utils/setting"
)

func init() {
	setting.Authn = "authn_db"
}

func SignUpTest(t *testing.T, user *dao.User) {
	b, err := json.Marshal(user)
	if err != nil {
		t.Error("marshal error.")
	}

	req, err := http.NewRequest("POST", "127.0.0.1:8080\\sinup", strings.NewReader(string(b)))
	if err != nil {
		t.Error(err)
	}

	//body := strings.NewReader(string(b))
	//rc := ioutil.NopCloser(body)
	//ctx.Req.Request.Body = rc
	ctx := &macaron.Context{Req: macaron.Request{req}}
	if rt, b := SignUp(ctx, &logs.BeeLogger{}); rt != http.StatusOK {
		t.Error(string(b))
	}
}

func Test_SignUp(t *testing.T) {
	openDB(t)

	user := &dao.User{
		Name:     "wangqilin",
		Email:    "wangqilin@huawei.com",
		Password: "wangqilin",
		RealName: "wangqilin",
		Comment:  "commnet",
	}
	SignUpTest(t, user)

	u := &dao.User{Name: user.Name}
	if b, err := u.Get(); err != nil {
		t.Error(err)
	} else if !b {
		t.Error("signup user error")
	} else {
		if u.Email != user.Email || u.RealName != user.RealName ||
			u.Comment != user.Comment {
			t.Error("singup:save not same with get")
		}
	}

	if err := user.Delete(); err != nil {
		t.Error(err)
	}
}
