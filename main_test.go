package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"

	// "strconv"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupRouter() *gin.Engine {
	db = DB{
		currentId: 1,
		users:     make([]User, 0),
		mu:        sync.Mutex{},
	}
	r := gin.Default()
	r.GET("/users", getUsers)
	r.GET("/users/:id", getUserById)
	r.POST("/users", createUser)
	r.PUT("/users/:id", updateUser)
	r.DELETE("/users/:id", deleteUser)
	return r
}

func TestGetUsers(t *testing.T) {
	r := setupRouter()
	db.mu.Lock()
	db.users = append(db.users, User{Id: 1, Name: "Ahmed Zaher", Email: "ahmed@zaher.com", Age: 25})
	db.mu.Unlock()

	req, _ := http.NewRequest("GET", "/users", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var users []User
	err := json.Unmarshal(w.Body.Bytes(), &users)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(users))
	assert.Equal(t, "Ahmed Zaher", users[0].Name)
}

func TestGetUserById(t *testing.T) {
	r := setupRouter()
	db.mu.Lock()
	db.users = append(db.users, User{Id: 1, Name: "Ahmed Zaher", Email: "ahmed@zaher.com", Age: 25})
	db.mu.Unlock()

	req, _ := http.NewRequest("GET", "/users/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var user User
	err := json.Unmarshal(w.Body.Bytes(), &user)
	assert.NoError(t, err)
	assert.Equal(t, "Ahmed Zaher", user.Name)
}

func TestCreateUser(t *testing.T) {
	r := setupRouter()

	newUser := User{Name: "Ahmed Zaher", Email: "ahmed@zaher", Age: 25}
	jsonValue, _ := json.Marshal(newUser)

	req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var createdUser User
	err := json.Unmarshal(w.Body.Bytes(), &createdUser)
	assert.NoError(t, err)
	assert.Equal(t, "Ahmed Zaher", createdUser.Name)
}

func TestUpdateUser(t *testing.T) {
	r := setupRouter()
	db.mu.Lock()
	db.users = append(db.users, User{Id: 1, Name: "Ahmed Zaher", Email: "ahmed@zaher.com", Age: 25})
	db.mu.Unlock()

	updatedUser := User{Name: "Ahmed Updated", Email: "ahmed.updated@zaher.com", Age: 35}
	jsonValue, _ := json.Marshal(updatedUser)

	req, _ := http.NewRequest("PUT", "/users/1", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var user User
	err := json.Unmarshal(w.Body.Bytes(), &user)
	assert.NoError(t, err)
	assert.Equal(t, "Ahmed Updated", user.Name)
}

func TestDeleteUser(t *testing.T) {
	r := setupRouter()
	db.mu.Lock()
	db.users = append(db.users, User{Id: 1, Name: "Ahmed Zaher", Email: "ahmed@zaher.com", Age: 25})
	db.mu.Unlock()

	req, _ := http.NewRequest("DELETE", "/users/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)

	db.mu.Lock()
	assert.Equal(t, 0, len(db.users))
	db.mu.Unlock()
}

func TestConcurrentRequests(t *testing.T) {
	r := setupRouter()

	var wg sync.WaitGroup
	const numRequests = 50

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			newUser := User{Name: "User " + strconv.Itoa(id), Email: "user" + strconv.Itoa(id) + "@example.com", Age: 25 + id}
			jsonValue, _ := json.Marshal(newUser)

			req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(jsonValue))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusCreated, w.Code)
		}(i)
	}

	wg.Wait()

	db.mu.Lock()
	assert.Equal(t, numRequests, len(db.users))
	db.mu.Unlock()
}
