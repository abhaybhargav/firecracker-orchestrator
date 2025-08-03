package firecracker

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/abhaybhargav/firecracker-orchestrator/internal/config"
	"github.com/abhaybhargav/firecracker-orchestrator/internal/database"
	"github.com/sirupsen/logrus"
)

// Manager handles Firecracker VM lifecycle
type Manager struct {
	config   *config.Config
	db       *database.Database
	logger   *logrus.Logger
	vms      map[string]*FirecrackerVM
	tapIndex int
}

// FirecrackerVM represents a running Firecracker VM
type FirecrackerVM struct {
	ID         string
	SocketPath string
	TAPDevice  string
	Process    *os.Process
	Config     *VMConfig
}

// VMConfig represents Firecracker VM configuration
type VMConfig struct {
	BootSource    BootSource     `json:"boot-source"`
	Drives        []Drive        `json:"drives"`
	MachineConfig MachineConfig  `json:"machine-config"`
	NetworkIfaces []NetworkIface `json:"network-interfaces"`
}

type BootSource struct {
	KernelImagePath string `json:"kernel_image_path"`
	BootArgs        string `json:"boot_args"`
}

type Drive struct {
	DriveID      string `json:"drive_id"`
	PathOnHost   string `json:"path_on_host"`
	IsRootDevice bool   `json:"is_root_device"`
	IsReadOnly   bool   `json:"is_read_only"`
}

type MachineConfig struct {
	VCPUCount  int   `json:"vcpu_count"`
	MemSizeMib int64 `json:"mem_size_mib"`
}

type NetworkIface struct {
	IfaceID     string `json:"iface_id"`
	GuestMAC    string `json:"guest_mac"`
	HostDevName string `json:"host_dev_name"`
}

// NewManager creates a new Firecracker manager
func NewManager(config *config.Config, db *database.Database, logger *logrus.Logger) *Manager {
	return &Manager{
		config:   config,
		db:       db,
		logger:   logger,
		vms:      make(map[string]*FirecrackerVM),
		tapIndex: 0,
	}
}

// CreateVM creates a new Firecracker VM
func (m *Manager) CreateVM(vm *database.VM) error {
	m.logger.Infof("Creating VM: %s", vm.ID)

	// Create socket directory
	if err := os.MkdirAll(m.config.SocketDir, 0755); err != nil {
		return fmt.Errorf("failed to create socket directory: %w", err)
	}

	// Generate unique socket path
	socketPath := filepath.Join(m.config.SocketDir, fmt.Sprintf("%s.sock", vm.ID))

	// Create TAP device
	tapDevice := fmt.Sprintf("%s%d", m.config.TAPDeviceBase, m.tapIndex)
	m.tapIndex++

	if err := m.createTAPDevice(tapDevice); err != nil {
		return fmt.Errorf("failed to create TAP device: %w", err)
	}

	// Assign IP address
	ipAddr := m.generateIPAddress()
	vm.IPAddress = ipAddr

	// Create VM configuration
	vmConfig := &VMConfig{
		BootSource: BootSource{
			KernelImagePath: m.config.KernelPath,
			BootArgs:        "console=ttyS0 reboot=k panic=1 pci=off",
		},
		Drives: []Drive{
			{
				DriveID:      "rootfs",
				PathOnHost:   m.config.RootfsPath,
				IsRootDevice: true,
				IsReadOnly:   false,
			},
		},
		MachineConfig: MachineConfig{
			VCPUCount:  vm.CPUs,
			MemSizeMib: vm.Memory,
		},
		NetworkIfaces: []NetworkIface{
			{
				IfaceID:     "eth0",
				GuestMAC:    m.generateMACAddress(),
				HostDevName: tapDevice,
			},
		},
	}

	// Save configuration to file
	configPath := filepath.Join(m.config.SocketDir, fmt.Sprintf("%s-config.json", vm.ID))
	configData, err := json.MarshalIndent(vmConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal VM config: %w", err)
	}

	if err := os.WriteFile(configPath, configData, 0644); err != nil {
		return fmt.Errorf("failed to write VM config: %w", err)
	}

	// Update VM status
	vm.Status = "created"
	if err := m.db.UpdateVM(vm); err != nil {
		return fmt.Errorf("failed to update VM in database: %w", err)
	}

	// Store VM reference
	fcVM := &FirecrackerVM{
		ID:         vm.ID,
		SocketPath: socketPath,
		TAPDevice:  tapDevice,
		Config:     vmConfig,
	}
	m.vms[vm.ID] = fcVM

	m.logger.Infof("VM %s created successfully", vm.ID)
	return nil
}

// StartVM starts a Firecracker VM
func (m *Manager) StartVM(vmID string) error {
	m.logger.Infof("Starting VM: %s", vmID)

	vm, err := m.db.GetVM(vmID)
	if err != nil {
		return fmt.Errorf("failed to get VM from database: %w", err)
	}

	fcVM, exists := m.vms[vmID]
	if !exists {
		return fmt.Errorf("VM %s not found in manager", vmID)
	}

	// Start Firecracker process
	cmd := exec.Command(
		m.config.FirecrackerBinary,
		"--api-sock", fcVM.SocketPath,
		"--config-file", filepath.Join(m.config.SocketDir, fmt.Sprintf("%s-config.json", vmID)),
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start Firecracker: %w", err)
	}

	fcVM.Process = cmd.Process

	// Update VM status
	vm.Status = "running"
	if err := m.db.UpdateVM(vm); err != nil {
		return fmt.Errorf("failed to update VM status: %w", err)
	}

	m.logger.Infof("VM %s started successfully with PID %d", vmID, cmd.Process.Pid)
	return nil
}

// StopVM stops a Firecracker VM
func (m *Manager) StopVM(vmID string) error {
	m.logger.Infof("Stopping VM: %s", vmID)

	vm, err := m.db.GetVM(vmID)
	if err != nil {
		return fmt.Errorf("failed to get VM from database: %w", err)
	}

	fcVM, exists := m.vms[vmID]
	if !exists {
		return fmt.Errorf("VM %s not found in manager", vmID)
	}

	if fcVM.Process != nil {
		if err := fcVM.Process.Kill(); err != nil {
			m.logger.Warnf("Failed to kill VM process: %v", err)
		}
		fcVM.Process = nil
	}

	// Clean up TAP device
	if err := m.deleteTAPDevice(fcVM.TAPDevice); err != nil {
		m.logger.Warnf("Failed to delete TAP device: %v", err)
	}

	// Update VM status
	vm.Status = "stopped"
	if err := m.db.UpdateVM(vm); err != nil {
		return fmt.Errorf("failed to update VM status: %w", err)
	}

	m.logger.Infof("VM %s stopped successfully", vmID)
	return nil
}

// DeleteVM deletes a Firecracker VM
func (m *Manager) DeleteVM(vmID string) error {
	m.logger.Infof("Deleting VM: %s", vmID)

	// Stop VM first if running
	if fcVM, exists := m.vms[vmID]; exists {
		if fcVM.Process != nil {
			if err := m.StopVM(vmID); err != nil {
				m.logger.Warnf("Failed to stop VM during deletion: %v", err)
			}
		}

		// Clean up files
		socketPath := fcVM.SocketPath
		configPath := filepath.Join(m.config.SocketDir, fmt.Sprintf("%s-config.json", vmID))

		os.Remove(socketPath)
		os.Remove(configPath)

		delete(m.vms, vmID)
	}

	// Remove from database
	if err := m.db.DeleteVM(vmID); err != nil {
		return fmt.Errorf("failed to delete VM from database: %w", err)
	}

	m.logger.Infof("VM %s deleted successfully", vmID)
	return nil
}

// ListVMs returns all VMs
func (m *Manager) ListVMs() ([]*database.VM, error) {
	return m.db.ListVMs()
}

// GetVM returns a specific VM
func (m *Manager) GetVM(vmID string) (*database.VM, error) {
	return m.db.GetVM(vmID)
}

// createTAPDevice creates a TAP network device
func (m *Manager) createTAPDevice(name string) error {
	// Note: This is a simplified implementation
	// In production, you'd want more sophisticated networking setup
	cmd := exec.Command("ip", "tuntap", "add", "dev", name, "mode", "tap")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create TAP device %s: %w", name, err)
	}

	cmd = exec.Command("ip", "link", "set", "dev", name, "up")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to bring up TAP device %s: %w", name, err)
	}

	return nil
}

// deleteTAPDevice deletes a TAP network device
func (m *Manager) deleteTAPDevice(name string) error {
	cmd := exec.Command("ip", "link", "delete", name)
	return cmd.Run()
}

// generateIPAddress generates a unique IP address for the VM
func (m *Manager) generateIPAddress() string {
	// Simple implementation - in production you'd want a proper IP pool manager
	return fmt.Sprintf("192.168.100.%d", 10+len(m.vms))
}

// generateMACAddress generates a unique MAC address for the VM
func (m *Manager) generateMACAddress() string {
	// Simple implementation - generates a locally administered MAC
	return fmt.Sprintf("02:00:00:00:00:%02x", len(m.vms)+1)
}
