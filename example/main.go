package main

import (
	"fmt"
	"time"

	"errors"
	"log"
	"os"

	db "github.com/fobilow/gitdb"
	"github.com/fobilow/gitdb/example/booking"
)

var dbconn *db.Connection
var logToFile bool

func init() {
	cfg := &db.Config{
		DbPath:        "./data",
		OnlineRemote:  os.Getenv("GITDB_REPO"),
		SyncInterval:  time.Second * 5,
		EncryptionKey: "XVlBzgbaiCMRAjWwhTHctcuAxhxKQFDa", //this has to be 32 bytes to select AES-256
		User:          db.NewUser("dev", "dev@gitdb.io"),
		GitDriver:     db.GitDriverBinary,
		//gitDriver: db.GitDriverGoGit,
	}

	if logToFile {
		runLogFile, err := os.OpenFile("./db.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
		if err == nil {
			logger := log.New(runLogFile, "GITDB: ", log.Ldate|log.Ltime|log.Lshortfile)
			db.SetLogger(logger)
		}
	}

	var err error
	dbconn, err = db.Open(cfg)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func main() {
	defer dbconn.Close()
	write()
	fetch()
}

func testTransaction() {
	t := dbconn.NewTransaction("booking")
	t.AddOperation(updateRoom)
	t.AddOperation(lockRoom)
	t.AddOperation(saveBooking)
	t.Commit()
}

func updateRoom() error  { println("updating room..."); return nil }
func lockRoom() error    { println("locking room"); return errors.New("cannot lock room") }
func saveBooking() error { println("saving booking"); return nil }

func testWrite() {
	ticker := time.NewTicker(time.Second * 4)
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
	bm.AutoId = dbconn.GenerateId(bm)

	err := dbconn.Insert(bm)
	if err != nil {
		fmt.Println(err)
	}
}

func read() {
	b := &booking.BookingModel{}
	err := dbconn.Get("Booking/201802/room_201802070030", b)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(b.NextOfKin)
	}
}

func delete() {
	err := dbconn.Delete("Booking/201801/room_201801111823")
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println("Deleted")
	}
}

func search() {
	searchParam := &db.SearchParam{Index: "CustomerId", Value: "customer_2"}
	rows, err := dbconn.Search("Booking", []*db.SearchParam{searchParam}, db.SEARCH_MODE_EQUALS)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		for _, r := range rows {
			fmt.Println(r)
		}
	}
}

func fetch() {
	rows, err := dbconn.Fetch("Booking")
	if err != nil {
		fmt.Println(err.Error())
	} else {
		bookings := []*booking.BookingModel{}
		for _, r := range rows {
			b := &booking.BookingModel{}
			r.Hydrate(b)
			bookings = append(bookings, b)

			fmt.Println(fmt.Sprintf("%s-%s", b.ID, b.CreatedAt))
		}
	}
}

func mail() {
	mails := dbconn.GetMails()
	for _, m := range mails {
		fmt.Println(m.Body)
	}
}
