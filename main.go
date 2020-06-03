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

type Objects struct {
}

type Users struct {
	ID        primitive.ObjectID `bson:"_id" json:"id,omitempty"`
	Email     string             ` json:"email"`
	LastName  string             `json:"last_name"`
	Country   string             `json:"country"`
	City      string             `json:"city"`
	Gender    string             `json:"gender"`
	BirthDate string             `json:"birth_date"`
}

type Pogination struct {
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

type Filter struct {
	Field    string `json:"field"`
	NewValue string `json:"neValue`
	Page     string `json:"page"`
}

func main() {

	router := mux.NewRouter().StrictSlash(true)

	router.HandleFunc("/user/add", CreateUser).Methods("POST")
	router.HandleFunc("/user/get_list", GetUser).Methods("POST")
	router.HandleFunc("/user/update_user", UpdateUser).Methods("POST")
	http.ListenAndServe(":6666", router)

}

func CreateUser(w http.ResponseWriter, r *http.Request) {
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

	var u Users
	err := dec.Decode(&u)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError

		switch {
		// Catch any syntax errors in the JSON and send an error message
		// which interpolates the location of the problem to make it
		// easier for the client to fix.
		case errors.As(err, &syntaxError):
			msg := fmt.Sprintf("Request body contains badly-formed JSON (at position %d)", syntaxError.Offset)
			http.Error(w, msg, http.StatusBadRequest)

		// In some circumstances Decode() may also return an
		// io.ErrUnexpectedEOF error for syntax errors in the JSON.
		case errors.Is(err, io.ErrUnexpectedEOF):
			msg := fmt.Sprintf("Request body contains badly-formed JSON")
			http.Error(w, msg, http.StatusBadRequest)

		// Catch any type errors, like trying to assign a string in the
		// JSON request body to a int field in our User struct.
		case errors.As(err, &unmarshalTypeError):
			msg := fmt.Sprintf("Request body contains an invalid value for the %q field (at position %d)", unmarshalTypeError.Field, unmarshalTypeError.Offset)
			http.Error(w, msg, http.StatusBadRequest)

		// Catch the error caused by extra unexpected fields in the request
		// body. We extract the field name from the error message and
		// interpolate it in our custom error message.
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			msg := fmt.Sprintf("Request body contains unknown field %s", fieldName)
			http.Error(w, msg, http.StatusBadRequest)

		// An io.EOF error is returned by Decode() if the request body is
		// empty.
		case errors.Is(err, io.EOF):
			msg := "Request body must not be empty"
			http.Error(w, msg, http.StatusBadRequest)

		// Catch the error caused by the request body being too large.
		case err.Error() == "http: request body too large":
			msg := "Request body must not be larger than 1MB"
			http.Error(w, msg, http.StatusRequestEntityTooLarge)

		// Otherwise default to logging the error and sending a 500 Internal
		// Server Error response.
		default:
			log.Println(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
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
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError

		switch {
		// Catch any syntax errors in the JSON and send an error message
		// which interpolates the location of the problem to make it
		// easier for the client to fix.
		case errors.As(err, &syntaxError):
			msg := fmt.Sprintf("Request body contains badly-formed JSON (at position %d)", syntaxError.Offset)
			http.Error(w, msg, http.StatusBadRequest)

		// In some circumstances Decode() may also return an
		// io.ErrUnexpectedEOF error for syntax errors in the JSON.
		case errors.Is(err, io.ErrUnexpectedEOF):
			msg := fmt.Sprintf("Request body contains badly-formed JSON")
			http.Error(w, msg, http.StatusBadRequest)

		// Catch any type errors, like trying to assign a string in the
		// JSON request body to a int field in our User struct.
		case errors.As(err, &unmarshalTypeError):
			msg := fmt.Sprintf("Request body contains an invalid value for the %q field (at position %d)", unmarshalTypeError.Field, unmarshalTypeError.Offset)
			http.Error(w, msg, http.StatusBadRequest)

		// Catch the error caused by extra unexpected fields in the request
		// body. We extract the field name from the error message and
		// interpolate it in our custom error message.
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			msg := fmt.Sprintf("Request body contains unknown field %s", fieldName)
			http.Error(w, msg, http.StatusBadRequest)

		// An io.EOF error is returned by Decode() if the request body is
		// empty.
		case errors.Is(err, io.EOF):
			msg := "Request body must not be empty"
			http.Error(w, msg, http.StatusBadRequest)

		// Catch the error caused by the request body being too large.
		case err.Error() == "http: request body too large":
			msg := "Request body must not be larger than 1MB"
			http.Error(w, msg, http.StatusRequestEntityTooLarge)

		// Otherwise default to logging the error and sending a 500 Internal
		// Server Error response.
		default:
			log.Println(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
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
	var u UpdUsers
	err := dec.Decode(&u)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError

		switch {
		// Catch any syntax errors in the JSON and send an error message
		// which interpolates the location of the problem to make it
		// easier for the client to fix.
		case errors.As(err, &syntaxError):
			msg := fmt.Sprintf("Request body contains badly-formed JSON (at position %d)", syntaxError.Offset)
			http.Error(w, msg, http.StatusBadRequest)

		// In some circumstances Decode() may also return an
		// io.ErrUnexpectedEOF error for syntax errors in the JSON.
		case errors.Is(err, io.ErrUnexpectedEOF):
			msg := fmt.Sprintf("Request body contains badly-formed JSON")
			http.Error(w, msg, http.StatusBadRequest)

		// Catch any type errors, like trying to assign a string in the
		// JSON request body to a int field in our User struct.
		case errors.As(err, &unmarshalTypeError):
			msg := fmt.Sprintf("Request body contains an invalid value for the %q field (at position %d)", unmarshalTypeError.Field, unmarshalTypeError.Offset)
			http.Error(w, msg, http.StatusBadRequest)

		// Catch the error caused by extra unexpected fields in the request
		// body. We extract the field name from the error message and
		// interpolate it in our custom error message.
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			msg := fmt.Sprintf("Request body contains unknown field %s", fieldName)
			http.Error(w, msg, http.StatusBadRequest)

		// An io.EOF error is returned by Decode() if the request body is
		// empty.
		case errors.Is(err, io.EOF):
			msg := "Request body must not be empty"
			http.Error(w, msg, http.StatusBadRequest)

		// Catch the error caused by the request body being too large.
		case err.Error() == "http: request body too large":
			msg := "Request body must not be larger than 1MB"
			http.Error(w, msg, http.StatusRequestEntityTooLarge)

		// Otherwise default to logging the error and sending a 500 Internal
		// Server Error response.
		default:
			log.Println(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
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
	if len(u.ID) == 0 {
		msg := "Please send user id"
		http.Error(w, msg, http.StatusUnprocessableEntity)
	}
	log.Print(u.ID)
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

	r := bson.D{{filter.Field, filter.NewValue}}
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

	filter.Field = "_id"
	updateResult, err := collection.UpdateOne(context.TODO(), bson.D{{"_id", s}}, bson.D{
		{"$set", bson.D{
			{"lastname", user.LastName}, {"email", user.Email}, {"city", user.City}, {"birthdate", user.BirthDate}, {"country", user.Country}, {"gender", user.Gender},
		}}})
	if err != nil {
		log.Fatal(err, s)
	}
	log.Printf("Matched %v documents and updated %v documents.\n", updateResult.MatchedCount, updateResult.ModifiedCount)
	return "successfully updated"
}

func AddUser(user Users) string {
	var checkUserInBase []Users
	collection := MongoInit()
	filter := Filter{"email", user.Email, "1"}
	checkUserInBase, _ = CheckUser(filter)
	log.Print(checkUserInBase)
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
