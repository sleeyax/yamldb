// Package yamldb is a simple disk-backed key-value store for YAML files.
package yamldb

import (
	"github.com/peterbourgon/diskv/v3"
	"gopkg.in/yaml.v2"
	"os"
	"path"
	"strings"
)

const extension = ".yaml"

type YamlDb struct {
	// Disk exposes the underlying diskv disk.
	// Set and/or get this field only for advanced usage.
	Disk *diskv.Diskv
}

type DiskPermissions struct {
	Path os.FileMode
	File os.FileMode
}

// DiskOptions are simplified diskv Disk options.
type DiskOptions struct {
	// Path to directory where files should be stored.
	BasePath string
	// Max in-memory cache size.
	CacheSizeMax uint64
	// Whether to append .yaml extension to keys (if not added yet).
	AppendExtension bool
	// Compression method to use on files before they are being stored on Disk.
	Compression Compression
	// Permissions to set on files and directories.
	Permissions DiskPermissions
	// If TempDir is set,
	// it will enable filesystem atomic writes by writing temporary files to that location before being moved to BasePath.
	// Note that TempDir MUST be on the same device/partition as BasePath.
	TempDir string
	// Enable sorting of keys in memory.
	SortKeys bool
	// Function that will be used to sort the keys when SortKeys is true.
	// Defaults to alphabetic sorting.
	SortOrderFunc OrderFunc
}

func New(options *DiskOptions) *YamlDb {
	diskOptions := diskv.Options{
		BasePath:     options.BasePath,
		CacheSizeMax: options.CacheSizeMax,
		AdvancedTransform: func(key string) *diskv.PathKey {
			if options.AppendExtension && !strings.HasSuffix(key, extension) {
				key += extension
			}
			split := strings.Split(key, "/")
			lastIndex := len(split) - 1
			return &diskv.PathKey{
				Path:     split[:lastIndex],
				FileName: split[lastIndex],
			}
		},
		InverseTransform: func(pathKey *diskv.PathKey) string {
			if options.AppendExtension && !strings.HasSuffix(pathKey.FileName, extension) {
				pathKey.FileName += extension
			}
			return path.Join(strings.Join(pathKey.Path, "/"), pathKey.FileName)
		},
		Compression: getCompression(options.Compression),
		PathPerm:    options.Permissions.Path,
		FilePerm:    options.Permissions.File,
		TempDir:     options.TempDir,
	}
	if options.SortKeys {
		diskOptions.Index = &diskv.BTreeIndex{}
		diskOptions.IndexLess = OrderAlphabetically
		if f := options.SortOrderFunc; f != nil {
			diskOptions.IndexLess = diskv.LessFunction(f)
		}
	}
	return &YamlDb{
		Disk: diskv.New(diskOptions),
	}
}

// Write writes a YAML-serializable struct to the database.
func (db *YamlDb) Write(key string, s interface{}) error {
	data, err := yaml.Marshal(s)
	if err != nil {
		return err
	}

	if err = db.Disk.Write(key, data); err != nil {
		return err
	}

	return nil
}

// WriteRaw writes a raw YAML resource to the database.
func (db *YamlDb) WriteRaw(key string, data []byte) error {
	return db.Disk.Write(key, data)
}

func (db *YamlDb) read(key string, out interface{}, strict bool) error {
	v, err := db.Disk.Read(key)
	if err != nil {
		return err
	}

	if strict {
		err = yaml.UnmarshalStrict(v, out)
	} else {
		err = yaml.Unmarshal(v, out)
	}

	if err != nil {
		return err
	}

	return nil
}

// Read unmarshals a resource from the database into given struct.
func (db *YamlDb) Read(key string, out interface{}) error {
	return db.read(key, out, false)
}

// ReadStrict strictly unmarshals a resource from the database into given struct.
func (db *YamlDb) ReadStrict(key string, out interface{}) error {
	return db.read(key, out, true)
}

// ReadRaw reads a raw YAML resource from the database.
func (db *YamlDb) ReadRaw(key string) ([]byte, error) {
	return db.Disk.Read(key)
}

// Delete removes a resource from the database.
func (db *YamlDb) Delete(key string) error {
	return db.Disk.Erase(key)
}

// Update updates a resource from the database with Values from provided update function.
func (db *YamlDb) Update(key string, t interface{}, updateFunc func(s interface{})) error {
	if err := db.Read(key, t); err != nil {
		return err
	}

	updateFunc(t)

	if err := db.Write(key, t); err != nil {
		return err
	}

	return nil
}

// Has returns whether the database stores a value for provided key.
func (db *YamlDb) Has(key string) bool {
	return db.Disk.Has(key)
}

// Purge removes all resources from the database that start with given prefix.
// This action cannot be undone!
func (db *YamlDb) Purge(prefix string) error {
	cancel := make(chan struct{})
	for key := range db.Disk.KeysPrefix(prefix, cancel) {
		if err := db.Delete(key); err != nil {
			close(cancel)
			return err
		}
	}
	return nil
}

// PurgeAll removes all data from the database.
// Note that *any* file that is found on disk will be deleted.
// This action cannot be undone!
func (db *YamlDb) PurgeAll() error {
	return db.Disk.EraseAll()
}

// Iterate iterates over each YAML resource found by provided key prefix.
//
// Returns a status bool indicating whether any resources with given prefix have been found,
// and thus have been iterated on or not.
func (db *YamlDb) Iterate(prefix string, callback func(key string, data []byte) error) (bool, error) {
	cancel := make(chan struct{})
	var hasIterated bool

	for key := range db.Disk.KeysPrefix(prefix, cancel) {
		if hasIterated == false {
			hasIterated = true
		}

		data, err := db.ReadRaw(key)
		if err != nil {
			close(cancel)
			return false, err
		}

		if err = callback(key, data); err != nil {
			close(cancel)
			return false, err
		}
	}

	return hasIterated, nil
}

// IterateSerialized iterates over each resource it can find in the given directory (relative to storage.Data),
// serializes it and passes it though specified callback for processing.
//
// Returns a status bool indicating whether any resources with given prefix have been found,
// and thus have been iterated on or not.
func (db *YamlDb) IterateSerialized(prefix string, out interface{}, callback func(p interface{}) error) (bool, error) {
	return db.Iterate(prefix, func(key string, data []byte) error {
		if err := yaml.Unmarshal(data, out); err != nil {
			return err
		}
		return callback(out)
	})
}

// GetOrderedKeys returns all keys - or those with given prefix - in order when SortKeys is enabled.
// Specify how many keys should be fetched from the underlying Index at a time through the chunks parameter.
// You can also start querying from a specific key.
func (db *YamlDb) GetOrderedKeys(prefix string, from string, chunks int) []string {
	var keysOrdered []string

	for keys := db.Disk.Index.Keys(from, chunks); len(keys) != 0; {
		for _, key := range keys {
			if strings.HasPrefix(key, prefix) {
				keysOrdered = append(keysOrdered, key)
			}
		}
		keys = db.Disk.Index.Keys(keys[len(keys)-1], chunks+1)
	}

	return keysOrdered
}
