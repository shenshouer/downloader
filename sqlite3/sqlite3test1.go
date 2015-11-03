package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"time"
	"github.com/fatih/structs"
	"strings"
	"reflect"
)

type(
	IEntity interface{
		TableName() string
		Initialize() IEntity
	}

	UserInfo struct {
		Uid int64			`json:"uid" db:"uid,pk"`
		Username string		`json:"username" db:"username"`
		Department string	`json:"departname" db:"departname"`
		Created time.Time	`json:"created" db:"created"`
	}
)

var(
	db *sql.DB
	sql_create_userinfo = `
	CREATE TABLE IF NOT EXISTS userinfo (
		uid INTEGER PRIMARY KEY AUTOINCREMENT,
		username VARCHAR(64) NULL,
		departname VARCHAR(64) NULL,
		created DATE NULL
	);`
)

func init() {
	log.SetFlags(log.Flags()|log.Lshortfile)

	var err error
	db, err = sql.Open("sqlite3", "./sqlite3test1.db")
	if err != nil{
		log.Fatal(err)
	}

	// create tables
	if _, err := db.Exec(sql_create_userinfo); err != nil{
		log.Fatal(err)
	}
}

func main() {
	var id int64  // use for test update
	// insert
	ui := &UserInfo{
		Username: "astaxie",
		Department: "研发部门",
		Created: time.Now(),
	}
	if res, err := Save(db, ui);err != nil{
		log.Fatal(err)
	}else{
		id, err = res.LastInsertId()
		if err != nil{
			log.Fatal(err)
		}
		log.Println(id)
	}

	// query
	if entities, err := Query(db, new(UserInfo)); err != nil{
		log.Fatal(err)
	}else{
		for i, entity := range entities{
			log.Println(i, entity)
		}
	}

	ui = &UserInfo{
		Uid: id,
		Username: "shenshouer",
		Department: "GT-基础部门",
	}

	// update
	if res, err := Update(db, ui);err != nil{
		log.Fatal(err)
	}else{
		affect, err := res.RowsAffected()
		if err != nil{
			log.Fatal(err)
		}

		log.Println(affect)
	}

	// query
	if entities, err := Query(db, new(UserInfo)); err != nil{
		log.Fatal(err)
	}else{
		for i, entity := range entities{
			log.Println(i, entity)
		}
	}

	// delete
	if res, err := Delete(&UserInfo{Uid:id}); err != nil{
		log.Fatal(err)
	}else{
		affect, err := res.RowsAffected()
		if err != nil{
			log.Fatal(err)
		}

		log.Println(affect)
	}


	defer db.Close()

}

// TODO 根据已经设置值的字段删除相关记录
// 目前只支持主键删除
func Delete(entity IEntity)(sql.Result, error){
	v := reflect.ValueOf(entity)
	if destEntity, ok := v.Interface().(IEntity); ok { // 类型断言成 IEntity
		destEntityValue := reflect.ValueOf(destEntity)
		destEntityValueType := destEntityValue.Type()
		if destEntityValueType.Kind() != reflect.Ptr || destEntityValueType.Elem().Kind() != reflect.Struct {
			panic(fmt.Errorf("dest must be pointer to struct; got %T", destEntityValueType))
		}

		sql := ""
		distElem := v.Elem()
		numField := distElem.NumField()
		values := make([]interface{}, 0)
		for i:=0; i < numField; i++{
			tags := strings.Split(distElem.Type().Field(i).Tag.Get("db"), ",")
			if len(tags) == 2{
				sql = fmt.Sprintf("DELETE FROM %s WHERE %s=?", entity.TableName() ,tags[0])
				values = append(values, distElem.Field(i).Interface())
			}
		}

		log.Println(sql, values)
		stmt, err := db.Prepare(sql)
		if err != nil{
			return nil, err
		}

		return stmt.Exec(values...)
	}

	return nil, fmt.Errorf("entity must be implemete IEntity")
}

// TODO 根据已配置的字段查询相关记录
// 目前只支持根据主键查询
func Query(db *sql.DB, entity IEntity)([]IEntity, error){
	v := reflect.Zero(reflect.TypeOf(entity)) // 获取IEntity类型的零值指针
	if destEntity, ok := v.Interface().(IEntity); ok{ // 类型断言成 IEntity
		destEntityValueType := reflect.ValueOf(destEntity).Type()
		if destEntityValueType.Kind() != reflect.Ptr || destEntityValueType.Elem().Kind() != reflect.Struct {
			panic(fmt.Errorf("dest must be pointer to struct; got %T", destEntityValueType))
		}

		// 组装sql
		querySql := fmt.Sprintf("SELECT * FROM %s", entity.TableName())
		rows, err := db.Query(querySql)
		if err != nil{
			return nil, err
		}

		entities := make([]IEntity, 0)
		columns, err := rows.Columns()
		if err != nil{
			return nil, err
		}

		for rows.Next() {
			values := make([]interface{},0)
			tmp := destEntity.Initialize()		// 创建新的指针实体
			val := reflect.ValueOf(tmp).Elem()
			numFiled := val.NumField()
			for i := 0; i < numFiled; i++{
				fieldValue := val.Field(i)
				fieldType := val.Type().Field(i)
				tags := strings.Split(fieldType.Tag.Get("db"), ",") // 去掉主键tag pk标识
				if fieldValue.CanSet(){
					for _, columnName := range columns{
						if len(tags) >= 1 && tags[0] == columnName{
							values = append(values, fieldValue.Addr().Interface())
						}
					}
				}
			}
			err := rows.Scan(values...)
			if err != nil{
				return nil, err
			}
			entities = append(entities, tmp)
		}
		return entities, nil
	}

	err := fmt.Errorf("entity must implemetes interface IEntity ")
	return nil, err
}

func Update(db *sql.DB, entity IEntity) (sql.Result, error){
	structInfo := structs.New(entity)
	fieldNames := structs.Names(entity)

	whereClause := ""
	var idValue interface{}
	setValueClause := ""
	values := make([]interface{}, 0)
	i := 0
	for _, fieldName := range fieldNames {
		field := structInfo.Field(fieldName)
		if field.IsExported() { // 公有变量作为数据库字段
			isPk := false
			tags := strings.Split(strings.Trim(field.Tag("db"), " "), ",")
			if len(tags) == 2 && tags[1] == "pk"{
				isPk = true
			}
			tag := tags[0]
			if isPk{ // 更新时,只认主键作为更新条件
				whereClause = fmt.Sprintf("WHERE %s = ?", tag)
				idValue = field.Value()
			} else {
				if !field.IsZero(){ // 零值字段不更新
					if i == 0{
						setValueClause += fmt.Sprintf(" %s=? ", tag)
					}else{
						setValueClause += fmt.Sprintf(", %s=? ", tag)
					}
					values = append(values, field.Value())
					i++
				}
			}
		}
	}
	// 检查主键值
	if idValue == nil || idValue == 0{
		return nil, fmt.Errorf("the value of pk must be set")
	}

	values = append(values, idValue)
	updateSql := fmt.Sprintf("UPDATE %s SET %s %s", entity.TableName(), setValueClause, whereClause)

	stmt, err := db.Prepare(updateSql)
	if err != nil{
		return nil, err
	}

	res, err := stmt.Exec(values...)
	if err != nil{
		return nil, err
	}

	return res, nil
}

func Save(db *sql.DB, entity IEntity) (sql.Result, error){
	structInfo := structs.New(entity)
	fieldNames := structs.Names(entity)

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

	sql := fmt.Sprintf("INSERT INTO %s %s VALUES %s", entity.TableName(), sqlFieldStr, sqlFieldPre)
	stmt, err := db.Prepare(sql)
	if err != nil{
		return nil, err
	}
	res, err := stmt.Exec(values...)
	if err != nil{
		return nil, err
	}

	return res, nil
}

func (*UserInfo) TableName() string {
	return "userinfo"
}

// 初始化当前指针,并将实体所有属性赋予零值
func (ui *UserInfo) Initialize() IEntity {
	ui = &UserInfo{}
	return ui
}