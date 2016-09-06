package db

import (
	//"database/sql"
	"fmt"
	//"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

type DB struct {
	db *gorm.DB
}

// single instance
var Instance *DB

func init() {
	Instance = &DB{}
}

// connet to mysql db
func InitDB(driver, user, passwd, URI, databaseName string) error {
	//var dbName string
	if driver == "mysql" {
		//dbName = "mysql"
	} else {
		return fmt.Errorf("only spport mysql driver now")
	}
	/*
		//1. connet to mysql db
		//[username[:password]@][protocol[(address)]]/dbname[?param1=value1&...&paramN=valueN]
		dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=True&loc=Local", user, passwd, URI, dbName)
		mysqlDB, err := sql.Open(driver, dsn)
		if err != nil {
			return fmt.Errorf("open db error:" + err.Error())
		}
		defer mysqlDB.Close()

		//2. create dockyard db
		if _, err := mysqlDB.Exec("SHOW CREATE DATABASE " + databaseName); err != nil {
			if _, err = mysqlDB.Exec("CREATE DATABASE " + databaseName); err != nil {
				return err
			}
		}
	*/
	//3. change to dockyard db
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=True&loc=Local", user, passwd, URI, databaseName)
	dockyardDB, err := gorm.Open(driver, dsn)
	if err != nil {
		return fmt.Errorf("open db error:" + err.Error())
	}
	dockyardDB.LogMode(false)
	Instance.db = dockyardDB

	Instance.initConnectedPool()

	return nil
}

func (db *DB) initConnectedPool() {
	//8. set connect pool
	db.db.DB().SetMaxIdleConns(10)
	db.db.DB().SetMaxOpenConns(100)
}
