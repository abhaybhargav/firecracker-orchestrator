package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/abhaybhargav/firecracker-orchestrator/internal/database"
	"github.com/abhaybhargav/firecracker-orchestrator/pkg/firecracker"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// Server represents the API server
type Server struct {
	vmManager *firecracker.Manager
	db        *database.Database
	logger    *logrus.Logger
}

// NewServer creates a new API server
func NewServer(vmManager *firecracker.Manager, db *database.Database, logger *logrus.Logger) *Server {
	return &Server{
		vmManager: vmManager,
		db:        db,
		logger:    logger,
	}
}

// SetupRoutes configures the API routes
func (s *Server) SetupRoutes(r *gin.Engine) {
	// Serve static files
	r.Static("/static", "./web/static")

	// Load HTML templates
	r.LoadHTMLGlob("web/templates/*")

	// Web UI routes
	r.GET("/", s.handleDashboard)
	r.GET("/vms", s.handleVMsPage)
	r.GET("/vms/new", s.handleNewVMPage)
	r.GET("/vms/:id", s.handleVMDetailPage)
	r.GET("/containers", s.handleContainersPage)
	r.GET("/containers/new", s.handleNewContainerPage)

	// API routes
	api := r.Group("/api/v1")
	{
		// Status and health
		api.GET("/status", s.handleStatus)
		api.GET("/health", s.handleHealth)
		api.GET("/stats", s.handleStats)

		// VM management
		api.GET("/vms", s.handleListVMs)
		api.POST("/vms", s.handleCreateVM)
		api.GET("/vms/:id", s.handleGetVM)
		api.PUT("/vms/:id", s.handleUpdateVM)
		api.DELETE("/vms/:id", s.handleDeleteVM)
		api.POST("/vms/:id/start", s.handleStartVM)
		api.POST("/vms/:id/stop", s.handleStopVM)

		// Container management
		api.GET("/containers", s.handleListContainers)
		api.POST("/containers", s.handleCreateContainer)
		api.GET("/containers/:id", s.handleGetContainer)
		api.PUT("/containers/:id", s.handleUpdateContainer)
		api.DELETE("/containers/:id", s.handleDeleteContainer)
		api.POST("/containers/:id/start", s.handleStartContainer)
		api.POST("/containers/:id/stop", s.handleStopContainer)
	}
}

// Web UI Handlers

func (s *Server) handleDashboard(c *gin.Context) {
	c.HTML(http.StatusOK, "layout.html", gin.H{
		"Title": "Dashboard",
		"Page":  "dashboard",
	})
}

func (s *Server) handleVMsPage(c *gin.Context) {
	c.HTML(http.StatusOK, "layout.html", gin.H{
		"Title": "Virtual Machines",
		"Page":  "vms",
	})
}

func (s *Server) handleNewVMPage(c *gin.Context) {
	c.HTML(http.StatusOK, "layout.html", gin.H{
		"Title": "Create VM",
		"Page":  "vms",
	})
}

func (s *Server) handleVMDetailPage(c *gin.Context) {
	vmID := c.Param("id")
	c.HTML(http.StatusOK, "layout.html", gin.H{
		"Title": "VM Details",
		"Page":  "vms",
		"VMID":  vmID,
	})
}

func (s *Server) handleContainersPage(c *gin.Context) {
	c.HTML(http.StatusOK, "layout.html", gin.H{
		"Title": "Containers",
		"Page":  "containers",
	})
}

func (s *Server) handleNewContainerPage(c *gin.Context) {
	c.HTML(http.StatusOK, "layout.html", gin.H{
		"Title": "Deploy Container",
		"Page":  "containers",
	})
}

// API Handlers

func (s *Server) handleStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"healthy":   true,
		"timestamp": time.Now().UTC(),
		"version":   "1.0.0",
	})
}

func (s *Server) handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "healthy"})
}

func (s *Server) handleStats(c *gin.Context) {
	vms, err := s.db.ListVMs()
	if err != nil {
		s.logger.Errorf("Failed to list VMs for stats: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load stats"})
		return
	}

	containers, err := s.db.ListContainers()
	if err != nil {
		s.logger.Errorf("Failed to list containers for stats: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load stats"})
		return
	}

	// Calculate statistics
	var runningVMs, runningContainers int
	for _, vm := range vms {
		if vm.Status == "running" {
			runningVMs++
		}
	}
	for _, container := range containers {
		if container.Status == "running" {
			runningContainers++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"totalVMs":          len(vms),
		"runningVMs":        runningVMs,
		"totalContainers":   len(containers),
		"runningContainers": runningContainers,
	})
}

// VM API Handlers

type CreateVMRequest struct {
	Name     string `json:"name" binding:"required"`
	Memory   int64  `json:"memory"`
	CPUs     int    `json:"cpus"`
	DiskSize int64  `json:"disk_size"`
}

func (s *Server) handleListVMs(c *gin.Context) {
	limit := c.Query("limit")

	vms, err := s.db.ListVMs()
	if err != nil {
		s.logger.Errorf("Failed to list VMs: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list VMs"})
		return
	}

	// Apply limit if specified
	if limit != "" {
		if limitInt, err := strconv.Atoi(limit); err == nil && limitInt > 0 && limitInt < len(vms) {
			vms = vms[:limitInt]
		}
	}

	c.JSON(http.StatusOK, vms)
}

func (s *Server) handleCreateVM(c *gin.Context) {
	var req CreateVMRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set defaults if not specified
	if req.Memory == 0 {
		req.Memory = 512
	}
	if req.CPUs == 0 {
		req.CPUs = 1
	}
	if req.DiskSize == 0 {
		req.DiskSize = 2
	}

	vm := &database.VM{
		ID:       uuid.New().String(),
		Name:     req.Name,
		Status:   "creating",
		Memory:   req.Memory,
		CPUs:     req.CPUs,
		DiskSize: req.DiskSize,
	}

	// Save to database first
	if err := s.db.CreateVM(vm); err != nil {
		s.logger.Errorf("Failed to create VM in database: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create VM"})
		return
	}

	// Create the VM with Firecracker
	if err := s.vmManager.CreateVM(vm); err != nil {
		s.logger.Errorf("Failed to create VM with Firecracker: %v", err)
		// Update status to error
		vm.Status = "error"
		s.db.UpdateVM(vm)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create VM"})
		return
	}

	s.logger.Infof("VM %s created successfully", vm.ID)
	c.JSON(http.StatusCreated, vm)
}

func (s *Server) handleGetVM(c *gin.Context) {
	vmID := c.Param("id")

	vm, err := s.db.GetVM(vmID)
	if err != nil {
		s.logger.Errorf("Failed to get VM %s: %v", vmID, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "VM not found"})
		return
	}

	c.JSON(http.StatusOK, vm)
}

func (s *Server) handleUpdateVM(c *gin.Context) {
	vmID := c.Param("id")

	vm, err := s.db.GetVM(vmID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "VM not found"})
		return
	}

	var req CreateVMRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	vm.Name = req.Name
	if req.Memory > 0 {
		vm.Memory = req.Memory
	}
	if req.CPUs > 0 {
		vm.CPUs = req.CPUs
	}
	if req.DiskSize > 0 {
		vm.DiskSize = req.DiskSize
	}

	if err := s.db.UpdateVM(vm); err != nil {
		s.logger.Errorf("Failed to update VM: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update VM"})
		return
	}

	c.JSON(http.StatusOK, vm)
}

func (s *Server) handleDeleteVM(c *gin.Context) {
	vmID := c.Param("id")

	if err := s.vmManager.DeleteVM(vmID); err != nil {
		s.logger.Errorf("Failed to delete VM %s: %v", vmID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete VM"})
		return
	}

	s.logger.Infof("VM %s deleted successfully", vmID)
	c.JSON(http.StatusOK, gin.H{"message": "VM deleted successfully"})
}

func (s *Server) handleStartVM(c *gin.Context) {
	vmID := c.Param("id")

	if err := s.vmManager.StartVM(vmID); err != nil {
		s.logger.Errorf("Failed to start VM %s: %v", vmID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start VM"})
		return
	}

	s.logger.Infof("VM %s started successfully", vmID)
	c.JSON(http.StatusOK, gin.H{"message": "VM started successfully"})
}

func (s *Server) handleStopVM(c *gin.Context) {
	vmID := c.Param("id")

	if err := s.vmManager.StopVM(vmID); err != nil {
		s.logger.Errorf("Failed to stop VM %s: %v", vmID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to stop VM"})
		return
	}

	s.logger.Infof("VM %s stopped successfully", vmID)
	c.JSON(http.StatusOK, gin.H{"message": "VM stopped successfully"})
}

// Container API Handlers

type CreateContainerRequest struct {
	Name        string            `json:"name" binding:"required"`
	Image       string            `json:"image" binding:"required"`
	VMID        string            `json:"vm_id" binding:"required"`
	Ports       map[string]string `json:"ports"`
	Environment map[string]string `json:"environment"`
}

func (s *Server) handleListContainers(c *gin.Context) {
	containers, err := s.db.ListContainers()
	if err != nil {
		s.logger.Errorf("Failed to list containers: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list containers"})
		return
	}

	c.JSON(http.StatusOK, containers)
}

func (s *Server) handleCreateContainer(c *gin.Context) {
	var req CreateContainerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify VM exists
	vm, err := s.db.GetVM(req.VMID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "VM not found"})
		return
	}

	if vm.Status != "running" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "VM must be running to deploy containers"})
		return
	}

	container := &database.Container{
		ID:     uuid.New().String(),
		Name:   req.Name,
		Image:  req.Image,
		Status: "creating",
		VMID:   req.VMID,
	}

	// Save to database
	if err := s.db.CreateContainer(container); err != nil {
		s.logger.Errorf("Failed to create container in database: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create container"})
		return
	}

	// TODO: Implement actual container creation in VM
	container.Status = "created"
	s.db.UpdateContainer(container)

	s.logger.Infof("Container %s created successfully", container.ID)
	c.JSON(http.StatusCreated, container)
}

func (s *Server) handleGetContainer(c *gin.Context) {
	containerID := c.Param("id")

	container, err := s.db.GetContainer(containerID)
	if err != nil {
		s.logger.Errorf("Failed to get container %s: %v", containerID, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Container not found"})
		return
	}

	c.JSON(http.StatusOK, container)
}

func (s *Server) handleUpdateContainer(c *gin.Context) {
	// TODO: Implement container update
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Container update not implemented yet"})
}

func (s *Server) handleDeleteContainer(c *gin.Context) {
	containerID := c.Param("id")

	if err := s.db.DeleteContainer(containerID); err != nil {
		s.logger.Errorf("Failed to delete container %s: %v", containerID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete container"})
		return
	}

	s.logger.Infof("Container %s deleted successfully", containerID)
	c.JSON(http.StatusOK, gin.H{"message": "Container deleted successfully"})
}

func (s *Server) handleStartContainer(c *gin.Context) {
	// TODO: Implement container start
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Container start not implemented yet"})
}

func (s *Server) handleStopContainer(c *gin.Context) {
	// TODO: Implement container stop
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Container stop not implemented yet"})
}
