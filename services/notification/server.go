package notification

import lastmilev1 "github.com/Dheeraj2209/Last_mile_go/gen/go/lastmile/v1"

type Server struct {
	lastmilev1.UnimplementedNotificationServiceServer
}

func NewServer() *Server {
	return &Server{}
}
