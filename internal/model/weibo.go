package model

import "github.com/Kamva/mgm/v2"

type WeiboHotTopics struct {
	mgm.DefaultModel `bson:",inline"`
	Time             int32        `json:"time" bson:"time"`
	Keywords         []Keyword    `json:"keywords" bson:"keywords"`
	Topics           []WeiboTopic `json:"topics" bson:"topics"`
}

type WeiboTopic struct {
	Heat     int32     `json:"heat" bson:"heat"`
	URL      string    `json:"url" bson:"url"`
	Title    string    `json:"title" bson:"title"`
	Keywords []Keyword `json:"keywords" bson:"keywords"`
}

type Keyword struct {
	Name   string  `json:"name" bson:"name"`
	Weight float64 `json:"weight" bson:"weight"`
	POS    string  `json:"pos" bson:"pos"`
}

func (z *WeiboHotTopics) CollectionName() string {
	return "weibo_hot_topics"
}
