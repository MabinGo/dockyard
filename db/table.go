package db

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/jinzhu/gorm"
)

//
func (db *DB) RegisterModel(models ...interface{}) error {
	for _, model := range models {
		if !db.db.HasTable(model) {
			if err := db.db.CreateTable(model).Error; err != nil {
				return fmt.Errorf("create table error:" + err.Error())
			}
		}
	}
	return nil
}

func (db *DB) AddUniqueIndex(model interface{}, indexName string, columns ...string) error {
	if err := db.db.Model(model).AddUniqueIndex(indexName, columns...).Error; err != nil {
		return err
	}
	return nil
}

// create foreign key: CASCADE or RESTRICT
func (db *DB) AddForeignKey(model interface{}, foreignKeyField string, destinationTable string, onDelete string, onUpdate string) error {
	if err := db.db.Model(model).AddForeignKey(foreignKeyField, destinationTable, onDelete, onUpdate).Error; err != nil {
		return err
	}
	return nil
}

//Get get the nums of records
func (db *DB) Count(value interface{}) (int64, error) {
	var count int64
	if err := db.db.Where(value).Find(value).Count(&count).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, nil
		}
		return 0, err
	}
	return count, nil
}

//Create create new data
func (db *DB) Create(value interface{}) error {
	tx := db.db.Begin()
	if err := tx.Create(value).Error; err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

//Update update value of fields in database, if the value doesn't have primary key, will insert it
func (db *DB) Update(value interface{}) error {
	tx := db.db.Begin()
	if err := tx.Model(value).Updates(value).Error; err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

func (db *DB) UpdateField(model interface{}, field string, value interface{}) error {
	return db.db.Model(model).Update(field, value).Error
}

//Save update all value in database, if the value doesn't have primary key, will insert it
func (db *DB) Save(value interface{}) error {
	tx := db.db.Begin()
	if err := tx.Save(value).Error; err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

//delete
//When delete a record, you need to ensure it's primary field has value,
//and GORM will use the primary key to delete the record, if primary field's blank,
//GORM will delete all records for the model
func (db *DB) Delete(value interface{}) error {
	tx := db.db.Begin()
	if err := tx.Unscoped().Delete(value).Error; err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

//softdelete
//When delete a record, you need to ensure it's primary field has value,
//and GORM will use the primary key to delete the record, if primary field's blank,
//GORM will delete all records for the model
func (db *DB) DeleteS(value interface{}) error {
	tx := db.db.Begin()
	if err := tx.Delete(value).Error; err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

//batchdelete
//Delete mutiple values by the given condition
func (db *DB) BatchDelete(model interface{}, query string) error {
	tx := db.db.Begin()
	if err := tx.Unscoped().Where(query).Delete(model).Error; err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

//query multi record
func (db *DB) QueryM(condition interface{}, results interface{}) error {
	if err := db.db.Where(condition).Find(results).Error; err != nil {
		return err
	}
	return nil
}

//fuzzy query
func (db *DB) QueryF(condition interface{}, results interface{}) error {
	var (
		name  string
		value []interface{}
	)
	scope := db.db.NewScope(condition)
	for _, field := range scope.New(condition).Fields() {
		if !field.IsIgnored && !field.IsBlank {

			switch field.Field.Type().Kind() {
			case reflect.String:
				name = name + fmt.Sprintf("%s like ? AND ", field.DBName)
				//value = append(value, "%"+scope.AddToVars(field.Field.Interface())+"%")
				value = append(value, "%"+field.Field.Interface().(string)+"%")

			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
				reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
				reflect.Float32, reflect.Float64,
				reflect.Bool:
				name = name + fmt.Sprintf("%s = ? AND ", field.DBName)
				value = append(value, field.Field.Interface())
				/*
					Uintptr
					Complex64
					Complex128
					Array
					Chan
					Func
					Interface
					Map
					Ptr
					Slice
					Struct
					UnsafePointer
				*/
			default:
				panic("unsupport type")
			}

		}
	}
	name = strings.TrimRight(name, "AND ")

	if err := db.db.Where(name, value...).Find(results).Error; err != nil {
		return err
	}
	return nil
}

func (db *DB) Raw(models interface{}, sql string, values ...interface{}) error {
	tx := db.db.Begin()
	if err := tx.Raw(sql, values...).Scan(models).Error; err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

func (db *DB) Exec(sql string, values ...interface{}) *gorm.DB {
	return db.db.Raw(sql, values...)
}
