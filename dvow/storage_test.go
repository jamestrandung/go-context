package dvow

import (
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "testing"
)

func TestDynamicOverwritingStorage_Get(t *testing.T) {
    valueMock := &MockValue{}

    storageMock := &MockStorage{}
    storageMock.On("Get", "existing_value_in_parent").Return(valueMock).Once()
    storageMock.On("Get", "non_existing_value").Return(nil).Once()

    variables := make(map[string]interface{})
    variables["existing_value"] = "value"

    storage := dynamicOverwritingStorage{
        parent:    storageMock,
        variables: variables,
    }

    value1 := storage.Get("existing_value")

    assert.NotNil(t, value1)
    assert.Equal(t, "value", value1.AsIs())

    value2 := storage.Get("existing_value_in_parent")

    assert.NotNil(t, value2)
    assert.Equal(t, valueMock, value2)

    value3 := storage.Get("non_existing_value")

    assert.Nil(t, value3)

    mock.AssertExpectationsForObjects(t, storageMock)
}