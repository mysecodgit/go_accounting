package building

type BuildingService struct {
	repo BuildingRepository
}

func NewBuildingService(repo BuildingRepository) *BuildingService {
	return &BuildingService{repo: repo}
}

func (s *BuildingService) CreateBuilding(building Building) (*Building, map[string]string, error) {
	// Field validation
	if errs := building.Validate(); errs != nil {
		return nil, errs, nil // validation errors
	}

	// Save to DB
	createdBuilding, err := s.repo.Create(building)
	if err != nil {
		return nil, nil, err // internal/server error
	}

	return &createdBuilding, nil, nil // success
}

func (s *BuildingService) GetAllBuildings() ([]Building, error) {
	buildings, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}

	return buildings, nil
}

func (s *BuildingService) GetBuildingByID(id int) (*Building, error) {
	building, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	return &building, nil
}

func (s *BuildingService) UpdateBuilding(id int, building Building) (*Building, map[string]string, error) {
	// Field validation
	if errs := building.Validate(); errs != nil {
		return nil, errs, nil // validation errors
	}

	building.ID = id

	// Update in DB
	updatedBuilding, err := s.repo.Update(building, id)
	if err != nil {
		return nil, nil, err // internal/server error
	}

	return &updatedBuilding, nil, nil // success
}
