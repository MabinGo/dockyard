package dao

import (
	"fmt"
	"strings"
	"time"

	"github.com/huawei-openlab/newdb/orm"
)

type Organization struct {
	//Id              int       `orm:"auto"`
	Name            string    `orm:"size(100);pk" json:"name"`
	Email           string    `orm:"size(100);null" json:"email,omitempty"`
	Comment         string    `orm:"size(100);null" json:"comment,omitempty"`
	URL             string    `orm:"column(url);size(100);null" json:"url,omitempty"`
	Location        string    `orm:"size(100);null" json:"location,omitempty"`
	MemberPrivilege int       `orm:"ingeter;default(3)" json:"memberprivilege,omitempty"` //WRITE,READ,NONE
	Status          int       `orm:"integer;default(0)" json:"status,omitempty"`          //status: active(0) or inactive(1)
	Created         time.Time `orm:"type(datetime);auto_now_add" json:"-"`
	Updated         time.Time `orm:"type(datetime);auto_now" json:"-"`
}

//return "false & nil" means no row
func (org *Organization) Get() (bool, error) {
	o := orm.NewOrm()
	if err := o.Read(org); err != nil {
		if err == orm.ErrNoRows {
			return false, nil
		} else {
			return false, err
		}
	} else {
		return true, nil
	}
}

//org name can't be same with username
func (org *Organization) Save() error {
	if strings.TrimSpace(org.Name) == "" {
		return fmt.Errorf(StatusBadRequest + ": orgnization name can't be null")
	}
	//judge legitimacy of organization name
	if err := JudgeNameLegitimacy(org.Name); err != nil {
		return err
	}
	if org.MemberPrivilege == WRITE || org.MemberPrivilege == READ || org.MemberPrivilege == NONE {
	} else {
		return fmt.Errorf(StatusBadRequest + ": orgnization member privilege is error")
	}
	u := &User{Name: org.Name}
	if b, err := u.Get(); err != nil {
		return err
	} else if b {
		return fmt.Errorf(StatusBadRequest + ": orgnization name can't same with user name")
	}

	o := orm.NewOrm()
	if _, err := o.Insert(org); err != nil {
		return err
	} else {
		return nil
	}
}

//org name can't be same with username
func (org *Organization) Update(field ...string) error {
	if strings.TrimSpace(org.Name) == "" {
		return fmt.Errorf(StatusBadRequest + ": orgnization name can't be null")
	}
	//judge legitimacy of organization name
	if err := JudgeNameLegitimacy(org.Name); err != nil {
		return err
	}
	if org.MemberPrivilege == WRITE || org.MemberPrivilege == READ || org.MemberPrivilege == NONE {
	} else {
		return fmt.Errorf(StatusBadRequest + ": orgnization member privilege is error")
	}

	organiz := &Organization{Name: org.Name}
	if exist, err := organiz.Get(); err != nil {
		return err
	} else if !exist {
		return fmt.Errorf(StatusNotFound+": organization %s not exist", org.Name)
	} else if organiz.Status != ACTIVE {
		return fmt.Errorf(StatusBadRequest+": organization %s is not active", org.Name)
	}

	o := orm.NewOrm()
	field = append(field, "updated")
	if num, err := o.Update(org, field...); err != nil {
		return err
	} else if num <= 0 {
		return fmt.Errorf(StatusBadRequest + ": parameter is same with db")
	} else {
		return nil
	}
}

func (org *Organization) Delete() error {
	o := orm.NewOrm()
	if num, err := o.Delete(org); err != nil {
		return err
	} else if num <= 0 {
		return fmt.Errorf(StatusNotFound+": organization %s not exist", org.Name)
	} else {
		return nil
	}
}

// Deactive updates all status relation to organization to inactive
func (org *Organization) Deactive() error {
	return DeactiveOrganization(org)
}

// Active updates all status relation to organization to active
func (org *Organization) Active() error {
	return ActiveOrganization(org)
}

//Get Organization list
func (org *Organization) List() ([]Organization, error) {
	o := orm.NewOrm()
	orglist := []Organization{}
	if _, err := o.QueryTable("organization").All(&orglist); err != nil {
		return nil, err
	} else {
		return orglist, nil
	}
}

func (org *Organization) activeOrg() error {
	if exist, err := org.Get(); err != nil {
		return err
	} else if !exist {
		return fmt.Errorf(StatusNotFound+": organization %s not exist", org.Name)
	} else if org.Status == ACTIVE {
		return fmt.Errorf(StatusBadRequest+": organization %s is already active", org.Name)
	}

	o := orm.NewOrm()
	org.Status = ACTIVE
	if num, err := o.Update(org, "status"); err != nil {
		return err
	} else if num <= 0 {
		return fmt.Errorf(StatusNotFound+": organization %s not exist", org.Name)
	} else {
		return nil
	}
}
