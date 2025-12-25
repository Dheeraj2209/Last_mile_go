package storage

import (
	"context"
	"errors"

	lastmilev1 "github.com/Dheeraj2209/Last_mile_go/gen/go/lastmile/v1"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoStationStore struct {
	collection *mongo.Collection
}

func NewMongoStationStore(client *mongo.Client, dbName, collectionName string) *MongoStationStore {
	if client == nil {
		return nil
	}
	if dbName == "" {
		dbName = "lastmile"
	}
	if collectionName == "" {
		collectionName = "stations"
	}
	return &MongoStationStore{collection: client.Database(dbName).Collection(collectionName)}
}

func (s *MongoStationStore) Upsert(ctx context.Context, station *lastmilev1.Station) error {
	if station == nil || station.StationId == "" {
		return ErrInvalidArgument
	}
	doc := stationDoc{
		ID:            station.StationId,
		Name:          station.Name,
		Location:      toLatLngDoc(station.Location),
		NearbyAreaIDs: append([]string(nil), station.NearbyAreaIds...),
	}
	_, err := s.collection.UpdateOne(
		ctx,
		bson.M{"_id": station.StationId},
		bson.M{"$set": doc},
		options.Update().SetUpsert(true),
	)
	if err != nil {
		return err
	}
	return nil
}

func (s *MongoStationStore) Get(ctx context.Context, stationID string) (*lastmilev1.Station, error) {
	if stationID == "" {
		return nil, ErrInvalidArgument
	}
	var doc stationDoc
	err := s.collection.FindOne(ctx, bson.M{"_id": stationID}).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return doc.toStation(), nil
}

func (s *MongoStationStore) List(ctx context.Context, offset, limit int) ([]*lastmilev1.Station, int, error) {
	if offset < 0 || limit <= 0 {
		return nil, 0, ErrInvalidArgument
	}
	total, err := s.collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, 0, err
	}
	if offset >= int(total) {
		return nil, -1, nil
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "_id", Value: 1}}).
		SetSkip(int64(offset)).
		SetLimit(int64(limit))

	cursor, err := s.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	stations := make([]*lastmilev1.Station, 0, limit)
	for cursor.Next(ctx) {
		var doc stationDoc
		if err := cursor.Decode(&doc); err != nil {
			return nil, 0, err
		}
		stations = append(stations, doc.toStation())
	}
	if err := cursor.Err(); err != nil {
		return nil, 0, err
	}

	next := -1
	if offset+len(stations) < int(total) {
		next = offset + len(stations)
	}
	return stations, next, nil
}

type stationDoc struct {
	ID            string    `bson:"_id"`
	Name          string    `bson:"name"`
	Location      latLngDoc `bson:"location"`
	NearbyAreaIDs []string  `bson:"nearby_area_ids,omitempty"`
}

type latLngDoc struct {
	Latitude  float64 `bson:"latitude"`
	Longitude float64 `bson:"longitude"`
}

func toLatLngDoc(latlng *lastmilev1.LatLng) latLngDoc {
	if latlng == nil {
		return latLngDoc{}
	}
	return latLngDoc{
		Latitude:  latlng.Latitude,
		Longitude: latlng.Longitude,
	}
}

func (d stationDoc) toStation() *lastmilev1.Station {
	return &lastmilev1.Station{
		StationId: d.ID,
		Name:      d.Name,
		Location: &lastmilev1.LatLng{
			Latitude:  d.Location.Latitude,
			Longitude: d.Location.Longitude,
		},
		NearbyAreaIds: append([]string(nil), d.NearbyAreaIDs...),
	}
}
