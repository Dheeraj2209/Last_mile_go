package storage

import (
	"context"
	"sync"

	lastmilev1 "github.com/Dheeraj2209/Last_mile_go/gen/go/lastmile/v1"
)

type RiderStore interface {
	CreateRider(ctx context.Context, profile *lastmilev1.RiderProfile) error
	GetRider(ctx context.Context, riderID string) (*lastmilev1.RiderProfile, error)
}

type DriverStore interface {
	CreateDriver(ctx context.Context, profile *lastmilev1.DriverProfile) error
	GetDriver(ctx context.Context, driverID string) (*lastmilev1.DriverProfile, error)
}

type MemoryUserStore struct {
	mu      sync.RWMutex
	riders  map[string]*lastmilev1.RiderProfile
	drivers map[string]*lastmilev1.DriverProfile
}

func NewMemoryUserStore() *MemoryUserStore {
	return &MemoryUserStore{
		riders:  make(map[string]*lastmilev1.RiderProfile),
		drivers: make(map[string]*lastmilev1.DriverProfile),
	}
}

func (s *MemoryUserStore) CreateRider(_ context.Context, profile *lastmilev1.RiderProfile) error {
	if profile == nil || profile.RiderId == "" {
		return ErrInvalidArgument
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.riders[profile.RiderId]; exists {
		return ErrAlreadyExists
	}
	s.riders[profile.RiderId] = cloneRiderProfile(profile)
	return nil
}

func (s *MemoryUserStore) GetRider(_ context.Context, riderID string) (*lastmilev1.RiderProfile, error) {
	if riderID == "" {
		return nil, ErrInvalidArgument
	}
	s.mu.RLock()
	profile, ok := s.riders[riderID]
	s.mu.RUnlock()
	if !ok {
		return nil, ErrNotFound
	}
	return cloneRiderProfile(profile), nil
}

func (s *MemoryUserStore) CreateDriver(_ context.Context, profile *lastmilev1.DriverProfile) error {
	if profile == nil || profile.DriverId == "" {
		return ErrInvalidArgument
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.drivers[profile.DriverId]; exists {
		return ErrAlreadyExists
	}
	s.drivers[profile.DriverId] = cloneDriverProfile(profile)
	return nil
}

func (s *MemoryUserStore) GetDriver(_ context.Context, driverID string) (*lastmilev1.DriverProfile, error) {
	if driverID == "" {
		return nil, ErrInvalidArgument
	}
	s.mu.RLock()
	profile, ok := s.drivers[driverID]
	s.mu.RUnlock()
	if !ok {
		return nil, ErrNotFound
	}
	return cloneDriverProfile(profile), nil
}

func cloneRiderProfile(profile *lastmilev1.RiderProfile) *lastmilev1.RiderProfile {
	if profile == nil {
		return nil
	}
	return &lastmilev1.RiderProfile{
		RiderId: profile.RiderId,
		Name:    profile.Name,
		Phone:   profile.Phone,
	}
}

func cloneDriverProfile(profile *lastmilev1.DriverProfile) *lastmilev1.DriverProfile {
	if profile == nil {
		return nil
	}
	return &lastmilev1.DriverProfile{
		DriverId:  profile.DriverId,
		Name:      profile.Name,
		Phone:     profile.Phone,
		VehicleId: profile.VehicleId,
	}
}
