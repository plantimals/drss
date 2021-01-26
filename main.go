package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/mmcdole/gofeed"
)

type Config struct {
	StoragePath string
	FeedURL     string
}

func parseFlags() *Config {
	var storagePath string
	var feedURL string
	flag.StringVar(&storagePath, "storage", "./feed", "path to construct feed")
	flag.StringVar(&feedURL, "feedURL", "https://feeds.transistor.fm/the-vance-crowe-podcast", "feed URL")
	flag.Parse()

	/*s, err := os.Stat(storagePath)
	if err != nil {
		panic(err)
	}
	if s.IsDir() {
		panic(err)
	}*/

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
	fn, err := os.Stat(config.StoragePath)
	if fn != nil {
		panic(fmt.Errorf("there's already a directory at %s\n", config.StoragePath))
	}
	err = os.Mkdir(config.StoragePath, 0700)
	if err != nil {
		panic(err)
	}
	addEpisode(feed.Items[0], fmt.Sprintf("%s/%s", config.StoragePath, "0"))
	publishToIPFS(config)
	return nil
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
	addText(path, "title", episode.Title)
}

func addText(path string, key string, value string) {
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

}

func publishToIPFS(config *Config) {
	fmt.Println("unimplemented")
}
