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
