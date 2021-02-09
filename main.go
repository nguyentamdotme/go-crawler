package main

import (
	"strings"

	// "encoding/hex"
	"fmt"
	"go-module/database"
	"go-module/processXml"
	"io"
	"net/http"
	"os"
	// "regexp"
	// "math/rand"
	"time"
	// "md5"
	// "hex"

	"github.com/gocolly/colly"
	"github.com/gosimple/slug"
	"github.com/joho/godotenv"
	// "github.com/gocolly/colly/proxy"

)

type Post struct {
	id    string
	title string
}

func visitLink(urlSet processXml.Urlset) {
	var limit int
	limit = len(urlSet.Urls)
	limit = 1
	for i := 0; i < limit; i++ {
		var link string
		link = urlSet.Urls[i].Loc
		getPost(link)
		//if isPost := getPost(link); isPost != true {
		//	getCategory(link)
		//}
	}
}

func getPost(link string) bool {
	random := rand.Intn(5-1) + 1
	time.Sleep(time.Duration(random) * time.Second)
	fmt.Println("Start Crawl Post")
	c := colly.NewCollector(
		colly.AllowedDomains("tech12h.com"),
	)

	c.OnHTML("body.page-taxonomy", func(e *colly.HTMLElement) {
		var avatar string
		var title string
		var slugName string
		var content string
		// Lấy tiêu đề bài viết
		title = e.ChildText(".view-content h1")

		// Lấy thuộc tích src của ảnh đại diện bài viết dùng để download ảnh
		// avataUrl := e.ChildAttr("#content_left > div.post-head > figure > div > img", "src")

		// Tải ảnh avata
		avatar = downloadImage("https://tech12h.com/sites/all/themes/bartik/chu.png", title)
		// avata := "AVATAR"

		// Lấy content bài viết
		content = ""
		//e.ForEach(".view-content", func(_ int, m *colly.HTMLElement) {
		//	contentOrigin := regexp.MustCompile(`\n`)
		//	contentConverted := contentOrigin.ReplaceAllString(m.Text, "<br/>")
		//	content += "<p>" + contentConverted + "</p>"
		//})

		fmt.Printf("Crawling post type 1: %s \n", title)

		// Tạo slug name dựa trên tiêu đề bài viết
		slugName = slug.Make(title)
		fmt.Printf("Crawling SLUG: %s \n", slugName)

		category := e.ChildText(".duong_dan a")
		for j := 0; j <= len(category); j++ {
			fmt.Println(category[j])
			// fmt.Printf("Category %d is: %v", j, category[j])
		}

		// Lưu dữ liệu bài viết vào DB
		insertData(title, content, slugName, avatar)
		fmt.Printf("Inserted: %q\n", content)

	})

	c.OnHTML(".duong_dan a", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		fmt.Printf("Link found: %q -> %s\n", e.Text, link)
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Crawling... ", r.URL)
	})

	c.Visit(link)
	return true
}

func getCategory(link string) bool {
	random := rand.Intn(5-1) + 1
	time.Sleep(time.Duration(random) * time.Second)
	fmt.Printf("Start Crawl Post")
	c := colly.NewCollector(
		colly.AllowedDomains("tech12h.com"),
	)
	c.OnHTML(".content_left .content_term_1", func(e *colly.HTMLElement) {

		// Lấy tiêu đề bài viết
		title2 := e.ChildText(".h2_title")
		fmt.Println("Crawling post type 2... ", title2)

		// avata := downloadImage("https://tech12h.com/sites/all/themes/bartik/chu.png", title2)
		// avata := "AVATAR"

		// Lấy content bài viết
		content := ""
		e.ForEach(".nd_my", func(_ int, m *colly.HTMLElement) {
			contentOrigin := regexp.MustCompile(`\n`)
			contentConverted := contentOrigin.ReplaceAllString(m.Text, "<br/>")
			content += "<p>" + contentConverted + "</p>"
		})

		fmt.Printf("Crawling Title: %s \n", title2)

		// Tạo slug name dựa trên tiêu đề bài viết
		slugName := strings.TrimSpace(slug.Make(title2))
		fmt.Printf("Crawling SLUG: %s \n", slugName)

		category := strings.TrimSpace(e.ChildText(".duong_dan a:last-child"))
		fmt.Printf("Category: %s\n", category)

		// for j := 0; j <= len(category); j++ {
		// 	fmt.Printf("Category %d is: %q\n", j, category[j].Text)
		// }

		// // Lưu dữ liệu bài viết vào DB

		// var post Post
		// err := db.QueryRow("SELECT id, title FROM posts where slug = ?", slugName).Scan(&post.id, &post.title)
		// if err != nil {
		// 	println(err.Error())
		// 	// panic(err.Error())
		// 	insPost, err := db.Prepare("INSERT INTO posts (title, avata, content, slug, category) VALUES(?, ?, ?, ?, ?)")
		// 	handleError(err)
		// 	res, err := insPost.Exec(title2, "img/"+ avata + ".png", content, slugName, category)
		// 	lastId, err := res.LastInsertId()
		// 	fmt.Printf("Inserted: %d\n", lastId)
		// } else {
		// 	fmt.Printf("Post exited with ID : %s\n", slugName)

		// }
	})
	c.OnHTML(".duong_dan a", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		fmt.Printf("Link found: %q -> %s\n", e.Text, link)
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Crawling... ", r.URL)
	})

	c.Visit(link)
	return true
}

func insertData(title string, content string, slugName string, avatar string) {
	db := database.DBConn()
	insPost, err := db.Prepare("INSERT INTO posts (title, avata, content, slug) VALUES(?, ?, ?, ?)")
	handleError(err)
	insPost.Exec(title, "img/"+avatar+".png", content, slugName)
	defer db.Close()
}

func main() {
	godotenv.Load()
	links := processXml.ReadSiteMap("sitemap.xml")
	visitLink(links)
}

func handleError(e error) {
	if e != nil {
		panic(e)
	}
}

// Tải luôn hình đại diện bài viết

func downloadImage(src, title string) (image string) {
	// Tạo hasname cho ảnh tránh bị trùng lặp
	// md5HashInBytes := md5.Sum([]byte(title))
	// image = hex.EncodeToString(md5HashInBytes[:])
	img, _ := os.Create("img/" + image + ".jpg")
	defer img.Close()

	resp, _ := http.Get(src)
	defer resp.Body.Close()

	b, _ := io.Copy(img, resp.Body)
	fmt.Println("Saved image ! Size: ", b)

	return
}
