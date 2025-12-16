package leases

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type LeaseHandler struct {
	service *LeaseService
}

func NewLeaseHandler(service *LeaseService) *LeaseHandler {
	return &LeaseHandler{service: service}
}

// GET /buildings/:id/leases/customers
func (h *LeaseHandler) GetCustomers(c *gin.Context) {
	buildingIDStr := c.Param("id")
	buildingID, err := strconv.Atoi(buildingIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid building ID"})
		return
	}

	customers, err := h.service.GetCustomersByBuildingID(buildingID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, customers)
}

// GET /buildings/:id/leases/customers-with-units
func (h *LeaseHandler) GetCustomersWithLeaseUnits(c *gin.Context) {
	buildingIDStr := c.Param("id")
	buildingID, err := strconv.Atoi(buildingIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid building ID"})
		return
	}

	customers, err := h.service.GetCustomersWithLeaseUnits(buildingID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, customers)
}

// GET /buildings/:id/leases/available-units
func (h *LeaseHandler) GetAvailableUnits(c *gin.Context) {
	buildingIDStr := c.Param("id")
	buildingID, err := strconv.Atoi(buildingIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid building ID"})
		return
	}

	// Get optional include_unit_id query parameter (for edit mode)
	includeUnitIDStr := c.Query("include_unit_id")
	var includeUnitID *int
	if includeUnitIDStr != "" {
		unitID, err := strconv.Atoi(includeUnitIDStr)
		if err == nil {
			includeUnitID = &unitID
		}
	}

	units, err := h.service.GetAvailableUnits(buildingID, includeUnitID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, units)
}

// GET /buildings/:id/leases/units-by-people/:peopleId
func (h *LeaseHandler) GetUnitsByPeopleID(c *gin.Context) {
	buildingIDStr := c.Param("id")
	buildingID, err := strconv.Atoi(buildingIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid building ID"})
		return
	}

	peopleIDStr := c.Param("peopleId")
	peopleID, err := strconv.Atoi(peopleIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid people ID"})
		return
	}

	units, err := h.service.GetUnitsByPeopleID(buildingID, peopleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, units)
}

// GET /buildings/:id/leases/unit/:unitId
func (h *LeaseHandler) GetLeasesByUnitID(c *gin.Context) {
	unitIDStr := c.Param("unitId")
	unitID, err := strconv.Atoi(unitIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid unit ID"})
		return
	}

	leases, err := h.service.GetLeasesByUnitID(unitID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, leases)
}

// POST /buildings/:id/leases
func (h *LeaseHandler) CreateLease(c *gin.Context) {
	var req CreateLeaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	buildingIDStr := c.Param("id")
	buildingID, err := strconv.Atoi(buildingIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid building ID"})
		return
	}
	req.BuildingID = buildingID

	if req.Status == "" {
		req.Status = "1"
	}

	response, err := h.service.CreateLease(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// PUT /buildings/:id/leases/:leaseId
func (h *LeaseHandler) UpdateLease(c *gin.Context) {
	var req UpdateLeaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	leaseIDStr := c.Param("leaseId")
	leaseID, err := strconv.Atoi(leaseIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid lease ID"})
		return
	}
	req.ID = leaseID

	buildingIDStr := c.Param("id")
	buildingID, err := strconv.Atoi(buildingIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid building ID"})
		return
	}
	req.BuildingID = buildingID

	response, err := h.service.UpdateLease(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GET /buildings/:id/leases/:leaseId
func (h *LeaseHandler) GetLeaseByID(c *gin.Context) {
	leaseIDStr := c.Param("leaseId")
	leaseID, err := strconv.Atoi(leaseIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid lease ID"})
		return
	}

	response, err := h.service.GetLeaseByID(leaseID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GET /buildings/:id/leases
func (h *LeaseHandler) GetLeasesByBuildingID(c *gin.Context) {
	buildingIDStr := c.Param("id")
	buildingID, err := strconv.Atoi(buildingIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid building ID"})
		return
	}

	leases, err := h.service.GetLeasesByBuildingID(buildingID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, leases)
}

// DELETE /buildings/:id/leases/:leaseId
func (h *LeaseHandler) DeleteLease(c *gin.Context) {
	leaseIDStr := c.Param("leaseId")
	leaseID, err := strconv.Atoi(leaseIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid lease ID"})
		return
	}

	err = h.service.DeleteLease(leaseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Lease deleted successfully"})
}

// POST /buildings/:id/leases/:leaseId/files
func (h *LeaseHandler) UploadLeaseFile(c *gin.Context) {
	leaseIDStr := c.Param("leaseId")
	leaseID, err := strconv.Atoi(leaseIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid lease ID"})
		return
	}

	buildingIDStr := c.Param("id")
	buildingID, err := strconv.Atoi(buildingIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid building ID"})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file provided"})
		return
	}

	// Create upload directory
	uploadPath := GetUploadPath(buildingID)
	if err := os.MkdirAll(uploadPath, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload directory"})
		return
	}

	// Generate unique filename
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%d_%s%s", leaseID, generateUniqueID(), ext)
	filePath := filepath.Join(uploadPath, filename)

	// Save file
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file"})
		return
	}
	defer src.Close()

	dst, err := os.Create(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create file"})
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	// Get file type
	fileType := strings.ToLower(ext[1:]) // Remove the dot
	if fileType == "" {
		fileType = "unknown"
	}

	// Save file record
	leaseFile, err := h.service.UploadLeaseFile(leaseID, filename, file.Filename, filePath, fileType, file.Size)
	if err != nil {
		// If database insert fails, delete the uploaded file
		if _, statErr := os.Stat(filePath); statErr == nil {
			os.Remove(filePath)
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, leaseFile)
}

// GET /buildings/:id/leases/:leaseId/files/:fileId/download
func (h *LeaseHandler) DownloadLeaseFile(c *gin.Context) {
	fileIDStr := c.Param("fileId")
	fileID, err := strconv.Atoi(fileIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file ID"})
		return
	}

	leaseFile, err := h.service.GetLeaseFileByID(fileID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	// Check if file exists
	if _, err := os.Stat(leaseFile.FilePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found on disk"})
		return
	}

	c.FileAttachment(leaseFile.FilePath, leaseFile.OriginalName)
}

// DELETE /buildings/:id/leases/:leaseId/files/:fileId
func (h *LeaseHandler) DeleteLeaseFile(c *gin.Context) {
	fileIDStr := c.Param("fileId")
	fileID, err := strconv.Atoi(fileIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file ID"})
		return
	}

	err = h.service.DeleteLeaseFile(fileID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "File deleted successfully"})
}

// Helper function to generate unique ID
func generateUniqueID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
