package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
)

type User struct {
	Name string `json:"name"`
}

// this acts as the in-memory "db"
var userCache = make(map[int]User)

// this enables many goroutines to READ data at the same time
// but only allows ONE goroutine to write data at a time
// this is only responsbile for the Lock() and so on "stuff"
var cacheMutex sync.RWMutex

var lastID int // for the in-memory "db" auto-inc bug incase of deletion

func main() {
	// creates the router called mux (Multiplexer)
	mux := http.NewServeMux()
	mux.HandleFunc("/", handleRoot)

	mux.HandleFunc("POST /users", createUser)
	mux.HandleFunc("GET /users/{id}", getUser)
	mux.HandleFunc("DELETE /users/{id}", deleteUser)

	fmt.Println("Server running on :8080")
	http.ListenAndServe(":8080", mux)
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello World\n")
}

func deleteUser(w http.ResponseWriter, r *http.Request) {
	// the converstion of the users req + handle the errors
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// if above is a success, look for the user in the map using the id provided
	// plus do the error checking
	if _, ok := userCache[id]; !ok {
		http.Error(w, "User not found", http.StatusBadRequest)
		return
	}
	// if the above is a success, write to the in-mem "db"
	// Lock() gives the responsible go routine write access and blocks other readers and writers
	cacheMutex.Lock()
	defer cacheMutex.Unlock() // once done, unlock read and write access
	delete(userCache, id)     // delete the user

	w.WriteHeader(http.StatusOK) // tell the clients browser everything went ok
}

func getUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// lock the WRITING PART ONLY, NO ONE can write to it (hence the R for read in the name)
	cacheMutex.RLock()
	defer cacheMutex.RUnlock() // unlock when done
	user, ok := userCache[id]
	if !ok {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	// if the above is successful, give back to the client
	// data in form of JSON as the header
	// NOTE: this doesn't send any bytes over the net yet...
	w.Header().Set("Content-Type", "application/json")
	// parse(encode) data to JSON (marshaling)
	data, err := json.Marshal(user)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// quick note: u only need to call w.WriteHeader only when u want to return a status code without..
	// an err body.e.g w.WriteHeader(http.StatusNoContent)
	w.Write(data) // this does the returning the status code + sends over the data via the internet
}

func createUser(w http.ResponseWriter, r *http.Request) {
	var user User
	// takes in the body we got from the requst and decodes it to JSON and decodes it...
	// to the struct
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if user.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}
	// lock mutation
	cacheMutex.Lock()
	lastID++                           // increament by 1
	defer cacheMutex.Unlock()          // unlock when below is done
	userCache[len(userCache)+1] = user // do the increamentation to the user
	w.WriteHeader(http.StatusCreated)  // send a created successfully status code

	// encode the response struct/data in JSON format and send it back to the client
	json.NewEncoder(w).Encode(user)
}
