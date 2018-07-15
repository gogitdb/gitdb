package db

//import (
//	"time"
//	"testing"
//	"fmt"
//	"github.com/fobilow/gitdb/example/booking"
//	"os"
//)
//
//
//func TestDbLocking(t *testing.T) {
//	cfg := &Config{
//		DbPath:         "/tmp/data",
//		OnlineRemote:   "",
//		SshKey:         "/Users/okeugwu/.ssh/id_rsa",
//		Factory:        make,
//		SyncInterval:   time.Second * 61,
//		EncryptionKey:  "hellobjkdkdjkdjdkjkdjooo",
//	}
//
//	runLogFile, err := os.OpenFile("./db.log", os.O_WRONLY | os.O_CREATE | os.O_APPEND, 0666)
//	if err == nil {
//		cfg.Logger = log.New(runLogFile, "GITDB: ", log.Ldate|log.Ltime|log.Lshortfile)
//	}
//
//	User = NewUser("dev", "dev@gitdb.io")
//	Start(cfg)
//
//	ticker := time.NewTicker(time.Second*4)
//	for {
//		select {
//		case <-ticker.C:
//			//populate model
//			bm := booking.NewBookingModel()
//			bm.Type = booking.Room
//			bm.CheckInDate = time.Now()
//			bm.CheckOutDate = time.Now()
//			bm.CustomerId = "customer_1"
//			bm.Guests = 2
//			bm.CardsIssued = 1
//			bm.NextOfKin = "Kin 1"
//			bm.Status = booking.CheckedIn
//			bm.UserId = "user_1"
//			bm.RoomId = "room_1"
//			bm.AutoId = GenerateId(bm)
//
//			err := Insert(bm)
//			if err != nil {
//				fmt.Println(err)
//			}
//		}
//	}
//}
//
//func make(modelName string) Model {
//	var m Model
//	switch modelName {
//	case "Booking":
//		m = &booking.BookingModel{}
//	}
//
//	return m
//}
