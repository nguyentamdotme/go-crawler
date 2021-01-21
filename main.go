package main

import (
	// "crypto/md5"
	"database/sql"
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
    id   string
    title string
}

func visitLink(urlSet processXml.Urlset, db *sql.DB) {
	var limit int
	limit = len(urlSet.Urls)
	limit = 1;
	for i := 0; i < limit; i++ {
		// random := rand.Intn(5 - 1) + 1
		// time.Sleep(time.Duration(random) * time.Second)

		c := colly.NewCollector(
			colly.AllowedDomains("tech12h.com"),
		)
		c.Limit(&colly.LimitRule{
			Delay: 1 * time.Second,
			RandomDelay: 1 * time.Second,
		})

		// rp, err := proxy.RoundRobinProxySwitcher("socks5://93.91.112.247:41258")
		// if err != nil {
		// 	fmt.Println("error proxy")
		// }
		// c.SetProxyFunc(rp)

		// c.OnHTML("body", func(e *colly.HTMLElement) {

		// });
		fmt.Println("Start")
		c.OnHTML(".content_left.a_xanh", func(e *colly.HTMLElement) {

			// Lấy tiêu đề bài viết
			title := e.ChildText(".view-content h1")

			// Lấy thuộc tích src của ảnh đại diện bài viết dùng để download ảnh
			// avataUrl := e.ChildAttr("#content_left > div.post-head > figure > div > img", "src")

			// Tải ảnh avata
			avata := downloadImage("https://tech12h.com/sites/all/themes/bartik/chu.png", title)
			// avata := "AVATAR"

            // Lấy content bài viết
			// content := DOM.Html()
			content := e.ChildText(".view-content > .views-row").DOM.Html()
			fmt.Printf(content)
			// e.ForEach(".view-content > .views-row", func(_ int, m *colly.HTMLElement) {
			// 	contentOrigin := regexp.MustCompile(`\n`)
			// 	contentConverted := contentOrigin.ReplaceAllString(m.Text, "<br/>")
			// 	content += "<p>" + contentConverted + "</p>"
			// })

			fmt.Printf("Crawling post type 1: %s \n", title);

			// Tạo slug name dựa trên tiêu đề bài viết
			slugName := strings.TrimSpace(slug.Make(title))
			fmt.Printf("Crawling SLUG: %s \n", slugName);

			category := strings.TrimSpace(e.ChildText(".duong_dan a:last-child"))
			fmt.Printf("Category: %s\n", category)

			categoryParent := strings.TrimSpace(e.ChildText(".duong_dan a:nth-last-child(2)"))
			fmt.Printf("CategoryParent: %s\n", categoryParent)

			// Lưu dữ liệu bài viết vào DB
			insPost, err := db.Prepare("INSERT INTO posts (title, avata, content, slug, category, categoryParent) VALUES(?, ?, ?, ?, ?, ?)")
			handleError(err)
			res, err := insPost.Exec(title, "img/"+ avata + ".png", content, slugName, category, categoryParent)
			lastId, err := res.LastInsertId()
			fmt.Printf("Inserted ID: %d\n", lastId)

		})


		c.OnHTML(".content_left .content_term_1", func(e *colly.HTMLElement) {

			// Lấy tiêu đề bài viết
			title2 := e.ChildText(".h2_title")
			fmt.Println("Crawling post type 2... ", title2)

			avata := downloadImage("https://tech12h.com/sites/all/themes/bartik/chu.png", title2)
			// avata := "AVATAR"

            // Lấy content bài viết
			content := e.OnHTML(".nd_my", func(text *colly.HTMLElement) {
				fmt.Printf(text.DOM.Html())
			})
			// e.ForEach(".nd_my", func(_ int, m *colly.HTMLElement) {
			// 	contentOrigin := regexp.MustCompile(`\n`)
			// 	contentConverted := contentOrigin.ReplaceAllString(m.Text, "<br/>")
			// 	content += "<p>" + contentConverted + "</p>"
			// })

			fmt.Printf("Crawling Title: %s \n", title2);

			// Tạo slug name dựa trên tiêu đề bài viết
			slugName := strings.TrimSpace(slug.Make(title2))
			fmt.Printf("Crawling SLUG: %s \n", slugName);

			category := strings.TrimSpace(e.ChildText(".duong_dan a:last-child"))
			fmt.Printf("Category: %s\n", category)

			categoryParent := strings.TrimSpace(e.ChildText(".duong_dan a:nth-last-child(2)"))
			fmt.Printf("CategoryParent: %s\n", categoryParent)


			// e.ForEach(".duong_dan a", func(_ int, elem *colly.HTMLElement) {
			// 	fmt.Println(elem.Text)
			// })


			// // Lưu dữ liệu bài viết vào DB

			// var post Post
			// err := db.QueryRow("SELECT id, title FROM posts where slug = ?", slugName).Scan(&post.id, &post.title)
			// println("check DB");
			// if err != nil {
			// 	println(err.Error())
				insPost, err := db.Prepare("INSERT INTO posts (title, avata, content, slug, category, categoryParent) VALUES(?, ?, ?, ?, ?, ?)")
				handleError(err)
				res, err := insPost.Exec(title2, "img/"+ avata + ".png", content, slugName, category, categoryParent)
				lastId, err := res.LastInsertId()
				fmt.Printf("Inserted ID: %d\n", lastId)
			// } else {
			// 	fmt.Printf("Post exited with ID : %s\n", slugName)

			// }
		})

		// c.OnHTML(".duong_dan a", func(e *colly.HTMLElement) {
		// 	link := e.Attr("href")
		// 	fmt.Printf("Link found: %q -> %s\n", e.Text, link)
		// })


		c.OnRequest(func(r *colly.Request) {
			fmt.Println("Crawling... ", r.URL)
		})

		c.Visit(urlSet.Urls[i].Loc)
	}
}

func main() {
	godotenv.Load()
	db := database.DBConn()
	defer db.Close()
	links := processXml.ReadSiteMap("sitemap.xml")
	visitLink(links, db)
}


func handleError(e error) {
	if e != nil {
		panic(e)
	}
}

// Tải luôn hình đại diện bài viết

func downloadImage(src, title string) (image string){
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
