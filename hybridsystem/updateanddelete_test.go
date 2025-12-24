package hybridsystem_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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

func TestHybridHandler3_UpdateUserHandler3(t *testing.T) {
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

	_, err = handle.MySQL.DB.Exec("INSERT INTO users (id ,name , email) VALUES (1,'Akash' ,'akash@gmail.com') ON DUPLICATE KEY UPDATE name='Akash',email='akash@gmail.com'")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name     string // description of this test case
		id       string
		user     hybridsystem.User2
		willpass bool
	}{
		{
			name: "valid update",
			user: hybridsystem.User2{
				ID:    1,
				Name:  "Akash paul",
				Email: "akashpaul@gmail.com",
			},
			willpass: true,
		},
		{
			name: "valid name and invalid email",
			user: hybridsystem.User2{
				ID:    1,
				Name:  "Akash paul",
				Email: "",
			},
			willpass: false,
		},
		{
			name: " invalid id format",
			id:   "5347",
			user: hybridsystem.User2{

				Name:  "akash",
				Email: "akash@gmail.com",
			},
			willpass: false,
		},
		{
			name: " invalid name",
			user: hybridsystem.User2{
				ID:    1,
				Name:  "",
				Email: "akash@gmail.com",
			},
			willpass: false,
		},
		{
			name: "valid name and invalid   email",
			user: hybridsystem.User2{
				ID:    1,
				Name:  "akash",
				Email: "",
			},
			willpass: false,
		},
		{
			name: "withspace-only name and valid  email",
			user: hybridsystem.User2{
				ID:    1,
				Name:  "   ",
				Email: "abc@gmail.com",
			},
			willpass: false,
		},
		{
			name: "nonexistent user",
			id:   "5336",
			user: hybridsystem.User2{

				Name:  "Ghost",
				Email: "ghost@gmail.com",
			},
			willpass: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userBytes, err := json.Marshal(tt.user)
			if err != nil {
				log.Panic("failed to marshal")
			}
			r := httptest.NewRequest(http.MethodPut, "/users/"+fmt.Sprint(tt.user.ID), bytes.NewBuffer(userBytes))
			r = mux.SetURLVars(r, map[string]string{"id": fmt.Sprint(tt.user.ID)})
			w := httptest.NewRecorder()

			handle.UpdateUserHandler3(w, r)

			if tt.willpass {
				if w.Code != http.StatusOK {
					t.Fatalf("Expected status ok , got %d", w.Code)
				}
				var updated hybridsystem.User2
				if err := json.NewDecoder(w.Body).Decode(&updated); err != nil {
					t.Fatalf("failed to decode response : %v", err)
				}
				if updated.Name != tt.user.Name {
					t.Fatalf("Expected name %s, got %s", tt.user.Name, updated.Name)
				}
				if updated.Email != tt.user.Email {
					t.Fatalf("Expected email %s, got %s", tt.user.Email, updated.Email)
				}

			} else {
				if w.Code == http.StatusOK {
					t.Fatalf("Expected status not ok , got %d", w.Code)
				}
			}
		})
	}
}

func TestHybridHandler3_DeleteUserHandler3(t *testing.T) {
	tests := []struct {
		name     string // description of this test case
		id       int
		willpass bool
	}{
		{
			name:     "valid id exists in sql",
			id:       1,
			willpass: true,
		},
		{
			name:     "invalid id not found ",
			id:       6345,
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
	_, err = handle.MySQL.DB.Exec("INSERT INTO users(id, name, email) VALUES(1, 'Akash', 'akash@gmail.com') " +
		"ON DUPLICATE KEY UPDATE name='Akash', email='akash@gmail.com'")
	if err != nil {
		t.Fatalf("failed to seed mysql: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// prepare input
			r := httptest.NewRequest(http.MethodDelete, "/users/"+strconv.Itoa(tt.id), nil)
			r = mux.SetURLVars(r, map[string]string{"id": strconv.Itoa(tt.id)})
			w := httptest.NewRecorder()

			handle.DeleteUserHandler3(w, r)

			if tt.willpass {
				if w.Code != http.StatusOK {
					t.Fatalf("expected status ok , got %d", w.Code)
				}
				if w.Body.String() != "user deleted" {
					t.Errorf("Expected 'user deleted', got %s", w.Body.String())
				}
			} else {
				if w.Code == http.StatusOK {
					t.Fatalf("Expected not ok status, got %d", w.Code)
				}
				return

			}
		})
	}
}

func TestHybridHandler3_UpdateUserHandler4(t *testing.T) {

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
	handler := &hybridsystem.HybridHandler3{Mongo: mongoInstance, Redis: redisInstance, Ctx: context.Background()}

	res, err := handler.Mongo.Persons.InsertOne(context.Background(), hybridsystem.Person{
		Name:  "abc",
		Email: "akashpaul@gmail.com",
	})
	if err != nil {
		t.Error("failed to create user")
	}
	id := res.InsertedID.(primitive.ObjectID)
	validID := id.Hex()

	tests := []struct {
		name     string // description of this test case
		id       string
		person   hybridsystem.Person
		willpass bool
	}{
		{
			name: "valid name and valid email",
			id:   validID,
			person: hybridsystem.Person{
				ID:    id,
				Name:  "Akash",
				Email: "akash@gmail.com",
			},
			willpass: true,
		},
		{
			name: "invalid name and valid email",
			id:   validID,
			person: hybridsystem.Person{
				ID:    id,
				Name:  "",
				Email: "akash@gmail.com",
			},
			willpass: false,
		},
		{
			name: "valid name and invalid email",
			id:   validID,
			person: hybridsystem.Person{
				ID:    id,
				Name:  "valid",
				Email: "",
			},
			willpass: false,
		},
		{
			name: "invalid id format",
			id:   "5347",
			person: hybridsystem.Person{
				Name:  "Akash",
				Email: "akash@gmail.com",
			},
			willpass: false,
		},
		{
			name: "nonexistent user",
			id:   primitive.NewObjectID().Hex(),
			person: hybridsystem.Person{

				Name:  "Ghost",
				Email: "ghost@gmail.com",
			},
			willpass: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userBytes, err := json.Marshal(tt.person)
			if err != nil {
				log.Panic("failed tp marshal")
			}
			buffer := bytes.NewBuffer(userBytes)
			r := httptest.NewRequest(http.MethodPut, "/persons/"+tt.id, buffer)
			r = mux.SetURLVars(r, map[string]string{"id": tt.id})
			w := httptest.NewRecorder()

			handler.UpdateUserHandler4(w, r)

			if tt.willpass {
				if w.Code != http.StatusOK {
					t.Fatalf("Expected ok status , got %d", w.Code)
				}
				var person hybridsystem.Person
				if err := json.NewDecoder(w.Body).Decode(&person); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				if person.Email != tt.person.Email {
					t.Fatalf("expected email %s, got %s", tt.person.Email, person.Email)
				}
				if person.Name != tt.person.Name {
					t.Fatalf("expected name %s, got %s", tt.person.Name, person.Name)
				}
				if person.ID.IsZero() {
					t.Fatalf("response id is empty")
				}
			} else {
				if w.Code == http.StatusOK {
					t.Fatalf(" Expected not ok status  , got %d", w.Code)

				}
			}
		})
	}
}

func TestHybridHandler3_DeleteuserHandler4(t *testing.T) {
	tests := []struct {
		name     string // description of this test case
		id       string
		willpass bool
	}{
		{
			name:     "valid id exists in mongodb",
			id:       primitive.NewObjectID().Hex(),
			willpass: true,
		},
		{
			name:     "invalid id not found ",
			id:       "6345",
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
	handler := &hybridsystem.HybridHandler3{Mongo: mongoInstance, Redis: redisInstance, Ctx: context.Background()}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.willpass {
				objID, _ := primitive.ObjectIDFromHex(tt.id)
				testUser := hybridsystem.Person{
					ID:    objID,
					Name:  "Akash",
					Email: "akash@gmail.com",
				}
				_, err := handler.Mongo.Persons.InsertOne(handler.Ctx, testUser)
				if err != nil {
					t.Fatalf("failed to insert test user: %v", err)
				}
			}
			// Build request with mux vars
			r := httptest.NewRequest(http.MethodPut, "/persons/"+tt.id, nil)
			r = mux.SetURLVars(r, map[string]string{"id": tt.id})
			w := httptest.NewRecorder()

			handler.DeleteuserHandler4(w, r)

			if tt.willpass {
				if w.Code != http.StatusOK {
					t.Fatalf("expected status ok , got %d", w.Code)
				}
				if w.Body.String() != "user Deleted!" {
					t.Errorf("Expected 'user Deleted!', got %s", w.Body.String())
				}
			} else {
				if w.Code == http.StatusOK {
					t.Fatalf("Expected not ok status, got %d", w.Code)
				}
				return

			}
		})
	}
}
