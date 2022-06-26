package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
  log "github.com/sirupsen/logrus"
)

type App struct {
	Router *chi.Mux
	DB     *sqlx.DB
	err    error
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

	if err := a.DB.Ping(); err != nil {
		a.err = fmt.Errorf("error pinging database: %s", err)
	}

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	usersRepo := NewUsersRepository(a.DB)
	r.Mount("/users", UsersResource{Repository: usersRepo}.Routes())

	a.Router = r
}

func (a *App) Run(port int) {
	if a.err != nil {
		return
	}
	log.Println(fmt.Sprintf("Listening at port %d", port))
	twoMinutes := time.Minute * 2
  srv := &http.Server{ReadTimeout: twoMinutes, WriteTimeout: twoMinutes,
    IdleTimeout: twoMinutes, Addr: fmt.Sprintf(":%d", port), Handler: a.Router}
	log.Fatal(srv.ListenAndServe())
}

func main() {
	a := App{}
	a.Initialize("postgres", "postgres", "gophernews")
	defer func() {
		if err := a.DB.Close(); err != nil {
			a.err = fmt.Errorf("error closing database: %s", err)
		}
	}()
	a.Run(8080)
	if a.err != nil {
		log.Fatalf("error running application: %s", a.err)
	}
}
