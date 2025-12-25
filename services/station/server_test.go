package station

import (
	"context"
	"testing"

	lastmilev1 "github.com/Dheeraj2209/Last_mile_go/gen/go/lastmile/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestUpsertStationValidation(t *testing.T) {
	server := NewServer()
	station := &lastmilev1.Station{
		Name:     "Main",
		Location: &lastmilev1.LatLng{Latitude: 1, Longitude: 2},
	}
	cases := []struct {
		name string
		req  *lastmilev1.UpsertStationRequest
	}{
		{name: "nil request", req: nil},
		{name: "nil station", req: &lastmilev1.UpsertStationRequest{}},
		{name: "missing name", req: &lastmilev1.UpsertStationRequest{Station: &lastmilev1.Station{Location: station.Location}}},
		{name: "missing location", req: &lastmilev1.UpsertStationRequest{Station: &lastmilev1.Station{Name: station.Name}}},
		{name: "bad latitude", req: &lastmilev1.UpsertStationRequest{Station: &lastmilev1.Station{Name: station.Name, Location: &lastmilev1.LatLng{Latitude: 100, Longitude: 2}}}},
		{name: "bad longitude", req: &lastmilev1.UpsertStationRequest{Station: &lastmilev1.Station{Name: station.Name, Location: &lastmilev1.LatLng{Latitude: 1, Longitude: 200}}}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := server.UpsertStation(context.Background(), tc.req)
			assertStatusCode(t, err, codes.InvalidArgument)
		})
	}
}

func TestUpsertStationSuccess(t *testing.T) {
	server := NewServer()
	resp, err := server.UpsertStation(context.Background(), &lastmilev1.UpsertStationRequest{
		Station: &lastmilev1.Station{
			Name:      "  Central ",
			Location:  &lastmilev1.LatLng{Latitude: 12.3, Longitude: 45.6},
			StationId: " ",
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Station == nil {
		t.Fatalf("expected station in response")
	}
	if resp.Station.StationId == "" {
		t.Fatalf("expected generated station_id")
	}
	if resp.Station.Name != "Central" {
		t.Fatalf("expected trimmed name, got %q", resp.Station.Name)
	}
}

func TestGetStationErrors(t *testing.T) {
	server := NewServer()
	_, err := server.GetStation(context.Background(), nil)
	assertStatusCode(t, err, codes.InvalidArgument)

	_, err = server.GetStation(context.Background(), &lastmilev1.GetStationRequest{StationId: " "})
	assertStatusCode(t, err, codes.InvalidArgument)

	_, err = server.GetStation(context.Background(), &lastmilev1.GetStationRequest{StationId: "missing"})
	assertStatusCode(t, err, codes.NotFound)
}

func TestListStationsValidation(t *testing.T) {
	server := NewServer()
	_, err := server.ListStations(context.Background(), &lastmilev1.ListStationsRequest{PageSize: -1})
	assertStatusCode(t, err, codes.InvalidArgument)

	_, err = server.ListStations(context.Background(), &lastmilev1.ListStationsRequest{PageToken: "bad"})
	assertStatusCode(t, err, codes.InvalidArgument)
}

func TestListStationsPagination(t *testing.T) {
	server := NewServer()
	for _, st := range []*lastmilev1.Station{
		{StationId: "b", Name: "B", Location: &lastmilev1.LatLng{Latitude: 1, Longitude: 1}},
		{StationId: "a", Name: "A", Location: &lastmilev1.LatLng{Latitude: 1, Longitude: 1}},
		{StationId: "c", Name: "C", Location: &lastmilev1.LatLng{Latitude: 1, Longitude: 1}},
	} {
		if _, err := server.UpsertStation(context.Background(), &lastmilev1.UpsertStationRequest{Station: st}); err != nil {
			t.Fatalf("unexpected upsert error: %v", err)
		}
	}

	resp, err := server.ListStations(context.Background(), &lastmilev1.ListStationsRequest{PageSize: 2, PageToken: "0"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Stations) != 2 {
		t.Fatalf("expected 2 stations, got %d", len(resp.Stations))
	}
	if resp.Stations[0].StationId != "a" || resp.Stations[1].StationId != "b" {
		t.Fatalf("unexpected order: %q, %q", resp.Stations[0].StationId, resp.Stations[1].StationId)
	}
	if resp.NextPageToken != "2" {
		t.Fatalf("expected next token 2, got %q", resp.NextPageToken)
	}

	resp, err = server.ListStations(context.Background(), &lastmilev1.ListStationsRequest{PageSize: 2, PageToken: resp.NextPageToken})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Stations) != 1 {
		t.Fatalf("expected 1 station, got %d", len(resp.Stations))
	}
	if resp.Stations[0].StationId != "c" {
		t.Fatalf("expected station c, got %q", resp.Stations[0].StationId)
	}
	if resp.NextPageToken != "" {
		t.Fatalf("expected empty next token, got %q", resp.NextPageToken)
	}
}

func assertStatusCode(t *testing.T, err error, code codes.Code) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error with code %s", code.String())
	}
	statusErr, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected status error, got %v", err)
	}
	if statusErr.Code() != code {
		t.Fatalf("expected code %s, got %s", code.String(), statusErr.Code().String())
	}
}
