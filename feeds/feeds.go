package feeds

import (
	"github.com/ipfs/go-cid"
	"github.com/mmcdole/gofeed"
)

type Config struct {
	StoragePath string
	FeedURL     string
}

type IPItem struct {
	Item       *gofeed.Item   `json:"item"`
	Enclosures []*IPEnclosure `json:"enclosures,omitempty"`
}

type IPEnclosure struct {
	URL      string  `json:"url"`
	FileType string  `json:"fileType"`
	File     cid.Cid `json:"file"`
}

type IPFeed struct {
	IPItems     []*IPItem
	Feed        *gofeed.Feed
	Title       string       `json:"title"`
	Description string       `json:"description,omitempty"`
	Link        string       `json:"link,omitempty"`
	Updated     string       `json:"updated,omitempty"`
	Image       *IPEnclosure `json:"image,omitempty"`
}
