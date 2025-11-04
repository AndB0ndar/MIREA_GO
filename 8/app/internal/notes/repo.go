package notes

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	ErrNotFound     = errors.New("note not found")
	ErrDuplicateKey = errors.New("duplicate title")
	ErrInvalidID    = errors.New("invalid id format")
)

type Repo struct {
	col *mongo.Collection
}

func NewRepo(db *mongo.Database) (*Repo, error) {
	col := db.Collection("notes")

	// Создаем уникальный индекс на поле title
	_, err := col.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys:    bson.D{{Key: "title", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return nil, err
	}

	return &Repo{col: col}, nil
}

func (r *Repo) Create(ctx context.Context, title, content string) (Note, error) {
	now := time.Now()
	n := Note{
		Title:     title,
		Content:   content,
		CreatedAt: now,
		UpdatedAt: now,
	}

	res, err := r.col.InsertOne(ctx, n)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return Note{}, ErrDuplicateKey
		}
		return Note{}, err
	}

	n.ID = res.InsertedID.(primitive.ObjectID)
	return n, nil
}

func (r *Repo) ByID(ctx context.Context, idHex string) (Note, error) {
	oid, err := primitive.ObjectIDFromHex(idHex)
	if err != nil {
		return Note{}, ErrInvalidID
	}

	var n Note
	if err := r.col.FindOne(ctx, bson.M{"_id": oid}).Decode(&n); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return Note{}, ErrNotFound
		}
		return Note{}, err
	}

	return n, nil
}

func (r *Repo) List(ctx context.Context, q string, limit, skip int64) ([]Note, error) {
	filter := bson.M{}
	if q != "" {
		filter["title"] = bson.M{"$regex": q, "$options": "i"}
	}

	findOptions := options.Find().
		SetLimit(limit).
		SetSkip(skip).
		SetSort(bson.D{{Key: "createdAt", Value: -1}})

	cur, err := r.col.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var notes []Note
	for cur.Next(ctx) {
		var n Note
		if err := cur.Decode(&n); err != nil {
			return nil, err
		}
		notes = append(notes, n)
	}

	if err := cur.Err(); err != nil {
		return nil, err
	}

	return notes, nil
}

func (r *Repo) Update(ctx context.Context, idHex string, title, content *string) (Note, error) {
	oid, err := primitive.ObjectIDFromHex(idHex)
	if err != nil {
		return Note{}, ErrInvalidID
	}

	updateFields := bson.M{"updatedAt": time.Now()}
	if title != nil {
		updateFields["title"] = *title
	}
	if content != nil {
		updateFields["content"] = *content
	}

	after := options.After
	opts := options.FindOneAndUpdate().
		SetReturnDocument(after)

	var updated Note
	err = r.col.FindOneAndUpdate(
		ctx,
		bson.M{"_id": oid},
		bson.M{"$set": updateFields},
		opts,
	).Decode(&updated)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return Note{}, ErrNotFound
		}
		if mongo.IsDuplicateKeyError(err) {
			return Note{}, ErrDuplicateKey
		}
		return Note{}, err
	}

	return updated, nil
}

func (r *Repo) Delete(ctx context.Context, idHex string) error {
	oid, err := primitive.ObjectIDFromHex(idHex)
	if err != nil {
		return ErrInvalidID
	}

	res, err := r.col.DeleteOne(ctx, bson.M{"_id": oid})
	if err != nil {
		return err
	}

	if res.DeletedCount == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *Repo) Count(ctx context.Context) (int64, error) {
	return r.col.CountDocuments(ctx, bson.M{})
}
