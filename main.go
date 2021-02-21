package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"log"

	"github.com/ipfs/go-cid"
	shell "github.com/ipfs/go-ipfs-api"
	"github.com/ipfs/go-ipfs-api/options"
	"github.com/mmcdole/gofeed"
	ext "github.com/mmcdole/gofeed/extensions"
)

type Config struct {
	StoragePath string
	FeedURL     string
}

//from gofeed
type Person struct {
	Name  string `json:"name,omitempty"`
	Email string `json:"email,omitempty"`
}

//from gofeed
type Image struct {
	URL   string `json:"url,omitempty"`
	Title string `json:"title,omitempty"`
}

//from gofeed
type Enclosure struct {
	URL    string `json:"url,omitempty"`
	Length string `json:"length,omitempty"`
	Type   string `json:"type,omitempty"`
}

type IPItem struct {
	RSSItem    cid.Cid   `json:"rssitem"`
	Enclosures []cid.Cid `json:"enclosures,omitempty"`
}

type IPEnclosure struct {
	URL      string  `json:"url"`
	FileType string  `json:"fileType"`
	File     cid.Cid `json:"file"`
}

type IPFeed struct {
	Items           []cid.Cid
	Title           string                   `json:"title,omitempty"`
	Description     string                   `json:"description,omitempty"`
	Link            string                   `json:"link,omitempty"`
	FeedLink        string                   `json:"feedLink,omitempty"`
	Updated         string                   `json:"updated,omitempty"`
	Content         string                   `json:"content,omitempty"`
	UpdatedParsed   *time.Time               `json:"updatedParsed,omitempty"`
	Published       string                   `json:"published,omitempty"`
	PublishedParsed *time.Time               `json:"publishedParsed,omitempty"`
	Author          *Person                  `json:"author,omitempty"`
	Language        string                   `json:"language,omitempty"`
	Image           *Image                   `json:"image,omitempty"`
	Copyright       string                   `json:"copyright,omitempty"`
	Generator       string                   `json:"generator,omitempty"`
	Categories      []string                 `json:"categories,omitempty"`
	DublinCoreExt   *ext.DublinCoreExtension `json:"dcExt,omitempty"`
	ITunesExt       *ext.ITunesFeedExtension `json:"itunesExt,omitempty"`
	Extensions      ext.Extensions           `json:"extensions,omitempty"`
	Custom          map[string]string        `json:"custom,omitempty"`
	FeedType        string                   `json:"feedType,omitempty"`
	FeedVersion     string                   `json:"feedVersion,omitempty"`
}

func parseFlags() *Config {
	var storagePath string
	var feedURL string
	flag.StringVar(&storagePath, "storage", "./feed", "path to construct feed")
	flag.StringVar(&feedURL, "feedURL", "https://feeds.transistor.fm/the-vance-crowe-podcast", "feed URL")
	flag.Parse()

	_, err := url.ParseRequestURI(feedURL)
	if err != nil {
		panic(err)
	}
	return &Config{StoragePath: storagePath, FeedURL: feedURL}
}

func main() {
	config := parseFlags()
	rssToISS(config)
}

func rssToISS(config *Config) error {
	feed, err := getFeed(config.FeedURL)
	if err != nil {
		panic(err)
	}
	s := shell.NewShell("localhost:5001")

	var itemNodes []*cid.Cid
	for _, i := range feed.Items {
		itemNodes = append(itemNodes, getItemNode(i, s))
	}

	fs := getFeedNode(feed, itemNodes, s)
	feedj, err := fs.MarshalJSON()
	if err != nil {
		panic(err)
	}
	fmt.Println(string(feedj))
	return nil
}

func getItemNode(i *gofeed.Item, s *shell.Shell) *cid.Cid {
	ij, err := getItemJSON(i)
	if err != nil {
		panic(err)
	}
	return dagPut(ij, s)
}

func dagPut(json []byte, s *shell.Shell) *cid.Cid {
	ms, err := s.DagPutWithOpts(
		json,
		options.Dag.InputEnc("json"),
		options.Dag.Kind("cbor"),
		options.Dag.Hash("sha2-256"),
	)
	if err != nil {
		panic(err)
	}
	c, err := cid.Decode(ms)
	if err != nil {
		panic(err)
	}
	return &c
}

func getItemJSON(i *gofeed.Item) ([]byte, error) {
	answer, err := json.Marshal(i)
	if err != nil {
		return nil, err
	}
	return answer, nil
}

func getFeed(url string) (*gofeed.Feed, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	fp := gofeed.NewParser()
	feed, err := fp.ParseURLWithContext(url, ctx)
	if err != nil {
		return nil, err
	}
	fmt.Printf("found %v items\n", len(feed.Items))

	return feed, nil
}

func getFeedNode(f *gofeed.Feed, items []*cid.Cid, s *shell.Shell) *cid.Cid {
	ipf := IPFeed{
		//Links:       items,
		Title:           f.Title,
		Description:     f.Description,
		Link:            f.Link,
		FeedLink:        f.FeedLink,
		Updated:         f.Updated,
		UpdatedParsed:   f.UpdatedParsed,
		Published:       f.Published,
		PublishedParsed: f.PublishedParsed,
		Author:          (*Person)(f.Author),
		Language:        f.Language,
		Image:           (*Image)(f.Image),
		Copyright:       f.Copyright,
		Generator:       f.Generator,
		Categories:      f.Categories,
		DublinCoreExt:   f.DublinCoreExt,
		ITunesExt:       f.ITunesExt,
		Extensions:      f.Extensions,
		Custom:          f.Custom,
		FeedType:        f.FeedType,
		FeedVersion:     f.FeedVersion,
	}
	for _, cid := range items {
		ipf.Items = append(ipf.Items, *cid)
	}
	j, err := json.Marshal(ipf)
	if err != nil {
		panic(err)
	}
	return dagPut(j, s)
}

func addEpisode(episode *gofeed.Item, path string) {
	fmt.Printf("path=%s\n", path)
	err := os.Mkdir(path, 0700)
	if err != nil {
		panic(err)
	}

	//addText(path, "original-json", episode.Extensions)
	addText(path, "title", episode.Title)
	addText(path, "description", episode.Description)
	addText(path, "published", episode.Published)
	addText(path, "author", episode.Author.Name)
	addText(path, "link", episode.Link)
	addText(path, "summary", episode.Custom["itunes:summary"])
	addText(path, "explicit", episode.Custom["itunes:explicit"])
	addText(path, "keywords", episode.Custom["itunes:keywords"])
	addFile(path, "episode.mp3", episode.Enclosures[0].URL)
	addFile(path, "image.jpg", episode.Image.URL)
}

func addText(path string, key string, value string) {
	if len(value) == 0 {
		log.Print("no value for key: ", key)
		return
	}
	f, err := os.Create(fmt.Sprintf("%s/%s", path, key))
	if err != nil {
		panic(err)
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	_, err = w.WriteString(value)
	w.Flush()

}

func addFile(path string, key string, url string) {
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	out, err := os.Create(fmt.Sprintf("%s/%s", path, key))
	if err != nil {
		panic(err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		panic(err)
	}
}

func publishToIPFS(config *Config) {
	fmt.Println("unimplemented")
}
