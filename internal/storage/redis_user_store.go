package storage

import (
	"context"
	"fmt"

	lastmilev1 "github.com/Dheeraj2209/Last_mile_go/gen/go/lastmile/v1"
	"github.com/redis/go-redis/v9"
	"google.golang.org/protobuf/encoding/protojson"
)

type RedisUserStore struct {
	client *redis.Client
	prefix string
}

func NewRedisUserStore(client *redis.Client, prefix string) *RedisUserStore {
	if client == nil {
		return nil
	}
	if prefix == "" {
		prefix = "lastmile"
	}
	return &RedisUserStore{client: client, prefix: prefix}
}

func (s *RedisUserStore) CreateRider(ctx context.Context, profile *lastmilev1.RiderProfile) error {
	if profile == nil || profile.RiderId == "" {
		return ErrInvalidArgument
	}
	key := s.riderKey(profile.RiderId)
	payload, err := protojson.Marshal(profile)
	if err != nil {
		return err
	}
	ok, err := s.client.SetNX(ctx, key, payload, 0).Result()
	if err != nil {
		return err
	}
	if !ok {
		return ErrAlreadyExists
	}
	return nil
}

func (s *RedisUserStore) GetRider(ctx context.Context, riderID string) (*lastmilev1.RiderProfile, error) {
	if riderID == "" {
		return nil, ErrInvalidArgument
	}
	data, err := s.client.Get(ctx, s.riderKey(riderID)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, ErrNotFound
		}
		return nil, err
	}
	var profile lastmilev1.RiderProfile
	if err := protojson.Unmarshal(data, &profile); err != nil {
		return nil, err
	}
	return &profile, nil
}

func (s *RedisUserStore) CreateDriver(ctx context.Context, profile *lastmilev1.DriverProfile) error {
	if profile == nil || profile.DriverId == "" {
		return ErrInvalidArgument
	}
	key := s.driverKey(profile.DriverId)
	payload, err := protojson.Marshal(profile)
	if err != nil {
		return err
	}
	ok, err := s.client.SetNX(ctx, key, payload, 0).Result()
	if err != nil {
		return err
	}
	if !ok {
		return ErrAlreadyExists
	}
	return nil
}

func (s *RedisUserStore) GetDriver(ctx context.Context, driverID string) (*lastmilev1.DriverProfile, error) {
	if driverID == "" {
		return nil, ErrInvalidArgument
	}
	data, err := s.client.Get(ctx, s.driverKey(driverID)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, ErrNotFound
		}
		return nil, err
	}
	var profile lastmilev1.DriverProfile
	if err := protojson.Unmarshal(data, &profile); err != nil {
		return nil, err
	}
	return &profile, nil
}

func (s *RedisUserStore) riderKey(riderID string) string {
	return fmt.Sprintf("%s:rider:%s", s.prefix, riderID)
}

func (s *RedisUserStore) driverKey(driverID string) string {
	return fmt.Sprintf("%s:driver:%s", s.prefix, driverID)
}
