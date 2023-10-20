package main

import (
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/vancho-go/effectivemobile-testtask/internal/config"
	"github.com/vancho-go/effectivemobile-testtask/internal/db"
	"github.com/vancho-go/effectivemobile-testtask/internal/handlers"
	"github.com/vancho-go/effectivemobile-testtask/internal/logger"
	"net/http"
)

const flagLogLevel = "Debug"

func main() {
	err := logger.Initialize(flagLogLevel)
	if err != nil {
		panic(errors.New("error initializing logger"))
	}

	logger.Log.Info("Parsing env config")
	envVars, err := config.ParseEnvVars()
	if err != nil {
		panic(err)
	}

	err = db.Initialize(
		envVars.DBUsername,
		envVars.DBPassword,
		envVars.DBName,
		envVars.DBHost,
		envVars.DBPort,
	)
	if err != nil {
		panic(errors.New("error initializing DB"))
	}

	logger.Log.Info("Starting server")
	r := chi.NewRouter()
	r.Get("/{id}", logger.RequestLogger(handlers.GetPeopleHandler))
	r.Post("/", logger.RequestLogger(handlers.CreatePersonHandler))
	r.Put("/{id}", logger.RequestLogger(handlers.UpdatePersonHandler))
	r.Delete("/{id}", logger.RequestLogger(handlers.DeletePersonHandler))

	err = http.ListenAndServe(envVars.ServerHost, r)
	if err != nil {
		panic(errors.New("error starting server"))
	}
}
