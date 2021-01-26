package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type Enclosure struct {
	url    string `json:"@url"`
	length string `json:"@length"`
	etype  string `json:"@type"`
}

type PodcastEpisode struct {
	title       string    `json:"title"`
	enclosure   Enclosure `json:"enclosure"`
	description string    `json:"description"`
	link        string    `json:"link"`
}

func main() {

	input, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	fmt.Println(string(input))
	var pce PodcastEpisode
	err = json.Unmarshal(input, &pce)
	if err != nil {
		panic(err)
	}
	fmt.Println(pce)
	pp, err := json.MarshalIndent(pce, "", "\t")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(pp))
}
