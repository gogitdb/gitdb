package main

import (
	"gitdb"
	"gitdb/example/booking"
)

func Make(modelName string) db.ModelInterface {
	var m db.ModelInterface
	switch modelName {
	case "Booking":
		m := &booking.BookingModel{}
		m.GetSchema().SetDef(m)
		break
	}

	return m
}
