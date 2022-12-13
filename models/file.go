package models

import (
	"time"

	"github.com/CPU-commits/Intranet_BFiles/db"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

const FILES_COLLECTION = "files"

type File struct {
	ID          primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	Filename    string             `json:"filename" bson:"filename"`
	Key         string             `json:"key" bson:"key"`
	URL         string             `json:"url" bson:"url"`
	Title       string             `json:"title" bson:"title"`
	Type        string             `json:"type" bson:"type"`
	User        primitive.ObjectID `json:"user" bson:"user"`
	Status      bool               `json:"status" bson:"status"`
	Permissions string             `json:"permissions" bson:"permissions"`
	Date        primitive.DateTime `json:"date" bson:"date"`
}

type FilesModel struct{}

func (f *FilesModel) Use() *mongo.Collection {
	return DbConnect.GetCollection(FILES_COLLECTION)
}

func (f *FilesModel) Transaction(todo func() error) error {
	wc := writeconcern.New(writeconcern.WMajority())
	rc := readconcern.Snapshot()
	txnOpts := options.Transaction().SetWriteConcern(wc).SetReadConcern(rc)

	client := db.NewMongoClient(settingsData.MONGO_HOST)
	session, err := client.StartSession()
	if err != nil {
		panic(err)
	}
	defer session.EndSession(db.Ctx)
	err = mongo.WithSession(db.Ctx, session, func(sc mongo.SessionContext) error {
		if err = session.StartTransaction(txnOpts); err != nil {
			return err
		}
		err = todo()
		if err != nil {
			return nil
		}
		return nil
	})
	if err != nil {
		if abortErr := session.AbortTransaction(db.Ctx); abortErr != nil {
			panic(abortErr)
		}
		return err
	}
	return nil
}

func (f *FilesModel) NewModel(filename, key, url, title, typeFile, idUser, permissions string) (*File, error) {
	file := &File{
		Filename:    filename,
		Key:         key,
		URL:         url,
		Title:       title,
		Type:        typeFile,
		Status:      true,
		Permissions: permissions,
		Date:        primitive.NewDateTimeFromTime(time.Now()),
	}
	if idUser != "" {
		idObjUser, err := primitive.ObjectIDFromHex(idUser)
		if err != nil {
			return nil, err
		}
		file.User = idObjUser
	}
	return file, nil
}

func init() {
	collections, errC := DbConnect.GetCollections()
	if errC != nil {
		panic(errC)
	}
	for _, collection := range collections {
		if collection == FILES_COLLECTION {
			return
		}
	}
	var jsonSchema = bson.M{
		"bsonType": "object",
		"required": []string{
			"filename",
			"key",
			"title",
			"url",
			"type",
			"status",
			"date",
		},
		"properties": bson.M{
			"user":     bson.M{"bsonType": "objectId"},
			"filename": bson.M{"bsonType": "string"},
			"title": bson.M{
				"bsonType":  "string",
				"maxLength": 100,
			},
			"key":         bson.M{"bsonType": "string"},
			"img":         bson.M{"bsonType": "objectId"},
			"url":         bson.M{"bsonType": "string"},
			"permissions": bson.M{"enum": bson.A{"private", "public", "public_classroom"}},
			"status":      bson.M{"bsonType": "bool"},
			"date":        bson.M{"bsonType": "date"},
			"type":        bson.M{"bsonType": "string"},
		},
	}
	var validators = bson.M{
		"$jsonSchema": jsonSchema,
	}
	opts := &options.CreateCollectionOptions{
		Validator: validators,
	}
	err := DbConnect.CreateCollection(FILES_COLLECTION, opts)
	if err != nil {
		panic(err)
	}
}
