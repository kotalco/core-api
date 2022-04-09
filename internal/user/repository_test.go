package user

import (
	"github.com/google/uuid"
	"github.com/kotalco/cloud-api/pkg/security"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

var (
	repo = NewRepository()
)

func cleanUp(t *testing.T) {
	dbClient.Exec("TRUNCATE TABLE users;")
}

func TestRepository_Create(t *testing.T) {
	t.Run("Create_Should_Pass", func(t *testing.T) {
		user := createUser(t)
		assert.NotNil(t, user)
		cleanUp(t)
	})

	t.Run("Create_Should_Throw_If_Email_Already_Exits", func(t *testing.T) {
		user := createUser(t)

		user.ID = uuid.New().String()
		restErr := repo.Create(user)

		assert.EqualValues(t, "email already exits", restErr.Message)
		assert.EqualValues(t, http.StatusConflict, restErr.Status)

		cleanUp(t)
	})
}

func TestRepository_GetByEmail(t *testing.T) {
	t.Run("Get_By_Email_Should_Pass", func(t *testing.T) {
		user := createUser(t)

		result, restErr := repo.GetByEmail(user.Email)

		assert.Nil(t, restErr)
		assert.EqualValues(t, user.Email, result.Email)

		cleanUp(t)
	})

	t.Run("Get_By_Email_Should_Throw_If_User_With_This_Email_Doesnt'_Exit", func(t *testing.T) {
		result, restErr := repo.GetByEmail("test")
		assert.Nil(t, result)
		assert.EqualValues(t, "can't find user with email  test", restErr.Message)
		assert.EqualValues(t, http.StatusNotFound, restErr.Status)
	})
}

func TestRepository_GetById(t *testing.T) {
	t.Run("Get_By_Id_Should_Pass", func(t *testing.T) {
		user := createUser(t)

		result, restErr := repo.GetById(user.ID)

		assert.Nil(t, restErr)
		assert.EqualValues(t, user.ID, result.ID)

		cleanUp(t)
	})

	t.Run("Get_By_Id_Should_Throw_If_Id_Does't_Exits", func(t *testing.T) {
		result, restErr := repo.GetById("")

		assert.Nil(t, result)
		assert.EqualValues(t, "no such user", restErr.Message)
	})
}

func TestRepository_Update(t *testing.T) {
	t.Run("Update_Should_Pass", func(t *testing.T) {
		user := createUser(t)
		user.Email = security.GenerateRandomString(5) + "@test.com"

		restErr := repo.Update(user)
		assert.Nil(t, restErr)
	})
	//cleanUp(t)
}

func createUser(t *testing.T) *User {
	user := new(User)
	user.ID = uuid.New().String()
	user.Email = security.GenerateRandomString(10) + "@test.com"
	restErr := repo.Create(user)
	assert.Nil(t, restErr)
	return user
}
