package model

import "github.com/Kamva/mgm/v2"

type HotTopicList struct {
	mgm.DefaultModel `bson:",inline"`
	Data             []HotTopic `json:"data" bson:"data"`
}

type HotTopic struct {
	Title string `json:"title" bson:"title"`
	Heat  int    `json:"heat" bson:"heat"`
	URL   string `json:"url" bson:"url"`
}
