package station

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"sort"
	"strconv"
	"sync"

	lastmilev1 "github.com/Dheeraj2209/Last_mile_go/gen/go/lastmile/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	lastmilev1.UnimplementedStationServiceServer
	mu       sync.RWMutex
	stations map[string]*lastmilev1.Station
}

func NewServer() *Server {
	return &Server{stations: make(map[string]*lastmilev1.Station)}
}

func (s *Server) UpsertStation(_ context.Context, req *lastmilev1.UpsertStationRequest) (*lastmilev1.UpsertStationResponse, error) {
	if req == nil || req.Station == nil {
		return nil, status.Error(codes.InvalidArgument, "station is required")
	}
	station := cloneStation(req.Station)
	if station.StationId == "" {
		station.StationId = newID("station")
	}

	s.mu.Lock()
	s.stations[station.StationId] = station
	s.mu.Unlock()

	return &lastmilev1.UpsertStationResponse{Station: cloneStation(station)}, nil
}

func (s *Server) GetStation(_ context.Context, req *lastmilev1.GetStationRequest) (*lastmilev1.GetStationResponse, error) {
	if req == nil || req.StationId == "" {
		return nil, status.Error(codes.InvalidArgument, "station_id is required")
	}

	s.mu.RLock()
	station, ok := s.stations[req.StationId]
	s.mu.RUnlock()
	if !ok {
		return nil, status.Error(codes.NotFound, "station not found")
	}

	return &lastmilev1.GetStationResponse{Station: cloneStation(station)}, nil
}

func (s *Server) ListStations(_ context.Context, req *lastmilev1.ListStationsRequest) (*lastmilev1.ListStationsResponse, error) {
	pageSize := int32(50)
	pageToken := "0"
	if req != nil {
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

	s.mu.RLock()
	ids := make([]string, 0, len(s.stations))
	for id := range s.stations {
		ids = append(ids, id)
	}
	s.mu.RUnlock()

	sort.Strings(ids)
	if offset >= len(ids) {
		return &lastmilev1.ListStationsResponse{Stations: nil, NextPageToken: ""}, nil
	}

	end := offset + int(pageSize)
	if end > len(ids) {
		end = len(ids)
	}

	stations := make([]*lastmilev1.Station, 0, end-offset)
	s.mu.RLock()
	for _, id := range ids[offset:end] {
		stations = append(stations, cloneStation(s.stations[id]))
	}
	s.mu.RUnlock()

	nextToken := ""
	if end < len(ids) {
		nextToken = strconv.Itoa(end)
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
