package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Haiku represents a haiku poem with environmental parameters.
type Haiku struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	Date         time.Time          `bson:"date"`
	Text         string             `bson:"text"`
	Moisture     int                `bson:"moisture"`
	Temperature  int                `bson:"temperature"`
	Illumination int                `bson:"illumination"`
	PH           int                `bson:"ph"`
}
