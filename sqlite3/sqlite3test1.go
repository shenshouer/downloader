package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"time"
	"github.com/shenshouer/structs"
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
	typeRegistry = make(map[string]reflect.Type)	// 用于orm查询时返回对应的struct实体

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

	Registry(new(UserInfo))
	var err error
	db, err = sql.Open("sqlite3", "./sqlite3test1.db")
	if err != nil{
		log.Fatal(err)
	}

	// create tables
	if result, err := db.Exec(sql_create_userinfo); err != nil{
		log.Fatal(err)
	}else{
		log.Println(result)
	}
}

func main() {
	// insert
//	ui := &UserInfo{
//		Username: "astaxie",
//		Department: "研发部门",
//		Created: time.Now(),
//	}
//	res, err := Save(db, ui)
//	if err != nil{
//		log.Fatal(err)
//	}
//
//	id, err := res.LastInsertId()
//	if err != nil{
//		log.Fatal(err)
//	}
//
//	fmt.Println(id)
	// query
	entities, err := Query(db, new(UserInfo))
	log.Println(err)
	log.Println(entities)

//	ui = &UserInfo{
//		Uid: id,
//		Username: "shenshouer",
//		Department: "GT-基础部门",
//	}
//
//	// update
//	_, err = Update(db, ui)
//	if err != nil{
//		log.Fatal(err)
//	}
//
//	affect, err := res.RowsAffected()
//	if err != nil{
//		log.Fatal(err)
//	}
//
//	fmt.Println(affect)
//
//	// query
//	rows, err = db.Query("SELECT * FROM userinfo")
//	if err != nil{
//		log.Fatal(err)
//	}
//
//	for rows.Next() {
//		ui := &UserInfo{}
//		err = rows.Scan(&ui.Uid, &ui.Username, &ui.Department, &ui.Created)
//		if err != nil{
//			log.Fatal(err)
//		}
//		log.Println(ui)
//	}

	// delete
//	stmt, err = db.Prepare("delete from userinfo where uid=?")
//	if err != nil{
//		log.Fatal(err)
//	}
//
//	res, err = stmt.Exec(id)
//	if err != nil{
//		log.Fatal(err)
//	}
//
//	affect, err = res.RowsAffected()
//	if err != nil{
//		log.Fatal(err)
//	}
//
//	fmt.Println(affect)

	defer db.Close()

}

func Query(db *sql.DB, entity IEntity)([]IEntity, error){
	structName := structs.Name(entity)
	var registryStruct reflect.Type
	var ok bool
	if registryStruct, ok = typeRegistry[structName]; !ok{
		return nil, fmt.Errorf("%s can not be registried, pleace registry before used!", structName)
	}

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

	colNum := len(columns)
	var values = make([]interface{}, colNum)
	for i, _ := range values {
		var ii interface{}
		values[i] = &ii
	}

	v := reflect.Zero(registryStruct)
	for rows.Next() {
//		v := reflect.New(registryStruct).Elem()
		if tmp, ok := v.Interface().(IEntity); ok{
			tmp = tmp.Initialize()
			val := reflect.ValueOf(tmp).Elem()
//			tmpType := reflect.TypeOf(val)
			numFiled := val.NumField()
			for i := 0; i < numFiled; i++{
				fieldValue := val.Field(i)
				fieldValueType := fieldValue.Type()
				fieldType := val.Type().Field(i)
				log.Println(fieldValueType, fieldValue, fieldType.Tag)
			}

			entities = append(entities, tmp)
		}


//		log.Println(v.Kind())
//		log.Println("==>", tmp, ok)
//		log.Println(tmp)
//		s := structs.New(tmp)
//		err := rows.Scan(values...)
//		if err != nil{
//			return nil, err
//		}
//		for i, fieldName := range columns{
//			fields := s.Fields()
//			for _, field := range fields{
//				log.Println(fieldName,"field", field.Name(), "tag" , strings.Split(field.Tag("db"), ",")[0])
//				if strings.Split(field.Tag("db"), ",")[0] == fieldName{
//					log.Println(fieldName, strings.Split(field.Tag("db"), ",")[0], values[i])
//					field.Set(values[i])
//					log.Println("=====>> field.Value", field.Value())
//					break
//				}
//			}
//		}

//		log.Println(tmp)
//		entities = append(entities, tmp)
	}
	return entities, nil
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
//	return nil, nil
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
	log.Println(ui)
	return ui
}

// 注册需要用到ORM的实体
func Registry(v interface{}){
	if _, ok := v.(IEntity); !ok{
		panic("Registry struce must be implement the interface of IEntity")
	}
	structName := structs.Name(v)
	if _, ok := typeRegistry[structName]; ok{
		panic(fmt.Sprintf("%s has registried!"))
	}else{
		typeRegistry[structName] = reflect.TypeOf(v)
	}
}