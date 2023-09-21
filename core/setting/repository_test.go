package setting

import (
	"fmt"
	"github.com/google/uuid"
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
	sqlclient.OpenDBConnection().Where("key = ?", key).Delete(&Setting{})
}

func TestRepository_Get(t *testing.T) {
	t.Run("get setting should pass", func(t *testing.T) {
		err := repo.WithoutTransaction().Create("testKey", "value")
		returnedValue, err := repo.WithoutTransaction().Get("testKey")
		assert.Nil(t, err)
		assert.EqualValues(t, "value", returnedValue)
		cleanUp("testKey")
	})
	t.Run("get should throw if key not found", func(t *testing.T) {
		key := uuid.NewString()
		_, err := repo.WithoutTransaction().Get(key)
		assert.EqualValues(t, fmt.Sprintf("can't find config for the key  %s", key), err.Error())
	})

}

func TestRepository_Create(t *testing.T) {
	t.Run("create should pass", func(t *testing.T) {
		key := uuid.NewString()
		err := repo.Create(key, "value")
		assert.Nil(t, err)
		cleanUp(key)
	})
	t.Run("create should throw duplicateErr", func(t *testing.T) {
		key := uuid.NewString()
		err := repo.WithoutTransaction().Create(key, "value")
		err2 := repo.WithoutTransaction().Create(key, "value")
		assert.Nil(t, err)
		assert.EqualValues(t, "key already exists", err2.Error())
		cleanUp(key)
	})

}

func TestRepository_Update(t *testing.T) {
	t.Run("update setting should pass", func(t *testing.T) {
		key := uuid.NewString()
		err := repo.WithoutTransaction().Create(key, "value")
		assert.Nil(t, err)
		err = repo.Update(key, "value2")
		assert.Nil(t, err)
	})
}
func TestRepository_Find(t *testing.T) {
	t.Run("find should return list of settings", func(t *testing.T) {
		_, err := repo.WithoutTransaction().Find()
		assert.Nil(t, err)
	})
}
