package dvow

import (
    "context"
)

//go:generate mockery --name IOverwritingOps --case underscore --inpkg
// IOverwritingOps ...
type IOverwritingOps interface {
    // ExtractOverwritingStorage returns the Storage currently associated with ctx, or
    // nil if no such Storage could be found.
    ExtractOverwritingStorage(ctx context.Context) Storage
    // GetOverwrittenValue returns the Value of the variable under this name if it was overwritten
    GetOverwrittenValue(ctx context.Context, name string) Value
}

type overwritingOps struct{}

// ExtractOverwritingStorage returns the Storage currently associated with ctx, or
// nil if no such Storage could be found.
func (overwritingOps) ExtractOverwritingStorage(ctx context.Context) Storage {
    return ExtractOverwritingStorage(ctx)
}

// GetOverwrittenValue returns the Value of the variable under this name if it was overwritten
func (overwritingOps) GetOverwrittenValue(ctx context.Context, name string) Value {
    return GetOverwrittenValue(ctx, name)
}

// Ops provides a wrapper around all overwriting-related functions provided by the library.
// It can be mocked to help write tests more fluently.
var Ops IOverwritingOps = overwritingOps{}

// MockOps can be used in tests to perform monkey-patching on Ops
func MockOps() (*MockIOverwritingOps, func()) {
   old := Ops
   mock := &MockIOverwritingOps{}

   Ops = mock
   return mock, func() {
       Ops = old
   }
}
