package main

import (
	"context"
	"flag"
	"fmt"
	"net/url"
	"time"

	"github.com/ipfs/go-cid"
	shell "github.com/ipfs/go-ipfs-api"
	"github.com/mmcdole/gofeed"
	"github.com/plantimals/ipfsrss/feeds"
)

func parseFlags() *feeds.Config {
	var storagePath string
	var feedURL string
	flag.StringVar(&storagePath, "storage", "./feed", "path to construct feed")
	flag.StringVar(&feedURL, "feedURL", "https://feeds.transistor.fm/the-vance-crowe-podcast", "feed URL")
	flag.Parse()

	_, err := url.ParseRequestURI(feedURL)
	if err != nil {
		panic(err)
	}
	return &feeds.Config{StoragePath: storagePath, FeedURL: feedURL}
}

func main() {
	config := parseFlags()
	cid := rssToISS(config)
	json, err := cid.MarshalJSON()
	if err != nil {
		panic(err)
	}
	fmt.Println(string(json))
}

func rssToISS(config *feeds.Config) *cid.Cid {
	feed, err := getFeed(config.FeedURL)
	if err != nil {
		panic(err)
	}
	s := shell.NewShell("localhost:5001")

	var itemNodes []*cid.Cid
	for _, i := range feed.Items {
		itemNodes = append(itemNodes, feeds.GetItemNode(i, s))
	}
	return feeds.PutFeedNode(feed, itemNodes, s)
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
