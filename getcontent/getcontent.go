package getcontent

import (
	"bufio"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"fmt"
	"github.com/yangyouwei/xiaoshuo_content/read_conf"
	"io"
	"strings"
	"log"
	"os"
	"regexp"
	"sync"
)

type booksinfo struct {
	Id              int    `db:"id"`
	Sourcesfilename string `db:"Sourcesfilename"`
	Chapterdone     int    `db:"chapterdone"`
}

type chapter struct {
	Id int64 `db:"id"`
	BookId      int    `db:"booksId"`
	ChapterId   int    `db:"chapterId"`
	Content     string `db:"content"`
	Size        string `db:"size"`
	Chapterline int64  `db:"chapterlines"`
	start       int64
	end         int64
}

type LineOffsetstr struct {
	start int
	end   int
}

var fullContent []string
var bookinfos = make(chan booksinfo ,100)
//var chapterContent []string
var bookId []int

func GetContent(dbc *sql.DB) {
	//并发
	c := read_conf.Main_str.Concurrent
	wg := sync.WaitGroup{}
	wg.Add(c+1)
	go getbookinfs(dbc, bookinfos,&wg)

	for i := 0; i < c; i++ {
		go func(wg *sync.WaitGroup) {
			for {
				b, isclose := <-bookinfos //判断chan是否关闭，关闭了就退出循环不在取文件名结束程序
				if !isclose {                     //判断通道是否关闭，关闭则退出循环
					return
				}
				//获取一本书籍信息
				//小说全部内容
				var fc *[]string
				fc = readfullcontent(b.Sourcesfilename)
				var chapterInfo *[]chapter
				//获取该本小说的全部章节信息，并更新章节start  end 行数
				chapterInfo = getchapterinfo(dbc, b)
				chapterInfo = dooffset(chapterInfo)
				for _, k := range *chapterInfo {
					//取出章节内容写入数据库
					updatechapter(dbc, k, fc)
				}
			}
			wg.Done()
		}(&wg)
	}
	wg.Wait()
}

func getbookinfs(dbc *sql.DB, c chan booksinfo,wg *sync.WaitGroup) {
	//查询总数
	n := 0
	sqltext := "select id from books order by id DESC limit 1;"
	err := dbc.QueryRow(sqltext).Scan(&n)
	if err != nil {
		panic(err)
	}
	fmt.Println("booksnum: ",n)

	for i:= 1;i <= n ;i++ {
		booksql := fmt.Sprintf("SELECT id, Sourcesfilename FROM books WHERE id=%v",i)
		res, err := dbc.Query(booksql)
		if err != nil {
			fmt.Println(err)
			continue
		}
		for res.Next() {
			//注意这里的Scan括号中的参数顺序，和 SELECT 的字段顺序要保持一致。
			a := booksinfo{}
			if err := res.Scan(&a.Id, &a.Sourcesfilename); err != nil {
				log.Fatal(err)
			}
			c <- a
		}
	}
	close(c)
	wg.Done()
}

func getchapterinfo(dbc *sql.DB, book booksinfo)  *[]chapter {
	var chinfo []chapter
	chaptersql := fmt.Sprintf("SELECT id,booksId,content,chapterlines FROM chapter_%v WHERE booksId=%v",book.Id%100+1,book.Id)
	rows, err := dbc.Query(chaptersql)
	if err != nil {
		panic(err)
	}
	c := chapter{}
	for  rows.Next()  {
		if err := rows.Scan(&c.Id,&c.BookId,&c.Content,&c.Chapterline); err != nil {
			log.Fatal(err)
		}
		chinfo = append(chinfo,c)
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
	rows.Close()
	return  &chinfo
}

func dooffset(c *[]chapter) *[]chapter  {
	num := len(*c)
	a := *c
	for n,v := range *c  {
		if n == num - 1 {
			a[n].start = v.Chapterline
			a[n].end = 0
			return &a
		}
		a[n].start = v.Chapterline
		a[n].end = a[n+1].Chapterline - 1
	}
	return &a
}

func readfullcontent(fp string) *[]string {
	fi, err := os.Open(fp)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
	}
	defer fi.Close()

	br := bufio.NewReader(fi)
	var tmp []string
	for {
		a, _, c := br.ReadLine()
		if c == io.EOF {
			break
		}
		//fmt.Println(string(a))
		tmp = append(tmp,string(a))
		//fmt.Println(cap(tmp))
	}
	return &tmp
}

//取出章节内容合并，更新章节表的内容。
func updatechapter(dbc *sql.DB, c chapter, fc *[]string){
	fmt.Println(c)
	cs := *fc
	var a []string
	if c.end == 0 {
		a = cs[c.start:]
	}else {
		a = cs[c.start:c.end]
	}
	
	var content string = "&nbsp&nbsp&nbsp&nbsp"
	for _,v := range a {
		if len(v) == 0 {
			continue
		}

		isok , err := regexp.Match(`^(\s+)$`,[]byte(v))
		if err != nil {
			fmt.Println(err)
		}
		if isok {
			continue
		}

		isok1 := strings.HasPrefix(v,"更多精彩，更多好书，尽在新奇书网—http://www.xqishu.com")
		if isok1 {
			isok , err := regexp.Match(`^(.*)(更多精彩，更多好书，尽在新奇书网—http://www.xqishu.com)$`,[]byte(v))
			if err != nil {
				fmt.Println(err)
			}
			if isok {
				continue
			}
			reg := regexp.MustCompile(`^(.*)(更多精彩，更多好书，尽在新奇书网—http://www.xqishu.com)(.+$)`)
			result := reg.FindAllStringSubmatch(v,-1)
			v = result[0][3]
		}

		content = content + v +"</br></br>"
	}
	//替换标签
	//fmt.Printf("chapterID: %v  start: %v end:  %v \n",c.Id,c.start,c.end)
	replacecharacter(&content)
	//fmt.Println(content)

	//写入数据库
	//fmt.Println(c.BookId)
	sqlupdate(dbc,c,content)
}

func sqlupdate(dbcon *sql.DB,c chapter,content string)  {
	contentsql := fmt.Sprintf("UPDATE chapter_%v SET content=? WHERE id=?",c.BookId%100+1)
	stmt, err := dbcon.Prepare(contentsql)
	if err != nil {
		log.Println(err)
	}
	_, err = stmt.Exec(content,c.Id)
	defer stmt.Close()
	if err != nil {
		log.Println(err)
	}
}


func replacecharacter(s *string) *string {
	//去行首空白字符
	isok , err := regexp.Match(`^(.+)(\s+)(.+)$`,[]byte(*s))
	if err != nil {
		fmt.Println(err)
	}
	if isok {
		reg := regexp.MustCompile(`^(.+)(\s+)(.+$)`)
		result := reg.FindAllStringSubmatch(*s,-1)
		*s = result[0][1]+result[0][3]
	}
	return s
}
