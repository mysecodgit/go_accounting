package user

import "fmt"

type UserService struct {
	repo UserRepository
}

func NewUserService(repo UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) Register(user RegisterUserRequest) (*UserResponse, map[string]string, error) {

	// Field validation
	if errs := user.Validate(); errs != nil {
		return nil, errs, nil // validation errors
	}

	// DTO -> Model
	mappedUser := user.ToUser()

	// Save to DB
	createdUser, err := s.repo.Create(mappedUser)
	if err != nil {
		return nil, nil, err // internal/server error
	}

	// Model -> DTO response
	response := createdUser.ToUserResponse()

	return &response, nil, nil // success
}

func (s *UserService) GetAllUsers() ([]UserResponse, error) {
	users, err := s.repo.GetAll()

	if err != nil {
		return nil, err
	}

	mappedUsers := []UserResponse{}

	for _, user := range users {
		u := user.ToUserResponse()
		mappedUsers = append(mappedUsers, u)
	}

	return mappedUsers,nil
}

func (s *UserService) GetUserByID(id int) (*UserResponse, error) {
	user, err := s.repo.GetByID(id)
	fmt.Println("The user is ", user)
	if err != nil {
		return nil, err
	}

	response := user.ToUserResponse()
	return &response, nil
}
