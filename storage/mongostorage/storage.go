package mongostorage

import (
	"context"
	storageInterface "decentralizedProject/storage"
	"fmt"
	"github.com/golang/geo/s1"
	"github.com/golang/geo/s2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx"
	"os"
	"strconv"
	"strings"
	"time"
)

var dbName = os.Getenv("MONGO_DBNAME")

const (
	collName      = "orders"
	earthRadiusKm = 6371.01
)

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
	size storageInterface.Size, pointStart string,
	pointEnd string, orderReadyDate storageInterface.ISOTimestamp) ([]storageInterface.OrderInfoForClient, error) {

	itemFast := storageInterface.Order{
		Id:             primitive.NewObjectID(),
		Weight:         weight,
		Size:           size,
		DeparturePoint: storageInterface.DeparturePoint(pointStart),
		ReceivePoint:   storageInterface.ReceivePoint(pointEnd),
		OrderReadyDate: orderReadyDate,
		Price:          CalculatePrice(pointStart, pointEnd, 1.0, weight, size),
		DeliveryDate:   CalculateDeliveryDate(1),
		OrderType:      storageInterface.OrderType("fast_Delivery"),
		OrderStatus:    "NotConfirmed",
	}
	itemSlow := storageInterface.Order{
		Id:             primitive.NewObjectID(),
		Weight:         weight,
		Size:           size,
		DeparturePoint: storageInterface.DeparturePoint(pointStart),
		ReceivePoint:   storageInterface.ReceivePoint(pointEnd),
		OrderReadyDate: orderReadyDate,
		Price:          CalculatePrice(pointStart, pointEnd, 0.0, weight, size),
		DeliveryDate:   CalculateDeliveryDate(0),
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

func CalculatePrice(pointStart string,
	pointEnd string, orderType float64, weight storageInterface.Weight, size storageInterface.Size) storageInterface.Price {
	//var from, to []string
	from := strings.Split(pointStart, ",")
	to := strings.Split(pointEnd, ",")

	fromLat, err := strconv.ParseFloat(from[0], 64)
	fromLon, err := strconv.ParseFloat(from[1], 64)

	toLat, err := strconv.ParseFloat(to[0], 64)
	toLon, err := strconv.ParseFloat(to[1], 64)
	if err != nil {
		return storageInterface.Price(-1)
	}
	fromPoint := s2.LatLngFromDegrees(fromLat, fromLon)
	toPoint := s2.LatLngFromDegrees(toLat, toLon)
	var result float64
	result = angleToKm(fromPoint.Distance(toPoint))
	return storageInterface.Price(result + orderType*500.0 + float64(weight) + float64(size)*2)
	//if from[0] == "" || to[1] == "" {
	//	fmt.Errorf("problems with parsing cords")
	//	return storageInterface.Price(-1)
	//}
	//return storageInterface.Price(fromLat + fromLon + toLon + toLat)
}

func kmToAngle(km float64) s1.Angle {
	return s1.Angle(km / earthRadiusKm)
}

func angleToKm(angle s1.Angle) float64 {
	return earthRadiusKm * float64(angle)
}

func CalculateDeliveryDate(orderType int) storageInterface.ISOTimestamp {
	if orderType == 1 {
		return storageInterface.ISOTimestamp(time.Now().Add(time.Hour * 24 * 5).UTC().Format(time.RFC3339))
	}
	return storageInterface.ISOTimestamp(time.Now().Add(time.Hour * 24 * 7).UTC().Format(time.RFC3339))
}
