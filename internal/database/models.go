package database

import (
	"database/sql"
	"time"
)

// VM represents a Firecracker virtual machine
type VM struct {
	ID        string    `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Status    string    `json:"status" db:"status"` // creating, running, stopped, error
	Memory    int64     `json:"memory" db:"memory"` // MB
	CPUs      int       `json:"cpus" db:"cpus"`
	DiskSize  int64     `json:"disk_size" db:"disk_size"` // GB
	IPAddress string    `json:"ip_address" db:"ip_address"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Container represents a Docker container running in a VM
type Container struct {
	ID          string    `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Image       string    `json:"image" db:"image"`
	Status      string    `json:"status" db:"status"` // creating, running, stopped, error
	VMID        string    `json:"vm_id" db:"vm_id"`
	ContainerID string    `json:"container_id" db:"container_id"` // Docker container ID
	Ports       string    `json:"ports" db:"ports"`               // JSON string of port mappings
	Environment string    `json:"environment" db:"environment"`   // JSON string of env vars
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// Database handles SQLite operations
type Database struct {
	db *sql.DB
}

// NewDatabase creates a new database connection
func NewDatabase(dbPath string) (*Database, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	database := &Database{db: db}
	if err := database.createTables(); err != nil {
		return nil, err
	}

	return database, nil
}

// createTables creates the necessary database tables
func (d *Database) createTables() error {
	vmTable := `
	CREATE TABLE IF NOT EXISTS vms (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		status TEXT NOT NULL,
		memory INTEGER NOT NULL,
		cpus INTEGER NOT NULL,
		disk_size INTEGER NOT NULL,
		ip_address TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	containerTable := `
	CREATE TABLE IF NOT EXISTS containers (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		image TEXT NOT NULL,
		status TEXT NOT NULL,
		vm_id TEXT NOT NULL,
		container_id TEXT,
		ports TEXT,
		environment TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (vm_id) REFERENCES vms (id)
	);`

	if _, err := d.db.Exec(vmTable); err != nil {
		return err
	}

	if _, err := d.db.Exec(containerTable); err != nil {
		return err
	}

	return nil
}

// Close closes the database connection
func (d *Database) Close() error {
	return d.db.Close()
}

// CreateVM inserts a new VM into the database
func (d *Database) CreateVM(vm *VM) error {
	query := `
		INSERT INTO vms (id, name, status, memory, cpus, disk_size, ip_address, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	vm.CreatedAt = time.Now()
	vm.UpdatedAt = time.Now()

	_, err := d.db.Exec(query, vm.ID, vm.Name, vm.Status, vm.Memory, vm.CPUs, vm.DiskSize, vm.IPAddress, vm.CreatedAt, vm.UpdatedAt)
	return err
}

// UpdateVM updates an existing VM in the database
func (d *Database) UpdateVM(vm *VM) error {
	query := `
		UPDATE vms SET name=?, status=?, memory=?, cpus=?, disk_size=?, ip_address=?, updated_at=?
		WHERE id=?`

	vm.UpdatedAt = time.Now()

	_, err := d.db.Exec(query, vm.Name, vm.Status, vm.Memory, vm.CPUs, vm.DiskSize, vm.IPAddress, vm.UpdatedAt, vm.ID)
	return err
}

// GetVM retrieves a VM by ID
func (d *Database) GetVM(id string) (*VM, error) {
	query := `SELECT id, name, status, memory, cpus, disk_size, ip_address, created_at, updated_at FROM vms WHERE id=?`

	vm := &VM{}
	err := d.db.QueryRow(query, id).Scan(&vm.ID, &vm.Name, &vm.Status, &vm.Memory, &vm.CPUs, &vm.DiskSize, &vm.IPAddress, &vm.CreatedAt, &vm.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return vm, nil
}

// ListVMs retrieves all VMs
func (d *Database) ListVMs() ([]*VM, error) {
	query := `SELECT id, name, status, memory, cpus, disk_size, ip_address, created_at, updated_at FROM vms ORDER BY created_at DESC`

	rows, err := d.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var vms []*VM
	for rows.Next() {
		vm := &VM{}
		err := rows.Scan(&vm.ID, &vm.Name, &vm.Status, &vm.Memory, &vm.CPUs, &vm.DiskSize, &vm.IPAddress, &vm.CreatedAt, &vm.UpdatedAt)
		if err != nil {
			return nil, err
		}
		vms = append(vms, vm)
	}

	return vms, nil
}

// DeleteVM removes a VM from the database
func (d *Database) DeleteVM(id string) error {
	query := `DELETE FROM vms WHERE id=?`
	_, err := d.db.Exec(query, id)
	return err
}

// CreateContainer inserts a new container into the database
func (d *Database) CreateContainer(container *Container) error {
	query := `
		INSERT INTO containers (id, name, image, status, vm_id, container_id, ports, environment, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	container.CreatedAt = time.Now()
	container.UpdatedAt = time.Now()

	_, err := d.db.Exec(query, container.ID, container.Name, container.Image, container.Status, container.VMID, container.ContainerID, container.Ports, container.Environment, container.CreatedAt, container.UpdatedAt)
	return err
}

// UpdateContainer updates an existing container in the database
func (d *Database) UpdateContainer(container *Container) error {
	query := `
		UPDATE containers SET name=?, image=?, status=?, vm_id=?, container_id=?, ports=?, environment=?, updated_at=?
		WHERE id=?`

	container.UpdatedAt = time.Now()

	_, err := d.db.Exec(query, container.Name, container.Image, container.Status, container.VMID, container.ContainerID, container.Ports, container.Environment, container.UpdatedAt, container.ID)
	return err
}

// GetContainer retrieves a container by ID
func (d *Database) GetContainer(id string) (*Container, error) {
	query := `SELECT id, name, image, status, vm_id, container_id, ports, environment, created_at, updated_at FROM containers WHERE id=?`

	container := &Container{}
	err := d.db.QueryRow(query, id).Scan(&container.ID, &container.Name, &container.Image, &container.Status, &container.VMID, &container.ContainerID, &container.Ports, &container.Environment, &container.CreatedAt, &container.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return container, nil
}

// ListContainers retrieves all containers
func (d *Database) ListContainers() ([]*Container, error) {
	query := `SELECT id, name, image, status, vm_id, container_id, ports, environment, created_at, updated_at FROM containers ORDER BY created_at DESC`

	rows, err := d.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var containers []*Container
	for rows.Next() {
		container := &Container{}
		err := rows.Scan(&container.ID, &container.Name, &container.Image, &container.Status, &container.VMID, &container.ContainerID, &container.Ports, &container.Environment, &container.CreatedAt, &container.UpdatedAt)
		if err != nil {
			return nil, err
		}
		containers = append(containers, container)
	}

	return containers, nil
}

// ListContainersByVM retrieves containers for a specific VM
func (d *Database) ListContainersByVM(vmID string) ([]*Container, error) {
	query := `SELECT id, name, image, status, vm_id, container_id, ports, environment, created_at, updated_at FROM containers WHERE vm_id=? ORDER BY created_at DESC`

	rows, err := d.db.Query(query, vmID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var containers []*Container
	for rows.Next() {
		container := &Container{}
		err := rows.Scan(&container.ID, &container.Name, &container.Image, &container.Status, &container.VMID, &container.ContainerID, &container.Ports, &container.Environment, &container.CreatedAt, &container.UpdatedAt)
		if err != nil {
			return nil, err
		}
		containers = append(containers, container)
	}

	return containers, nil
}

// DeleteContainer removes a container from the database
func (d *Database) DeleteContainer(id string) error {
	query := `DELETE FROM containers WHERE id=?`
	_, err := d.db.Exec(query, id)
	return err
}
