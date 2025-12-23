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
