package dvow

//go:generate mockery --name Storage --case underscore --inpkg
// Storage is the container of all overwritten variables
type Storage interface {
    // Get returns the Value of the variable under this name if it was overwritten
    Get(name string) Value
}

type dynamicOverwritingStorage struct {
    parent Storage // from parent context.Context
    variables map[string]interface{}
}

// Get returns the Value of the variable under this name if it was overwritten
func (s dynamicOverwritingStorage) Get(name string) Value {
    if value, isPresent := s.variables[name]; isPresent {
        return overwriteValue{
            value: value,
        }
    }

    if s.parent != nil {
        return s.parent.Get(name)
    }

    return nil
}

