package dao

import (
	"fmt"
	"time"

	"github.com/huawei-openlab/newdb/orm"
)

type OrganizationUserMap struct {
	Id      int           `orm:"auto"`
	User    *User         `orm:"column(user_name);rel(fk);on_delete(cascade)"`
	Role    int           `orm:"integer;default(4)"` //role of user in organization, owner or member
	Org     *Organization `orm:"column(org_name);rel(fk);on_delete(cascade)"`
	Status  int           `orm:"integer;default(0)"` //status: active(0) or inactive(1)
	Created time.Time     `orm:"type(datetime);auto_now_add"`
	Updated time.Time     `orm:"type(datetime);auto_now"`
}

//define unique index
func (orgUserMap *OrganizationUserMap) TableUnique() [][]string {
	return [][]string{
		[]string{"User", "Org"},
	}
}

func (orgUserMap *OrganizationUserMap) Save() error {
	if orgUserMap.Role == ORGADMIN || orgUserMap.Role == ORGMEMBER {
	} else {
		return fmt.Errorf(StatusBadRequest + ": user role error")
	}

	o := orm.NewOrm()
	if _, err := o.Insert(orgUserMap); err != nil {
		return err
	} else {
		return nil
	}
}

func (orgUserMap *OrganizationUserMap) Update(field ...string) error {
	if orgUserMap.Role == ORGADMIN || orgUserMap.Role == ORGMEMBER {
	} else {
		return fmt.Errorf(StatusBadRequest + ": user role error")
	}

	oum := &OrganizationUserMap{
		Org:  orgUserMap.Org,
		User: orgUserMap.User,
	}
	if exist, err := oum.Get(); err != nil {
		return err
	} else if !exist {
		return fmt.Errorf(StatusNotFound+": organization %s and user %s relation not exist", oum.Org.Name, oum.User.Name)
	} else if oum.Status != ACTIVE {
		return fmt.Errorf(StatusBadRequest+": organization %s and user %s relation is not active", oum.Org.Name, oum.User.Name)
	}

	o := orm.NewOrm()
	orgUserMap.Id = oum.Id
	field = append(field, "updated")
	if num, err := o.Update(orgUserMap, field...); err != nil {
		return err
	} else if num <= 0 {
		return fmt.Errorf(StatusBadRequest + ": parameter is same with db")
	} else {
		return nil
	}
}

//return "false & nil" means no row
func (orgUserMap *OrganizationUserMap) Get() (bool, error) {
	o := orm.NewOrm()
	qs := o.QueryTable("organization_user_map")
	err := qs.Filter("org_name", orgUserMap.Org.Name).Filter("user_name", orgUserMap.User.Name).RelatedSel().One(orgUserMap)
	if err == orm.ErrNoRows {
		return false, nil
	} else if err != nil {
		return false, err
	} else {
		return true, nil
	}
}

func (orgUserMap *OrganizationUserMap) Delete() error {
	if exist, err := orgUserMap.Get(); err != nil {
		return err
	} else if exist {
		o := orm.NewOrm()
		if _, err := o.Delete(orgUserMap); err != nil {
			return err
		} else {
			return nil
		}
	} else {
		return fmt.Errorf(StatusNotFound+": organization %s and user %s relation not exist", orgUserMap.Org.Name, orgUserMap.User.Name)
	}
}

// Deactive updates all status relation to the user in the organization
// to inactive
func (orgUserMap *OrganizationUserMap) Deactive() error {
	return DeactiveOrgUserMap(orgUserMap)
}

// Active updates all status relation to the user in the organization
// to active
func (orgUserMap *OrganizationUserMap) Active() error {
	return ActiveOrgUserMap(orgUserMap)
}

//Get user list from organization
func (orgUserMap *OrganizationUserMap) List() ([]OrganizationUserMap, error) {
	o := orm.NewOrm()
	oumlist := []OrganizationUserMap{}
	if _, err := o.QueryTable("organization_user_map").Filter("org_name", orgUserMap.Org.Name).All(&oumlist); err != nil {
		return nil, err
	} else {
		return oumlist, nil
	}
}

func (orgUserMap *OrganizationUserMap) activeOrgUser() error {
	if exist, err := orgUserMap.Get(); err != nil {
		return err
	} else if !exist {
		return fmt.Errorf(StatusNotFound+": organization %s and user %s relation not exist", orgUserMap.Org.Name, orgUserMap.User.Name)
	} else if orgUserMap.Status == ACTIVE {
		return fmt.Errorf(StatusBadRequest+": organization %s and user %s relation is already active", orgUserMap.Org.Name, orgUserMap.User.Name)
	}

	o := orm.NewOrm()
	orgUserMap.Status = ACTIVE
	if num, err := o.Update(orgUserMap, "status"); err != nil {
		return err
	} else if num <= 0 {
		return fmt.Errorf(StatusNotFound+": organization %s and user %s relation not exist", orgUserMap.Org.Name, orgUserMap.User.Name)
	} else {
		return nil
	}
}
