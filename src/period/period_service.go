package period

type PeriodService struct {
	repo PeriodRepository
}

func NewPeriodService(repo PeriodRepository) *PeriodService {
	return &PeriodService{repo: repo}
}

func (s *PeriodService) CreatePeriod(period Period) (*Period, map[string]string, error) {
	// Field validation
	if errs := period.Validate(); errs != nil {
		return nil, errs, nil // validation errors
	}

	// Check if building exists
	buildingExists, err := s.repo.BuildingIDExists(period.BuildingID)
	if err != nil {
		return nil, nil, err
	}
	if !buildingExists {
		return nil, map[string]string{"building_id": "Building does not exist"}, nil
	}

	// Save to DB
	createdPeriod, err := s.repo.Create(period)
	if err != nil {
		return nil, nil, err // internal/server error
	}

	return &createdPeriod, nil, nil // success
}

func (s *PeriodService) GetAllPeriods() ([]PeriodResponse, error) {
	periods, buildings, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}

	responses := []PeriodResponse{}
	for i, period := range periods {
		response := period.ToPeriodResponse(buildings[i])
		responses = append(responses, response)
	}

	return responses, nil
}

func (s *PeriodService) GetPeriodsByBuildingID(buildingID int) ([]PeriodResponse, error) {
	periods, buildings, err := s.repo.GetByBuildingID(buildingID)
	if err != nil {
		return nil, err
	}

	responses := []PeriodResponse{}
	for i, period := range periods {
		response := period.ToPeriodResponse(buildings[i])
		responses = append(responses, response)
	}

	return responses, nil
}

func (s *PeriodService) GetPeriodByID(id int) (*PeriodResponse, error) {
	period, building, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	response := period.ToPeriodResponse(building)
	return &response, nil
}

func (s *PeriodService) UpdatePeriod(id int, period Period) (*Period, map[string]string, error) {
	// Field validation
	if errs := period.Validate(); errs != nil {
		return nil, errs, nil // validation errors
	}

	// Check if building exists
	buildingExists, err := s.repo.BuildingIDExists(period.BuildingID)
	if err != nil {
		return nil, nil, err
	}
	if !buildingExists {
		return nil, map[string]string{"building_id": "Building does not exist"}, nil
	}

	period.ID = id

	// Update in DB
	updatedPeriod, err := s.repo.Update(period, id)
	if err != nil {
		return nil, nil, err // internal/server error
	}

	return &updatedPeriod, nil, nil // success
}

