package people_types

import "fmt"

type PeopleTypeService struct {
	repo PeopleTypeRepository
}

func NewPeopleTypeService(repo PeopleTypeRepository) *PeopleTypeService {
	return &PeopleTypeService{repo: repo}
}

func (s *PeopleTypeService) CreatePeopleType(peopleType PeopleType) (*PeopleType, map[string]string, error) {
	// Field validation
	if errs := peopleType.Validate(); errs != nil {
		return nil, errs, nil // validation errors
	}

	// Check if title already exists in this building
	exists, err := s.repo.TitleExistsInBuilding(peopleType.Title, peopleType.BuildingID)
	if err != nil {
		return nil, nil, err // internal/server error
	}
	if exists {
		errors := make(map[string]string)
		errors["title"] = fmt.Sprintf("this title exists in this building")
		return nil, errors, nil
	}

	// Save to DB
	createdPeopleType, err := s.repo.Create(peopleType)
	if err != nil {
		return nil, nil, err // internal/server error
	}

	return &createdPeopleType, nil, nil // success
}

func (s *PeopleTypeService) GetAllPeopleTypes() ([]PeopleTypeResponse, error) {
	peopleTypes, buildings, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}

	responses := []PeopleTypeResponse{}
	for i, peopleType := range peopleTypes {
		response := peopleType.ToPeopleTypeResponse(buildings[i])
		responses = append(responses, response)
	}

	return responses, nil
}

func (s *PeopleTypeService) GetPeopleTypeByID(id int) (*PeopleTypeResponse, error) {
	peopleType, building, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	response := peopleType.ToPeopleTypeResponse(building)
	return &response, nil
}

func (s *PeopleTypeService) UpdatePeopleType(id int, updateReq UpdatePeopleTypeRequest) (*PeopleType, map[string]string, error) {
	// Field validation
	if errs := updateReq.Validate(); errs != nil {
		return nil, errs, nil // validation errors
	}

	// Get existing people_type to check building_id for duplicate title check
	existingPeopleType, _, err := s.repo.GetByID(id)
	if err != nil {
		return nil, nil, err // internal/server error
	}

	// Check if title already exists in this building (excluding current record)
	exists, err := s.repo.TitleExistsInBuildingExcludingID(updateReq.Title, existingPeopleType.BuildingID, id)
	if err != nil {
		return nil, nil, err // internal/server error
	}
	if exists {
		errors := make(map[string]string)
		errors["title"] = "this title exists in this building"
		return nil, errors, nil
	}

	// Update only title in DB
	updatedPeopleType, err := s.repo.UpdateTitle(updateReq.Title, id)
	if err != nil {
		return nil, nil, err // internal/server error
	}

	return &updatedPeopleType, nil, nil // success
}

