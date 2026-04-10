# zenoh-go development notes

## Building

zenoh-go uses CGo and requires the zenoh-c library. Build and install it from the submodule:

```sh
git submodule init && git submodule update
cd zenoh-c && mkdir -p build && cd build
cmake .. -DZENOHC_BUILD_WITH_UNSTABLE_API=ON -DCMAKE_INSTALL_PREFIX="$HOME/local"
cmake --build . --target install --config Release
```

Then build/test with:

```sh
CGO_CFLAGS="-I$HOME/local/include" CGO_LDFLAGS="-L$HOME/local/lib -lzenohc" go build ./...
CGO_CFLAGS="-I$HOME/local/include" CGO_LDFLAGS="-L$HOME/local/lib -lzenohc" go test ./...
```

## CGo Conventions

### C pointer conversion methods

Methods that convert a Go struct to a C pointer are named `toCPtr` and live in the same file as the Go type. They return a pointer to a C-owned struct (e.g. `*C.z_owned_foo_t`). Prefer stack allocation over `C.malloc`:

```go
func (t Foo) toCPtr() *C.z_owned_foo_t {
    var owned C.z_owned_foo_t
    // ... fill owned from t ...
    return &owned // escapes to heap via Go escape analysis
}
```

Never use `C.malloc` / `C.free` for temporary C structs — Go's escape analysis handles heap promotion.

### Lifetime bounds with runtime.Pinner

When a C pointer is passed to a C function (e.g. via `z_foo_move`), use a `runtime.Pinner` at the call site to make the lifetime explicit:

```go
pinner := runtime.Pinner{}
defer pinner.Unpin()
ownedPtr := val.toCPtr()
pinner.Pin(ownedPtr)
cOpts.foo = C.z_foo_move(ownedPtr)
```

The pinner can be reused for multiple fields in the same scope.

### Optional fields

Fields that may or may not be present in the C API should be typed as `option.Option[T]` (from `github.com/BooleanCat/option`), not as a zero-value sentinel like `""` or `0`. This applies to both struct fields and method return types. Use `option.Some(v)` when the value is present and `option.None[T]()` (or zero-value `option.Option[T]{}`) when absent.
