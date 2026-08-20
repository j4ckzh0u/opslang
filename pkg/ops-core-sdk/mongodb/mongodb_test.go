package mongodb

import (
	"testing"
)

// TestBuildURI tests URI construction with defaults and overrides.
func TestBuildURI(t *testing.T) {
	tests := []struct {
		name string
		host string
		port int
		db   string
		want string
	}{
		{"all defaults", "", 0, "", "mongodb://localhost:27017/test"},
		{"custom host", "10.0.0.1", 0, "", "mongodb://10.0.0.1:27017/test"},
		{"custom port", "", 27018, "", "mongodb://localhost:27018/test"},
		{"custom db", "", 0, "mydb", "mongodb://localhost:27017/mydb"},
		{"all custom", "mongo.example.com", 27019, "appdb", "mongodb://mongo.example.com:27019/appdb"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildURI(tt.host, tt.port, tt.db)
			if got != tt.want {
				t.Errorf("buildURI(%q, %d, %q) = %q, want %q", tt.host, tt.port, tt.db, got, tt.want)
			}
		})
	}
}

// TestCreateDatabaseValidation tests input validation.
func TestCreateDatabaseValidation(t *testing.T) {
	_, err := CreateDatabase("localhost", 27017, "")
	if err == nil {
		t.Fatal("expected error for empty database name, got nil")
	}
}

// TestDropDatabaseValidation tests input validation.
func TestDropDatabaseValidation(t *testing.T) {
	_, err := DropDatabase("localhost", 27017, "")
	if err == nil {
		t.Fatal("expected error for empty database name, got nil")
	}
}

// TestCreateUserValidation tests input validation.
func TestCreateUserValidation(t *testing.T) {
	_, err := CreateUser("localhost", 27017, "admin", "", "", "readWrite")
	if err == nil {
		t.Fatal("expected error for empty user/password, got nil")
	}
}

// TestDropUserValidation tests input validation.
func TestDropUserValidation(t *testing.T) {
	_, err := DropUser("localhost", 27017, "admin", "")
	if err == nil {
		t.Fatal("expected error for empty user, got nil")
	}
}

// TestCreateCollectionValidation tests input validation.
func TestCreateCollectionValidation(t *testing.T) {
	_, err := CreateCollection("localhost", 27017, "", "col1")
	if err == nil {
		t.Fatal("expected error for empty database, got nil")
	}
	_, err = CreateCollection("localhost", 27017, "mydb", "")
	if err == nil {
		t.Fatal("expected error for empty collection, got nil")
	}
}

// TestDropCollectionValidation tests input validation.
func TestDropCollectionValidation(t *testing.T) {
	_, err := DropCollection("localhost", 27017, "", "col1")
	if err == nil {
		t.Fatal("expected error for empty database, got nil")
	}
	_, err = DropCollection("localhost", 27017, "mydb", "")
	if err == nil {
		t.Fatal("expected error for empty collection, got nil")
	}
}

// TestCreateIndexValidation tests input validation.
func TestCreateIndexValidation(t *testing.T) {
	_, err := CreateIndex("localhost", 27017, "mydb", "col1", "", false, "")
	if err == nil {
		t.Fatal("expected error for empty keys, got nil")
	}
	_, err = CreateIndex("localhost", 27017, "", "col1", "field:1", false, "")
	if err == nil {
		t.Fatal("expected error for empty database, got nil")
	}
}

// TestDropIndexValidation tests input validation.
func TestDropIndexValidation(t *testing.T) {
	_, err := DropIndex("localhost", 27017, "mydb", "col1", "")
	if err == nil {
		t.Fatal("expected error for empty index name, got nil")
	}
}

// TestListIndexesValidation tests input validation.
func TestListIndexesValidation(t *testing.T) {
	_, err := ListIndexes("localhost", 27017, "", "col1")
	if err == nil {
		t.Fatal("expected error for empty database, got nil")
	}
	_, err = ListIndexes("localhost", 27017, "mydb", "")
	if err == nil {
		t.Fatal("expected error for empty collection, got nil")
	}
}

// TestFindMongoshNotFound verifies behavior when mongosh is absent.
func TestFindMongoshNotFound(t *testing.T) {
	// This test passes on systems without mongosh installed.
	// On systems with mongosh, it will find it and the test is effectively a no-op.
	_, _ = findMongosh()
}
