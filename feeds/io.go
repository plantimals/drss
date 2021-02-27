package feeds

import (
	"net/http"

	"github.com/ipfs/go-cid"
	shell "github.com/ipfs/go-ipfs-api"
	"github.com/ipfs/go-ipfs-api/options"
	"github.com/mmcdole/gofeed"
)

func GetFeedNode(feedCid string, s *shell.Shell) (*IPFeed, error) {
	answer := &IPFeed{}
	err := s.DagGet(feedCid, answer)
	if err != nil {
		return nil, err
	}
	return answer, nil
}

func GetIPFeed(f *gofeed.Feed, ipItems []*IPItem, s *shell.Shell) (*IPFeed, error) {
	ipFeed := &IPFeed{
		Title:       f.Title,
		Description: f.Description,
		Link:        f.Link,
		Updated:     f.Updated,
	}
	if f.Image != nil {
		imageCID, err := storeFile(f.Image.URL, s)
		if err != nil {
			panic(err)
		}
		*ipFeed.Image = IPEnclosure{
			URL:      f.Image.URL,
			FileType: "image",
			File:     *imageCID,
		}

	}
	for _, ipItem := range ipItems {
		ipFeed.IPItems = append(ipFeed.IPItems, ipItem)
	}
	return ipFeed, nil
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

func GetIPItem(i *gofeed.Item, s *shell.Shell) (*IPItem, error) {

	ipItem := &IPItem{
		Item: i,
	}
	for _, e := range i.Enclosures {
		ipEnc, err := getHeavyEnclosure(e, s)
		if err != nil {
			return nil, err
		}
		ipItem.Enclosures = append(ipItem.Enclosures, ipEnc)
	}
	return ipItem, nil
}

func getHeavyEnclosure(e *gofeed.Enclosure, s *shell.Shell) (*IPEnclosure, error) {
	cid, err := storeFile(e.URL, s)
	if err != nil {
		return nil, err
	}
	return &IPEnclosure{
		URL:      e.URL,
		FileType: e.Type,
		File:     *cid,
	}, nil
}
