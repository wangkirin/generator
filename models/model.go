package models

import (
	"errors"
	"reflect"

	"github.com/containerops/wrench/db"
)

//Save struct to db
func Save(obj interface{}) (err error) {

	refValue := reflect.ValueOf(obj)
	refType := reflect.TypeOf(obj)

	key := refValue.FieldByName("UUID")
	fieldCound := refType.NumField()

	for i := 0; i < fieldCound; i++ {
		objField := refType.Field(i)
		fieldName := objField.Name
		fieldValue := refValue.FieldByName(fieldName)
		if objField.Type.Kind() == reflect.String {
			db.Client.HSet(key.String(), fieldName, fieldValue.String())
		} else if objField.Type.Kind() == reflect.Int64 {
			db.Client.HSet(key.String(), fieldName, fieldValue.String())
		}
	}

	return nil
}

//Get struct from db
func Del(key string) (err error) {
	delCount, err := db.Client.Del(key).Result()
	if err != nil {
		return err
	}
	if delCount < 1 {
		return errors.New("can't del key")
	}
	return nil
}

//Get struct from db
func Get(key string, obj interface{}) (err error) {

	refValue := reflect.ValueOf(obj)
	refType := reflect.TypeOf(obj)

	fieldCound := refType.NumField()
	for i := 0; i < fieldCound; i++ {
		objField := refType.Field(i)
		fieldName := objField.Name
		isExist, err := db.Client.Exists(key).Result()

		if err != nil {
			return err
		}
		if !isExist {
			return errors.New("key not exist")
		}

		dbValue := db.Client.HGet(key, fieldName)
		if objField.Type.Kind() == reflect.String {
			v, e := dbValue.Result()
			if e != nil {
				return e
			}
			refValue.FieldByName(fieldName).SetString(v)
		} else if objField.Type.Kind() == reflect.Int64 {
			v, e := dbValue.Int64()
			if e != nil {
				return e
			}
			refValue.FieldByName(fieldName).SetInt(v)
		}
	}

	return nil
}
