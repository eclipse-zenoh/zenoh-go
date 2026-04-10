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
