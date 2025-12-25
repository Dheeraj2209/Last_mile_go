package user

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"sync"

	lastmilev1 "github.com/Dheeraj2209/Last_mile_go/gen/go/lastmile/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	lastmilev1.UnimplementedUserServiceServer
	mu      sync.RWMutex
	riders  map[string]*lastmilev1.RiderProfile
	drivers map[string]*lastmilev1.DriverProfile
}

func NewServer() *Server {
	return &Server{
		riders:  make(map[string]*lastmilev1.RiderProfile),
		drivers: make(map[string]*lastmilev1.DriverProfile),
	}
}

func (s *Server) CreateRiderProfile(_ context.Context, req *lastmilev1.CreateRiderProfileRequest) (*lastmilev1.CreateRiderProfileResponse, error) {
	if req == nil || req.Profile == nil {
		return nil, status.Error(codes.InvalidArgument, "profile is required")
	}
	profile := cloneRiderProfile(req.Profile)
	if profile.RiderId == "" {
		profile.RiderId = newID("rider")
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.riders[profile.RiderId]; exists {
		return nil, status.Error(codes.AlreadyExists, "rider already exists")
	}
	s.riders[profile.RiderId] = profile

	return &lastmilev1.CreateRiderProfileResponse{Profile: cloneRiderProfile(profile)}, nil
}

func (s *Server) CreateDriverProfile(_ context.Context, req *lastmilev1.CreateDriverProfileRequest) (*lastmilev1.CreateDriverProfileResponse, error) {
	if req == nil || req.Profile == nil {
		return nil, status.Error(codes.InvalidArgument, "profile is required")
	}
	profile := cloneDriverProfile(req.Profile)
	if profile.DriverId == "" {
		profile.DriverId = newID("driver")
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.drivers[profile.DriverId]; exists {
		return nil, status.Error(codes.AlreadyExists, "driver already exists")
	}
	s.drivers[profile.DriverId] = profile

	return &lastmilev1.CreateDriverProfileResponse{Profile: cloneDriverProfile(profile)}, nil
}

func (s *Server) GetRiderProfile(_ context.Context, req *lastmilev1.GetRiderProfileRequest) (*lastmilev1.GetRiderProfileResponse, error) {
	if req == nil || req.RiderId == "" {
		return nil, status.Error(codes.InvalidArgument, "rider_id is required")
	}

	s.mu.RLock()
	profile, ok := s.riders[req.RiderId]
	s.mu.RUnlock()
	if !ok {
		return nil, status.Error(codes.NotFound, "rider not found")
	}

	return &lastmilev1.GetRiderProfileResponse{Profile: cloneRiderProfile(profile)}, nil
}

func (s *Server) GetDriverProfile(_ context.Context, req *lastmilev1.GetDriverProfileRequest) (*lastmilev1.GetDriverProfileResponse, error) {
	if req == nil || req.DriverId == "" {
		return nil, status.Error(codes.InvalidArgument, "driver_id is required")
	}

	s.mu.RLock()
	profile, ok := s.drivers[req.DriverId]
	s.mu.RUnlock()
	if !ok {
		return nil, status.Error(codes.NotFound, "driver not found")
	}

	return &lastmilev1.GetDriverProfileResponse{Profile: cloneDriverProfile(profile)}, nil
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

func newID(prefix string) string {
	buf := make([]byte, 8)
	_, _ = rand.Read(buf)
	return prefix + "_" + hex.EncodeToString(buf)
}
