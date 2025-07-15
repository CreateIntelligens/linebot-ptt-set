package models

import (
	"encoding/json"
	"log"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/bson/primitive"
)



type Model struct {
	MongoClient            *mongo.Client
	Database               *mongo.Database
	Collection             *mongo.Collection
	CollectionUserFavorite *mongo.Collection
	Log                    *log.Logger
}

type MessageCount struct {
	All     int `json:"all" bson:"all"`
	Boo     int `json:"boo" bson:"boo"`
	Count   int `json:"count" bson:"count"`
	Neutral int `json:"neutral" bson:"neutral"`
	Push    int `json:"push" bson:"push"`
}

type ArticleDocument struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	ArticleID    string             `json:"article_id" bson:"article_id"`
	ArticleTitle string             `json:"article_title" bson:"article_title"`
	Author       string             `json:"author" bson:"author"`
	Board        string             `json:"board" bson:"board"`
	Content      string             `json:"content" bson:"content"`
	Date         string             `json:"date" bson:"date"`
	IP           string             `json:"ip" bson:"ip"`
	MessageCount MessageCount       `bson:"message_count"`
	Messages     []interface{}      `json:"messages" bson:"messages"`
	Timestamp    int                `json:"timestamp" bson:"timestamp"`
	URL          string             `json:"url" bson:"url"`
	ImageLinks   []string           `json:"image_links" bson:"image_links"`
}

// Note: These methods will be replaced with new MongoDB driver implementations in controllers

func (d *ArticleDocument) ToString() (info string) {
	b, err := json.Marshal(d)
	if err != nil {
		//fmt.Println(err)
		return
	}
	return string(b)
}
