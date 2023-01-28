package cext

import (
	"context"
	"github.com/jamestrandung/go-context/helper"
)

type contextKey struct{}

var breadcrumbKey = contextKey{}

// WithAcyclicBreadcrumb return a new context with the given breadcrumbID embedded inside and true
// if this ID has never been encountered in the execution path before. Otherwise, it returns a nil
// context.Context and false to indicate the execution is running in circle.
//
// Note: the provided breadcrumbID must be comparable and should not be of type string or any other
// built-in type to avoid collisions between packages using this context. You should define your
// own types for breadcrumbID similar to the best practices for using context.WithValue.
func WithAcyclicBreadcrumb[V comparable](ctx context.Context, breadcrumbID V) (context.Context, bool) {
	prevBreadcrumb := findPrevBreadcrumb(ctx, breadcrumbID)

	newBreadcrumb, ok := appendBreadcrumb(ctx, breadcrumbID, prevBreadcrumb)
	if !ok {
		return nil, false
	}

	return context.WithValue(ctx, breadcrumbKey, newBreadcrumb), true
}

type breadcrumb struct {
	parentCtx context.Context
	id        interface{}
	prev      *breadcrumb
}

// findPrevBreadcrumb returns the previous breadcrumb having ID with the same underlying type as
// the given breadcrumbID or nil if such breadcrumb does not exist.
func findPrevBreadcrumb[V comparable](ctx context.Context, breadcrumbID V) *breadcrumb {
	bc, ok := ctx.Value(breadcrumbKey).(*breadcrumb)
	if !ok {
		return nil
	}

	if helper.IsCastable[V](bc.id) {
		return bc
	}

	return findPrevBreadcrumb(bc.parentCtx, breadcrumbID)
}

// appendBreadcrumb returns a new breadcrumb appended to the end of the existing breadcrumb chain
// and true if no breadcrumb having the same ID exists in the chain. Otherwise, it returns nil and
// false, indicating the execution is running in circle.
func appendBreadcrumb[V comparable](ctx context.Context, breadcrumbID V, prev *breadcrumb) (*breadcrumb, bool) {
	cur := prev
	for cur != nil {
		if cur.id == breadcrumbID {
			return nil, false
		}

		cur = cur.prev
	}

	return &breadcrumb{
		parentCtx: ctx,
		id:        breadcrumbID,
		prev:      prev,
	}, true
}
