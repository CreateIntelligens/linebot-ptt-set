package controllers

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/mong0520/linebot-ptt-set/models"
	"github.com/mong0520/linebot-ptt-set/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserFavorite struct {
    UserId    string   `json:"user_id" bson:"user_id"`
    Favorites []string `json:"favorites" bson:"favorites"`
}

func GetOne(collection *mongo.Collection, query bson.M) (result *models.ArticleDocument, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	result = &models.ArticleDocument{}
	err = collection.FindOne(ctx, query).Decode(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func Get(collection *mongo.Collection, page int, perPage int) (results []models.ArticleDocument, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	query := bson.M{"article_title": bson.M{"$regex": ".*"}}
	
	opts := options.Find()
	opts.SetSort(bson.D{{"timestamp", -1}})
	opts.SetSkip(int64(page * perPage))
	opts.SetLimit(int64(perPage))
	
	cursor, err := collection.Find(ctx, query, opts)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer cursor.Close(ctx)
	
	err = cursor.All(ctx, &results)
	if err != nil {
		return nil, err
	}
	
	return results, nil
}

func GetAll(collection *mongo.Collection, query bson.M) (results []models.ArticleDocument, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	cursor, err := collection.Find(ctx, query)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	
	err = cursor.All(ctx, &results)
	if err != nil {
		return nil, err
	}
	
	return results, nil
}

func GetRandom(collection *mongo.Collection, count int, keyword string) (results []models.ArticleDocument, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	query := bson.M{}
	baseline_ts := 1420070400 // 2015年Jan/1/00:00:00 之後
	needRandom := true
	
	if keyword == "" {
		query = bson.M{
			"timestamp": bson.M{"$gte": baseline_ts}, 
			"article_title": bson.M{"$regex": "^(?!\\[公告\\]).*"},
		}
	} else {
		query = bson.M{
			"timestamp":     bson.M{"$gte": baseline_ts},
			"article_title": bson.M{"$regex": fmt.Sprintf("^(?!\\[公告\\]).*%s.*", strings.ToLower(keyword))},
		}
	}

	total, _ := collection.CountDocuments(ctx, query)
	fmt.Println("total = ", total)
	if total == 0 {
		return nil, errors.New("NotFound")
	} else if int(total) < count {
		count = int(total)
		needRandom = false
	}
	fmt.Println("count = ", count)
	
	if needRandom {
		randSkip := utils.GetRandomIntSet(int(total), count)
		for i := 0; i < count; i++ {
			skip := randSkip[i]
			opts := options.FindOne().SetSkip(int64(skip))
			result := &models.ArticleDocument{}
			collection.FindOne(ctx, query, opts).Decode(result)
			results = append(results, *result)
		}
		sort.Slice(results, func(i, j int) bool {
			return results[i].MessageCount.Push > results[j].MessageCount.Push
		})
	} else {
		opts := options.Find()
		opts.SetSort(bson.D{{"message_count.push", -1}})
		opts.SetLimit(int64(count))
		
		cursor, err := collection.Find(ctx, query, opts)
		if err != nil {
			return nil, err
		}
		defer cursor.Close(ctx)
		
		err = cursor.All(ctx, &results)
		if err != nil {
			return nil, err
		}
	}

	return results, nil
}

func GetMostLike(collection *mongo.Collection, count int, timestampOffset int) (results []models.ArticleDocument, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	query := bson.M{}
	if timestampOffset > 0 {
		now := time.Now()
		nowInSec := int(now.Unix())
		start := nowInSec - timestampOffset
		query = bson.M{
			"timestamp": bson.M{"$gte": start, "$lt": nowInSec}, 
			"article_title": bson.M{"$regex": "^(?!\\[公告\\]).*"},
		}
	} else {
		query = bson.M{"article_title": bson.M{"$regex": "^(?!\\[公告\\]).*"}}
	}
	
	opts := options.Find()
	opts.SetSort(bson.D{{"message_count.push", -1}})
	opts.SetLimit(int64(count))
	
	cursor, err := collection.Find(ctx, query, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	
	err = cursor.All(ctx, &results)
	if err != nil {
		return nil, err
	}
	
	return results, nil
}

func (u *UserFavorite) Add(meta *models.Model) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	if _, err := meta.CollectionUserFavorite.InsertOne(ctx, u); err != nil {
		meta.Log.Println(err)
	}
}

func (u *UserFavorite) Get(meta *models.Model) (result *UserFavorite, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	meta.Log.Println(u.UserId)
	query := bson.M{"user_id": u.UserId}
	
	result = &UserFavorite{}
	if err := meta.CollectionUserFavorite.FindOne(ctx, query).Decode(result); err != nil {
		meta.Log.Println(err)
		return nil, err
	}
	
	return result, nil
}

func (u *UserFavorite) Update(meta *models.Model) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	meta.Log.Println(u.UserId)
	query := bson.M{"user_id": u.UserId}
	update := bson.M{"$set": u}
	
	if _, err := meta.CollectionUserFavorite.UpdateOne(ctx, query, update); err != nil {
		meta.Log.Println(err)
		return err
	}
	
	return nil
}