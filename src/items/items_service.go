package items

type ItemService struct {
	repo ItemRepository
}

func NewItemService(repo ItemRepository) *ItemService {
	return &ItemService{repo: repo}
}

func (s *ItemService) CreateItem(item Item) (*Item, map[string]string, error) {
	// Field validation
	if errs := item.Validate(); errs != nil {
		return nil, errs, nil // validation errors
	}

	// Check if building exists
	buildingExists, err := s.repo.BuildingIDExists(item.BuildingID)
	if err != nil {
		return nil, nil, err
	}
	if !buildingExists {
		return nil, map[string]string{"building_id": "Building does not exist"}, nil
	}

	// Check if accounts exist (if provided)
	if item.AssetAccount != nil {
		accountExists, err := s.repo.AccountIDExists(*item.AssetAccount)
		if err != nil {
			return nil, nil, err
		}
		if !accountExists {
			return nil, map[string]string{"asset_account": "Asset account does not exist"}, nil
		}
	}

	if item.IncomeAccount != nil {
		accountExists, err := s.repo.AccountIDExists(*item.IncomeAccount)
		if err != nil {
			return nil, nil, err
		}
		if !accountExists {
			return nil, map[string]string{"income_account": "Income account does not exist"}, nil
		}
	}

	if item.COGSAccount != nil {
		accountExists, err := s.repo.AccountIDExists(*item.COGSAccount)
		if err != nil {
			return nil, nil, err
		}
		if !accountExists {
			return nil, map[string]string{"cogs_account": "COGS account does not exist"}, nil
		}
	}

	if item.ExpenseAccount != nil {
		accountExists, err := s.repo.AccountIDExists(*item.ExpenseAccount)
		if err != nil {
			return nil, nil, err
		}
		if !accountExists {
			return nil, map[string]string{"expense_account": "Expense account does not exist"}, nil
		}
	}

	// Save to DB
	createdItem, err := s.repo.Create(item)
	if err != nil {
		return nil, nil, err // internal/server error
	}

	return &createdItem, nil, nil // success
}

func (s *ItemService) GetAllItems() ([]ItemResponse, error) {
	items, buildings, assetAccounts, incomeAccounts, cogsAccounts, expenseAccounts, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}

	responses := []ItemResponse{}
	for i, item := range items {
		response := item.ToItemResponse(buildings[i], assetAccounts[i], incomeAccounts[i], cogsAccounts[i], expenseAccounts[i])
		responses = append(responses, response)
	}

	return responses, nil
}

func (s *ItemService) GetItemsByBuildingID(buildingID int) ([]ItemResponse, error) {
	items, buildings, assetAccounts, incomeAccounts, cogsAccounts, expenseAccounts, err := s.repo.GetByBuildingID(buildingID)
	if err != nil {
		return nil, err
	}

	responses := []ItemResponse{}
	for i, item := range items {
		response := item.ToItemResponse(buildings[i], assetAccounts[i], incomeAccounts[i], cogsAccounts[i], expenseAccounts[i])
		responses = append(responses, response)
	}

	return responses, nil
}

func (s *ItemService) GetItemByID(id int) (*ItemResponse, error) {
	item, building, assetAccount, incomeAccount, cogsAccount, expenseAccount, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	response := item.ToItemResponse(building, assetAccount, incomeAccount, cogsAccount, expenseAccount)
	return &response, nil
}

func (s *ItemService) UpdateItem(id int, item Item) (*Item, map[string]string, error) {
	// Field validation
	if errs := item.Validate(); errs != nil {
		return nil, errs, nil // validation errors
	}

	// Check if building exists
	buildingExists, err := s.repo.BuildingIDExists(item.BuildingID)
	if err != nil {
		return nil, nil, err
	}
	if !buildingExists {
		return nil, map[string]string{"building_id": "Building does not exist"}, nil
	}

	// Check if accounts exist (if provided)
	if item.AssetAccount != nil {
		accountExists, err := s.repo.AccountIDExists(*item.AssetAccount)
		if err != nil {
			return nil, nil, err
		}
		if !accountExists {
			return nil, map[string]string{"asset_account": "Asset account does not exist"}, nil
		}
	}

	if item.IncomeAccount != nil {
		accountExists, err := s.repo.AccountIDExists(*item.IncomeAccount)
		if err != nil {
			return nil, nil, err
		}
		if !accountExists {
			return nil, map[string]string{"income_account": "Income account does not exist"}, nil
		}
	}

	if item.COGSAccount != nil {
		accountExists, err := s.repo.AccountIDExists(*item.COGSAccount)
		if err != nil {
			return nil, nil, err
		}
		if !accountExists {
			return nil, map[string]string{"cogs_account": "COGS account does not exist"}, nil
		}
	}

	if item.ExpenseAccount != nil {
		accountExists, err := s.repo.AccountIDExists(*item.ExpenseAccount)
		if err != nil {
			return nil, nil, err
		}
		if !accountExists {
			return nil, map[string]string{"expense_account": "Expense account does not exist"}, nil
		}
	}

	item.ID = id

	// Update in DB
	updatedItem, err := s.repo.Update(item, id)
	if err != nil {
		return nil, nil, err // internal/server error
	}

	return &updatedItem, nil, nil // success
}

