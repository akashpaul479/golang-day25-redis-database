package hybridsystem

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
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
	if err := ValidateUser(users); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"Error": err.Error()})
		return
	}
	res, err := a.MySQL.DB.Exec("UPDATE users SET name=?,email=? WHERE id=?", users.Name, users.Email, users.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	rows, err := res.RowsAffected()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if rows == 0 {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}
	jsonData, err := json.Marshal(users)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	a.Redis.Client.Set(a.Ctx, fmt.Sprint(users.ID), jsonData, 10*time.Minute)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}
func (a *HybridHandler3) DeleteUserHandler3(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	idInt, _ := strconv.Atoi(id)

	res, err := a.MySQL.DB.Exec("DELETE FROM users WHERE id=?", idInt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	rows, err := res.RowsAffected()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if rows == 0 {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	a.Redis.Client.Del(a.Ctx, id)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	w.Write([]byte("user deleted"))
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
	if err := ValidateUser1(persons); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"Error": err.Error()})
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
	res, err := h.Mongo.Persons.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if res.MatchedCount == 0 {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}
	persons.ID = objID
	jsonData, err := json.Marshal(persons)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
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

	res, err := h.Mongo.Persons.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if res.DeletedCount == 0 {
		http.Error(w, "user not found", http.StatusNotFound)
	}

	h.Redis.Client.Del(h.Ctx, id)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("user Deleted!"))

}
