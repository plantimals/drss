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
	"strings"
	"time"

	"log"

	shell "github.com/ipfs/go-ipfs-api"
	"github.com/mmcdole/gofeed"
)

type Config struct {
	StoragePath string
	FeedURL     string
}

type Link struct {
	addr string
}

func (l *Link) toJSON() string {
	return fmt.Sprintf("{\"/\":\"%s\"", l.addr)
}

type IPFeed struct {
	item string `json:"item"`
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

	var itemNodes []string
	//for _, i := range feed.Items {
	itemNodes = append(itemNodes, getItemNode(feed.Items[0], s))
	//itemNodes = append(itemNodes, getItemNode(feed.Items[1], s))
	//}
	fmt.Println(getFeedNode(itemNodes, s))
	return nil
}

func getFeedNode(f *gofeed.Feed, items []string, s *shell.Shell) string {

}

func getItemNode(i *gofeed.Item, s *shell.Shell) string {
	ib, err := getItemJSON(i)
	c, err := s.DagPut(ib, "json", "cbor")
	if err != nil {
		panic(err)
	}
	return strings.ToUpper(c)
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

	return feed, nil
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
