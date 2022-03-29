package verification

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/kotalco/api/pkg/logger"
	"github.com/kotalco/cloud-api/pkg/config"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"net/http"
	"sync"
	"testing"
)

var (
	testingClientOnce sync.Once
	DbTestingClient   *gorm.DB
	repo              IRepository
)

func OpenTestingDBConnection() *gorm.DB {
	testingClientOnce.Do(func() {
		db, err := gorm.Open(postgres.Open(config.EnvironmentConf["DB_TESTING_SERVER_URL"]), &gorm.Config{})
		if err != nil {
			go logger.Panic("TESTING_DATABASE_CONNECTION_ERROR", err)
			panic(err)
		}
		DbTestingClient = db
	})
	return DbTestingClient
}

func setupTest(t *testing.T) func(t *testing.T) {
	repo = NewRepository()

	dbClient = OpenTestingDBConnection()
	err := dbClient.AutoMigrate(Verification{})
	if err != nil {
		panic(err.Error())
	}

	return func(t *testing.T) {
		dbClient = OpenTestingDBConnection()
		dbClient.Exec("TRUNCATE TABLE users;")
	}
}

func TestRepository_Create(t *testing.T) {
	cleanUp := setupTest(t)
	t.Run("Create_Should_Pass", func(t *testing.T) {
		verification := createVerification(t)
		assert.EqualValues(t, false, verification.Completed)
		cleanUp(t)
	})

	t.Run("Create_Should_Throw_If_Already_Exits", func(t *testing.T) {
		verification := createVerification(t)
		restErr := repo.Create(verification)
		assert.EqualValues(t, "verification already exits", restErr.Message)
	})
}

func TestRepository_GetByUserId(t *testing.T) {
	cleanUp := setupTest(t)
	t.Run("Get_Use_By_Id_Should_Pass", func(t *testing.T) {
		verification := createVerification(t)
		result, restErr := repo.GetByUserId(verification.UserId)
		assert.Nil(t, restErr)
		assert.EqualValues(t, verification.ID, result.ID)
	})
	t.Run("Get_User_By_Id_Should_Throw_If_Verification_With_User_Id_Does't_Exit", func(t *testing.T) {
		verification, restErr := repo.GetByUserId("")
		fmt.Println(verification, restErr)
		assert.Nil(t, verification)
		assert.EqualValues(t, fmt.Sprintf("can't find verification with userId  %s", ""), restErr.Message)
		assert.EqualValues(t, http.StatusNotFound, restErr.Status)
	})
	cleanUp(t)
}

func TestRepository_Update(t *testing.T) {
	cleanUp := setupTest(t)
	t.Run("Update_Should_Pass", func(t *testing.T) {
		verification := createVerification(t)
		verification.Completed = true

		restErr := repo.Update(verification)

		assert.Nil(t, restErr)
		cleanUp(t)
	})
}

func createVerification(t *testing.T) *Verification {
	verification := new(Verification)
	verification.ID = uuid.New().String()
	verification.UserId = uuid.New().String()
	restErr := repo.Create(verification)
	assert.Nil(t, restErr)
	return verification
}
