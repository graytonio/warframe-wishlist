package repository

import (
	"context"
	"time"

	"github.com/graytonio/warframe-wishlist/internal/database"
	"github.com/graytonio/warframe-wishlist/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const wishlistCollection = "wishlists"

type WishlistRepository struct {
	db         *database.MongoDB
	collection *mongo.Collection
}

func NewWishlistRepository(db *database.MongoDB) *WishlistRepository {
	return &WishlistRepository{
		db:         db,
		collection: db.Collection(wishlistCollection),
	}
}

func (r *WishlistRepository) GetByUserID(ctx context.Context, userID string) (*models.Wishlist, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	filter := bson.M{"userId": userID}
	var wishlist models.Wishlist

	err := r.collection.FindOne(ctx, filter).Decode(&wishlist)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &wishlist, nil
}

func (r *WishlistRepository) Create(ctx context.Context, wishlist *models.Wishlist) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	wishlist.CreatedAt = time.Now()
	wishlist.UpdatedAt = time.Now()
	if wishlist.Items == nil {
		wishlist.Items = []models.WishlistItem{}
	}

	result, err := r.collection.InsertOne(ctx, wishlist)
	if err != nil {
		return err
	}

	wishlist.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *WishlistRepository) AddItem(ctx context.Context, userID string, item models.WishlistItem) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	filter := bson.M{"userId": userID}
	update := bson.M{
		"$push": bson.M{"items": item},
		"$set":  bson.M{"updatedAt": time.Now()},
	}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

func (r *WishlistRepository) RemoveItem(ctx context.Context, userID, uniqueName string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	filter := bson.M{"userId": userID}
	update := bson.M{
		"$pull": bson.M{"items": bson.M{"uniqueName": uniqueName}},
		"$set":  bson.M{"updatedAt": time.Now()},
	}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

func (r *WishlistRepository) UpdateItemQuantity(ctx context.Context, userID, uniqueName string, quantity int) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	filter := bson.M{
		"userId":           userID,
		"items.uniqueName": uniqueName,
	}
	update := bson.M{
		"$set": bson.M{
			"items.$.quantity": quantity,
			"updatedAt":        time.Now(),
		},
	}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

func (r *WishlistRepository) Upsert(ctx context.Context, wishlist *models.Wishlist) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	filter := bson.M{"userId": wishlist.UserID}
	wishlist.UpdatedAt = time.Now()

	opts := options.Update().SetUpsert(true)
	update := bson.M{
		"$set": bson.M{
			"items":     wishlist.Items,
			"updatedAt": wishlist.UpdatedAt,
		},
		"$setOnInsert": bson.M{
			"userId":    wishlist.UserID,
			"createdAt": time.Now(),
		},
	}

	_, err := r.collection.UpdateOne(ctx, filter, update, opts)
	return err
}
