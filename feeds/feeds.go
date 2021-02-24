package feeds

import (
	"time"

	"github.com/ipfs/go-cid"
	ext "github.com/mmcdole/gofeed/extensions"
)

type Config struct {
	StoragePath string
	FeedURL     string
}

//from gofeed
type Person struct {
	Name  string `json:"name,omitempty"`
	Email string `json:"email,omitempty"`
}

//from gofeed
type Image struct {
	URL   string `json:"url,omitempty"`
	Title string `json:"title,omitempty"`
}

//from gofeed
type Enclosure struct {
	URL    string `json:"url,omitempty"`
	Length string `json:"length,omitempty"`
	Type   string `json:"type,omitempty"`
}

type IPItem struct {
	RSSItem    cid.Cid   `json:"rssitem"`
	Enclosures []cid.Cid `json:"enclosures,omitempty"`
}

type IPEnclosure struct {
	URL      string  `json:"url"`
	FileType string  `json:"fileType"`
	File     cid.Cid `json:"file"`
}

type IPFeed struct {
	_id             string                   `json:"_id"`
	Items           []cid.Cid                `json:"items,omitempty"`
	Title           string                   `json:"title,omitempty"`
	Description     string                   `json:"description,omitempty"`
	Link            string                   `json:"link,omitempty"`
	FeedLink        string                   `json:"feedLink,omitempty"`
	Updated         string                   `json:"updated,omitempty"`
	Content         string                   `json:"content,omitempty"`
	UpdatedParsed   *time.Time               `json:"updatedParsed,omitempty"`
	Published       string                   `json:"published,omitempty"`
	PublishedParsed *time.Time               `json:"publishedParsed,omitempty"`
	Author          *Person                  `json:"author,omitempty"`
	Language        string                   `json:"language,omitempty"`
	Image           *Image                   `json:"image,omitempty"`
	Copyright       string                   `json:"copyright,omitempty"`
	Generator       string                   `json:"generator,omitempty"`
	Categories      []string                 `json:"categories,omitempty"`
	DublinCoreExt   *ext.DublinCoreExtension `json:"dcExt,omitempty"`
	ITunesExt       *ext.ITunesFeedExtension `json:"itunesExt,omitempty"`
	Extensions      ext.Extensions           `json:"extensions,omitempty"`
	Custom          map[string]string        `json:"custom,omitempty"`
	FeedType        string                   `json:"feedType,omitempty"`
	FeedVersion     string                   `json:"feedVersion,omitempty"`
}
