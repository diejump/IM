package dao

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
)

var db *sql.DB

func Initmysql(username, passwd, host, port string) {
	var err error
	//dsn := "root:123@tcp(127.0.0.1:3306)/im?charset=utf8mb4&parseTime=True&loc=Local"
	dsn := username + ":" + passwd + "@tcp(" + host + ":" + port + ")/im?charset=utf8mb4&parseTime=True&loc=Local"

	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalln(err, "数据库连接失败")
	}

	err = db.Ping()
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("数据库连接成功")
	return
}

func AddUser(Account, Username string, Password []byte) bool {
	sqlstr := "insert into user (username,account,password) values (?,?,?)"
	_, err := db.Exec(sqlstr, Username, Account, Password)
	if err != nil {
		fmt.Printf("用户注册失败, err:%v\n", err)
		return false
	}
	log.Println("用户注册成功")
	return true
}

func SelectPasswordFromAccount(Account string) []byte {
	sqlstr := "select password from user where account=?"
	var password []byte
	db.QueryRow(sqlstr, Account).Scan(&password)
	return password
}

func FindUsernameFromAccount(account any) string {
	sqlstr := "select username from user where account = ?"
	var username string
	db.QueryRow(sqlstr, account).Scan(&username)
	return username
}

func GetAllUser() map[string]string {
	sqlStr := "select account,username from user"
	rows, err := db.Query(sqlStr)
	if err != nil {
		fmt.Printf("查找用户失败, err:%v\n", err)
		return nil
	}
	defer rows.Close()

	user := make(map[string]string)

	for rows.Next() {
		var a, b string
		err := rows.Scan(&a, &b)
		fmt.Println(a, b)
		if err != nil {
			fmt.Printf("读取失败, err:%v\n", err)
			return nil
		}
		user[a] = b
	}
	return user
}
