package booking

import (
	"gitdb"
)

type BookingModel struct {
	//extends..
	db.BaseModel
	BookingSchema
}

func NewBookingModel() *BookingModel {
	bm := &BookingModel{}

	bm.Init(bm)
	return bm
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

	//TODO move this to a better place
	//timestamp the data


	return true
}
