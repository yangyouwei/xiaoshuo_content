package getcontent

import (
	"bufio"
	"database/sql"
	"fmt"
	"github.com/yangyouwei/xiaoshuo_content/read_conf"
	"os"
)

type booksinfo struct {
	Id int	`db:"id"`
	Sourcesfilename string `db:"Sourcesfilename"`
	Chapterdone int `db:"chapterdone"`
}

type chapter struct {
	BookId int `db:"booksId"`
	ChapterId int `db:"chapterId"`
	Content string `db:"content"`
	Size string `db:"size"`
	Chapterline int64 `db:"chapterlines"`
}

var booksinfo_ch = make(chan booksinfo)
var contentoffset []int

func GetContent(db *sql.DB)  {
	concurrent := read_conf.Main_str.Concurrent
	filespaht := read_conf.Main_str.Filepath

	getcontent()
}

func getbooks()  {

}

func getcontent(fp string)(content string) {
	for {
		star := 1
		end := 10
		fi, err := os.Open(fp)
		if err != nil {
			fmt.Printf("Error: %s\n", err)
			return ""
		}
		defer fi.Close()

		br := bufio.NewScanner(fi)
		for br.Scan(){
			if star == end{
				return br.Text()
			}
			star++
		}

	}
	return content
}

func replaceContent(s string) (rs string)  {

	return rs
}