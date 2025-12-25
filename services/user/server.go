package user

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"

	lastmilev1 "github.com/Dheeraj2209/Last_mile_go/gen/go/lastmile/v1"
	"github.com/Dheeraj2209/Last_mile_go/internal/storage"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	lastmilev1.UnimplementedUserServiceServer
	riders  storage.RiderStore
	drivers storage.DriverStore
}

func NewServer() *Server {
	mem := storage.NewMemoryUserStore()
	return NewServerWithStores(mem, mem)
}

func NewServerWithStores(riders storage.RiderStore, drivers storage.DriverStore) *Server {
	if riders == nil || drivers == nil {
		mem := storage.NewMemoryUserStore()
		riders = mem
		drivers = mem
	}
	return &Server{
		riders:  riders,
		drivers: drivers,
	}
}

func (s *Server) CreateRiderProfile(ctx context.Context, req *lastmilev1.CreateRiderProfileRequest) (*lastmilev1.CreateRiderProfileResponse, error) {
	if req == nil || req.Profile == nil {
		return nil, status.Error(codes.InvalidArgument, "profile is required")
	}
	profile := cloneRiderProfile(req.Profile)
	name := strings.TrimSpace(profile.Name)
	if name == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}
	phone := strings.TrimSpace(profile.Phone)
	if phone == "" {
		return nil, status.Error(codes.InvalidArgument, "phone is required")
	}
	profile.Name = name
	profile.Phone = phone

	riderID := strings.TrimSpace(profile.RiderId)
	if riderID == "" {
		profile.RiderId = newID("rider")
	} else {
		profile.RiderId = riderID
	}

	if err := s.riders.CreateRider(ctx, profile); err != nil {
		if errors.Is(err, storage.ErrAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, "rider already exists")
		}
		if errors.Is(err, storage.ErrInvalidArgument) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, "storage error")
	}

	return &lastmilev1.CreateRiderProfileResponse{Profile: cloneRiderProfile(profile)}, nil
}

func (s *Server) CreateDriverProfile(ctx context.Context, req *lastmilev1.CreateDriverProfileRequest) (*lastmilev1.CreateDriverProfileResponse, error) {
	if req == nil || req.Profile == nil {
		return nil, status.Error(codes.InvalidArgument, "profile is required")
	}
	profile := cloneDriverProfile(req.Profile)
	name := strings.TrimSpace(profile.Name)
	if name == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}
	phone := strings.TrimSpace(profile.Phone)
	if phone == "" {
		return nil, status.Error(codes.InvalidArgument, "phone is required")
	}
	vehicleID := strings.TrimSpace(profile.VehicleId)
	if vehicleID == "" {
		return nil, status.Error(codes.InvalidArgument, "vehicle_id is required")
	}
	profile.Name = name
	profile.Phone = phone
	profile.VehicleId = vehicleID

	driverID := strings.TrimSpace(profile.DriverId)
	if driverID == "" {
		profile.DriverId = newID("driver")
	} else {
		profile.DriverId = driverID
	}

	if err := s.drivers.CreateDriver(ctx, profile); err != nil {
		if errors.Is(err, storage.ErrAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, "driver already exists")
		}
		if errors.Is(err, storage.ErrInvalidArgument) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, "storage error")
	}

	return &lastmilev1.CreateDriverProfileResponse{Profile: cloneDriverProfile(profile)}, nil
}

func (s *Server) GetRiderProfile(ctx context.Context, req *lastmilev1.GetRiderProfileRequest) (*lastmilev1.GetRiderProfileResponse, error) {
	if req == nil || strings.TrimSpace(req.RiderId) == "" {
		return nil, status.Error(codes.InvalidArgument, "rider_id is required")
	}

	profile, err := s.riders.GetRider(ctx, strings.TrimSpace(req.RiderId))
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "rider not found")
		}
		if errors.Is(err, storage.ErrInvalidArgument) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, "storage error")
	}

	return &lastmilev1.GetRiderProfileResponse{Profile: cloneRiderProfile(profile)}, nil
}

func (s *Server) GetDriverProfile(ctx context.Context, req *lastmilev1.GetDriverProfileRequest) (*lastmilev1.GetDriverProfileResponse, error) {
	if req == nil || strings.TrimSpace(req.DriverId) == "" {
		return nil, status.Error(codes.InvalidArgument, "driver_id is required")
	}

	profile, err := s.drivers.GetDriver(ctx, strings.TrimSpace(req.DriverId))
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "driver not found")
		}
		if errors.Is(err, storage.ErrInvalidArgument) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, "storage error")
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
