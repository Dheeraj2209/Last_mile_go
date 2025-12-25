package user

import (
	"context"
	"testing"

	lastmilev1 "github.com/Dheeraj2209/Last_mile_go/gen/go/lastmile/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestCreateRiderProfileValidation(t *testing.T) {
	server := NewServer()
	cases := []struct {
		name string
		req  *lastmilev1.CreateRiderProfileRequest
	}{
		{name: "nil request", req: nil},
		{name: "nil profile", req: &lastmilev1.CreateRiderProfileRequest{}},
		{name: "missing name", req: &lastmilev1.CreateRiderProfileRequest{Profile: &lastmilev1.RiderProfile{Phone: "123"}}},
		{name: "missing phone", req: &lastmilev1.CreateRiderProfileRequest{Profile: &lastmilev1.RiderProfile{Name: "Rita"}}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := server.CreateRiderProfile(context.Background(), tc.req)
			assertStatusCode(t, err, codes.InvalidArgument)
		})
	}
}

func TestCreateRiderProfileSuccess(t *testing.T) {
	server := NewServer()
	resp, err := server.CreateRiderProfile(context.Background(), &lastmilev1.CreateRiderProfileRequest{
		Profile: &lastmilev1.RiderProfile{
			Name:    "  Alice  ",
			Phone:   " 555 ",
			RiderId: " ",
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Profile == nil {
		t.Fatalf("expected profile in response")
	}
	if resp.Profile.RiderId == "" {
		t.Fatalf("expected generated rider_id")
	}
	if resp.Profile.Name != "Alice" {
		t.Fatalf("expected trimmed name, got %q", resp.Profile.Name)
	}
	if resp.Profile.Phone != "555" {
		t.Fatalf("expected trimmed phone, got %q", resp.Profile.Phone)
	}
}

func TestCreateRiderProfileDuplicate(t *testing.T) {
	server := NewServer()
	req := &lastmilev1.CreateRiderProfileRequest{
		Profile: &lastmilev1.RiderProfile{
			RiderId: "r1",
			Name:    "Rita",
			Phone:   "123",
		},
	}
	if _, err := server.CreateRiderProfile(context.Background(), req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err := server.CreateRiderProfile(context.Background(), req)
	assertStatusCode(t, err, codes.AlreadyExists)
}

func TestGetRiderProfileErrors(t *testing.T) {
	server := NewServer()
	_, err := server.GetRiderProfile(context.Background(), nil)
	assertStatusCode(t, err, codes.InvalidArgument)

	_, err = server.GetRiderProfile(context.Background(), &lastmilev1.GetRiderProfileRequest{RiderId: " "})
	assertStatusCode(t, err, codes.InvalidArgument)

	_, err = server.GetRiderProfile(context.Background(), &lastmilev1.GetRiderProfileRequest{RiderId: "missing"})
	assertStatusCode(t, err, codes.NotFound)
}

func TestCreateDriverProfileValidation(t *testing.T) {
	server := NewServer()
	cases := []struct {
		name string
		req  *lastmilev1.CreateDriverProfileRequest
	}{
		{name: "nil request", req: nil},
		{name: "nil profile", req: &lastmilev1.CreateDriverProfileRequest{}},
		{name: "missing name", req: &lastmilev1.CreateDriverProfileRequest{Profile: &lastmilev1.DriverProfile{Phone: "123", VehicleId: "v1"}}},
		{name: "missing phone", req: &lastmilev1.CreateDriverProfileRequest{Profile: &lastmilev1.DriverProfile{Name: "Dana", VehicleId: "v1"}}},
		{name: "missing vehicle", req: &lastmilev1.CreateDriverProfileRequest{Profile: &lastmilev1.DriverProfile{Name: "Dana", Phone: "123"}}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := server.CreateDriverProfile(context.Background(), tc.req)
			assertStatusCode(t, err, codes.InvalidArgument)
		})
	}
}

func TestCreateDriverProfileSuccess(t *testing.T) {
	server := NewServer()
	resp, err := server.CreateDriverProfile(context.Background(), &lastmilev1.CreateDriverProfileRequest{
		Profile: &lastmilev1.DriverProfile{
			Name:      "  Dana ",
			Phone:     " 999 ",
			VehicleId: " van-1 ",
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Profile == nil {
		t.Fatalf("expected profile in response")
	}
	if resp.Profile.DriverId == "" {
		t.Fatalf("expected generated driver_id")
	}
	if resp.Profile.Name != "Dana" {
		t.Fatalf("expected trimmed name, got %q", resp.Profile.Name)
	}
	if resp.Profile.Phone != "999" {
		t.Fatalf("expected trimmed phone, got %q", resp.Profile.Phone)
	}
	if resp.Profile.VehicleId != "van-1" {
		t.Fatalf("expected trimmed vehicle_id, got %q", resp.Profile.VehicleId)
	}
}

func TestGetDriverProfileErrors(t *testing.T) {
	server := NewServer()
	_, err := server.GetDriverProfile(context.Background(), nil)
	assertStatusCode(t, err, codes.InvalidArgument)

	_, err = server.GetDriverProfile(context.Background(), &lastmilev1.GetDriverProfileRequest{DriverId: " "})
	assertStatusCode(t, err, codes.InvalidArgument)

	_, err = server.GetDriverProfile(context.Background(), &lastmilev1.GetDriverProfileRequest{DriverId: "missing"})
	assertStatusCode(t, err, codes.NotFound)
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
