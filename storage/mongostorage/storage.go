package mongostorage

import (
	"context"
	storageInterface "decentralizedProject/storage"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx"
	"os"
	"time"
)

var dbName = os.Getenv("MONGO_DBNAME")

const collName = "orders"

type storage struct {
	posts *mongo.Collection
}

func NewStorage(mongoURL string) *storage {
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURL))
	if err != nil {
		panic(err)
	}

	collection := client.Database(dbName).Collection(collName)

	ensureIndexes(ctx, collection)

	return &storage{
		posts: collection,
	}
}

func ensureIndexes(ctx context.Context, collection *mongo.Collection) {
	indexModels := []mongo.IndexModel{
		{
			Keys: bsonx.Doc{{Key: "_id", Value: bsonx.Int32(1)}},
		},
	}
	opts := options.CreateIndexes().SetMaxTime(10 * time.Second)

	_, err := collection.Indexes().CreateMany(ctx, indexModels, opts)
	if err != nil {
		panic(fmt.Errorf("failed to ensure indexes %w", err))
	}
}

func (s storage) GetOrder(ctx context.Context, weight storageInterface.Weight,
	size storageInterface.Size, pointStart storageInterface.DeparturePoint,
	pointEnd storageInterface.ReceivePoint, orderReadyDate storageInterface.ISOTimestamp) ([]storageInterface.OrderInfoForClient, error) {

	itemFast := storageInterface.Order{
		Id:             primitive.NewObjectID(),
		Weight:         weight,
		Size:           size,
		DeparturePoint: pointStart,
		ReceivePoint:   pointEnd,
		OrderReadyDate: orderReadyDate,
		Price:          CalculatePrice(),
		DeliveryDate:   CalculateDeliveryDate(),
		OrderType:      storageInterface.OrderType("fast_Delivery"),
		OrderStatus:    "NotConfirmed",
	}
	itemSlow := storageInterface.Order{
		Id:             primitive.NewObjectID(),
		Weight:         weight,
		Size:           size,
		DeparturePoint: pointStart,
		ReceivePoint:   pointEnd,
		OrderReadyDate: orderReadyDate,
		Price:          CalculatePrice(),
		DeliveryDate:   CalculateDeliveryDate(),
		OrderType:      storageInterface.OrderType("slow_Delivery"),
		OrderStatus:    "NotConfirmed",
	}

	_, err := s.posts.InsertOne(ctx, itemFast)
	if err != nil {
		return []storageInterface.OrderInfoForClient{}, fmt.Errorf("something went wrong with saving data to storage")
	}

	_, err = s.posts.InsertOne(ctx, itemSlow)
	if err != nil {
		return []storageInterface.OrderInfoForClient{}, fmt.Errorf("something went wrong with saving data to storage")
	}

	var orders []storageInterface.OrderInfoForClient
	answerFast := storageInterface.OrderInfoForClient{
		Id:           itemFast.Id,
		Price:        itemFast.Price,
		DeliveryDate: itemFast.DeliveryDate,
		OrderType:    itemFast.OrderType,
	}
	answerSlow := storageInterface.OrderInfoForClient{
		Id:           itemSlow.Id,
		Price:        itemSlow.Price,
		DeliveryDate: itemSlow.DeliveryDate,
		OrderType:    itemSlow.OrderType,
	}
	orders = append(orders, answerFast, answerSlow)
	return orders, nil
}

func (s storage) AcceptDelivery(ctx context.Context, id storageInterface.OrderId) (string, error) {
	valueId, _ := primitive.ObjectIDFromHex(string(id))
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	filter := bson.D{{"_id", valueId}}
	update := bson.D{
		{"$set", bson.D{{"orderStatus", "Confirmed"}}},
	}
	order := s.posts.FindOneAndUpdate(
		ctx,
		filter,
		update,
		opts,
	)
	result := storageInterface.Order{}
	err := order.Decode(&result)
	if err != nil {
		return "Not Confirmed, error occured", fmt.Errorf("%w",
			storageInterface.StorageError)
	}
	return "OK", nil
}

func CalculatePrice() storageInterface.Price {
	return storageInterface.Price(123.0)
}

func CalculateDeliveryDate() storageInterface.ISOTimestamp {
	currentTime := storageInterface.ISOTimestamp(time.Now().UTC().Format(time.RFC3339))
	return currentTime
}
