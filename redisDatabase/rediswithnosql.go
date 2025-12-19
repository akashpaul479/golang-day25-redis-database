package redisDatabase

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type RedisInstance struct {
	Client *redis.Client
}
type MongoInstance struct {
	Client *mongo.Client
	DB     *mongo.Database
	Users  *mongo.Collection
}

type HybridHandler struct {
	Redis *RedisInstance
	Mongo *MongoInstance
	Ctx   context.Context
}

type User1 struct {
	ID    primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Name  string             `json:"name" bson:"name"`
	Email string             `json:"email" bson:"email"`
}

func (h *HybridHandler) CreateUserHandlers1(w http.ResponseWriter, r *http.Request) {
	var users User1
	if err := json.NewDecoder(r.Body).Decode(&users); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	ctx, cancel := context.WithTimeout(h.Ctx, 5*time.Second)
	defer cancel()

	res, err := h.Mongo.Users.InsertOne(ctx, users)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	users.ID = res.InsertedID.(primitive.ObjectID)

	jsonData, _ := json.Marshal(users)
	h.Redis.Client.Set(h.Ctx, users.ID.Hex(), jsonData, 10*time.Minute)

	json.NewEncoder(w).Encode(users)
}
func (h *HybridHandler) GetUserHandler1(w http.ResponseWriter, r *http.Request) {
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
	var users User1
	ctx, cancel := context.WithTimeout(h.Ctx, 5*time.Second)
	defer cancel()

	err = h.Mongo.Users.FindOne(ctx, bson.M{"_id": objID}).Decode(&users)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsondata, _ := json.Marshal(users)
	h.Redis.Client.Set(h.Ctx, id, jsondata, 10*time.Minute)

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsondata)
}
func (h *HybridHandler) UpdateUserHandler1(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var users User1
	if err := json.NewDecoder(r.Body).Decode(&users); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	objID, _ := primitive.ObjectIDFromHex(id)
	ctx, cancel := context.WithTimeout(h.Ctx, 5*time.Second)
	defer cancel()
	update := bson.M{
		"$set": bson.M{
			"name":  users.Name,
			"email": users.Email,
		},
	}

	res, err := h.Mongo.Users.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if res.MatchedCount == 0 {
		http.Error(w, "user not found", http.StatusInternalServerError)
		return
	}
	users.ID = objID
	jsonData, _ := json.Marshal(users)
	h.Redis.Client.Set(h.Ctx, id, jsonData, 10*time.Minute)

	w.Header().Set("content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}
func (h *HybridHandler) DeleteuserHandler1(w http.ResponseWriter, r *http.Request) {
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
func Connectredis() (*RedisInstance, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_ADDR"),
		DB:   0,
	})
	return &RedisInstance{Client: rdb}, nil
}
func connectMongo() (*MongoInstance, error) {
	clientOPtions := options.Client().ApplyURI(os.Getenv("MONGO_URI"))
	client, err := mongo.NewClient(clientOPtions)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		return nil, err
	}
	db := client.Database(os.Getenv("MONGO_DB"))
	return &MongoInstance{
		Client: client,
		DB:     db,
		Users:  db.Collection("users"),
	}, nil
}
func CRUDoperations1() {
	godotenv.Load()

	redisInstance, err := Connectredis()
	if err != nil {
		log.Fatal(err)
	}
	mongoInstance, err := connectMongo()
	if err != nil {
		log.Fatal(err)
	}
	handle := &HybridHandler{Mongo: mongoInstance, Redis: redisInstance, Ctx: context.Background()}
	r := mux.NewRouter()

	r.HandleFunc("/users", handle.CreateUserHandlers1).Methods("POST")
	r.HandleFunc("/users/{id}", handle.GetUserHandler1).Methods("GET")
	r.HandleFunc("/users/{id}", handle.UpdateUserHandler1).Methods("PUT")
	r.HandleFunc("/users/{id}", handle.DeleteuserHandler1).Methods("DELETE")

	log.Println("Server running on port :8080")
	http.ListenAndServe(":8080", r)
}
