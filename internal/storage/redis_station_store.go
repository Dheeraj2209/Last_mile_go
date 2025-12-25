package storage

import (
	"context"
	"fmt"

	lastmilev1 "github.com/Dheeraj2209/Last_mile_go/gen/go/lastmile/v1"
	"github.com/redis/go-redis/v9"
	"google.golang.org/protobuf/encoding/protojson"
)

type RedisStationStore struct {
	client *redis.Client
	prefix string
}

func NewRedisStationStore(client *redis.Client, prefix string) *RedisStationStore {
	if client == nil {
		return nil
	}
	if prefix == "" {
		prefix = "lastmile"
	}
	return &RedisStationStore{client: client, prefix: prefix}
}

func (s *RedisStationStore) Upsert(ctx context.Context, station *lastmilev1.Station) error {
	if station == nil || station.StationId == "" {
		return ErrInvalidArgument
	}
	payload, err := protojson.Marshal(station)
	if err != nil {
		return err
	}
	pipe := s.client.TxPipeline()
	pipe.Set(ctx, s.stationKey(station.StationId), payload, 0)
	pipe.ZAdd(ctx, s.indexKey(), redis.Z{Score: 0, Member: station.StationId})
	_, err = pipe.Exec(ctx)
	return err
}

func (s *RedisStationStore) Get(ctx context.Context, stationID string) (*lastmilev1.Station, error) {
	if stationID == "" {
		return nil, ErrInvalidArgument
	}
	data, err := s.client.Get(ctx, s.stationKey(stationID)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, ErrNotFound
		}
		return nil, err
	}
	var station lastmilev1.Station
	if err := protojson.Unmarshal(data, &station); err != nil {
		return nil, err
	}
	return &station, nil
}

func (s *RedisStationStore) List(ctx context.Context, offset, limit int) ([]*lastmilev1.Station, int, error) {
	if offset < 0 || limit <= 0 {
		return nil, 0, ErrInvalidArgument
	}
	total, err := s.client.ZCard(ctx, s.indexKey()).Result()
	if err != nil {
		return nil, 0, err
	}
	if offset >= int(total) {
		return nil, -1, nil
	}

	stop := int64(offset + limit - 1)
	ids, err := s.client.ZRange(ctx, s.indexKey(), int64(offset), stop).Result()
	if err != nil {
		return nil, 0, err
	}
	if len(ids) == 0 {
		return nil, -1, nil
	}

	keys := make([]string, len(ids))
	for i, id := range ids {
		keys[i] = s.stationKey(id)
	}
	values, err := s.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, 0, err
	}

	stations := make([]*lastmilev1.Station, 0, len(values))
	for _, value := range values {
		if value == nil {
			continue
		}
		var data []byte
		switch v := value.(type) {
		case string:
			data = []byte(v)
		case []byte:
			data = v
		default:
			continue
		}
		var station lastmilev1.Station
		if err := protojson.Unmarshal(data, &station); err != nil {
			return nil, 0, err
		}
		stations = append(stations, &station)
	}

	next := -1
	if offset+len(ids) < int(total) {
		next = offset + len(ids)
	}
	return stations, next, nil
}

func (s *RedisStationStore) stationKey(stationID string) string {
	return fmt.Sprintf("%s:station:%s", s.prefix, stationID)
}

func (s *RedisStationStore) indexKey() string {
	return fmt.Sprintf("%s:stations", s.prefix)
}
