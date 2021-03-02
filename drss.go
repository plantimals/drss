package drss

import (
	"context"
	"net/http"
	"time"

	"github.com/alecthomas/jsonschema"
	"github.com/ipfs/go-cid"
	shell "github.com/ipfs/go-ipfs-api"
	"github.com/ipfs/go-ipfs-api/options"
	"github.com/mmcdole/gofeed"
)

//DItem distributed item
type DItem struct {
	Item       *gofeed.Item  `json:"item,omitempty"`
	Enclosures []*DEnclosure `json:"enclosures,omitempty"`
}

//DEnclosure distributed enclosure
type DEnclosure struct {
	URL      string   `json:"url,omitempty"`
	FileType string   `json:"fileType"`
	File     *cid.Cid `json:"file"`
}

//DFeed distributed feed
type DFeed struct {
	DItems      []*DItem
	Feed        *gofeed.Feed
	Title       string      `json:"title"`
	Description string      `json:"description,omitempty"`
	Link        string      `json:"link,omitempty"`
	Updated     string      `json:"updated,omitempty"`
	Image       *DEnclosure `json:"image,omitempty"`
}

//ReadDFeed fetches and unmarshals a DFeed from a CID address
func ReadDFeed(feedCid string, s *shell.Shell) (*DFeed, error) {
	answer := &DFeed{}
	err := s.DagGet(feedCid, answer)
	if err != nil {
		return nil, err
	}
	return answer, nil
}

//CreateDFeedFromRSS takes an RSS/ATOM/JSON feed URL,
//downloads the contents, converts it to a DFeed, then
//pushes that feed into IPFS and returns a DFeed object
func CreateDFeedFromRSS(RSSURL string, s *shell.Shell) (*DFeed, error) {
	feed, err := GetRSSFeed(RSSURL)
	if err != nil {
		return nil, err
	}
	return CreateDFeed(feed, s)
}

//CreateDFeed takes a gofeed.Feed object, creates and
//uploads a DFeed to IPFS, and returns a DFeed object
func CreateDFeed(feed *gofeed.Feed, s *shell.Shell) (*DFeed, error) {

	dFeed := &DFeed{
		Title:       feed.Title,
		Description: feed.Description,
		Link:        feed.Link,
		Updated:     feed.Updated,
		Feed:        feed,
	}
	if feed.Image != nil && feed.Image.URL != "" {
		imageCID, err := storeFile(feed.Image.URL, s)
		if err != nil {
			panic(err)
		}
		dFeed.Image = &DEnclosure{
			URL:      feed.Image.URL,
			FileType: "image",
			File:     imageCID,
		}

	}
	for _, i := range feed.Items {
		dItem, err := CreateDItem(i, s)
		if err != nil {
			panic(err)
		}
		dFeed.DItems = append(dFeed.DItems, dItem)
	}
	return dFeed, nil
}

func storeFile(URL string, s *shell.Shell) (*cid.Cid, error) {
	resp, err := http.Get(URL)
	if err != nil {
		return nil, err
	}
	hash, err := s.Add(resp.Body)
	if err != nil {
		return nil, err
	}
	answer, err := cid.Decode(hash)
	if err != nil {
		return nil, err
	}
	return &answer, nil
}

//CreateDag pushes a json object into IPFS, returning a cid address
func CreateDag(json []byte, s *shell.Shell) *cid.Cid {
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

//CreateDItem accepts a gofeed.Item and turns it into a DItem
func CreateDItem(i *gofeed.Item, s *shell.Shell) (*DItem, error) {
	dItem := &DItem{
		Item: i,
	}
	for _, e := range i.Enclosures {
		dEnc, err := getHeavyEnclosure(e, s)
		if err != nil {
			return nil, err
		}
		dItem.Enclosures = append(dItem.Enclosures, dEnc)
	}
	return dItem, nil
}

func getHeavyEnclosure(e *gofeed.Enclosure, s *shell.Shell) (*DEnclosure, error) {
	cid, err := storeFile(e.URL, s)
	if err != nil {
		return nil, err
	}
	return &DEnclosure{
		URL:      e.URL,
		FileType: e.Type,
		File:     cid,
	}, nil
}

func GetJSONSchema() *jsonschema.Schema {
	return jsonschema.Reflect(&DFeed{})
}

func GetRSSFeed(url string) (*gofeed.Feed, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	fp := gofeed.NewParser()
	feed, err := fp.ParseURLWithContext(url, ctx)
	if err != nil {
		return nil, err
	}
	return feed, nil
}
