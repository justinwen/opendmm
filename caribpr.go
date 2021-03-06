package opendmm

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/deckarep/golang-set"
	"github.com/golang/glog"
	"github.com/junzh0u/httpx"
)

func caribprSearch(query string, wg *sync.WaitGroup, metach chan MovieMeta) {
	keywords := caribprGuess(query)
	for keyword := range keywords.Iter() {
		wg.Add(1)
		go func(keyword string) {
			defer wg.Done()
			caribprSearchKeyword(keyword, metach)
		}(keyword.(string))
	}
}

func caribprGuess(query string) mapset.Set {
	re := regexp.MustCompile("(\\d{6})_(\\d{3})")
	matches := re.FindAllStringSubmatch(query, -1)
	keywords := mapset.NewSet()
	for _, match := range matches {
		keywords.Add(fmt.Sprintf("%s_%s", match[1], match[2]))
	}
	return keywords
}

func caribprGuessFull(query string) mapset.Set {
	keywords := mapset.NewSet()
	for keyword := range caribprGuess(query).Iter() {
		keywords.Add(fmt.Sprintf("Caribpr %s", keyword))
	}
	return keywords
}

func caribprSearchKeyword(keyword string, metach chan MovieMeta) {
	glog.Info("Keyword: ", keyword)
	urlstr := fmt.Sprintf(
		"http://www.caribbeancompr.com/moviepages/%s/index.html",
		url.QueryEscape(keyword),
	)
	caribprParse(urlstr, keyword, metach)
}

func caribprParse(urlstr string, keyword string, metach chan MovieMeta) {
	glog.V(2).Info("Product page: ", urlstr)
	doc, err := newDocument(urlstr, httpx.GetContentInUTF8(http.Get))
	if err != nil {
		glog.V(2).Infof("Error parsing %s: %v", urlstr, err)
		return
	}

	var meta MovieMeta
	meta.Code = fmt.Sprintf("Caribpr %s", keyword)
	meta.Page = urlstr

	var urlbase *url.URL
	urlbase, err = url.Parse(urlstr)
	if err != nil {
		return
	}
	var urlcover *url.URL
	urlcover, err = urlbase.Parse("./images/l_l.jpg")
	if err == nil {
		meta.CoverImage = urlcover.String()
	}
	var urlthumbnail *url.URL
	urlthumbnail, err = urlbase.Parse("./images/main_b.jpg")
	if err == nil {
		meta.ThumbnailImage = urlthumbnail.String()
	}

	meta.Title = doc.Find("#main-content > div.main-content-movieinfo > div.video-detail").Text()
	meta.Description = doc.Find("#main-content > div.main-content-movieinfo > div.movie-comment").Text()
	doc.Find("#main-content > div.detail-content.detail-content-gallery > ul > li > div > a").Each(
		func(i int, a *goquery.Selection) {
			href, ok := a.Attr("href")
			if ok {
				if !strings.Contains(href, "/member/") {
					meta.SampleImages = append(meta.SampleImages, href)
				}
			}
		})

	doc.Find("#main-content > div.main-content-movieinfo > div.movie-info > dl").Each(
		func(i int, dl *goquery.Selection) {
			dt := dl.Find("dt")
			if strings.Contains(dt.Text(), "出演") {
				meta.Actresses = dl.Find("dd a").Map(
					func(i int, a *goquery.Selection) string {
						return a.Text()
					})
			} else if strings.Contains(dt.Text(), "カテゴリー") {
				meta.Categories = dl.Find("dd a").Map(
					func(i int, a *goquery.Selection) string {
						return a.Text()
					})
			} else if strings.Contains(dt.Text(), "販売日") {
				meta.ReleaseDate = dl.Find("dd").Text()
			} else if strings.Contains(dt.Text(), "再生時間") {
				meta.MovieLength = dl.Find("dd").Text()
			} else if strings.Contains(dt.Text(), "スタジオ") {
				meta.Maker = dl.Find("dd").Text()
			} else if strings.Contains(dt.Text(), "シリーズ") {
				meta.Series = dl.Find("dd").Text()
			}
		})

	metach <- meta
}
