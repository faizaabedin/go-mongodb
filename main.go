package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"../gopkg.in/mgo.v2"
	"../github.com/gorilla/context"
	"../gopkg.in/mgo.v2/bson"
)

func main() {

	// connect to the database
	db, err := mgo.Dial("localhost")
	if err != nil {
		log.Fatal("cannot dial mongo", err)
	}
	defer db.Close() // clean up when we're done

	// Adapt our handle function using withDB
	h := Adapt(http.HandlerFunc(handle), withDB(db))

	// add the handler
	http.Handle("/companies", context.ClearHandler(h))

	// start the server
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}

}

type Adapter func(http.Handler) http.Handler

func Adapt(h http.Handler, adapters ...Adapter) http.Handler {
	for _, adapter := range adapters {
		h = adapter(h)
	}
	return h
}

func handle(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		handleRead(w, r)
	case "POST":
		handleInsert(w, r)
	default:
		http.Error(w, "Not supported", http.StatusMethodNotAllowed)
	}
}

type company struct {
	ID     bson.ObjectId `json:"id" bson:"_id"`
	Name string        `json:"name" bson:"name"`
	Description   string        `json:"description" bson:"description"`
	Floor   int `json:"floor" bson:"floor"`
	Unit int `json:"unit" bson:"unit"`
	When   time.Time     `json:"when" bson:"when"`
}

func handleInsert(w http.ResponseWriter, r *http.Request) {
	db := context.Get(r, "database").(*mgo.Session)

	// decode the request body
	var c company
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// give the company a unique ID
	c.ID = bson.NewObjectId()
	c.When = time.Now()

	// insert it into the database
	if err := db.DB("companiesapp").C("companies").Insert(&c); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// redirect to it
	http.Redirect(w, r, "/companies/"+c.ID.Hex(), http.StatusTemporaryRedirect)
}
func handleRead(w http.ResponseWriter, r *http.Request) {
	db := context.Get(r, "database").(*mgo.Session)

	// load the companies
	var companies []*company
	if err := db.DB("companiesapp").C("companies").
		Find(nil).Sort("-when").Limit(100).All(&companies); err != nil {

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// write it out
	if err := json.NewEncoder(w).Encode(companies); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func withDB(db *mgo.Session) Adapter {

	// return the Adapter
	return func(h http.Handler) http.Handler {

		// the adapter (when called) should return a new handler
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			// copy the database session
			dbsession := db.Copy()
			defer dbsession.Close() // clean up

			// save it in the mux context
			context.Set(r, "database", dbsession)

			// pass execution to the original handler
			h.ServeHTTP(w, r)

		})
	}
}
