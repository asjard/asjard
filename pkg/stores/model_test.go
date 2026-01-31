package stores

import (
	"context"
	"testing"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/pkg/tools"
	"github.com/stretchr/testify/assert"
)

type TestModel struct {
	*Model
	*TestTable
}

func (model *TestModel) Get() (*TestTable, error) {
	return model.TestTable.Get()
}

func (model *TestModel) Update() (*TestTable, error) {
	return model.TestTable.Update()
}

func (model *TestModel) Del() error {
	return model.TestTable.Del()
}

func (model *TestModel) Search() ([]*TestTable, error) {
	return model.TestTable.Search()
}

// 测试表
type TestTable struct {
	Id   int64
	Name string
	Age  int32
}

func (TestTable) TableName() string {
	return "test_model_table"
}

func (TestTable) Database() string {
	return "example_database"
}

// 测试model
func (table *TestTable) ModelName() string {
	return table.Database() + "_" + table.TableName()
}

func (TestTable) Get() (*TestTable, error) {
	return &TestTable{}, nil
}

func (TestTable) Update() (*TestTable, error) {
	return &TestTable{}, nil
}

func (TestTable) Del() error {
	return nil
}

func (TestTable) Search() ([]*TestTable, error) {
	return []*TestTable{}, nil
}

func TestMain(m *testing.M) {
	if err := config.Load(-1); err != nil {
		panic(err)
	}
	tools.DefaultTW.Start()
	m.Run()
}

func TestGetData(t *testing.T) {
	model := &TestModel{}

	from := "test_from"

	getFunc := func() (any, error) {
		return &from, nil
	}
	t.Run("OutNotNil", func(t *testing.T) {
		assert.NotNil(t, model.GetData(context.Background(), nil, nil, getFunc))
	})
	t.Run("CacheNil", func(t *testing.T) {
		var to string
		assert.Nil(t, model.GetData(context.Background(), &to, nil, getFunc))
		assert.Equal(t, from, to)
	})
}

func TestSetData(t *testing.T) {
	model := &TestModel{}
	setFunc := func() error {
		return nil
	}
	t.Run("CacheNil", func(t *testing.T) {
		assert.Nil(t, model.SetData(context.Background(), setFunc, nil))
	})
}

func TestCopy(t *testing.T) {
	model := &TestModel{
		Model: &Model{},
	}
	from := "test_from"
	var to string
	t.Run("OutNotPtr", func(t *testing.T) {
		assert.NotNil(t, model.copy(context.Background(), &from, to))
	})
	t.Run("DiffType", func(t *testing.T) {
		assert.NotNil(t, model.copy(context.Background(), from, &to))
	})
	t.Run("Success", func(t *testing.T) {
		var to string
		assert.Nil(t, model.copy(context.Background(), &from, &to))
		assert.Equal(t, from, to)
	})
}
