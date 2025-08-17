package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

type Task struct {
	ID          int    `json:"id"`
	Text        string `json:"text"`
	Description string `json:"description"`
	Date        string `json:"date"`
	Time        string `json:"time"`
	Done        bool   `json:"done"`
}

var db *sql.DB

func main() {
	var err error
	db, err = sql.Open("mysql", "root@tcp(127.0.0.1:3306)/todolist")
	if err != nil {
		log.Fatal("Erro ao conectar no MySQL:", err)
	}
	defer db.Close()
	http.HandleFunc("/api/tasks", tasksHandler)
	http.HandleFunc("/api/tasks/", taskHandler)
	http.Handle("/", http.FileServer(http.Dir("./frontend/build")))
	log.Println("API rodando em http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func tasksHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case "GET":
		rows, err := db.Query("SELECT id, text, description, date, time, done FROM tasks")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()
		tasks := make([]Task, 0)
		for rows.Next() {
			var t Task
			err := rows.Scan(&t.ID, &t.Text, &t.Description, &t.Date, &t.Time, &t.Done)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			tasks = append(tasks, t)
		}
		json.NewEncoder(w).Encode(tasks)
	case "POST":
		var t Task
		if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		res, err := db.Exec("INSERT INTO tasks (text, description, date, time, done) VALUES (?, ?, ?, ?, ?)", t.Text, t.Description, t.Date, t.Time, t.Done)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		id, _ := res.LastInsertId()
		t.ID = int(id)
		json.NewEncoder(w).Encode(t)
	default:
		http.Error(w, "Método não suportado", http.StatusMethodNotAllowed)
	}
}

func taskHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id := 0
	_, err := fmt.Sscanf(r.URL.Path, "/api/tasks/%d", &id)
	if err != nil || id == 0 {
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}
	switch r.Method {
	case "PUT":
		var t Task
		if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		_, err := db.Exec("UPDATE tasks SET text=?, description=?, date=?, time=?, done=? WHERE id=?", t.Text, t.Description, t.Date, t.Time, t.Done, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		t.ID = id
		json.NewEncoder(w).Encode(t)
	case "DELETE":
		_, err := db.Exec("DELETE FROM tasks WHERE id=?", id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		http.Error(w, "Método não suportado", http.StatusMethodNotAllowed)
	}
}
