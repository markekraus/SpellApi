package db

import (
	"context"

	"github.com/honeycombio/beeline-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DB struct {
	*mongo.Client
}

// Connect to the specified mongo instance using the context for timeout
func ConnectDb(uri string) (*DB, error) {
	ctx := context.Background()
	clientOptions := options.Client().ApplyURI(uri).SetDirect(true)
	c, err := mongo.NewClient(clientOptions)
	if err != nil {
		return nil, err
	}

	err = c.Connect(ctx)
	if err != nil {
		return nil, err
	}

	err = c.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}

	return &DB{c}, nil
}

//	collection := mc.Database("reminders").Collection("reminders")

func runQuery(ctx context.Context, mc *mongo.Collection, query interface{}) ([]bson.M, error) {

	ctx, span := beeline.StartSpan(ctx, "Mongo.RunQuery")
	defer span.Send()

	beeline.AddField(ctx, "Mongo.RunQuery.Collection", mc.Name())
	beeline.AddField(ctx, "Mongo.RunQuery.Database", mc.Database().Name())
	beeline.AddField(ctx, "Mongo.RunQuery.Query", query)

	cursor, err := mc.Find(ctx, query)
	if err != nil {
		beeline.AddField(ctx, "Mongo.RunQuery.Error", err)
		return nil, err
	}

	var results []bson.M
	if err = cursor.All(ctx, &results); err != nil {
		beeline.AddField(ctx, "Mongo.RunQuery.Error", err)
		return nil, err
	}

	beeline.AddField(ctx, "Mongo.RunQuery.Results.Count", len(results))
	beeline.AddField(ctx, "Mongo.RunQuery.Results.Raw", results)

	return results, nil
}

func writeDbObject(ctx context.Context, mc *mongo.Collection, obj []byte) error {

	ctx, span := beeline.StartSpan(ctx, "Mongo.WriteObject")
	defer span.Send()

	beeline.AddField(ctx, "Mongo.WriteObject.Collection", mc.Name())
	beeline.AddField(ctx, "Mongo.WriteObject.Database", mc.Database().Name())
	beeline.AddField(ctx, "Mongo.WriteObject.Object", obj)

	res, err := mc.InsertOne(ctx, obj)
	if err != nil {
		beeline.AddField(ctx, "Mongo.WriteObject.Error", err)
		return err
	}

	beeline.AddField(ctx, "Mongo.WriteObject.Id", res.InsertedID)

	return nil
}

func deleteDbObject(ctx context.Context, mc *mongo.Collection, query interface{}) error {

	ctx, span := beeline.StartSpan(ctx, "Mongo.DeleteDbObject")
	defer span.Send()

	beeline.AddField(ctx, "Mongo.DeleteDbObject.Collection", mc.Name())
	beeline.AddField(ctx, "Mongo.DeleteDbObject.Database", mc.Database().Name())
	beeline.AddField(ctx, "Mongo.DeleteDbObject.Query", query)

	deleted, err := mc.DeleteOne(ctx, query)
	if err != nil {
		beeline.AddField(ctx, "Mongo.DeleteDbObject.Error", err)
		return err
	}

	beeline.AddField(ctx, "Mongo.DeleteDbObject.DeletedCount", deleted)

	return nil
}

func (db *DB) GetSpell(ctx context.Context, search bson.M) ([]bson.M, error) {

	ctx, span := beeline.StartSpan(ctx, "Mongo.GetSpell")
	defer span.Send()

	beeline.AddField(ctx, "Mongo.GetSpell.Query", search)

	collection := db.Database("spellapi").Collection("spells")

	result, err := runQuery(ctx, collection, search)
	if err != nil {
		beeline.AddField(ctx, "Mongo.GetSpell.Error", err)
		return nil, err
	}

	beeline.AddField(ctx, "Mongo.GetSpell.Result", result)

	return result, nil
}

func (db *DB) AddSpell(ctx context.Context, spell []byte) error {

	ctx, span := beeline.StartSpan(ctx, "Mongo.AddSpell")
	defer span.Send()

	beeline.AddField(ctx, "Mongo.AddSpell.Spell", spell)

	collection := db.Database("spellapi").Collection("spells")

	err := writeDbObject(ctx, collection, spell)
	if err != nil {
		beeline.AddField(ctx, "Mongo.AddSpell.Error", err)
		return err
	}

	return nil
}

func (db *DB) DeleteSpell(ctx context.Context, spell bson.M) error {

	ctx, span := beeline.StartSpan(ctx, "Mongo.DeleteSpell")
	defer span.Send()

	beeline.AddField(ctx, "Mongo.DeleteSpell.Spell", spell)

	collection := db.Database("spellapi").Collection("spells")

	err := deleteDbObject(ctx, collection, spell)
	if err != nil {
		beeline.AddField(ctx, "Mongo.DeleteSpell.Error", err)
		return err
	}

	return nil
}
