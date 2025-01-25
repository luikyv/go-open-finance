package consent

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Storage struct {
	collection *mongo.Collection
}

func NewStorage(db *mongo.Database) Storage {
	return Storage{
		collection: db.Collection("consents"),
	}
}

func (st Storage) save(ctx context.Context, consent Consent) error {
	shouldUpsert := true
	filter := bson.D{{Key: "_id", Value: consent.ID}}
	if _, err := st.collection.ReplaceOne(ctx, filter, consent, &options.ReplaceOptions{
		Upsert: &shouldUpsert,
	}); err != nil {
		return err
	}

	return nil
}

func (st Storage) consent(ctx context.Context, id string) (Consent, error) {

	filter := bson.D{{Key: "_id", Value: id}}
	result := st.collection.FindOne(ctx, filter)
	if result.Err() != nil {
		return Consent{}, result.Err()
	}

	var consent Consent
	if err := result.Decode(&consent); err != nil {
		return Consent{}, err
	}

	return consent, nil
}
