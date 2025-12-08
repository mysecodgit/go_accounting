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

	// Check if title already exists globally
	exists, err := s.repo.TitleExists(peopleType.Title)
	if err != nil {
		return nil, nil, err // internal/server error
	}
	if exists {
		errors := make(map[string]string)
		errors["title"] = fmt.Sprintf("this title already exists")
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
	peopleTypes, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}

	responses := []PeopleTypeResponse{}
	for _, peopleType := range peopleTypes {
		response := PeopleTypeResponse{
			ID:    peopleType.ID,
			Title: peopleType.Title,
		}
		responses = append(responses, response)
	}

	return responses, nil
}

func (s *PeopleTypeService) GetPeopleTypeByID(id int) (*PeopleTypeResponse, error) {
	peopleType, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	response := PeopleTypeResponse{
		ID:    peopleType.ID,
		Title: peopleType.Title,
	}
	return &response, nil
}

func (s *PeopleTypeService) UpdatePeopleType(id int, updateReq UpdatePeopleTypeRequest) (*PeopleType, map[string]string, error) {
	// Field validation
	if errs := updateReq.Validate(); errs != nil {
		return nil, errs, nil // validation errors
	}

	// Check if title already exists globally (excluding current record)
	exists, err := s.repo.TitleExistsExcludingID(updateReq.Title, id)
	if err != nil {
		return nil, nil, err // internal/server error
	}
	if exists {
		errors := make(map[string]string)
		errors["title"] = "this title already exists"
		return nil, errors, nil
	}

	// Update only title in DB
	updatedPeopleType, err := s.repo.UpdateTitle(updateReq.Title, id)
	if err != nil {
		return nil, nil, err // internal/server error
	}

	return &updatedPeopleType, nil, nil // success
}

