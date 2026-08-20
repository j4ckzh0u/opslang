// Package mongodb provides MongoDB database management operations.
// Equivalent to Ansible's mongodb modules (mongodb_user, mongodb_db, etc).
package mongodb

import (
	"fmt"
	"os/exec"
	"strings"
)

// Result represents a generic operation result.
type Result struct {
	Status  string `json:"status"`
	Changed bool   `json:"changed"`
	Output  string `json:"output,omitempty"`
	Error   string `json:"error,omitempty"`
}

// DatabaseResult represents a database operation result.
type DatabaseResult struct {
	Status  string `json:"status"`
	Changed bool   `json:"changed"`
	Name    string `json:"name,omitempty"`
	SizeMB  int64  `json:"size_mb,omitempty"`
	Msg     string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// UserResult represents a user operation result.
type UserResult struct {
	Status   string `json:"status"`
	Changed  bool   `json:"changed"`
	User     string `json:"user,omitempty"`
	Database string `json:"database,omitempty"`
	Roles    string `json:"roles,omitempty"`
	Msg      string `json:"message,omitempty"`
	Error    string `json:"error,omitempty"`
}

// CollectionResult represents a collection operation result.
type CollectionResult struct {
	Status     string `json:"status"`
	Changed    bool   `json:"changed"`
	Database   string `json:"database,omitempty"`
	Collection string `json:"collection,omitempty"`
	Msg        string `json:"message,omitempty"`
	Error      string `json:"error,omitempty"`
}

// InfoResult represents info about a database.
type InfoResult struct {
	Status      string `json:"status"`
	Name        string `json:"name"`
	SizeMB      int64  `json:"size_mb,omitempty"`
	Collections int    `json:"collections,omitempty"`
	Indexes     int    `json:"indexes,omitempty"`
}

func findMongosh() (string, error) {
	if p, err := exec.LookPath("mongosh"); err == nil {
		return p, nil
	}
	if p, err := exec.LookPath("mongo"); err == nil {
		return p, nil
	}
	return "", fmt.Errorf("mongosh not found")
}

func runQuery(uri string, db string, query string) (string, error) {
	mongosh, err := findMongosh()
	if err != nil {
		return "", err
	}
	args := []string{}
	if uri != "" {
		args = append(args, uri)
	} else {
		if db != "" {
			args = append(args, db)
		}
	}
	args = append(args, "--quiet", "--eval", query)
	cmd := exec.Command(mongosh, args...)
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

// CreateDatabase creates a new database (by creating a collection in it).
func CreateDatabase(host string, port int, name string) (DatabaseResult, error) {
	if name == "" {
		return DatabaseResult{Status: "failed", Error: "database name is required"}, fmt.Errorf("database name is required")
	}
	uri := buildURI(host, port, name)
	query := fmt.Sprintf(`db = db.getSiblingDB('%s'); db.createCollection('_init')`, name)
	out, err := runQuery(uri, "", query)
	if err != nil {
		return DatabaseResult{Status: "failed", Name: name, Error: fmt.Sprintf("create database: %v", err)}, err
	}
	return DatabaseResult{Status: "success", Changed: true, Name: name, Msg: out}, nil
}

// DropDatabase drops a database.
func DropDatabase(host string, port int, name string) (DatabaseResult, error) {
	if name == "" {
		return DatabaseResult{Status: "failed", Error: "database name is required"}, fmt.Errorf("database name is required")
	}
	uri := buildURI(host, port, name)
	query := `db.dropDatabase()`
	out, err := runQuery(uri, "", query)
	if err != nil {
		return DatabaseResult{Status: "failed", Name: name, Error: fmt.Sprintf("drop database: %v", err)}, err
	}
	return DatabaseResult{Status: "success", Changed: true, Name: name, Msg: out}, nil
}

// ListDatabases lists all databases.
func ListDatabases(host string, port int) ([]InfoResult, error) {
	uri := buildURI(host, port, "admin")
	query := `JSON.stringify(db.adminCommand('listDatabases').databases.map(d => ({name: d.name, size_mb: Math.round(d.sizeOnDisk / 1048576)})))`
	out, err := runQuery(uri, "", query)
	if err != nil {
		return nil, fmt.Errorf("list databases: %w", err)
	}
	var results []InfoResult
	// Parse simplified JSON output
	out = strings.Trim(out, "[]")
	if out == "" {
		return results, nil
	}
	for _, entry := range strings.Split(out, "},{") {
		entry = strings.Trim(entry, "{}")
		info := InfoResult{Status: "success"}
		for _, pair := range strings.Split(entry, ",") {
			parts := strings.SplitN(pair, ":", 2)
			if len(parts) == 2 {
				key := strings.Trim(parts[0], "\" ")
				val := strings.Trim(parts[1], "\" ")
				switch key {
				case "name":
					info.Name = val
				case "size_mb":
					// parse size
				}
			}
		}
		if info.Name != "" {
			results = append(results, info)
		}
	}
	return results, nil
}

// CreateUser creates a new user.
func CreateUser(host string, port int, database string, user string, password string, roles string) (UserResult, error) {
	if user == "" || password == "" {
		return UserResult{Status: "failed", Error: "user and password are required"}, fmt.Errorf("user and password are required")
	}
	if database == "" {
		database = "admin"
	}
	uri := buildURI(host, port, database)
	roleStr := "[]"
	if roles != "" {
		// Convert comma-separated roles to MongoDB role format
		roleParts := strings.Split(roles, ",")
		roleList := make([]string, 0, len(roleParts))
		for _, r := range roleParts {
			r = strings.TrimSpace(r)
			if r != "" {
				roleList = append(roleList, fmt.Sprintf(`"%s"`, r))
			}
		}
		roleStr = "[" + strings.Join(roleList, ",") + "]"
	}
	query := fmt.Sprintf(`db.createUser({user: "%s", pwd: "%s", roles: %s})`, user, password, roleStr)
	out, err := runQuery(uri, "", query)
	if err != nil {
		return UserResult{Status: "failed", User: user, Database: database, Error: fmt.Sprintf("create user: %v", err)}, err
	}
	return UserResult{Status: "success", Changed: true, User: user, Database: database, Roles: roles, Msg: out}, nil
}

// DropUser drops a user.
func DropUser(host string, port int, database string, user string) (UserResult, error) {
	if user == "" {
		return UserResult{Status: "failed", Error: "user is required"}, fmt.Errorf("user is required")
	}
	if database == "" {
		database = "admin"
	}
	uri := buildURI(host, port, database)
	query := fmt.Sprintf(`db.dropUser("%s")`, user)
	out, err := runQuery(uri, "", query)
	if err != nil {
		return UserResult{Status: "failed", User: user, Database: database, Error: fmt.Sprintf("drop user: %v", err)}, err
	}
	return UserResult{Status: "success", Changed: true, User: user, Database: database, Msg: out}, nil
}

// ListUsers lists users in a database.
func ListUsers(host string, port int, database string) ([]map[string]string, error) {
	if database == "" {
		database = "admin"
	}
	uri := buildURI(host, port, database)
	query := `JSON.stringify(db.system.users.find().toArray().map(u => ({user: u.user, db: u.db, roles: JSON.stringify(u.roles)})))`
	out, err := runQuery(uri, "", query)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	var results []map[string]string
	// Simplified parsing
	out = strings.Trim(out, "[]")
	if out == "" {
		return results, nil
	}
	for _, entry := range strings.Split(out, "},{") {
		entry = strings.Trim(entry, "{}")
		info := make(map[string]string)
		for _, pair := range strings.Split(entry, ",") {
			parts := strings.SplitN(pair, ":", 2)
			if len(parts) == 2 {
				key := strings.Trim(parts[0], "\" ")
				val := strings.Trim(parts[1], "\" ")
				info[key] = val
			}
		}
		if len(info) > 0 {
			results = append(results, info)
		}
	}
	return results, nil
}

// CreateCollection creates a new collection.
func CreateCollection(host string, port int, database string, collection string) (CollectionResult, error) {
	if database == "" || collection == "" {
		return CollectionResult{Status: "failed", Error: "database and collection are required"}, fmt.Errorf("database and collection are required")
	}
	uri := buildURI(host, port, database)
	query := fmt.Sprintf(`db = db.getSiblingDB('%s'); db.createCollection('%s')`, database, collection)
	out, err := runQuery(uri, "", query)
	if err != nil {
		return CollectionResult{Status: "failed", Database: database, Collection: collection, Error: fmt.Sprintf("create collection: %v", err)}, err
	}
	return CollectionResult{Status: "success", Changed: true, Database: database, Collection: collection, Msg: out}, nil
}

// DropCollection drops a collection.
func DropCollection(host string, port int, database string, collection string) (CollectionResult, error) {
	if database == "" || collection == "" {
		return CollectionResult{Status: "failed", Error: "database and collection are required"}, fmt.Errorf("database and collection are required")
	}
	uri := buildURI(host, port, database)
	query := fmt.Sprintf(`db = db.getSiblingDB('%s'); db.%s.drop()`, database, collection)
	out, err := runQuery(uri, "", query)
	if err != nil {
		return CollectionResult{Status: "failed", Database: database, Collection: collection, Error: fmt.Sprintf("drop collection: %v", err)}, err
	}
	return CollectionResult{Status: "success", Changed: true, Database: database, Collection: collection, Msg: out}, nil
}

// ListCollections lists collections in a database.
func ListCollections(host string, port int, database string) ([]string, error) {
	if database == "" {
		database = "admin"
	}
	uri := buildURI(host, port, database)
	query := `JSON.stringify(db.getCollectionNames())`
	out, err := runQuery(uri, "", query)
	if err != nil {
		return nil, fmt.Errorf("list collections: %w", err)
	}
	out = strings.Trim(out, "[]\"")
	if out == "" {
		return nil, nil
	}
	parts := strings.Split(out, "\",\"")
	return parts, nil
}

// IndexResult represents an index.
type IndexResult struct {
	Name    string   `json:"name"`
	Keys    []string `json:"keys"`
	Unique  bool     `json:"unique"`
	Sparse  bool     `json:"sparse"`
}

// CreateIndex creates an index on a collection.
func CreateIndex(host string, port int, database string, collection string, keys string, unique bool, name string) (Result, error) {
	if database == "" || collection == "" || keys == "" {
		return Result{Status: "failed", Error: "database, collection, and keys are required"}, fmt.Errorf("database, collection, and keys are required")
	}
	uri := buildURI(host, port, database)
	// Parse keys: "field1:1,field2:-1"
	keyParts := strings.Split(keys, ",")
	keyDoc := make([]string, 0, len(keyParts))
	for _, kp := range keyParts {
		kv := strings.SplitN(strings.TrimSpace(kp), ":", 2)
		if len(kv) == 2 {
			keyDoc = append(keyDoc, fmt.Sprintf(`"%s": %s`, kv[0], kv[1]))
		}
	}
	keyStr := "{" + strings.Join(keyDoc, ",") + "}"
	query := fmt.Sprintf(`db = db.getSiblingDB('%s'); db.%s.createIndex(%s`, database, collection, keyStr)
	if unique {
		query += `, {unique: true`
		if name != "" {
			query += fmt.Sprintf(`, name: "%s"`, name)
		}
		query += "})"
	} else if name != "" {
		query += fmt.Sprintf(`, {name: "%s"})`, name)
	} else {
		query += ")"
	}
	out, err := runQuery(uri, "", query)
	if err != nil {
		return Result{Status: "failed", Output: out, Error: fmt.Sprintf("create index: %v", err)}, err
	}
	return Result{Status: "success", Changed: true, Output: out}, nil
}

// DropIndex drops an index.
func DropIndex(host string, port int, database string, collection string, indexName string) (Result, error) {
	if database == "" || collection == "" || indexName == "" {
		return Result{Status: "failed", Error: "database, collection, and index name are required"}, fmt.Errorf("database, collection, and index name are required")
	}
	uri := buildURI(host, port, database)
	query := fmt.Sprintf(`db = db.getSiblingDB('%s'); db.%s.dropIndex("%s")`, database, collection, indexName)
	out, err := runQuery(uri, "", query)
	if err != nil {
		return Result{Status: "failed", Output: out, Error: fmt.Sprintf("drop index: %v", err)}, err
	}
	return Result{Status: "success", Changed: true, Output: out}, nil
}

// ListIndexes lists indexes on a collection.
func ListIndexes(host string, port int, database string, collection string) ([]IndexResult, error) {
	if database == "" || collection == "" {
		return nil, fmt.Errorf("database and collection are required")
	}
	uri := buildURI(host, port, database)
	query := fmt.Sprintf(`db = db.getSiblingDB('%s'); JSON.stringify(db.%s.getIndexes())`, database, collection)
	out, err := runQuery(uri, "", query)
	if err != nil {
		return nil, fmt.Errorf("list indexes: %w", err)
	}
	// Return raw output for simplicity
	var results []IndexResult
	results = append(results, IndexResult{
		Name: out,
	})
	return results, nil
}

// ServerStatus returns server status info.
func ServerStatus(host string, port int) (map[string]interface{}, error) {
	uri := buildURI(host, port, "admin")
	query := `JSON.stringify(db.serverStatus())`
	out, err := runQuery(uri, "", query)
	if err != nil {
		return nil, fmt.Errorf("server status: %w", err)
	}
	return map[string]interface{}{"status": out}, nil
}

// ReplicaSetStatus returns replica set status.
func ReplicaSetStatus(host string, port int) (map[string]interface{}, error) {
	uri := buildURI(host, port, "admin")
	query := `JSON.stringify(rs.status())`
	out, err := runQuery(uri, "", query)
	if err != nil {
		return nil, fmt.Errorf("replica set status: %w", err)
	}
	return map[string]interface{}{"status": out}, nil
}

func buildURI(host string, port int, db string) string {
	if host == "" {
		host = "localhost"
	}
	if port == 0 {
		port = 27017
	}
	if db == "" {
		db = "test"
	}
	return fmt.Sprintf("mongodb://%s:%d/%s", host, port, db)
}
