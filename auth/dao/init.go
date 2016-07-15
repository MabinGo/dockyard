package dao

import (
	"database/sql"
	"fmt"

	"github.com/containerops/dockyard/utils/setting"
)

var setfkcmds = []string{
	"alter table organization_user_map add constraint fk_oum_user_name foreign key(user_name) references user(name) on delete cascade",
	"alter table organization_user_map add constraint fk_oum_org_name foreign key(org_name) references organization(name) on delete cascade",
	"alter table repository_ex add constraint fk_repo_org_name foreign key(org_name) references organization(name) on delete cascade",
	"alter table repository_ex add constraint fk_repo_user_name foreign key(user_name) references user(name) on delete cascade",
	"alter table team add constraint fk_team_org_name foreign key(org_name) references organization(name) on delete cascade",
	"alter table team_repository_map add constraint fk_trm_team_id  foreign key(team_id)  references  team(id) on delete  cascade",
	"alter table team_repository_map add constraint fk_trm_repo_id  foreign key(repo_id)  references  repository_ex(id) on delete  cascade",
	"alter table team_user_map add constraint fk_tum_team_id foreign key(team_id) references team(id) on delete cascade",
	"alter table team_user_map add constraint fk_tum_user_name foreign key(user_name) references organization_user_map(user_name) on delete cascade"}

var queryfkcmds = []string{
	"SELECT constraint_name FROM `information_schema`.`KEY_COLUMN_USAGE` where constraint_name='fk_oum_user_name'",
	"SELECT constraint_name FROM `information_schema`.`KEY_COLUMN_USAGE` where constraint_name='fk_oum_org_name'",
	"SELECT constraint_name FROM `information_schema`.`KEY_COLUMN_USAGE` where constraint_name='fk_repo_org_name'",
	"SELECT constraint_name FROM `information_schema`.`KEY_COLUMN_USAGE` where constraint_name='fk_repo_user_name'",
	"SELECT constraint_name FROM `information_schema`.`KEY_COLUMN_USAGE` where constraint_name='fk_team_org_name'",
	"SELECT constraint_name FROM `information_schema`.`KEY_COLUMN_USAGE` where constraint_name='fk_trm_team_id'",
	"SELECT constraint_name FROM `information_schema`.`KEY_COLUMN_USAGE` where constraint_name='fk_trm_repo_id'",
	"SELECT constraint_name FROM `information_schema`.`KEY_COLUMN_USAGE` where constraint_name='fk_tum_team_id'",
	"SELECT constraint_name FROM `information_schema`.`KEY_COLUMN_USAGE` where constraint_name='fk_tum_user_name'"}

func InitDAO() error {
	//use sql command to cascade tables in order to avoid the problem that orm method is unavailable
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8", setting.DBUser, setting.DBPasswd, setting.DBURI, setting.DBName)
	switch setting.DBDriver {
	case "mysql":
		db, err := sql.Open("mysql", dsn)
		if err != nil {
			return err
		}
		defer db.Close()

		if err := db.Ping(); err != nil {
			return err
		}

		var constraint_name string
		for i := 0; i < len(queryfkcmds); i++ {
			err := db.QueryRow(queryfkcmds[i]).Scan(&constraint_name)
			if err != nil {
				if _, err := db.Exec(setfkcmds[i]); err != nil {
					return err
				}
			}
		}

	default:
		//
	}

	u := &User{
		Name:     "root",
		Email:    "root@rootroot56789.com",
		Password: "root",
		Comment:  "administrator for dockyard system",
		Status:   0,
		Role:     SYSADMIN,
	}
	if exist, err := u.Get(); err != nil {
		return err
	} else if !exist {
		if err := u.Save(); err != nil {
			return err
		}
	}

	return nil
}
