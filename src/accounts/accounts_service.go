package accounts

type AccountService struct {
	repo AccountRepository
}

func NewAccountService(repo AccountRepository) *AccountService {
	return &AccountService{repo: repo}
}

func (s *AccountService) CreateAccount(account Account) (*Account, map[string]string, error) {
	// Field validation
	if errs := account.Validate(); errs != nil {
		return nil, errs, nil // validation errors
	}

	// Check if account type exists
	accountTypeExists, err := s.repo.AccountTypeIDExists(account.AccountType)
	if err != nil {
		return nil, nil, err
	}
	if !accountTypeExists {
		return nil, map[string]string{"account_type": "Account type does not exist"}, nil
	}

	// Check if building exists
	buildingExists, err := s.repo.BuildingIDExists(account.BuildingID)
	if err != nil {
		return nil, nil, err
	}
	if !buildingExists {
		return nil, map[string]string{"building_id": "Building does not exist"}, nil
	}

	// Save to DB
	createdAccount, err := s.repo.Create(account)
	if err != nil {
		return nil, nil, err // internal/server error
	}

	return &createdAccount, nil, nil // success
}

func (s *AccountService) GetAllAccounts() ([]AccountResponse, error) {
	accounts, accountTypes, buildings, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}

	responses := []AccountResponse{}
	for i, account := range accounts {
		response := account.ToAccountResponse(accountTypes[i], buildings[i])
		responses = append(responses, response)
	}

	return responses, nil
}

func (s *AccountService) GetAccountsByBuildingID(buildingID int) ([]AccountResponse, error) {
	accounts, accountTypes, buildings, err := s.repo.GetByBuildingID(buildingID)
	if err != nil {
		return nil, err
	}

	responses := []AccountResponse{}
	for i, account := range accounts {
		response := account.ToAccountResponse(accountTypes[i], buildings[i])
		responses = append(responses, response)
	}

	return responses, nil
}

func (s *AccountService) GetAccountByID(id int) (*AccountResponse, error) {
	account, accountType, building, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	response := account.ToAccountResponse(accountType, building)
	return &response, nil
}

func (s *AccountService) UpdateAccount(id int, account Account) (*Account, map[string]string, error) {
	// Field validation
	if errs := account.Validate(); errs != nil {
		return nil, errs, nil // validation errors
	}

	// Check if account type exists
	accountTypeExists, err := s.repo.AccountTypeIDExists(account.AccountType)
	if err != nil {
		return nil, nil, err
	}
	if !accountTypeExists {
		return nil, map[string]string{"account_type": "Account type does not exist"}, nil
	}

	// Check if building exists
	buildingExists, err := s.repo.BuildingIDExists(account.BuildingID)
	if err != nil {
		return nil, nil, err
	}
	if !buildingExists {
		return nil, map[string]string{"building_id": "Building does not exist"}, nil
	}

	account.ID = id

	// Update in DB
	updatedAccount, err := s.repo.Update(account, id)
	if err != nil {
		return nil, nil, err // internal/server error
	}

	return &updatedAccount, nil, nil // success
}



