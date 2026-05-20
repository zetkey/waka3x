package services

import (
	"errors"
	"testing"

	"github.com/zetkey/waka3x/mocks"
	"github.com/zetkey/waka3x/models"
	"github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/suite"
)

const (
	TestUserID = "muety"
	TestAPIKey = "full-access-key-from-user-model"
)

type UserServiceTestSuite struct {
	suite.Suite
	TestUser        *models.User
	KeyValueService *mocks.KeyValueServiceMock
	MailService     *mocks.MailServiceMock
	ApiKeyService   *mocks.MockApiKeyService
	UserRepo        *mocks.UserRepositoryMock
}

func (suite *UserServiceTestSuite) SetupSuite() {
	suite.TestUser = &models.User{ID: TestUserID, ApiKey: TestAPIKey}
}

func (suite *UserServiceTestSuite) BeforeTest(suiteName, testName string) {
	suite.KeyValueService = new(mocks.KeyValueServiceMock)
	suite.MailService = new(mocks.MailServiceMock)
	suite.ApiKeyService = new(mocks.MockApiKeyService)
	suite.UserRepo = new(mocks.UserRepositoryMock)
}

func TestUserServiceTestSuite(t *testing.T) {
	suite.Run(t, new(UserServiceTestSuite))
}

func (suite *UserServiceTestSuite) TestUserService_GetByEmail_Empty() {
	sut := NewUserService(suite.KeyValueService, suite.MailService, suite.ApiKeyService, suite.UserRepo)

	result, err := sut.GetUserByEmail("")

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(err, errors.New("email must not be empty"))
}

func (suite *UserServiceTestSuite) TestUserService_GetByEmail_Invalid() {
	sut := NewUserService(suite.KeyValueService, suite.MailService, suite.ApiKeyService, suite.UserRepo)

	result, err := sut.GetUserByEmail("notanemailaddress")

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(err, errors.New("not a valid email"))
}

func (suite *UserServiceTestSuite) TestUserService_GetByEmail_Valid() {
	const testEmail = "foo@bar.com"

	suite.UserRepo.On("FindOne", models.User{Email: testEmail}).Return(suite.TestUser, nil)

	sut := NewUserService(suite.KeyValueService, suite.MailService, suite.ApiKeyService, suite.UserRepo)
	result, err := sut.GetUserByEmail(testEmail)

	suite.Equal(suite.TestUser, result)
	suite.Nil(err)
}

func (suite *UserServiceTestSuite) TestUserService_GetByEmptyKey_Failed() {
	sut := NewUserService(suite.KeyValueService, suite.MailService, suite.ApiKeyService, suite.UserRepo)

	result, err := sut.GetUserByKey("", false)

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(err, errors.New("key must not be empty"))
}

func (suite *UserServiceTestSuite) TestUserService_GetByKeyFromCache_Success() {
	userCached := &models.User{ID: TestUserID, ApiKey: "cached-key"}

	userCache := cache.New(cache.NoExpiration, cache.NoExpiration)
	userCache.SetDefault(TestAPIKey, userCached)

	sut := &UserService{cache: userCache}

	result, err := sut.GetUserByKey(TestAPIKey, false)
	suite.Nil(err)
	suite.NotNil(result)
	suite.Equal(1, userCache.ItemCount())
	suite.Equal(userCached, result)
	suite.Equal(userCached.ApiKey, result.ApiKey)
}

func (suite *UserServiceTestSuite) TestUserService_GetByKeyFromUserModel_Success() {
	sut := NewUserService(suite.KeyValueService, suite.MailService, suite.ApiKeyService, suite.UserRepo)

	suite.UserRepo.On("FindOne", models.User{ApiKey: TestAPIKey}).Return(suite.TestUser, nil)

	result, err := sut.GetUserByKey(TestAPIKey, false)
	suite.Nil(err)
	suite.NotNil(result)
	suite.Equal(suite.TestUser, result)
	suite.Equal(suite.TestUser.ApiKey, result.ApiKey)
}

func (suite *UserServiceTestSuite) TestUserService_GetByKeyFromAdditionalApiKeys_Success() {
	sut := NewUserService(suite.KeyValueService, suite.MailService, suite.ApiKeyService, suite.UserRepo)

	suite.UserRepo.On("FindOne", models.User{ApiKey: TestAPIKey}).Return(nil, errors.New("not found"))
	suite.ApiKeyService.On("GetByApiKey", TestAPIKey, true).Return(&models.ApiKey{User: suite.TestUser}, nil)

	result, err := sut.GetUserByKey(TestAPIKey, true)
	suite.Nil(err)
	suite.NotNil(result)
	suite.Equal(suite.TestUser, result)
	suite.Equal(suite.TestUser.ApiKey, result.ApiKey)
}

func (suite *UserServiceTestSuite) TestUserService_GetByKeyFromAdditionalApiKeys_Failed() {
	sut := NewUserService(suite.KeyValueService, suite.MailService, suite.ApiKeyService, suite.UserRepo)

	suite.UserRepo.On("FindOne", models.User{ApiKey: TestAPIKey}).Return(nil, errors.New("not found"))
	suite.ApiKeyService.On("GetByApiKey", TestAPIKey, true).Return(nil, errors.New("not found"))

	result, err := sut.GetUserByKey(TestAPIKey, true)
	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(err, errors.New("not found"))
}

func (suite *UserServiceTestSuite) TestUserService_GetUserById_Empty() {
	sut := NewUserService(suite.KeyValueService, suite.MailService, suite.ApiKeyService, suite.UserRepo)

	result, err := sut.GetUserById("")

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(err, errors.New("user id must not be empty"))
}

func (suite *UserServiceTestSuite) TestUserService_GetUserById_FromCache() {
	userCached := &models.User{ID: TestUserID, ApiKey: "cached-key"}

	userCache := cache.New(cache.NoExpiration, cache.NoExpiration)
	userCache.SetDefault(TestUserID, userCached)

	sut := &UserService{cache: userCache}

	result, err := sut.GetUserById(TestUserID)
	suite.Nil(err)
	suite.NotNil(result)
	suite.Equal(userCached, result)
}

func (suite *UserServiceTestSuite) TestUserService_GetUserById_FromRepo() {
	sut := NewUserService(suite.KeyValueService, suite.MailService, suite.ApiKeyService, suite.UserRepo)

	suite.UserRepo.On("FindOne", models.User{ID: TestUserID}).Return(suite.TestUser, nil)

	result, err := sut.GetUserById(TestUserID)
	suite.Nil(err)
	suite.NotNil(result)
	suite.Equal(suite.TestUser, result)
}

func (suite *UserServiceTestSuite) TestUserService_GetUserByWebAuthnID_Empty() {
	sut := NewUserService(suite.KeyValueService, suite.MailService, suite.ApiKeyService, suite.UserRepo)

	result, err := sut.GetUserByWebAuthnID("")

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(err, errors.New("webauthn id must not be empty"))
}

func (suite *UserServiceTestSuite) TestUserService_GetUserByWebAuthnID_Valid() {
	const testWebAuthnID = "test-webauthn-id"

	suite.UserRepo.On("FindOne", models.User{WebauthnID: testWebAuthnID}).Return(suite.TestUser, nil)

	sut := NewUserService(suite.KeyValueService, suite.MailService, suite.ApiKeyService, suite.UserRepo)
	result, err := sut.GetUserByWebAuthnID(testWebAuthnID)

	suite.Equal(suite.TestUser, result)
	suite.Nil(err)
}

func (suite *UserServiceTestSuite) TestUserService_GetUserByResetToken_Empty() {
	sut := NewUserService(suite.KeyValueService, suite.MailService, suite.ApiKeyService, suite.UserRepo)

	result, err := sut.GetUserByResetToken("")

	suite.Nil(result)
	suite.NotNil(err)
	suite.Equal(err, errors.New("reset token must not be empty"))
}

func (suite *UserServiceTestSuite) TestUserService_GetUserByResetToken_Valid() {
	const testResetToken = "test-reset-token"

	suite.UserRepo.On("FindOne", models.User{ResetToken: testResetToken}).Return(suite.TestUser, nil)

	sut := NewUserService(suite.KeyValueService, suite.MailService, suite.ApiKeyService, suite.UserRepo)
	result, err := sut.GetUserByResetToken(testResetToken)

	suite.Equal(suite.TestUser, result)
	suite.Nil(err)
}

func (suite *UserServiceTestSuite) TestUserService_CreateOrGet_NewUser() {
	signup := &models.Signup{
		Username: "newuser",
		Email:    "new@example.com",
		Password: "password123",
		Location: "US/Pacific",
	}

	suite.UserRepo.On("InsertOrGet", &models.User{
		ID:       signup.Username,
		Email:    signup.Email,
		Location: signup.Location,
		Password: signup.Password,
	}).Return(suite.TestUser, true, nil)

	sut := NewUserService(suite.KeyValueService, suite.MailService, suite.ApiKeyService, suite.UserRepo)
	result, created, err := sut.CreateOrGet(signup, false)

	suite.Nil(err)
	suite.NotNil(result)
	suite.True(created)
}

func (suite *UserServiceTestSuite) TestUserService_Update_Success() {
	updatedUser := &models.User{ID: TestUserID, Email: "updated@example.com"}

	suite.UserRepo.On("Update", suite.TestUser).Return(updatedUser, nil)

	sut := NewUserService(suite.KeyValueService, suite.MailService, suite.ApiKeyService, suite.UserRepo)
	result, err := sut.Update(suite.TestUser)

	suite.Nil(err)
	suite.NotNil(result)
	suite.Equal(updatedUser, result)
}

func (suite *UserServiceTestSuite) TestUserService_ResetApiKey_Success() {
	updatedUser := &models.User{ID: TestUserID, ApiKey: "new-api-key"}

	suite.UserRepo.On("Update", suite.TestUser).Return(updatedUser, nil)

	sut := NewUserService(suite.KeyValueService, suite.MailService, suite.ApiKeyService, suite.UserRepo)
	result, err := sut.ResetApiKey(suite.TestUser)

	suite.Nil(err)
	suite.NotNil(result)
	suite.NotEqual(TestAPIKey, result.ApiKey)
}
