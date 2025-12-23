package hybridsystem

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MySQLInstance1 struct {
	DB *sql.DB
}
type RedisInstance1 struct {
	Client *redis.Client
}
type MongoInstance1 struct {
	Client *mongo.Client
	DB     *mongo.Database
	Users  *mongo.Collection
}

type HybridHandler3 struct {
	Redis *RedisInstance1
	MySQL *MySQLInstance1
	Mongo *MongoInstance1
	Ctx   context.Context
}

type User2 struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type Person struct {
	ID    primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Name  string             `json:"name" bson:"name"`
	Email string             `json:"email" bson:"email"`
}

func Connectredis1() (*RedisInstance1, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_ADDR"),
		DB:   0,
	})
	return &RedisInstance1{Client: rdb}, nil
}
func ConnectMySQL1() (*MySQLInstance1, error) {
	db, err := sql.Open("mysql", os.Getenv("MYSQL_DSN"))
	if err != nil {
		return nil, err
	}
	return &MySQLInstance1{DB: db}, nil
}
func ConnectMongo1() (*MongoInstance1, error) {
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
	return &MongoInstance1{
		Client: client,
		DB:     db,
		Users:  db.Collection("users"),
	}, nil
}
func CRUDoperations2() {
	godotenv.Load()

	redisInstance, err := Connectredis1()
	if err != nil {
		log.Fatal(err)
	}
	mySQLInstance, err := ConnectMySQL1()
	if err != nil {
		log.Fatal(err)
	}
	mongoInstance, err := ConnectMongo1()
	if err != nil {
		log.Fatal(err)
	}
	handle := &HybridHandler3{Mongo: mongoInstance, MySQL: mySQLInstance, Redis: redisInstance, Ctx: context.Background()}
	r := mux.NewRouter()
	// for MySQL routes
	r.HandleFunc("/users", handle.CreateUserHandler3).Methods("POST")
	r.HandleFunc("/users/{id}", handle.GetUserHandler3).Methods("GET")
	r.HandleFunc("/users/{id}", handle.UpdateUserHandler3).Methods("PUT")
	r.HandleFunc("/users/{id}", handle.DeleteUserHandler3).Methods("DELETE")
	//  for MongoDB routes
	r.HandleFunc("/persons", handle.CreateUserHandlers4).Methods("POST")
	r.HandleFunc("/persons/{id}", handle.GetUserHandler4).Methods("GET")
	r.HandleFunc("/persons/{id}", handle.UpdateUserHandler4).Methods("PUT")
	r.HandleFunc("/persons/{id}", handle.DeleteuserHandler4).Methods("DELETE")

	log.Println("Server running on port :8080")
	http.ListenAndServe(":8080", r)
}
