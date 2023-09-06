package endpointactivity

import (
	"github.com/google/uuid"
	"github.com/kotalco/cloud-api/pkg/sqlclient"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var repo = NewRepository()

func init() {
	err := sqlclient.OpenDBConnection().AutoMigrate(new(Activity))
	if err != nil {
		panic(err)
	}
}

func cleanUp(activity Activity) {
	sqlclient.OpenDBConnection().Delete(activity)
}

func TestRepository_Create(t *testing.T) {
	t.Run("Update_should_pass_and_create_new_record", func(t *testing.T) {
		record := createActivityRecord(t)
		assert.NotNil(t, record)
		cleanUp(record)
	})
	t.Run("update_should_throw_if_db_throws", func(t *testing.T) {
		record := new(Activity)
		restErr := repo.WithoutTransaction().Create(record)
		assert.EqualValues(t, "something went wrong", restErr.Error())
	})
}

func TestRepository_FindMany(t *testing.T) {
	t.Run("find many should pass", func(t *testing.T) {
		record := createActivityRecord(t)
		list, err := repo.WithoutTransaction().FindMany(queryGetMonthlyActivity, record.EndpointId)
		assert.Nil(t, err)
		assert.EqualValues(t, 1, len(list))
		cleanUp(record)
	})
	t.Run("find many should return empty list if endpointId don't match", func(t *testing.T) {
		list, err := repo.WithoutTransaction().FindMany(queryGetMonthlyActivity, "")
		assert.Nil(t, err)
		assert.EqualValues(t, 0, len(list))
	})
	t.Run("find many should throw if db throws", func(t *testing.T) {
		_, err := repo.WithoutTransaction().FindMany("", "")
		assert.EqualValues(t, "something went wrong", err.Error())
	})
}

func createActivityRecord(t *testing.T) Activity {
	record := new(Activity)
	record.ID = uuid.NewString()
	record.EndpointId = uuid.NewString()
	record.Timestamp = time.Now().Unix()
	restErr := repo.WithoutTransaction().Create(record)
	assert.Nil(t, restErr)
	return *record
}
