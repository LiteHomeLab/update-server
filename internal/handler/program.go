package handler

import (
	"docufiller-update-server/internal/models"
	"docufiller-update-server/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ProgramHandler struct {
	programSvc *service.ProgramService
}

func NewProgramHandler(programSvc *service.ProgramService) *ProgramHandler {
	return &ProgramHandler{programSvc: programSvc}
}

// CreateProgram 创建程序
func (h *ProgramHandler) CreateProgram(c *gin.Context) {
	var req struct {
		ProgramID   string `json:"programId" binding:"required"`
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		IconURL     string `json:"iconUrl"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	program := &models.Program{
		ProgramID:   req.ProgramID,
		Name:        req.Name,
		Description: req.Description,
		IconURL:     req.IconURL,
		IsActive:    true,
	}

	if err := h.programSvc.CreateProgram(program); err != nil {
		c.JSON(500, gin.H{"error": "failed to create program"})
		return
	}

	c.JSON(http.StatusOK, program)
}

// ListPrograms 列出所有程序
func (h *ProgramHandler) ListPrograms(c *gin.Context) {
	programs, err := h.programSvc.ListPrograms()
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to list programs"})
		return
	}

	c.JSON(http.StatusOK, programs)
}

// GetProgram 获取程序详情
func (h *ProgramHandler) GetProgram(c *gin.Context) {
	programID := c.Param("programId")

	program, err := h.programSvc.GetProgramByID(programID)
	if err != nil {
		c.JSON(404, gin.H{"error": "program not found"})
		return
	}

	c.JSON(http.StatusOK, program)
}
