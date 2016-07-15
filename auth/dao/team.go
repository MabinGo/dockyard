package dao

import (
	"fmt"
	"strings"
	"time"

	"github.com/huawei-openlab/newdb/orm"
)

type Team struct {
	Id   int    `orm:"auto"`
	Name string `orm:"size(100)"`
	//IsVisible bool          `orm:"bool"` //true:visible,false:Secret.Secret teams are only visible to members of that team and to the organization’s owners.
	Comment string        `orm:"size(100);null"`
	Org     *Organization `orm:"column(org_name);rel(fk);on_delete(cascade)"`
	Status  int           `orm:"integer;default(0)"` //status: active(0) or inactive(1)
	Created time.Time     `orm:"type(datetime);auto_now_add"`
	Updated time.Time     `orm:"type(datetime);auto_now"`
}

//define unique index
func (team *Team) TableUnique() [][]string {
	return [][]string{
		[]string{"Name", "Org"},
	}
}

func (team *Team) Save() error {
	if strings.TrimSpace(team.Name) == "" {
		return fmt.Errorf(StatusBadRequest + ": team name can't be null")
	}
	//judge legitimacy of team name
	if err := JudgeNameLegitimacy(team.Name); err != nil {
		return err
	}
	o := orm.NewOrm()
	if _, err := o.Insert(team); err != nil {
		return err
	} else {
		return nil
	}
}

func (team *Team) Update(field ...string) error {
	if strings.TrimSpace(team.Name) == "" {
		return fmt.Errorf(StatusBadRequest + ": team name can't be null")
	}
	//judge legitimacy of team name
	if err := JudgeNameLegitimacy(team.Name); err != nil {
		return err
	}
	t := &Team{
		Name: team.Name,
		Org:  team.Org,
	}
	if exist, err := t.Get(); err != nil {
		return err
	} else if !exist {
		return fmt.Errorf(StatusNotFound+": team %s not exist", team.Name)
	} else if t.Status != ACTIVE {
		return fmt.Errorf(StatusBadRequest+": team %s is not active", team.Name)
	}

	o := orm.NewOrm()
	team.Id = t.Id
	field = append(field, "updated")
	if num, err := o.Update(team, field...); err != nil {
		return err
	} else if num <= 0 {
		return fmt.Errorf(StatusBadRequest + ": parameter is same with db")
	} else {
		return nil
	}
}

func (team *Team) Delete() error {
	if exist, err := team.Get(); err != nil {
		return err
	} else if exist {
		o := orm.NewOrm()
		if _, err := o.Delete(team); err != nil {
			return err
		} else {
			return nil
		}
	} else {
		return fmt.Errorf(StatusNotFound+": team %s not exist", team.Name)
	}
}

// Deactive updates all status relation to the team to inactive
func (team *Team) Deactive() error {
	return DeactiveTeam(team)
}

// Active updates all status relation to the team to active
func (team *Team) Active() error {
	return ActiveTeam(team)
}

//return "false & nil" means no row
func (team *Team) Get() (bool, error) {
	o := orm.NewOrm()
	qs := o.QueryTable("team")
	err := qs.Filter("name", team.Name).Filter("org_name", team.Org.Name).RelatedSel().One(team)
	if err == orm.ErrNoRows {
		return false, nil
	} else if err != nil {
		return false, err
	} else {
		return true, nil
	}
}

//Get team list from organization
func (team *Team) List() ([]Team, error) {
	o := orm.NewOrm()
	teamlist := []Team{}
	if _, err := o.QueryTable("team").Filter("org_name", team.Org.Name).All(&teamlist); err != nil {
		return nil, err
	} else {
		return teamlist, err
	}
}

func (team *Team) activeTeam() error {
	if exist, err := team.Get(); err != nil {
		return err
	} else if !exist {
		return fmt.Errorf(StatusNotFound+": team %s not exist", team.Name)
	} else if team.Status == ACTIVE {
		return fmt.Errorf(StatusBadRequest+": team %s is already active", team.Name)
	}

	o := orm.NewOrm()
	team.Status = ACTIVE
	if num, err := o.Update(team, "status"); err != nil {
		return err
	} else if num <= 0 {
		return fmt.Errorf(StatusNotFound+": team %s not exist", team.Name)
	} else {
		return nil
	}
}
