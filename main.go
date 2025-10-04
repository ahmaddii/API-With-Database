package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
)

var db *sql.DB

// Create a secret key for jwt token access
var jwtkey = []byte("12345") // secret key

type User struct {
	Id    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Login function for generating jwt token
func Login(w http.ResponseWriter, r *http.Request) {

	var Creds credentials

	err := json.NewDecoder(r.Body).Decode(&Creds) // body se jo data aya use credes se pakr liya
	if err != nil {
		http.Error(w, "Invalid Request", http.StatusBadRequest)
		return
	}

	// just use hardcode later we do it with mysql
	if Creds.Email != "admin@gmail.com" || Creds.Password != "yaad12345" {
		http.Error(w, "Invalid Credentials", http.StatusUnauthorized)
		return
	}

	// otherwise if email password auth correct then create a jwt token with payload includes
	claims := jwt.MapClaims{
		"email": Creds.Email,
		"role":  "admin",
		"exp":   time.Now().Add(1 * time.Hour).Unix(), // after 1 hour ke bad expire hojaye ga token
	}

	// now create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(jwtkey) // signned the token with jwt key
	if err != nil {
		http.Error(w, "Error while Creating a Token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	// encode response with token
	json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
}

// use middleware: first one checks authentitcated user and validates the token
func authMiddelWare(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// yahan per hum dekhe ge token validation before passing to next handler
		tokenstr := r.Header.Get("Authorization") // get the header auth from http

		if tokenstr == "" {
			http.Error(w, "Missing Token", http.StatusUnauthorized) // this mean client dosent log in and dosent provide auth info thats why
			return
		}

		// validate token
		token, err := jwt.Parse(tokenstr, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok { // security check to verify the token was sign with HMAC or not
				return nil, fmt.Errorf("Not Expected Signing Method ")
			}
			return jwtkey, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Invalid Token Request", http.StatusUnauthorized)
			return
		}

		// if token valid then proceed to next handler
		next.ServeHTTP(w, r)
	})
}

func main() {

	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
<<<<<<< HEAD

		//dsn = "backend-programmer"

=======
		dsn = "backend-programmer:example-password@tcp(127.0.0.1:3306)/testdb"
>>>>>>> 9283fd4 (Jwt Auth and Authorization with Postman Check for Secure APIS)
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

	// Public Route for Auth
	route.HandleFunc("/login", Login).Methods("POST")

	// Protected Routes which require middleware jwt check
	// use Handle (not HandleFunc) because authMiddelWare returns http.Handler
	route.Handle("/users", authMiddelWare(http.HandlerFunc(GetUsers))).Methods("GET")
	route.Handle("/users/{id}", authMiddelWare(http.HandlerFunc(GetUser))).Methods("GET")
	route.Handle("/users", authMiddelWare(http.HandlerFunc(createUser))).Methods("POST")
	route.Handle("/users/{id}", authMiddelWare(http.HandlerFunc(updateUser))).Methods("PUT")
	route.Handle("/users/{id}", authMiddelWare(http.HandlerFunc(deleteUser))).Methods("DELETE")

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
		var u User                                                  // create variable of User struct
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

	params := mux.Vars(r)               // incoming url r se jo id ayegi us ko pakra
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

// create user
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

	params := mux.Vars(r)               // get id from the url request
	id, _ := strconv.Atoi(params["id"]) // convert string to int

	_, err := db.Exec("DELETE FROM users where id=?", id) // delete from sepecic id
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent) // means give 204 that data deleted
}
