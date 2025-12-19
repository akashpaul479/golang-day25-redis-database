package redisDatabase

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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

func (a *App) createUserHandler(w http.ResponseWriter, r *http.Request) {
	var users User
	if err := json.NewDecoder(r.Body).Decode(&users); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
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
	_, err := a.DB.Exec("UPDATE users SET name=?,email=? WHERE id=?", users.Name, users.Email, users.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonData, _ := json.Marshal(users)
	a.RDB.Set(a.Ctx, fmt.Sprint(users.ID), jsonData, 10*time.Minute)
}
func (a *App) DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	_, err := a.DB.Exec("DELETE FROM users WHERE id=?", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	a.RDB.Del(a.Ctx, id)

	w.Write([]byte("User deleted"))
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
	r.HandleFunc("/users", app.createUserHandler).Methods("POST")
	r.HandleFunc("/users/{id}", app.GetUserHandler).Methods("GET")
	r.HandleFunc("/users/{id}", app.UpdateUserHandler).Methods("PUT")
	r.HandleFunc("/users/{id}", app.DeleteUserHandler).Methods("DELETE")

	log.Println("Server running on port :8080")
	http.ListenAndServe(":8080", r)

}
