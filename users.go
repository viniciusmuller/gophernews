package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator"
	log "github.com/sirupsen/logrus"
	renderPkg "github.com/unrolled/render"
)

var (
	render    = renderPkg.New()
	Validator = validator.New()
)

var (
	ValidationErrorType = "ValidationError"
	UniqueConstraint    = "UniqueConstraint"
)

type UserBase struct {
	Id       string `db:"id" json:"id,omitempty"`
	Username string `db:"username" json:"username,omitempty" validate:"required,min=3,max=20"`
	Email    string `db:"email" json:"email,omitempty" validate:"required,email,max=254"`
}

type UserId struct {
	Id string `validate:"required,uuid"`
}

type UserPrivate struct {
	UserBase
	PasswordHash         string
	CreationDate         time.Time
	LastModificationDate time.Time
}

type UserWithPassword struct {
	UserBase
	Password string `json:"password" validate:"required,min=8"`
}

type ErrorResponse struct {
	ErrorType string      `json:"errorType"`
	Data      interface{} `json:"data,omitempty"`
}

type UsersResource struct {
	Repository UsersRepositorier
}

func (rs UsersResource) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", rs.List)    // GET /todos - read a list of todos
	r.Post("/", rs.Create) // POST /todos - create a new todo and persist it

	r.Route("/{id}", func(r chi.Router) {
		r.Get("/", rs.Get)       // GET /todos/{id} - read a single todo by :id
		r.Put("/", rs.Update)    // PUT /todos/{id} - update a single todo by :id
		r.Delete("/", rs.Delete) // DELETE /todos/{id} - delete a single todo by :id
	})

	return r
}

func (rs UsersResource) List(w http.ResponseWriter, r *http.Request) {
	users, err := rs.Repository.ListUsers()
	if err != nil {
		log.Println(err)
		writeStatus(w, http.StatusInternalServerError)
		return
	}
	render.JSON(w, http.StatusOK, users)
}

func (rs UsersResource) Create(w http.ResponseWriter, r *http.Request) {
	userWithPassword := UserWithPassword{}
	err := json.NewDecoder(r.Body).Decode(&userWithPassword)
	if err != nil {
		writeStatus(w, http.StatusBadRequest)
		return
	}

	err = Validator.Struct(userWithPassword)
	if err != nil {
		errorResponse := handleValidationErr(err)
		render.JSON(w, http.StatusUnprocessableEntity, errorResponse)
		return
	}

	user, err := rs.Repository.CreateUser(userWithPassword)
	if err != nil {
		if errors.Is(err, ErrUniqueConstraint) {
			errorResponse := ErrorResponse{
				ErrorType: UniqueConstraint,
				Data:      err.Error(),
			}
			render.JSON(w, http.StatusUnprocessableEntity, errorResponse)
			return
		}

		log.Println(err)
		writeStatus(w, http.StatusInternalServerError)
		return
	}
	render.JSON(w, http.StatusCreated, user)
}

func (rs UsersResource) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	err := validateUserId(id)
	if err != nil {
		errorsResponse := handleValidationErr(err)
		render.JSON(w, http.StatusUnprocessableEntity, errorsResponse)
		return
	}

	user, err := rs.Repository.GetUser(id)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			writeStatus(w, http.StatusNotFound)
			return
		}

		log.Println(err)
		writeStatus(w, http.StatusInternalServerError)
		return
	}
	render.JSON(w, http.StatusOK, user)
}

func (rs UsersResource) Update(w http.ResponseWriter, r *http.Request) {
	user := UserBase{}
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		writeStatus(w, http.StatusBadRequest)
		return
	}

	var errorResponse ErrorResponse

	id := chi.URLParam(r, "id")
	err = validateUserId(id)
	if err != nil {
		errorResponse = handleValidationErr(err)
		render.JSON(w, http.StatusUnprocessableEntity, errorResponse)
		return
	}

	err = Validator.Struct(user)
	if err != nil {
		errorResponse = handleValidationErr(err)
		render.JSON(w, http.StatusUnprocessableEntity, errorResponse)
		return
	}

	user.Id = id
	user, err = rs.Repository.UpdateUser(user)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			writeStatus(w, http.StatusNotFound)
			return
		}

		if errors.Is(err, ErrUniqueConstraint) {
			errorResponse := ErrorResponse{
				ErrorType: UniqueConstraint,
				Data:      err.Error(),
			}
			render.JSON(w, http.StatusUnprocessableEntity, errorResponse)
      return
		}

		log.Println(err)
		writeStatus(w, http.StatusInternalServerError)
		return
	}
	render.JSON(w, http.StatusOK, user)
}

func (rs UsersResource) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	err := rs.Repository.DeleteUser(id)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			writeStatus(w, http.StatusNotFound)
			return
		}

		log.Println(err)
		writeStatus(w, http.StatusInternalServerError)
	}
}

func writeStatus(w http.ResponseWriter, statusCode int) {
	response := ErrorResponse{ErrorType: http.StatusText(statusCode)}
	render.JSON(w, statusCode, response)
}

func validateUserId(id string) error {
	uId := UserId{Id: id}
	return Validator.Struct(uId)
}

func handleValidationErr(err error) ErrorResponse {
	var validationErrors []string
	for _, e := range err.(validator.ValidationErrors) {
		validationErrors = append(validationErrors, fmt.Sprint(e))
	}
	return ErrorResponse{Data: validationErrors, ErrorType: ValidationErrorType}
}
