package dao

import (
	"crypto/md5"
	"crypto/rand"
	"fmt"
	"strings"
	"time"

	"github.com/huawei-openlab/newdb/orm"
)

type User struct {
	//Id       int    `orm:"auto"`
	Name     string    `orm:"size(100);pk" json:"name"`
	Email    string    `orm:"size(100)"  json:"email"`
	Password string    `orm:"size(100)" json:"password"`
	RealName string    `orm:"size(100);null" json:"realname,omitempty"`
	Comment  string    `orm:"size(100);null" json:"comment,omitempty"`    //0:actived,1:deleted
	Status   int       `orm:"integer;default(0)" json:"status,omitempty"` //status: active(0) or inactive(1)
	Role     int       `orm:"integer;default(2)" json:"role"`             //SYSADMIN(1) or SYSMEMBER(2),default is SYSMEMBER(2)
	Salt     string    `orm:"size(100)" json:"-"`
	Created  time.Time `orm:"type(datetime);auto_now_add" json:"-"`
	Updated  time.Time `orm:"type(datetime);auto_now" json:"-"`
}

/*
*return false, nil means no row
 */
func (user *User) Get() (bool, error) {
	o := orm.NewOrm()
	if err := o.Read(user); err != nil {
		if err == orm.ErrNoRows {
			return false, nil
		} else {
			return false, err
		}
	} else {
		return true, nil
	}
}

//user name can't be same with organization name
func (user *User) Save() error {
	if strings.TrimSpace(user.Name) == "" {
		return fmt.Errorf(StatusBadRequest + ": user name can't be null")
	}

	//judge legitimacy of user name
	if err := JudgeNameLegitimacy(user.Name); err != nil {
		return err
	}

	if user.Role == SYSADMIN || user.Role == SYSMEMBER {
	} else {
		return fmt.Errorf(StatusBadRequest + ": user role error")
	}

	//encrypt pwd
	if pwd, salt, err := generatePwdAndSalt(user.Password); err != nil {
		return fmt.Errorf("encrypt pwd error")
	} else {
		user.Password = pwd
		user.Salt = salt
	}

	org := Organization{Name: user.Name}
	if b, err := org.Get(); err != nil {
		return err
	} else if b {
		return fmt.Errorf(StatusBadRequest + ": user name can't same with organization name")
	}

	o := orm.NewOrm()
	if _, err := o.Insert(user); err != nil {
		return err
	} else {
		return nil
	}
}

func (user *User) Update(field ...string) error {
	if strings.TrimSpace(user.Name) == "" {
		return fmt.Errorf("user name can't be null")
	}
	//judge legitimacy of user name
	if err := JudgeNameLegitimacy(user.Name); err != nil {
		return err
	}

	//encrypt pwd
	if pwd, salt, err := generatePwdAndSalt(user.Password); err != nil {
		return fmt.Errorf("encrypt pwd error")
	} else {
		user.Password = pwd
		user.Salt = salt
	}

	u := &User{Name: user.Name}
	if exist, err := u.Get(); err != nil {
		return err
	} else if !exist {
		return fmt.Errorf(StatusNotFound+": user %s not exist", user.Name)
	} else if u.Status != ACTIVE {
		return fmt.Errorf(StatusBadRequest+": user %s  is not active", user.Name)
	}

	o := orm.NewOrm()
	field = append(field, "updated")
	if num, err := o.Update(user, field...); err != nil {
		return err
	} else if num <= 0 {
		return fmt.Errorf(StatusBadRequest + ": parameter is same with db")
	} else {
		return nil
	}
}

//List is to get user list
func (user *User) List() ([]User, error) {
	o := orm.NewOrm()
	userlist := []User{}
	if _, err := o.QueryTable("user").All(&userlist); err != nil {
		return nil, err
	}
	return userlist, nil
}

func (user *User) Delete() error {
	o := orm.NewOrm()
	if num, err := o.Delete(user); err != nil {
		return err
	} else if num <= 0 {
		return fmt.Errorf(StatusNotFound+": user %s not exist", user.Name)
	} else {
		return nil
	}
}

// Deactive updates all status relation to the user to inactive
func (user *User) Deactive() error {
	return DeactiveUser(user)
}

// Active updates all status relation to the user to active
func (user *User) Active() error {
	return ActiveUser(user)
}

func (user *User) activeUser() error {
	if exist, err := user.Get(); err != nil {
		return err
	} else if !exist {
		return fmt.Errorf(StatusNotFound+": user %s not exist", user.Name)
	} else if user.Status == ACTIVE {
		return fmt.Errorf(StatusBadRequest+": user %s  is already active", user.Name)
	}

	o := orm.NewOrm()
	user.Status = ACTIVE
	if num, err := o.Update(user, "status"); err != nil {
		return err
	} else if num <= 0 {
		return fmt.Errorf(StatusNotFound+": user %s not exist", user.Name)
	} else {
		return nil
	}
}

// Pseudo Random Number Generator
func PRNG(size int) ([]byte, error) {
	b := make([]byte, size)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func generatePwdAndSalt(pwd string) (string, string, error) {
	b, err := PRNG(16)
	if err != nil {
		return "", "", err
	}
	salt := fmt.Sprintf("%x", b)

	md5Pwd := fmt.Sprintf("%x", md5.Sum([]byte(pwd)))
	md5Salt := fmt.Sprintf("%x", md5.Sum([]byte(salt)))
	password := fmt.Sprintf("%x", md5.Sum([]byte(md5Pwd+md5Salt)))

	return password, salt, nil
}

func GeneratePwdBySalt(pwd, salt string) string {
	md5Pwd := fmt.Sprintf("%x", md5.Sum([]byte(pwd)))
	md5Salt := fmt.Sprintf("%x", md5.Sum([]byte(salt)))
	password := fmt.Sprintf("%x", md5.Sum([]byte(md5Pwd+md5Salt)))

	return password
}
