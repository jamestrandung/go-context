package dvow

import (
    "context"
)

type contextKey struct{}

var overwritingStorageKey = contextKey{}

// WithOverwrittenVariables returns a new context.Context that holds a reference to
// the given overwritten variables.
//
// Note: This implementation creates a shallow copy of the input map to hold a reference
// to all of its key-value pairs. The expectation is that clients must only use the lib
// to **get** overwritten values from the context. If the value clients receive is a
// pointer or a complex struct containing some pointers or a pointer-like object such as
// an array or a map, they should NOT update this value since the context is most likely
// passed into many go-routines running in parallel. As a consequence, clients may run into
// a race condition if things goes wrong.
func WithOverwrittenVariables(ctx context.Context, overwrittenVariables map[string]interface{}) context.Context {
    if len(overwrittenVariables) == 0 {
        return ctx
    }

    // Make a copy so that our storage wouldn't be affected by changes to the input map
    clone := make(map[string]interface{}, len(overwrittenVariables))
    for name, value := range overwrittenVariables {
        clone[name] = value
    }

    derivedStorage := dynamicOverwritingStorage{
        parent: Ops.ExtractOverwritingStorage(ctx),
        variables: clone,
    }

    return context.WithValue(ctx, overwritingStorageKey, derivedStorage)
}

// ExtractOverwritingStorage returns the Storage currently associated with ctx, or
// nil if no such Storage could be found.
func ExtractOverwritingStorage(ctx context.Context) Storage {
    val := ctx.Value(overwritingStorageKey)
    if s, ok := val.(Storage); ok {
        return s
    }

    return nil
}

// GetOverwrittenValue returns the Value of the variable under this name if it was overwritten
func GetOverwrittenValue(ctx context.Context, name string) Value {
    storage := Ops.ExtractOverwritingStorage(ctx)
    if storage == nil {
        return nil
    }

    return storage.Get(name)
}
