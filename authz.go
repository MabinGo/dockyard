package controller

import (
	"fmt"

	"github.com/astaxie/beego/logs"
	"gopkg.in/macaron.v1"

	"github.com/containerops/dockyard/auth/authz"
)

// Get handles GET request, it checks the http header for user credentials
// and parse service and scope based on docker registry v2 standard,
// checkes the permission agains local DB and generates jwt token.

func GetAuthorize(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {

	userName, password, _ := ctx.Req.BasicAuth()

	service := ctx.Query("service")
	scope := ctx.Query("scope")
	//log.Info("service: %v,scope: %v", service, scope)

	status, token := authz.GetAuthorize(userName, password, service, scope)
	ctx.Resp.Header().Set("Content-Type", "application/json; charset=utf-8")
	ctx.Resp.Header().Set("Content-Length", fmt.Sprint(len(token)))

	//log.Info("token: %v", string(token))
	//log.Info("status: %d", status)
	return status, token
}

func DeleteAuthorize(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {

	userName, password, _ := ctx.Req.BasicAuth()

	service := ctx.Query("service")
	scope := ctx.Query("scope")
	isdel := ctx.Query("delete")
	//log.Info("service: %v,scope: %v", service, scope)

	status, token := authz.DeleteAuthorize(userName, password, service, scope, isdel)
	ctx.Resp.Header().Set("Content-Type", "application/json; charset=utf-8")
	ctx.Resp.Header().Set("Content-Length", fmt.Sprint(len(token)))

	//log.Info("token: %v", string(token))
	//log.Info("status: %d", status)
	return status, token
}

func PostAuthorize(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	userName, password, _ := ctx.Req.BasicAuth()
	service := ctx.Query("service")
	scope := ctx.Query("scope")
	//log.Info("service: %v,scope: %v", service, scope)

	status, token := authz.PostAuthorize(userName, password, service, scope)
	ctx.Resp.Header().Set("Content-Type", "application/json; charset=utf-8")
	ctx.Resp.Header().Set("Content-Length", fmt.Sprint(len(token)))

	//log.Info("token: %v", string(token))
	//log.Info("status: %d", status)
	return status, token
}
