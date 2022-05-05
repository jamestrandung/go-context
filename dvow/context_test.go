package dvow

import (
    "context"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "testing"
)

func TestWithOverwrittenVariables(t *testing.T) {
    opsMock, cleanup := MockOps()
    defer cleanup()

    scenarios := []struct {
        desc string
        test func(t *testing.T)
    }{
        {
            desc: "input map is empty/nil",
            test: func(t *testing.T) {
                ctx := context.Background()

                var overwrittenVariables map[string]interface{}

                actual := WithOverwrittenVariables(ctx, overwrittenVariables)

                assert.Equal(t, ctx, actual)
                assert.Nil(t, actual.Value(overwritingStorageKey))
            },
        },
        {
            desc: "input map contains some values",
            test: func(t *testing.T) {
                ctx := context.Background()

                overwrittenVariables := make(map[string]interface{})
                overwrittenVariables["test"] = "random"
                overwrittenVariables["test2"] = nil

                storageMock := &MockStorage{}
                opsMock.On("ExtractOverwritingStorage", ctx).Return(storageMock).Once()

                actual := WithOverwrittenVariables(ctx, overwrittenVariables)

                expectedStorage := dynamicOverwritingStorage{
                    parent: storageMock,
                    variables: overwrittenVariables,
                }

                assert.Equal(t, expectedStorage, actual.Value(overwritingStorageKey))
                mock.AssertExpectationsForObjects(t, opsMock)

                overwrittenVariables["new_key"] = "random"
                assert.NotEqual(t, expectedStorage, actual.Value(overwritingStorageKey), "changes to the input map must not affect our Storage")
            },
        },
    }

    for _, scenario := range scenarios {
        sc := scenario
        t.Run(sc.desc, func(t *testing.T) {
            sc.test(t)
        })
    }
}

func TestExtractOverwritingStorage(t *testing.T) {
    scenarios := []struct {
        desc string
        test func(t *testing.T)
    }{
        {
            desc: "ctx does not contain a Storage",
            test: func(t *testing.T) {
                ctx := context.Background()

                actual := ExtractOverwritingStorage(ctx)

                assert.Nil(t, actual)
            },
        },
        {
            desc: "value under overwritingStorageKey is not a Storage",
            test: func(t *testing.T) {
                ctx := context.WithValue(context.Background(), overwritingStorageKey, "dummy")

                actual := ExtractOverwritingStorage(ctx)

                assert.Nil(t, actual)
            },
        },
        {
            desc: "value under overwritingStorageKey is a Storage",
            test: func(t *testing.T) {
                storageMock := &MockStorage{}

                ctx := context.WithValue(context.Background(), overwritingStorageKey, storageMock)

                actual := ExtractOverwritingStorage(ctx)

                assert.Equal(t, storageMock, actual)
            },
        },
    }

    for _, scenario := range scenarios {
        sc := scenario
        t.Run(sc.desc, func(t *testing.T) {
            sc.test(t)
        })
    }
}

func TestGetOverwrittenValue(t *testing.T) {
    opsMock, cleanup := MockOps()
    defer cleanup()

    scenarios := []struct {
        desc string
        test func(t *testing.T)
    }{
        {
            desc: "ctx does NOT contain a Storage",
            test: func(t *testing.T) {
                ctx := context.Background()

                opsMock.On("ExtractOverwritingStorage", ctx).Return(nil).Once()

                actual := GetOverwrittenValue(ctx, "name")

                assert.Nil(t, actual)
                mock.AssertExpectationsForObjects(t, opsMock)
            },
        },
        {
            desc: "ctx contains a Storage",
            test: func(t *testing.T) {
                ctx := context.Background()
                storageMock := &MockStorage{}
                valueMock := &MockValue{}
                varName := "name"

                opsMock.On("ExtractOverwritingStorage", ctx).Return(storageMock).Once()
                storageMock.On("Get", varName).Return(valueMock).Once()

                actual := GetOverwrittenValue(ctx, varName)

                assert.Equal(t, valueMock, actual)
                mock.AssertExpectationsForObjects(t, opsMock, storageMock)
            },
        },
    }

    for _, scenario := range scenarios {
        sc := scenario
        t.Run(sc.desc, func(t *testing.T) {
            sc.test(t)
        })
    }
}