# Dynamic Value Overwriting

## Why context?

**Dynamic Value Overwriting** is usually an API-level requirement in which we need to dynamically replace the values of 
specific variables in one execution of our code with some values that our clients provide.

In essence, it means these overwritten values are bound to the request scope, making context a suitable storage for
such values on paper. 

Another good reason why we should use context is that dynamic value overwriting requires a cross-cutting implementation 
similar to OpenTracing spans. If you take a step back and think about it, each service is a combination of many code 
modules, each is responsible for one single task (e.g. calculate NSNSFare, calculate PTPFare, fetch coefficients, etc.). 
The variables we need to replace can be inside any of these code modules.

We can create a simple map of overwritten variables and pass it down every single code path to overwrite wherever necessary.
However, this design requires us to add a map parameter to all functions/methods in our code path which is messy. In fact,
this map might not be provided at all as most API calls don't require dynamic overwriting of values. 

On the other hand, passing context down to every lower-level code module is a recommended coding practice so that each 
module can take into account timeout, deadline and handle cross-cutting concerns such as span tracing. Similar to how
we only create a tracing span if necessary in the code path that we need to measure performance, by using context, we 
can choose to retrieve overwritten variables only in those places where we want clients to overwrite dynamically.

## How to use

First, you need to receive a map of overwritten variables from clients via your gRPC or HTTP endpoints. This requires an 
update to the Protobuf definition of your APIs. In addition, you must define a list of overwritable variables including
their <u>fixed & unique names</u> and <u>expected data types</u> so that clients know what to send into your APIs.

Subsequently, when you receive an API call, you should have a `map[string]interface{}` on hand to execute the following 
function.

```go
// WithOverwrittenVariables returns a new context.Context that holds a reference to
// the given overwritten variables.
func WithOverwrittenVariables(ctx context.Context, overwrittenVariables map[string]interface{}) context.Context
```

After getting back a context from this function, you can pass it down to lower-level code, which is probably what you've
already done in existing code.

To get an overwritten value out of this context, you can execute the following function.

```go
// GetOverwrittenValue returns the Value of the variable under this name if it was overwritten
func GetOverwrittenValue(ctx context.Context, name string) Value
```

Another way is to extract the overwriting storage by calling the function below.

```go
// ExtractOverwritingStorage returns the Storage currently associated with ctx, or
// nil if no such Storage could be found.
func ExtractOverwritingStorage(ctx context.Context) Storage
```

Subsequently, you can call the `Get()` method of this `Storage` to get an overwritten value.

```go
// Get returns the Value of the variable under this name if it was overwritten
Get(name string) Value
```

Once you got a `Value`, take advantage of the provided helper methods to implement the overwriting behavior cleanly.

```go
// Value wraps a raw interface{} value
type Value interface {
    // AsIs returns the wrapped value as-is.
    AsIs() interface{}
    // AsString typecast to string. Returns zero value if not possible to cast.
    AsString() string
    // AsBool typecast to bool. Returns zero value if not possible to cast.
    AsBool() bool
    // AsFloat typecast to float64. Returns zero value if not possible to cast.
    // Note: Try not to use a raw value of type float32 if possible.
    // https://stackoverflow.com/questions/67145364/golang-losing-precision-while-converting-float32-to-float64
    AsFloat() float64
    // AsInt typecast to int64. Returns zero value if not possible to cast.
    // NOTE: JSON by default unmarshal to numbers which are treated as float.
    // Using this method, your float will lose precision as an int64.
    AsInt() int64
    // Unmarshal into the given type. t should be a pointer to a struct.
    Unmarshal(t interface{}) error
}
```