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
	drss "github.com/plantimals/drss/feeds"
)

type config struct {
	StoragePath string
	FeedURL     string
	DumpSchema  bool
}

func parseFlags() *config {
	var storagePath string
	var feedURL string
	var dumpSchema bool
	flag.StringVar(&storagePath, "storage", "./feed", "path to construct feed")
	flag.StringVar(&feedURL, "feedURL", "https://ipfs.io/blog/index.xml", "feed URL")
	flag.BoolVar(&dumpSchema, "schema", false, "dump the IPFeed jsonschema")
	flag.Parse()

	_, err := url.ParseRequestURI(feedURL)
	if err != nil {
		panic(err)
	}
	return &config{
		StoragePath: storagePath,
		FeedURL:     feedURL,
		DumpSchema:  dumpSchema,
	}
}

func main() {
	config := parseFlags()
	if config.DumpSchema {
		schema := drss.GetJSONSchema()
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

func rssToDRSS(config *config) *cid.Cid {
	feed, err := getFeed(config.FeedURL)
	if err != nil {
		panic(err)
	}
	s := shell.NewShell("localhost:5001")

	var dItems []*drss.DItem
	for _, i := range feed.Items {
		dItem, err := drss.CreateDItem(i, s)
		if err != nil {
			panic(err)
		}
		dItems = append(dItems, dItem)
	}
	dFeed, err := drss.CreateDFeed(feed, dItems, s)
	if err != nil {
		panic(err)
	}
	dFeedJSON, err := json.Marshal(dFeed)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(dFeedJSON))
	return drss.CreateDag(dFeedJSON, s)
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
