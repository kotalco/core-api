package setting

import (
	"github.com/kotalco/cloud-api/pkg/sqlclient"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	repo = NewRepository()
)

func init() {
	err := sqlclient.OpenDBConnection().AutoMigrate(new(Setting))
	if err != nil {
		panic(err.Error())
	}
}

func cleanUp(key string) {
	sqlclient.DbClient.Where("key = ?", key).Delete(&Setting{})
}

func TestRepository_GetAndSet(t *testing.T) {
	t.Run("set key store should pass", func(t *testing.T) {
		err := repo.Set("testKey", "testValue")
		assert.Nil(t, err)

		value, err := repo.Get("testKey")
		assert.Nil(t, err)
		assert.EqualValues(t, "testValue", value)
		cleanUp("testKey")
	})
}
