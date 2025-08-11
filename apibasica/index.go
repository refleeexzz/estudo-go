package main

import (
    "fmt"
    "net/http"
    "golang.org/x/crypto/bcrypt"
    "os"
)

func helloHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintln(w, "Hello, API!")
}

func hashHandler(w http.ResponseWriter, r *http.Request) {
    password := r.URL.Query().Get("password")
    if password == "" {
        http.Error(w, "Senha não informada", http.StatusBadRequest)
        return
    }
    hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        http.Error(w, "Erro ao gerar hash", http.StatusInternalServerError)
        return
    }
    fmt.Fprintf(w, "Hash gerado: %s", hash)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    data, err := os.ReadFile("index.html")
    if err != nil {
        http.Error(w, "Arquivo não encontrado", http.StatusInternalServerError)
        return
    }
    w.Write(data)
}

func main() {
    http.HandleFunc("/", indexHandler)
    http.HandleFunc("/hello", helloHandler)
    http.HandleFunc("/hash", hashHandler)
    fmt.Println("Servidor rodando em http://localhost:8080")
    http.ListenAndServe(":8080", nil)
}