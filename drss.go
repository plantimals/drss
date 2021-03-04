package drss

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	b64 "encoding/base64"

	"github.com/alecthomas/jsonschema"
	"github.com/ipfs/go-cid"
	shell "github.com/ipfs/go-ipfs-api"
	"github.com/ipfs/go-ipfs-api/options"
	"github.com/mmcdole/gofeed"
)

//DFeedID is the sha2 256 hash of feed url, used to identify a feed
type DFeedID string

//DItem distributed item
type DItem struct {
	Item       *gofeed.Item  `json:"item,omitempty"`
	Enclosures []*DEnclosure `json:"enclosures,omitempty"`
}

//DEnclosure distributed enclosure
type DEnclosure struct {
	URL      string   `json:"url,omitempty"`
	FileType string   `json:"fileType"`
	File     *cid.Cid `json:"file,omitempty"`
}

//DFeed distributed feed
type DFeed struct {
	DItems       []*DItem
	Feed         *gofeed.Feed
	Title        string        `json:"title"`
	Description  string        `json:"description,omitempty"`
	Link         string        `json:"link,omitempty"`
	Updated      string        `json:"updated,omitempty"`
	Image        *DEnclosure   `json:"image,omitempty"`
	FeedID       DFeedID       `json:"dFeedID"`
	OriginalFile []*DEnclosure `json:"enclosures"`
}

//GetHash converts a feed URL to a DFeedID
func GetHash(URL string) DFeedID {
	hash := sha256.Sum256([]byte(URL))
	base := b64.RawURLEncoding.EncodeToString(hash[:])
	fmt.Printf("encoded URL: %s\n", base)
	return DFeedID(base)
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
func CreateDFeedFromRSS(RSSURL string, s *shell.Shell) (*cid.Cid, error) {
	feed, err := GetRSSFeed(RSSURL)
	if err != nil {
		return nil, err
	}
	dFeed, err := CreateDFeed(feed)
	if err != nil {
		return nil, err
	}
	return PushDFeedtoIPFS(dFeed, s)
}

func PushDFeedtoIPFS(dFeed *DFeed, s *shell.Shell) (*cid.Cid, error) {
	if dFeed.Feed.Image != nil && dFeed.Feed.Image.URL != "" {
		imageCID, err := storeFile(dFeed.Feed.Image.URL, s)
		if err != nil {
			panic(err)
		}
		dFeed.Image = &DEnclosure{
			URL:      dFeed.Feed.Image.URL,
			FileType: "image",
			File:     imageCID,
		}
	}

	originalFile, err := EncloseOriginalFile(dFeed, s)
	if err != nil {
		return nil, err
	}
	dFeed.OriginalFile = append(dFeed.OriginalFile, originalFile)

	for _, dItem := range dFeed.DItems {
		for _, dEnc := range dItem.Enclosures {
			cid, err := storeFile(dEnc.URL, s)
			if err != nil {
				return nil, err
			}
			dEnc.File = cid
		}
		if dItem.Item.Image != nil && dItem.Item.Image.URL != "" {
			imageCID, err := storeFile(dItem.Item.Image.URL, s)
			if err != nil {
				panic(err)
			}
			dFeed.Image = &DEnclosure{
				URL:      dFeed.Feed.Image.URL,
				FileType: "image",
				File:     imageCID,
			}
		}
	}

	jf, err := json.Marshal(dFeed)
	if err != nil {
		return nil, err
	}
	return CreateDag(jf, s)
}

func EncloseOriginalFile(dFeed *DFeed, s *shell.Shell) (*DEnclosure, error) {
	cid, err := storeFile(dFeed.Feed.FeedLink, s)
	if err != nil {
		return nil, err
	}
	return &DEnclosure{
		File:     cid,
		FileType: dFeed.Feed.FeedType,
		URL:      dFeed.Feed.FeedLink,
	}, nil
}

//CreateDFeed takes a gofeed.Feed object, creates and
//uploads a DFeed to IPFS, and returns a DFeed object
func CreateDFeed(feed *gofeed.Feed) (*DFeed, error) {
	dFeed := &DFeed{
		FeedID:      GetHash(feed.FeedLink),
		Title:       feed.Title,
		Description: feed.Description,
		Link:        feed.Link,
		Updated:     feed.Updated,
		Feed:        feed,
	}

	for _, i := range feed.Items {
		dItem, err := CreateDItem(i)
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
func CreateDag(json []byte, s *shell.Shell) (*cid.Cid, error) {
	ms, err := s.DagPutWithOpts(
		json,
		options.Dag.InputEnc("json"),
		options.Dag.Kind("cbor"),
		options.Dag.Hash("sha2-256"),
	)
	if err != nil {
		return nil, err
	}
	c, err := cid.Decode(ms)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

//CreateDItem accepts a gofeed.Item and turns it into a DItem
func CreateDItem(i *gofeed.Item) (*DItem, error) {
	dItem := &DItem{
		Item: i,
	}
	for _, e := range i.Enclosures {
		dEnc, err := CreateLightEnclosure(e)
		if err != nil {
			return nil, err
		}
		dItem.Enclosures = append(dItem.Enclosures, dEnc)
	}
	return dItem, nil
}

func CreateLightEnclosure(e *gofeed.Enclosure) (*DEnclosure, error) {
	return &DEnclosure{
		URL:      e.URL,
		FileType: e.Type,
	}, nil
}

func CreateHeavyFromLight(le *DEnclosure, s *shell.Shell) error {
	cid, err := storeFile(le.URL, s)
	if err != nil {
		return err
	}
	le.File = cid
	return nil
}

func CreateHeavyEnclosure(e *gofeed.Enclosure, s *shell.Shell) (*DEnclosure, error) {
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
