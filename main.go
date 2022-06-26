package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type App struct {
	Router *chi.Mux;
	DB     *sqlx.DB;
}

const (
	port = 8080
)

func (a *App) Initialize(user, password, dbname string) {
	connectionString := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", user, password, dbname)

	var err error
	a.DB, err = sqlx.Connect("postgres", connectionString)
	if err != nil {
		log.Fatal(err)
	}

	r := chi.NewRouter()

  r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

  r.Mount("/users", UsersResource{DB: a.DB}.Routes())

  a.Router = r
}

func (a *App) Run(port int) {
	fmt.Println(fmt.Sprintf("Listening at port %d", port))
	http.ListenAndServe(fmt.Sprintf(":%d", port), a.Router)
}

func main() {
  a := App{}
	a.Initialize("postgres", "postgres", "gophernews")
		// os.Getenv("APP_DB_USERNAME"),
		// os.Getenv("APP_DB_PASSWORD"),
		// os.Getenv("APP_DB_NAME"))
	a.Run(8080)
}
