package authz

import (
	"fmt"

	"github.com/containerops/dockyard/auth/dao"
)

type OrganizationNSPermission struct {
	User               *dao.User
	NameSpace          string
	Repo               string
	OrgMemberPrivilege int
	Actions            []string
}

//user *dao.User, nameSpace string, repo string, memberPrivilege int, actions []string
//if namespace is org name,and login user is in organizaion:
//user is org's admin, then namespace permission is W.
//user is org's member, then namespace permission is same with  member's privilege
//else is defined by repository
func (orgnsp *OrganizationNSPermission) GetPermission() (string, error) {
	permission := ""
	canCreateRepo := false
	IsRepoExist := false

	//only push can create repo if repo is not exit
	for _, a := range orgnsp.Actions {
		if a == "push" {
			canCreateRepo = true
		}
	}

	if orgnsp.User.Role == dao.SYSADMIN {
		canCreateRepo = canCreateRepo && true
		permission = "ADMIN"
	} else {
		//get org user role
		orgUserMap := &dao.OrganizationUserMap{
			Org:  &dao.Organization{Name: orgnsp.NameSpace},
			User: &dao.User{Name: orgnsp.User.Name},
		}
		exist, err := orgUserMap.Get()
		if err != nil {
			return "", err
		}
		//user in organization
		if exist {
			// status is active
			if orgUserMap.Status == dao.ACTIVE {
				//if org admin, permission is W
				if orgUserMap.Role == dao.ORGADMIN {
					canCreateRepo = canCreateRepo && true
					permission = "ADMIN"
				} else {
					canCreateRepo = false
					//memberPrivilege
					switch orgnsp.OrgMemberPrivilege {
					//case dao.ADMIN:
					//	permission = "W"
					case dao.WRITE:
						permission = "W"
					case dao.READ:
						permission = "R"
					case dao.NONE:
						permission = ""
					default:
						return "", fmt.Errorf("user's organization privilege is error")
					}
				}
			} else {
				permission = ""
				canCreateRepo = canCreateRepo && false
			}
		} else {
			permission = ""
			canCreateRepo = canCreateRepo && false
		}
	}

	//delete repo
	if orgnsp.Actions[0] == "*" {
		return permission, nil
	}

	//get repo privilege,create it if repo is not exist
	r := dao.RepositoryEx{
		Name:     orgnsp.Repo,
		IsOrgRep: true,
		Org:      &dao.Organization{Name: orgnsp.NameSpace},
	}
	if exist, err := r.Get(); err != nil {
		return "", err
	} else if exist {
		if r.Status == dao.ACTIVE {
			IsRepoExist = true
		} else {
			return "", fmt.Errorf(dao.StatusBadRequest + ": repository is inactive")
		}
	} else {
		IsRepoExist = false
	}

	if IsRepoExist {
		if r.IsPublic {
			permission = permission + "R"
		}
		//add team perimit
		if p, err := orgnsp.getTeamRepoPermission(orgnsp.User.Name, r.Id); err != nil {
			return "", err
		} else {
			permission = permission + p
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

//repo is associated with team
func (orgnsp *OrganizationNSPermission) getTeamRepoPermission(userName string, repoID int) (string, error) {
	permission := ""

	if teamPermit, err := dao.GetTeamRepoPermit(repoID, userName); err != nil {
		return "", err
	} else {
		for _, p := range teamPermit {
			if p == dao.WRITE {
				permission = permission + "W"
			} else {
				permission = permission + "R"
			}
		}
	}
	return permission, nil
}
