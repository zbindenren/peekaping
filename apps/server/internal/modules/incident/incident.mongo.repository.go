package incident

import (
	"context"
	"peekaping/internal/config"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var _ Repository = (*MongoRepositoryImpl)(nil)

type mongoModel struct {
	ID           primitive.ObjectID `bson:"_id"`
	StatusPageID primitive.ObjectID `bson:"status_page_id"`
	Title        string             `bson:"title"`
	Content      string             `bson:"content"`
	Style        string             `bson:"style"`
	Active       bool               `bson:"active"`
	ResolvedAt   *time.Time         `bson:"resolved_at,omitempty"`
	CreatedAt    time.Time          `bson:"created_at"`
	UpdatedAt    time.Time          `bson:"updated_at"`
}

type mongoUpdateModel struct {
	Title      *string `bson:"title,omitempty"`
	Content    *string `bson:"content,omitempty"`
	Style      *string `bson:"style,omitempty"`
	Active     *bool   `bson:"active,omitempty"`
	ResolvedAt *string `bson:"resolved_at,omitempty"`
	UpdatedAt  *string `bson:"updated_at,omitempty"`
}

func toDomainModel(mm *mongoModel) *Model {
	return &Model{
		ID:           mm.ID.Hex(),
		StatusPageID: mm.StatusPageID.Hex(),
		Title:        mm.Title,
		Content:      mm.Content,
		Style:        mm.Style,
		Active:       mm.Active,
		ResolvedAt:   mm.ResolvedAt,
		CreatedAt:    mm.CreatedAt,
		UpdatedAt:    mm.UpdatedAt,
	}
}

type MongoRepositoryImpl struct {
	client     *mongo.Client
	db         *mongo.Database
	collection *mongo.Collection
}

func NewMongoRepository(client *mongo.Client, cfg *config.Config) Repository {
	db := client.Database(cfg.DBName)
	collection := db.Collection("incidents")

	_, err := collection.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys: bson.D{{Key: "status_page_id", Value: 1}},
	})
	if err != nil {
		panic("Failed to create index on incidents collection: " + err.Error())
	}

	return &MongoRepositoryImpl{client, db, collection}
}

func (r *MongoRepositoryImpl) Create(ctx context.Context, entity *CreateDto) (*Model, error) {
	// check that the status page exists
	var sp bson.M
	err := r.db.Collection("status_pages").FindOne(ctx, bson.M{"_id": mustObjectID(entity.StatusPageID)}).Decode(&sp)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, mongo.ErrNoDocuments
		}
		return nil, err
	}

	mm := &mongoModel{
		ID:           primitive.NewObjectID(),
		StatusPageID: mustObjectID(entity.StatusPageID),
		Title:        entity.Title,
		Content:      entity.Content,
		Style:        entity.Style,
		Active:       true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	_, err = r.collection.InsertOne(ctx, mm)
	if err != nil {
		return nil, err
	}

	return toDomainModel(mm), nil
}

func (r *MongoRepositoryImpl) FindByID(ctx context.Context, id string) (*Model, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var mm mongoModel
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&mm)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return toDomainModel(&mm), nil
}

func (r *MongoRepositoryImpl) FindAll(ctx context.Context, page int, limit int, q string) ([]*Model, error) {
	skip := int64(page * limit)
	limit64 := int64(limit)

	opts := &options.FindOptions{
		Skip:  &skip,
		Limit: &limit64,
		Sort:  bson.D{{Key: "created_at", Value: -1}},
	}

	filter := bson.M{}
	if q != "" {
		filter["title"] = bson.M{"$regex": q, "$options": "i"}
	}

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var models []*Model
	for cursor.Next(ctx) {
		var mm mongoModel
		if err := cursor.Decode(&mm); err != nil {
			return nil, err
		}
		models = append(models, toDomainModel(&mm))
	}
	return models, cursor.Err()
}

func (r *MongoRepositoryImpl) FindByStatusPageID(ctx context.Context, statusPageID string, page int, limit int) ([]*Model, error) {
	skip := int64(page * limit)
	limit64 := int64(limit)

	opts := &options.FindOptions{
		Skip:  &skip,
		Limit: &limit64,
		Sort:  bson.D{{Key: "created_at", Value: -1}},
	}

	statusPageObjectID := mustObjectID(statusPageID)
	cursor, err := r.collection.Find(ctx, bson.M{"status_page_id": statusPageObjectID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var models []*Model
	for cursor.Next(ctx) {
		var mm mongoModel
		if err := cursor.Decode(&mm); err != nil {
			return nil, err
		}
		models = append(models, toDomainModel(&mm))
	}
	return models, cursor.Err()
}

func (r *MongoRepositoryImpl) FindActiveByStatusPageID(ctx context.Context, statusPageID string) ([]*Model, error) {
	opts := &options.FindOptions{
		Sort: bson.D{{Key: "created_at", Value: -1}},
	}

	statusPageObjectID := mustObjectID(statusPageID)
	cursor, err := r.collection.Find(ctx, bson.M{"status_page_id": statusPageObjectID, "active": true}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var models []*Model
	for cursor.Next(ctx) {
		var mm mongoModel
		if err := cursor.Decode(&mm); err != nil {
			return nil, err
		}
		models = append(models, toDomainModel(&mm))
	}
	return models, cursor.Err()
}

func (r *MongoRepositoryImpl) FindByStatusPageSlug(ctx context.Context, slug string, page int, limit int) ([]*Model, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "status_pages"},
			{Key: "localField", Value: "status_page_id"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "status_page"},
		}}},
		{{Key: "$unwind", Value: "$status_page"}},
		{{Key: "$match", Value: bson.M{
			"status_page.slug":      slug,
			"status_page.published": true,
		}}},
		{{Key: "$sort", Value: bson.D{{Key: "created_at", Value: -1}}}},
		{{Key: "$skip", Value: int64(page * limit)}},
		{{Key: "$limit", Value: int64(limit)}},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var models []*Model
	for cursor.Next(ctx) {
		var mm mongoModel
		if err := cursor.Decode(&mm); err != nil {
			return nil, err
		}
		models = append(models, toDomainModel(&mm))
	}
	return models, cursor.Err()
}

func (r *MongoRepositoryImpl) Update(ctx context.Context, id string, entity *UpdateDto) (*Model, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	nowStr := now.Format(time.RFC3339)

	update := &mongoUpdateModel{
		Title:     entity.Title,
		Content:   entity.Content,
		Style:     entity.Style,
		Active:    entity.Active,
		UpdatedAt: &nowStr,
	}

	if entity.Active != nil {
		if *entity.Active {
			// reactivating: clear resolved_at by unsetting it
			filter := bson.M{"_id": objectID}
			updateDoc := bson.M{
				"$set":   update,
				"$unset": bson.M{"resolved_at": ""},
			}
			_, err = r.collection.UpdateOne(ctx, filter, updateDoc)
			if err != nil {
				return nil, err
			}
			return r.FindByID(ctx, id)
		}
		// deactivating: set resolved_at
		resolvedStr := nowStr
		update.ResolvedAt = &resolvedStr
	}

	filter := bson.M{"_id": objectID}
	updateDoc := bson.M{"$set": update}

	_, err = r.collection.UpdateOne(ctx, filter, updateDoc)
	if err != nil {
		return nil, err
	}

	return r.FindByID(ctx, id)
}

func (r *MongoRepositoryImpl) Resolve(ctx context.Context, id string) (*Model, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC().Format(time.RFC3339)
	update := bson.M{"$set": bson.M{
		"active":      false,
		"resolved_at": now,
		"updated_at":  now,
	}}

	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	if err != nil {
		return nil, err
	}

	return r.FindByID(ctx, id)
}

func (r *MongoRepositoryImpl) Delete(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	return err
}

func (r *MongoRepositoryImpl) DeleteByStatusPageID(ctx context.Context, statusPageID string) error {
	statusPageObjectID := mustObjectID(statusPageID)
	_, err := r.collection.DeleteMany(ctx, bson.M{"status_page_id": statusPageObjectID})
	return err
}

// mustObjectID converts a hex string to an ObjectID, returning a zero ObjectID on error.
func mustObjectID(hex string) primitive.ObjectID {
	oid, _ := primitive.ObjectIDFromHex(hex)
	return oid
}
