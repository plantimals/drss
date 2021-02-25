package feeds

import (
	"encoding/json"

	"github.com/ipfs/go-cid"
	shell "github.com/ipfs/go-ipfs-api"
	"github.com/ipfs/go-ipfs-api/options"
	"github.com/mmcdole/gofeed"
)

func GetFeedNode(feedCid string, s *shell.Shell) (*IPFeed, error) {
	answer := IPFeed{}
	err := s.DagGet(feedCid, answer)
	if err != nil {
		return nil, err
	}
	return &answer, nil
}

func PutFeedNode(f *gofeed.Feed, items []*cid.Cid, s *shell.Shell) *cid.Cid {
	ipf := IPFeed{
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
	return DagPut(j, s)
}

func DagPut(json []byte, s *shell.Shell) *cid.Cid {
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

func GetItemNode(i *gofeed.Item, s *shell.Shell) *cid.Cid {
	ij, err := getItemJSON(i)
	if err != nil {
		panic(err)
	}
	return DagPut(ij, s)
}

func getItemJSON(i *gofeed.Item) ([]byte, error) {
	answer, err := json.Marshal(i)
	if err != nil {
		return nil, err
	}
	return answer, nil
}
