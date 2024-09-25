package user_api

import (
	user_query "recorder/pkg/mariadb/user"

	"github.com/gin-gonic/gin"
)

type checkRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
}

type addRequest struct {
	Username   string `json:"username"`
	Email      string `json:"email"`
	Permission int    `json:"permission"`
}

type editRequest struct {
	Username   string `json:"username"`
	Permission int    `json:"permission"`
}

type deleteRequest struct {
	Username string `json:"username"`
}

func Check(c *gin.Context) {
	var req checkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	permission, err := user_query.CheckUser(req.Username)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	if permission == -1 {
		user_query.AddUser(req.Username, req.Email, 0) // default permission is 0 (guest)
		permission = 0
	}
	c.JSON(200, gin.H{"permission": permission})
}

func Add(c *gin.Context) {
	var req addRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	err := user_query.AddUser(req.Username, req.Email, req.Permission)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "success"})
}

func List(c *gin.Context) {
	users, err := user_query.ListUser()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"users": users})
}

func Edit(c *gin.Context) {
	var req editRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	err := user_query.EditUserPermission(req.Username, req.Permission)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "success"})
}

func Delete(c *gin.Context) {
	var req deleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	err := user_query.DeleteUser(req.Username)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "success"})
}
