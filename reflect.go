package main

import (
	"time"
	"reflect"
	"log"
	"github.com/fatih/structs"
	"fmt"
	"strings"
)

type(
	IEntity interface {
		TableName() string
	}

	UserInfo struct {
		test 	string
		Uid int				`json:"uid" db:"uid,pk"`
		Username string		`json:"username" db:"username"`
		Department string	`json:"department" db:"department"`
		Created time.Time	`json:"created" db:"created"`
	}
)

func main() {
	log.SetFlags(log.Flags()|log.Lshortfile)

	var ui IEntity = &UserInfo{
		test:"ss",
		Uid:1,
		Username:"shenshouer",
		Department:"GT-基础平台",
		Created:time.Now(),
	}

	//rawReflect(ui)
	otherReflect(ui)
}

func(*UserInfo) TableName() string{
	return "userinfo"
}

func otherReflect(ui IEntity) {
	structName := structs.Name(ui)
	log.Println("structName", structName)

	fieldNames := structs.Names(ui)

	structInfo := structs.New(ui)

	values := make([]interface{}, 0)
	sqlFieldStr := ""
	sqlFieldPre := ""
	i := 0
	for _, fieldName := range fieldNames{
		field := structInfo.Field(fieldName)
		if field.IsExported(){
			tags := strings.Split(strings.Trim(field.Tag("db"), " "), ",")
			isPk := false
			if len(tags) == 2 && tags[1] == "pk"{
				isPk = true
			}
			tag := tags[0]
			if(!isPk){
				values = append(values, field.Value())
				if i == 0{
					sqlFieldStr += "("+strings.ToLower(tag)
					sqlFieldPre += "(?"
				}else{
					sqlFieldStr += ","+strings.ToLower(tag)
					sqlFieldPre += ",?"
				}
				i++
			}
		}
	}
	if (len(values) > 1){
		sqlFieldStr += ")"
		sqlFieldPre += ")"
	}

	sql := fmt.Sprintf("INSERT INTO %s %s VALUES %s", ui.TableName(), sqlFieldStr, sqlFieldPre)
	fmt.Println(sql)
	fmt.Println(values)
}

func rawReflect(ui IEntity){

//	val := reflect.Indirect(reflect.ValueOf(ui))
//
//	numField := val.NumField()
//
//	log.Println("numField:", numField, "type:", val.Type().Name())
//	for i := 0; i < numField; i++{
//		fieldValue := val.Field(i)
//		log.Println("fieldValue:", fieldValue, "field Type:",fieldValue.Type().Name(), "field tag", fieldValue.)
//	}

	val := reflect.ValueOf(ui).Elem()
	numField := val.NumField()

	log.Println("numField:", numField, "type:", val.Type().Name())
	for i := 0; i < numField; i++{
		fieldValue := val.Field(i)				// 获取 字段value的值
		fieldValueType := fieldValue.Type()		// 获取 字段value的类型
		fieldType := val.Type().Field(i)		// 获取 字段type
		log.Println("fieldValue:", fieldValue, "field Type:",fieldValueType.Name(),"fied name", fieldType.Name, "field tag:", fieldType.Tag)
	}
}