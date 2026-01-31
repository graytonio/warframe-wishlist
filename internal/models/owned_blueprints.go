package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type OwnedBlueprint struct {
	UniqueName string    `json:"uniqueName" bson:"uniqueName"`
	AddedAt    time.Time `json:"addedAt" bson:"addedAt"`
}

type OwnedBlueprints struct {
	ID         primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	UserID     string             `json:"userId" bson:"userId"`
	Blueprints []OwnedBlueprint   `json:"blueprints" bson:"blueprints"`
	CreatedAt  time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt  time.Time          `json:"updatedAt" bson:"updatedAt"`
}

type AddBlueprintRequest struct {
	UniqueName string `json:"uniqueName"`
}

type BulkAddBlueprintsRequest struct {
	UniqueNames []string `json:"uniqueNames"`
}
