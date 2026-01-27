package repository

import (
	"context"
	"time"

	"github.com/graytonio/warframe-wishlist/internal/database"
	"github.com/graytonio/warframe-wishlist/internal/models"
	"github.com/graytonio/warframe-wishlist/pkg/logger"
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
	logger.Debug(ctx, "repo: WishlistRepository.GetByUserID called", "userID", userID)

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	filter := bson.M{"userId": userID}
	var wishlist models.Wishlist

	err := r.collection.FindOne(ctx, filter).Decode(&wishlist)
	if err == mongo.ErrNoDocuments {
		logger.Debug(ctx, "repo: WishlistRepository.GetByUserID - no wishlist found for user")
		return nil, nil
	}
	if err != nil {
		logger.Error(ctx, "repo: WishlistRepository.GetByUserID - error querying database", "error", err)
		return nil, err
	}

	logger.Debug(ctx, "repo: WishlistRepository.GetByUserID - found wishlist", "itemCount", len(wishlist.Items))
	return &wishlist, nil
}

func (r *WishlistRepository) Create(ctx context.Context, wishlist *models.Wishlist) error {
	logger.Debug(ctx, "repo: WishlistRepository.Create called", "userID", wishlist.UserID)

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	wishlist.CreatedAt = time.Now()
	wishlist.UpdatedAt = time.Now()
	if wishlist.Items == nil {
		wishlist.Items = []models.WishlistItem{}
	}

	result, err := r.collection.InsertOne(ctx, wishlist)
	if err != nil {
		logger.Error(ctx, "repo: WishlistRepository.Create - error inserting wishlist", "error", err)
		return err
	}

	wishlist.ID = result.InsertedID.(primitive.ObjectID)
	logger.Info(ctx, "repo: WishlistRepository.Create - wishlist created", "wishlistID", wishlist.ID.Hex())
	return nil
}

func (r *WishlistRepository) AddItem(ctx context.Context, userID string, item models.WishlistItem) error {
	logger.Debug(ctx, "repo: WishlistRepository.AddItem called", "userID", userID, "uniqueName", item.UniqueName, "quantity", item.Quantity)

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	filter := bson.M{"userId": userID}
	update := bson.M{
		"$push": bson.M{"items": item},
		"$set":  bson.M{"updatedAt": time.Now()},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		logger.Error(ctx, "repo: WishlistRepository.AddItem - error updating wishlist", "error", err)
		return err
	}

	logger.Debug(ctx, "repo: WishlistRepository.AddItem - completed", "matchedCount", result.MatchedCount, "modifiedCount", result.ModifiedCount)
	return nil
}

func (r *WishlistRepository) RemoveItem(ctx context.Context, userID, uniqueName string) error {
	logger.Debug(ctx, "repo: WishlistRepository.RemoveItem called", "userID", userID, "uniqueName", uniqueName)

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	filter := bson.M{"userId": userID}
	update := bson.M{
		"$pull": bson.M{"items": bson.M{"uniqueName": uniqueName}},
		"$set":  bson.M{"updatedAt": time.Now()},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		logger.Error(ctx, "repo: WishlistRepository.RemoveItem - error updating wishlist", "error", err)
		return err
	}

	logger.Debug(ctx, "repo: WishlistRepository.RemoveItem - completed", "matchedCount", result.MatchedCount, "modifiedCount", result.ModifiedCount)
	return nil
}

func (r *WishlistRepository) UpdateItemQuantity(ctx context.Context, userID, uniqueName string, quantity int) error {
	logger.Debug(ctx, "repo: WishlistRepository.UpdateItemQuantity called", "userID", userID, "uniqueName", uniqueName, "quantity", quantity)

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

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		logger.Error(ctx, "repo: WishlistRepository.UpdateItemQuantity - error updating wishlist", "error", err)
		return err
	}

	logger.Debug(ctx, "repo: WishlistRepository.UpdateItemQuantity - completed", "matchedCount", result.MatchedCount, "modifiedCount", result.ModifiedCount)
	return nil
}

func (r *WishlistRepository) Upsert(ctx context.Context, wishlist *models.Wishlist) error {
	logger.Debug(ctx, "repo: WishlistRepository.Upsert called", "userID", wishlist.UserID, "itemCount", len(wishlist.Items))

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

	result, err := r.collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		logger.Error(ctx, "repo: WishlistRepository.Upsert - error upserting wishlist", "error", err)
		return err
	}

	logger.Debug(ctx, "repo: WishlistRepository.Upsert - completed", "matchedCount", result.MatchedCount, "modifiedCount", result.ModifiedCount, "upsertedCount", result.UpsertedCount)
	return nil
}
