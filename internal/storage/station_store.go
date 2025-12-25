package storage

import (
	"context"
	"sort"
	"sync"

	lastmilev1 "github.com/Dheeraj2209/Last_mile_go/gen/go/lastmile/v1"
)

type StationStore interface {
	Upsert(ctx context.Context, station *lastmilev1.Station) error
	Get(ctx context.Context, stationID string) (*lastmilev1.Station, error)
	List(ctx context.Context, offset, limit int) ([]*lastmilev1.Station, int, error)
}

type MemoryStationStore struct {
	mu       sync.RWMutex
	stations map[string]*lastmilev1.Station
}

func NewMemoryStationStore() *MemoryStationStore {
	return &MemoryStationStore{stations: make(map[string]*lastmilev1.Station)}
}

func (s *MemoryStationStore) Upsert(_ context.Context, station *lastmilev1.Station) error {
	if station == nil || station.StationId == "" {
		return ErrInvalidArgument
	}
	s.mu.Lock()
	s.stations[station.StationId] = cloneStation(station)
	s.mu.Unlock()
	return nil
}

func (s *MemoryStationStore) Get(_ context.Context, stationID string) (*lastmilev1.Station, error) {
	if stationID == "" {
		return nil, ErrInvalidArgument
	}
	s.mu.RLock()
	station, ok := s.stations[stationID]
	s.mu.RUnlock()
	if !ok {
		return nil, ErrNotFound
	}
	return cloneStation(station), nil
}

func (s *MemoryStationStore) List(_ context.Context, offset, limit int) ([]*lastmilev1.Station, int, error) {
	if offset < 0 || limit <= 0 {
		return nil, 0, ErrInvalidArgument
	}
	s.mu.RLock()
	ids := make([]string, 0, len(s.stations))
	for id := range s.stations {
		ids = append(ids, id)
	}
	s.mu.RUnlock()

	sort.Strings(ids)
	if offset >= len(ids) {
		return nil, -1, nil
	}

	end := offset + limit
	if end > len(ids) {
		end = len(ids)
	}

	stations := make([]*lastmilev1.Station, 0, end-offset)
	s.mu.RLock()
	for _, id := range ids[offset:end] {
		stations = append(stations, cloneStation(s.stations[id]))
	}
	s.mu.RUnlock()

	next := -1
	if end < len(ids) {
		next = end
	}
	return stations, next, nil
}

func cloneStation(station *lastmilev1.Station) *lastmilev1.Station {
	if station == nil {
		return nil
	}
	return &lastmilev1.Station{
		StationId:     station.StationId,
		Name:          station.Name,
		Location:      cloneLatLng(station.Location),
		NearbyAreaIds: append([]string(nil), station.NearbyAreaIds...),
	}
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
