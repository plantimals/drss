package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type Enclosure struct {
	Url    string `json:"@url"`
	Length string `json:"@length"`
	Etype  string `json:"@type"`
}

type PodcastEpisode struct {
	Title       string `json:"title"`
	En          Enclosure
	Description string `json:"description"`
	Link        string `json:"link"`
}

func main() {
	input, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	var pce PodcastEpisode
	err = json.Unmarshal(input, &pce)
	if err != nil {
		panic(err)
	}
	pp, err := json.MarshalIndent(pce, "", "\t")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(pp))
}
