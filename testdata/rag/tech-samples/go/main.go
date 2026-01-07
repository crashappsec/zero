// Test file: Go detection patterns
// Should detect: Go, Gin/Echo framework, GORM patterns

package main

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username string `json:"username" gorm:"unique"`
	Email    string `json:"email" gorm:"unique"`
}

var db *gorm.DB

func main() {
	// Database connection
	dsn := os.Getenv("DATABASE_URL")
	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	db.AutoMigrate(&User{})

	// Router setup
	r := gin.Default()

	r.GET("/api/users", getUsers)
	r.POST("/api/users", createUser)
	r.GET("/api/users/:id", getUser)
	r.PUT("/api/users/:id", updateUser)
	r.DELETE("/api/users/:id", deleteUser)

	r.Run(":8080")
}

func getUsers(c *gin.Context) {
	var users []User
	db.Find(&users)
	c.JSON(http.StatusOK, users)
}

func getUser(c *gin.Context) {
	var user User
	if err := db.First(&user, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	c.JSON(http.StatusOK, user)
}

func createUser(c *gin.Context) {
	var user User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db.Create(&user)
	c.JSON(http.StatusCreated, user)
}

func updateUser(c *gin.Context) {
	var user User
	if err := db.First(&user, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	c.ShouldBindJSON(&user)
	db.Save(&user)
	c.JSON(http.StatusOK, user)
}

func deleteUser(c *gin.Context) {
	var user User
	if err := db.First(&user, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	db.Delete(&user)
	c.JSON(http.StatusOK, gin.H{"message": "User deleted"})
}

// JWT token validation middleware
func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
		c.Next()
	}
}
