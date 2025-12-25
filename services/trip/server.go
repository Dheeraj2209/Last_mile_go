package trip

import lastmilev1 "github.com/Dheeraj2209/Last_mile_go/gen/go/lastmile/v1"

type Server struct {
	lastmilev1.UnimplementedTripServiceServer
}

func NewServer() *Server {
	return &Server{}
}
