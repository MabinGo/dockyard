package authz

import (
	"fmt"

	"github.com/containerops/dockyard/auth/dao"
)

type UserNSPermission struct {
	User      *dao.User
	NameSpace string
	Repo      string
	Actions   []string
}

//if namespace is user name, namespace permission is W,else is defined by repository
func (usernsp *UserNSPermission) GetPermission() (string, error) {
	permission := ""
	canCreateRepo := false
	IsRepoExist := false

	//only push can create repo if repo is not exit
	for _, a := range usernsp.Actions {
		if a == "push" {
			canCreateRepo = true
		}
	}

	if usernsp.User.Name == usernsp.NameSpace || usernsp.User.Role == dao.SYSADMIN {
		permission = "ADMIN"
		canCreateRepo = canCreateRepo && true
	} else {
		permission = ""
		canCreateRepo = false
	}

	//delete repo
	if usernsp.Actions[0] == "*" {
		return permission, nil
	}

	//get repo privilege,create it if repo is not exist
	r := dao.RepositoryEx{
		Name:     usernsp.Repo,
		IsOrgRep: false,
		User:     &dao.User{Name: usernsp.NameSpace},
	}
	if exist, err := r.Get(); err != nil {
		return "", err
	} else if exist {
		if r.Status == dao.ACTIVE {
			IsRepoExist = true
		} else {
			return "", fmt.Errorf(dao.StatusBadRequest + "repository is inactive")
		}
	} else {
		IsRepoExist = false
	}

	if IsRepoExist {
		if r.IsPublic {
			permission = permission + "R"
		}
	} else if !IsRepoExist && canCreateRepo {
		r.IsPublic = true
		if err := r.Save(); err != nil {
			return "", err
		}
		permission = permission + "R"
	} else {
		return "", fmt.Errorf(dao.StatusUnauthorized + ": repository is not exist and login user didn't have privilege to create it")
	}

	return permission, nil
}
