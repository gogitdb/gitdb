package booking

type Status string

const (
	Cancelled  Status = "cancelled"
	CheckedOut Status = "checkedout"
	CheckedIn  Status = "checkedin"
	Booked     Status = "booked"
)

type PaymentMode string

const (
	Daily  PaymentMode = "daily"
	Hourly PaymentMode = "hourly"
)


type RoomType string

const (
	Room  RoomType = "room"
	Hall RoomType = "hall"
)