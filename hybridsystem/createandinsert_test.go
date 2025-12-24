package hybridsystem_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	hybridsystem "redisDatabase/Hybridsystem"
	"strconv"
	"testing"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestHybridHandler3_CreateUserHandler3(t *testing.T) {
	tests := []struct {
		name     string // description of this test case
		user     hybridsystem.User2
		willpass bool
	}{
		{
			name: "valid name and valid email",
			user: hybridsystem.User2{
				Name:  "Akash",
				Email: "akash@gmail.com",
			},
			willpass: true,
		},
		{
			name: "invalid name and valid email",
			user: hybridsystem.User2{
				Name:  "",
				Email: "akash@gmail.com",
			},
			willpass: false,
		},
		{
			name: "valid name and invalid email",
			user: hybridsystem.User2{
				Name:  "Akash",
				Email: "",
			},
			willpass: false,
		},
		{
			name: "valid name and email without prefix",
			user: hybridsystem.User2{
				Name:  "Akash",
				Email: "@gmail.com",
			},
			willpass: false,
		},
		{
			name: "withspace only name and valid email",
			user: hybridsystem.User2{
				Name:  "   ",
				Email: "akash@gmail.com",
			},
			willpass: false,
		},
	}
	os.Setenv("REDIS_ADDR", "localhost:6379")
	os.Setenv("MYSQL_DSN", "root:root@tcp(127.0.0.1:3306)/go_users")

	redisInstance, err := hybridsystem.Connectredis1()
	if err != nil {
		log.Fatal(err)
	}
	mySQLInstance, err := hybridsystem.ConnectMySQL1()
	if err != nil {
		log.Fatal(err)
	}
	handle := &hybridsystem.HybridHandler3{MySQL: mySQLInstance, Redis: redisInstance}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userBytes, err := json.Marshal(tt.user)
			if err != nil {
				log.Panic("Failed to marshal", err)
			}
			buffer := bytes.NewBuffer(userBytes)
			r := httptest.NewRequest(http.MethodPost, "/users", buffer)
			w := httptest.NewRecorder()

			handle.CreateUserHandler3(w, r)

			if tt.willpass {
				if w.Code != http.StatusCreated {
					t.Fatalf("expected status OK, got %d", w.Code)
				}
				// validate response
				var user hybridsystem.User2
				if err := json.NewDecoder(w.Body).Decode(&user); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				if user.Name != tt.user.Name {
					t.Fatalf("expected name %s, got %s", tt.user.Name, user.Name)
				}
				if user.Email != tt.user.Email {
					t.Fatalf("expected email %s, got %s", tt.user.Email, user.Email)
				}
				if user.ID == 0 {
					t.Fatalf("expected non zero ID")
				}
			} else {
				if w.Code != http.StatusOK {
					t.Fatalf("expected status not OK, got %d", w.Code)

				}
			}
		})
	}
}

func TestHybridHandler3_GetUserHandler3(t *testing.T) {
	tests := []struct {
		name     string // description of this test case
		id       int
		willpass bool
	}{
		{
			name:     "valid id exsist in sql",
			id:       1,
			willpass: true,
		},
		{
			name:     "invalid id not found",
			id:       4567,
			willpass: false,
		},
	}
	os.Setenv("REDIS_ADDR", "localhost:6379")
	os.Setenv("MYSQL_DSN", "root:root@tcp(127.0.0.1:3306)/go_users")

	redisInstance, err := hybridsystem.Connectredis1()
	if err != nil {
		log.Fatal(err)
	}
	mySQLInstance, err := hybridsystem.ConnectMySQL1()
	if err != nil {
		log.Fatal(err)
	}
	handle := &hybridsystem.HybridHandler3{MySQL: mySQLInstance, Redis: redisInstance, Ctx: context.Background()}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// 🔹 CLEAN DB & REDIS
			handle.MySQL.DB.Exec("DELETE FROM users")
			handle.Redis.Client.FlushAll(handle.Ctx)

			userID := tt.id
			if tt.willpass {
				res, err := handle.MySQL.DB.Exec("INSERT INTO users (name , email) VALUES (? , ?)", "Akash", "akash@gmail.com")
				if err != nil {
					t.Fatal(err)
				}
				id, _ := res.LastInsertId()
				userID = int(id)
			}
			r := httptest.NewRequest(http.MethodGet, "/users/"+strconv.Itoa(userID), nil)
			r = mux.SetURLVars(r, map[string]string{"id": strconv.Itoa(userID)})
			w := httptest.NewRecorder()

			handle.GetUserHandler3(w, r)

			if tt.willpass {
				if w.Code != http.StatusOK {
					t.Fatalf("Expected status ok, got %d", w.Code)
				}
				var user hybridsystem.User2
				if err := json.NewDecoder(w.Body).Decode(&user); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				if user.ID != userID {
					t.Fatalf("Expected id %d , got %d", userID, user.ID)
				}
			} else {
				if w.Code != http.StatusNotFound {
					t.Fatalf("Expected status not ok , got %d", w.Code)
				}
			}
		})
	}
}

func TestHybridHandler3_CreateUserHandlers4(t *testing.T) {
	tests := []struct {
		name     string // description of this test case
		person   hybridsystem.Person
		willpass bool
	}{
		{
			name: "valid name and valid email",
			person: hybridsystem.Person{
				Name:  "Akash",
				Email: "akash@gmail.com",
			},
			willpass: true,
		},
		{
			name: "valid name and invalid email",
			person: hybridsystem.Person{
				Name:  "Akash",
				Email: "abc",
			},
			willpass: false,
		},
		{
			name: "valid name and invalid empty email",
			person: hybridsystem.Person{
				Name:  "rjesh",
				Email: "",
			},
			willpass: false,
		},
		{
			name: "valid name and invalid empty email",
			person: hybridsystem.Person{
				Name:  "akash",
				Email: "akash@yahoo.com",
			},
			willpass: false,
		},
		{
			name: "valid name and gmail without prefix",
			person: hybridsystem.Person{
				Name:  "akash",
				Email: "@gmail.com",
			},
			willpass: false,
		},
		{
			name: "invalid name and valid  email",
			person: hybridsystem.Person{
				Name:  "",
				Email: "abc@gmail.com",
			},
			willpass: false,
		},
		{
			name: "withspace-only name and valid  email",
			person: hybridsystem.Person{
				Name:  "   ",
				Email: "abc@gmail.com",
			},
			willpass: false,
		},
	}

	os.Setenv("MONGO_URI", "mongodb://localhost:27017")
	os.Setenv("MONGO_DB", "go_users")
	os.Setenv("REDIS_ADDR", "localhost:6379")

	redisInstance, err := hybridsystem.Connectredis1()
	if err != nil {
		log.Fatal(err)
	}
	mongoInstance, err := hybridsystem.ConnectMongo1()
	if err != nil {
		log.Fatal(err)
	}
	handle := &hybridsystem.HybridHandler3{Mongo: mongoInstance, Redis: redisInstance, Ctx: context.Background()}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userBytes, err := json.Marshal(tt.person)
			if err != nil {
				log.Panic("failed to marshal")
			}
			buffer := bytes.NewBuffer(userBytes)
			r := httptest.NewRequest(http.MethodPost, "localhost:8080/users", buffer)
			w := httptest.NewRecorder()
			handle.CreateUserHandlers4(w, r)

			if tt.willpass {
				if w.Code != http.StatusOK {
					t.Fatalf("Expected ok status , got %d", w.Code)
				}
				// validate response
				person := new(hybridsystem.Person)
				if err := json.NewDecoder(w.Body).Decode(person); err != nil {
					t.Fatalf("decode error: %v", err)
				}
				if person.Name != tt.person.Name {
					t.Fatalf("Expected name %s, got %s", tt.person.Name, person.Name)
				}
				if person.Email != tt.person.Email {
					t.Fatalf("Expected email %s, got %s", tt.person.Email, person.Email)
				}
				if person.ID.IsZero() {
					log.Panic("response id is empty")
				}
			} else {
				if w.Code == http.StatusOK {
					t.Fatalf("Expected  not ok status, got %d", w.Code)
				}
			}
		})
	}
}

func TestHybridHandler3_GetUserHandler4(t *testing.T) {
	tests := []struct {
		name     string // description of this test case
		id       string
		willpass bool
	}{
		{
			name:     "valid id exists in mongoDB",
			id:       primitive.NewObjectID().Hex(),
			willpass: true,
		},
		{
			name:     "invalid id format",
			id:       "5357",
			willpass: false,
		},
	}

	os.Setenv("MONGO_URI", "mongodb://localhost:27017")
	os.Setenv("MONGO_DB", "go_users")
	os.Setenv("REDIS_ADDR", "localhost:6379")

	redisInstance, err := hybridsystem.Connectredis1()
	if err != nil {
		log.Fatal(err)
	}
	mongoInstance, err := hybridsystem.ConnectMongo1()
	if err != nil {
		log.Fatal(err)
	}
	handle := &hybridsystem.HybridHandler3{Mongo: mongoInstance, Redis: redisInstance, Ctx: context.Background()}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.willpass {
				objID, _ := primitive.ObjectIDFromHex(tt.id)
				testPerson := hybridsystem.Person{
					ID:    objID,
					Name:  "Akash",
					Email: "akash@gmail.com",
				}
				_, err := handle.Mongo.Persons.InsertOne(handle.Ctx, testPerson)
				if err != nil {
					t.Fatalf("Failed to insert test persons: %v", err)
				}
			}
			// build request with mux wars
			r := httptest.NewRequest(http.MethodGet, "/persons/"+tt.id, nil)
			r = mux.SetURLVars(r, map[string]string{"id": tt.id})
			w := httptest.NewRecorder()

			handle.GetUserHandler4(w, r)

			if tt.willpass {
				if w.Code != http.StatusOK {
					t.Fatalf("Expected ok status, got %d", w.Code)
				}
				person := new(hybridsystem.Person)
				if err := json.NewDecoder(w.Body).Decode(&person); err != nil {
					t.Fatalf("decode error: %v", err)
				}
				if person.ID.Hex() != tt.id {
					t.Fatalf("Expected id %s, got %s", person.ID.Hex(), tt.id)
				}
			} else {
				if w.Code == http.StatusOK {
					t.Fatalf("Expected not ok status,got %d ", w.Code)
				}
			}
		})
	}
}
