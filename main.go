package main

import (
	"context"
	"encoding/json"
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
	var dumpSchema bool
	flag.StringVar(&storagePath, "storage", "./feed", "path to construct feed")
	flag.StringVar(&feedURL, "feedURL", "https://feeds.transistor.fm/the-vance-crowe-podcast", "feed URL")
	flag.BoolVar(&dumpSchema, "schema", false, "dump the IPFeed jsonschema")
	flag.Parse()

	_, err := url.ParseRequestURI(feedURL)
	if err != nil {
		panic(err)
	}
	return &feeds.Config{StoragePath: storagePath, FeedURL: feedURL, DumpSchema: dumpSchema}
}

func main() {
	config := parseFlags()
	if config.DumpSchema {
		schema := feeds.GetJSONSchema()
		json, err := schema.MarshalJSON()
		if err != nil {
			panic(err)
		}
		fmt.Println(string(json))
	} else {
		cid := rssToDRSS(config)
		json, err := cid.MarshalJSON()
		if err != nil {
			panic(err)
		}
		fmt.Println(string(json))
	}
}

func rssToDRSS(config *feeds.Config) *cid.Cid {
	feed, err := getFeed(config.FeedURL)
	if err != nil {
		panic(err)
	}
	s := shell.NewShell("localhost:5001")

	var IPItems []*feeds.IPItem
	for _, i := range feed.Items {
		ipItem, err := feeds.GetIPItem(i, s)
		if err != nil {
			panic(err)
		}
		IPItems = append(IPItems, ipItem)
	}
	ipFeed, err := feeds.GetIPFeed(feed, IPItems, s)
	if err != nil {
		panic(err)
	}
	ipFeedJSON, err := json.Marshal(ipFeed)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(ipFeedJSON))
	return feeds.DagPut(ipFeedJSON, s)
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
