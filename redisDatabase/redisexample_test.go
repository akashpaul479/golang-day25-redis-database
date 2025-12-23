package redisDatabase_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	redisDatabase "redisDatabase/redisDatabase"
	"strconv"
	"testing"

	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
)

func TestApp_createUserHandler(t *testing.T) {
	tests := []struct {
		name     string // description of this test case
		user     redisDatabase.User
		willpass bool
	}{
		{
			name: "valid name and valid email",
			user: redisDatabase.User{
				Name:  "akash",
				Email: "akash@gmail.com",
			},
			willpass: true,
		},
		{
			name: "invalid name and valid email",
			user: redisDatabase.User{
				Name:  "",
				Email: "akash@gmail.com",
			},
			willpass: false,
		},
		{
			name: "valid name and invalid email",
			user: redisDatabase.User{
				Name:  "akash",
				Email: "",
			},
			willpass: false,
		},
		{
			name: "valid name and gmail without prefix",
			user: redisDatabase.User{
				Name:  "akash",
				Email: "@gmail.com",
			},
			willpass: false,
		},
		{
			name: "withspace only name and valid email",
			user: redisDatabase.User{
				Name:  "   ",
				Email: "akash@gmail.com",
			},
			willpass: false,
		},
	}
	dsn := "root:root@tcp(127.0.0.1:3306)/go_users"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   0,
	})
	app := &redisDatabase.App{
		DB:  db,
		RDB: rdb,
		Ctx: context.Background(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// prepare inputs
			usersBytes, err := json.Marshal(tt.user)
			if err != nil {
				log.Fatalf("fail to marshal!")
			}
			buffer := bytes.NewBuffer(usersBytes)
			r := httptest.NewRequest(http.MethodPost, "localhost:8080/users", buffer)
			w := httptest.NewRecorder()

			app.CreateUserHandler(w, r)

			if tt.willpass {
				if w.Code != http.StatusOK {
					t.Errorf("Expected ok status , got %d", w.Code)
				}
			} else {
				if w.Code == http.StatusOK {
					t.Errorf("Expected not ok status, got %d", w.Code)
				}
			}

			// validate response

			var got redisDatabase.User
			if err := json.NewDecoder(w.Body).Decode(&got); err == nil {
				if got.Email != tt.user.Email {
					t.Errorf("%s: expected email %s, got %s", tt.name, tt.user.Email, got.Email)
				}
				if got.Name != tt.user.Name {
					t.Errorf("%s: expected name %s, got %s", tt.name, tt.user.Name, got.Name)
				}
			}

		})
	}
}

func TestApp_GetUserHandler(t *testing.T) {
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
			id:       9648,
			willpass: false,
		},
	}
	dsn := "root:root@tcp(127.0.0.1:3306)/go_users"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   0,
	})
	app := &redisDatabase.App{
		DB:  db,
		RDB: rdb,
		Ctx: context.Background(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// prepare input
			r := httptest.NewRequest(http.MethodGet, "/users"+strconv.Itoa(tt.id), nil)
			r = mux.SetURLVars(r, map[string]string{"id": strconv.Itoa(tt.id)})
			w := httptest.NewRecorder()

			app.GetUserHandler(w, r)

			if tt.willpass {
				if w.Code != http.StatusOK {
					t.Fatalf("Expected status ok, got %d", w.Code)
				}
				var user redisDatabase.User
				if err := json.NewDecoder(w.Body).Decode(&user); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				if user.ID != tt.id {
					t.Errorf("Expected id %d, got %d", tt.id, user.ID)
				}

			} else {
				if w.Code == http.StatusOK {
					t.Fatalf("Expected not found , got %d", w.Code)
				}
			}
		})
	}
}

func TestApp_UpdateUserHandler(t *testing.T) {
	dsn := "root:root@tcp(127.0.0.1:3306)/go_users"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   0,
	})
	app := &redisDatabase.App{
		DB:  db,
		RDB: rdb,
		Ctx: context.Background(),
	}
	_, err = db.Exec("INSERT INTO users(id , name , email) VALUES(1, 'Akash','akash@gmail.com') ON DUPLICATE KEY UPDATE name='Akash',email='akash@gmail.com'")
	if err != nil {
		t.Fatalf("failed to seed mysql: %v", err)
	}
	tests := []struct {
		name     string // description of this test case
		user     redisDatabase.User
		willpass bool
	}{
		{
			name: "valid name and valid email",
			user: redisDatabase.User{
				ID:    1,
				Name:  "Akash paul",
				Email: "akashpaul@gmail.com",
			},
			willpass: true,
		},
		{
			name: " invalid id format",
			user: redisDatabase.User{
				ID:    6548,
				Name:  "ghost",
				Email: "ghost@gmail.com",
			},
			willpass: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyBytes, _ := json.Marshal(tt.user)
			req := httptest.NewRequest(http.MethodPut, "/users/"+fmt.Sprint(tt.user.ID), bytes.NewBuffer(bodyBytes))
			req = mux.SetURLVars(req, map[string]string{"id": fmt.Sprint(tt.user.ID)})
			w := httptest.NewRecorder()

			app.UpdateUserHandler(w, req)

			if tt.willpass {
				if w.Code != http.StatusOK {
					t.Fatalf("Expected status OK, got %d", w.Code)
				}
				var updated redisDatabase.User
				if err := json.NewDecoder(w.Body).Decode(&updated); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				if updated.Name != tt.user.Name {
					t.Errorf("Expected name %s, got %s", tt.user.Name, updated.Name)
				}
				if updated.Email != tt.user.Email {
					t.Errorf("Expected email %s, got %s", tt.user.Email, updated.Email)
				}
			} else {
				if w.Code == http.StatusOK {
					t.Fatalf("Expected error status, got %d", w.Code)
				}
			}

		})
	}
}

func TestApp_DeleteUserHandler(t *testing.T) {
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
			id:       9648,
			willpass: false,
		},
	}
	dsn := "root:root@tcp(127.0.0.1:3306)/go_users"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   0,
	})
	app := &redisDatabase.App{
		DB:  db,
		RDB: rdb,
		Ctx: context.Background(),
	}
	_, err = db.Exec("INSERT INTO users(id, name, email) VALUES(1, 'Akash', 'akash@gmail.com') " +
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
			app.DeleteUserHandler(w, r)

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
