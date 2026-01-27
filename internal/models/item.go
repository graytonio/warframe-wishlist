package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Component struct {
	UniqueName   string `json:"uniqueName" bson:"uniqueName"`
	Name         string `json:"name" bson:"name"`
	ItemCount    int    `json:"itemCount" bson:"itemCount"`
	IsPrime      bool   `json:"isPrime,omitempty" bson:"isPrime,omitempty"`
	Description  string `json:"description,omitempty" bson:"description,omitempty"`
	ImageName    string `json:"imageName,omitempty" bson:"imageName,omitempty"`
	Tradable     bool   `json:"tradable,omitempty" bson:"tradable,omitempty"`
	Drops        []Drop `json:"drops,omitempty" bson:"drops,omitempty"`
}

type Drop struct {
	Location string  `json:"location" bson:"location"`
	Type     string  `json:"type" bson:"type"`
	Rarity   string  `json:"rarity,omitempty" bson:"rarity,omitempty"`
	Chance   float64 `json:"chance,omitempty" bson:"chance,omitempty"`
}

type Item struct {
	ID               primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	UniqueName       string             `json:"uniqueName" bson:"uniqueName"`
	Name             string             `json:"name" bson:"name"`
	Description      string             `json:"description,omitempty" bson:"description,omitempty"`
	Type             string             `json:"type,omitempty" bson:"type,omitempty"`
	Category         string             `json:"category,omitempty" bson:"category,omitempty"`
	ImageName        string             `json:"imageName,omitempty" bson:"imageName,omitempty"`
	Tradable         bool               `json:"tradable,omitempty" bson:"tradable,omitempty"`
	IsPrime          bool               `json:"isPrime,omitempty" bson:"isPrime,omitempty"`
	MasteryReq       int                `json:"masteryReq,omitempty" bson:"masteryReq,omitempty"`
	BuildPrice       int                `json:"buildPrice,omitempty" bson:"buildPrice,omitempty"`
	BuildTime        int                `json:"buildTime,omitempty" bson:"buildTime,omitempty"`
	SkipBuildTimePrice int              `json:"skipBuildTimePrice,omitempty" bson:"skipBuildTimePrice,omitempty"`
	BuildQuantity    int                `json:"buildQuantity,omitempty" bson:"buildQuantity,omitempty"`
	ConsumeOnBuild   bool               `json:"consumeOnBuild,omitempty" bson:"consumeOnBuild,omitempty"`
	Components       []Component        `json:"components,omitempty" bson:"components,omitempty"`
	Drops            []Drop             `json:"drops,omitempty" bson:"drops,omitempty"`
	WikiaThumbnail   string             `json:"wikiaThumbnail,omitempty" bson:"wikiaThumbnail,omitempty"`
	WikiaURL         string             `json:"wikiaUrl,omitempty" bson:"wikiaUrl,omitempty"`
	Collection       string             `json:"_collection,omitempty" bson:"_collection,omitempty"`
}

type ItemSearchResult struct {
	UniqueName  string `json:"uniqueName" bson:"uniqueName"`
	Name        string `json:"name" bson:"name"`
	Description string `json:"description,omitempty" bson:"description,omitempty"`
	Category    string `json:"category,omitempty" bson:"category,omitempty"`
	ImageName   string `json:"imageName,omitempty" bson:"imageName,omitempty"`
	Collection  string `json:"_collection,omitempty" bson:"_collection,omitempty"`
}

type SearchParams struct {
	Query    string
	Category string
	Limit    int
	Offset   int
}
