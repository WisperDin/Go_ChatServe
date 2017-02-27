// sql_Oper
package sql_Oper

import (
	"database/sql"
	"fmt"
	"time"
)

var db *sql.DB

func SetupDB() {
	var err error
	rootDbPwd := "root"
	connStr := "root:" + rootDbPwd + "@/mysql?charset=utf8&loc=Local&parseTime=true" //登录
	db, err = sql.Open("mysql", connStr)
	if err != nil {
		panic(err.Error())
	}
	err = db.Ping()
	if err != nil {
		panic(err.Error())
	}
	cr_db := "CREATE DATABASE IF NOT EXISTS qnearBE DEFAULT CHARSET utf8 COLLATE utf8_general_ci;"
	stmt, err := db.Prepare(cr_db)
	if err != nil {
		panic(err.Error())
	}
	_, err = stmt.Exec()
	if err != nil {
		fmt.Println(err.Error())
	}
	stmt.Close()
	grantSQL := "grant all on qnearBE.* to cstAdmin identified by 'cstDb4ever';"
	stmt, err = db.Prepare(grantSQL)
	if err != nil {
		panic(err.Error())
	}
	_, err = stmt.Exec()
	if err != nil {
		fmt.Println(err.Error())
	}
	stmt.Close()
	grantSQL = "grant all on qnearBE.* to cstAdmin@'' identified by 'cstDb4ever';"
	stmt, err = db.Prepare(grantSQL)
	if err != nil {
		panic(err.Error())
	}
	_, err = stmt.Exec()
	if err != nil {
		fmt.Println(err.Error())
	}
	stmt.Close()
	grantSQL = "grant all on qnearBE.* to cstAdmin@'localhost' identified by 'cstDb4ever';"
	stmt, err = db.Prepare(grantSQL)
	if err != nil {
		panic(err.Error())
	}
	_, err = stmt.Exec()
	if err != nil {
		fmt.Println(err.Error())
	}
	stmt.Close()
	grantSQL = "grant all on qnearBE.* to cstAdmin@'127.0.0.1' identified by 'cstDb4ever';"
	stmt, err = db.Prepare(grantSQL)
	if err != nil {
		panic(err.Error())
	}
	_, err = stmt.Exec()
	if err != nil {
		fmt.Println(err.Error())
	}
	stmt.Close()
	db.Close()
	dbPwd := "cstDb4ever"
	connStr = "cstAdmin:" + dbPwd + "@/qnearBE?charset=utf8&loc=Local&parseTime=true"
	db, err = sql.Open("mysql", connStr)
	if err != nil {
		panic(err.Error())
	}
	err = db.Ping()
	if err != nil {
		panic(err.Error())
	}

	cr_table0 := "create table if not exists t_msg(msg_id int auto_increment primary key, peer varchar(64),msg varchar(128),recvTime datetime not null default 0)"
	stmt, err = db.Prepare(cr_table0)
	if err != nil {
		panic(err.Error())
	}
	defer stmt.Close()
	_, err = stmt.Exec()
	if err != nil {
		fmt.Println(err.Error())
	}

	cr_table1 := "create table if not exists t_user(user_id int, name varchar(64))"
	stmt, err = db.Prepare(cr_table1)
	if err != nil {
		panic(err.Error())
	}
	defer stmt.Close()
	_, err = stmt.Exec()
	if err != nil {
		fmt.Println(err.Error())
	}
}
func ImpUser(userList map[string]int) {
	sql := "select * from t_user"
	stmt, err := db.Prepare(sql)
	if err != nil {
		panic(err.Error())
	}
	defer stmt.Close()
	if err != nil {
		panic(err.Error())
	}
	rows, err := stmt.Query()
	if err != nil {
		panic(err.Error())
	}
	defer rows.Close()
	name := ""
	id := -1
	//msgList := ""
	for rows.Next() {
		rows.Scan(&id, &name)
		userList[name] = id
		//msgList = msgList + recvTime.Format("15:04:05 ") + msg + "\r\n"
	}
}

//注册用户
func InitUser(arg_id int, arg_user string) {
	sql := "insert IGNORE into t_user(user_id,name) values(?,?)" //ignore 的作用是无则插入
	stmt, err := db.Prepare(sql)
	if err != nil {
		panic(err.Error())
	}
	defer stmt.Close()
	_, err = stmt.Exec(arg_id, arg_user)
	if err != nil {
		fmt.Println(err.Error())
	}
}

//保存数据
func keepMsg(arg_msg string, arg_peer string) {
	sql := "insert into t_msg(peer,msg) values(?,?)"
	stmt, err := db.Prepare(sql)
	if err != nil {
		panic(err.Error())
	}
	defer stmt.Close()
	_, err = stmt.Exec(arg_peer, arg_msg)
	if err != nil {
		fmt.Println(err.Error())
	}
}
func queryMsg(arg_peer string) string {
	sql := "select msg,recvTime from t_msg"
	stmt, err := db.Prepare(sql)
	if err != nil {
		panic(err.Error())
	}
	defer stmt.Close()
	if err != nil {
		panic(err.Error())
	}
	rows, err := stmt.Query()
	if err != nil {
		panic(err.Error())
	}
	defer rows.Close()
	msg := ""
	recvTime := time.Now()
	msgList := ""

	for rows.Next() {
		rows.Scan(&msg, &recvTime)
		msgList = msgList + recvTime.Format("15:04:05 ") + msg + "\r\n"
	}
	return msgList
}
