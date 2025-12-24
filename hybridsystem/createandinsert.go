package hybridsystem

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func ValidateUser(user User2) error {
	if user.Email == "" {
		return fmt.Errorf("email is invalid and empty")
	}
	if strings.TrimSpace(user.Name) == "" {
		return fmt.Errorf("name is invalid and empty")
	}
	if !strings.HasSuffix(user.Email, "@gmail.com") {
		return fmt.Errorf("email is invalid and does not contain @gmail.com")
	}
	prefix := strings.TrimSuffix(user.Email, "@gmail.com")
	if prefix == "" {
		return fmt.Errorf("email must contains a prefix before the @gmail.com ")
	}
	return nil
}

// create and get users for mysql Databases with redis
func (a *HybridHandler3) CreateUserHandler3(w http.ResponseWriter, r *http.Request) {
	var users User2
	if err := json.NewDecoder(r.Body).Decode(&users); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := ValidateUser(users); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"Error": err.Error()})
		return
	}
	res, err := a.MySQL.DB.Exec("INSERT INTO users (name , email) VALUES (? , ?)", users.Name, users.Email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	id, err := res.LastInsertId()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	users.ID = int(id)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
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
	jsonData, err := json.Marshal(users)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	a.Redis.Client.Set(a.Ctx, id, jsonData, 10*time.Minute)

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}
func ValidateUser1(person Person) error {
	if person.Email == "" {
		return fmt.Errorf("email is invalid and empty")
	}
	if strings.TrimSpace(person.Name) == "" {
		return fmt.Errorf("name is invalid and empty")
	}
	if !strings.HasSuffix(person.Email, "@gmail.com") {
		return fmt.Errorf("email is invalid and does not contain @gmail.com")
	}
	prefix := strings.TrimSuffix(person.Email, "@gmail.com")
	if prefix == "" {
		return fmt.Errorf("email must contains a prefix before the @gmail.com ")
	}
	return nil
}

// create and Get users for mongodb with redis
func (h *HybridHandler3) CreateUserHandlers4(w http.ResponseWriter, r *http.Request) {
	var persons Person
	if err := json.NewDecoder(r.Body).Decode(&persons); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := ValidateUser1(persons); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"Error": err.Error()})
		return
	}
	ctx, cancel := context.WithTimeout(h.Ctx, 5*time.Second)
	defer cancel()

	res, err := h.Mongo.Persons.InsertOne(ctx, persons)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	persons.ID = res.InsertedID.(primitive.ObjectID)

	jsonData, _ := json.Marshal(persons)
	h.Redis.Client.Set(h.Ctx, persons.ID.Hex(), jsonData, 10*time.Minute)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(persons)
	w.WriteHeader(http.StatusCreated)

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
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		http.Error(w, "invalid id format", http.StatusBadRequest)
		return
	}
	var persons Person
	ctx, cancel := context.WithTimeout(h.Ctx, 5*time.Second)
	defer cancel()

	err = h.Mongo.Persons.FindOne(ctx, bson.M{"_id": objID}).Decode(&persons)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsondata, err := json.Marshal(persons)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.Redis.Client.Set(h.Ctx, id, jsondata, 10*time.Minute)

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsondata)
}
