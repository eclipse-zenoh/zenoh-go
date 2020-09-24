..
.. Copyright (c) 2017, 2020 ADLINK Technology Inc.
..
.. This program and the accompanying materials are made available under the
.. terms of the Eclipse Public License 2.0 which is available at
.. http://www.eclipse.org/legal/epl-2.0, or the Apache License, Version 2.0
.. which is available at https://www.apache.org/licenses/LICENSE-2.0.
..
.. SPDX-License-Identifier: EPL-2.0 OR Apache-2.0
..
.. Contributors:
..   ADLINK zenoh team, <zenoh@adlink-labs.tech>
..

********
zenoh-go
********

The *zenoh-go* library provides a Go `zenoh client API <zenoh-api.html>`_ for zenoh.

An introduction to zenoh and its concepts is available on `zenoh.io <https://zenoh.io>`_.

.. warning::
    zenoh has been subjet to a complete rewrite with major protocol updates between 
    versions 0.4.2 and 0.5.0. The Go API does not yet integrate those changes and is 
    only compatible with version 0.4.2 of the zenoh daemon and the underlying zenoh-c 
    stack.

Note that this library also provides a low-level API (`zenoh-net <zenoh-net-api.html>`_)
that gives access to the zenoh protocol primitives and allow some
advanced use cases where a fine tuning of the protocol is required.


 * `zenoh API <https://godoc.org/github.com/eclipse-zenoh/zenoh-go>`_
 * `zenoh-net API <https://godoc.org/github.com/eclipse-zenoh/zenoh-go/net>`_


