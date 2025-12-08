package people

type PersonService struct {
	repo PersonRepository
}

func NewPersonService(repo PersonRepository) *PersonService {
	return &PersonService{repo: repo}
}

func (s *PersonService) CreatePerson(person Person) (*Person, map[string]string, error) {
	// Field validation
	if errs := person.Validate(); errs != nil {
		return nil, errs, nil // validation errors
	}

	// Check if type_id exists (people_types are now universal)
	typeExists, _, err := s.repo.TypeIDExists(person.TypeID)
	if err != nil {
		return nil, nil, err // internal/server error
	}
	if !typeExists {
		errors := make(map[string]string)
		errors["type_id"] = "type_id does not exist"
		return nil, errors, nil
	}

	// Check if person name already exists in this building
	exists, err := s.repo.PersonNameExists(person.Name, person.BuildingID, 0)
	if err != nil {
		return nil, nil, err // database error
	}
	if exists {
		return nil, map[string]string{"name": "A person with this name already exists in this building"}, nil
	}

	// building_id must be provided in the request (no longer comes from people_types)

	// Save to DB
	createdPerson, err := s.repo.Create(person)
	if err != nil {
		return nil, nil, err // internal/server error
	}

	return &createdPerson, nil, nil // success
}

func (s *PersonService) GetAllPeople() ([]PersonResponse, error) {
	people, peopleTypes, buildings, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}

	responses := []PersonResponse{}
	for i, person := range people {
		response := person.ToPersonResponse(peopleTypes[i], buildings[i])
		responses = append(responses, response)
	}

	return responses, nil
}

func (s *PersonService) GetPeopleByBuildingID(buildingID int) ([]PersonResponse, error) {
	people, peopleTypes, buildings, err := s.repo.GetByBuildingID(buildingID)
	if err != nil {
		return nil, err
	}

	responses := []PersonResponse{}
	for i, person := range people {
		response := person.ToPersonResponse(peopleTypes[i], buildings[i])
		responses = append(responses, response)
	}

	return responses, nil
}

func (s *PersonService) GetPersonByID(id int) (*PersonResponse, error) {
	person, peopleType, building, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	response := person.ToPersonResponse(peopleType, building)
	return &response, nil
}

func (s *PersonService) UpdatePerson(id int, updateReq UpdatePersonRequest) (*Person, map[string]string, error) {
	// Field validation
	if errs := updateReq.Validate(); errs != nil {
		return nil, errs, nil // validation errors
	}

	// Check if type_id exists (people_types are now universal)
	typeExists, _, err := s.repo.TypeIDExists(updateReq.TypeID)
	if err != nil {
		return nil, nil, err // internal/server error
	}
	if !typeExists {
		errors := make(map[string]string)
		errors["type_id"] = "type_id does not exist"
		return nil, errors, nil
	}

	// Get existing person to get building_id
	existingPerson, _, _, err := s.repo.GetByID(id)
	if err != nil {
		return nil, nil, err // person not found or database error
	}

	// Check if person name already exists in this building (excluding current person)
	exists, err := s.repo.PersonNameExists(updateReq.Name, existingPerson.BuildingID, id)
	if err != nil {
		return nil, nil, err // database error
	}
	if exists {
		return nil, map[string]string{"name": "A person with this name already exists in this building"}, nil
	}

	// Update name, phone, and type_id in DB
	updatedPerson, err := s.repo.UpdateNameAndPhone(updateReq.Name, updateReq.Phone, updateReq.TypeID, id)
	if err != nil {
		return nil, nil, err // internal/server error
	}

	return &updatedPerson, nil, nil // success
}
