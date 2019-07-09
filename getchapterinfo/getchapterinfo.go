package getchapterinfo

import (
	"bufio"
	"database/sql"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

//book信息
type Bookinfo struct {
	Bookid         int    `db:"id"`
	Bookname       string `db:"booksName"`
	Boookauthor    string `db:"author"`
	Bookcahtpernum int    `db:"chapters"`
	Bookcomment    string `db:"summary"`
	chapterdone    int    `db:"chapterdone"`
	//Bookstorepath   string `db:"bookstorepath"`
}

//章节信息
type chapterinfo struct {
	Bookid      int    `db:"booksId"`
	Chapterid   int    `db:"chapterId"`
	Chaptername string `db:"chapterName"`
	Chapter     string `db:"content"`
	//Storepath string `db:"storepath"`
	Chapterlines int64 `db:"chapterlines"`
}

//文件名chan
var filenamech = make(chan string, 100)
var dbcon *sql.DB

func GetChapterInfo(db *sql.DB)  {
	dbcon = db
	//获取绝对路径
	pathname, err := filepath.Abs("/data/zip/unrarfull")
	if err != nil {
		fmt.Println("path error")
		return
	}
	//并发数量
	concurrenc := 20

	wg := sync.WaitGroup{} //控制主程序等待，以便goroutines运行完
	wg.Add(concurrenc + 1)
	//开启一个go 程去取文件名（全路径）
	go func(wg *sync.WaitGroup, filenamech chan string) {
		GetAllFile(pathname, filenamech)
		close(filenamech) //关闭通道，以便读取通道的程序知道通道已经关闭。
		wg.Done()         //一定在函数的内部的最后一行运行。否则可能函数没有执行完毕。
	}(&wg, filenamech)
	//for循环开启并发的go程
	for i := 0; i < concurrenc; i++ {
		go func(wg *sync.WaitGroup, filenamech chan string) {
			for { //for不断重chan中取文件路径
				filename, isclose := <-filenamech //判断chan是否关闭，关闭了就退出循环不在取文件名结束程序
				if !isclose {                     //判断通道是否关闭，关闭则退出循环
					break
				}
				//真正干事的
				dosomewrork(filename)
			}
			wg.Done()
		}(&wg, filenamech)
	}
	wg.Wait()
}

func dosomewrork(fp string) {
	//读取文本
	fi, err := os.Open(fp)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
	}
	defer fi.Close()

	br := bufio.NewReader(fi)
	var lines int64    //文本的行数
	var chapternum int //找到章节数，作为章节id
	var pre_chaptername string
	booksrt := Bookinfo{}
	for {
		a, _, c := br.ReadLine()
		if c == io.EOF {
			break
		}
		lines++
		//获取一个chapter结构实例，并存入数据库
		res := getchapterinfo(string(a), fp, lines)
		//去重匹配规则
		if res.Chaptername != "" && res.Bookid != 0 {
			//去重
			if quchong(pre_chaptername) == quchong(res.Chaptername) {
				continue
			}
			//匹配章节规则
			preisok, err := regexp.Match("^\\s*第.{1,9}章.*", []byte(pre_chaptername))
			if err != nil {
				fmt.Println(err)
			}
			nowisok, err := regexp.Match("^\\s*第.{1,9}章.*", []byte(res.Chaptername))
			if err != nil {
				fmt.Println(err)
			}
			if preisok {
				if preisok != nowisok {
					continue
				}
			}

			//匹配章节规则
			preisok1, err := regexp.Match("^\\s*\\d+\\s*", []byte(pre_chaptername))
			if err != nil {
				fmt.Println(err)
			}
			nowisok1, err := regexp.Match("^\\s*\\d+\\s*", []byte(res.Chaptername))
			if err != nil {
				fmt.Println(err)
			}
			if preisok1 {
				if preisok1 != nowisok1 {
					continue
				}
			}
			//
			preisok2, err := regexp.Match("^\\s*第.{1,9}节.*", []byte(pre_chaptername))
			if err != nil {
				fmt.Println(err)
			}
			nowisok2, err := regexp.Match("^\\s*第.{1,9}节.*", []byte(res.Chaptername))
			if err != nil {
				fmt.Println(err)
			}
			if preisok2 {
				if preisok2 != nowisok2 {
					continue
				}
			}

			//
			preisok3 , err := regexp.Match("^(\\s*)(卷.*)",[]byte(pre_chaptername))
			if err != nil {
				fmt.Println(err)
			}
			nowisok3 , err := regexp.Match("^(\\s*)(卷.*)",[]byte(res.Chaptername))
			if err != nil {
				fmt.Println(err)
			}
			if preisok3 {
				if preisok3 != nowisok3 {
					continue
				}
			}

			//章节id
			chapternum++
			res.Chapterid = chapternum
			fmt.Println(res.Chaptername)
			res.writetodb()

			//update 状态用
			b := &pre_chaptername
			*b = res.Chaptername
			booksrt.Bookid = res.Bookid
			booksrt.Bookcahtpernum = res.Chapterid
		}
	}

	//更新小说章节状态
	sqlupdate := fmt.Sprintf("UPDATE books set chapterdone=? WHERE id=?")
	stmt, err := dbcon.Prepare(sqlupdate)
	_, err = stmt.Exec(1, booksrt.Bookid)
	defer stmt.Close()
	if err != nil {
		fmt.Println(err)
	}

	//更新book 章节数
	sqlupdate = fmt.Sprintf("UPDATE books set chapters=? WHERE id=?")
	stmt, err = dbcon.Prepare(sqlupdate)
	_, err = stmt.Exec(booksrt.Bookcahtpernum, booksrt.Bookid)
	defer stmt.Close()
	if err != nil {
		fmt.Println(err)
	}
	return
}

func quchong(s string) string {
	isok , err := regexp.Match("^(\\s*)(第)(.{1,9})(章)(.*)",[]byte(s))
	if err != nil {
		fmt.Println(err)
	}

	if isok {
		reg := regexp.MustCompile("^(\\s*)(第.{1,9}章)(.*)$")
		result := reg.FindAllStringSubmatch(s,-1)
		a := result[0][2]
		return a
	}

	isok , err = regexp.Match("^(\\s*)(\\d+)(\\s*)",[]byte(s))
	if err != nil {
		fmt.Println(err)
	}

	if isok {
		reg := regexp.MustCompile("^(\\s*)(\\d+)(\\s*)")
		result := reg.FindAllStringSubmatch(s,-1)
		a := result[0][2]
		return a
	}

	isok , err = regexp.Match("^(\\s*)(卷.*)",[]byte(s))
	if err != nil {
		fmt.Println(err)
	}

	if isok {
		reg := regexp.MustCompile("^(\\s*)(卷.*)$")
		result := reg.FindAllStringSubmatch(s,-1)
		a := result[0][2]
		return a
	}

	isok , err = regexp.Match("^(\\s*)(第.{1,9}节)",[]byte(s))
	if err != nil {
		fmt.Println(err)
	}

	if isok {
		reg := regexp.MustCompile("^(\\s*)(第.{1,9}节)(.*)$")
		result := reg.FindAllStringSubmatch(s,-1)
		a := result[0][2]
		return a
	}

	return ""
}

func getchapterinfo(s string, fp string, lines int64) *chapterinfo {
	c := chapterinfo{}
	c.Bookid = getbookid(fp)
	c.Chaptername = getchaptername(s)
	c.Chapterlines = lines
	return &c
}

func getchaptername(s string)( a string ){
	isok , err := regexp.Match("^(\\s*)(第)(.{1,9})(章)(.*)",[]byte(s))
	if err != nil {
		fmt.Println(err)
	}

	if isok {
		reg := regexp.MustCompile("^(\\s*)(第.{1,9}章)(.*)$")
		result := reg.FindAllStringSubmatch(s,-1)
		a = result[0][0]
		//fmt.Println(a)
		return a
	}

	isok1 , err := regexp.Match("^(\\s*)(\\d+)(\\s*)",[]byte(s))
	if err != nil {
		fmt.Println(err)
	}

	if isok1 {
		reg := regexp.MustCompile("^(\\s*)(\\d+)(\\s*)")
		result := reg.FindAllStringSubmatch(s,-1)
		a = result[0][0]
		//fmt.Println(a)
		return a
	}

	isok2 , err := regexp.Match("^(\\s*)(卷.*)",[]byte(s))
	if err != nil {
		fmt.Println(err)
	}

	if isok2 {
		reg := regexp.MustCompile("^(\\s*)(卷.*)")
		result := reg.FindAllStringSubmatch(s,-1)
		a = result[0][0]
		//fmt.Println(a)
		return a
	}

	isok2 , err = regexp.Match("^(\\s*)(第.{1,9}节)",[]byte(s))
	if err != nil {
		fmt.Println(err)
	}

	if isok2 {
		reg := regexp.MustCompile("^(\\s*)(第.{1,9}节)(.*)")
		result := reg.FindAllStringSubmatch(s,-1)
		a = result[0][0]
		//fmt.Println(a)
		return a
	}

	return a
}

func getbookid(fp string) (bookid int) {
	db := dbcon
	bn := strings.Split(filepath.Base(fp), ".")
	bookname := bn[2]
	err := db.QueryRow("SELECT id FROM books WHERE booksName = ?", bookname).Scan(&bookid)
	check(err)
	return
}

func (c *chapterinfo) writetodb() {
	db := dbcon
	bookid := "chapter_" + strconv.Itoa(c.Bookid%100+1)
	insertsql := fmt.Sprintf("INSERT %v ( booksId, chapterName, content,chapterlines,chapterId) VALUES (?,?,?,?,?)", bookid)
	stmt, err := db.Prepare(insertsql)
	check(err)

	res, err := stmt.Exec(c.Bookid, c.Chaptername, c.Chapter, c.Chapterlines, c.Chapterid)
	check(err)
	defer stmt.Close()

	_, err = res.LastInsertId() //必须是自增id的才可以正确返回。
	check(err)
}

func (c *chapterinfo) printcahpter() {
	fmt.Println(c.Chaptername, c.Chapterlines)
}

func check(err error) {
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
}

func GetAllFile(pathname string, fn_ch chan string) {
	rd, err := ioutil.ReadDir(pathname)
	if err != nil {
		fmt.Println("read dir fail:", err)
	}
	for _, fi := range rd {
		if fi.IsDir() {
			fullDir := pathname + "/" + fi.Name()
			GetAllFile(fullDir, fn_ch)
			if err != nil {
				fmt.Println("read dir fail:", err)
			}
		} else {
			fullName := pathname + "/" + fi.Name()
			fn_ch <- fullName
		}
	}
}