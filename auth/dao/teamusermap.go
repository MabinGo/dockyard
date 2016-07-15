package dao

import (
	"fmt"
	"time"

	"github.com/huawei-openlab/newdb/orm"
)

type TeamUserMap struct {
	Id      int       `orm:"auto"`
	Team    *Team     `orm:"column(team_id);rel(fk);on_delete(cascade)"`
	User    *User     `orm:"column(user_name);rel(fk);on_delete(cascade)"`
	Role    int       `orm:"integer"`            //role of user in team, owner or member
	Status  int       `orm:"integer;default(0)"` //status: active(0) or inactive(1)
	Created time.Time `orm:"type(datetime);auto_now_add"`
	Updated time.Time `orm:"type(datetime);auto_now"`
}

//define unique index
func (teamUserMap *TeamUserMap) TableUnique() [][]string {
	return [][]string{
		[]string{"Team", "User"},
	}
}

//input: team name, team's org name;  user name
//user should in org
func (teamUserMap *TeamUserMap) Save() error {
	if teamUserMap.Role == TEAMADMIN || teamUserMap.Role == TEAMMEMBER {
	} else {
		return fmt.Errorf(StatusBadRequest + ": user role error")
	}

	if err := teamUserMap.getID(); err != nil {
		return err
	}

	o := orm.NewOrm()
	if _, err := o.Insert(teamUserMap); err != nil {
		return err
	} else {
		return nil
	}
}

//input: team name, team's org name;  user name
//user should in org
func (teamUserMap *TeamUserMap) Update(field ...string) error {
	if teamUserMap.Role == TEAMADMIN || teamUserMap.Role == TEAMMEMBER {
	} else {
		return fmt.Errorf(StatusBadRequest + ": user role error")
	}

	tum := &TeamUserMap{
		Team: teamUserMap.Team,
		User: teamUserMap.User,
	}
	if exist, err := tum.Get(); err != nil {
		return err
	} else if !exist {
		return fmt.Errorf(StatusNotFound+": team %s and user %s relation not exist",
			teamUserMap.Team.Org.Name+"_"+teamUserMap.Team.Name, teamUserMap.User.Name)
	} else if tum.Status != ACTIVE {
		return fmt.Errorf(StatusBadRequest+": team %s and user %s relation is not active",
			teamUserMap.Team.Org.Name+"_"+teamUserMap.Team.Name, teamUserMap.User.Name)
	}

	o := orm.NewOrm()
	teamUserMap.Id = tum.Id
	field = append(field, "updated")
	if num, err := o.Update(teamUserMap, field...); err != nil {
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
func (teamUserMap *TeamUserMap) Get() (bool, error) {
	if err := teamUserMap.getID(); err != nil {
		return false, err
	}

	o := orm.NewOrm()
	qs := o.QueryTable("team_user_map")
	err := qs.Filter("team_id", teamUserMap.Team.Id).
		Filter("user_name", teamUserMap.User.Name).RelatedSel().One(teamUserMap)
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
func (teamUserMap *TeamUserMap) Delete() error {
	if exist, err := teamUserMap.Get(); err != nil {
		return err
	} else if exist {
		o := orm.NewOrm()
		if _, err := o.Delete(teamUserMap); err != nil {
			return err
		} else {
			return nil
		}
	} else {
		return fmt.Errorf(StatusNotFound+": team %s and user %s relation not exist",
			teamUserMap.Team.Org.Name+"_"+teamUserMap.Team.Name, teamUserMap.User.Name)
	}
}

//input: team name, team's org name;  repo name,repo's org name
//team's org name should same with repo's org name
func (teamUserMap *TeamUserMap) Deactive() error {
	if _, err := teamUserMap.Get(); err != nil {
		return err
	}
	teamUserMap.Status = INACTIVE
	if err := teamUserMap.Update("Status"); err != nil {
		return err
	}
	return nil
}

// Active can recovery inactive team user to active
func (teamUserMap *TeamUserMap) Active() error {
	if exist, err := teamUserMap.Get(); err != nil {
		return err
	} else if !exist {
		return fmt.Errorf(StatusNotFound+": team %s and user %s relation not exist",
			teamUserMap.Team.Org.Name+"_"+teamUserMap.Team.Name, teamUserMap.User.Name)
	} else if teamUserMap.Status == ACTIVE {
		return fmt.Errorf(StatusBadRequest+": team %s and user %s relation is already active",
			teamUserMap.Team.Org.Name+"_"+teamUserMap.Team.Name, teamUserMap.User.Name)
	}

	o := orm.NewOrm()
	teamUserMap.Status = ACTIVE
	if num, err := o.Update(teamUserMap, "status"); err != nil {
		return err
	} else if num <= 0 {
		return fmt.Errorf(StatusNotFound+": team %s and user %s relation not exist",
			teamUserMap.Team.Org.Name+"_"+teamUserMap.Team.Name, teamUserMap.User.Name)
	} else {
		return nil
	}
}

func (teamUserMap *TeamUserMap) getID() error {
	if exist, err := teamUserMap.Team.Get(); err != nil {
		return err
	} else if !exist {
		return fmt.Errorf("team is not exist: organization name:%s, team name:%s", teamUserMap.Team.Org.Name, teamUserMap.Team.Name)
	} else {
		return nil
	}
}

//Get user list from team
func (teamUserMap *TeamUserMap) List() ([]TeamUserMap, error) {
	o := orm.NewOrm()
	tumlist := []TeamUserMap{}

	if exist, err := teamUserMap.Team.Get(); err != nil {
		return nil, err
	} else if !exist {
		return nil, fmt.Errorf("team is not exist: organization name:%s, team name:%s", teamUserMap.Team.Org.Name, teamUserMap.Team.Name)
	}
	if _, err := o.QueryTable("team_user_map").Filter("team_id", teamUserMap.Team.Id).All(&tumlist); err != nil {
		return nil, err
	} else {
		return tumlist, nil
	}
}
