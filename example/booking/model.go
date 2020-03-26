package booking

import (
	"time"

	db "github.com/fobilow/gitdb"
)

type BookingModel struct {
	//extends..
	db.BaseModel
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
	AutoId       int64
}

func NewBookingModel() *BookingModel {
	return &BookingModel{}
}

func (b *BookingModel) GetSchema() *db.Schema {

	//Name of schema
	name := func() string {
		return "Booking"
	}

	//Block of schema
	block := func() string {
		return b.CreatedAt.Format("200601")
	}

	//Record of schema
	record := func() string {
		return string(b.Type) + "_" + b.CreatedAt.Format("20060102150405")
	}

	//Indexes speed up searching
	indexes := func() map[string]interface{} {
		indexes := make(map[string]interface{})

		indexes["RoomId"] = b.RoomId
		indexes["Guests"] = b.Guests
		indexes["CustomerId"] = b.CustomerId
		indexes["CreationDate"] = b.CreatedAt.Format("2006-01-02")

		return indexes
	}

	return db.NewSchema(name, block, record, indexes)
}

func (b *BookingModel) GetLockFileNames() []string {
	var names []string
	names = append(names, "lock_"+b.CheckInDate.Format("2006-01-02")+"_"+b.RoomId)
	return names
}

func (b *BookingModel) NumberOfHours() int {
	return int(b.CheckOutDate.Sub(b.CheckInDate).Hours())
}

func (b *BookingModel) NumberOfNights() int {
	n := int(b.CheckOutDate.Sub(b.CheckInDate).Hours() / 24)
	if n <= 0 {
		n = 1
	}

	return n
}

func (b *BookingModel) Validate() bool {
	//write validation logic here
	return true
}

func (b *BookingModel) ShouldEncrypt() bool {
	return true
}
