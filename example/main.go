package main

import (
	"fmt"
	"gitdb"
	"gitdb/example/booking"
	"time"
)

var cfg *db.Config

func init() {
	cfg = &db.Config{
		DbPath:         "./data",
		OfflineRepoDir: "./Repo/app.git",
		OnlineRemote:   "",
		OfflineRemote:  "",
		SshKey:         "",
		Factory:        Make,
	}
}

func main() {
	write()
	//delete()
	//search()
	fetch()

	//db.Start(cfg)
	//db.User = db.NewUser("dev", "dev@gitdb.io")
	//db.StartGUI()
}

func write() {
	bm := booking.NewBookingModel()

	db.Start(cfg)
	db.User = db.NewUser("dev", "dev@gitdb.io")

	//populate model
	bm.Type = booking.Room
	bm.CheckInDate = time.Now()
	bm.CheckOutDate = time.Now()
	bm.CustomerId = "customer_1"
	bm.Guests = 2
	bm.CardsIssued = 1
	bm.NextOfKin = "Kin 1"
	bm.Status = booking.CheckedIn
	bm.UserId = "user_1"
	bm.RoomId = "room_1"

	err := db.Insert(bm)
	fmt.Println(err)

}

func read() {
	db.Start(cfg)
	db.User = db.NewUser("dev", "dev@gitdb.io")

	r, err := db.Get("Booking/201801/room_201801111512")
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(r.GetID())
	}
}

func delete() {
	db.Start(cfg)
	db.User = db.NewUser("dev", "dev@gitdb.io")

	r, err := db.Delete("Booking/201801/room_201801111823")
	if err != nil {
		fmt.Println(err.Error())
	} else {
		if r {
			fmt.Println("Deleted")
		} else {
			fmt.Println("NOT Deleted")
		}
	}
}

func search() {
	db.Start(cfg)
	db.User = db.NewUser("dev", "dev@gitdb.io")

	rows, err := db.Search("Booking", []string{"CustomerId"}, []string{"customer_2"}, db.SEARCH_MODE_EQUALS)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		for _, r := range rows {
			fmt.Println(r)
		}
	}
}

func fetch() {
	db.Start(cfg)
	db.User = db.NewUser("dev", "dev@gitdb.io")

	rows, err := db.Fetch("Booking")
	if err != nil {
		fmt.Println(err.Error())
	} else {
		for _, r := range rows {
			fmt.Println(r)
		}
	}
}

func Make(modelName string) db.ModelSchema {
	var m db.ModelSchema
	switch modelName {
	case "Booking":
		m := &booking.BookingModel{}
		m.Init(m)
		return m
		break
	}

	return m
}
