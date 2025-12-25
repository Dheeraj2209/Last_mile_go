package storage

import (
	"context"
	"errors"

	lastmilev1 "github.com/Dheeraj2209/Last_mile_go/gen/go/lastmile/v1"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoUserStore struct {
	riders  *mongo.Collection
	drivers *mongo.Collection
}

func NewMongoUserStore(client *mongo.Client, dbName, riderCollection, driverCollection string) *MongoUserStore {
	if client == nil {
		return nil
	}
	if dbName == "" {
		dbName = "lastmile"
	}
	if riderCollection == "" {
		riderCollection = "riders"
	}
	if driverCollection == "" {
		driverCollection = "drivers"
	}
	db := client.Database(dbName)
	return &MongoUserStore{
		riders:  db.Collection(riderCollection),
		drivers: db.Collection(driverCollection),
	}
}

func (s *MongoUserStore) CreateRider(ctx context.Context, profile *lastmilev1.RiderProfile) error {
	if profile == nil || profile.RiderId == "" {
		return ErrInvalidArgument
	}
	doc := riderDoc{
		ID:    profile.RiderId,
		Name:  profile.Name,
		Phone: profile.Phone,
	}
	_, err := s.riders.InsertOne(ctx, doc)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return ErrAlreadyExists
		}
		return err
	}
	return nil
}

func (s *MongoUserStore) GetRider(ctx context.Context, riderID string) (*lastmilev1.RiderProfile, error) {
	if riderID == "" {
		return nil, ErrInvalidArgument
	}
	var doc riderDoc
	err := s.riders.FindOne(ctx, bson.M{"_id": riderID}).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &lastmilev1.RiderProfile{
		RiderId: doc.ID,
		Name:    doc.Name,
		Phone:   doc.Phone,
	}, nil
}

func (s *MongoUserStore) CreateDriver(ctx context.Context, profile *lastmilev1.DriverProfile) error {
	if profile == nil || profile.DriverId == "" {
		return ErrInvalidArgument
	}
	doc := driverDoc{
		ID:        profile.DriverId,
		Name:      profile.Name,
		Phone:     profile.Phone,
		VehicleID: profile.VehicleId,
	}
	_, err := s.drivers.InsertOne(ctx, doc)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return ErrAlreadyExists
		}
		return err
	}
	return nil
}

func (s *MongoUserStore) GetDriver(ctx context.Context, driverID string) (*lastmilev1.DriverProfile, error) {
	if driverID == "" {
		return nil, ErrInvalidArgument
	}
	var doc driverDoc
	err := s.drivers.FindOne(ctx, bson.M{"_id": driverID}).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &lastmilev1.DriverProfile{
		DriverId:  doc.ID,
		Name:      doc.Name,
		Phone:     doc.Phone,
		VehicleId: doc.VehicleID,
	}, nil
}

type riderDoc struct {
	ID    string `bson:"_id"`
	Name  string `bson:"name"`
	Phone string `bson:"phone"`
}

type driverDoc struct {
	ID        string `bson:"_id"`
	Name      string `bson:"name"`
	Phone     string `bson:"phone"`
	VehicleID string `bson:"vehicle_id"`
}
