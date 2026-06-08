<img src="https://raw.githubusercontent.com/eclipse-zenoh/zenoh/main/zenoh-dragon.png" height="150">

[![CI](https://github.com/eclipse-zenoh/zenoh-go/workflows/CI/badge.svg)](https://github.com/eclipse-zenoh/zenoh-go/actions?query=workflow%3A%22CI%22)
[![Discussion](https://img.shields.io/badge/discussion-on%20github-blue)](https://github.com/eclipse-zenoh/roadmap/discussions)
[![Discord](https://img.shields.io/badge/chat-on%20discord-blue)](https://discord.gg/2GJ958VuHs)
[![License](https://img.shields.io/badge/License-EPL%202.0-blue)](https://choosealicense.com/licenses/epl-2.0/)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

# Eclipse Zenoh

The Eclipse Zenoh: Zero Overhead Pub/sub, Store/Query and Compute.

Zenoh (pronounce _/zeno/_) unifies data in motion, data at rest and computations. It carefully blends traditional pub/sub with geo-distributed storages, queries and computations, while retaining a level of time and space efficiency that is well beyond any of the mainstream stacks.

Check the website [zenoh.io](http://zenoh.io) and the [roadmap](https://github.com/eclipse-zenoh/roadmap) for more detailed information.

-------------------------------

# Zenoh Go Binding

This repository contains the Go bindings for Zenoh.

## Dependencies

Before building and running the examples, you need to have the following dependencies installed:

- **Zenoh-C**: The C implementation of the Zenoh protocol.

### Installing Zenoh-C

Zenoh-C is included as a git submodule. It must be built and installed **before** building the Go examples, with unstable API support enabled (`-DZENOHC_BUILD_WITH_UNSTABLE_API=ON`).

You can also follow the upstream instructions in the [Zenoh-C repository](https://github.com/eclipse-zenoh/zenoh-c).

## Building the Examples

This project includes several examples located in the `examples` directory. Each example is in a subdirectory prefixed with `z_`. You can build all the examples using the provided `Makefile`.

The Go bindings use CGo and need the installed Zenoh-C headers and libraries on the include/link path. Run the commands below in order from the repository root.

### Build All Examples

```bash
# 1. Initialize and update the zenoh-c submodule
git submodule init
git submodule update

# 2. Configure and build zenoh-c; install into ./build/ (gitignored)
mkdir -p build && cd build
cmake ../zenoh-c \
  -DZENOHC_BUILD_WITH_UNSTABLE_API=ON \
  -DCMAKE_INSTALL_PREFIX="$PWD"
cmake --build . --target install --config Release
cd ..

# 3. Point CGo at the installed zenoh-c, then build all examples
export CGO_CFLAGS="-I$(pwd)/build/include"
export CGO_LDFLAGS="-L$(pwd)/build/lib"
export LD_LIBRARY_PATH="$(pwd)/build/lib"
make all
```

Zenoh-C is installed under `build/` (`build/include`, `build/lib`).

### Build a Specific Example

To build a specific example, use the example's name. For instance, to build the `z_pub` example:

```bash
make z_pub
```

## Running the Examples

After building the examples, you can run them from the `bin` directory.
Description of each example can be found [here](./examples/README.md).

### Run a Specific Example

To run a specific example, navigate to the `bin` directory and execute the binary. For example, to run the `z_pub` example:

```bash
./bin/z_pub
```

### Run Examples Directly with go run ###

You can also run the examples directly using go run without building the binaries. For example, to run the z_sub example:

```bash
go run examples/z_pub/z_pub.go
```

## Project Structure

- `examples/`: This directory contains all the example subdirectories. Each example has its own subdirectory prefixed with `z_`.
- `bin/`: This directory will contain the compiled binaries for the examples.
- `build/`: Local install prefix for zenoh-c (headers and libraries) after building the submodule.

## Makefile

The provided `Makefile` includes the following targets:

- `all`: Builds all the examples.
- `fmt`: Formats the source code using `go fmt`.
- `clean`: Cleans up all generated binaries.

### Makefile Usage

- **Build all binaries**: `make all`
- **Build specific binary**: `make <example_name>`
- **Format the source code**: `make fmt`
- **Clean up binaries**: `make clean`