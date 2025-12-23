package redisDatabase

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
)

type App struct {
	DB  *sql.DB
	RDB *redis.Client
	Ctx context.Context
}

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func (a *App) CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	var users User
	if err := json.NewDecoder(r.Body).Decode(&users); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Validation
	if strings.TrimSpace(users.Name) == "" {
		http.Error(w, "name cannot be empty", http.StatusBadRequest)
		return
	}
	if !strings.Contains(users.Email, "@") || strings.TrimSpace(users.Email) == "" {
		http.Error(w, "invalid email", http.StatusBadRequest)
		return
	}
	prefix := strings.TrimSuffix(users.Email, "@gmail.com")
	if prefix == "" {
		http.Error(w, "Email must contains a prefix before @gmail.com ", http.StatusBadRequest)
		return
	}

	res, err := a.DB.Exec("INSERT INTO users (name , email) VALUES (? , ?)", users.Name, users.Email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	id, _ := res.LastInsertId()
	users.ID = int(id)

	json.NewEncoder(w).Encode(users)
}

func (a *App) GetUserHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	value, err := a.RDB.Get(a.Ctx, id).Result()
	if err == nil {
		log.Println("cache hit")
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(value))
		return
	}
	log.Println("cache miss, guerying mysql... ")
	row := a.DB.QueryRow("SELECT id ,name , email FROM users WHERE id=?", id)

	var users User
	if err := row.Scan(&users.ID, &users.Name, &users.Email); err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}
	jsonData, _ := json.Marshal(users)
	a.RDB.Set(a.Ctx, id, jsonData, 10*time.Minute)

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}
func (a *App) UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	var users User
	if err := json.NewDecoder(r.Body).Decode(&users); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	res, err := a.DB.Exec("UPDATE users SET name=?,email=? WHERE id=?", users.Name, users.Email, users.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}
	jsonData, _ := json.Marshal(users)
	a.RDB.Set(a.Ctx, fmt.Sprint(users.ID), jsonData, 10*time.Minute)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)

}
func (a *App) DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	idInt, _ := strconv.Atoi(id)

	res, err := a.DB.Exec("DELETE FROM users WHERE id=?", idInt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}
	a.RDB.Del(a.Ctx, id)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("user deleted"))
}
func Redisexample() {
	dsn := "root:root@tcp(127.0.0.1:3306)/go_users"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   0,
	})
	app := &App{
		DB:  db,
		RDB: rdb,
		Ctx: context.Background(),
	}
	r := mux.NewRouter()
	r.HandleFunc("/users", app.CreateUserHandler).Methods("POST")
	r.HandleFunc("/users/{id}", app.GetUserHandler).Methods("GET")
	r.HandleFunc("/users/{id}", app.UpdateUserHandler).Methods("PUT")
	r.HandleFunc("/users/{id}", app.DeleteUserHandler).Methods("DELETE")

	log.Println("Server running on port :8080")
	http.ListenAndServe(":8080", r)

}
