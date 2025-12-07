package account_types

type AccountTypeService struct {
	repo AccountTypeRepository
}

func NewAccountTypeService(repo AccountTypeRepository) *AccountTypeService {
	return &AccountTypeService{repo: repo}
}

func (s *AccountTypeService) CreateAccountType(accountType AccountType) (*AccountType, map[string]string, error) {
	// Field validation
	if errs := accountType.Validate(); errs != nil {
		return nil, errs, nil // validation errors
	}

	// Save to DB
	createdAccountType, err := s.repo.Create(accountType)
	if err != nil {
		return nil, nil, err // internal/server error
	}

	return &createdAccountType, nil, nil // success
}

func (s *AccountTypeService) GetAllAccountTypes() ([]AccountTypeResponse, error) {
	accountTypes, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}

	responses := []AccountTypeResponse{}
	for _, accountType := range accountTypes {
		response := accountType.ToAccountTypeResponse()
		responses = append(responses, response)
	}

	return responses, nil
}

func (s *AccountTypeService) GetAccountTypeByID(id int) (*AccountTypeResponse, error) {
	accountType, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	response := accountType.ToAccountTypeResponse()
	return &response, nil
}

func (s *AccountTypeService) UpdateAccountType(id int, accountType AccountType) (*AccountType, map[string]string, error) {
	// Field validation
	if errs := accountType.Validate(); errs != nil {
		return nil, errs, nil // validation errors
	}

	accountType.ID = id

	// Update in DB
	updatedAccountType, err := s.repo.Update(accountType, id)
	if err != nil {
		return nil, nil, err // internal/server error
	}

	return &updatedAccountType, nil, nil // success
}

