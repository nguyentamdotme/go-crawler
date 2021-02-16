package main

import (
	"strings"
	"sync"

	// "encoding/hex"
	"fmt"
	"go-module/database"
	"go-module/processXml"
	"io"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"time"

	// "md5"
	// "hex"
	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	"github.com/gosimple/slug"
	"github.com/joho/godotenv"
	// "github.com/gocolly/colly/proxy"
)

type Post struct {
	postType string
	title string
	alias string
	content string
	slugName string
	avatar	string
	category string
	categoryParent string
}

func main() {
	godotenv.Load()
	links := processXml.ReadSiteMap("sitemap.xml")
	visitLink(links)
}

func visitLink(urlSet processXml.Urlset) {
	wg := new(sync.WaitGroup)
	queueLink := make(chan string, 5)
	var limit int
	limit = len(urlSet.Urls)
	limit = 2
	for i := 0; i < limit; i++ {
		linkTemp := urlSet.Urls[i].Loc;
		fmt.Println("Queue Link: ", linkTemp)
		queueLink <- linkTemp
	}

	for i := 1; i<= 5; i++ {
		wg.Add(1)
		go fetchURL(wg, queueLink)
	}
	wg.Wait()
}

func fetchURL(wg *sync.WaitGroup, queueLink <-chan string) string {
	defer wg.Done()
	link := <- queueLink
	random := rand.Intn(5-1) + 1
	time.Sleep(time.Duration(random) * time.Second)
	fmt.Println("Start Crawl Post")
	c := colly.NewCollector(
		colly.AllowedDomains("tech12h.com"),
	)

	c.Limit(&colly.LimitRule{
		Parallelism: 2,
		RandomDelay: 3 * time.Second,
	})

	re := regexp.MustCompile("[^/]+$")
	alias := re.FindString(link);
	alias = strings.Replace(alias, ".html", "", 1)
	fmt.Println("Alias: ", alias)
	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Crawling... ", r.URL)
	})

	c.OnHTML("html body.page-node", func(e *colly.HTMLElement) {
		getPostPage(link, alias, e);
	})

	c.OnHTML("html body.page-taxonomy", func(e *colly.HTMLElement) {
		getCategoryPage(link, alias, e);
	})

	c.Visit(link)
	// <- link
	return alias;
}

func getPostPage(link string, alias string, e *colly.HTMLElement) bool {
	var avatar string
	var title string
	var slugName string
	querySelection := e.DOM

	title = querySelection.Find(".view-content h1:first-child").Text()
	slugName = slug.Make(title)

	avatar = getThumbnail(slugName, querySelection);

	content, err := querySelection.Find(".nd_1").Html()
	if(err != nil) {
		fmt.Println("Cant get Content at Link: ", link)
	}

	category, categoryParent := getCategory(querySelection)

	var postInfo = Post{
		postType: "POST",
		title: title,
		alias: alias,
		content: content,
		slugName: slugName,
		avatar: avatar,
		category: category,
		categoryParent: categoryParent}

	// Lưu dữ liệu bài viết vào DB
	insertData(postInfo)
	fmt.Printf("Inserted: %q\n", title)
	return true
}

func getCategoryPage(link string, alias string, e *colly.HTMLElement) bool {
	var avatar string
	var title string
	var slugName string
	// var content string
	querySelection := e.DOM

	title = querySelection.Find(".nd_my h2:first-child").Text()
	slugName = slug.Make(title)

	avatar = getThumbnail(slugName, querySelection);

	content, err := querySelection.Find(".nd_my").Html()
	if(err != nil) {
		fmt.Println("Cant get Content at Link: ", link)
	}

	category, categoryParent := getCategory(querySelection)

	var postInfo = Post{
		postType: "CATEGORY",
		title: title,
		alias: alias,
		content: content,
		slugName: slugName,
		avatar: avatar,
		category: category,
		categoryParent: categoryParent}

	// Lưu dữ liệu bài viết vào DB
	insertData(postInfo)
	fmt.Printf("Inserted: %q\n", title)
	return true
}

func getCategory(querySelection *goquery.Selection) (string, string) {
	re := regexp.MustCompile("[^/]+$")
	category := ""
	categoryParent := ""
	categories := querySelection.Find(".duong_dan a")
	categories.Each(func(i int, s *goquery.Selection) {
		linkOranic, _ := s.Attr("href");
		linkCat := re.FindString(linkOranic);
		if(linkCat != "") {
			if(i==1) {
				categoryParent = linkCat
			}

			if(i==2) {
				category = linkCat
			}
		}
	})
	return category, categoryParent
}

func getThumbnail(slugName string, querySelection *goquery.Selection) string {
	avatarURL := ""
	metaLink := querySelection.ParentsUntil("~").Find("meta")
	metaLink.Each(func(_ int, s *goquery.Selection) {
		property, _ := s.Attr("property")
		if strings.EqualFold(property, "og:image") {
			img, exist := s.Attr("content")
			fmt.Println("IMG:", img)
			if(exist) {
				avatarURL = img;
			}
		}

	})
	avatar := ""
	if(avatarURL != "") {
		avatar = downloadImage(avatarURL, slugName)
	}
	return avatar
}

func insertData(data Post) {
	db := database.DBConn()
	insPost, err := db.Prepare("INSERT INTO posts (post_type, title, alias, avatar, content, slug, category, category_parent) VALUES(?, ?, ?, ?, ?, ?, ?, ?)")
	handleError(err)
	insPost.Exec(data.postType, data.title, data.alias, data.avatar, data.content, data.slugName, data.category, data.categoryParent)
	defer db.Close()
}

func handleError(e error) {
	if e != nil {
		panic(e)
	}
}

func downloadImage(src, title string) (image string) {
	dir := "img/" + title + ".jpg";
	img, _ := os.Create(dir)
	defer img.Close()

	resp, _ := http.Get(src)
	defer resp.Body.Close()

	b, _ := io.Copy(img, resp.Body)
	fmt.Println("Saved image ! Size: ", b)

	return dir
}
