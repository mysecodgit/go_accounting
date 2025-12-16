package leases

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mysecodgit/go_accounting/src/people"
	"github.com/mysecodgit/go_accounting/src/people_types"
)

type LeaseService struct {
	leaseRepo      LeaseRepository
	leaseFileRepo  LeaseFileRepository
	peopleRepo     people.PersonRepository
	peopleTypeRepo people_types.PeopleTypeRepository
	db             *sql.DB
}

func NewLeaseService(
	leaseRepo LeaseRepository,
	leaseFileRepo LeaseFileRepository,
	peopleRepo people.PersonRepository,
	peopleTypeRepo people_types.PeopleTypeRepository,
	db *sql.DB,
) *LeaseService {
	return &LeaseService{
		leaseRepo:      leaseRepo,
		leaseFileRepo:  leaseFileRepo,
		peopleRepo:     peopleRepo,
		peopleTypeRepo: peopleTypeRepo,
		db:             db,
	}
}

// GetCustomersByBuildingID gets only people with type "customer"
func (s *LeaseService) GetCustomersByBuildingID(buildingID int) ([]people.Person, error) {
	// Get all people for the building
	peopleList, typesList, _, err := s.peopleRepo.GetByBuildingID(buildingID)
	if err != nil {
		return nil, err
	}

	// Find the "customer" type
	customerTypeID := 0
	for _, pt := range typesList {
		if strings.ToLower(pt.Title) == "customer" {
			customerTypeID = pt.ID
			break
		}
	}

	if customerTypeID == 0 {
		return []people.Person{}, nil
	}

	// Filter people to only customers
	customers := []people.Person{}
	for _, p := range peopleList {
		if p.TypeID == customerTypeID {
			customers = append(customers, p)
		}
	}

	return customers, nil
}

// GetCustomersWithLeaseUnits gets customers with their active lease unit information
func (s *LeaseService) GetCustomersWithLeaseUnits(buildingID int) ([]map[string]interface{}, error) {
	// Get all people for the building
	peopleList, typesList, _, err := s.peopleRepo.GetByBuildingID(buildingID)
	if err != nil {
		return nil, err
	}

	// Find the "customer" type
	customerTypeID := 0
	for _, pt := range typesList {
		if strings.ToLower(pt.Title) == "customer" {
			customerTypeID = pt.ID
			break
		}
	}

	if customerTypeID == 0 {
		return []map[string]interface{}{}, nil
	}

	// Filter people to only customers
	customers := []people.Person{}
	for _, p := range peopleList {
		if p.TypeID == customerTypeID {
			customers = append(customers, p)
		}
	}

	// Get active leases with unit information for these customers
	query := `
		SELECT l.people_id, l.unit_id, u.name as unit_name
		FROM leases l
		INNER JOIN units u ON l.unit_id = u.id
		WHERE l.building_id = ? AND l.status = '1'
	`
	rows, err := s.db.Query(query, buildingID)
	if err != nil {
		return nil, fmt.Errorf("failed to query leases: %v", err)
	}
	defer rows.Close()

	// Map people_id to unit information
	leaseUnits := make(map[int]map[string]interface{})
	for rows.Next() {
		var peopleID, unitID int
		var unitName string
		if err := rows.Scan(&peopleID, &unitID, &unitName); err != nil {
			return nil, fmt.Errorf("failed to scan lease: %v", err)
		}
		leaseUnits[peopleID] = map[string]interface{}{
			"unit_id":   unitID,
			"unit_name": unitName,
		}
	}

	// Build result with customers and their unit information
	result := []map[string]interface{}{}
	for _, customer := range customers {
		customerData := map[string]interface{}{
			"id":   customer.ID,
			"name": customer.Name,
		}
		if unitInfo, hasLease := leaseUnits[customer.ID]; hasLease {
			customerData["unit_id"] = unitInfo["unit_id"]
			customerData["unit_name"] = unitInfo["unit_name"]
		} else {
			customerData["unit_id"] = nil
			customerData["unit_name"] = nil
		}
		result = append(result, customerData)
	}

	return result, nil
}

// GetAvailableUnits gets units that don't have active leases, optionally including a specific unit ID
func (s *LeaseService) GetAvailableUnits(buildingID int, includeUnitID *int) ([]interface{}, error) {
	// Build query to get units without active leases
	query := `
		SELECT DISTINCT u.id, u.name, u.building_id
		FROM units u
		WHERE u.building_id = ?
			AND u.id NOT IN (
				SELECT DISTINCT l.unit_id
				FROM leases l
				WHERE l.building_id = ? AND l.status = '1'
			)
	`
	args := []interface{}{buildingID, buildingID}

	// If includeUnitID is provided, also include that unit
	if includeUnitID != nil && *includeUnitID > 0 {
		query = `
			SELECT DISTINCT u.id, u.name, u.building_id
			FROM units u
			WHERE u.building_id = ?
				AND (
					u.id NOT IN (
						SELECT DISTINCT l.unit_id
						FROM leases l
						WHERE l.building_id = ? AND l.status = '1'
					)
					OR u.id = ?
				)
		`
		args = []interface{}{buildingID, buildingID, *includeUnitID}
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query available units: %v", err)
	}
	defer rows.Close()

	units := []interface{}{}
	for rows.Next() {
		var id, buildingID int
		var name string
		if err := rows.Scan(&id, &name, &buildingID); err != nil {
			return nil, fmt.Errorf("failed to scan unit: %v", err)
		}
		units = append(units, map[string]interface{}{
			"id":          id,
			"name":        name,
			"building_id": buildingID,
		})
	}

	return units, nil
}

// GetUnitsByPeopleID gets units for a specific people based on their active leases
func (s *LeaseService) GetUnitsByPeopleID(buildingID, peopleID int) ([]interface{}, error) {
	query := `
		SELECT DISTINCT u.id, u.name, u.building_id
		FROM units u
		INNER JOIN leases l ON u.id = l.unit_id
		WHERE l.building_id = ? AND l.people_id = ? AND l.status = '1'
		ORDER BY u.name
	`

	rows, err := s.db.Query(query, buildingID, peopleID)
	if err != nil {
		return nil, fmt.Errorf("failed to query units for people: %v", err)
	}
	defer rows.Close()

	units := []interface{}{}
	for rows.Next() {
		var id, buildingID int
		var name string
		if err := rows.Scan(&id, &name, &buildingID); err != nil {
			return nil, fmt.Errorf("failed to scan unit: %v", err)
		}
		units = append(units, map[string]interface{}{
			"id":          id,
			"name":        name,
			"building_id": buildingID,
		})
	}

	return units, nil
}

func (s *LeaseService) CreateLease(req CreateLeaseRequest) (*LeaseResponse, error) {
	lease := Lease{
		PeopleID:      req.PeopleID,
		BuildingID:    req.BuildingID,
		UnitID:        req.UnitID,
		StartDate:     req.StartDate,
		EndDate:       req.EndDate,
		RentAmount:    req.RentAmount,
		DepositAmount: req.DepositAmount,
		ServiceAmount: req.ServiceAmount,
		LeaseTerms:    req.LeaseTerms,
		Status:        req.Status,
	}

	if errors := lease.Validate(); errors != nil {
		return nil, fmt.Errorf("validation failed: %v", errors)
	}

	createdLease, err := s.leaseRepo.Create(lease)
	if err != nil {
		return nil, fmt.Errorf("failed to create lease: %v", err)
	}

	// Get lease files (empty initially)
	leaseFiles, _ := s.leaseFileRepo.GetByLeaseID(createdLease.ID)

	return &LeaseResponse{
		Lease:      createdLease,
		LeaseFiles: leaseFiles,
	}, nil
}

func (s *LeaseService) UpdateLease(req UpdateLeaseRequest) (*LeaseResponse, error) {
	lease := Lease{
		ID:            req.ID,
		PeopleID:      req.PeopleID,
		BuildingID:    req.BuildingID,
		UnitID:        req.UnitID,
		StartDate:     req.StartDate,
		EndDate:       req.EndDate,
		RentAmount:    req.RentAmount,
		DepositAmount: req.DepositAmount,
		ServiceAmount: req.ServiceAmount,
		LeaseTerms:    req.LeaseTerms,
		Status:        req.Status,
	}

	if errors := lease.Validate(); errors != nil {
		return nil, fmt.Errorf("validation failed: %v", errors)
	}

	updatedLease, err := s.leaseRepo.Update(lease)
	if err != nil {
		return nil, fmt.Errorf("failed to update lease: %v", err)
	}

	// Get lease files
	leaseFiles, _ := s.leaseFileRepo.GetByLeaseID(updatedLease.ID)

	return &LeaseResponse{
		Lease:      updatedLease,
		LeaseFiles: leaseFiles,
	}, nil
}

func (s *LeaseService) GetLeaseByID(id int) (*LeaseResponse, error) {
	lease, err := s.leaseRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	leaseFiles, _ := s.leaseFileRepo.GetByLeaseID(id)

	return &LeaseResponse{
		Lease:      lease,
		LeaseFiles: leaseFiles,
	}, nil
}

func (s *LeaseService) GetLeasesByBuildingID(buildingID int) ([]LeaseListItem, error) {
	leases, err := s.leaseRepo.GetByBuildingID(buildingID)
	if err != nil {
		return nil, err
	}

	result := []LeaseListItem{}
	for _, lease := range leases {
		item := LeaseListItem{
			Lease: lease,
		}

		// Fetch people information if people_id exists
		if lease.PeopleID > 0 {
			person, _, _, err := s.peopleRepo.GetByID(lease.PeopleID)
			if err == nil {
				item.People = &person
			}
		}

		result = append(result, item)
	}

	return result, nil
}

func (s *LeaseService) GetLeasesByUnitID(unitID int) ([]LeaseListItem, error) {
	leases, err := s.leaseRepo.GetByUnitID(unitID)
	if err != nil {
		return nil, err
	}

	result := []LeaseListItem{}
	for _, lease := range leases {
		item := LeaseListItem{
			Lease: lease,
		}

		// Fetch people information if people_id exists
		if lease.PeopleID > 0 {
			person, _, _, err := s.peopleRepo.GetByID(lease.PeopleID)
			if err == nil {
				item.People = &person
			}
		}

		result = append(result, item)
	}

	return result, nil
}

func (s *LeaseService) DeleteLease(id int) error {
	// Start transaction to ensure atomicity
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %v", err)
	}

	committed := false
	defer func() {
		if !committed {
			tx.Rollback()
		}
	}()

	// Get all associated files first
	leaseFiles, err := s.leaseFileRepo.GetByLeaseID(id)
	if err != nil {
		// If we can't get files, still try to delete the lease
	} else {
		// Delete all files within transaction
		for _, file := range leaseFiles {
			// Delete database record first (within transaction)
			_, err = tx.Exec("DELETE FROM lease_files WHERE id = ?", file.ID)
			if err != nil {
				return fmt.Errorf("failed to delete file record %d: %v", file.ID, err)
			}
			// Delete physical file (outside transaction, but if DB delete succeeds, we should delete file)
			if _, statErr := os.Stat(file.FilePath); statErr == nil {
				os.Remove(file.FilePath)
			}
		}
	}

	// Delete lease within transaction
	_, err = tx.Exec("UPDATE leases SET status = '0' WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete lease: %v", err)
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}
	committed = true

	return nil
}

func (s *LeaseService) UploadLeaseFile(leaseID int, filename, originalName, filePath, fileType string, fileSize int64) (*LeaseFile, error) {
	// Verify lease exists before uploading file
	_, err := s.leaseRepo.GetByID(leaseID)
	if err != nil {
		// If lease doesn't exist, delete the uploaded file
		if _, statErr := os.Stat(filePath); statErr == nil {
			os.Remove(filePath)
		}
		return nil, fmt.Errorf("lease not found: %v", err)
	}

	leaseFile := LeaseFile{
		LeaseID:      leaseID,
		Filename:     filename,
		OriginalName: originalName,
		FilePath:     filePath,
		FileType:     fileType,
		FileSize:     fileSize,
	}

	createdFile, err := s.leaseFileRepo.Create(leaseFile)
	if err != nil {
		// If database insert fails, delete the uploaded file
		if _, statErr := os.Stat(filePath); statErr == nil {
			os.Remove(filePath)
		}
		return nil, fmt.Errorf("failed to save file record: %v", err)
	}

	return &createdFile, nil
}

func (s *LeaseService) GetLeaseFileByID(id int) (*LeaseFile, error) {
	file, err := s.leaseFileRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	return &file, nil
}

func (s *LeaseService) DeleteLeaseFile(id int) error {
	file, err := s.leaseFileRepo.GetByID(id)
	if err != nil {
		return err
	}

	// Delete physical file
	if _, err := os.Stat(file.FilePath); err == nil {
		if err := os.Remove(file.FilePath); err != nil {
			return fmt.Errorf("failed to delete file: %v", err)
		}
	}

	// Delete database record
	return s.leaseFileRepo.Delete(id)
}

// GetUploadPath returns the path where lease files should be stored
func GetUploadPath(buildingID int) string {
	return filepath.Join("uploads", "leases", fmt.Sprintf("building_%d", buildingID))
}
