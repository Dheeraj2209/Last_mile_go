package station

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"math"
	"strconv"
	"strings"

	lastmilev1 "github.com/Dheeraj2209/Last_mile_go/gen/go/lastmile/v1"
	"github.com/Dheeraj2209/Last_mile_go/internal/storage"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	lastmilev1.UnimplementedStationServiceServer
	store storage.StationStore
}

func NewServer() *Server {
	return NewServerWithStore(storage.NewMemoryStationStore())
}

func NewServerWithStore(store storage.StationStore) *Server {
	if store == nil {
		store = storage.NewMemoryStationStore()
	}
	return &Server{store: store}
}

func (s *Server) UpsertStation(ctx context.Context, req *lastmilev1.UpsertStationRequest) (*lastmilev1.UpsertStationResponse, error) {
	if req == nil || req.Station == nil {
		return nil, status.Error(codes.InvalidArgument, "station is required")
	}
	station := cloneStation(req.Station)
	station.Name = strings.TrimSpace(station.Name)
	if station.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}
	if err := validateLatLng(station.Location); err != nil {
		return nil, err
	}

	stationID := strings.TrimSpace(station.StationId)
	if stationID == "" {
		station.StationId = newID("station")
	} else {
		station.StationId = stationID
	}

	if err := s.store.Upsert(ctx, station); err != nil {
		if errors.Is(err, storage.ErrInvalidArgument) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, "storage error")
	}

	return &lastmilev1.UpsertStationResponse{Station: cloneStation(station)}, nil
}

func (s *Server) GetStation(ctx context.Context, req *lastmilev1.GetStationRequest) (*lastmilev1.GetStationResponse, error) {
	if req == nil || strings.TrimSpace(req.StationId) == "" {
		return nil, status.Error(codes.InvalidArgument, "station_id is required")
	}

	station, err := s.store.Get(ctx, strings.TrimSpace(req.StationId))
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "station not found")
		}
		if errors.Is(err, storage.ErrInvalidArgument) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, "storage error")
	}

	return &lastmilev1.GetStationResponse{Station: cloneStation(station)}, nil
}

func (s *Server) ListStations(ctx context.Context, req *lastmilev1.ListStationsRequest) (*lastmilev1.ListStationsResponse, error) {
	pageSize := int32(50)
	pageToken := "0"
	if req != nil {
		if req.PageSize < 0 {
			return nil, status.Error(codes.InvalidArgument, "page_size must be positive")
		}
		if req.PageSize > 0 {
			pageSize = req.PageSize
		}
		if req.PageToken != "" {
			pageToken = req.PageToken
		}
	}
	if pageSize > 100 {
		pageSize = 100
	}
	if pageSize <= 0 {
		return nil, status.Error(codes.InvalidArgument, "page_size must be positive")
	}

	offset, err := strconv.Atoi(pageToken)
	if err != nil || offset < 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid page_token")
	}

	stations, nextOffset, err := s.store.List(ctx, offset, int(pageSize))
	if err != nil {
		if errors.Is(err, storage.ErrInvalidArgument) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, "storage error")
	}

	nextToken := ""
	if nextOffset >= 0 {
		nextToken = strconv.Itoa(nextOffset)
	}

	return &lastmilev1.ListStationsResponse{Stations: stations, NextPageToken: nextToken}, nil
}

func cloneStation(station *lastmilev1.Station) *lastmilev1.Station {
	if station == nil {
		return nil
	}
	clone := &lastmilev1.Station{
		StationId:     station.StationId,
		Name:          station.Name,
		Location:      cloneLatLng(station.Location),
		NearbyAreaIds: append([]string(nil), station.NearbyAreaIds...),
	}
	return clone
}

func cloneLatLng(latlng *lastmilev1.LatLng) *lastmilev1.LatLng {
	if latlng == nil {
		return nil
	}
	return &lastmilev1.LatLng{
		Latitude:  latlng.Latitude,
		Longitude: latlng.Longitude,
	}
}

func newID(prefix string) string {
	buf := make([]byte, 8)
	_, _ = rand.Read(buf)
	return prefix + "_" + hex.EncodeToString(buf)
}

func validateLatLng(latlng *lastmilev1.LatLng) error {
	if latlng == nil {
		return status.Error(codes.InvalidArgument, "location is required")
	}
	if math.IsNaN(latlng.Latitude) || math.IsNaN(latlng.Longitude) {
		return status.Error(codes.InvalidArgument, "location has invalid coordinates")
	}
	if math.IsInf(latlng.Latitude, 0) || math.IsInf(latlng.Longitude, 0) {
		return status.Error(codes.InvalidArgument, "location has invalid coordinates")
	}
	if latlng.Latitude < -90 || latlng.Latitude > 90 {
		return status.Error(codes.InvalidArgument, "latitude out of range")
	}
	if latlng.Longitude < -180 || latlng.Longitude > 180 {
		return status.Error(codes.InvalidArgument, "longitude out of range")
	}
	return nil
}
