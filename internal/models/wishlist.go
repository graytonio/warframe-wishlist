package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type WishlistItem struct {
	UniqueName string    `json:"uniqueName" bson:"uniqueName"`
	Quantity   int       `json:"quantity" bson:"quantity"`
	AddedAt    time.Time `json:"addedAt" bson:"addedAt"`
}

type Wishlist struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	UserID    string             `json:"userId" bson:"userId"`
	Items     []WishlistItem     `json:"items" bson:"items"`
	CreatedAt time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt time.Time          `json:"updatedAt" bson:"updatedAt"`
}

type AddItemRequest struct {
	UniqueName string `json:"uniqueName"`
	Quantity   int    `json:"quantity,omitempty"`
}

type UpdateQuantityRequest struct {
	Quantity int `json:"quantity"`
}

type MaterialRequirement struct {
	UniqueName  string `json:"uniqueName"`
	Name        string `json:"name"`
	TotalCount  int    `json:"totalCount"`
	ImageName   string `json:"imageName,omitempty"`
	Description string `json:"description,omitempty"`
}

type MaterialsResponse struct {
	Materials    []MaterialRequirement `json:"materials"`
	TotalCredits int                   `json:"totalCredits"`
}
