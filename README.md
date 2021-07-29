# YAML Database (yamldb)
Simple disk-backed key-value store for YAML files. 

You probably shouldn't use this as your go-to database solution in most cases.
A realistic use-case for yamldb is when you have a small app that requires a few human-readable configuration files to be stored locally.

## Usage
The basics:
```go
package main

import (
	"github.com/sleeyax/yamldb"
	"log"
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
	err := db.Write("subdir/mykey", MyData{
		Ok:    true,
		Value: "hello world!",
	})
	if err != nil {
		log.Fatal(err)
	}
	
	// read data back into our struct
	var out MyData
	if err = db.Read("subdir/mykey", &out); err != nil {
		log.Fatal(err)
	}
	
	log.Println(out.Value)
}
```

See [tests](./yamldb_test.go) for more examples.
