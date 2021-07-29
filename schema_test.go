package yamldb

import (
	"testing"
)

// One User can have many Posts.
type User struct {
	Id   int
	Name string
}

// Only one Post can belong to one User.
type Post struct {
	Id      int
	Message string
}

const (
	userKey = "users/1"
	postKey = "posts/1"
)

func getUserSchema() (Schema, *YamlDb, error) {
	db := New(&DiskOptions{
		BasePath:        basePath,
		AppendExtension: true,
	})

	// write user schema
	err := db.Write(userKey, Schema{
		Key: userKey,
		Reference: &SchemaReference{
			Key: postKey, // references the post below
			Constraints: Constraints{
				Delete: Cascade,
				Update: Cascade,
			},
		},
		Data: User{
			Id:   1,
			Name: "David",
		},
	})
	if err != nil {
		return Schema{}, nil, err
	}

	// write post schema
	err = db.Write(postKey, Schema{
		Key: postKey,
		Data: Post{
			Id:      1,
			Message: "Hello World!",
		},
	})

	// unserialize user schema
	var schema Schema
	if err = db.Read(userKey, &schema); err != nil {
		return Schema{}, nil, err
	}

	return schema, db, nil
}

func TestSchema_Delete(t *testing.T) {
	schema, db, err := getUserSchema()
	if err != nil {
		t.Fatal(err)
	}
	defer db.PurgeAll()

	// delete the user
	if err = schema.Delete(db); err != nil {
		t.Fatal(err)
	}

	// verify whether the related post has also been deleted
	if db.Has(userKey) || db.Has(postKey) {
		t.FailNow()
	}
}

func TestSchema_Update(t *testing.T) {
	schema, db, err := getUserSchema()
	if err != nil {
		t.Fatal(err)
	}
	defer db.PurgeAll()

	// update the user schema key
	// NOTE: in a realistic scenario you would also move the local file to a new location beforehand.
	if err = schema.Update(db, "users/2", nil); err != nil && err != UpdateRestrictedError {
		t.Fatal(err)
	}
}
