package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

var db *sql.DB
var sessions = make(map[string]string) // token -> username

func helloHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}
	cookie, err := r.Cookie("session")
	if err != nil || cookie.Value == "" {
		http.Error(w, "Acesso negado: faça login para testar a API.", http.StatusUnauthorized)
		return
	}
	username, ok := sessions[cookie.Value]
	if !ok || username == "" {
		http.Error(w, "Sessão inválida. Faça login novamente.", http.StatusUnauthorized)
		return
	}
	fmt.Fprintln(w, "Hello, API!")
}

func hashHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}
	cookie, err := r.Cookie("session")
	if err != nil || cookie.Value == "" {
		http.Error(w, "Acesso negado: faça login para testar a API.", http.StatusUnauthorized)
		return
	}
	username, ok := sessions[cookie.Value]
	if !ok || username == "" {
		http.Error(w, "Sessão inválida. Faça login novamente.", http.StatusUnauthorized)
		return
	}
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

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}
	username := r.FormValue("username")
	password := r.FormValue("password")
	w.Header().Set("Content-Type", "application/json")
	if username == "" || password == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"success":false, "error":"Usuário ou senha não informados"}`)
		return
	}
	var hash string
	err := db.QueryRow("SELECT password_hash FROM users WHERE username = ?", username).Scan(&hash)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, `{"success":false, "error":"Usuário não encontrado"}`)
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, `{"success":false, "error":"Senha incorreta"}`)
		return
	}
	// Gera token de sessão único
	tokenBytes := make([]byte, 16)
	_, err = rand.Read(tokenBytes)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, `{"success":false, "error":"Erro ao gerar sessão"}`)
		return
	}
	token := hex.EncodeToString(tokenBytes)
	sessions[token] = username
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    token,
		Path:     "/",
		MaxAge:   3600,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	fmt.Fprint(w, `{"success":true}`)
}
func logoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")
	if err == nil && cookie.Value != "" {
		delete(sessions, cookie.Value)
		http.SetCookie(w, &http.Cookie{
			Name:     "session",
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, `{"success":true, "message":"Logout realizado"}`)
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"error":"Método GET não suportado para registro"}`)
		return
	}
	if r.Method == "POST" {
		w.Header().Set("Content-Type", "application/json")
		// Bloqueia registro se já estiver logado
		cookie, err := r.Cookie("session")
		if err == nil && cookie.Value != "" {
			usernameSess, ok := sessions[cookie.Value]
			if ok && usernameSess != "" {
				w.WriteHeader(http.StatusForbidden)
				fmt.Fprint(w, `{"success":false, "error":"Você já está logado. Não é possível registrar outro usuário."}`)
				return
			}
		}
		username := r.FormValue("username")
		password := r.FormValue("password")
		email := r.FormValue("email")
		if username == "" || password == "" || email == "" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `{"success":false, "error":"Usuário, e-mail ou senha não informados"}`)
			return
		}
		// Verifica se já existe username ou email
		var exists int
		errCheck := db.QueryRow("SELECT COUNT(*) FROM users WHERE username = ? OR email = ?", username, email).Scan(&exists)
		if errCheck != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, `{"success":false, "error":"Erro ao verificar usuário/email"}`)
			return
		}
		if exists > 0 {
			w.WriteHeader(http.StatusConflict)
			fmt.Fprint(w, `{"success":false, "error":"Usuário ou e-mail já registrado"}`)
			return
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, `{"success":false, "error":"Erro ao gerar hash"}`)
			return
		}
		_, err = db.Exec("INSERT INTO users (username, password_hash, email) VALUES (?, ?, ?)", username, hash, email)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, `{"success":false, "error":"Erro ao registrar usuário"}`)
			return
		}
		fmt.Fprint(w, `{"success":true, "message":"Usuário registrado com sucesso!"}`)
	}
}

func main() {
	var err error
	db, err = sql.Open("mysql", "root@tcp(127.0.0.1:3306)/golangapi")
	if err != nil {
		panic(err)
	}
	defer db.Close()
	// Cria tabela se não existir
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS users (
		id INT AUTO_INCREMENT PRIMARY KEY,
		username VARCHAR(255) UNIQUE,
		password_hash VARCHAR(255),
		email VARCHAR(255)
	)`)
	if err != nil {
		panic(err)
	}
	// Servir arquivos estáticos (CSS, imagens, etc)
	fs := http.FileServer(http.Dir("."))
	http.Handle("/style-land.css", fs)
	http.Handle("/style-auth.css", fs)
	http.Handle("/style-api.css", fs)
	http.Handle("/auth.html", fs)
	// Protege api.html: só serve se logado
	http.HandleFunc("/api.html", func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session")
		if err != nil || cookie.Value == "" {
			http.Redirect(w, r, "/auth.html", http.StatusFound)
			return
		}
		username, ok := sessions[cookie.Value]
		if !ok || username == "" {
			http.Redirect(w, r, "/auth.html", http.StatusFound)
			return
		}
		http.ServeFile(w, r, "api.html")
	})

	// Endpoint para verificar se está logado
	http.HandleFunc("/me", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		cookie, err := r.Cookie("session")
		if err != nil || cookie.Value == "" {
			fmt.Fprint(w, `{"logged":false}`)
			return
		}
		username, ok := sessions[cookie.Value]
		if !ok || username == "" {
			fmt.Fprint(w, `{"logged":false}`)
			return
		}
		fmt.Fprintf(w, `{"logged":true, "username":"%s"}`, username)
	})
	// ...existing code...
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/hello", helloHandler)
	http.HandleFunc("/hash", hashHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/logout", logoutHandler)
	// Endpoint para checar disponibilidade de username
	http.HandleFunc("/check-username", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		username := r.URL.Query().Get("username")
		if username == "" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `{"success":false, "error":"Username não informado"}`)
			return
		}
		var exists int
		err := db.QueryRow("SELECT COUNT(*) FROM users WHERE username = ?", username).Scan(&exists)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, `{"success":false, "error":"Erro ao consultar banco"}`)
			return
		}
		if exists > 0 {
			fmt.Fprint(w, `{"success":true, "available":false, "message":"Username já está em uso"}`)
		} else {
			fmt.Fprint(w, `{"success":true, "available":true, "message":"Username disponível"}`)
		}
	})

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/hello", helloHandler)
	http.HandleFunc("/hash", hashHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/logout", logoutHandler)
	fmt.Println("Servidor rodando em http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
