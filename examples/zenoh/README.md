# Zenoh Go examples

## Prerequisites

   The [Zenoh](https://zenoh.io) C client library must be installed on your host.  
   See installation instructions on https://zenoh.io or clone, build and install it yourself from https://github.com/eclipse-zenoh/zenoh-c.

## Start instructions
   
   ```bash
   go run <example/example.go>
   ```

## Examples description

### z_add_storage

   Add a storage in the Zenoh router it's connected to.

   Usage:
   ```bash
   go run z_add_storage/z_add_storage.go [selector] [storage-id] [locator]
   ```
   where the optional arguments are:
   - **selector** :  the selector matching the keys (path) that have to be stored.  
                     Default value: `/demo/example/**`
   - **storage-id** : the storage identifier.  
                      Default value: `Demo` 
   - **locator** : the locator of the Zenoh router to connect.  
                   Default value: none, meaning the Zenoh router is found via multicast.

   Note that his example doesn't specify the Backend that Zenoh has to use for storage creation.  
   Therefore, Zenoh will automatically select the memory backend, meaning the storage will be in memory
   (i.e. not persistent).

### z_put

   Put a key/value into Zenoh.  
   The key/value will be stored by all the storages with a selector that matches the key.
   It will also be received by all the matching subscribers (see [z_sub](#z_sub) below).  
   Note that if no storage and no subscriber are matching the key, the key/value will be dropped.
   Therefore, you probably should run [z_add_storage](#z_add_storage) and/or [z_sub](#z_sub) before z_put.

   Usage:
   ```bash
   go run z_put/z_put.go [path] [value] [locator]
   ```
   where the optional arguments are:
   - **path** : the path used as a key for the value.  
                Default value: `/demo/example/zenoh-go-put` 
   - **value** : the value (as a string).  
                Default value: `"Put from Zenoh Go!"` 
   - **locator** : the locator of the Zenoh router to connect.  
                   Default value: none, meaning the Zenoh router is found via multicast.

### z_get

   Get a list of keys/values from Zenoh.  
   The values will be retrieved from the Storages containing paths that match the specified selector.  
   The Eval functions (see [z_eval](#z_eval) below) registered with a path matching the selector
   will also be triggered.

   Usage:
   ```bash
   go run z_get/z_get.go [selector] [locator]
   ```
   where the optional arguments are:
   - **selector** : the selector that all replies shall match.  
                    Default value: `/demo/example/**` 
   - **locator** : the locator of the Zenoh router to connect.  
                   Default value: none, meaning the Zenoh router is found via multicast.

### z_remove

   Remove a key and its associated value from Zenoh.  
   Any storage that store the key/value will drop it.  
   The subscribers with a selector matching the key will also receive a notification of this removal.

   Usage:
   ```bash
   go run z_remove/z_remove.go [path] [locator]
   ```
   where the optional arguments are:
   - **path** : the key to be removed.  
                Default value: `/demo/example/zenoh-go-put` 
   - **locator** : the locator of the Zenoh router to connect.  
                   Default value: none, meaning the Zenoh router is found via multicast.

### z_sub

   Register a subscriber with a selector.  
   The subscriber will be notified of each put/remove made on any path matching the selector,
   and will print this notification.

   Usage:
   ```bash
   go run z_sub/z_sub.go [selector] [locator]
   ```
   where the optional arguments are:
   - **selector** : the subscription selector.  
                    Default value: `/demo/example/**` 
   - **locator** : the locator of the Zenoh router to connect.  
                   Default value: none, meaning the Zenoh router is found via multicast.

### z_eval

   Register an evaluation function with a path.  
   This evaluation function will be triggered by each call to a get operation on Zenoh 
   with a selector that matches the path. In this example, the function returns a string value.
   See the code for more details.

   Usage:
   ```bash
   go run z_eval/z_eval.go [selector] [locator]
   ```
   where the optional arguments are:
   - **path** : the eval path.  
                Default value: `/demo/example/zenoh-go-eval` 
   - **locator** : the locator of the Zenoh router to connect.  
                   Default value: none, meaning the Zenoh router is found via multicast.

### z_put_thr & z_sub_thr

   Pub/Sub throughput test.
   This example allows to perform throughput measurements between a pubisher performing
   put operations and a subscriber receiving notifications of those put.
   Note that you can run this example with or without any storage.

   Publisher usage:
   ```bash
   go run z_put_thr/z_put_thr.go <payload-size> [locator]
   ```
   where the arguments are:
   - **payload-size** : the size of the payload in bytes.  
   - **locator** : the locator of the Zenoh router to connect.  
                   Default value: none, meaning the Zenoh router is found via multicast.

   Subscriber usage:
   ```bash
   go run z_sub_thr/z_sub_thr.go [locator]
   ```
   where the optional arguments are:
   - **locator** : the locator of the Zenoh router to connect.  
                   Default value: none, meaning the Zenoh router is found via multicast.
