![zenoh banner](./zenoh-dragon.png)

![Build](https://github.com/eclipse-zenoh/zenoh-go/workflows/Go/badge.svg)
[![GoReport Status](https://goreportcard.com/badge/github.com/eclipse-zenoh/zenoh-go)](https://goreportcard.com/report/github.com/eclipse-zenoh/zenoh-go)
[![Documentation Status](https://readthedocs.org/projects/zenoh-go/badge/?version=latest)](https://zenoh-go.readthedocs.io/en/latest/?badge=latest)
[![Gitter](https://badges.gitter.im/atolab/zenoh.svg)](https://gitter.im/atolab/zenoh?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge)
[![License](https://img.shields.io/badge/License-EPL%202.0-blue)](https://choosealicense.com/licenses/epl-2.0/)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

# Eclipse zenoh Go API

[Eclipse zenoh](http://zenoh.io) is an extremely efficient and fault-tolerant [Named Data Networking](http://named-data.net) (NDN) protocol 
that is able to scale down to extremely constrainded devices and networks. 

The Go API is for pure clients, in other terms does not support peer-to-peer communication, can be easily tested against a zenoh router running in a Docker container (see https://github.com/eclipse-zenoh/zenoh#how-to-test-it). 

-------------------------------
## How to install it

The zenoh-go library relies on the zenoh-c library. Please install it, either [installing the **libzenohc-dev** package](https://github.com/eclipse-zenoh/zenoh-c#how-to-install-it), either [building zenoh-c](https://github.com/eclipse-zenoh/zenoh-c#how-to-build-it) by yourself

Supported Go version: **1.14.0** minimum.

Install the zenoh-go library via the usual `go get`command:
  ```bash
  $ go get github.com/eclipse-zenoh/zenoh-go
  ```

-------------------------------
## Running the Examples

The simplest way to run some of the example is to get a Docker image of the **zenoh** network router (see https://github.com/eclipse-zenoh/zenoh#how-to-test-it) and then to run the examples on your machine.

Then, run the zenoh-go examples following the instructions in [examples/zenoh/README.md](https://github.com/eclipse-zenoh/zenoh-go/blob/master/examples/zenoh/README.md)

