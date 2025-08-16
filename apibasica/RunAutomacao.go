package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type Task struct {
	ID      int
	Name    string
	Type    string // "api" ou "db"
	Payload string // parâmetros para a automação
	Status  string // pendente, executando, concluída, erro
}

func RunAutomacao() {
	// Conexão com o banco
	db, err := sql.Open("mysql", "root@tcp(127.0.0.1:3306)/golangapi")
	if err != nil {
		fmt.Println("Erro ao conectar no banco:", err)
		return
	}
	defer db.Close()

	rows, err := db.Query("SELECT id, name, type, payload, status FROM tasks WHERE status = 'pendente'")
	if err != nil {
		fmt.Println("Erro ao buscar tarefas:", err)
		return
	}
	defer rows.Close()

	tasks := []Task{}
	for rows.Next() {
		t := Task{}
		if err := rows.Scan(&t.ID, &t.Name, &t.Type, &t.Payload, &t.Status); err == nil {
			tasks = append(tasks, t)
		}
	}

	for _, task := range tasks {
		fmt.Printf("Executando tarefa %d: %s\n", task.ID, task.Name)
		db.Exec("UPDATE tasks SET status = 'executando' WHERE id = ?", task.ID)

		if task.Type == "api" {
			executeAPI(task)
		} else if task.Type == "db" {
			executeDB(db, task)
		}

		db.Exec("UPDATE tasks SET status = 'concluida' WHERE id = ?", task.ID)
		fmt.Printf("Tarefa %d concluída!\n", task.ID)
	}
}

func executeAPI(task Task) {
	var params map[string]string
	json.Unmarshal([]byte(task.Payload), &params)
	endpoint := params["endpoint"]
	method := params["method"]
	if endpoint == "" {
		endpoint = "http://localhost:8080/hello"
	}
	if method == "" {
		method = "GET"
	}

	req, err := http.NewRequest(method, endpoint, nil)
	if err != nil {
		fmt.Println("Erro ao criar request:", err)
		return
	}
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Erro ao chamar API:", err)
		return
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Printf("Resposta da API: %s\n", string(body))
}

func executeDB(db *sql.DB, task Task) {
	_, err := db.Exec(task.Payload)
	if err != nil {
		fmt.Println("Erro ao executar SQL:", err)
	} else {
		fmt.Println("Comando SQL executado com sucesso!")
	}
}
