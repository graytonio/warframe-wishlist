package repository

import (
	"context"
	"time"

	"github.com/graytonio/warframe-wishlist/internal/database"
	"github.com/graytonio/warframe-wishlist/internal/models"
	"github.com/graytonio/warframe-wishlist/pkg/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var ItemCollections = []string{
	"warframes", "melee", "primary", "secondary", "arch_gun", "arch_melee",
	"archwing", "pets", "sentinels", "sentinelweapons", "railjack", "arcanes",
	"mods", "resources", "gear", "misc", "fish", "glyphs", "sigils", "skins",
	"relics", "quests", "node", "enemy",
}

type ItemRepository struct {
	db *database.MongoDB
}

func NewItemRepository(db *database.MongoDB) *ItemRepository {
	return &ItemRepository{db: db}
}

func (r *ItemRepository) Search(ctx context.Context, params models.SearchParams) ([]models.ItemSearchResult, error) {
	logger.Debug(ctx, "repo: ItemRepository.Search called", "query", params.Query, "category", params.Category, "limit", params.Limit, "offset", params.Offset)

	var results []models.ItemSearchResult

	limit := params.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	offset := params.Offset
	if offset < 0 {
		offset = 0
	}

	filter := bson.M{}
	if params.Query != "" {
		filter["name"] = bson.M{"$regex": primitive.Regex{Pattern: params.Query, Options: "i"}}
	}

	collections := ItemCollections
	if params.Category != "" {
		collections = []string{params.Category}
	}

	findOptions := options.Find().
		SetProjection(bson.M{
			"uniqueName":  1,
			"name":        1,
			"description": 1,
			"category":    1,
			"imageName":   1,
		}).
		SetLimit(int64(limit)).
		SetSkip(int64(offset))

	logger.Debug(ctx, "repo: ItemRepository.Search - searching collections", "collectionCount", len(collections))
	for _, collName := range collections {
		collection := r.db.Collection(collName)

		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		cursor, err := collection.Find(ctx, filter, findOptions)
		cancel()
		if err != nil {
			logger.Debug(ctx, "repo: ItemRepository.Search - error querying collection", "collection", collName, "error", err)
			continue
		}

		var items []models.ItemSearchResult
		if err := cursor.All(ctx, &items); err != nil {
			logger.Debug(ctx, "repo: ItemRepository.Search - error decoding results", "collection", collName, "error", err)
			cursor.Close(ctx)
			continue
		}
		cursor.Close(ctx)

		for i := range items {
			items[i].Collection = collName
		}

		logger.Debug(ctx, "repo: ItemRepository.Search - found items in collection", "collection", collName, "count", len(items))
		results = append(results, items...)

		if len(results) >= limit {
			results = results[:limit]
			break
		}
	}

	logger.Debug(ctx, "repo: ItemRepository.Search - completed", "totalResults", len(results))
	return results, nil
}

func (r *ItemRepository) FindByUniqueName(ctx context.Context, uniqueName string) (*models.Item, error) {
	logger.Debug(ctx, "repo: ItemRepository.FindByUniqueName called", "uniqueName", uniqueName)

	filter := bson.M{"uniqueName": uniqueName}

	for _, collName := range ItemCollections {
		collection := r.db.Collection(collName)

		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		var item models.Item
		err := collection.FindOne(ctx, filter).Decode(&item)
		cancel()

		if err == nil {
			item.Collection = collName
			logger.Debug(ctx, "repo: ItemRepository.FindByUniqueName - found item", "uniqueName", uniqueName, "collection", collName, "itemName", item.Name)
			return &item, nil
		}
	}

	logger.Debug(ctx, "repo: ItemRepository.FindByUniqueName - item not found", "uniqueName", uniqueName)
	return nil, nil
}

func (r *ItemRepository) FindByUniqueNames(ctx context.Context, uniqueNames []string) (map[string]*models.Item, error) {
	logger.Debug(ctx, "repo: ItemRepository.FindByUniqueNames called", "count", len(uniqueNames))

	result := make(map[string]*models.Item)

	if len(uniqueNames) == 0 {
		logger.Debug(ctx, "repo: ItemRepository.FindByUniqueNames - empty input, returning empty result")
		return result, nil
	}

	filter := bson.M{"uniqueName": bson.M{"$in": uniqueNames}}

	for _, collName := range ItemCollections {
		collection := r.db.Collection(collName)

		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		cursor, err := collection.Find(ctx, filter)
		cancel()
		if err != nil {
			logger.Debug(ctx, "repo: ItemRepository.FindByUniqueNames - error querying collection", "collection", collName, "error", err)
			continue
		}

		var items []models.Item
		if err := cursor.All(ctx, &items); err != nil {
			logger.Debug(ctx, "repo: ItemRepository.FindByUniqueNames - error decoding results", "collection", collName, "error", err)
			cursor.Close(ctx)
			continue
		}
		cursor.Close(ctx)

		if len(items) > 0 {
			logger.Debug(ctx, "repo: ItemRepository.FindByUniqueNames - found items in collection", "collection", collName, "count", len(items))
		}
		for i := range items {
			items[i].Collection = collName
			result[items[i].UniqueName] = &items[i]
		}
	}

	logger.Debug(ctx, "repo: ItemRepository.FindByUniqueNames - completed", "foundCount", len(result))
	return result, nil
}
