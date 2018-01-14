package booking

import (
	"gitdb"
	"time"
)

type BookingDefinition struct {
	db.BaseDef
	Type         RoomType
	CheckInDate  time.Time
	CheckOutDate time.Time
	CheckedOutAt time.Time
	Guests       int
	CardsIssued  int
	Status       Status
	PaymentMode  PaymentMode
	RoomPrice    float64
	RoomId       string
	CustomerId   string
	UserId       string
	NextOfKin    string
	Purpose      string
}

//ID: Booking/201801/room_201801101100

//Name of schema
func (s *BookingDefinition) Name() string {
	return "Booking"
}

//Block of schema
func (s *BookingDefinition) Block() string {
	return s.CreatedAt.Format("200601")
}

//Record of schema
func (s *BookingDefinition) Record() string {
	return string(s.Type) + "_" + s.CreatedAt.Format("200601021504")
}

//Indexes speed up searching
func (s *BookingDefinition) Indexes() map[string]interface{} {
	indexes := make(map[string]interface{})

	indexes["RoomId"] = s.RoomId
	indexes["Guests"] = s.Guests
	indexes["CustomerId"] = s.CustomerId
	indexes["CreationDate"] = s.CreatedAt.Format("2006-01-02")

	return indexes
}
