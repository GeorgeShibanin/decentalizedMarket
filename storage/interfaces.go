package storage

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type OrderId string

type Weight float64
type Size float64
type DeparturePoint string
type ReceivePoint string

type Price float64
type OrderType string
type ISOTimestamp string

var (
	StorageError = errors.New("storage")
	//ErrCollision = fmt.Errorf("%w.collision", StorageError)
	//ErrNotFound  = fmt.Errorf("%w.not_found", StorageError)
)

type Order struct {
	Id primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`

	Weight         Weight         `json:"weight" bson:"weight"`
	Size           Size           `json:"volume" bson:"volume"`
	DeparturePoint DeparturePoint `json:"from" bson:"from"`
	ReceivePoint   ReceivePoint   `json:"to" bson:"to"`
	OrderReadyDate ISOTimestamp   `json:"time" bson:"time"`

	Price        Price        `json:"price" bson:"price"`
	DeliveryDate ISOTimestamp `json:"deliveryDate" bson:"deliveryDate"`
	OrderType    OrderType    `json:"OrderType" bson:"OrderType"`
	OrderStatus  string       `json:"orderStatus" bson:"orderStatus"`
}

type OrderInfoForClient struct {
	Id           primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Price        Price              `json:"price" bson:"price"`
	DeliveryDate ISOTimestamp       `json:"deliveryDate" bson:"deliveryDate"`
	OrderType    OrderType          `json:"OrderType" bson:"OrderType"`
}

type Storage interface {
	GetOrder(ctx context.Context, weight Weight, size Size,
		pointStart DeparturePoint, pointEnd ReceivePoint, orderReadyDate ISOTimestamp) ([]OrderInfoForClient, error)
	AcceptDelivery(ctx context.Context, id OrderId) (string, error)
}
