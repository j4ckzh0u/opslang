// Package postgresql provides PostgreSQL database management operations.
package postgresql

import (
	"fmt"
	"os/exec"
	"strings"
)

// DB represents a PostgreSQL database.
type DB struct {
	Name  string `json:"name"`
	Owner string `json:"owner"`
}

// DBListResult represents the result of listing databases.
type DBListResult struct {
	Databases []DB `json:"databases"`
}

// ActionResult represents the result of a PostgreSQL action.
type ActionResult struct {
	Changed bool   `json:"changed"`
	Message string `json:"message"`
}

// User represents a PostgreSQL user.
type User struct {
	Name string `json:"name"`
}

// UserListResult represents the result of listing users.
type UserListResult struct {
	Users []User `json:"users"`
}

// runPSQL executes a psql command and returns the output.
func runPSQL(query string) (string, error) {
	cmd := exec.Command("psql", "-t", "-A", "-c", query)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("psql failed: %w (output: %s)", err, string(out))
	}
	return strings.TrimSpace(string(out)), nil
}

// Databases returns all databases.
func Databases() (*DBListResult, error) {
	out, err := runPSQL("SELECT datname || '|' || pg_catalog.pg_get_userbyid(datdba) FROM pg_database WHERE datistemplate = false ORDER BY datname")
	if err != nil {
		return nil, err
	}

	result := &DBListResult{Databases: make([]DB, 0)}
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "|", 2)
		if len(parts) >= 2 {
			result.Databases = append(result.Databases, DB{
				Name:  parts[0],
				Owner: parts[1],
			})
		}
	}
	return result, nil
}

// CreateDatabase creates a new database.
func CreateDatabase(name string) (*ActionResult, error) {
	_, err := runPSQL(fmt.Sprintf("CREATE DATABASE %q", name))
	if err != nil {
		return nil, err
	}
	return &ActionResult{Changed: true, Message: fmt.Sprintf("Created database %s", name)}, nil
}

// DropDatabase drops a database.
func DropDatabase(name string) (*ActionResult, error) {
	_, err := runPSQL(fmt.Sprintf("DROP DATABASE %q", name))
	if err != nil {
		return nil, err
	}
	return &ActionResult{Changed: true, Message: fmt.Sprintf("Dropped database %s", name)}, nil
}

// Users returns all users.
func Users() (*UserListResult, error) {
	out, err := runPSQL("SELECT rolname FROM pg_roles WHERE rolcanlogin = true ORDER BY rolname")
	if err != nil {
		return nil, err
	}

	result := &UserListResult{Users: make([]User, 0)}
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			result.Users = append(result.Users, User{Name: line})
		}
	}
	return result, nil
}

// CreateUser creates a new user.
func CreateUser(user string, password string) (*ActionResult, error) {
	query := fmt.Sprintf("CREATE USER %q WITH PASSWORD '%s'", user, strings.ReplaceAll(password, "'", "''"))
	_, err := runPSQL(query)
	if err != nil {
		return nil, err
	}
	return &ActionResult{Changed: true, Message: fmt.Sprintf("Created user %s", user)}, nil
}

// DropUser drops a user.
func DropUser(user string) (*ActionResult, error) {
	_, err := runPSQL(fmt.Sprintf("DROP USER %q", user))
	if err != nil {
		return nil, err
	}
	return &ActionResult{Changed: true, Message: fmt.Sprintf("Dropped user %s", user)}, nil
}

// Grant grants privileges on a database to a user.
func Grant(privileges string, database string, user string) (*ActionResult, error) {
	query := fmt.Sprintf("GRANT %s ON DATABASE %q TO %q", privileges, database, user)
	_, err := runPSQL(query)
	if err != nil {
		return nil, err
	}
	return &ActionResult{Changed: true, Message: fmt.Sprintf("Granted %s on %s to %s", privileges, database, user)}, nil
}
