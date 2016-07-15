package dao

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/huawei-openlab/newdb/orm"
)

func CreateOrganization(org *Organization, oum *OrganizationUserMap) error {
	o := orm.NewOrm()

	if err := o.Begin(); err != nil {
		return err
	}
	if err := org.Save(); err != nil {
		return err
	}

	if err := oum.Save(); err != nil {
		o.Rollback()
		return err
	}
	o.Commit()
	return nil
}

func CreateTeam(team *Team, tum *TeamUserMap) error {
	o := orm.NewOrm()

	if err := o.Begin(); err != nil {
		return err
	}
	if err := team.Save(); err != nil {
		return err
	}

	if err := tum.Save(); err != nil {
		o.Rollback()
		return err
	}
	o.Commit()
	return nil
}

func GetTeamRepoPermit(repoID int, userName string) ([]int, error) {
	SQL := "select permit from team_repository_map inner join team_user_map on team_user_map.team_id=team_repository_map.team_id and team_user_map.status=0 and team_repository_map.status=0 where repo_id=? and user_name=?"
	list := orm.ParamsList{}
	o := orm.NewOrm()
	if _, err := o.Raw(SQL, repoID, userName).ValuesFlat(&list); err != nil {
		return nil, err
	} else {
		permit := []int{}
		for _, a := range list {
			if i, err := strconv.Atoi(a.(string)); err != nil {
				return nil, err
			} else {
				permit = append(permit, i)
			}
		}
		return permit, nil
	}
}

func JudgeNameLegitimacy(Name string) error {
	//judge legitimacy of signup name
	reg, err := regexp.Compile("[a-z0-9](?:-*[a-z0-9])*(?:[._][a-z0-9](?:-*[a-z0-9])*)*")
	if err != nil {
		return fmt.Errorf(StatusBadRequest + ": failed to compile regexp object")
	}
	match := false
	regstr := reg.FindAllString(Name, -1)
	for _, v := range regstr {
		if strings.Compare(Name, v) == 0 {
			match = true
		}
	}
	if match == false {
		return fmt.Errorf(StatusBadRequest + ": name is not legal")
	}
	return nil
}

// DeactiveTeam updates all status relation to the team to inactive
func DeactiveTeam(team *Team) error {
	sqlTum := "UPDATE team_user_map SET status=1 where team_id =?"
	sqlTrm := "UPDATE team_repository_map SET status=1 where team_id =?"

	o := orm.NewOrm()
	if err := o.Begin(); err != nil {
		return err
	}

	if _, err := team.Get(); err != nil {
		return err
	}
	team.Status = INACTIVE
	if err := team.Update("Status"); err != nil {
		return err
	}

	if _, err := o.Raw(sqlTum, team.Id).Exec(); err != nil {
		o.Rollback()
		return err
	}
	if _, err := o.Raw(sqlTrm, team.Id).Exec(); err != nil {
		o.Rollback()
		return err
	}

	o.Commit()
	return nil
}

// ActiveTeam updates all status relation to the team to active
func ActiveTeam(team *Team) error {
	sqlTum := "UPDATE team_user_map SET status=0 where team_id =?"
	sqlTrm := "UPDATE team_repository_map SET status=0 where team_id =?"

	o := orm.NewOrm()
	if err := o.Begin(); err != nil {
		return err
	}

	if _, err := team.Get(); err != nil {
		return err
	}
	if err := team.activeTeam(); err != nil {
		return err
	}

	if _, err := o.Raw(sqlTum, team.Id).Exec(); err != nil {
		o.Rollback()
		return err
	}
	if _, err := o.Raw(sqlTrm, team.Id).Exec(); err != nil {
		o.Rollback()
		return err
	}

	o.Commit()
	return nil
}

// DeactiveRepo updates all status relation to the repository to inactive
func DeactiveRepo(repo *RepositoryEx) error {
	sql := "UPDATE team_repository_map trm inner join team on team.id=trm.team_id SET trm.status=1 where repo_id=? and org_name=?"
	o := orm.NewOrm()

	if err := o.Begin(); err != nil {
		return err
	}

	if _, err := repo.Get(); err != nil {
		return err
	}
	repo.Status = INACTIVE
	if err := repo.Update("Status"); err != nil {
		return err
	}

	if _, err := o.Raw(sql, repo.Id, repo.Org.Name).Exec(); err != nil {
		o.Rollback()
		return err
	}

	o.Commit()
	return nil
}

// ActiveRepo updates all status relation to the repository to active
func ActiveRepo(repo *RepositoryEx) error {
	sql := "UPDATE team_repository_map trm inner join team on team.id=trm.team_id SET trm.status=0 where repo_id=? and org_name=?"
	o := orm.NewOrm()

	if err := o.Begin(); err != nil {
		return err
	}

	if _, err := repo.Get(); err != nil {
		return err
	}
	if err := repo.activeRepo(); err != nil {
		return err
	}

	if _, err := o.Raw(sql, repo.Id, repo.Org.Name).Exec(); err != nil {
		o.Rollback()
		return err
	}

	o.Commit()
	return nil
}

// DeactiveOrgUserMap updates all status relation to the user in the organization to inactive
func DeactiveOrgUserMap(orgUserMap *OrganizationUserMap) error {
	sql := "UPDATE team_user_map tum inner join team on team.id=tum.team_id SET tum.status=1 where user_name=? and org_name=?"
	o := orm.NewOrm()

	if err := o.Begin(); err != nil {
		return err
	}

	if _, err := orgUserMap.Get(); err != nil {
		return err
	}
	orgUserMap.Status = INACTIVE
	if err := orgUserMap.Update("Status"); err != nil {
		return err
	}

	if _, err := o.Raw(sql, orgUserMap.User.Name, orgUserMap.Org.Name).Exec(); err != nil {
		o.Rollback()
		return err
	}

	o.Commit()
	return nil
}

// ActiveOrgUserMap updates all status relation to the user in the organization to active
func ActiveOrgUserMap(orgUserMap *OrganizationUserMap) error {
	sql := "UPDATE team_user_map tum inner join team on team.id=tum.team_id SET tum.status=0 where user_name=? and org_name=?"
	o := orm.NewOrm()

	if err := o.Begin(); err != nil {
		return err
	}

	if _, err := orgUserMap.Get(); err != nil {
		return err
	}
	if err := orgUserMap.activeOrgUser(); err != nil {
		return err
	}

	if _, err := o.Raw(sql, orgUserMap.User.Name, orgUserMap.Org.Name).Exec(); err != nil {
		o.Rollback()
		return err
	}

	o.Commit()
	return nil
}

// DeactiveOrganization updates all status relation to organization to inactive
func DeactiveOrganization(org *Organization) error {
	sqlOum := "UPDATE organization_user_map SET status=1 where org_name =?"
	sqlRepo := "UPDATE repository_ex SET status=1 where org_name =?"
	sqlTeam := "UPDATE team SET status=1 where org_name =?"
	sqlTum := "UPDATE team_user_map tum inner join team on team.id=tum.team_id SET tum.status=1 where org_name=?"
	sqlTrm := "UPDATE team_repository_map trm inner join team on team.id=trm.team_id SET trm.status=1 where org_name=?"

	o := orm.NewOrm()
	if err := o.Begin(); err != nil {
		return err
	}

	if _, err := org.Get(); err != nil {
		return err
	}
	org.Status = INACTIVE
	if err := org.Update("Status"); err != nil {
		return err
	}

	if _, err := o.Raw(sqlOum, org.Name).Exec(); err != nil {
		o.Rollback()
		return err
	}

	if _, err := o.Raw(sqlRepo, org.Name).Exec(); err != nil {
		o.Rollback()
		return err
	}

	if _, err := o.Raw(sqlTeam, org.Name).Exec(); err != nil {
		o.Rollback()
		return err
	}

	if _, err := o.Raw(sqlTum, org.Name).Exec(); err != nil {
		o.Rollback()
		return err
	}

	if _, err := o.Raw(sqlTrm, org.Name).Exec(); err != nil {
		o.Rollback()
		return err
	}

	o.Commit()
	return nil
}

// ActiveOrganization updates all status relation to organization to active
func ActiveOrganization(org *Organization) error {
	sqlOum := "UPDATE organization_user_map SET status=0 where org_name = ?"
	sqlRepo := "UPDATE repository_ex SET status=0 where org_name = ?"
	sqlTeam := "UPDATE team SET status=0 where org_name = ?"
	sqlTum := "UPDATE team_user_map tum inner join team on team.id=tum.team_id SET tum.status=0 where org_name=?"
	sqlTrm := "UPDATE team_repository_map trm inner join team on team.id=trm.team_id SET trm.status=0 where org_name=?"

	o := orm.NewOrm()
	if err := o.Begin(); err != nil {
		return err
	}

	if _, err := org.Get(); err != nil {
		return err
	}
	if err := org.activeOrg(); err != nil {
		return err
	}

	if _, err := o.Raw(sqlOum, org.Name).Exec(); err != nil {
		o.Rollback()
		return err
	}
	if _, err := o.Raw(sqlRepo, org.Name).Exec(); err != nil {
		o.Rollback()
		return err
	}
	if _, err := o.Raw(sqlTeam, org.Name).Exec(); err != nil {
		o.Rollback()
		return err
	}
	if _, err := o.Raw(sqlTum, org.Name).Exec(); err != nil {
		o.Rollback()
		return err
	}
	if _, err := o.Raw(sqlTrm, org.Name).Exec(); err != nil {
		o.Rollback()
		return err
	}

	o.Commit()
	return nil
}

// DeactiveUser updates all status relation to the user to inactive
func DeactiveUser(user *User) error {
	sqlRepo := "UPDATE repository_ex SET status=1 where user_name = ?"
	sqlOum := "UPDATE organization_user_map SET status=1 where user_name = ?"
	sqlTum := "UPDATE team_user_map tum SET tum.status=1 where user_name=?"

	o := orm.NewOrm()
	if err := o.Begin(); err != nil {
		return err
	}

	if _, err := user.Get(); err != nil {
		return err
	}
	user.Status = INACTIVE
	if err := user.Update("Status"); err != nil {
		return err
	}

	if _, err := o.Raw(sqlRepo, user.Name).Exec(); err != nil {
		o.Rollback()
		return err
	}
	if _, err := o.Raw(sqlOum, user.Name).Exec(); err != nil {
		o.Rollback()
		return err
	}
	if _, err := o.Raw(sqlTum, user.Name).Exec(); err != nil {
		o.Rollback()
		return err
	}

	o.Commit()
	return nil
}

// ActiveUser updates all status relation to the user to active
func ActiveUser(user *User) error {
	sqlRepo := "UPDATE repository_ex SET status=0 where user_name = ?"
	sqlOum := "UPDATE organization_user_map SET status=0 where user_name = ?"
	sqlTum := "UPDATE team_user_map tum SET tum.status=0 where user_name=?"

	o := orm.NewOrm()
	if err := o.Begin(); err != nil {
		return err
	}

	if _, err := user.Get(); err != nil {
		return err
	}
	if err := user.activeUser(); err != nil {
		return err
	}

	if _, err := o.Raw(sqlRepo, user.Name).Exec(); err != nil {
		o.Rollback()
		return err
	}
	if _, err := o.Raw(sqlOum, user.Name).Exec(); err != nil {
		o.Rollback()
		return err
	}
	if _, err := o.Raw(sqlTum, user.Name).Exec(); err != nil {
		o.Rollback()
		return err
	}

	o.Commit()
	return nil
}
