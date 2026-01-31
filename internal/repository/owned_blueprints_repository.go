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

const ownedBlueprintsCollection = "owned_blueprints"

type OwnedBlueprintsRepository struct {
	db         *database.MongoDB
	collection *mongo.Collection
}

func NewOwnedBlueprintsRepository(db *database.MongoDB) *OwnedBlueprintsRepository {
	return &OwnedBlueprintsRepository{
		db:         db,
		collection: db.Collection(ownedBlueprintsCollection),
	}
}

func (r *OwnedBlueprintsRepository) GetByUserID(ctx context.Context, userID string) (*models.OwnedBlueprints, error) {
	logger.Debug(ctx, "repo: OwnedBlueprintsRepository.GetByUserID called", "userID", userID)

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	filter := bson.M{"userId": userID}
	var ownedBlueprints models.OwnedBlueprints

	err := r.collection.FindOne(ctx, filter).Decode(&ownedBlueprints)
	if err == mongo.ErrNoDocuments {
		logger.Debug(ctx, "repo: OwnedBlueprintsRepository.GetByUserID - no owned blueprints found for user")
		return nil, nil
	}
	if err != nil {
		logger.Error(ctx, "repo: OwnedBlueprintsRepository.GetByUserID - error querying database", "error", err)
		return nil, err
	}

	logger.Debug(ctx, "repo: OwnedBlueprintsRepository.GetByUserID - found owned blueprints", "blueprintCount", len(ownedBlueprints.Blueprints))
	return &ownedBlueprints, nil
}

func (r *OwnedBlueprintsRepository) Create(ctx context.Context, ownedBlueprints *models.OwnedBlueprints) error {
	logger.Debug(ctx, "repo: OwnedBlueprintsRepository.Create called", "userID", ownedBlueprints.UserID)

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	ownedBlueprints.CreatedAt = time.Now()
	ownedBlueprints.UpdatedAt = time.Now()
	if ownedBlueprints.Blueprints == nil {
		ownedBlueprints.Blueprints = []models.OwnedBlueprint{}
	}

	result, err := r.collection.InsertOne(ctx, ownedBlueprints)
	if err != nil {
		logger.Error(ctx, "repo: OwnedBlueprintsRepository.Create - error inserting owned blueprints", "error", err)
		return err
	}

	ownedBlueprints.ID = result.InsertedID.(primitive.ObjectID)
	logger.Info(ctx, "repo: OwnedBlueprintsRepository.Create - owned blueprints created", "id", ownedBlueprints.ID.Hex())
	return nil
}

func (r *OwnedBlueprintsRepository) AddBlueprint(ctx context.Context, userID string, blueprint models.OwnedBlueprint) error {
	logger.Debug(ctx, "repo: OwnedBlueprintsRepository.AddBlueprint called", "userID", userID, "uniqueName", blueprint.UniqueName)

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	filter := bson.M{"userId": userID}
	update := bson.M{
		"$push": bson.M{"blueprints": blueprint},
		"$set":  bson.M{"updatedAt": time.Now()},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		logger.Error(ctx, "repo: OwnedBlueprintsRepository.AddBlueprint - error updating owned blueprints", "error", err)
		return err
	}

	logger.Debug(ctx, "repo: OwnedBlueprintsRepository.AddBlueprint - completed", "matchedCount", result.MatchedCount, "modifiedCount", result.ModifiedCount)
	return nil
}

func (r *OwnedBlueprintsRepository) RemoveBlueprint(ctx context.Context, userID, uniqueName string) error {
	logger.Debug(ctx, "repo: OwnedBlueprintsRepository.RemoveBlueprint called", "userID", userID, "uniqueName", uniqueName)

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	filter := bson.M{"userId": userID}
	update := bson.M{
		"$pull": bson.M{"blueprints": bson.M{"uniqueName": uniqueName}},
		"$set":  bson.M{"updatedAt": time.Now()},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		logger.Error(ctx, "repo: OwnedBlueprintsRepository.RemoveBlueprint - error updating owned blueprints", "error", err)
		return err
	}

	logger.Debug(ctx, "repo: OwnedBlueprintsRepository.RemoveBlueprint - completed", "matchedCount", result.MatchedCount, "modifiedCount", result.ModifiedCount)
	return nil
}

func (r *OwnedBlueprintsRepository) BulkAddBlueprints(ctx context.Context, userID string, blueprints []models.OwnedBlueprint) error {
	logger.Debug(ctx, "repo: OwnedBlueprintsRepository.BulkAddBlueprints called", "userID", userID, "count", len(blueprints))

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	filter := bson.M{"userId": userID}
	update := bson.M{
		"$push": bson.M{"blueprints": bson.M{"$each": blueprints}},
		"$set":  bson.M{"updatedAt": time.Now()},
	}

	opts := options.Update().SetUpsert(true)
	result, err := r.collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		logger.Error(ctx, "repo: OwnedBlueprintsRepository.BulkAddBlueprints - error updating owned blueprints", "error", err)
		return err
	}

	logger.Debug(ctx, "repo: OwnedBlueprintsRepository.BulkAddBlueprints - completed", "matchedCount", result.MatchedCount, "modifiedCount", result.ModifiedCount, "upsertedCount", result.UpsertedCount)
	return nil
}

func (r *OwnedBlueprintsRepository) ClearAll(ctx context.Context, userID string) error {
	logger.Debug(ctx, "repo: OwnedBlueprintsRepository.ClearAll called", "userID", userID)

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	filter := bson.M{"userId": userID}
	update := bson.M{
		"$set": bson.M{
			"blueprints": []models.OwnedBlueprint{},
			"updatedAt":  time.Now(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		logger.Error(ctx, "repo: OwnedBlueprintsRepository.ClearAll - error clearing owned blueprints", "error", err)
		return err
	}

	logger.Debug(ctx, "repo: OwnedBlueprintsRepository.ClearAll - completed", "matchedCount", result.MatchedCount, "modifiedCount", result.ModifiedCount)
	return nil
}
