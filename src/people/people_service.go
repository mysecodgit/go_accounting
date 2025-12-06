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

	// Check if type_id exists and get its building_id
	typeExists, typeBuildingID, err := s.repo.TypeIDExists(person.TypeID)
	if err != nil {
		return nil, nil, err // internal/server error
	}
	if !typeExists {
		errors := make(map[string]string)
		errors["type_id"] = "type_id does not exist"
		return nil, errors, nil
	}

	// Set building_id from people_types
	person.BuildingID = typeBuildingID

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

	// Update only name and phone in DB
	updatedPerson, err := s.repo.UpdateNameAndPhone(updateReq.Name, updateReq.Phone, id)
	if err != nil {
		return nil, nil, err // internal/server error
	}

	return &updatedPerson, nil, nil // success
}
