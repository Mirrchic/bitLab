// package main

// import (
// 	"context"
// 	"encoding/json"
// 	"fmt"
// 	"io/ioutil"
// 	"log"
// 	"os"

// 	"go.mongodb.org/mongo-driver/mongo"
// 	"go.mongodb.org/mongo-driver/mongo/options"
// )

// type Users struct {
// 	Objects []struct {
// 		Email     string `json:"email"`
// 		LastName  string `json:"last_name"`
// 		Country   string `json:"country"`
// 		City      string `json:"city"`
// 		Gender    string `json:"gender"`
// 		BirthDate string `json:"birth_date"`
// 	} `json:"objects"`
// }

// type Trainer struct {
// 	Name string
// 	Age  int
// 	City string
// }

// func main() {
// 	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://127.0.0.1:27017"))
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	err = client.Connect(context.TODO())
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	err = client.Ping(context.TODO(), nil)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	fmt.Println("Connected to MongoDB!")
// 	collection := client.Database("test").Collection("trainers")

// 	var user Users
// 	jsonFile, err := os.Open("users_go.json")
// 	if err != nil {
// 		log.Fatalln(err)
// 	}
// 	byteJson, _ := ioutil.ReadAll(jsonFile)
// 	defer jsonFile.Close()
// 	json.Unmarshal(byteJson, &user)
// 	for i := range user.Objects {
// 		collection.InsertOne(context.TODO(), user.Objects[i])
// 	}
// 	log.Print("Collection is creatade")

// }

"objects":[ 
{ 
"email":"Logan_Devonport3313@tonsy.org",
"last_name":"Devonport",
"country":"Oman",
"city":"Madrid",
"gender":"Male",
"birth_date":"Friday, April 4, 8527 8:45 AM"
},]}
curl -d@"objects":[ :[  http://127.0.0.1:8080/user/get_list
