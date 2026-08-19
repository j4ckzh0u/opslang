// Package mysql provides MySQL database management operations.
package mysql

import (
	"fmt"
	"os/exec"
	"strings"
)

// Database represents a MySQL database.
type Database struct {
	Name string `json:"name"`
}

// DatabasesResult represents the result of listing databases.
type DatabasesResult struct {
	Databases []Database `json:"databases"`
}

// User represents a MySQL user.
type User struct {
	User string `json:"user"`
	Host string `json:"host"`
}

// UsersResult represents the result of listing users.
type UsersResult struct {
	Users []User `json:"users"`
}

// ActionResult represents the result of a MySQL action.
type ActionResult struct {
	Changed bool   `json:"changed"`
	Message string `json:"message"`
}

func runMySQL(query string) (string, error) {
	cmd := exec.Command("mysql", "-N", "-e", query)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("mysql query failed: %w (output: %s)", err, string(out))
	}
	return strings.TrimSpace(string(out)), nil
}

// Databases returns all databases.
func Databases() (*DatabasesResult, error) {
	out, err := runMySQL("SHOW DATABASES")
	if err != nil {
		return nil, err
	}

	result := &DatabasesResult{Databases: make([]Database, 0)}
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line != "" && line != "information_schema" && line != "mysql" && line != "performance_schema" && line != "sys" {
			result.Databases = append(result.Databases, Database{Name: line})
		}
	}

	return result, nil
}

// CreateDatabase creates a new database.
func CreateDatabase(name string) (*ActionResult, error) {
	_, err := runMySQL(fmt.Sprintf("CREATE DATABASE `%s`", name))
	if err != nil {
		return nil, err
	}

	return &ActionResult{
		Changed: true,
		Message: fmt.Sprintf("Created database %s", name),
	}, nil
}

// DropDatabase drops a database.
func DropDatabase(name string) (*ActionResult, error) {
	_, err := runMySQL(fmt.Sprintf("DROP DATABASE `%s`", name))
	if err != nil {
		return nil, err
	}

	return &ActionResult{
		Changed: true,
		Message: fmt.Sprintf("Dropped database %s", name),
	}, nil
}

// Users returns all users.
func Users() (*UsersResult, error) {
	out, err := runMySQL("SELECT User, Host FROM mysql.user")
	if err != nil {
		return nil, err
	}

	result := &UsersResult{Users: make([]User, 0)}
	for _, line := range strings.Split(out, "\n") {
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			result.Users = append(result.Users, User{User: parts[0], Host: parts[1]})
		}
	}

	return result, nil
}

// CreateUser creates a new user.
func CreateUser(user string, host string, password string) (*ActionResult, error) {
	if host == "" {
		host = "localhost"
	}

	query := fmt.Sprintf("CREATE USER '%s'@'%s' IDENTIFIED BY '%s'", user, host, password)
	_, err := runMySQL(query)
	if err != nil {
		return nil, err
	}

	return &ActionResult{
		Changed: true,
		Message: fmt.Sprintf("Created user %s@%s", user, host),
	}, nil
}

// DropUser drops a user.
func DropUser(user string, host string) (*ActionResult, error) {
	if host == "" {
		host = "localhost"
	}

	query := fmt.Sprintf("DROP USER '%s'@'%s'", user, host)
	_, err := runMySQL(query)
	if err != nil {
		return nil, err
	}

	return &ActionResult{
		Changed: true,
		Message: fmt.Sprintf("Dropped user %s@%s", user, host),
	}, nil
}

// Grant grants privileges to a user.
func Grant(privileges string, database string, user string, host string) (*ActionResult, error) {
	if host == "" {
		host = "localhost"
	}

	query := fmt.Sprintf("GRANT %s ON `%s`.* TO '%s'@'%s'", privileges, database, user, host)
	_, err := runMySQL(query)
	if err != nil {
		return nil, err
	}

	return &ActionResult{
		Changed: true,
		Message: fmt.Sprintf("Granted %s on %s to %s@%s", privileges, database, user, host),
	}, nil
}
