package redisDatabase_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	redisDatabase "redisDatabase/redisDatabase"
	"testing"

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
