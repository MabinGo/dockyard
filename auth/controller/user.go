package controller

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/astaxie/beego/logs"
	"gopkg.in/macaron.v1"

	"github.com/containerops/dockyard/auth/authn"
	"github.com/containerops/dockyard/auth/dao"
)

//1.user manager
//SignUp  POST
//SignIn  POST
//SignOut GET
//ResetPassword  POST
//UpdatePassword POST

//POST
func SignUp(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	body, err := ctx.Req.Body().Bytes()
	if err != nil {
		log.Error("[SignUp] Failed to get body: %v", err.Error())
		return http.StatusBadRequest, []byte("Request signup message body error")
	}

	u := dao.User{}
	if err := json.Unmarshal(body, &u); err != nil {
		log.Error("[SignUp] Failed to unmarshal request body: %v", err.Error())
		return http.StatusBadRequest, []byte("Failed to unmarshal request body")
	}
	//status and role can't be set by user
	u.Status = 0
	u.Role = dao.SYSMEMBER
	if err := u.Save(); err != nil {
		log.Error("[SignUp] Failed to save user to db: %v", err.Error())
		return ConvertError(err)
	}
	return http.StatusOK, []byte("")
}

//GET
func SignIn(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	userName, password, ok := ctx.Req.BasicAuth()
	if !ok {
		log.Error("[SignIn] Failed to decode Basic Authentication")
		return http.StatusUnauthorized, []byte("Decode Basic Authentication error")
	}
	user, err := authn.Login(userName, password)
	if err != nil {
		log.Error("[SignIn] Failed to login: %v", err.Error())
		return http.StatusUnauthorized, []byte("User or pwd error")
	}

	u := map[string]interface{}{
		"name":     user.Name,
		"email":    user.Email,
		"realname": user.RealName,
		"comment":  user.Comment,
		"status":   user.Status,
		"Role":     user.Role,
	}
	if result, err := json.Marshal(u); err != nil {
		log.Error("[SignIn] Failed to marshal user: %v", err.Error())
		return http.StatusInternalServerError, []byte(err.Error())
	} else {
		return http.StatusOK, result
	}
}

//GET
func SignOut(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	return http.StatusOK, []byte("")

}

//PUT
func UpdatePassword(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	return http.StatusOK, []byte("")
}

//GET
func ResetPassword(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	return http.StatusOK, []byte("")
}

func UpdateUser(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	//authn login user
	userName, password, ok := ctx.Req.BasicAuth()
	if !ok {
		log.Error("[Update User] Failed to decode Basic Authentication")
		return http.StatusUnauthorized, []byte("Decode Basic Authentication error")
	}
	user, err := authn.Login(userName, password)
	if err != nil {
		log.Error("[Update User] Failed to login: %v", err.Error())
		return http.StatusUnauthorized, []byte(err.Error())
	}

	body, err := ctx.Req.Body().Bytes()
	if err != nil {
		log.Error("[Update User] Failed to get body: %v", err.Error())
		return http.StatusBadRequest, []byte("Request signup message body error")
	}

	//get update field
	var dat map[string]interface{}
	if err := json.Unmarshal(body, &dat); err != nil {
		log.Error("[Update User] Failed to unmarshal request body: %v", err.Error())
		return http.StatusBadRequest, []byte("Failed to unmarshal request body")
	}
	for k := range dat {
		if k != "name" && k != "role" && k != "email" && k != "realname" && k != "comment" && k != "password" {
			log.Error("[Update User] Request body error, unknown field: " + k)
			return http.StatusBadRequest, []byte(fmt.Sprintf("Request body error, unknown field: " + k))
		}
	}

	fields := []string{}
	u := &dao.User{}
	if val, ok := dat["name"]; ok {
		u.Name = val.(string)
	}
	if val, ok := dat["role"]; ok {
		if user.Name == "root" && u.Name != "root" {
			u.Role = int(val.(float64))
			fields = append(fields, "role")
		} else {
			log.Error("[Update User] Only root can update user role and root can't change its role")
			return http.StatusBadRequest, []byte("Only root can update user role and root can't change its role")
		}
	}

	if val, ok := dat["email"]; ok {
		u.Email = val.(string)
		fields = append(fields, "email")
	}
	if val, ok := dat["realname"]; ok {
		u.RealName = val.(string)
		fields = append(fields, "realname")
	}
	if val, ok := dat["comment"]; ok {
		u.Comment = val.(string)
		fields = append(fields, "comment")
	}
	if val, ok := dat["password"]; ok {
		u.Password = val.(string)
		fields = append(fields, "password", "salt")
	}

	uTemp := dao.User{Name: u.Name}
	if exist, err := uTemp.Get(); err != nil {
		log.Error("[Update User] Get user %s error:%v", u.Name, err)
		return http.StatusInternalServerError, []byte("Get updated user error")
	} else if !exist {
		log.Error("[Update User] Updated user is inexist")
		return http.StatusBadRequest, []byte("Updated user is inexist")
	}

	if user.Name == u.Name {
	} else if user.Name == "root" {
	} else if user.Role == dao.SYSADMIN && uTemp.Role == dao.SYSMEMBER {
	} else {
		log.Error("[Update User] Unauthorized to update user info")
		return http.StatusUnauthorized, []byte("Unauthorized to update user info")
	}

	if err := u.Update(fields...); err != nil {
		log.Error("[Update User] Failed to update user info to db: %v", err.Error())
		return ConvertError(err)
	}
	return http.StatusOK, []byte("")
}

//POST
func CreateUser(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	//authn login user
	userName, password, ok := ctx.Req.BasicAuth()
	if !ok {
		log.Error("[Create User] Failed to decode Basic Authentication")
		return http.StatusUnauthorized, []byte("Decode Basic Authentication error")
	}
	user, err := authn.Login(userName, password)
	if err != nil {
		log.Error("[Create User] Failed to login: %v", err.Error())
		return http.StatusUnauthorized, []byte(err.Error())
	}

	if user.Role != dao.SYSADMIN {
		log.Error("[Create User] Only system admin can create user")
		return http.StatusUnauthorized, []byte("Only system admin can create user")
	}

	body, err := ctx.Req.Body().Bytes()
	if err != nil {
		log.Error("[Create User] Failed to get body: %v", err.Error())
		return http.StatusBadRequest, []byte("Request signup message body error")
	}

	u := dao.User{}
	if err := json.Unmarshal(body, &u); err != nil {
		log.Error("[Create User] Failed to unmarshal request body: %v", err.Error())
		return http.StatusBadRequest, []byte("Failed to unmarshal request body")
	}

	//only local host can create system admin
	remoteAddr := ctx.RemoteAddr()
	if u.Role == dao.SYSADMIN && (remoteAddr != "127.0.0.1" || user.Name != "root") {
		log.Error("[Create User] Only root user in local host can create system admin")
		return http.StatusUnauthorized, []byte("Only root user in local host can create system admin")
	}

	//status can't be set by user
	u.Status = 0
	if err := u.Save(); err != nil {
		log.Error("[Create User] Failed to save user to db: %v", err.Error())
		return ConvertError(err)
	}
	return http.StatusOK, []byte("")
}

func DeactiveUser(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	// authn login user
	userName, password, ok := ctx.Req.BasicAuth()
	if !ok {
		log.Error("[Deactive User] Failed to decode Basic Authentication")
		return http.StatusUnauthorized, []byte("Decode Basic Authentication error")
	}
	user, err := authn.Login(userName, password)
	if err != nil {
		log.Error("[Deactive User] Failed to login: %v", err.Error())
		return http.StatusUnauthorized, []byte(err.Error())
	}
	if user.Role != dao.SYSADMIN {
		log.Error("[Deactive User] Only system admin can deactive user")
		return http.StatusBadRequest, []byte("Only system admin can deactive user")
	}

	// get deactive user
	deactivedUserName := ctx.Params(":user")
	if deactivedUserName == "root" {
		log.Error("[Deactive User] Root can not be deactived")
		return http.StatusBadRequest, []byte("Root can not be deactived")
	}
	u := dao.User{Name: deactivedUserName}
	if exist, err := u.Get(); err != nil {
		log.Error("[Deactive User] Get user %s error:%v", deactivedUserName, err)
		return http.StatusInternalServerError, []byte("Get deactived user error")
	} else if !exist {
		log.Error("[Deactive User] Deactived user is inexist")
		return http.StatusBadRequest, []byte("Deactived user is inexist")
	}

	// only local host can inactive system admin
	remoteAddr := ctx.RemoteAddr()
	if u.Role == dao.SYSADMIN && (remoteAddr != "127.0.0.1" || user.Name != "root") {
		log.Error("[Deactive User] Only root user in local host can deactived system admin")
		return http.StatusUnauthorized, []byte("Only root user in local host can deactived system admin")
	}

	//deactive user
	if err := u.Deactive(); err != nil {
		log.Error("[Deactive User] Failed to deactive user: %v", err.Error())
		return ConvertError(err)
	}

	return http.StatusOK, []byte("")
}

// ActiveUser returns response of active according to the request
func ActiveUser(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	// authn login user
	userName, password, ok := ctx.Req.BasicAuth()
	if !ok {
		log.Error("[Active User] Failed to decode Basic Authentication")
		return http.StatusUnauthorized, []byte("Decode Basic Authentication error")
	}
	user, err := authn.Login(userName, password)
	if err != nil {
		log.Error("[Active User] Failed to login: " + err.Error())
		return http.StatusUnauthorized, []byte(err.Error())
	}
	if user.Role != dao.SYSADMIN {
		log.Error("[Active User] Only system admin can active user")
		return http.StatusBadRequest, []byte("Only system admin can active user")
	}

	// get active user
	deactivedUserName := ctx.Params(":user")
	if deactivedUserName == "root" {
		log.Error("[Active User] Root can not be actived")
		return http.StatusBadRequest, []byte("Root can not be actived")
	}
	u := dao.User{Name: deactivedUserName}
	if exist, err := u.Get(); err != nil {
		log.Error("[Active User] Get user %s error: "+deactivedUserName, err)
		return http.StatusInternalServerError, []byte("Get actived user error")
	} else if !exist {
		log.Error("[Active User] Actived user is inexist")
		return http.StatusBadRequest, []byte("Actived user is inexist")
	}

	// only local host can inactive system admin
	remoteAddr := ctx.RemoteAddr()
	if u.Role == dao.SYSADMIN && (remoteAddr != "127.0.0.1" || user.Name != "root") {
		log.Error("[Active User] Only root user in local host can actived system admin")
		return http.StatusUnauthorized, []byte("Only root user in local host can actived system admin")
	}

	//active user
	if err := u.Active(); err != nil {
		log.Error("[Active User] Failed to active user: " + err.Error())
		return ConvertError(err)
	}

	return http.StatusOK, []byte("")
}

//GetUserList is to get list of user
func GetUserList(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	// authn login user
	userName, password, ok := ctx.Req.BasicAuth()
	if !ok {
		log.Error("[Get User List] Failed to decode Basic Authentication")
		return http.StatusUnauthorized, []byte("Decode Basic Authentication error")
	}
	user, err := authn.Login(userName, password)
	if err != nil {
		log.Error("[Get User List] Failed to login: " + err.Error())
		return http.StatusUnauthorized, []byte(err.Error())
	}
	if user.Role != dao.SYSADMIN {
		log.Error("[Get User List] Only system admin can get user list")
		return http.StatusBadRequest, []byte("Only system admin can get user list")
	}

	// get user list
	var result []byte
	user = &dao.User{}
	if userlist, err := user.List(); err != nil {
		log.Error("[Get User List] Failed to get user list: " + err.Error())
		return ConvertError(err)
	} else if rst, err := json.Marshal(userlist); err != nil {
		log.Error("[Get User List] Failed to marshal user list: " + err.Error())
		return http.StatusInternalServerError, []byte(err.Error())
	} else {
		result = rst
	}

	return http.StatusOK, result
}

func DeleteUser(ctx *macaron.Context, log *logs.BeeLogger) (int, []byte) {
	// authn login user
	userName, password, ok := ctx.Req.BasicAuth()
	if !ok {
		log.Error("[Delete User] Failed to decode Basic Authentication")
		return http.StatusUnauthorized, []byte("Decode Basic Authentication error")
	}
	user, err := authn.Login(userName, password)
	if err != nil {
		log.Error("[Delete User] Failed to login: %v", err.Error())
		return http.StatusUnauthorized, []byte(err.Error())
	}
	if user.Role != dao.SYSADMIN {
		log.Error("[Delete User] Only system admin can delete user")
		return http.StatusBadRequest, []byte("Only system admin can delete user")
	}

	// get deleted user
	deletedUserName := ctx.Params(":user")
	if deletedUserName == "root" {
		log.Error("[Delete User] Root can not be deleted")
		return http.StatusBadRequest, []byte("Root can not be deleted")
	}
	u := dao.User{Name: deletedUserName}
	if exist, err := u.Get(); err != nil {
		log.Error("[Delete User] Get user %s error:%v", deletedUserName, err)
		return http.StatusInternalServerError, []byte("Get deleted user error")
	} else if !exist {
		log.Error("[Delete User] Deleted user is inexist")
		return http.StatusBadRequest, []byte("Deleted user is inexist")
	}

	// only local host can delete system admin
	remoteAddr := ctx.RemoteAddr()
	if u.Role == dao.SYSADMIN && (remoteAddr != "127.0.0.1" || user.Name != "root") {
		log.Error("[Delete User] Only root user in local host can delete system admin")
		return http.StatusUnauthorized, []byte("Only root user in local host can delete system admin")
	}

	//delete user
	if err := u.Delete(); err != nil {
		log.Error("[Delete User] Failed to delete user: %v", err.Error())
		return ConvertError(err)
	}

	return http.StatusOK, []byte("")
}
