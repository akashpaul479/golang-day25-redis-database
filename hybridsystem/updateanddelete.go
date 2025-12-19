package hybridsystem

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// update and delete users using mysql with redis
func (a *HybridHandler3) UpdateUserHandler3(w http.ResponseWriter, r *http.Request) {
	var users User2
	if err := json.NewDecoder(r.Body).Decode(&users); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	_, err := a.MySQL.DB.Exec("UPDATE users SET name=?,email=? WHERE id=?", users.Name, users.Email, users.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonData, _ := json.Marshal(users)
	a.Redis.Client.Set(a.Ctx, fmt.Sprint(users.ID), jsonData, 10*time.Minute)
}
func (a *HybridHandler3) DeleteUserHandler3(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	_, err := a.MySQL.DB.Exec("DELETE FROM users WHERE id=?", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	a.Redis.Client.Del(a.Ctx, id)

	w.Write([]byte("User deleted"))
}

// update and delete users using mongoDB with redis
func (h *HybridHandler3) UpdateUserHandler4(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var persons Person
	if err := json.NewDecoder(r.Body).Decode(&persons); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	objID, _ := primitive.ObjectIDFromHex(id)
	ctx, cancel := context.WithTimeout(h.Ctx, 5*time.Second)
	defer cancel()
	update := bson.M{
		"$set": bson.M{
			"name":  persons.Name,
			"email": persons.Email,
		},
	}

	res, err := h.Mongo.Users.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if res.MatchedCount == 0 {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}
	persons.ID = objID
	jsonData, _ := json.Marshal(persons)
	h.Redis.Client.Set(h.Ctx, id, jsonData, 10*time.Minute)

	w.Header().Set("content-Type", "application/json")
	json.NewEncoder(w).Encode(persons)
}
func (h *HybridHandler3) DeleteuserHandler4(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	objID, _ := primitive.ObjectIDFromHex(id)
	ctx, cancel := context.WithTimeout(h.Ctx, 5*time.Second)
	defer cancel()

	_, err := h.Mongo.Users.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.Redis.Client.Del(h.Ctx, id)
	w.Write([]byte("user Deleted!"))
}
