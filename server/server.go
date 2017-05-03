package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"sync"
)

type address string

type user struct {
	X       int64   `json:"x"`
	Y       int64   `json:"y"`
	Address address `json:"address"`
}

var users map[address]user

type point struct {
	ID      int64   `json:"id"`
	X       int64   `json:"x"`
	Y       int64   `json:"y"`
	Fee     float64 `json:"fee"`
	Address address `json:"address"`
}

var routes map[address][]point

var muUser sync.Mutex
var muRout sync.Mutex

// upload POST BODY json string
// {
// 	"x": 120444444,
// 	"y": 30111111,
// 	"address": "1604860a06bd66054cbbe4c547291cfa3897a8da"
// }
func uploadPosition(w http.ResponseWriter, req *http.Request) {
	log.Println("uploadPosition")

	body, _ := ioutil.ReadAll(req.Body)
	fmt.Println(string(body))
	var newUser user
	err := json.Unmarshal(body, &newUser)
	if err != nil {
		log.Println(err)
	}

	muUser.Lock()
	users[newUser.Address] = newUser
	muUser.Unlock()

	bytes, _ := json.Marshal(newUser)
	fmt.Fprint(w, string(bytes))
}

// download GET
// http://121.42.212.54:8080/download?address=1604860a06bd66054cbbe4c547291cfa3897a8da
func downloadPosition(w http.ResponseWriter, req *http.Request) {
	log.Println("downloadPosition")

	query := req.URL.Query()
	fmt.Println("GET: ", query["address"][0])

	var newUser user
	newUser.Address = address(query["address"][0])
	muUser.Lock()
	newUser = users[newUser.Address]
	muUser.Unlock()

	bytes, err := json.Marshal(newUser)
	if err != nil {
		log.Println(err)
	}
	fmt.Fprint(w, string(bytes))
}

// calculateFee POST BODY json string
// {
//  "id": 0
// 	"x": 120444444,
// 	"y": 30111111,
// 	"address": "1604860a06bd66054cbbe4c547291cfa3897a8da"
// }
// when id equals 0, it will reset the route. That is, 0 stands for start point
func calculateFee(w http.ResponseWriter, req *http.Request) {
	log.Println("calculateFee")

	body, _ := ioutil.ReadAll(req.Body)
	fmt.Println(string(body))
	var newPoint point
	err := json.Unmarshal(body, &newPoint)
	if err != nil {
		log.Println(err)
	}

	addr := newPoint.Address

	muRout.Lock()
	if newPoint.ID == 0 { //reset the route
		routes[addr] = nil
	} else {
		last := newPoint.ID - 1
		newPoint.Fee = routes[addr][last].Fee + math.Sqrt((math.Pow(float64(routes[addr][last].X-newPoint.X), 2)+math.Pow(float64(routes[addr][last].Y-newPoint.Y), 2)))/100
	}
	routes[addr] = append(routes[addr], newPoint)
	muRout.Unlock()

	var newUser user
	newUser.Address = newPoint.Address
	newUser.X = newPoint.X
	newUser.Y = newPoint.Y
	muUser.Lock()
	users[newUser.Address] = newUser
	muUser.Unlock()

	fmt.Println(routes[addr])

	bytes, _ := json.Marshal(newPoint)
	fmt.Fprint(w, string(bytes))
}

// getRoute GET
// http://121.42.212.54:8080/route?address=1604860a06bd66054cbbe4c547291cfa3897a8da
func getRoute(w http.ResponseWriter, req *http.Request) {
	log.Println("getRoute")

	query := req.URL.Query()
	fmt.Println("GET:", query["address"][0])

	var route []point
	addr := address(query["address"][0])
	muRout.Lock()
	route = routes[addr]
	muRout.Unlock()

	fmt.Println(routes[addr])

	bytes, err := json.Marshal(route)
	if err != nil {
		log.Println(err)
	}
	fmt.Fprint(w, string(bytes))
}

func main() {
	users = make(map[address]user)
	routes = make(map[address][]point)

	http.HandleFunc("/upload", uploadPosition)
	http.HandleFunc("/download", downloadPosition)
	http.HandleFunc("/calculate", calculateFee)
	http.HandleFunc("/route", getRoute)

	fmt.Println("Listening in port 8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

}
