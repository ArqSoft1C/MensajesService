// test
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"goji.io"
	"goji.io/pat"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var Newid int

func ErrorWithJSON(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	fmt.Fprintf(w, "{message: %q}", message)
}

func ResponseWithJSON(w http.ResponseWriter, json []byte, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	w.Write(json)
}

type MESSAGE struct {
	ID      bson.ObjectId `bson:"_id,omitempty"`
	User1   string        `json:"user1"`
	User2   string        `json:"user2"`
	Subject string        `json:"subject"`
	Content string        `json:"content"`
}

func main() {

	session, err := mgo.Dial("messages-db:27017")
	if err != nil {
		fmt.Printf("no funciono")
		panic(err)
	}
	defer session.Close()

	session.SetMode(mgo.Monotonic, true)

	mux := goji.NewMux()

	mux.HandleFunc(pat.Get("/message"), allMessages(session))
	mux.HandleFunc(pat.Post("/message"), addMessage(session))
	//mux.HandleFunc(pat.Put("/mensajes/:id"), updateBook(session))
	mux.HandleFunc(pat.Delete("/message/:id"), deleteMessage(session))

	http.ListenAndServe(":4003", mux)
	//s := &http.Server{
	//	Addr:           ":4003",
	//	Handler:        mux,
	//	ReadTimeout:    10 * time.Second,
	//	WriteTimeout:   10 * time.Second,
	//	MaxHeaderBytes: 1 << 20,
	//}

	//s.ListenAndServe()

}

func addMessage(s *mgo.Session) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		session := s.Copy()
		defer session.Close()

		var message MESSAGE

		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&message)
		if err != nil {
			ErrorWithJSON(w, "Incorrect body", http.StatusBadRequest)
			return
		}
		// connect AutoIncrement to collection "counters"

		c := session.DB("Message_db").C("mensajes")
		message.ID = bson.NewObjectId()
		err = c.Insert(message)
		if err != nil {
			if mgo.IsDup(err) {
				ErrorWithJSON(w, "message with this ID exists", http.StatusBadRequest)
				return
			}

			ErrorWithJSON(w, "Database error", http.StatusInternalServerError)
			log.Println("Failed insert message: ", err)
			return
		}
		respBody, err := json.MarshalIndent(message, "", "  ")
		if err != nil {
			log.Fatal(err)
		}
		ResponseWithJSON(w, respBody, http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Location", r.URL.Path+"/"+string(message.ID))
		w.WriteHeader(http.StatusCreated)
	}
}

func allMessages(s *mgo.Session) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		session := s.Copy()
		defer session.Close()

		c := session.DB("Message_db").C("mensajes")

		var messages []MESSAGE
		err := c.Find(bson.M{}).All(&messages)
		if err != nil {
			ErrorWithJSON(w, "Database error", http.StatusInternalServerError)
			log.Println("Failed get all messages: ", err)
			return
		}

		respBody, err := json.MarshalIndent(messages, "", "  ")
		if err != nil {
			log.Fatal(err)
		}

		ResponseWithJSON(w, respBody, http.StatusOK)
	}
}

func deleteMessage(s *mgo.Session) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		session := s.Copy()
		defer session.Close()

		id := pat.Param(r, "id")
		//bsonObjectID := bson.ObjectIdHex(id)

		c := session.DB("Message_db").C("mensajes")

		err := c.Remove(bson.M{"_id": bson.ObjectIdHex(string(id))})
		if err != nil {
			switch err {
			default:
				ErrorWithJSON(w, "Database error", http.StatusInternalServerError)
				log.Println("Failed deleting message: ", err)
				return
			case mgo.ErrNotFound:
				ErrorWithJSON(w, "message not found", http.StatusNotFound)
				return
			}
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

/*
func ensureIndex(s *mgo.Session) {
	session := s.Copy()
	defer session.Close()

	c := session.DB("Message_db").C("mensajes")

	index := mgo.Index{
		Key:        []string{"id"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}
	err := c.EnsureIndex(index)
	if err != nil {
		panic(err)
	}
}
*/
