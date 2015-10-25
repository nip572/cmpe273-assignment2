package main

import (
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"strings"
	"time"
)

type RequestLocation struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	City    string `json:"city"`
	State   string `json:"state"`
	Zip     string `json:"zip"`
}

type ResponseLocation struct {
	ID         bson.ObjectId `json:"id" bson:"_id,omitempty"`
	Name       string        `json:"name"`
	Address    string        `json:"address"`
	City       string        `json:"city"`
	State      string        `json:"state"`
	Zip        string        `json:"zip"`
	Coordinate struct {
		Lat float64 `json:"lat"`
		Lng float64 `json:"lng"`
	} `json:"coordinate"`
}

type GeoLocation struct {
	Results []struct {
		AddressComponents []struct {
			LongName  string   `json:"long_name"`
			ShortName string   `json:"short_name"`
			Types     []string `json:"types"`
		} `json:"address_components"`
		FormattedAddress string `json:"formatted_address"`
		Geometry         struct {
			Location struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			} `json:"location"`
			LocationType string `json:"location_type"`
			Viewport     struct {
				Northeast struct {
					Lat float64 `json:"lat"`
					Lng float64 `json:"lng"`
				} `json:"northeast"`
				Southwest struct {
					Lat float64 `json:"lat"`
					Lng float64 `json:"lng"`
				} `json:"southwest"`
			} `json:"viewport"`
		} `json:"geometry"`
		PlaceID string   `json:"place_id"`
		Types   []string `json:"types"`
	} `json:"results"`
	Status string `json:"status"`
}

type LocationResponseByGoogle struct {
	Results []struct {
		AddressComponents []struct {
			LongName  string   `json:"long_name"`
			ShortName string   `json:"short_name"`
			Types     []string `json:"types"`
		} `json:"address_components"`
		FormattedAddress string `json:"formatted_address"`
		Geometry         struct {
			Location struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			} `json:"location"`
			LocationType string `json:"location_type"`
			Viewport     struct {
				Northeast struct {
					Lat float64 `json:"lat"`
					Lng float64 `json:"lng"`
				} `json:"northeast"`
				Southwest struct {
					Lat float64 `json:"lat"`
					Lng float64 `json:"lng"`
				} `json:"southwest"`
			} `json:"viewport"`
		} `json:"geometry"`
		PlaceID string   `json:"place_id"`
		Types   []string `json:"types"`
	} `json:"results"`
	Status string `json:"status"`
}

var mgoCollection *mgo.Collection
var Location_Response ResponseLocation

const (
	timeout = time.Duration(time.Second * 100)
)

//CONNECT TO MONGO INSTANCE

func ConnectToMongo() {
	
	uri := "mongodb://nipun:nipun@ds045464.mongolab.com:45464/db2"
	ses, err := mgo.Dial(uri)

	if err != nil {
		fmt.Printf("Problem Connecting to Mongo Instance  %v\n", err)
	} else {
		ses.SetSafe(&mgo.Safe{})
	
		mgoCollection = ses.DB("db2").C("qwerty")
	}
}
// GET LOCATION FROM GOOGLE
func getLocationFromGoogle(address string) (geoLoc GeoLocation) {

	client := http.Client{Timeout: timeout}

	url := fmt.Sprintf("http://maps.google.com/maps/api/geocode/json?address=%s", address)

	res, err := client.Get(url)

	if err != nil {
		fmt.Errorf("Wrong Address %v", err)
	}
	defer res.Body.Close()

	decoder := json.NewDecoder(res.Body)

	err = decoder.Decode(&geoLoc)
	if err != nil {
		fmt.Errorf("Error in getting location %v", err)
	}

	return geoLoc
}

//POST LOCATION 
func PostLocation(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {

	var temporaryLocationRequest RequestLocation
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&temporaryLocationRequest)
	if err != nil {
		fmt.Errorf("Wrong Input provided, please enter it correctly %v", err)
	}
	address := temporaryLocationRequest.Address + " " + temporaryLocationRequest.City + " " + temporaryLocationRequest.State + " " + temporaryLocationRequest.Zip

	address = strings.Replace(address, " ", "%20", -1)

	locationDetails := getLocationFromGoogle(address)

	Location_Response.ID = bson.NewObjectId()

	Location_Response.Address = temporaryLocationRequest.Address

	Location_Response.City = temporaryLocationRequest.City

	Location_Response.Name = temporaryLocationRequest.Name

	Location_Response.State = temporaryLocationRequest.State

	Location_Response.Zip = temporaryLocationRequest.Zip

	Location_Response.Coordinate.Lat = locationDetails.Results[0].Geometry.Location.Lat

	Location_Response.Coordinate.Lng = locationDetails.Results[0].Geometry.Location.Lng

	err = mgoCollection.Insert(Location_Response)
	if err != nil {
		fmt.Printf("Problem inserting document: %v\n", err)
	}

	err = mgoCollection.FindId(Location_Response.ID).One(&Location_Response)
	if err != nil {
		fmt.Printf("Could not find doc %v\n")
	}
	rw.Header().Set("Content-Type", "application/json; charset=UTF-8")
	rw.WriteHeader(201)
	json.NewEncoder(rw).Encode(Location_Response)
}


//GET LOCATION FUNCTION
func getLocation(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {

	id := bson.ObjectIdHex(p.ByName("locationID"))

	err := mgoCollection.FindId(id).One(&Location_Response)

	if err != nil {
		fmt.Printf("Could not find doc%v\n")
	}
	rw.Header().Set("Content-Type", "application/json; charset=UTF-8")

	rw.WriteHeader(200)

	json.NewEncoder(rw).Encode(Location_Response)
}




// UPDATE OPERATION - UPDATE LOCATION

func updateLocation(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {

	var GetTempResponse ResponseLocation

	var Location_Response ResponseLocation

	id := bson.ObjectIdHex(p.ByName("locationID"))

	err := mgoCollection.FindId(id).One(&Location_Response)
	if err != nil {
		fmt.Printf("Could not find doc %v\n")
	}
	GetTempResponse.Name = Location_Response.Name

	GetTempResponse.Address = Location_Response.Address

	GetTempResponse.City = Location_Response.City

	GetTempResponse.State = Location_Response.State

	GetTempResponse.Zip = Location_Response.Zip

	decoder := json.NewDecoder(req.Body)

	err = decoder.Decode(&GetTempResponse)

	if err != nil {
		fmt.Errorf("Error, Input could not be decoded properly %v", err)
	}

	address := GetTempResponse.Address + " " + GetTempResponse.City + " " + GetTempResponse.State + " " + GetTempResponse.Zip

	address = strings.Replace(address, " ", "%20", -1)

	locationDetails := getLocationFromGoogle(address)

	GetTempResponse.Coordinate.Lat = locationDetails.Results[0].Geometry.Location.Lat

	GetTempResponse.Coordinate.Lng = locationDetails.Results[0].Geometry.Location.Lng

	err = mgoCollection.UpdateId(id, GetTempResponse)
	if err != nil {
		fmt.Printf("Error Please check document %v\n")
	}

	err = mgoCollection.FindId(id).One(&Location_Response)
	if err != nil {
		fmt.Printf("got an error finding a doc %v\n")
	}
	rw.Header().Set("Content-Type", "application/json; charset=UTF-8")

	rw.WriteHeader(201)

	json.NewEncoder(rw).Encode(Location_Response)
}


// DELETE LOCATION FROM MONGODB
func deleteLocation(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {

	idLoc := bson.ObjectIdHex(p.ByName("locationID"))

	err := mgoCollection.RemoveId(idLoc)
	if err != nil {
		fmt.Printf("got an error deleting a doc %v\n")
	}
	rw.WriteHeader(200)
}


// MAIN FUNCTION START
func main() {
	mux := httprouter.New()

	mux.GET("/locations/:locationID", getLocation)

	mux.POST("/locations", PostLocation)

	mux.PUT("/locations/:locationID", updateLocation)

	mux.DELETE("/locations/:locationID", deleteLocation)

	server := http.Server{
		Addr:    "0.0.0.0:8080",
		Handler: mux,
	}
	ConnectToMongo()
	server.ListenAndServe()
}
