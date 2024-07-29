package repositories

import (
	"context"

	"github.com/luikyv/go-opf/gopf/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Consent struct {
	collection *mongo.Collection
}

func NewConsent(db *mongo.Database) Consent {
	return Consent{
		collection: db.Collection("consents"),
	}
}

func (repo *Consent) Save(ctx context.Context, consent models.Consent) error {
	shouldUpsert := true
	filter := bson.D{{Key: "_id", Value: consent.ID}}
	if _, err := repo.collection.ReplaceOne(ctx, filter, consent, &options.ReplaceOptions{Upsert: &shouldUpsert}); err != nil {
		return err
	}

	return nil
}

func (repo *Consent) Get(ctx context.Context, id string) (models.Consent, error) {
	filter := bson.D{{Key: "_id", Value: id}}

	result := repo.collection.FindOne(ctx, filter)
	if result.Err() != nil {
		return models.Consent{}, result.Err()
	}

	var consent models.Consent
	if err := result.Decode(&consent); err != nil {
		return models.Consent{}, err
	}

	return consent, nil
}

func (repo *Consent) Delete(ctx context.Context, id string) error {
	filter := bson.D{{Key: "_id", Value: id}}
	if _, err := repo.collection.DeleteOne(ctx, filter); err != nil {
		return err
	}

	return nil
}
