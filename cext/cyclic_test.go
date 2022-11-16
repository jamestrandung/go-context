package cext

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestWithAcyclicBreadcrumb(t *testing.T) {
	// New breadcrumb with ID as 1
	ctxWithAcyclicBreadcrumb, ok := WithAcyclicBreadcrumb(context.Background(), 1)
	assert.NotNil(t, ctxWithAcyclicBreadcrumb)
	assert.True(t, ok)

	// New breadcrumb with ID as 2
	ctxWithAcyclicBreadcrumb, ok = WithAcyclicBreadcrumb(ctxWithAcyclicBreadcrumb, 2)
	assert.NotNil(t, ctxWithAcyclicBreadcrumb)
	assert.True(t, ok)

	// Old breadcrumb with ID as 1
	ctxWithBadBreadcrumb, ok := WithAcyclicBreadcrumb(ctxWithAcyclicBreadcrumb, 1)
	assert.Nil(t, ctxWithBadBreadcrumb)
	assert.False(t, ok)

	// New breadcrumb with ID as "a"
	ctxWithAcyclicBreadcrumb, ok = WithAcyclicBreadcrumb(ctxWithAcyclicBreadcrumb, "a")
	assert.NotNil(t, ctxWithAcyclicBreadcrumb)
	assert.True(t, ok)

	// Old breadcrumb with ID as 1
	ctxWithBadBreadcrumb, ok = WithAcyclicBreadcrumb(ctxWithAcyclicBreadcrumb, 1)
	assert.Nil(t, ctxWithBadBreadcrumb)
	assert.False(t, ok)
}
