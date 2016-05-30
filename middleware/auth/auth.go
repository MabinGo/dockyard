package auth

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/astaxie/beego/logs"
	"gopkg.in/macaron.v1"

	"github.com/containerops/dockyard/middleware"
	"github.com/containerops/dockyard/models"
	"github.com/containerops/dockyard/module"
	"github.com/containerops/dockyard/utils/setting"
)

type authorization struct{}

type Challenge interface {
	error
	SetHeaders(w http.ResponseWriter)
}

type Resource struct {
	Type string
	Name string
}

type Access struct {
	Resource
	Action string
}

type AccessController interface {
	InitFunc() error
	Authorized(authorization string, access ...Access) (string, error)
}

var accessControllers = map[string]AccessController{}
var Log *logs.BeeLogger

func init() {
	middleware.Register("auth", middleware.HandlerInterface(&authorization{}))
}

func Register(name string, acctrl AccessController) error {
	if _, exists := accessControllers[name]; exists {
		return fmt.Errorf("name already registered: %s", name)
	}

	accessControllers[name] = acctrl

	return nil
}

func (a *authorization) InitFunc() error {
	if setting.RunMode == "dev" {
		Log = logs.NewLogger(4096)
		Log.SetLogger("console", "")
	}

	return accessControllers[setting.Authmode].InitFunc()
}

func (a *authorization) Handler(ctx *macaron.Context) {
	//skip router of authorization center
	if strings.Contains(ctx.Req.RequestURI, "/uam/") {
		return
	}

	//exception handler,request by command,just like get _catalog, delete mainifest or blob
	if ignorecmd, err := cmdReqHandler(ctx); err != nil {
		Log.Error("Authorized err: %v", err.Error())
		ctx.Resp.WriteHeader(http.StatusUnauthorized)
		return
	} else if !ignorecmd {
		Log.Info("Authorized successfully")
		return
	} else {
		//do not use command to request,call the default handler
	}

	//default handler,request by docker daemon
	if err := authorized(ctx); err != nil {
		ctx.Resp.WriteHeader(http.StatusUnauthorized)
	}
}

//add for IT department
func cmdReqHandler(ctx *macaron.Context) (bool, error) {
	//filter docker client request
	if strings.Compare(ctx.Req.RequestURI, "/v2/") == 0 {
		return true, nil
	}
	author := ctx.Req.Header.Get("Authorization")
	parts := strings.Split(author, " ")
	partslen := len(parts)
	sign := strings.ToLower(parts[0])
	if partslen == 2 && sign == "bearer" {
		return true, nil
	}

	if partslen != 2 || sign != "basic" {
		return false, fmt.Errorf("invalid user name or password")
	}

	if strings.Compare(ctx.Req.RequestURI, "/syn/drc") == 0 {
		if err := drcVerifyHandler(author); err != nil {
			return false, err
		}

		return false, nil
	}

	var repo string
	var accessRecords []Access
	w := ctx.Resp
	r := ctx.Req.Request

	//TODO: support domains/repo format soon
	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")

	if namespace == "" || repository == "" {
		repo = ""
	} else {
		repo = fmt.Sprintf("%v/%v", namespace, repository)
	}
	if repo != "" {
		accessRecords = appendAccessRecords(accessRecords, r.Method, repo)
	} else {
		if nameRequired(r) {
			return false, fmt.Errorf("name required")
		}
		accessRecords = appendCatalogAccessRecord(accessRecords, r)
	}

	var isdel string
	acclen := len(accessRecords)
	if acclen <= 0 {
		return false, fmt.Errorf("bad access")
	}
	account, _, _ := ctx.Req.BasicAuth()
	typ := accessRecords[0].Type
	name := accessRecords[0].Name
	action := accessRecords[0].Action
	for i := 1; i < acclen; i++ {
		action = action + "," + accessRecords[i].Action
	}
	service := setting.Service
	realm := setting.Realm
	if r.Method == "DELETE" && ctx.Params(":reference") != "" {
		rep := new(models.Repository)
		if exists, err := rep.Get(namespace, repository); err != nil {
			return false, err
		} else if !exists {
			return false, fmt.Errorf("not found %v", repo)
		}

		if tagslist := rep.GetTagslist(); len(tagslist) <= 1 {
			isdel = "delete=true"
		} else {
			isdel = "delete=false"
		}
	}
	params := fmt.Sprintf("?account=%v&scope=%v:%v:%v&service=%v&%v",
		account, typ, name, action, service, isdel)

	authType := r.Method
	switch authType {
	case "GET", "HEAD":
		authType = "GET"
	case "DELETE":
	case "POST", "PUT", "PATCH":
		authType = "POST"
	default:
		return false, fmt.Errorf("not support auth type %s", authType)
	}

	resp, err := module.SendHttpRequest(authType, realm+params, nil, author)
	if err != nil {
		return false, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	token := map[string]string{}
	if err = json.Unmarshal(body, &token); err != nil {
		return false, err
	}
	bearer := "Bearer" + " " + token["token"]
	if _, err := accessControllers[setting.Authmode].Authorized(bearer, accessRecords...); err != nil {
		switch err := err.(type) {
		case Challenge:
			err.SetHeaders(w)
		default:
			w.WriteHeader(http.StatusBadRequest)
		}

		return false, err
	}

	return false, nil
}

func drcVerifyHandler(author string) error {
	rawurl := fmt.Sprintf("%s://%s/uam/user/signin", setting.ListenMode, setting.Domains)
	resp, err := module.SendHttpRequest("GET", rawurl, nil, author)
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	u := map[string]interface{}{}
	if err := json.Unmarshal(body, &u); err != nil {
		return err
	}

	if u["Role"].(float64) != 1 {
		return fmt.Errorf("no authority")
	}

	return nil
}

func authorized(ctx *macaron.Context) error {
	var repo string
	var accessRecords []Access

	namespace := ctx.Params(":namespace")
	repository := ctx.Params(":repository")
	author := ctx.Req.Header.Get("Authorization")

	w := ctx.Resp
	r := ctx.Req.Request

	if namespace == "" || repository == "" {
		repo = ""
	} else {
		repo = namespace + "/" + repository
	}

	if repo != "" {
		accessRecords = appendAccessRecords(accessRecords, r.Method, repo)
		if fromRepo := r.FormValue("from"); fromRepo != "" {
			accessRecords = appendAccessRecords(accessRecords, "GET", fromRepo)
		}
	} else {
		if nameRequired(r) {
			return fmt.Errorf("forbidden: no repository name")
		}
		accessRecords = appendCatalogAccessRecord(accessRecords, r)
	}

	_, err := accessControllers[setting.Authmode].Authorized(author, accessRecords...)
	if err != nil {
		switch err := err.(type) {
		case Challenge:
			err.SetHeaders(w)
		default:
			w.WriteHeader(http.StatusBadRequest)
		}

		return err
	}

	return nil
}

func nameRequired(r *http.Request) bool {
	if strings.Compare(r.RequestURI, "/v2/") != 0 && strings.Compare(r.RequestURI, "/v2/_catalog") != 0 {
		return true
	}

	return false
}

func appendAccessRecords(records []Access, method string, repo string) []Access {
	resource := Resource{
		Type: "repository",
		Name: repo,
	}

	switch method {
	case "GET", "HEAD":
		records = append(records,
			Access{
				Resource: resource,
				Action:   "pull",
			})
	case "POST", "PUT", "PATCH":
		records = append(records,
			Access{
				Resource: resource,
				Action:   "pull",
			},
			Access{
				Resource: resource,
				Action:   "push",
			})
	case "DELETE":
		records = append(records,
			Access{
				Resource: resource,
				Action:   "*",
			})
	}
	return records
}

func appendCatalogAccessRecord(accessRecords []Access, r *http.Request) []Access {
	if strings.Compare(r.RequestURI, "/v2/_catalog") == 0 {
		resource := Resource{
			Type: "registry",
			Name: "catalog",
		}

		accessRecords = append(accessRecords,
			Access{
				Resource: resource,
				Action:   "*",
			})
	}
	return accessRecords
}
