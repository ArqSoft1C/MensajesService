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

type MENSAJE struct {
	ID        string `json:"id"`
	Usuario1  string `json:"usuario1"`
	Usuario2  string `json:"usuario2"`
	Asunto    string `json:"asunto"`
	Contenido string `json:"contenido"`
}

func main() {

	session, err := mgo.Dial("mongo_db:27017")
	if err != nil {
		fmt.Printf("no funciono")
		panic(err)
	}
	defer session.Close()

	session.SetMode(mgo.Monotonic, true)
	ensureIndex(session)

	mux := goji.NewMux()

	mux.HandleFunc(pat.Get("/mensajes"), allMessages(session))
	mux.HandleFunc(pat.Post("/mensajes"), addMessage(session))
	//mux.HandleFunc(pat.Put("/mensajes/:id"), updateBook(session))
	mux.HandleFunc(pat.Delete("/mensajes/:id"), deleteBook(session))

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

		var mensaje MENSAJE
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&mensaje)
		if err != nil {
			ErrorWithJSON(w, "Incorrect body", http.StatusBadRequest)
			return
		}

		c := session.DB("Message_db").C("mensajes")

		err = c.Insert(mensaje)
		if err != nil {
			if mgo.IsDup(err) {
				ErrorWithJSON(w, "Book with this ISBN already exists", http.StatusBadRequest)
				return
			}

			ErrorWithJSON(w, "Database error", http.StatusInternalServerError)
			log.Println("Failed insert book: ", err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Location", r.URL.Path+"/"+mensaje.ID)
		w.WriteHeader(http.StatusCreated)
	}
}

func allMessages(s *mgo.Session) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		session := s.Copy()
		defer session.Close()

		c := session.DB("Message_db").C("mensajes")

		var mensajes []MENSAJE
		err := c.Find(bson.M{}).All(&mensajes)
		if err != nil {
			ErrorWithJSON(w, "Database error", http.StatusInternalServerError)
			log.Println("Failed get all messages: ", err)
			return
		}

		respBody, err := json.MarshalIndent(mensajes, "", "  ")
		if err != nil {
			log.Fatal(err)
		}

		ResponseWithJSON(w, respBody, http.StatusOK)
	}
}

func deleteBook(s *mgo.Session) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		session := s.Copy()
		defer session.Close()

		id := pat.Param(r, "id")

		c := session.DB("Message_db").C("mensajes")

		err := c.Remove(bson.M{"id": id})
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