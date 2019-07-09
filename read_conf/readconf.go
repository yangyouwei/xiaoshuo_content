package read_conf

import (
	"github.com/Unknwon/goconfig"
	"log"
	"path/filepath"
	"strconv"
)

type Mysql_conf struct {
	Username string
	Password string
	Ipaddress string
	Port	string
	DatabaseName	string
}

type mainS struct {
	Concurrent	int
	Mode string
	Filepath string
}

type BookinfoStr struct {
	Rules map[string]string
}

type ChapterIonfStr struct {
	Rules map[string]string
}

type ContentStr struct {
	Rules map[string]string
}
var Mysql_conf_str Mysql_conf
var Bookinfo_str BookinfoStr
var Chapterinfo_str ChapterIonfStr
var Content_str ContentStr
var Main_str mainS

func init()  {
	cfg, err := goconfig.LoadConfigFile("conf")
	if err != nil {
		log.Println("读取配置文件失败[config.ini]")
		panic(err)
	}

	Mysql_conf_str.Mysql_fun(cfg,err)
	Bookinfo_str.Bookinfo_fun(cfg,err)
	Chapterinfo_str.Chapter_fun(cfg,err)
	Content_str.Content_fun(cfg,err)
	Main_str.main_fun(cfg,err)
}

func (this *Mysql_conf)Mysql_fun(c *goconfig.ConfigFile,err error) {
	this.Username, err = c.GetValue("main", "username")
	if err != nil {
		log.Fatalf("无法获取键值（%s）：%s", "username", err)
		panic(err)
	}

	this.Password, err = c.GetValue("main", "password")
	if err != nil {
		log.Fatalf("无法获取键值（%s）：%s", "password", err)
		panic(err)
	}

	this.Ipaddress, err = c.GetValue("main", "addr")
	if err != nil {
		log.Fatalf("无法获取键值（%s）：%s", "addr", err)
		panic(err)
	}

	this.Port, err = c.GetValue("main", "port")
	if err != nil {
		log.Fatalf("无法获取键值（%s）：%s", "prot", err)
		panic(err)
	}

	this.DatabaseName, err = c.GetValue("main", "databasename")
	if err != nil {
		log.Fatalf("无法获取键值（%s）：%s", "databasename", err)
		panic(err)
	}
}

func (this *BookinfoStr)Bookinfo_fun(c *goconfig.ConfigFile,err error)  {
	this.Rules, err = c.GetSection("getbookinfo")
	if err != nil {
		log.Fatalf("无法获取键值section（%s）：%s", "getbookinfo", err)
		panic(err)
	}
}

func (this *ChapterIonfStr)Chapter_fun(c *goconfig.ConfigFile,err error)  {
	this.Rules, err = c.GetSection("getchapterinfo")
	if err != nil {
		log.Fatalf("无法获取键值section（%s）：%s", "getchapterinfo", err)
		panic(err)
	}
}

func (this *ContentStr)Content_fun(c *goconfig.ConfigFile,err error)  {
	this.Rules, err = c.GetSection("getcontent")
	if err != nil {
		log.Fatalf("无法获取键值section（%s）：%s", "getcontent", err)
		panic(err)
	}
}

func (this *mainS)main_fun(c *goconfig.ConfigFile,err error)  {
	n,err := c.GetValue("main","concurrent")
	if err != nil {
		log.Fatalf("无法获取键值section（%s）：%s", "concurrent", err)
		panic(err)
	}
	this.Concurrent,err = strconv.Atoi(n)
	if err != nil {
		log.Fatalf("%s）：%s,无效", "concurrent", err)
		panic(err)
	}

	a,err := c.GetValue("main","filepath")
	this.Filepath,err = filepath.Abs(a)
	if err != nil {
		log.Fatalf("%s）：%s,无效", "filepath", err)
		panic(err)
	}

	this.Mode,err = c.GetValue("main","mode")
}

