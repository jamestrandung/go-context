package dvow

import "errors"

var (
    // ErrPointerArgumentRequired ...
    ErrPointerArgumentRequired = errors.New("value type should be a pointer to struct")
)