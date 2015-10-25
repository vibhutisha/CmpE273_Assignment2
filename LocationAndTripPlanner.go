package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type LocationAndTripResults struct {
	Results []struct {
		AddressComponents []struct {
			LongName  string   `json:"long_name"`
			ShortName string   `json:"short_name"`
			Types     []string `json:"types"`
		} `json:"address_components"`
		FormattedAddress string `json:"formatted_address"`
		Geometry         struct {
			Location struct {
				Latitude  float64 `json:"lat"`
				Longitude float64 `json:"lng"`
			} `json:"location"`
			LocationType string `json:"location_type"`
			Viewport     struct {
				Northeast struct {
					Latitude  float64 `json:"lat"`
					Longitude float64 `json:"lng"`
				} `json:"northeast"`
				Southwest struct {
					Latitude  float64 `json:"lat"`
					Longitude float64 `json:"lng"`
				} `json:"southwest"`
			} `json:"viewport"`
		} `json:"geometry"`
		PlaceID string   `json:"place_id"`
		Types   []string `json:"types"`
	} `json:"results"`
	Status string `json:"status"`
}

type Coordinate struct {
	Latitude  float64 `json:"lat" bson:"lat"`
	Longitude float64 `json:"lng" bson:"lng"`
}

type LocationAndTripResponse struct {
	Id      bson.ObjectId `json:"id" bson:"_id"`
	Name    string        `json:"name" bson:"name"`
	Address string        `json:"address" bson:"address"`
	City    string        `json:"city" bson:"city"`
	State   string        `json:"state" bson:"state"`
	Zip     string        `json:"zip" bson:"zip"`
	LatLong Coordinate    `json:"coordinate" bson:"coordinate"`
}

type LatLongConnect struct {
	session *mgo.Session
}

func NewLatLongConnect(s *mgo.Session) *LatLongConnect {
	return &LatLongConnect{s}
}

func (latlongc LatLongConnect) LatLongConnectGet(w http.ResponseWriter, r *http.Request, i httprouter.Params) {

	id := i.ByName("id")

	if !bson.IsObjectIdHex(id) {
		w.WriteHeader(404)
		fmt.Println("Error in Query Input... Please Enter again in Correct Format!!")
		return
	}

	oid := bson.ObjectIdHex(id)

	lntr := LocationAndTripResponse{}

	if err := latlongc.session.DB("cmpe273").C("LocationAndTripPlanner").FindId(oid).One(&lntr); err != nil {
		w.WriteHeader(404)
		fmt.Println("Error in Query Input")
		return
	}

	json.NewDecoder(r.Body).Decode(lntr)

	uj, _ := json.Marshal(lntr)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	fmt.Fprintf(w, "%s", uj)
}

func (latlongc LatLongConnect) UpdateLatLongConnect(w http.ResponseWriter, r *http.Request, i httprouter.Params) {

	id := i.ByName("id")

	if !bson.IsObjectIdHex(id) {
		w.WriteHeader(404)
		fmt.Println("Error!!!")
		return
	}

	oid := bson.ObjectIdHex(id)

	lntr := LocationAndTripResponse{}
	lnts := LocationAndTripResponse{}

	lnts.Id = oid

	json.NewDecoder(r.Body).Decode(&lnts)

	if err := latlongc.session.DB("cmpe273").C("LocationAndTripPlanner").FindId(oid).One(&lntr); err != nil {
		w.WriteHeader(404)
		return
	}

	nameValue := lntr.Name

	connections := latlongc.session.DB("cmpe273").C("LocationAndTripPlanner")

	lntr = getvaluefrompost(&lnts)
	connections.Update(bson.M{"_id": oid}, bson.M{"$set": bson.M{"address": lnts.Address, "city": lnts.City, "state": lnts.State, "zip": lnts.Zip, "coordinate": bson.M{"lat": lnts.LatLong.Latitude, "lng": lntr.LatLong.Longitude}}})

	lntr.Name = nameValue

	uj, _ := json.Marshal(lntr)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	fmt.Fprintf(w, "%s", uj)

}

func (latlongc LatLongConnect) RemoveLatLongConnect(w http.ResponseWriter, r *http.Request, i httprouter.Params) {

	id := i.ByName("id")

	if !bson.IsObjectIdHex(id) {
		w.WriteHeader(404)
		return
	}

	oid := bson.ObjectIdHex(id)

	if err := latlongc.session.DB("cmpe273").C("LocationAndTripPlanner").RemoveId(oid); err != nil {
		w.WriteHeader(404)
		return
	}

	w.WriteHeader(200)
}

func (latlongc LatLongConnect) CreateLatLongConnect(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

	postlntr := LocationAndTripResponse{}

	json.NewDecoder(r.Body).Decode(&postlntr)

	getValue := getvaluefrompost(&postlntr)

	getValue.Id = bson.NewObjectId()

	latlongc.session.DB("cmpe273").C("LocationAndTripPlanner").Insert(getValue)

	uj, _ := json.Marshal(getValue)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	fmt.Fprintf(w, "%s", uj)
}

func getvaluefrompost(rep *LocationAndTripResponse) LocationAndTripResponse {

	addrs := rep.Address
	cityy := rep.City

	stat := strings.Replace(rep.State, " ", "+", -1)
	repladd := strings.Replace(addrs, " ", "+", -1)
	replcity := strings.Replace(cityy, " ", "+", -1)

	uri := "http://maps.google.com/maps/api/geocode/json?address=" + repladd + "+" + replcity + "+" + stat + "&sensor=false"

	resp, _ := http.Get(uri)

	body, _ := ioutil.ReadAll(resp.Body)

	C := LocationAndTripResults{}

	err := json.Unmarshal(body, &C)
	if err != nil {
		panic(err)
	}

	for _, Sample := range C.Results {
		rep.LatLong.Latitude = Sample.Geometry.Location.Latitude
		rep.LatLong.Longitude = Sample.Geometry.Location.Longitude
	}

	return *rep
}

func getSession() *mgo.Session {

	s, err := mgo.Dial("mongodb://vibhuti:vibhuti@ds045464.mongolab.com:45464/cmpe273")

	if err != nil {
		panic(err)
	}
	return s
}

func main() {

	r := httprouter.New()
	latlongc := NewLatLongConnect(getSession())

	r.GET("/locations/:id", latlongc.LatLongConnectGet)
	r.POST("/locations", latlongc.CreateLatLongConnect)
	r.PUT("/locations/:id", latlongc.UpdateLatLongConnect)
	r.DELETE("/locations/:id", latlongc.RemoveLatLongConnect)

	http.ListenAndServe("localhost:3022", r)

}

//Commands
//curl -XPOST -H 'Content-Type: application/json' -d '{"name": "Vibhuti Sharma", "address": "123 Main Street", "city" : "San Jose", "state" : "CA", "zip": "12345"}' http://localhost:3022/locations
//curl http://localhost:3022/locations/562c6545d5cc55178495951
//curl -XPUT -X 'Content-Type: application/json' -i '{"name": "VSid Sharma", "address": "123 Main Street", "city" : "San Jose", "state" : "CA", "zip": "12345"}' http://localhost:3022/locations/562c6545d5cc55178495951
//curl -XDELETE http://localhost:3022/locations/562c6545d5cc55178495951
//
