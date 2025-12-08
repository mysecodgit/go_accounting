package unit

type UnitService struct {
	repo UnitRepository
}

func NewUnitService(repo UnitRepository) *UnitService {
	return &UnitService{repo: repo}
}

func (s *UnitService) CreateUnit(unit Unit) (*Unit, map[string]string, error) {
	// Field validation
	if errs := unit.Validate(); errs != nil {
		return nil, errs, nil // validation errors
	}

	// Check if unit name already exists in this building
	exists, err := s.repo.UnitNameExists(unit.Name, unit.BuildingID, 0)
	if err != nil {
		return nil, nil, err // database error
	}
	if exists {
		return nil, map[string]string{"name": "A unit with this name already exists in this building"}, nil
	}

	// Save to DB
	createdUnit, err := s.repo.Create(unit)
	if err != nil {
		return nil, nil, err // internal/server error
	}

	return &createdUnit, nil, nil // success
}

func (s *UnitService) GetAllUnits() ([]UnitResponse, error) {
	units, buildings, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}

	responses := []UnitResponse{}
	for i, unit := range units {
		response := unit.ToUnitResponse(buildings[i])
		responses = append(responses, response)
	}

	return responses, nil
}

func (s *UnitService) GetUnitsByBuildingID(buildingID int) ([]UnitResponse, error) {
	units, buildings, err := s.repo.GetByBuildingID(buildingID)
	if err != nil {
		return nil, err
	}

	responses := []UnitResponse{}
	for i, unit := range units {
		response := unit.ToUnitResponse(buildings[i])
		responses = append(responses, response)
	}

	return responses, nil
}

func (s *UnitService) GetUnitByID(id int) (*UnitResponse, error) {
	unit, building, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	response := unit.ToUnitResponse(building)
	return &response, nil
}

func (s *UnitService) UpdateUnit(id int, unit Unit) (*Unit, map[string]string, error) {
	// Field validation
	if errs := unit.Validate(); errs != nil {
		return nil, errs, nil // validation errors
	}

	// Check if unit name already exists in this building (excluding current unit)
	exists, err := s.repo.UnitNameExists(unit.Name, unit.BuildingID, id)
	if err != nil {
		return nil, nil, err // database error
	}
	if exists {
		return nil, map[string]string{"name": "A unit with this name already exists in this building"}, nil
	}

	unit.ID = id

	// Update in DB
	updatedUnit, err := s.repo.Update(unit, id)
	if err != nil {
		return nil, nil, err // internal/server error
	}

	return &updatedUnit, nil, nil // success
}

