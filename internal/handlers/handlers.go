package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/vancho-go/effectivemobile-testtask/internal/db"
	"net/http"
	"strconv"
)

type Person struct {
	Name       string `json:"name"`
	Surname    string `json:"surname"`
	Patronymic string `json:"patronymic,omitempty"`
}

type EnrichedPerson struct {
	Person
	Age         int    `json:"age"`
	Gender      string `json:"gender"`
	Nationality string `json:"nationality"`
}

type AgeResponse struct {
	Count int    `json:"count"`
	Name  string `json:"name"`
	Age   int    `json:"age"`
}

type GenderResponse struct {
	Count       int     `json:"count"`
	Name        string  `json:"name"`
	Gender      string  `json:"gender"`
	Probability float64 `json:"probability"`
}

type NationalityResponse struct {
	Count   int    `json:"count"`
	Name    string `json:"name"`
	Country []struct {
		CountryId   string  `json:"country_id"`
		Probability float64 `json:"probability"`
	} `json:"country"`
}

func enrichPerson(person Person) (EnrichedPerson, error) {
	enrichedPerson := EnrichedPerson{
		Person: Person{
			Name:       person.Name,
			Surname:    person.Surname,
			Patronymic: person.Patronymic,
		},
	}

	age, err := getAge(enrichedPerson.Name)
	if err != nil {
		return enrichedPerson, err
	}
	enrichedPerson.Age = age

	gender, err := getGender(person.Name)
	if err != nil {
		return enrichedPerson, err
	}
	enrichedPerson.Gender = gender

	nationality, err := getNationality(person.Name)
	if err != nil {
		return enrichedPerson, err
	}
	enrichedPerson.Nationality = nationality

	return enrichedPerson, nil
}

func getAge(name string) (int, error) {
	url := fmt.Sprintf("https://api.agify.io/?name=%s", name)

	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var ageResponse AgeResponse
	err = json.NewDecoder(resp.Body).Decode(&ageResponse)
	if err != nil {
		return 0, err
	}

	return ageResponse.Age, nil
}

func getGender(name string) (string, error) {
	url := fmt.Sprintf("https://api.genderize.io/?name=%s", name)

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var genderResponse GenderResponse
	err = json.NewDecoder(resp.Body).Decode(&genderResponse)
	if err != nil {
		return "", err
	}
	return genderResponse.Gender, nil
}

func getNationality(name string) (string, error) {
	url := fmt.Sprintf("https://api.nationalize.io/?name=%s", name)

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var nationalityResponse NationalityResponse
	err = json.NewDecoder(resp.Body).Decode(&nationalityResponse)
	if err != nil {
		return "", err
	}

	if len(nationalityResponse.Country) > 0 {
		return nationalityResponse.Country[0].CountryId, nil
	}

	return "", nil
}
func addPerson(ep EnrichedPerson) error {
	_, err := db.DB.Conn.Exec("INSERT INTO people (name, surname, patronymic, age, gender, nationality) VALUES ($1, $2, $3, $4, $5, $6)",
		ep.Name, ep.Surname, ep.Patronymic, ep.Age, ep.Gender, ep.Nationality)
	return err
}

func deletePerson(id int) error {
	_, err := db.DB.Conn.Exec("DELETE FROM people WHERE id = $1",
		id)
	return err
}

func getPerson(id int) (EnrichedPerson, error) {
	var ep EnrichedPerson
	query := "SELECT name, surname, patronymic, age, nationality, gender FROM people WHERE id = $1"
	row := db.DB.Conn.QueryRow(query, id)

	err := row.Scan(&ep.Name, &ep.Surname, &ep.Patronymic, &ep.Age, &ep.Nationality, &ep.Gender)
	if err != nil {
		if err == sql.ErrNoRows {
			return ep, errors.New("Person not found")
		} else {
			return ep, errors.New("Database query error")
		}
	}
	return ep, nil
}

func updatePerson(id int, person EnrichedPerson) error {
	query := "UPDATE people SET name=$1, surname=$2, patronymic=$3, age=$4, gender=$5, nationality=$6 WHERE id=$7"
	_, err := db.DB.Conn.Exec(query, person.Name, person.Surname, person.Patronymic, person.Age, person.Gender, person.Nationality, id)
	if err != nil {
		return errors.New("Database update error")
	}
	return nil
}

func GetPeopleHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Something wrong with ID", http.StatusBadRequest)
	}
	person, err := getPerson(id)
	if err != nil {
		http.Error(w, "Can't get person with this ID", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(person)
}

func CreatePersonHandler(w http.ResponseWriter, r *http.Request) {
	var person Person
	err := json.NewDecoder(r.Body).Decode(&person)
	if err != nil {
		http.Error(w, "Problem with decoding person", http.StatusBadRequest)
		return
	}
	if person.Name == "" || person.Surname == "" {
		http.Error(w, "Invalid person data", http.StatusBadRequest)
		return
	}
	enrichedPerson, err := enrichPerson(person)
	if err != nil {
		http.Error(w, "Problem enriching person", http.StatusBadRequest)
	}
	err = addPerson(enrichedPerson)
	if err != nil {
		http.Error(w, "Error adding person", http.StatusBadRequest)
	}
	w.WriteHeader(http.StatusOK)
}

func UpdatePersonHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Something wrong with ID", http.StatusBadRequest)
	}

	var person EnrichedPerson
	err = json.NewDecoder(r.Body).Decode(&person)
	if err != nil {
		http.Error(w, "Error in Json", http.StatusBadRequest)
		return
	}
	if person.Name == "" || person.Surname == "" || person.Gender == "" || person.Age == 0 || person.Nationality == "" {
		http.Error(w, "Not all fields presented in JSON", http.StatusBadRequest)
		return
	}
	err = updatePerson(id, person)
	if err != nil {
		http.Error(w, "Database update error", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func DeletePersonHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Something wrong with ID", http.StatusBadRequest)
	}
	err = deletePerson(id)
	if err != nil {
		http.Error(w, "Can't delete person with this ID", http.StatusBadRequest)
	}
	w.WriteHeader(http.StatusOK)
}
