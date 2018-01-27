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
	//write()
	//delete()

	db.Start(cfg)
	db.User = db.NewUser("dev", "dev@gitdb.io")
	db.StartGUI()
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
	bm.TimeStamp()

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
		fmt.Println("Found: " + r.Id())
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
