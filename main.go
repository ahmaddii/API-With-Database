package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

var db *sql.DB

type User struct {
	Id    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email`
}

func main() {

	dsn := os.Getenv("MYSQL_DSN")

	if dsn == "" {

		//dsn = "backend-programmer"

	}

	var err error

	db, err = sql.Open("mysql", dsn)

	if err != nil {

		log.Fatalf("Cannot Open Db, %v", err)

	}

	defer db.Close()

	// testing the connection

	if err := db.Ping(); err != nil {

		log.Fatalf("Cannot Connect to DB %v", err)

	}

	fmt.Println("SQL Connect Succesfully")

	route := mux.NewRouter()

	route.HandleFunc("/users", GetUsers).Methods("GET")
	route.HandleFunc("/users/{id}", GetUser).Methods("GET")
	route.HandleFunc("/users", createUser).Methods("POST")
	route.HandleFunc("/users/{id}", updateUser).Methods("PUT")
	route.HandleFunc("/users/{id}", deleteUser).Methods("DELETE")

	fmt.Println("Server Starting On PORT:8080")
	log.Fatal(http.ListenAndServe(":8080", route))

}

// Get all users

func GetUsers(w http.ResponseWriter, r *http.Request) {

	rows, err := db.Query("SELECT id,namee,email FROM users")

	if err != nil {

		http.Error(w, err.Error(), http.StatusInternalServerError) // db connection ke doran error aya to wo yahan per handle howa by giving 500 error

		return
	}

	defer rows.Close() // close the rows if error comes so that we ensure that memory dosent leak

	users := []User{} // empty slice ke ander user ko store kiay

	for rows.Next() {

		var u User // create variable of User struct

		if err := rows.Scan(&u.Id, &u.Name, &u.Email); err != nil { // then read the data from the row of struct if error occur then give error

			http.Error(w, err.Error(), http.StatusInternalServerError)

		}

		users = append(users, u)

	}

	// what data show to user

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(users) // users data show in the form of json by encoding it
}

// now next function is GetUsers{id} with their specifc id

func GetUser(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r) // incoming url r se jo id ayegi us ko pakra

	id, _ := strconv.Atoi(params["id"]) // string to int converstion

	var u User

	err := db.QueryRow("SELECT id,namee,email FROM users where id = ?", id).Scan(&u.Id, &u.Name, &u.Email) // stores data in user strut by getting one specifc row with specic id from mysql db

	// 2 conditions aik to user ho hi na with specifc id or dosra koi or a skta ha

	if err != nil {

		if err == sql.ErrNoRows {

			http.Error(w, "User Not Found", http.StatusNotFound) // give 400 error

		} else {

			http.Error(w, err.Error(), http.StatusInternalServerError) // give 500 errro to the client side

		}

		return

	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(u) // converts the data to json to show the client

}
func createUser(w http.ResponseWriter, r *http.Request) {
	var users []User
	err := json.NewDecoder(r.Body).Decode(&users) // decode array
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	for _, u := range users {
		_, err := db.Exec("INSERT INTO users(namee, email) VALUES(?, ?)", u.Name, u.Email)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// now Update User function with its specifc id

func updateUser(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]

	var u User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Make sure column names match your DB data
	_, err := db.Exec("UPDATE users SET namee=?, email=? WHERE id=?", u.Name, u.Email, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	u.Id, _ = strconv.Atoi(id)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(u)
}

// last function for deleting a specific data with respect to its id

func deleteUser(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r) // get id from the url request

	id, _ := strconv.Atoi(params["id"]) // convert string to int

	_, err := db.Exec("DELETE FROM users where id=?", id) // delete from sepecic id

	if err != nil {

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent) // means give 204 that data deleted

}
