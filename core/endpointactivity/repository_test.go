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
	t.Run("create_should_pass_and_create_new_record", func(t *testing.T) {
		record := createActivityRecord(t)
		assert.NotNil(t, record)
		cleanUp(record)
	})
	t.Run("create_should_throw_if_db_throws", func(t *testing.T) {
		record := new(Activity)
		restErr := repo.WithoutTransaction().CreateInBatches([]*Activity{
			record,
		})
		assert.EqualValues(t, "something went wrong", restErr.Error())
	})
}

func TestRepository_RawQuery(t *testing.T) {
	t.Run("raw_query_should_pass", func(t *testing.T) {
		record := createActivityRecord(t)
		assert.NotNil(t, record)
		type dailyAggregation struct {
			Day   int
			Count int
		}
		now := time.Now()
		currentYear, currentMonth, _ := now.Date()
		location := now.Location()
		firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, location)
		lastOfMonth := firstOfMonth.AddDate(0, 1, -1)

		dest := new([]dailyAggregation)
		restErr := repo.WithoutTransaction().RawQuery(rawDailyStatsQuery, dest, record.EndpointId, firstOfMonth, lastOfMonth)
		assert.Nil(t, restErr)
		cleanUp(record)
	})
}

func createActivityRecord(t *testing.T) Activity {
	record := new(Activity)
	record.ID = uuid.NewString()
	record.EndpointId = uuid.NewString()
	record.Timestamp = time.Now()
	activities := []*Activity{
		record,
	}
	restErr := repo.WithoutTransaction().CreateInBatches(activities)
	assert.Nil(t, restErr)
	return *record
}
