package db

import (
	"net/http"
	"fmt"
	"net"
)

func startServer(){
	appHost := "localhost"
	port := 4120

	http.HandleFunc("/app/version", handler)

	http.Handle("/", http.FileServer(http.Dir("./gui")))

	addr := fmt.Sprintf("%s:%d", appHost, port)
	fmt.Println("GitDB GUI will run at http://" + addr)

	l, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Println(err.Error())
	}

	err = http.Serve(l, nil)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Server started!")
	}
}

func handler(w http.ResponseWriter, r *http.Request) {

}

