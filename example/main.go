package main

import (
	"fmt"
	"time"

	db "github.com/fobilow/gitdb"
	"github.com/fobilow/gitdb/example/booking"
	"log"
	"os"
)

var cfg *db.Config
var logToFile bool

func init() {
	cfg = &db.Config{
		DbPath:         "./data",
		OnlineRemote:   os.Getenv("GITDB_REPO"),
		SshKey:         os.Getenv("GITDB_SSH_KEY"),
		Factory:        make,
		SyncInterval:   time.Second * 5,
		EncryptionKey:  "put_your_encryption_key_here",
		//GitDriver: &db.GitBinary{},
		GitDriver: &db.GoGit{},
	}

	if logToFile {
		runLogFile, err := os.OpenFile("./db.log", os.O_WRONLY | os.O_CREATE | os.O_APPEND, 0666)
		if err == nil {
			cfg.Logger = log.New(runLogFile, "GITDB: ", log.Ldate|log.Ltime|log.Lshortfile)
		}
	}

	db.User = db.NewUser("dev", "dev@gitdb.io")
	db.Start(cfg)
}

func main() {
	testWrite()
}

func testGUI(){
	db.StartGUI()
}

func testWrite() {
	ticker := time.NewTicker(time.Second*4)
	for {
		select {
		case <-ticker.C:
			write()
		}
	}
}

func write() {
	//populate model
	bm := booking.NewBookingModel()
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
	bm.AutoId = db.GenerateId(bm)

	err := db.Insert(bm)
	if err != nil {
		fmt.Println(err)
	}
}

func read() {
	b := &booking.BookingModel{}
	err := db.Get("Booking/201802/room_201802070030", b)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(b.NextOfKin)
	}
}

func delete() {
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
	rows, err := db.Fetch("Booking")
	if err != nil {
		fmt.Println(err.Error())
	} else {
		for _, r := range rows {
			b := &booking.BookingModel{}
			db.GetModel(r, b)

			fmt.Println(b.CustomerId)
		}
	}
}

func make(modelName string) db.Model {
	var m db.Model
	switch modelName {
	case "Booking":
		m = &booking.BookingModel{}
	}

	return m
}
