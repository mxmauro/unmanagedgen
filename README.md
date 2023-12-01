# unmanagedgen

This tool generates non-GC versions of Golang structs. They use an Allocator interface to allocate and free memory.

## The generator

You can add a directive like this:

```golang
//go:generate go run cmd\main.go --file mystruct.go
``` 

Or a code like this:

```golang
import (
	"github.com/mxmauro/unmanagedgen/processor"
)

func createFiles() {
	// ...
	err := processor.ProcessFolder(settings.FileMask, ...)
    // ...
}
```

## The generated code

Let's take the following struct as an example:

```golang
type Sample struct {
	A int
	B string
}
```

After running the generator, a similar struct is generated:

```golang
type UnmanagedSample struct {
	A int
	B string
}
```

And several helper methods:

* The struct allocator
```golang
func NewUnmanagedSample(alloc allocator.Allocator) *UnmanagedSample
```

* A method to initialize stack or embedded structs.

```golang
func (v *UnmanagedSample) InitAllocator(alloc allocator.Allocator)
```

* The method that frees all structure memory including the used by its fields.

```golang
func (v *UnmanagedSample) Free()
```

* Setter helpers, used mainly by string, slice and pointer fields.

```golang
func (v *UnmanagedSample) SetA(value string)
```

## Final notes:

* **UNMANAGED DATA MUST BE HANDLED WITH CARE**. For example, in Golang, when a string or slice is copied, only the
  header actually copied. The underlying memory that contains data remains the same. If no precautions are taken,
  you may end using memory references after freeing them.

* It is recommended to download this library in a separated folder and run the processor test. It will generate a
  large variety of setters because the example contains a lot of field types.

* For slices, helpers that sets the length and capacity are generated too.

* Also, for slices and arrays, setters of elements by index are created.

* The library aims to support native Go types and structs. Pointers, slices, arrays and their combinations are also
  supported.


## LICENSE

[MIT](/LICENSE.txt)
