package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func addAgent(c *gin.Context) {
	var input struct {
		Name  string `json:"name" binding:"required"`
		Phone string `json:"phone" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	agent := DeliveryAgent{
		Name:   input.Name,
		Phone:  input.Phone,
		Status: AgentStatusAvailable,
	}

	if err := DB.Create(&agent).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create agent"})
		return
	}

	c.JSON(http.StatusCreated, agent)
}

func getAvailableAgents(c *gin.Context) {
	var agents []DeliveryAgent
	if err := DB.Where("status = ?", AgentStatusAvailable).Find(&agents).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch agents"})
		return
	}
	
	if agents == nil {
		agents = []DeliveryAgent{}
	}

	c.JSON(http.StatusOK, agents)
}

func updateTaskStatus(c *gin.Context) {
	id := c.Param("id")
	var input struct {
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var task DeliveryTask
	if err := DB.First(&task, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	task.Status = input.Status
	DB.Save(&task)

	// If delivered, mark agent as available again
	if input.Status == TaskStatusDelivered {
		var agent DeliveryAgent
		if err := DB.First(&agent, task.DeliveryAgentID).Error; err == nil {
			agent.Status = AgentStatusAvailable
			DB.Save(&agent)
		}
	}

	c.JSON(http.StatusOK, task)
}
