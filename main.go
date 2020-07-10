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

func CreateUser(w http.ResponseWriter, r *http.Request) {

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	var u Users
	err := dec.Decode(&u)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = AddUser(u)
	if err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
}

func GetUser(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "" {
		value, _ := header.ParseValueAndParams(r.Header, "Content-Type")
		if value != "application/json" {
			msg := "Content-Type header is not application/json"
			http.Error(w, msg, http.StatusUnsupportedMediaType)
			return
		}
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1048576)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	var u Filter
	err := dec.Decode(&u)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		http.Error(w, err.Error(), http.StatusBadRequest)
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
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	log.Print(u.ID)
	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		msg := "Request body must only contain a single JSON object"
		http.Error(w, msg, http.StatusBadRequest)
		return
	}
	result, err := UserUpdate(u.ID, u)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprint(w, result)
}

func CheckUser(filter Filter) (lists []Users, info PaginationData) {
	collection, err := MongoInit()

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

func UserUpdate(id string, user UpdUsers) (string, error) {
	s, _ := primitive.ObjectIDFromHex(user.ID)
	collection, err := MongoInit()
	if err != nil {
		return "", err
	}
	updateResult, err := collection.UpdateOne(context.TODO(), bson.D{{"_id", s}}, bson.D{
		{"$set", bson.D{
			{"lastname", user.LastName}, {"email", user.Email}, {"city", user.City}, {"birthdate", user.BirthDate}, {"country", user.Country}, {"gender", user.Gender},
		}}})
	if err != nil {
		return "", err
	}
	if updateResult.MatchedCount == 0 {
		return "", errors.New("invalid id")
	}
	log.Printf("Matched %v documents and updated %v documents.\n", updateResult.MatchedCount, updateResult.ModifiedCount)
	return "user successfully updated", nil
}

func AddUser(user Users) error {
	var checkUserInBase []Users
	collection, err := MongoInit()
	if err != nil {
		return err
	}
	filter := Filter{"email", user.Email, "1"}
	checkUserInBase, _ = CheckUser(filter)
	if len(checkUserInBase) != 0 {
		errors.New("something wen't wromg")
	}
	if len(checkUserInBase) == 0 {
		user.ID = primitive.NewObjectID()
		_, err := collection.InsertOne(context.TODO(), user)
		if err != nil {
			return err
		}
		log.Print("successfully added user: ", user)
		return nil
	}
	return errors.New("something wen't wromg")
}

func MongoInit() (*mongo.Collection, error) {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://127.0.0.1:27017"))
	if err != nil {
		return nil, err
	}
	err = client.Connect(context.TODO())
	if err != nil {
		return nil, err
	}
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		return nil, err
	}
	log.Printf("Connected to MongoDB!")
	collection := client.Database("test").Collection("trainers")
	return collection, nil
}
