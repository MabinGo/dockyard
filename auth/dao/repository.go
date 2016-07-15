package dao

import (
	"fmt"
	"strings"
	"time"

	"github.com/huawei-openlab/newdb/orm"
)

type RepositoryEx struct {
	Id       int           `orm:"auto" json:"id"`
	Name     string        `orm:"size(100)" json:"name"`
	IsPublic bool          `orm:"bool" json:"ispublic"` //true:public; false:private
	Comment  string        `orm:"size(100);null" json:"comment,omitempty"`
	IsOrgRep bool          `orm:"bool" json:"isorgrep"` //belong to organization or user
	Org      *Organization `orm:"column(org_name);rel(fk);null;on_delete(cascade)" json:"org,omitempty"`
	User     *User         `orm:"column(user_name);rel(fk);null;on_delete(cascade)" json:"user,omitempty"`
	Status   int           `orm:"integer;default(0)" json:"status,omitempty"` //status: active(0) or inactive(1)
	Created  time.Time     `orm:"type(datetime);auto_now_add" json:"-"`
	Updated  time.Time     `orm:"type(datetime);auto_now" json:"-"`
}

//define unique index
func (repo *RepositoryEx) TableUnique() [][]string {
	return [][]string{
		[]string{"Name", "Org"},
		[]string{"Name", "User"},
	}
}

func (repo *RepositoryEx) Save() error {
	if strings.TrimSpace(repo.Name) == "" {
		return fmt.Errorf(StatusBadRequest + ": repo name can't be null")
	}
	//judge legitimacy of repository name
	if err := JudgeNameLegitimacy(repo.Name); err != nil {
		return err
	}

	o := orm.NewOrm()
	if _, err := o.Insert(repo); err != nil {
		return err
	} else {
		return nil
	}
}

func (repo *RepositoryEx) Update(field ...string) error {
	if strings.TrimSpace(repo.Name) == "" {
		return fmt.Errorf(StatusBadRequest + ": repo name can't be null")
	}
	//judge legitimacy of repository name
	if err := JudgeNameLegitimacy(repo.Name); err != nil {
		return err
	}

	r := &RepositoryEx{
		Name:     repo.Name,
		IsOrgRep: repo.IsOrgRep,
	}
	if r.IsOrgRep == true {
		r.Org = repo.Org
	} else {
		r.User = repo.User
	}

	if exist, err := r.Get(); err != nil {
		return err
	} else if !exist {
		return fmt.Errorf(StatusNotFound+": Repository %s not exist", r.Name)
	} else if r.Status != ACTIVE {
		return fmt.Errorf(StatusBadRequest+": Repository %s is not active", r.Name)
	}

	o := orm.NewOrm()
	repo.Id = r.Id
	field = append(field, "updated")
	if num, err := o.Update(repo, field...); err != nil {
		return err
	} else if num <= 0 {
		return fmt.Errorf(StatusBadRequest + ": parameter is same with db")
	} else {
		return nil
	}
}

func (repo *RepositoryEx) Delete() error {
	if exist, err := repo.Get(); err != nil {
		return err
	} else if exist {
		o := orm.NewOrm()
		if _, err := o.Delete(repo); err != nil {
			return err
		} else {
			return nil
		}
	} else {
		return fmt.Errorf(StatusNotFound+": Repository %s not existed", repo.Name)
	}
}

// Deactive updates all status relation to the repository to inactive
func (repo *RepositoryEx) Deactive() error {
	return DeactiveRepo(repo)
}

// Active updates all status relation to the repository to active
func (repo *RepositoryEx) Active() error {
	return ActiveRepo(repo)
}

//return "false & nil" means no row
func (repo *RepositoryEx) Get() (bool, error) {
	o := orm.NewOrm()
	qs := o.QueryTable("repository_ex")
	var err error
	if repo.IsOrgRep {
		err = qs.Filter("name", repo.Name).Filter("org_name", repo.Org.Name).RelatedSel().One(repo)
	} else {
		err = qs.Filter("name", repo.Name).Filter("user_name", repo.User.Name).RelatedSel().One(repo)
	}
	if err == orm.ErrNoRows {
		return false, nil
	} else if err != nil {
		return false, err
	} else {
		return true, nil
	}
}

//Get repository list
func (repo *RepositoryEx) List() ([]RepositoryEx, error) {
	o := orm.NewOrm()
	repolist := []RepositoryEx{}
	if repo.IsOrgRep == true {
		if _, err := o.QueryTable("repository_ex").Filter("org_name", repo.Org.Name).Filter("is_org_rep", true).All(&repolist); err != nil {
			return nil, err
		} else {
			return repolist, nil
		}
	} else {
		if _, err := o.QueryTable("repository_ex").Filter("user_name", repo.User.Name).Filter("is_org_rep", false).All(&repolist); err != nil {
			return nil, err
		} else {
			return repolist, nil
		}
	}
}

//Get fuzzy repository list
func (repo *RepositoryEx) FuzzyList() ([]RepositoryEx, error) {
	o := orm.NewOrm()
	repolist := []RepositoryEx{}
	if _, err := o.QueryTable("repository_ex").Filter("name__icontains", repo.Name).All(&repolist); err != nil {
		return nil, err
	} else {
		return repolist, nil
	}
}

func (repo *RepositoryEx) activeRepo() error {
	if exist, err := repo.Get(); err != nil {
		return err
	} else if !exist {
		return fmt.Errorf(StatusNotFound+": repo %s not exist", repo.Name)
	} else if repo.Status == ACTIVE {
		return fmt.Errorf(StatusBadRequest+": Repository %s is already active", repo.Name)
	}

	o := orm.NewOrm()
	repo.Status = ACTIVE
	if num, err := o.Update(repo, "status"); err != nil {
		return err
	} else if num <= 0 {
		return fmt.Errorf(StatusNotFound+": repo %s not exist", repo.Name)
	} else {
		return nil
	}
}
