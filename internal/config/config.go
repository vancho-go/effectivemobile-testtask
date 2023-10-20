package config

import (
	"errors"
	"os"
)

type Env struct {
	DBUsername string
	DBPassword string
	DBName     string
	DBHost     string
	DBPort     string
	ServerHost string
}

func ParseEnvVars() (Env, error) {
	var env Env

	DBUsername, DBPassword, DBName, DBHost, DBPort, ServerHost :=
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB"),
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("SERVER_HOST")

	//DBUsername, DBPassword, DBName, DBHost, DBPort, ServerHost :=
	//	"vancho",
	//	"vancho_pswd",
	//	"vancho_db",
	//	"localhost",
	//	"5432",
	//	"localhost:8080"

	if DBUsername != "" && DBPassword != "" && DBName != "" && DBHost != "" && DBPort != "" && ServerHost != "" {
		env.DBUsername = DBUsername
		env.DBPassword = DBPassword
		env.DBName = DBName
		env.DBHost = DBHost
		env.DBPort = DBPort
		env.ServerHost = ServerHost
		return env, nil
	}
	return env, errors.New("Missing environmental variables")
}
