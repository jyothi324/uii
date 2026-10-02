package handlers

import (

	"github.com/gin-gonic/gin"
	"inventory/services"
)


func GeminiImageHandler(c *gin.Context) {
	file, _ := c.FormFile("image")
	texture, _ := c.FormFile("texture_image")

	target := c.PostForm("target")
	notes := c.PostForm("design_notes")

	result, err := services.ProcessGeminiImage(file, texture, target, notes)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"image": result,
	})
}