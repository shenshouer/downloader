package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"time"
	"github.com/fatih/structs"
	"strings"
)

type(
	IEntity interface{
		TableName() string
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
	if result, err := db.Exec(sql_create_userinfo); err != nil{
		log.Fatal(err)
	}else{
		log.Println(result)
	}
}

func main() {
	// insert
	ui := &UserInfo{
		Username: "astaxie",
		Department: "研发部门",
		Created: time.Now(),
	}
	res, err := Save(db, ui)
	if err != nil{
		log.Fatal(err)
	}

	id, err := res.LastInsertId()
	if err != nil{
		log.Fatal(err)
	}

	fmt.Println(id)
	// query
	rows, err := db.Query("SELECT * FROM userinfo")
	if err != nil{
		log.Fatal(err)
	}

	for rows.Next() {
		ui := &UserInfo{}
		err = rows.Scan(&ui.Uid, &ui.Username, &ui.Department, &ui.Created)
		if err != nil{
			log.Fatal(err)
		}
		log.Println(ui)
	}

	ui = &UserInfo{
		Uid: id,
		Username: "shenshouer",
		Department: "GT-基础部门",
	}

	// update
	_, err = Update(db, ui)
	if err != nil{
		log.Fatal(err)
	}

//	affect, err := res.RowsAffected()
//	if err != nil{
//		log.Fatal(err)
//	}
//
//	fmt.Println(affect)

	// query
	rows, err = db.Query("SELECT * FROM userinfo")
	if err != nil{
		log.Fatal(err)
	}

	for rows.Next() {
		ui := &UserInfo{}
		err = rows.Scan(&ui.Uid, &ui.Username, &ui.Department, &ui.Created)
		if err != nil{
			log.Fatal(err)
		}
		log.Println(ui)
	}

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

	db.Close()

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

func (ui *UserInfo) TableName() string {
	return "userinfo"
}