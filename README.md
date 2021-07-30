# yamldb (YAML Database)
Simple disk-backed key-value store for YAML files. 

You probably shouldn't use this as your go-to database solution in most cases.
A realistic use-case for yamldb is when you have a small app that requires a few human-readable configuration files to be stored locally.

## Features
- [x] CRUD data to disk
- [x] Reference other YAML files with constraints (think like SQL foreign keys)
- [x] Ordered keys through sorting, compression and anything else that's made possible with [diskv](https://github.com/peterbourgon/diskv)

## Usage
The best way to learn more about this API is to check out the tests: [yamldb](./yamldb_test.go) & [schemas](./schema_test.go).

But here's a very basic 'read-write arbitrary data' example to get you started:
```go
package main

import (
	"github.com/sleeyax/yamldb"
	"fmt"
)

type MyData struct {
	Ok    bool
	Value string
}

func main() {
	// create yaml db
	db := yamldb.New(&yamldb.DiskOptions{
		BasePath:        "./data",        // files will be stored in relative 'data' directory.
		AppendExtension: true,            // automatically append .yaml extensions (if not present).
		CacheSizeMax:    1024 * 1024 * 5, // 5MB in-memory caching 
	})
	
	// write some data
	// YAML will be written to: ./data/subdir/mykey.yaml
	db.Write("subdir/mykey", MyData{
		Ok:    true,
		Value: "hello world!",
	})
	
	// read data back into our struct
	var out MyData
	db.Read("subdir/mykey", &out)
	
	fmt.Println(out.Value)
}
```
