package main

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"fmt"
	"github.com/yangyouwei/xiaoshuo_content/getbookinfo"
	"github.com/yangyouwei/xiaoshuo_content/getchapterinfo"
	"github.com/yangyouwei/xiaoshuo_content/getcontent"
	"github.com/yangyouwei/xiaoshuo_content/read_conf"
	"log"
	"os"
)

var Db *sql.DB
var err error

//读取配置文件，初始化数据库连接
func init()  {
	var datasourcename string = read_conf.Mysql_conf_str.Username + ":" + read_conf.Mysql_conf_str.Password + "@tcp(" + read_conf.Mysql_conf_str.Ipaddress + ":" + read_conf.Mysql_conf_str.Port + ")/" + read_conf.Mysql_conf_str.DatabaseName
	Db, err = sql.Open("mysql", datasourcename)
	check(err)
}

func check(err error) {
	if err != nil {
		log.Panicln(err)
	}
}

//main
func main()  {
	workmode := read_conf.Main_str.Mode
	switch workmode {
	case "getbookinfo":
		getbookinfo.GetBookinfo(Db)
	case "getchapterinfo":
		getchapterinfo.GetChapterInfo(Db)
	case "getconent":
		getcontent.GetContent(Db)
	default:
		fmt.Println("workmod error.")
		os.Exit(1)
	}
}