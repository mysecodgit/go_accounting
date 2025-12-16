package readings

import (
	"database/sql"
	"fmt"

	"github.com/mysecodgit/go_accounting/src/items"
	"github.com/mysecodgit/go_accounting/src/leases"
	"github.com/mysecodgit/go_accounting/src/people"
	"github.com/mysecodgit/go_accounting/src/unit"
)

type ReadingService struct {
	readingRepo ReadingRepository
	itemRepo    items.ItemRepository
	unitRepo    unit.UnitRepository
	leaseRepo   leases.LeaseRepository
	peopleRepo  people.PersonRepository
	db          *sql.DB
}

func NewReadingService(
	readingRepo ReadingRepository,
	itemRepo items.ItemRepository,
	unitRepo unit.UnitRepository,
	leaseRepo leases.LeaseRepository,
	peopleRepo people.PersonRepository,
	db *sql.DB,
) *ReadingService {
	return &ReadingService{
		readingRepo: readingRepo,
		itemRepo:    itemRepo,
		unitRepo:    unitRepo,
		leaseRepo:   leaseRepo,
		peopleRepo:  peopleRepo,
		db:          db,
	}
}

func (s *ReadingService) CreateReading(req CreateReadingRequest) (*ReadingResponse, error) {
	reading := Reading{
		ItemID:        req.ItemID,
		UnitID:        req.UnitID,
		LeaseID:       req.LeaseID,
		ReadingMonth:  req.ReadingMonth,
		ReadingYear:   req.ReadingYear,
		ReadingDate:   req.ReadingDate,
		PreviousValue: req.PreviousValue,
		CurrentValue:  req.CurrentValue,
		UnitPrice:     req.UnitPrice,
		TotalAmount:   req.TotalAmount,
		Notes:         req.Notes,
		Status:        req.Status,
	}

	if req.Status == "" {
		reading.Status = "1"
	}

	if errors := reading.Validate(); errors != nil {
		return nil, fmt.Errorf("validation failed: %v", errors)
	}

	createdReading, err := s.readingRepo.Create(reading)
	if err != nil {
		return nil, fmt.Errorf("failed to create reading: %v", err)
	}

	// Fetch related entities
	item, _, _, _, _, _, err := s.itemRepo.GetByID(createdReading.ItemID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch item: %v", err)
	}

	unitData, _, err := s.unitRepo.GetByID(createdReading.UnitID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch unit: %v", err)
	}

	var leaseData *leases.Lease
	if createdReading.LeaseID != nil {
		lease, err := s.leaseRepo.GetByID(*createdReading.LeaseID)
		if err == nil {
			leaseData = &lease
		}
	}

	return &ReadingResponse{
		Reading: createdReading,
		Item:    item,
		Unit:    unitData,
		Lease:   leaseData,
	}, nil
}

func (s *ReadingService) UpdateReading(req UpdateReadingRequest) (*ReadingResponse, error) {
	reading := Reading{
		ID:            req.ID,
		ItemID:        req.ItemID,
		UnitID:        req.UnitID,
		LeaseID:       req.LeaseID,
		ReadingMonth:  req.ReadingMonth,
		ReadingYear:   req.ReadingYear,
		ReadingDate:   req.ReadingDate,
		PreviousValue: req.PreviousValue,
		CurrentValue:  req.CurrentValue,
		UnitPrice:     req.UnitPrice,
		TotalAmount:   req.TotalAmount,
		Notes:         req.Notes,
		Status:        req.Status,
	}

	if errors := reading.Validate(); errors != nil {
		return nil, fmt.Errorf("validation failed: %v", errors)
	}

	updatedReading, err := s.readingRepo.Update(reading)
	if err != nil {
		return nil, fmt.Errorf("failed to update reading: %v", err)
	}

	// Fetch related entities
	item, _, _, _, _, _, err := s.itemRepo.GetByID(updatedReading.ItemID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch item: %v", err)
	}

	unitData, _, err := s.unitRepo.GetByID(updatedReading.UnitID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch unit: %v", err)
	}

	var leaseData *leases.Lease
	if updatedReading.LeaseID != nil {
		lease, err := s.leaseRepo.GetByID(*updatedReading.LeaseID)
		if err == nil {
			leaseData = &lease
		}
	}

	return &ReadingResponse{
		Reading: updatedReading,
		Item:    item,
		Unit:    unitData,
		Lease:   leaseData,
	}, nil
}

func (s *ReadingService) GetReadingByID(id int) (*ReadingResponse, error) {
	reading, err := s.readingRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Fetch related entities
	item, _, _, _, _, _, err := s.itemRepo.GetByID(reading.ItemID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch item: %v", err)
	}

	unitData, _, err := s.unitRepo.GetByID(reading.UnitID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch unit: %v", err)
	}

	var leaseData *leases.Lease
	if reading.LeaseID != nil {
		lease, err := s.leaseRepo.GetByID(*reading.LeaseID)
		if err == nil {
			leaseData = &lease
		}
	}

	return &ReadingResponse{
		Reading: reading,
		Item:    item,
		Unit:    unitData,
		Lease:   leaseData,
	}, nil
}

func (s *ReadingService) GetReadingsByBuildingID(buildingID int, status *string) ([]ReadingListItem, error) {
	readings, err := s.readingRepo.GetByBuildingID(buildingID, status)
	if err != nil {
		return nil, err
	}

	result := []ReadingListItem{}
	for _, reading := range readings {
		// Fetch related entities
		item, _, _, _, _, _, err := s.itemRepo.GetByID(reading.ItemID)
		if err != nil {
			continue // Skip if item not found
		}

		unitData, _, err := s.unitRepo.GetByID(reading.UnitID)
		if err != nil {
			continue // Skip if unit not found
		}

		var leaseData *LeaseWithPeople
		if reading.LeaseID != nil && *reading.LeaseID > 0 {
			lease, err := s.leaseRepo.GetByID(*reading.LeaseID)
			if err == nil {
				leaseWithPeople := &LeaseWithPeople{
					Lease: lease,
				}
				// Fetch people information if people_id exists
				if lease.PeopleID > 0 {
					person, _, _, err := s.peopleRepo.GetByID(lease.PeopleID)
					if err == nil {
						leaseWithPeople.People = &person
					}
				}
				leaseData = leaseWithPeople
			}
		}

		result = append(result, ReadingListItem{
			Reading: reading,
			Item:    item,
			Unit:    unitData,
			Lease:   leaseData,
		})
	}

	return result, nil
}

func (s *ReadingService) GetReadingsByUnitID(unitID int) ([]ReadingListItem, error) {
	readings, err := s.readingRepo.GetByUnitID(unitID)
	if err != nil {
		return nil, err
	}

	var readingListItems []ReadingListItem
	for _, reading := range readings {
		// Only include active readings
		if reading.Status != "1" {
			continue
		}

		item, _, _, _, _, _, err := s.itemRepo.GetByID(reading.ItemID)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch item for reading %d: %v", reading.ID, err)
		}

		unitData, _, err := s.unitRepo.GetByID(reading.UnitID)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch unit for reading %d: %v", reading.ID, err)
		}

		var leaseData *LeaseWithPeople
		if reading.LeaseID != nil {
			lease, err := s.leaseRepo.GetByID(*reading.LeaseID)
			if err == nil {
				leaseWithPeople := &LeaseWithPeople{
					Lease: lease,
				}
				// Fetch people information if people_id exists
				if lease.PeopleID > 0 {
					person, _, _, err := s.peopleRepo.GetByID(lease.PeopleID)
					if err == nil {
						leaseWithPeople.People = &person
					}
				}
				leaseData = leaseWithPeople
			}
		}

		readingListItems = append(readingListItems, ReadingListItem{
			Reading: reading,
			Item:    item,
			Unit:    unitData,
			Lease:   leaseData,
		})
	}

	return readingListItems, nil
}

func (s *ReadingService) GetLatestReadingByItemAndUnit(itemID, unitID int) (*Reading, error) {
	return s.readingRepo.GetLatestByItemAndUnit(itemID, unitID)
}

func (s *ReadingService) BulkImportReadings(req BulkImportReadingsRequest) (*BulkImportReadingsResponse, error) {
	// Start transaction
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %v", err)
	}

	committed := false
	defer func() {
		if !committed {
			tx.Rollback()
		}
	}()

	successCount := 0
	failedCount := 0
	errors := []string{}

	for i, readingReq := range req.Readings {
		// Convert request to Reading model
		reading := Reading{
			ItemID:        readingReq.ItemID,
			UnitID:        readingReq.UnitID,
			LeaseID:       readingReq.LeaseID,
			ReadingMonth:  readingReq.ReadingMonth,
			ReadingYear:   readingReq.ReadingYear,
			ReadingDate:   readingReq.ReadingDate,
			PreviousValue: readingReq.PreviousValue,
			CurrentValue:  readingReq.CurrentValue,
			UnitPrice:     readingReq.UnitPrice,
			TotalAmount:   readingReq.TotalAmount,
			Notes:         readingReq.Notes,
			Status:        readingReq.Status,
		}

		if reading.Status == "" {
			reading.Status = "1"
		}

		// Validate
		if validationErrors := reading.Validate(); validationErrors != nil {
			failedCount++
			errors = append(errors, fmt.Sprintf("Row %d: validation failed - %v", i+1, validationErrors))
			continue
		}

		// Create with transaction
		_, err := s.readingRepo.CreateWithTx(tx, reading)
		if err != nil {
			failedCount++
			errors = append(errors, fmt.Sprintf("Row %d: failed to create reading - %v", i+1, err))
			continue
		}

		successCount++
	}

	// If any row failed, rollback everything
	if failedCount > 0 {
		tx.Rollback()
		return &BulkImportReadingsResponse{
			SuccessCount: 0,
			FailedCount:  failedCount,
			Errors:       errors,
		}, fmt.Errorf("import failed: %d rows failed", failedCount)
	}

	// Commit transaction if all rows succeeded
	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %v", err)
	}

	committed = true

	return &BulkImportReadingsResponse{
		SuccessCount: successCount,
		FailedCount:  0,
		Errors:       nil,
	}, nil
}

func (s *ReadingService) DeleteReading(id int) error {
	return s.readingRepo.Delete(id)
}
