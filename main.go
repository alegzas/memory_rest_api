package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

type Score struct {
	Name  string `json:"name"`
	Score int    `json:"score"`
	Date  string `json:"date"`
}

type JsonResponse struct {
	Type    string      `json:"type"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type Scores []Score

var db *sql.DB

func main() {

	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	var err error
	dbInfo := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", dbUser, dbPassword, dbName)
	db, err = sql.Open("postgres", dbInfo)
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	router := mux.NewRouter()

	router.HandleFunc("/", index).Methods("GET")
	router.HandleFunc("/getScores", getScores).Methods("GET")
	router.HandleFunc("/sendScore", sendScore).Methods("POST")

	fmt.Println("Running on port 60")
	log.Fatal(http.ListenAndServe(":60", router))

}

func index(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello World"))
}

func getScores(w http.ResponseWriter, r *http.Request) {
	var scores Scores
	var response JsonResponse
	fmt.Println("Getting scores")
	rows, err := db.Query("SELECT name, score, DATE(date) FROM scores ORDER BY score DESC LIMIT 10")
	if err != nil {
		fmt.Println(err)
	}
	defer rows.Close()
	for rows.Next() {
		var score Score
		err := rows.Scan(&score.Name, &score.Score, &score.Date)
		if err != nil {
			response = JsonResponse{Type: "error", Message: "Error getting scores"}
			fmt.Println(err)

		}
		scores = append(scores, score)
	}
	err = rows.Err()
	if err != nil {
		response = JsonResponse{Type: "error", Message: "Error getting scores"}
		fmt.Println(err)
	}

	w.Header().Set("Content-Type", "application/json")
	response = JsonResponse{Type: "success", Message: "Scores retrieved", Data: scores}
	json.NewEncoder(w).Encode(response)
}

func sendScore(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	score := r.FormValue("score")
	var date string
	if r.FormValue("date") == "" {
		date = time.Now().Format("2006-01-02")
	} else {
		date = r.FormValue("date")
	}
	var response JsonResponse

	if name == "" || score == "" {
		response = JsonResponse{Type: "error", Message: "Missing name or score"}
	} else {
		fmt.Println("Inserting score for " + name + " with score " + score + " on " + date)
		response = JsonResponse{Type: "success", Message: "Score inserted"}
		_, err := db.Exec("INSERT INTO scores (name, score, date) VALUES ($1, $2, $3)", name, score, date)
		if err != nil {
			response = JsonResponse{Type: "error", Message: "Error inserting score"}
			fmt.Println(err)
		}

	}
	json.NewEncoder(w).Encode(response)
}
