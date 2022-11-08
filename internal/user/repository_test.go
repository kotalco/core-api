package user

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/kotalco/cloud-api/pkg/security"
	"github.com/kotalco/cloud-api/pkg/sqlclient"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

var (
	repo = NewRepository()
)

func init() {
	err := sqlclient.OpenDBConnection().AutoMigrate(new(User))
	if err != nil {
		panic(err.Error())
	}
}

func cleanUp(user User) {
	sqlclient.DbClient.Delete(user)
}

func TestRepository_Create(t *testing.T) {
	t.Run("Create_Should_Pass", func(t *testing.T) {
		user := createUser(t)
		assert.NotNil(t, user)
		cleanUp(user)
	})

	t.Run("Create_Should_Throw_If_Email_Already_Exits", func(t *testing.T) {
		user := createUser(t)

		user.ID = uuid.New().String()
		restErr := repo.Create(&user)

		assert.EqualValues(t, "email already exits", restErr.Message)
		assert.EqualValues(t, http.StatusConflict, restErr.Status)

		cleanUp(user)
	})
}

func TestRepository_GetByEmail(t *testing.T) {
	t.Run("Get_By_Email_Should_Pass", func(t *testing.T) {
		user := createUser(t)

		result, restErr := repo.GetByEmail(user.Email)

		assert.Nil(t, restErr)
		fmt.Println(result, user)
		assert.EqualValues(t, user.Email, result.Email)

		cleanUp(user)
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

		cleanUp(user)
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

		restErr := repo.Update(&user)
		assert.Nil(t, restErr)
		cleanUp(user)
	})
}

func TestRepository_FindWhereIdInSlice(t *testing.T) {
	t.Run("find-where-in-array-should-pass", func(t *testing.T) {
		user1 := createUser(t)
		user2 := createUser(t)
		list := []*User{&user1, &user2}

		users, err := repo.FindWhereIdInSlice([]string{user1.ID, user2.ID})
		assert.Nil(t, err)
		assert.ElementsMatch(t, list, users)
	})
}

func TestRepository_Count(t *testing.T) {
	t.Run("count users should pass", func(t *testing.T) {
		user1 := createUser(t)
		count, err := repo.Count()
		assert.Nil(t, err)
		assert.NotEqual(t, 0, count)
		cleanUp(user1)
	})
}

func createUser(t *testing.T) User {
	user := new(User)
	user.ID = uuid.New().String()
	user.Email = security.GenerateRandomString(10) + "@test.com"
	restErr := repo.Create(user)
	assert.Nil(t, restErr)
	return *user
}
