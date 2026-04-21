package dbops

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	constants "go-web-blog/constants"
	"go-web-blog/models"
)

var collection *mongo.Collection
var ctx = context.TODO()

var MONGOURI = constants.MONGOURI

// var MONGOURI = os.Getenv("MONGO_URI")
// var MONGOURI = "mongodb://mongo:27017"

func Init() (context.Context, *mongo.Client) {
	clientOptions := options.Client().ApplyURI(MONGOURI)
	client, err := mongo.Connect(clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to MongoDB!")
	return ctx, client
}

func Write(collection string, data interface{}) (bson.ObjectID, error) {
	ctx, client := Init()

	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			log.Fatal(err)
		}
		fmt.Println("Connection to MongoDB closed.")
	}()

	coll := client.Database(constants.DBNAME).Collection(collection)

	res, err := coll.InsertOne(ctx, data)
	if err != nil {
		return bson.ObjectID{}, err
	}
	insertedID, ok := res.InsertedID.(bson.ObjectID)
	if !ok {
		return bson.ObjectID{}, err
	}
	return insertedID, nil
}

func Read(ctx context.Context, client *mongo.Client, collect *mongo.Collection, user string) (models.User, error) {

	// w, r = http_headers.W, http_headers.R

	filter := bson.D{{Key: "username", Value: user}}
	cursor, err := collect.Find(ctx, filter)
	if err != nil {
		return models.User{}, fmt.Errorf("Error: User not found %s", err)
	}
	var res []models.User
	var userdet models.User
	if err = cursor.All(context.TODO(), &res); err != nil {
		fmt.Printf("Error %s\n", err)
	}

	for _, result := range res {
		res, _ := bson.MarshalExtJSON(result, false, false)
		err := json.Unmarshal(res, &userdet)
		if err != nil {
			return models.User{}, fmt.Errorf("Error: %s", err)
		}
	}
	return userdet, nil
}

// func Read(ctx context.Context, client *mongo.Client, collect_ *mongo.Collection) {
// 	user := "John Marshton"

// 	filter := bson.D{{Key: "username", Value: user}}
// 	cursor, err := collect_.Find(ctx, filter)
// 	if err != nil {
// 		fmt.Printf("Error %s\n", err)
// 	}
// 	var res []User
// 	var userdet User
// 	if err = cursor.All(context.TODO(), &res); err != nil {
// 		fmt.Printf("Error %s\n", err)
// 	}
// 	for _, result := range res {
// 		res, _ := bson.MarshalExtJSON(result, false, false)
// 		err := json.Unmarshal(res, &userdet)
// 		if err == nil {
// 			fmt.Printf("Username: %s\n", userdet.Username)
// 			fmt.Printf("Password: %s\n", userdet.Password)
// 		}
// 	}
// }
