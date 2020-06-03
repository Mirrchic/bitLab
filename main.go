package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"

	. "github.com/gobeam/mongo-go-pagination"
	"github.com/golang/gddo/httputil/header"
	"github.com/gorilla/mux"

	"go.mongodb.org/mongo-driver/bson"

	"go.mongodb.org/mongo-driver/mongo"
)

type Users struct {
	ID        primitive.ObjectID `bson:"_id" json:"id,omitempty"`
	Email     string             ` json:"email"`
	LastName  string             `json:"last_name"`
	Country   string             `json:"country"`
	City      string             `json:"city"`
	Gender    string             `json:"gender"`
	BirthDate string             `json:"birth_date"`
}

//type getting page request<
type Filter struct {
	Field        string `json:"field"`
	QueringValue string `json:"neValue`
	Page         string `json:"page"`
}

type PageResponse struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Email     string             ` json:"email"`
	LastName  string             `json:"last_name"`
	Country   string             `json:"country"`
	City      string             `json:"city"`
	Gender    string             `json:"gender"`
	BirthDate string             `json:"birth_date"`
	Page      string             `json:"Page"`
	Limit     string             `json:"Limit"`
}

func main() {

	router := mux.NewRouter().StrictSlash(true)
	//forr adding user send a jason in the request with the parameters of the new user, example:
	// 	{
	//                 "email": "Didsdsddsachsasda@gmail.com",
	//                 "last_name": "Mor",
	//                 "country": "Vietna,",
	//                 "city": "Borispol",
	//                 "gender": "Male",
	//                 "birth_date": "Friday, April 4, 8527 8:45 AM"
	// }
	router.HandleFunc("/user/add", CreateUser).Methods("POST")
	// to get data about users, send a request with the field and data and page,
	// in the answer you will receive user data
	//example of request:
	// 	{
	//         "field": "lastname",
	//         "newValue": "Joe"
	// }
	// just add number of page to keep moving through pages
	router.HandleFunc("/user/get_list", GetUser).Methods("POST")
	//send data user data in request with his ID to update his info
	//example:
	//{
	// 		"id": "5ed6d4c1ded14738cece7e9e",
	// 		"email": "Didsdsddsachsasda@gmail.com",
	// 		"last_name": "Mor",
	// 		"country": "Vietna,",
	// 		"city": "Borispol",
	// 		"gender": "Male",
	// 		"birth_date": "Friday, April 4, 8527 8:45 AM"
	// }
	router.HandleFunc("/user/update_user", UpdateUser).Methods("POST")
	http.ListenAndServe(":6666", router)

}

func JsonCheck(err error) string {

	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError

		switch {
		// Catch any syntax errors in the JSON and send an error message
		// which interpolates the location of the problem to make it
		// easier for the client to fix.
		case errors.As(err, &syntaxError):
			return "Request body contains badly-formed JSON "

		// In some circumstances Decode() may also return an
		// io.ErrUnexpectedEOF error for syntax errors in the JSON.
		case errors.Is(err, io.ErrUnexpectedEOF):
			return "Request body contains badly-formed JSON"

		// Catch any type errors, like trying to assign a string in the
		// JSON request body to a int field in our User struct.
		case errors.As(err, &unmarshalTypeError):
			return "Request body contains an invalid value for the field "

		// Catch the error caused by extra unexpected fields in the request
		// body. We extract the field name from the error message and
		// interpolate it in our custom error message.
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			return "Request body contains unknown field "

		// An io.EOF error is returned by Decode() if the request body is
		// empty.
		case errors.Is(err, io.EOF):
			return "Request body must not be empty"

		// Catch the error caused by the request body being too large.
		case err.Error() == "http: request body too large":
			return "Request body must not be larger than 1MB"

		// Otherwise default to logging the error and sending a 500 Internal
		// Server Error response.
		default:
			return "enternal server error"
		}
	}

	// Call decode again, using a pointer to an empty anonymous struct as
	// the destination. If the request body only contained a single JSON
	// object this will return an io.EOF error. So if we get anything else,
	// we know that there is additional data in the request body.
	return ""
}

func CreateUser(w http.ResponseWriter, r *http.Request) {

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	var u Users
	err := dec.Decode(&u)
	if err != nil {
		msg := JsonCheck(err)
		http.Error(w, msg, http.StatusBadRequest)
	}

	// Call decode again, using a pointer to an empty anonymous struct as
	// the destination. If the request body only contained a single JSON
	// object this will return an io.EOF error. So if we get anything else,
	// we know that there is additional data in the request body.
	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		msg := "Request body must only contain a single JSON object"
		http.Error(w, msg, http.StatusBadRequest)
		return
	}
	emailCheck := AddUser(u)
	if emailCheck != "user successfully added" {
		msg := emailCheck
		http.Error(w, msg, http.StatusConflict)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
}

func GetUser(w http.ResponseWriter, r *http.Request) {
	// If the Content-Type header is present, check that it has the value
	// application/json.
	if r.Header.Get("Content-Type") != "" {

		value, _ := header.ParseValueAndParams(r.Header, "Content-Type")
		if value != "application/json" {
			msg := "Content-Type header is not application/json"
			http.Error(w, msg, http.StatusUnsupportedMediaType)
			return
		}
	}

	// Use http.MaxBytesReader to enforce a maximum read of 1MB
	r.Body = http.MaxBytesReader(w, r.Body, 1048576)

	// Setup the decoder and call the DisallowUnknownFields() method on it.
	// This will cause Decode() to return a "json: unknown field ..." error
	// if it encounters any extra unexpected fields in the JSON. Strictly
	// speaking, it returns an error for "keys which do not match any
	// non-ignored, exported fields in the destination".
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	var u Filter
	err := dec.Decode(&u)
	if err != nil {
		msg := JsonCheck(err)
		http.Error(w, msg, http.StatusBadRequest)
	}

	// Call decode again, using a pointer to an empty anonymous struct as
	// the destination. If the request body only contained a single JSON
	// object this will return an io.EOF error. So if we get anything else,
	// we know that there is additional data in the request body.
	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		msg := "Request body must only contain a single JSON object"
		http.Error(w, msg, http.StatusBadRequest)
		return
	}
	usersFound, paginInfo := CheckUser(u)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(usersFound)
	json.NewEncoder(w).Encode(paginInfo)
}

type UpdUsers struct {
	ID        string ` json:"id,omitempty"`
	Email     string `bson:"" json:"email"`
	LastName  string `json:"last_name"`
	Country   string `json:"country"`
	City      string `json:"city"`
	Gender    string `json:"gender"`
	BirthDate string `json:"birth_date"`
}

func UpdateUser(w http.ResponseWriter, r *http.Request) {
	var u UpdUsers
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	err := dec.Decode(&u)
	if err != nil {
		msg := JsonCheck(err)
		http.Error(w, msg, http.StatusBadRequest)
	}
	log.Print(u.ID)
	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		msg := "Request body must only contain a single JSON object"
		http.Error(w, msg, http.StatusBadRequest)
		return
	}
	result := UserUpdate(u.ID, u)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprint(w, result)
}

func CheckUser(filter Filter) (lists []Users, info PaginationData) {
	collection := MongoInit()
	limitPerUser := bson.D{
		{"email", 1},
		{"lastname", 1},
		{"country", 1},
		{"gender", 1},
		{"birthdate", 1},
		{"city", 1},
	}
	var limit int64
	limit = 10

	r := bson.D{{filter.Field, filter.QueringValue}}
	page, _ := strconv.ParseInt(filter.Page, 10, 64)
	paginatedData, err := New(collection).Limit(limit).Page(page).Select(limitPerUser).Filter(r).Find()
	if err != nil {
		panic(err)
	}
	for _, raw := range paginatedData.Data {
		var product *Users
		if marshallErr := bson.Unmarshal(raw, &product); marshallErr == nil {
			lists = append(lists, *product)
		}
	}
	info = paginatedData.Pagination
	return lists, info
}

//unfortunately forced to create another type because:
//cannot transform type string to a BSON Document: WriteString can only write while positioned on a Element or Value but is positioned on a TopLevel

func UserUpdate(id string, user UpdUsers) string {
	s, _ := primitive.ObjectIDFromHex(user.ID)
	collection := MongoInit()
	updateResult, err := collection.UpdateOne(context.TODO(), bson.D{{"_id", s}}, bson.D{
		{"$set", bson.D{
			{"lastname", user.LastName}, {"email", user.Email}, {"city", user.City}, {"birthdate", user.BirthDate}, {"country", user.Country}, {"gender", user.Gender},
		}}})
	if err != nil {
		log.Fatal(err, s)
	}
	if updateResult.MatchedCount == 0 {
		return "invalid id"
	}
	log.Printf("Matched %v documents and updated %v documents.\n", updateResult.MatchedCount, updateResult.ModifiedCount)
	return "user successfully updated"
}

func AddUser(user Users) string {
	var checkUserInBase []Users
	collection := MongoInit()
	filter := Filter{"email", user.Email, "1"}
	checkUserInBase, _ = CheckUser(filter)
	if len(checkUserInBase) != 0 {
		return "this email already used"
	}
	if len(checkUserInBase) == 0 {
		user.ID = primitive.NewObjectID()
		_, err := collection.InsertOne(context.TODO(), user)
		if err != nil {
			log.Fatal(err)
		}
		log.Print("successfully added user: ", user)
		return "user successfully added"
	}
	return "something went wrong"
}

func MongoInit() *mongo.Collection {

	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://127.0.0.1:27017"))
	if err != nil {
		log.Fatal(err)
	}
	err = client.Connect(context.TODO())
	if err != nil {
		log.Fatal(err)
	}
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Connected to MongoDB!")
	collection := client.Database("test").Collection("trainers")
	return collection
}
