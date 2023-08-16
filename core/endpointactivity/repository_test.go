package endpointactivity

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/kotalco/cloud-api/pkg/sqlclient"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
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

func TestRepository_GetEndpointById(t *testing.T) {
	t.Run("GetByEndpointId_should_pass", func(t *testing.T) {
		record := createActivityRecord(t)
		result, restErr := repo.WithoutTransaction().GetByEndpointId(record.EndpointId)
		assert.Nil(t, restErr)
		assert.EqualValues(t, result.ID, record.ID)
		cleanUp(record)
	})
	t.Run("GetByEndpointId_should_throw_if_record_doesn't_exist", func(t *testing.T) {
		record, restErr := repo.WithoutTransaction().GetByEndpointId("")
		assert.Nil(t, record)
		assert.EqualValues(t, fmt.Sprintf("can't find activity with endpointId  %s", ""), restErr.Error())
		assert.EqualValues(t, http.StatusNotFound, restErr.StatusCode())
	})
}

func TestRepository_Increment(t *testing.T) {
	t.Run("Update_should_pass_and_create_new_record", func(t *testing.T) {
		record := createActivityRecord(t)
		assert.NotNil(t, record)
		cleanUp(record)
	})
	t.Run("Update_should_pass_and_increment_counter", func(t *testing.T) {
		record := createActivityRecord(t)
		record.Counter++
		restErr := repo.WithoutTransaction().Update(&record)
		updated, err := repo.WithoutTransaction().GetByEndpointId(record.EndpointId)
		assert.Nil(t, err)
		assert.Nil(t, restErr)
		assert.EqualValues(t, 2, updated.Counter)
		cleanUp(*updated)
	})
	t.Run("update_should_throw_if_db_throws", func(t *testing.T) {
		record := new(Activity)
		restErr := repo.WithoutTransaction().Update(record)
		assert.EqualValues(t, "something went wrong", restErr.Error())
	})
}

func createActivityRecord(t *testing.T) Activity {
	record := new(Activity)
	record.ID = uuid.NewString()
	record.EndpointId = uuid.NewString()
	record.Counter = 1
	restErr := repo.WithoutTransaction().Update(record)
	assert.Nil(t, restErr)
	return *record
}
