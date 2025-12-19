package hybridsystem

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// create and get users for mysql Databases with redis
func (a *HybridHandler3) createUserHandler3(w http.ResponseWriter, r *http.Request) {
	var users User2
	if err := json.NewDecoder(r.Body).Decode(&users); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	res, err := a.MySQL.DB.Exec("INSERT INTO users (name , email) VALUES (? , ?)", users.Name, users.Email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	id, _ := res.LastInsertId()
	users.ID = int(id)

	json.NewEncoder(w).Encode(users)
}

func (a *HybridHandler3) GetUserHandler3(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	value, err := a.Redis.Client.Get(a.Ctx, id).Result()
	if err == nil {
		log.Println("cache hit")
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(value))
		return
	}
	log.Println("cache miss, guerying mysql... ")
	row := a.MySQL.DB.QueryRow("SELECT id ,name , email FROM users WHERE id=?", id)

	var users User2
	if err := row.Scan(&users.ID, &users.Name, &users.Email); err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}
	jsonData, _ := json.Marshal(users)
	a.Redis.Client.Set(a.Ctx, id, jsonData, 10*time.Minute)

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

// create and Get users for mongodb with redis
func (h *HybridHandler3) CreateUserHandlers4(w http.ResponseWriter, r *http.Request) {
	var persons Person
	if err := json.NewDecoder(r.Body).Decode(&persons); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	ctx, cancel := context.WithTimeout(h.Ctx, 5*time.Second)
	defer cancel()

	res, err := h.Mongo.Users.InsertOne(ctx, persons)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	persons.ID = res.InsertedID.(primitive.ObjectID)

	jsonData, _ := json.Marshal(persons)
	h.Redis.Client.Set(h.Ctx, persons.ID.Hex(), jsonData, 10*time.Minute)

	json.NewEncoder(w).Encode(persons)
}
func (h *HybridHandler3) GetUserHandler4(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	value, err := h.Redis.Client.Get(h.Ctx, id).Result()
	if err == nil {
		log.Println("Cache hit")
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(value))
		return
	}
	log.Println("cache miss, querying MongoDB...")
	objID, _ := primitive.ObjectIDFromHex(id)
	var persons Person
	ctx, cancel := context.WithTimeout(h.Ctx, 5*time.Second)
	defer cancel()

	err = h.Mongo.Users.FindOne(ctx, bson.M{"_id": objID}).Decode(&persons)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsondata, _ := json.Marshal(persons)
	h.Redis.Client.Set(h.Ctx, id, jsondata, 10*time.Minute)

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsondata)
}
