// Package runner implements the OpsLang universal runner that executes
// JSON instruction packages from stdin and outputs JSON results to stdout.
package runner

import (
	"encoding/json"
	"fmt"

	"github.com/opslang/opslang/pkg/ops-core-sdk/file"
	opsjson "github.com/opslang/opslang/pkg/ops-core-sdk/json"
	opsnet "github.com/opslang/opslang/pkg/ops-core-sdk/net"
	"github.com/opslang/opslang/pkg/ops-core-sdk/process"
	"github.com/opslang/opslang/pkg/ops-core-sdk/service"
	"github.com/opslang/opslang/pkg/ops-core-sdk/sys"
	optime "github.com/opslang/opslang/pkg/ops-core-sdk/time"
)

// InstructionPackage represents the JSON input from stdin.
type InstructionPackage struct {
	Version      string        `json:"version"`
	TaskID       string        `json:"task_id"`
	DryRun       bool          `json:"dry_run"`
	Instructions []Instruction `json:"instructions"`
}

// Instruction represents a single operation to execute.
type Instruction struct {
	Op     string                 `json:"op"`
	Args   map[string]interface{} `json:"args"`
	Assign string                 `json:"assign,omitempty"`
}

// Output represents the JSON output to stdout.
type Output struct {
	Status   string                 `json:"status"`
	Data     map[string]interface{} `json:"data"`
	Errors   []string               `json:"errors"`
	Warnings []string               `json:"warnings"`
}

// OperationFunc is a function that executes an operation with the given arguments.
type OperationFunc func(args map[string]interface{}) (interface{}, error)

// Registry holds all registered operations.
type Registry struct {
	ops map[string]OperationFunc
}

// NewRegistry creates a new registry with all standard operations registered.
func NewRegistry() *Registry {
	r := &Registry{
		ops: make(map[string]OperationFunc),
	}
	r.registerAll()
	return r
}

// registerAll registers all standard library operations.
func (r *Registry) registerAll() {
	// sys operations
	r.Register("sys.cpu.usage", opCPUUsage)
	r.Register("sys.memory.info", opMemoryInfo)
	r.Register("sys.disk.usage", opDiskUsage)
	r.Register("sys.load.avg", opLoadAvg)
	r.Register("sys.hostname", opHostname)

	// file operations
	r.Register("file.read", opFileRead)
	r.Register("file.write", opFileWrite)
	r.Register("file.copy", opFileCopy)
	r.Register("file.move", opFileMove)
	r.Register("file.delete", opFileDelete)
	r.Register("file.exists", opFileExists)

	// net operations
	r.Register("net.http.get", opHTTPGet)
	r.Register("net.http.post", opHTTPPost)
	r.Register("net.tcp.check", opTCPCheck)
	r.Register("net.dns.resolve", opDNSResolve)

	// process operations
	r.Register("process.list", opProcessList)
	r.Register("process.find.byname", opProcessFindByName)
	r.Register("process.exec", opProcessExec)

	// service operations
	r.Register("service.status", opServiceStatus)
	r.Register("service.start", opServiceStart)
	r.Register("service.stop", opServiceStop)
	r.Register("service.restart", opServiceRestart)
	r.Register("service.enable", opServiceEnable)

	// time operations
	r.Register("time.now", opTimeNow)
	r.Register("time.format", opTimeFormat)

	// json operations
	r.Register("json.encode", opJSONEncode)
	r.Register("json.decode", opJSONDecode)

	// special operations
	r.Register("report", opReport)
}

// Register adds an operation to the registry.
func (r *Registry) Register(name string, fn OperationFunc) {
	r.ops[name] = fn
}

// Get retrieves an operation from the registry.
func (r *Registry) Get(name string) (OperationFunc, bool) {
	fn, ok := r.ops[name]
	return fn, ok
}

// Execute runs a single instruction and returns the result.
func (r *Registry) Execute(inst *Instruction, dryRun bool) (interface{}, error) {
	fn, ok := r.ops[inst.Op]
	if !ok {
		return nil, fmt.Errorf("unknown operation: %q", inst.Op)
	}

	if dryRun {
		return map[string]interface{}{
			"dry_run":   true,
			"operation": inst.Op,
			"args":      inst.Args,
		}, nil
	}

	return fn(inst.Args)
}

// Run processes an instruction package and returns the output.
func Run(pkg *InstructionPackage, registry *Registry) *Output {
	output := &Output{
		Status:   "ok",
		Data:     make(map[string]interface{}),
		Errors:   []string{},
		Warnings: []string{},
	}

	var lastReport map[string]interface{}

	for i, inst := range pkg.Instructions {
		result, err := registry.Execute(&inst, pkg.DryRun)
		if err != nil {
			output.Errors = append(output.Errors, fmt.Sprintf("instruction %d (%s): %v", i, inst.Op, err))
			continue
		}

		if inst.Assign != "" {
			output.Data[inst.Assign] = result
		}

		if inst.Op == "report" {
			if report, ok := result.(map[string]interface{}); ok {
				lastReport = report
			}
		}
	}

	if len(output.Errors) > 0 {
		output.Status = "partial"
	}

	if lastReport != nil {
		output.Data = lastReport
	}

	return output
}

// ============== Operation Implementations ==============

// sys operations
func opCPUUsage(_ map[string]interface{}) (interface{}, error) {
	return sys.GetCPUUsage()
}

func opMemoryInfo(_ map[string]interface{}) (interface{}, error) {
	return sys.GetMemoryInfo()
}

func opDiskUsage(args map[string]interface{}) (interface{}, error) {
	path := "/"
	if p, ok := args["path"].(string); ok {
		path = p
	}
	return sys.GetDiskUsage(path)
}

func opLoadAvg(_ map[string]interface{}) (interface{}, error) {
	return sys.GetLoadAvg()
}

func opHostname(_ map[string]interface{}) (interface{}, error) {
	return sys.Hostname()
}

// file operations
func opFileRead(args map[string]interface{}) (interface{}, error) {
	path, ok := args["path"].(string)
	if !ok {
		return nil, fmt.Errorf("path is required")
	}
	return file.Read(path)
}

func opFileWrite(args map[string]interface{}) (interface{}, error) {
	path, ok := args["path"].(string)
	if !ok {
		return nil, fmt.Errorf("path is required")
	}
	content, ok := args["content"].(string)
	if !ok {
		return nil, fmt.Errorf("content is required")
	}
	return file.Write(path, content)
}

func opFileCopy(args map[string]interface{}) (interface{}, error) {
	src, ok := args["src"].(string)
	if !ok {
		return nil, fmt.Errorf("src is required")
	}
	dst, ok := args["dst"].(string)
	if !ok {
		return nil, fmt.Errorf("dst is required")
	}
	return file.Copy(src, dst)
}

func opFileMove(args map[string]interface{}) (interface{}, error) {
	src, ok := args["src"].(string)
	if !ok {
		return nil, fmt.Errorf("src is required")
	}
	dst, ok := args["dst"].(string)
	if !ok {
		return nil, fmt.Errorf("dst is required")
	}
	return file.Move(src, dst)
}

func opFileDelete(args map[string]interface{}) (interface{}, error) {
	path, ok := args["path"].(string)
	if !ok {
		return nil, fmt.Errorf("path is required")
	}
	return file.Delete(path)
}

func opFileExists(args map[string]interface{}) (interface{}, error) {
	path, ok := args["path"].(string)
	if !ok {
		return nil, fmt.Errorf("path is required")
	}
	return file.Exists(path)
}

// net operations
func opHTTPGet(args map[string]interface{}) (interface{}, error) {
	url, ok := args["url"].(string)
	if !ok {
		return nil, fmt.Errorf("url is required")
	}
	return opsnet.HTTPGet(url)
}

func opHTTPPost(args map[string]interface{}) (interface{}, error) {
	url, ok := args["url"].(string)
	if !ok {
		return nil, fmt.Errorf("url is required")
	}
	body, ok := args["body"].(string)
	if !ok {
		return nil, fmt.Errorf("body is required")
	}
	return opsnet.HTTPPost(url, body)
}

func opTCPCheck(args map[string]interface{}) (interface{}, error) {
	host, ok := args["host"].(string)
	if !ok {
		return nil, fmt.Errorf("host is required")
	}
	portF, ok := args["port"].(float64)
	if !ok {
		return nil, fmt.Errorf("port is required")
	}
	return opsnet.TCPConnect(host, int(portF))
}

func opDNSResolve(args map[string]interface{}) (interface{}, error) {
	host, ok := args["host"].(string)
	if !ok {
		return nil, fmt.Errorf("host is required")
	}
	return opsnet.DNSLookup(host)
}

// process operations
func opProcessList(_ map[string]interface{}) (interface{}, error) {
	return process.List()
}

func opProcessFindByName(args map[string]interface{}) (interface{}, error) {
	name, ok := args["name"].(string)
	if !ok {
		return nil, fmt.Errorf("name is required")
	}
	return process.FindByName(name)
}

func opProcessExec(args map[string]interface{}) (interface{}, error) {
	command, ok := args["command"].(string)
	if !ok {
		return nil, fmt.Errorf("command is required")
	}
	var procArgs []string
	if a, ok := args["args"].([]interface{}); ok {
		for _, v := range a {
			if s, ok := v.(string); ok {
				procArgs = append(procArgs, s)
			}
		}
	}
	return process.Exec(command, procArgs)
}

// service operations
func opServiceStatus(args map[string]interface{}) (interface{}, error) {
	name, ok := args["name"].(string)
	if !ok {
		return nil, fmt.Errorf("name is required")
	}
	return service.Status(name)
}

func opServiceStart(args map[string]interface{}) (interface{}, error) {
	name, ok := args["name"].(string)
	if !ok {
		return nil, fmt.Errorf("name is required")
	}
	return service.Start(name)
}

func opServiceStop(args map[string]interface{}) (interface{}, error) {
	name, ok := args["name"].(string)
	if !ok {
		return nil, fmt.Errorf("name is required")
	}
	return service.Stop(name)
}

func opServiceRestart(args map[string]interface{}) (interface{}, error) {
	name, ok := args["name"].(string)
	if !ok {
		return nil, fmt.Errorf("name is required")
	}
	return service.Restart(name)
}

func opServiceEnable(args map[string]interface{}) (interface{}, error) {
	name, ok := args["name"].(string)
	if !ok {
		return nil, fmt.Errorf("name is required")
	}
	return service.Enable(name)
}

// time operations
func opTimeNow(_ map[string]interface{}) (interface{}, error) {
	return optime.Now(), nil
}

func opTimeFormat(args map[string]interface{}) (interface{}, error) {
	layout, ok := args["layout"].(string)
	if !ok {
		return nil, fmt.Errorf("layout is required")
	}
	unixF, ok := args["unix"].(float64)
	if !ok {
		return nil, fmt.Errorf("unix timestamp is required")
	}
	return optime.Format(int64(unixF), layout)
}

// json operations
func opJSONEncode(args map[string]interface{}) (interface{}, error) {
	data, ok := args["data"]
	if !ok {
		return nil, fmt.Errorf("data is required")
	}
	return opsjson.Encode(data)
}

func opJSONDecode(args map[string]interface{}) (interface{}, error) {
	input, ok := args["input"].(string)
	if !ok {
		return nil, fmt.Errorf("input is required")
	}
	return opsjson.Decode(input)
}

// report operation - collects results for final output
func opReport(args map[string]interface{}) (interface{}, error) {
	result := make(map[string]interface{})
	for key, value := range args {
		result[key] = value
	}
	return result, nil
}

// ListOperations returns the names of all registered operations.
func (r *Registry) ListOperations() []string {
	names := make([]string, 0, len(r.ops))
	for name := range r.ops {
		names = append(names, name)
	}
	return names
}

// ValidatePackage checks if an instruction package is valid.
func ValidatePackage(pkg *InstructionPackage) error {
	if pkg.Version == "" {
		return fmt.Errorf("version is required")
	}
	if pkg.Version != "1.0" {
		return fmt.Errorf("unsupported version: %s", pkg.Version)
	}
	if len(pkg.Instructions) == 0 {
		return fmt.Errorf("at least one instruction is required")
	}
	for i, inst := range pkg.Instructions {
		if inst.Op == "" {
			return fmt.Errorf("instruction %d: op is required", i)
		}
	}
	return nil
}

// OutputToJSON marshals an Output to JSON bytes.
func OutputToJSON(output *Output) ([]byte, error) {
	return json.MarshalIndent(output, "", "  ")
}
