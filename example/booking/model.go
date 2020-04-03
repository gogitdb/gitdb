package booking

import (
	"time"

	"github.com/fobilow/gitdb/v2"
)

type BookingModel struct {
	//extends..
	gitdb.TimeStampedModel
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

//GetSchema example
func (b *BookingModel) GetSchema() *gitdb.Schema {

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

	return gitdb.NewSchema(name, block, record, indexes)
}

//GetLockFileNames example
//Todo add comment and date expansion function
func (b *BookingModel) GetLockFileNames() []string {
	var names []string
	names = append(names, "lock_"+b.CheckInDate.Format("2006-01-02")+"_"+b.RoomId)
	return names
}

//Validate example
func (b *BookingModel) Validate() error {
	//write validation logic here
	return nil
}

//IsLockable example
func (b *BookingModel) IsLockable() bool {
	return false
}

//ShouldEncrypt example
func (b *BookingModel) ShouldEncrypt() bool {
	return true
}
