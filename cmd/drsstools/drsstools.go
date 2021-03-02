package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/url"

	"github.com/ipfs/go-cid"
	shell "github.com/ipfs/go-ipfs-api"
	drss "github.com/plantimals/drss"
)

type config struct {
	StoragePath  string
	FeedURL      string
	DumpSchema   bool
	DumpJSONFeed bool
}

func parseFlags() *config {
	var storagePath string
	var feedURL string
	var dumpSchema bool
	var dumpJSONFeed bool
	flag.StringVar(&storagePath, "storage", "./feed", "path to construct feed")
	flag.StringVar(&feedURL, "feedURL", "https://ipfs.io/blog/index.xml", "feed URL")
	flag.BoolVar(&dumpSchema, "schema", false, "dump the IPFeed jsonschema")
	flag.BoolVar(&dumpJSONFeed, "toJSON", false, "dump the contents of the provided URL in jsonfeed format to stdout")
	flag.Parse()

	_, err := url.ParseRequestURI(feedURL)
	if err != nil {
		panic(err)
	}
	return &config{
		StoragePath:  storagePath,
		FeedURL:      feedURL,
		DumpSchema:   dumpSchema,
		DumpJSONFeed: dumpJSONFeed,
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
	} else if config.DumpJSONFeed {
		feed, err := drss.GetRSSFeed(config.FeedURL)
		if err != nil {
			panic(err)
		}
		json, err := json.Marshal(feed)
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
	feed, err := drss.GetRSSFeed(config.FeedURL)
	if err != nil {
		panic(err)
	}
	s := shell.NewShell("localhost:5001")

	dFeed, err := drss.CreateDFeed(feed, s)
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
