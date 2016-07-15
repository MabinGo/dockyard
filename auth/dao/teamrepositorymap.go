package dao

import (
	"fmt"
	"time"

	"github.com/huawei-openlab/newdb/orm"
)

type TeamRepositoryMap struct {
	Id      int           `orm:"auto"`
	Team    *Team         `orm:"column(team_id);rel(fk);on_delete(cascade)"`
	Repo    *RepositoryEx `orm:"column(repo_id);rel(fk);on_delete(cascade)"`
	Permit  int           `orm:"integer"`            //team's access permit for repository, WRITE or READ.
	Status  int           `orm:"integer;default(0)"` //status: active(0) or inactive(1)
	Created time.Time     `orm:"type(datetime);auto_now_add"`
	Updated time.Time     `orm:"type(datetime);auto_now"`
}

//define unique index
func (teamRepoMap *TeamRepositoryMap) TableUnique() [][]string {
	return [][]string{
		[]string{"Team", "Repo"},
	}
}

//input: team name, team's org name;  repo name,repo's org name
//team's org name should same with repo's org name
func (teamRepoMap *TeamRepositoryMap) Save() error {
	if teamRepoMap.Permit == WRITE || teamRepoMap.Permit == READ {
	} else {
		return fmt.Errorf(StatusBadRequest + ": team repo's permit error")
	}

	if err := teamRepoMap.getID(); err != nil {
		return err
	}

	o := orm.NewOrm()
	if _, err := o.Insert(teamRepoMap); err != nil {
		return err
	} else {
		return nil
	}
}

//input: team name, team's org name;  repo name,repo's org name
//team's org name should same with repo's org name
func (teamRepoMap *TeamRepositoryMap) Update(field ...string) error {
	if teamRepoMap.Permit == WRITE || teamRepoMap.Permit == READ {
	} else {
		return fmt.Errorf(StatusBadRequest + ": team repo's permit error")
	}

	trm := &TeamRepositoryMap{
		Team: teamRepoMap.Team,
		Repo: teamRepoMap.Repo,
	}

	if exist, err := trm.Get(); err != nil {
		return err
	} else if !exist {
		return fmt.Errorf(StatusNotFound+": team %s and repository %s relation not exist",
			teamRepoMap.Team.Org.Name+"_"+teamRepoMap.Team.Name, teamRepoMap.Repo.Name)
	} else if trm.Status != ACTIVE {
		return fmt.Errorf(StatusBadRequest+": team %s and repository %s relation is not active",
			teamRepoMap.Team.Org.Name+"_"+teamRepoMap.Team.Name, teamRepoMap.Repo.Name)
	}

	o := orm.NewOrm()
	teamRepoMap.Id = trm.Id
	field = append(field, "updated")
	if num, err := o.Update(teamRepoMap, field...); err != nil {
		return err
	} else if num <= 0 {
		return fmt.Errorf(StatusBadRequest + ": parameter is same with db")
	} else {
		return nil
	}
}

//input: team name, team's org name;  repo name,repo's org name
//team's org name should same with repo's org name
//return "false & nil" means no row
func (teamRepoMap *TeamRepositoryMap) Get() (bool, error) {
	if err := teamRepoMap.getID(); err != nil {
		return false, err
	}

	o := orm.NewOrm()
	qs := o.QueryTable("team_repository_map")
	err := qs.Filter("team_id", teamRepoMap.Team.Id).Filter("repo_id", teamRepoMap.Repo.Id).RelatedSel().One(teamRepoMap)
	if err == orm.ErrNoRows {
		return false, nil
	} else if err != nil {
		return false, err
	} else {
		return true, nil
	}
}

//input: team name, team's org name;  repo name,repo's org name
//team's org name should same with repo's org name
func (teamRepoMap *TeamRepositoryMap) Delete() error {
	if exist, err := teamRepoMap.Get(); err != nil {
		return err
	} else if exist {
		o := orm.NewOrm()
		if _, err := o.Delete(teamRepoMap); err != nil {
			return err
		} else {
			return nil
		}
	} else {
		return fmt.Errorf(StatusNotFound+": team %s and repository %s relation not exist",
			teamRepoMap.Team.Org.Name+"_"+teamRepoMap.Team.Name, teamRepoMap.Repo.Name)
	}
}

//input: team name, team's org name;  repo name,repo's org name
//team's org name should same with repo's org name
func (teamRepoMap *TeamRepositoryMap) Deactive() error {
	if _, err := teamRepoMap.Get(); err != nil {
		return err
	}
	teamRepoMap.Status = INACTIVE
	if err := teamRepoMap.Update("Status"); err != nil {
		return err
	}
	return nil
}

// Active can recovery inactive team repository to active
func (teamRepoMap *TeamRepositoryMap) Active() error {
	if exist, err := teamRepoMap.Get(); err != nil {
		return err
	} else if !exist {
		return fmt.Errorf(StatusNotFound+": team %s and repository %s relation not exist",
			teamRepoMap.Team.Org.Name+"_"+teamRepoMap.Team.Name, teamRepoMap.Repo.Name)
	} else if teamRepoMap.Status == ACTIVE {
		return fmt.Errorf(StatusBadRequest+": team %s and repository %s relation is already active",
			teamRepoMap.Team.Org.Name+"_"+teamRepoMap.Team.Name, teamRepoMap.Repo.Name)
	}

	o := orm.NewOrm()
	teamRepoMap.Status = ACTIVE
	if num, err := o.Update(teamRepoMap, "status"); err != nil {
		return err
	} else if num <= 0 {
		return fmt.Errorf(StatusNotFound+": team %s and repository %s relation not exist",
			teamRepoMap.Team.Org.Name+"_"+teamRepoMap.Team.Name, teamRepoMap.Repo.Name)
	} else {
		return nil
	}
}

func (teamRepoMap *TeamRepositoryMap) getID() error {
	if teamRepoMap.Team.Org.Name != teamRepoMap.Repo.Org.Name {
		return fmt.Errorf("team's org name is not same with repo's org name")
	}

	if exist, err := teamRepoMap.Team.Get(); err != nil {
		return err
	} else if !exist {
		return fmt.Errorf("team is not exist: organization name:%s, team name:%s", teamRepoMap.Team.Org.Name, teamRepoMap.Team.Name)
	}
	teamRepoMap.Repo.IsOrgRep = true
	if exist, err := teamRepoMap.Repo.Get(); err != nil {
		return err
	} else if !exist {
		return fmt.Errorf("repo is not exist: organization name:%s, repo name:%s", teamRepoMap.Team.Org.Name, teamRepoMap.Repo.Name)
	}
	return nil
}

//Get repository list from team
func (teamRepoMap *TeamRepositoryMap) List() ([]TeamRepositoryMap, error) {
	o := orm.NewOrm()
	trmlist := []TeamRepositoryMap{}

	if exist, err := teamRepoMap.Team.Get(); err != nil {
		return nil, err
	} else if !exist {
		return nil, fmt.Errorf("team is not exist: organization name:%s, team name:%s", teamRepoMap.Team.Org.Name, teamRepoMap.Team.Name)
	}
	if _, err := o.QueryTable("team_repository_map").Filter("team_id", teamRepoMap.Team.Id).All(&trmlist); err != nil {
		return nil, err
	} else {
		return trmlist, nil
	}
}
