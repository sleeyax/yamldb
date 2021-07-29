package yamldb

import (
	"errors"
	"fmt"
)

type Action string

const (
	// Cascade modifies the referenced Schema based on the constraint.
	Cascade Action = "cascade"
	// Restrict prevents modifications to the referenced Schema based on the constraint.
	Restrict Action = "restrict"
	// NoAction doesn't touch the referenced Schema when any of the constraints are triggered.
	NoAction Action = "no action"
)

var (
	UpdateRestrictedError = errors.New("can't update because there's still a reference to another resource (blocked by constraint)")
)

type Constraints struct {
	// Update triggers specified Action when Schema Data is modified.
	Update Action
	// Delete triggers specified Action when this Schema gets deleted.
	Delete Action
}

type SchemaReference struct {
	// Key to another YAML file in the database.
	Key string

	// Constraints trigger Actions that have an effect on the referenced Schema whenever this Schema changes.
	Constraints Constraints
}

// Schema represents a versioned YAML data structure that can also reference other files in the database.
// Think of it like an SQL table.
type Schema struct {
	// Version number.
	// This value should be increased whenever the schema data type changes.
	// You can then check this number later in your code to perform the appropriate upgrades or migrations.
	Version int

	// YAML file key as stored in the database.
	// The value of this key MUST be unique AND equal the file key,
	// as it will later be used by other Schemas to reference this one.
	Key string

	// Reference is a link to another Schema.
	Reference *SchemaReference

	// Inner YAML data to store.
	Data interface{}
}

// Delete deletes the schema and file on disk while also triggering any set constraints.
func (s *Schema) Delete(db *YamlDb) error {
	if s.Reference != nil && db.Has(s.Reference.Key) {
		switch s.Reference.Constraints.Delete {
		case Cascade:
			if err := db.Delete(s.Reference.Key); err != nil {
				return err
			}
		case Restrict:
			return fmt.Errorf("can't delete because there's still a reference to another resource (blocked by constraint)")
		case NoAction:
		default:
			break
		}
	}

	if db.Has(s.Key) {
		return db.Delete(s.Key)
	}

	return nil
}

// Update changes this and referenced schema key while also triggering any set constraints.
func (s *Schema) Update(db *YamlDb, updatedKey string, onUpdate func(schema *Schema)) error {
	if s.Reference != nil {
		if !db.Has(s.Reference.Key) {
			return fmt.Errorf("old reference to %s not found", s.Reference.Key)
		}
		switch s.Reference.Constraints.Update {
		case Cascade:
			// TODO: link back to referenced table somehow (probably gonna need reflection) and actually update the 'foreign key' to the new key
			fallthrough
		case Restrict:
			return UpdateRestrictedError
		case NoAction:
		default:
			break
		}
	}

	return db.Update(s.Key, s, func(i interface{}) {
		schema := i.(*Schema)
		schema.Key = updatedKey
		if onUpdate != nil {
			onUpdate(schema)
		}
	})
}
