# Zenoh Go examples

## Start instructions
   
   ```bash
   go run <example/example.go>
   ```

   Each example accepts the -h or --help option that provides a description of its arguments and their default values.

## Examples description

### z_add_storage

   Add a storage in the Zenoh router it's connected to.

   Usage:
   ```bash
   go run z_add_storage/z_add_storage.go [--selector SELECTOR] [--id ID] [--locator LOCATOR]
   ```

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
   go run z_put/z_put.go [--path PATH] [--locator LOCATOR] [--msg MSG]
   ```

### z_put_float

   Put a key/value into Zenoh where the value is a float.
   The key/value will be stored by all the storages with a selector that matches the key.
   It will also be received by all the matching subscribers (see [z_sub](#z_sub) below).
   Note that if no storage and no subscriber are matching the key, the key/value will be dropped.
   Therefore, you probably should run [z_add_storage](#z_add_storage) and/or [z_sub](#z_sub) before z_put_float.

   Usage:
   ```bash
   go run z_put_float/z_put_float.go [--path PATH] [--locator LOCATOR]
   ```

### z_get

   Get a list of keys/values from Zenoh.  
   The values will be retrieved from the Storages containing paths that match the specified selector.  
   The Eval functions (see [z_eval](#z_eval) below) registered with a path matching the selector
   will also be triggered.

   Usage:
   ```bash
   go run z_get/z_get.go [--selector SELECTOR] [--locator LOCATOR]
   ```

### z_remove

   Remove a key and its associated value from Zenoh.  
   Any storage that store the key/value will drop it.  
   The subscribers with a selector matching the key will also receive a notification of this removal.

   Usage:
   ```bash
   go run z_remove/z_remove.go [--path PATH] [--locator LOCATOR]
   ```

### z_sub

   Register a subscriber with a selector.  
   The subscriber will be notified of each put/remove made on any path matching the selector,
   and will print this notification.

   Usage:
   ```bash
   go run z_sub/z_sub.go [--selector SELECTOR] [--locator LOCATOR]
   ```

### z_eval

   Register an evaluation function with a path.  
   This evaluation function will be triggered by each call to a get operation on Zenoh 
   with a selector that matches the path. In this example, the function returns a string value.
   See the code for more details.

   Usage:
   ```bash
   go run z_eval/z_eval.go [--path PATH] [--locator LOCATOR]
   ```

### z_put_thr & z_sub_thr

   Pub/Sub throughput test.
   This example allows to perform throughput measurements between a pubisher performing
   put operations and a subscriber receiving notifications of those put.
   Note that you can run this example with or without any storage.

   Subscriber usage:
   ```bash
   go run z_sub_thr/z_sub_thr.go [--path PATH] [--locator LOCATOR]
   ```

   Publisher usage:
   ```bash
   go run z_put_thr/z_put_thr.go [--size SIZE] [--locator LOCATOR] [--path PATH]
   ```