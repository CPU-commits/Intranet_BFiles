package models

import (
	"github.com/CPU-commits/Intranet_BFiles/db"
	"github.com/CPU-commits/Intranet_BFiles/settings"
	"go.mongodb.org/mongo-driver/mongo"
)

var settingsData = settings.GetSettings()

var DbConnect = db.NewConnection(
	settingsData.MONGO_HOST,
	settingsData.MONGO_DB,
)

type Models interface {
	Use() *mongo.Collection
	NewModel() interface{}
}
