// Package rabbitmq provides RabbitMQ management via rabbitmqctl / rabbitmqadmin.
package rabbitmq

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// VhostResult is returned by vhost operations.
type VhostResult struct {
	Name    string `json:"name"`
	Success bool   `json:"success"`
	Changed bool   `json:"changed"`
	Error   string `json:"error,omitempty"`
}

// UserResult is returned by user operations.
type UserResult struct {
	Name    string `json:"name"`
	Success bool   `json:"success"`
	Changed bool   `json:"changed"`
	Tags    string `json:"tags,omitempty"`
	Error   string `json:"error,omitempty"`
}

// PermissionResult is returned by permission operations.
type PermissionResult struct {
	User    string `json:"user"`
	Vhost   string `json:"vhost"`
	Success bool   `json:"success"`
	Changed bool   `json:"changed"`
	Error   string `json:"error,omitempty"`
}

// PolicyResult is returned by policy operations.
type PolicyResult struct {
	Name    string `json:"name"`
	Vhost   string `json:"vhost"`
	Success bool   `json:"success"`
	Changed bool   `json:"changed"`
	Error   string `json:"error,omitempty"`
}

// QueueResult is returned by queue operations.
type QueueResult struct {
	Name    string `json:"name"`
	Vhost   string `json:"vhost"`
	Success bool   `json:"success"`
	Changed bool   `json:"changed"`
	Error   string `json:"error,omitempty"`
}

// ExchangeResult is returned by exchange operations.
type ExchangeResult struct {
	Name    string `json:"name"`
	Vhost   string `json:"vhost"`
	Success bool   `json:"success"`
	Changed bool   `json:"changed"`
	Error   string `json:"error,omitempty"`
}

// BindingResult is returned by binding operations.
type BindingResult struct {
	Queue    string `json:"queue"`
	Exchange string `json:"exchange"`
	Vhost    string `json:"vhost"`
	Success  bool   `json:"success"`
	Changed  bool   `json:"changed"`
	Error    string `json:"error,omitempty"`
}

// StatusResult is returned by cluster status check.
type StatusResult struct {
	Node    string   `json:"node"`
	Running bool     `json:"running"`
	Partitions []string `json:"partitions,omitempty"`
	Error   string   `json:"error,omitempty"`
}

// rabbitmqctl runs a rabbitmqctl subcommand.
func rabbitmqctl(args ...string) (string, error) {
	cmd := exec.Command("rabbitmqctl", args...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// rabbitmqadmin runs a rabbitmqadmin subcommand.
func rabbitmqadmin(args ...string) (string, error) {
	cmd := exec.Command("rabbitmqadmin", args...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// AddVhost creates a virtual host.
func AddVhost(name string) VhostResult {
	if name == "" {
		return VhostResult{Error: "name is required"}
	}
	// Check if already exists.
	out, _ := rabbitmqctl("list_vhosts", "--quiet")
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		if strings.TrimSpace(line) == name {
			return VhostResult{Name: name, Success: true, Changed: false}
		}
	}
	out, err := rabbitmqctl("add_vhost", name)
	if err != nil {
		return VhostResult{Name: name, Error: fmt.Sprintf("add_vhost failed: %s: %s", err, out)}
	}
	return VhostResult{Name: name, Success: true, Changed: true}
}

// DeleteVhost removes a virtual host.
func DeleteVhost(name string) VhostResult {
	if name == "" {
		return VhostResult{Error: "name is required"}
	}
	out, err := rabbitmqctl("delete_vhost", name)
	if err != nil {
		return VhostResult{Name: name, Error: fmt.Sprintf("delete_vhost failed: %s: %s", err, out)}
	}
	return VhostResult{Name: name, Success: true, Changed: true}
}

// ListVhosts lists all virtual hosts.
func ListVhosts() ([]string, error) {
	out, err := rabbitmqctl("list_vhosts", "--quiet")
	if err != nil {
		return nil, fmt.Errorf("list_vhosts failed: %w: %s", err, out)
	}
	var vhosts []string
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			vhosts = append(vhosts, line)
		}
	}
	return vhosts, nil
}

// AddUser creates a user with a password.
func AddUser(name, password, tags string) UserResult {
	if name == "" {
		return UserResult{Error: "name is required"}
	}
	if password == "" {
		return UserResult{Error: "password is required"}
	}
	// Check if already exists.
	out, _ := rabbitmqctl("list_users", "--quiet")
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		fields := strings.Fields(line)
		if len(fields) > 0 && fields[0] == name {
			return UserResult{Name: name, Success: true, Changed: false, Tags: tags}
		}
	}
	out, err := rabbitmqctl("add_user", name, password)
	if err != nil {
		return UserResult{Name: name, Error: fmt.Sprintf("add_user failed: %s: %s", err, out)}
	}
	if tags != "" {
		rabbitmqctl("set_user_tags", name, tags)
	}
	return UserResult{Name: name, Success: true, Changed: true, Tags: tags}
}

// DeleteUser removes a user.
func DeleteUser(name string) UserResult {
	if name == "" {
		return UserResult{Error: "name is required"}
	}
	out, err := rabbitmqctl("delete_user", name)
	if err != nil {
		return UserResult{Name: name, Error: fmt.Sprintf("delete_user failed: %s: %s", err, out)}
	}
	return UserResult{Name: name, Success: true, Changed: true}
}

// SetUserTags sets tags on a user.
func SetUserTags(name, tags string) UserResult {
	if name == "" {
		return UserResult{Error: "name is required"}
	}
	out, err := rabbitmqctl("set_user_tags", name, tags)
	if err != nil {
		return UserResult{Name: name, Error: fmt.Sprintf("set_user_tags failed: %s: %s", err, out)}
	}
	return UserResult{Name: name, Success: true, Changed: true, Tags: tags}
}

// ListUsers lists all users.
func ListUsers() ([]string, error) {
	out, err := rabbitmqctl("list_users", "--quiet")
	if err != nil {
		return nil, fmt.Errorf("list_users failed: %w: %s", err, out)
	}
	var users []string
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		fields := strings.Fields(line)
		if len(fields) > 0 {
			users = append(users, fields[0])
		}
	}
	return users, nil
}

// SetPermission sets permissions for a user on a vhost.
func SetPermission(user, vhost, configure, write, read string) PermissionResult {
	if user == "" || vhost == "" {
		return PermissionResult{Error: "user and vhost are required"}
	}
	out, err := rabbitmqctl("set_permissions", "-p", vhost, user, configure, write, read)
	if err != nil {
		return PermissionResult{User: user, Vhost: vhost, Error: fmt.Sprintf("set_permissions failed: %s: %s", err, out)}
	}
	return PermissionResult{User: user, Vhost: vhost, Success: true, Changed: true}
}

// ClearPermission removes permissions for a user on a vhost.
func ClearPermission(user, vhost string) PermissionResult {
	if user == "" || vhost == "" {
		return PermissionResult{Error: "user and vhost are required"}
	}
	out, err := rabbitmqctl("clear_permissions", "-p", vhost, user)
	if err != nil {
		return PermissionResult{User: user, Vhost: vhost, Error: fmt.Sprintf("clear_permissions failed: %s: %s", err, out)}
	}
	return PermissionResult{User: user, Vhost: vhost, Success: true, Changed: true}
}

// SetPolicy creates or updates a policy.
func SetPolicy(name, vhost, pattern, definition, applyTo string) PolicyResult {
	if name == "" || vhost == "" {
		return PolicyResult{Error: "name and vhost are required"}
	}
	if pattern == "" {
		pattern = ".*"
	}
	if applyTo == "" {
		applyTo = "all"
	}
	// Validate definition is JSON.
	if definition != "" {
		var js json.RawMessage
		if err := json.Unmarshal([]byte(definition), &js); err != nil {
			return PolicyResult{Name: name, Vhost: vhost, Error: fmt.Sprintf("definition is not valid JSON: %v", err)}
		}
	}
	out, err := rabbitmqctl("set_policy", "-p", vhost, "--apply-to", applyTo, name, pattern, definition)
	if err != nil {
		return PolicyResult{Name: name, Vhost: vhost, Error: fmt.Sprintf("set_policy failed: %s: %s", err, out)}
	}
	return PolicyResult{Name: name, Vhost: vhost, Success: true, Changed: true}
}

// DeletePolicy removes a policy.
func DeletePolicy(name, vhost string) PolicyResult {
	if name == "" || vhost == "" {
		return PolicyResult{Error: "name and vhost are required"}
	}
	out, err := rabbitmqctl("clear_policy", "-p", vhost, name)
	if err != nil {
		return PolicyResult{Name: name, Vhost: vhost, Error: fmt.Sprintf("clear_policy failed: %s: %s", err, out)}
	}
	return PolicyResult{Name: name, Vhost: vhost, Success: true, Changed: true}
}

// DeclareQueue declares a queue via rabbitmqadmin.
func DeclareQueue(name, vhost, queueType string, durable, autoDelete bool) QueueResult {
	if name == "" {
		return QueueResult{Error: "name is required"}
	}
	if vhost == "" {
		vhost = "/"
	}
	args := []string{"declare", "queue", fmt.Sprintf("name=%s", name), fmt.Sprintf("vhost=%s", vhost)}
	if durable {
		args = append(args, "durable=true")
	}
	if autoDelete {
		args = append(args, "auto_delete=true")
	}
	if queueType != "" {
		args = append(args, fmt.Sprintf("arguments={\"x-queue-type\":\"%s\"}", queueType))
	}
	out, err := rabbitmqadmin(args...)
	if err != nil {
		return QueueResult{Name: name, Vhost: vhost, Error: fmt.Sprintf("declare queue failed: %s: %s", err, out)}
	}
	return QueueResult{Name: name, Vhost: vhost, Success: true, Changed: true}
}

// DeleteQueue removes a queue.
func DeleteQueue(name, vhost string) QueueResult {
	if name == "" {
		return QueueResult{Error: "name is required"}
	}
	if vhost == "" {
		vhost = "/"
	}
	out, err := rabbitmqadmin("delete", "queue", fmt.Sprintf("name=%s", name), fmt.Sprintf("vhost=%s", vhost))
	if err != nil {
		return QueueResult{Name: name, Vhost: vhost, Error: fmt.Sprintf("delete queue failed: %s: %s", err, out)}
	}
	return QueueResult{Name: name, Vhost: vhost, Success: true, Changed: true}
}

// DeclareExchange declares an exchange.
func DeclareExchange(name, vhost, exType string, durable, autoDelete bool) ExchangeResult {
	if name == "" {
		return ExchangeResult{Error: "name is required"}
	}
	if vhost == "" {
		vhost = "/"
	}
	if exType == "" {
		exType = "direct"
	}
	args := []string{"declare", "exchange", fmt.Sprintf("name=%s", name), fmt.Sprintf("vhost=%s", vhost), fmt.Sprintf("type=%s", exType)}
	if durable {
		args = append(args, "durable=true")
	}
	if autoDelete {
		args = append(args, "auto_delete=true")
	}
	out, err := rabbitmqadmin(args...)
	if err != nil {
		return ExchangeResult{Name: name, Vhost: vhost, Error: fmt.Sprintf("declare exchange failed: %s: %s", err, out)}
	}
	return ExchangeResult{Name: name, Vhost: vhost, Success: true, Changed: true}
}

// DeleteExchange removes an exchange.
func DeleteExchange(name, vhost string) ExchangeResult {
	if name == "" {
		return ExchangeResult{Error: "name is required"}
	}
	if vhost == "" {
		vhost = "/"
	}
	out, err := rabbitmqadmin("delete", "exchange", fmt.Sprintf("name=%s", name), fmt.Sprintf("vhost=%s", vhost))
	if err != nil {
		return ExchangeResult{Name: name, Vhost: vhost, Error: fmt.Sprintf("delete exchange failed: %s: %s", err, out)}
	}
	return ExchangeResult{Name: name, Vhost: vhost, Success: true, Changed: true}
}

// BindQueue binds a queue to an exchange.
func BindQueue(queue, exchange, vhost, routingKey string) BindingResult {
	if queue == "" || exchange == "" {
		return BindingResult{Error: "queue and exchange are required"}
	}
	if vhost == "" {
		vhost = "/"
	}
	args := []string{"declare", "binding",
		fmt.Sprintf("source=%s", exchange),
		fmt.Sprintf("destination=%s", queue),
		fmt.Sprintf("vhost=%s", vhost),
		fmt.Sprintf("routing_key=%s", routingKey),
	}
	out, err := rabbitmqadmin(args...)
	if err != nil {
		return BindingResult{Queue: queue, Exchange: exchange, Vhost: vhost, Error: fmt.Sprintf("bind failed: %s: %s", err, out)}
	}
	return BindingResult{Queue: queue, Exchange: exchange, Vhost: vhost, Success: true, Changed: true}
}

// UnbindQueue removes a binding.
func UnbindQueue(queue, exchange, vhost, routingKey string) BindingResult {
	if queue == "" || exchange == "" {
		return BindingResult{Error: "queue and exchange are required"}
	}
	if vhost == "" {
		vhost = "/"
	}
	args := []string{"delete", "binding",
		fmt.Sprintf("source=%s", exchange),
		fmt.Sprintf("destination=%s", queue),
		fmt.Sprintf("vhost=%s", vhost),
		fmt.Sprintf("properties_key=%s", routingKey),
	}
	out, err := rabbitmqadmin(args...)
	if err != nil {
		return BindingResult{Queue: queue, Exchange: exchange, Vhost: vhost, Error: fmt.Sprintf("unbind failed: %s: %s", err, out)}
	}
	return BindingResult{Queue: queue, Exchange: exchange, Vhost: vhost, Success: true, Changed: true}
}

// GetStatus returns cluster status.
func GetStatus() StatusResult {
	out, err := rabbitmqctl("status")
	if err != nil {
		return StatusResult{Error: fmt.Sprintf("status failed: %s: %s", err, out)}
	}
	// Simple check: look for node name in output.
	node := "unknown"
	for _, line := range strings.Split(out, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "Runtime ") || strings.HasPrefix(trimmed, "Node ") {
			node = trimmed
			break
		}
	}
	return StatusResult{Node: node, Running: true}
}
