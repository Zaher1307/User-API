package main

import (
	"net/http"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
)

type User struct {
	Id    int    `json:"id"`
	Age   int    `json:"age" binding:"required"`
	Email string `json:"email" binding:"required"`
	Name  string `json:"name"  binding:"required"`
}

type DB struct {
	currentId int
	users     []User
	mu        sync.Mutex
}

var db DB

func main() {
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

	r.Run(":8080")
}

func getUsers(c *gin.Context) {
	// Why not db.mu.Lock() then defer db.mu.Unlock()
	// copying the users slice into different slice variable to reduce the size
	// of the critical section instead of blocknig until I/O network call to be
	// done.
	db.mu.Lock()
	users := db.users
	db.mu.Unlock()

	c.JSON(http.StatusOK, users)
}

func getUserById(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "id is not a number",
		})
		return
	}

	var user User
	found := false

	db.mu.Lock()
	for _, u := range db.users {
		if u.Id == id {
			user = u
			found = true
			break
		}
	}
	db.mu.Unlock()

	if !found {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "user not found",
		})
		return
	}

	c.JSON(http.StatusOK, user)
}

func createUser(c *gin.Context) {
	var createdUser User
	if err := c.ShouldBindJSON(&createdUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	db.mu.Lock()
	createdUser.Id = db.currentId
	db.currentId++
	db.users = append(db.users, createdUser)
	db.mu.Unlock()

	c.JSON(http.StatusCreated, createdUser)
}

func updateUser(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "id is not a number",
		})
		return
	}

	var updatedUser User
	if err := c.ShouldBindJSON(&updatedUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	updatedUser.Id = id

	found := false

	db.mu.Lock()
	for i, u := range db.users {
		if u.Id == id {
			db.users[i] = updatedUser
			found = true
			break
		}
	}
	db.mu.Unlock()

	if !found {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "user not found",
		})
		return
	}

	c.JSON(http.StatusOK, updatedUser)
}

func deleteUser(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "id is not a number",
		})
		return
	}

	found := false

	db.mu.Lock()
	for i, u := range db.users {
		if u.Id == id {
			db.users = append(db.users[:i], db.users[i+1:]...)
			found = true
			break
		}
	}
	db.mu.Unlock()

	if !found {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "user not found",
		})
		return
	}

	c.JSON(http.StatusNoContent, gin.H{
		"message": "user deleted successfully",
	})
}
