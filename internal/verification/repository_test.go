package verification

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/kotalco/cloud-api/pkg/sqlclient"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

var (
	repo = NewRepository()
)

func init() {
	err := sqlclient.OpenDBConnection().AutoMigrate(new(Verification))
	if err != nil {
		panic(err.Error())
	}
}

func cleanUp(verification Verification) {
	sqlclient.OpenDBConnection().Delete(verification)
}

func TestRepository_Create(t *testing.T) {
	t.Run("Create_Should_Pass", func(t *testing.T) {
		verification := createVerification(t)
		assert.EqualValues(t, false, verification.Completed)
		cleanUp(verification)
	})

	t.Run("Create_Should_Throw_If_Already_Exits", func(t *testing.T) {
		verification := createVerification(t)
		restErr := repo.WithoutTransaction().Create(&verification)
		assert.EqualValues(t, "verification already exits", restErr.Message)
	})
}

func TestRepository_GetByUserId(t *testing.T) {
	t.Run("Get_Use_By_Id_Should_Pass", func(t *testing.T) {
		verification := createVerification(t)
		result, restErr := repo.WithoutTransaction().GetByUserId(verification.UserId)
		assert.Nil(t, restErr)
		assert.EqualValues(t, verification.ID, result.ID)
		cleanUp(verification)
	})
	t.Run("Get_User_By_Id_Should_Throw_If_Verification_With_User_Id_Does't_Exit", func(t *testing.T) {
		verification, restErr := repo.WithoutTransaction().GetByUserId("")
		fmt.Println(verification, restErr)
		assert.Nil(t, verification)
		assert.EqualValues(t, fmt.Sprintf("can't find verification with userId  %s", ""), restErr.Message)
		assert.EqualValues(t, http.StatusNotFound, restErr.Status)
	})

}

func TestRepository_Update(t *testing.T) {
	t.Run("Update_Should_Pass", func(t *testing.T) {
		verification := createVerification(t)
		verification.Completed = true

		restErr := repo.WithoutTransaction().Update(&verification)

		assert.Nil(t, restErr)
		cleanUp(verification)
	})
}

func createVerification(t *testing.T) Verification {
	verification := new(Verification)
	verification.ID = uuid.New().String()
	verification.UserId = uuid.New().String()
	restErr := repo.WithoutTransaction().Create(verification)
	assert.Nil(t, restErr)
	return *verification
}
