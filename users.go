package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
)

type User struct {
	Id       string `db:"id" json:"id,omitempty"`
	Username string `db:"username" json:"username,omitempty"`
	Email    string `db:"email" json:"email,omitempty"`
}

type UserWithPassword struct {
  User
	Password string `json:"password"`
}

type UsersResource struct {
	DB *sqlx.DB
}

// Routes creates a REST router for the todos resource
func (rs UsersResource) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", rs.List)    // GET /todos - read a list of todos
	r.Post("/", rs.Create) // POST /todos - create a new todo and persist it
	r.Put("/", rs.Delete)

	r.Route("/{id}", func(r chi.Router) {
		r.Get("/", rs.Get)       // GET /todos/{id} - read a single todo by :id
		r.Put("/", rs.Update)    // PUT /todos/{id} - update a single todo by :id
		r.Delete("/", rs.Delete) // DELETE /todos/{id} - delete a single todo by :id
		r.Get("/sync", rs.Sync)
	})

	return r
}

func (rs UsersResource) List(w http.ResponseWriter, r *http.Request) {
	users := []User{}
	err := rs.DB.Select(&users, "select username, email, id from users")
	if err != nil {
		log.Fatal(err)
	}
	json, err := json.Marshal(users)
	if err != nil {
		log.Fatal(err)
	}
	w.Write(json)
}

func (rs UsersResource) Create(w http.ResponseWriter, r *http.Request) {
  userWithPassword := UserWithPassword{}
  err := json.NewDecoder(r.Body).Decode(&userWithPassword)
  if err != nil {
		log.Fatal(err)
  }
  w.Header().Set("content-type", "application/json")
  _, err = rs.DB.NamedExec(`insert into users (username,email,password_hash) values (:username,:email,:password)`, userWithPassword)
  if err != nil {
		log.Fatal(err)
  }
  json.NewEncoder(w).Encode(userWithPassword)
}

func (rs UsersResource) Get(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("todo get"))
}

func (rs UsersResource) Update(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("todo update"))
}

func (rs UsersResource) Delete(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("todo delete"))
}

func (rs UsersResource) Sync(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("todo sync"))
}
