// Package compiler implements the AOT compilation pipeline for OpsLang.
// It translates an AST into Go source code and compiles it into a static binary.
package compiler

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/opslang/opslang/internal/ast"
)

// sdkMapping maps OpsLang dotted function names to Go SDK calls. Names
// follow the canonical table in internal/opsspec; historical aliases are
// resolved through sdkAliases.
var sdkMapping = map[string]sdkFunc{
	// sys
	"sys.cpu.usage":       {pkg: "sys", goName: "GetCPUUsage"},
	"sys.cpu.count":       {pkg: "sys", goName: "GetCPUCount"},
	"sys.cpu.info":        {pkg: "sys", goName: "GetCPUInfo"},
	"sys.memory.info":     {pkg: "sys", goName: "GetMemoryInfo"},
	"sys.disk.usage":      {pkg: "sys", goName: "GetDiskUsage", args: true, params: []string{"s"}},
	"sys.disk.partitions": {pkg: "sys", goName: "GetDiskPartitions"},
	"sys.os":              {pkg: "sys", goName: "GetHostInfo"},
	"sys.hostname":        {pkg: "sys", goName: "Hostname"},
	"sys.load":            {pkg: "sys", goName: "GetLoadAvg"},
	"sys.net.interfaces":  {pkg: "sys", goName: "GetNetInterfaces"},
	"sys.users":           {pkg: "sys", goName: "Users"},
	"sys.uptime":          {pkg: "sys", goName: "Uptime"},

	// file
	"file.read":     {pkg: "file", goName: "Read", args: true, params: []string{"s"}},
	"file.write":    {pkg: "file", goName: "Write", args: true, params: []string{"s", "s"}},
	"file.append":   {pkg: "file", goName: "Append", args: true, params: []string{"s", "s"}},
	"file.exists":   {pkg: "file", goName: "Exists", args: true, params: []string{"s"}},
	"file.copy":     {pkg: "file", goName: "Copy", args: true, params: []string{"s", "s"}},
	"file.move":     {pkg: "file", goName: "Move", args: true, params: []string{"s", "s"}},
	"file.delete":   {pkg: "file", goName: "Delete", args: true, params: []string{"s"}},
	"file.stat":     {pkg: "file", goName: "Stat", args: true, params: []string{"s"}},
	"file.list":     {pkg: "file", goName: "List", args: true, params: []string{"s"}},
	"file.mkdir":    {pkg: "file", goName: "Mkdir", args: true, params: []string{"s"}},
	"file.chmod":    {pkg: "file", goName: "Chmod", args: true, params: []string{"s", "m"}},
	"file.checksum": {pkg: "file", goName: "Checksum", args: true, params: []string{"s", "s"}},
	"file.template": {pkg: "file", goName: "Template", args: true, params: []string{"s", "d"}},

	// net
	"net.http_get":   {pkg: "net", goName: "HTTPGet", args: true, params: []string{"s"}},
	"net.http_post":  {pkg: "net", goName: "HTTPPost", args: true, params: []string{"s", "s"}},
	"net.tcp_check":  {pkg: "net", goName: "TCPConnect", args: true, params: []string{"s", "i"}},
	"net.dns_lookup": {pkg: "net", goName: "DNSLookup", args: true, params: []string{"s"}},
	"net.interfaces": {pkg: "net", goName: "Interfaces"},

	// process
	"process.list":         {pkg: "process", goName: "List"},
	"process.find_by_name": {pkg: "process", goName: "FindByName", args: true, params: []string{"s"}},
	"process.find_by_port": {pkg: "process", goName: "FindByPort", args: true, params: []string{"i"}},
	"process.kill":         {pkg: "process", goName: "Kill", args: true, params: []string{"i", "s"}},
	"process.exec":         {pkg: "process", goName: "Exec", args: true, params: []string{"s", "l"}},

	// service
	"service.status":  {pkg: "service", goName: "Status", args: true, params: []string{"s"}},
	"service.start":   {pkg: "service", goName: "Start", args: true, params: []string{"s"}},
	"service.stop":    {pkg: "service", goName: "Stop", args: true, params: []string{"s"}},
	"service.restart": {pkg: "service", goName: "Restart", args: true, params: []string{"s"}},
	"service.enable":  {pkg: "service", goName: "Enable", args: true, params: []string{"s"}},
	"service.disable": {pkg: "service", goName: "Disable", args: true, params: []string{"s"}},

	// selinux
	"selinux.get": {pkg: "selinux", goName: "Get"},
	"selinux.set": {pkg: "selinux", goName: "Set", args: true, params: []string{"s"}},

	// pkg
	"pkg.install": {pkg: "pkg", goName: "Install", args: true, params: []string{"s"}},
	"pkg.remove":  {pkg: "pkg", goName: "Remove", args: true, params: []string{"s"}},
	"pkg.info":    {pkg: "pkg", goName: "Info", args: true, params: []string{"s"}},
	"pkg.list":    {pkg: "pkg", goName: "List"},

	// ntp
	"ntp.get": {pkg: "ntp", goName: "Get"},
	"ntp.set": {pkg: "ntp", goName: "Set", args: true, params: []string{"s"}},

	// time
	"time.now":    {pkg: "time", goName: "Now", noErr: true},
	"time.format": {pkg: "time", goName: "Format", args: true, params: []string{"i64", "s"}},
	"time.parse":  {pkg: "time", goName: "Parse", args: true, params: []string{"s", "s"}},
	"time.since":  {pkg: "time", goName: "Since", args: true, params: []string{"i64"}},
	"time.sleep":  {pkg: "time", goName: "Sleep", args: true, params: []string{"i"}},
	"time.diff":   {pkg: "time", goName: "Diff", args: true, params: []string{"i64", "i64"}},

	// json / yaml
	"json.encode": {pkg: "json", goName: "Encode", args: true, params: []string{"a"}},
	"json.decode": {pkg: "json", goName: "Decode", args: true, params: []string{"s"}},
	"yaml.encode": {pkg: "yaml", goName: "Encode", args: true, params: []string{"a"}},
	"yaml.decode": {pkg: "yaml", goName: "Decode", args: true, params: []string{"s"}},

	// git
	"git.clone": {pkg: "git", goName: "Clone", args: true, params: []string{"s", "s", "ms"}},
	"git.pull":  {pkg: "git", goName: "Pull", args: true, params: []string{"s", "s", "s"}},

	// user
	"user.info":    {pkg: "user", goName: "Info", args: true, params: []string{"s"}},
	"user.list":    {pkg: "user", goName: "List"},
	"user.add":     {pkg: "user", goName: "Add", args: true, params: []string{"s", "ms"}},
	"user.remove":  {pkg: "user", goName: "Remove", args: true, params: []string{"s", "b"}},
	"user.modify":  {pkg: "user", goName: "Modify", args: true, params: []string{"s", "ms"}},
	"user.exists":  {pkg: "user", goName: "Exists", args: true, params: []string{"s"}},

	// group
	"group.info":   {pkg: "group", goName: "Info", args: true, params: []string{"s"}},
	"group.list":   {pkg: "group", goName: "List"},
	"group.add":    {pkg: "group", goName: "Add", args: true, params: []string{"s", "ms"}},
	"group.remove": {pkg: "group", goName: "Remove", args: true, params: []string{"s"}},
	"group.exists": {pkg: "group", goName: "Exists", args: true, params: []string{"s"}},

	// cron
	"cron.list":   {pkg: "cron", goName: "List", args: true, params: []string{"s"}},
	"cron.add":    {pkg: "cron", goName: "Add", args: true, params: []string{"s", "entry"}},
	"cron.remove": {pkg: "cron", goName: "Remove", args: true, params: []string{"s", "s"}},

	// sysctl
	"sysctl.get":  {pkg: "sysctl", goName: "Get", args: true, params: []string{"s"}},
	"sysctl.set":  {pkg: "sysctl", goName: "Set", args: true, params: []string{"s", "s"}},
	"sysctl.list": {pkg: "sysctl", goName: "List"},

	// file extensions
	"file.lineinfile": {pkg: "file", goName: "LineInFile", args: true, params: []string{"s", "s", "b", "s"}},

	// net extensions
	"net.wait_for": {pkg: "net", goName: "WaitFor", args: true, params: []string{"s", "i", "i"}},

	// sys extensions
	"sys.hostname_set": {pkg: "sys", goName: "HostnameSet", args: true, params: []string{"s"}},
	"sys.mount":        {pkg: "sys", goName: "Mount", args: true, params: []string{"s", "s", "s", "ms"}},
	"sys.unmount":      {pkg: "sys", goName: "Unmount", args: true, params: []string{"s"}},
	"sys.list_mounts":  {pkg: "sys", goName: "ListMounts"},

	// firewall
	"firewall.rule": {pkg: "sys", goName: "FirewallRule", args: true, params: []string{"s", "s", "i", "s"}},

	// firewalld
	"firewalld.get":        {pkg: "firewalld", goName: "Get"},
	"firewalld.start":      {pkg: "firewalld", goName: "Start"},
	"firewalld.stop":       {pkg: "firewalld", goName: "Stop"},
	"firewalld.restart":    {pkg: "firewalld", goName: "Restart"},
	"firewalld.enable":     {pkg: "firewalld", goName: "Enable"},
	"firewalld.disable":    {pkg: "firewalld", goName: "Disable"},
	"firewalld.list_zones": {pkg: "firewalld", goName: "ListZones"},
	"firewalld.reload":     {pkg: "firewalld", goName: "Reload"},

	// archive
	"archive.create":  {pkg: "archive", goName: "Create", args: true, params: []string{"s", "l"}},
	"archive.extract": {pkg: "archive", goName: "Extract", args: true, params: []string{"s", "s"}},

	// disk
	"disk.filesystem": {pkg: "disk", goName: "FilesystemCreate", args: true, params: []string{"s", "s"}},
	"disk.part_list":  {pkg: "disk", goName: "PartList", args: true, params: []string{"s"}},

	// kernel
	"kernel.module_list":   {pkg: "kernel", goName: "ModuleList"},
	"kernel.module_load":   {pkg: "kernel", goName: "ModuleLoad", args: true, params: []string{"s"}},
	"kernel.module_unload": {pkg: "kernel", goName: "ModuleUnload", args: true, params: []string{"s"}},

	// known_hosts
	"known_hosts.list":   {pkg: "known_hosts", goName: "List"},
	"known_hosts.check":  {pkg: "known_hosts", goName: "Check", args: true, params: []string{"s"}},
	"known_hosts.add":    {pkg: "known_hosts", goName: "Add", args: true, params: []string{"s"}},
	"known_hosts.remove": {pkg: "known_hosts", goName: "Remove", args: true, params: []string{"s"}},

	// limits
	"limits.list":   {pkg: "limits", goName: "List"},
	"limits.get":    {pkg: "limits", goName: "Get", args: true, params: []string{"s"}},
	"limits.set":    {pkg: "limits", goName: "Set", args: true, params: []string{"s", "s", "s", "s"}},
	"limits.remove": {pkg: "limits", goName: "Remove", args: true, params: []string{"s"}},

	// ssh
	"ssh.authorized_key_add":    {pkg: "ssh", goName: "AuthorizedKeyAdd", args: true, params: []string{"s", "s", "b"}},
	"ssh.authorized_key_remove": {pkg: "ssh", goName: "AuthorizedKeyRemove", args: true, params: []string{"s", "s"}},
	"ssh.authorized_key_list":   {pkg: "ssh", goName: "AuthorizedKeyList", args: true, params: []string{"s"}},

	// docker
	"docker.container_list":   {pkg: "docker", goName: "ContainerList", args: true, params: []string{"b"}},
	"docker.container_exists": {pkg: "docker", goName: "ContainerExists", args: true, params: []string{"s"}},
	"docker.container_run":    {pkg: "docker", goName: "ContainerRun", args: true, params: []string{"s", "s", "ms"}},
	"docker.container_stop":   {pkg: "docker", goName: "ContainerStop", args: true, params: []string{"s"}},
	"docker.container_remove": {pkg: "docker", goName: "ContainerRemove", args: true, params: []string{"s", "b"}},
	"docker.image_list":       {pkg: "docker", goName: "ImageList"},
	"docker.image_pull":       {pkg: "docker", goName: "ImagePull", args: true, params: []string{"s"}},
	"docker.image_remove":     {pkg: "docker", goName: "ImageRemove", args: true, params: []string{"s", "b"}},

	// hosts
	"hosts.list":   {pkg: "hosts", goName: "List"},
	"hosts.exists": {pkg: "hosts", goName: "Exists", args: true, params: []string{"s"}},
	"hosts.add":    {pkg: "hosts", goName: "Add", args: true, params: []string{"s", "l"}},
	"hosts.remove": {pkg: "hosts", goName: "Remove", args: true, params: []string{"l"}},

	// locale
	"locale.get":       {pkg: "locale", goName: "Get"},
	"locale.available": {pkg: "locale", goName: "Available"},
	"locale.set":       {pkg: "locale", goName: "Set", args: true, params: []string{"s"}},

	// pip
	"pip.list":     {pkg: "pip", goName: "List"},
	"pip.exists":   {pkg: "pip", goName: "Exists", args: true, params: []string{"s"}},
	"pip.install":  {pkg: "pip", goName: "Install", args: true, params: []string{"s", "s"}},
	"pip.uninstall": {pkg: "pip", goName: "Uninstall", args: true, params: []string{"s"}},

	// apt_repo
	"apt_repo.list":   {pkg: "apt_repo", goName: "List"},
	"apt_repo.exists": {pkg: "apt_repo", goName: "Exists", args: true, params: []string{"s"}},
	"apt_repo.add":    {pkg: "apt_repo", goName: "Add", args: true, params: []string{"s", "s", "s"}},
	"apt_repo.remove": {pkg: "apt_repo", goName: "Remove", args: true, params: []string{"s"}},
	"apt_repo.update": {pkg: "apt_repo", goName: "Update"},

	// logrotate
	"logrotate.list":   {pkg: "logrotate", goName: "List"},
	"logrotate.get":    {pkg: "logrotate", goName: "Get", args: true, params: []string{"s"}},
	"logrotate.set":    {pkg: "logrotate", goName: "Set", args: true, params: []string{"s", "s", "s", "i", "b", "s"}},
	"logrotate.remove": {pkg: "logrotate", goName: "Remove", args: true, params: []string{"s"}},

	// resolv
	"resolv.get":              {pkg: "resolv", goName: "Get"},
	"resolv.set":              {pkg: "resolv", goName: "Set", args: true, params: []string{"l", "l", "l", "s"}},
	"resolv.add_nameserver":   {pkg: "resolv", goName: "AddNameserver", args: true, params: []string{"s"}},
	"resolv.remove_nameserver": {pkg: "resolv", goName: "RemoveNameserver", args: true, params: []string{"s"}},

	// yum_repo
	"yum_repo.list":   {pkg: "yum_repo", goName: "List"},
	"yum_repo.exists": {pkg: "yum_repo", goName: "Exists", args: true, params: []string{"s"}},
	"yum_repo.add":    {pkg: "yum_repo", goName: "Add", args: true, params: []string{"s", "s", "s", "b", "s"}},
	"yum_repo.remove": {pkg: "yum_repo", goName: "Remove", args: true, params: []string{"s"}},

	// file extensions
	"file.find":        {pkg: "file", goName: "FindFromArgs", args: true, params: []string{"l", "l", "s", "s", "i", "i64", "i64"}},
	"file.replace":     {pkg: "file", goName: "Replace", args: true, params: []string{"s", "s", "s", "s", "s"}},
	"file.blockinfile": {pkg: "file", goName: "BlockInFile", args: true, params: []string{"s", "s", "s", "b", "s", "s"}},
	"file.ini_get":     {pkg: "file", goName: "IniGet", args: true, params: []string{"s", "s", "s"}},
	"file.ini_set":     {pkg: "file", goName: "IniSet", args: true, params: []string{"s", "s", "s", "s"}},

	// net extensions
	"net.download":             {pkg: "net", goName: "Download", args: true, params: []string{"s", "s", "s", "s"}},
	"net.wait_for_connection":  {pkg: "net", goName: "WaitForConnection", args: true, params: []string{"s", "i", "i"}},

	// sys extensions
	"sys.timezone_get": {pkg: "sys", goName: "TimezoneGet"},
	"sys.timezone_set": {pkg: "sys", goName: "TimezoneSet", args: true, params: []string{"s"}},
	"sys.reboot":       {pkg: "sys", goName: "Reboot"},
}

// SDKMappingNames returns every canonical function name the code generator
// can translate. Used by cross-engine consistency tests.
func SDKMappingNames() []string {
	names := make([]string, 0, len(sdkMapping))
	for name := range sdkMapping {
		names = append(names, name)
	}
	return names
}

// sdkFunc describes how an OpsLang SDK call maps to Go.
type sdkFunc struct {
	pkg    string // short package key (e.g. "sys", "net")
	goName string // Go function name without package prefix
	args   bool   // whether the function takes arguments
	// params declares per-argument converters from dynamic interface{}
	// values to the Go parameter type. Codes:
	//   "s"  -> string          via opsStr
	//   "i"  -> int             via int(toFloat(..))
	//   "i64"-> int64           via int64(toFloat(..))
	//   "m"  -> uint32 file mode via opsParseMode
	//   "d"  -> map[string]interface{} via opsToMap
	//   "ms" -> map[string]string      via opsToStrMap
	//   "l"  -> []string        via opsStrList
	//   "a"  -> interface{}     as-is
	// A nil params (with args true) means all-string convention and is
	// only allowed where the SDK really takes strings.
	params []string
	// noErr: the SDK function returns only a value (no error), e.g.
	// time.Now(). Generated code must not unpack two returns from it.
	noErr bool
}

// generateSDKCall renders the converted Go call for an SDK function.
func (f sdkFunc) generateSDKCall(alias string, argExprs []string) string {
	callArgs := make([]string, len(argExprs))
	for i, a := range argExprs {
		var conv string
		if i < len(f.params) {
			conv = f.params[i]
		} else {
			conv = "a"
		}
		callArgs[i] = convertArg(a, conv)
	}
	return fmt.Sprintf("%s.%s(%s)", alias, f.goName, strings.Join(callArgs, ", "))
}

// convertArg wraps an interface{} expression into the target Go type.
func convertArg(expr, conv string) string {
	switch conv {
	case "s":
		return fmt.Sprintf("opsStr(%s)", expr)
	case "i":
		return fmt.Sprintf("int(toFloat(%s))", expr)
	case "i64":
		return fmt.Sprintf("int64(toFloat(%s))", expr)
	case "m":
		return fmt.Sprintf("opsParseMode(%s)", expr)
	case "d":
		return fmt.Sprintf("opsToMap(%s)", expr)
	case "ms":
		return fmt.Sprintf("opsToStrMap(%s)", expr)
	case "l":
		return fmt.Sprintf("opsStrList(%s)", expr)
	case "b":
		return fmt.Sprintf("opsBool(%s)", expr)
	case "entry":
		return fmt.Sprintf("opsToCronEntry(%s)", expr)
	default: // "a"
		return expr
	}
}

// pkgImportAlias maps our short package key to the import alias used in generated Go code.
var pkgImportAlias = map[string]string{
	"sys":     "sys",
	"file":    "file",
	"net":     "opsnet",
	"process": "process",
	"service": "service",
	"pkg":     "opspkg",
	"time":    "opstime",
	"json":    "opsjson",
	"yaml":    "opsyaml",
	"git":     "opsgit",
	"user":    "opsuser",
	"group":   "opsgrp",
	"cron":    "opscron",
	"sysctl":  "opsysctl",
	"archive": "opsarchive",
	"disk":    "opsdisk",
	"kernel":  "opskernel",
	"ssh":     "opsssh",
}

// pkgImportPath maps our short package key to the full import path.
var pkgImportPath = map[string]string{
	"sys":     "github.com/opslang/opslang/pkg/ops-core-sdk/sys",
	"file":    "github.com/opslang/opslang/pkg/ops-core-sdk/file",
	"net":     "github.com/opslang/opslang/pkg/ops-core-sdk/net",
	"process": "github.com/opslang/opslang/pkg/ops-core-sdk/process",
	"service": "github.com/opslang/opslang/pkg/ops-core-sdk/service",
	"pkg":     "github.com/opslang/opslang/pkg/ops-core-sdk/pkg",
	"time":    "github.com/opslang/opslang/pkg/ops-core-sdk/time",
	"json":    "github.com/opslang/opslang/pkg/ops-core-sdk/json",
	"yaml":    "github.com/opslang/opslang/pkg/ops-core-sdk/yaml",
	"git":     "github.com/opslang/opslang/pkg/ops-core-sdk/git",
	"user":    "github.com/opslang/opslang/pkg/ops-core-sdk/user",
	"group":   "github.com/opslang/opslang/pkg/ops-core-sdk/group",
	"cron":    "github.com/opslang/opslang/pkg/ops-core-sdk/cron",
	"sysctl":  "github.com/opslang/opslang/pkg/ops-core-sdk/sysctl",
	"archive": "github.com/opslang/opslang/pkg/ops-core-sdk/archive",
	"disk":    "github.com/opslang/opslang/pkg/ops-core-sdk/disk",
	"kernel":  "github.com/opslang/opslang/pkg/ops-core-sdk/kernel",
	"ssh":     "github.com/opslang/opslang/pkg/ops-core-sdk/ssh",
}

// CodeGenerator translates an AST Program into Go source code.
type CodeGenerator struct {
	indent    int
	buf       strings.Builder
	usedSDK   map[string]bool // tracks which SDK packages are used
	useOS     bool            // whether "os" is needed
	useSync   bool            // whether "sync" is needed (for parallel blocks)
	userFuncs []userFunc      // user-defined functions collected during generation
}

type userFunc struct {
	name   string
	params []ast.Parameter
	body   *ast.BlockStatement
}

// Generate takes an AST Program and returns a complete Go source string.
func (g *CodeGenerator) Generate(prog *ast.Program) (string, error) {
	// Compile-time privilege enforcement: reject mutating calls in scripts
	// whose declared privilege (default read_only) does not allow them,
	// before generating any code.
	if err := CheckPrivileges(prog); err != nil {
		return "", err
	}

	g.usedSDK = make(map[string]bool)
	g.userFuncs = nil
	g.useOS = false
	g.useSync = false

	// First pass: collect user-defined functions
	for _, stmt := range prog.Statements {
		if fn, ok := stmt.(*ast.FnStatement); ok {
			g.userFuncs = append(g.userFuncs, userFunc{
				name:   fn.Name.Name,
				params: fn.Params,
				body:   fn.Body,
			})
		}
	}

	// Second pass: generate main body into a temp buffer
	var mainBody strings.Builder
	savedBuf := g.buf
	savedIndent := g.indent
	g.buf = mainBody
	g.indent = 1

	for _, stmt := range prog.Statements {
		if _, ok := stmt.(*ast.FnStatement); ok {
			continue // already collected
		}
		if err := g.genStatement(stmt); err != nil {
			return "", err
		}
	}

	mainCode := g.buf.String()
	g.buf = savedBuf
	g.indent = savedIndent

	// Assemble the full file
	return g.assemble(mainCode)
}

// assemble builds the complete Go source file with imports, helpers, and main.
func (g *CodeGenerator) assemble(mainCode string) (string, error) {
	var b strings.Builder

	b.WriteString("// Code generated by OpsLang AOT compiler. DO NOT EDIT.\n")
	b.WriteString("package main\n\n")

	// Collect all imports
	b.WriteString("import (\n")
	// Standard library imports (always needed by helpers)
	b.WriteString("\t\"encoding/json\"\n")
	b.WriteString("\t\"fmt\"\n")
	b.WriteString("\t\"os\"\n")
	b.WriteString("\t\"sort\"\n")
	b.WriteString("\t\"strings\"\n")
	b.WriteString("\t\"sync\"\n")

	// SDK imports
	sdkOrder := []string{"sys", "file", "net", "process", "service", "pkg", "time", "json", "yaml", "git", "user", "group", "cron", "sysctl"}
	for _, pkg := range sdkOrder {
		if g.usedSDK[pkg] {
			alias := pkgImportAlias[pkg]
			path := pkgImportPath[pkg]
			b.WriteString(fmt.Sprintf("\t%s %q\n", alias, path))
		}
	}
	b.WriteString(")\n\n")

	// Suppress unused import warnings for always-imported stdlib packages
	b.WriteString("// Ensure standard library imports are used.\n")
	b.WriteString("var (\n")
	b.WriteString("\t_ = fmt.Println\n")
	b.WriteString("\t_ = json.Marshal\n")
	b.WriteString("\t_ = os.Stderr\n")
	b.WriteString("\t_ = sort.Strings\n")
	b.WriteString("\t_ = strings.Join\n")
	b.WriteString("\t_ = sync.Mutex{}\n")
	b.WriteString(")\n\n")

	// Runtime helpers
	g.writeHelpers(&b)

	// User-defined functions
	for _, fn := range g.userFuncs {
		if err := g.writeUserFunc(&b, fn); err != nil {
			return "", err
		}
	}

	// Main function
	b.WriteString("func main() {\n")
	b.WriteString("\t_output := make(map[string]interface{})\n")
	b.WriteString("\tvar _outputMu sync.Mutex\n")
	b.WriteString("\t_ = _output\n")
	b.WriteString("\t_ = _outputMu\n")
	b.WriteString(mainCode)
	b.WriteString("\n\t// Print final output as JSON\n")
	b.WriteString("\tif len(_output) > 0 {\n")
	b.WriteString("\t\tdata, _ := json.MarshalIndent(_output, \"\", \"  \")\n")
	b.WriteString("\t\tfmt.Println(string(data))\n")
	b.WriteString("\t}\n")
	b.WriteString("}\n")

	return b.String(), nil
}

// writeHelpers emits runtime helper functions used by generated code.
func (g *CodeGenerator) writeHelpers(b *strings.Builder) {
	b.WriteString(`func toFloat(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case int64:
		return float64(val)
	case int:
		return float64(val)
	case string:
		var f float64
		fmt.Sscanf(val, "%f", &f)
		return f
	case bool:
		if val {
			return 1.0
		}
		return 0.0
	case nil:
		return 0.0
	default:
		return 0.0
	}
}

func toInt(v interface{}) int64 {
	return int64(toFloat(v))
}

func formatValue(v interface{}) string {
	if v == nil {
		return "nil"
	}
	switch val := v.(type) {
	case bool:
		if val {
			return "true"
		}
		return "false"
	case int64:
		return fmt.Sprintf("%d", val)
	case float64:
		return fmt.Sprintf("%g", val)
	case string:
		return val
	case []interface{}:
		parts := make([]string, len(val))
		for i, elem := range val {
			parts[i] = formatValue(elem)
		}
		return "[" + strings.Join(parts, ", ") + "]"
	case map[string]interface{}:
		data, _ := json.Marshal(val)
		return string(data)
	default:
		return fmt.Sprintf("%v", val)
	}
}

func isTruthy(v interface{}) bool {
	if v == nil {
		return false
	}
	switch val := v.(type) {
	case bool:
		return val
	case int64:
		return val != 0
	case float64:
		return val != 0
	case string:
		return val != ""
	case []interface{}:
		return len(val) > 0
	case map[string]interface{}:
		return len(val) > 0
	default:
		return true
	}
}

// setOutput writes to the shared output map under lock: parallel blocks
// run goroutines that emit reports concurrently.
func setOutput(m *sync.Mutex, output map[string]interface{}, key string, value interface{}) {
	m.Lock()
	output[key] = value
	m.Unlock()
}

func opsStr(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return formatValue(v)
}

func opsParseMode(v interface{}) uint32 {
	var m uint64
	fmt.Sscanf(opsStr(v), "%o", &m)
	return uint32(m)
}

func opsToMap(v interface{}) map[string]interface{} {
	if m, ok := v.(map[string]interface{}); ok {
		return m
	}
	return map[string]interface{}{}
}

// opsNormalize converts SDK typed values (structs, typed slices) into
// generic interface{} shapes via a JSON round-trip. DSL list indexing and
// len() operate on []interface{}; a []ProcessInfo failed both silently.
func opsNormalize(v interface{}) interface{} {
	switch v.(type) {
	case nil, bool, string, int64, float64, int, []interface{}, map[string]interface{}:
		return v
	default:
		data, err := json.Marshal(v)
		if err != nil {
			return v
		}
		var out interface{}
		if json.Unmarshal(data, &out) == nil {
			return out
		}
		return v
	}
}

// opsLen is len() over normalized dynamic values.
func opsLen(v interface{}) int64 {
	switch c := opsNormalize(v).(type) {
	case string:
		return int64(len(c))
	case []interface{}:
		return int64(len(c))
	case map[string]interface{}:
		return int64(len(c))
	default:
		return int64(0)
	}
}

// opsToMapDeep converts SDK result structs into generic maps via a JSON
// round-trip, mirroring the interpreter structToMap. Without this,
// member access on a typed struct silently evaluated to nil.
func opsToMapDeep(v interface{}) map[string]interface{} {
	if m, ok := v.(map[string]interface{}); ok {
		return m
	}
	data, err := json.Marshal(v)
	if err != nil {
		return map[string]interface{}{}
	}
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		return map[string]interface{}{}
	}
	return m
}

func opsStrList(v interface{}) []string {
	list, ok := v.([]interface{})
	if !ok {
		return nil
	}
	out := make([]string, 0, len(list))
	for _, e := range list {
		out = append(out, opsStr(e))
	}
	return out
}

func opsToStrMap(v interface{}) map[string]string {
	m, ok := v.(map[string]interface{})
	if !ok {
		return map[string]string{}
	}
	out := make(map[string]string, len(m))
	for k, val := range m {
		out[k] = opsStr(val)
	}
	return out
}

func opsBool(v interface{}) bool {
	if b, ok := v.(bool); ok {
		return b
	}
	return false
}
`)
	// Only emit cron-dependent helper when cron is actually used,
	// to avoid an unused import error for opscron.
	if g.usedSDK["cron"] {
		b.WriteString(`
func opsToCronEntry(v interface{}) opscron.CronEntry {
	m, ok := v.(map[string]interface{})
	if !ok {
		return opscron.CronEntry{}
	}
	return opscron.CronEntry{
		Minute:     opsStr(m["minute"]),
		Hour:       opsStr(m["hour"]),
		DayOfMonth: opsStr(m["day_of_month"]),
		Month:      opsStr(m["month"]),
		DayOfWeek:  opsStr(m["day_of_week"]),
		Command:    opsStr(m["command"]),
	}
}
`)
	}
	b.WriteString(`
// opsFatal aborts the compiled script: runtime SDK errors must fail the
// deployment, not become string values that flow onward silently.
func opsFatal(err error) {
	fmt.Fprintf(os.Stderr, "runtime error: %v\n", err)
	os.Exit(1)
}

// opsEqual: numbers compare numerically (int64/float64/int), strings as
// strings, bools as bools. Values of different kinds are NOT equal (1 != "1")
// - cross-type string-coincidence matching hid real bugs. Matches the
// interpreter's isEqual exactly.
func opsEqual(l, r interface{}) bool {
	if l == nil || r == nil {
		return l == nil && r == nil
	}
	if lb, ok := l.(bool); ok {
		rb, rok := r.(bool)
		return rok && lb == rb
	}
	lf, lok := toNumber(l)
	rf, rok := toNumber(r)
	if lok && rok {
		return lf == rf
	}
	ls, lok := l.(string)
	rs, rsok := r.(string)
	if lok && rsok {
		return ls == rs
	}
	return false
}

func toNumber(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case int64:
		return float64(n), true
	case float64:
		return n, true
	case int:
		return float64(n), true
	default:
		return 0, false
	}
}

// typeName mirrors the interpreter's type() names.
func typeName(v interface{}) string {
	switch v.(type) {
	case nil:
		return "nil"
	case bool:
		return "bool"
	case int64, int:
		return "int"
	case float64:
		return "float"
	case string:
		return "string"
	case []interface{}:
		return "list"
	case map[string]interface{}:
		return "dict"
	default:
		return fmt.Sprintf("%T", v)
	}
}

`)
}

// writeUserFunc generates a Go function for a user-defined OpsLang function.
func (g *CodeGenerator) writeUserFunc(b *strings.Builder, fn userFunc) error {
	b.WriteString(fmt.Sprintf("func %s(", fn.name))
	for i, param := range fn.params {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(fmt.Sprintf("%s interface{}", sanitizeName(param.Name.Name)))
	}
	b.WriteString(") interface{} {\n")

	for _, stmt := range fn.body.Statements {
		if err := g.genStatementTo(b, stmt, 1); err != nil {
			return err
		}
	}

	b.WriteString("\treturn nil\n")
	b.WriteString("}\n\n")
	return nil
}

// genStatement generates Go code for a statement, appending to g.buf.
func (g *CodeGenerator) genStatement(stmt ast.Statement) error {
	return g.genStatementTo(&g.buf, stmt, g.indent)
}

func (g *CodeGenerator) genStatementTo(b *strings.Builder, stmt ast.Statement, indent int) error {
	prefix := strings.Repeat("\t", indent)

	switch s := stmt.(type) {
	case *ast.LetStatement:
		expr, err := g.genExpr(s.Value)
		if err != nil {
			return err
		}
		// Dynamic typing: declare every variable as interface{} so that a
		// later `x = <dynamic expr>` assignment always compiles. `x := int64(0)`
		// followed by `x = func() interface{}{...}()` was a compile error.
		b.WriteString(fmt.Sprintf("%svar %s interface{} = %s\n", prefix, sanitizeName(s.Name.Name), expr))

	case *ast.AssignStatement:
		target, err := g.genExpr(s.Target)
		if err != nil {
			return err
		}
		val, err := g.genExpr(s.Value)
		if err != nil {
			return err
		}
		b.WriteString(fmt.Sprintf("%s%s = %s\n", prefix, target, val))

	case *ast.ExpressionStatement:
		expr, err := g.genExpr(s.Expr)
		if err != nil {
			return err
		}
		b.WriteString(fmt.Sprintf("%s_ = %s\n", prefix, expr))

	case *ast.IfStatement:
		return g.genIf(b, s, indent)

	case *ast.ForStatement:
		return g.genFor(b, s, indent)

	case *ast.ForInStatement:
		return g.genForIn(b, s, indent)

	case *ast.BlockRescueStatement:
		return g.genBlockRescue(b, s, indent)

	case *ast.WhileStatement:
		return g.genWhile(b, s, indent)

	case *ast.ReturnStatement:
		if s.Value != nil {
			expr, err := g.genExpr(s.Value)
			if err != nil {
				return err
			}
			b.WriteString(fmt.Sprintf("%sreturn %s\n", prefix, expr))
		} else {
			b.WriteString(fmt.Sprintf("%sreturn nil\n", prefix))
		}

	case *ast.TaskStatement:
		// In AOT mode, execute task body directly
		for _, inner := range s.Body.Statements {
			if err := g.genStatementTo(b, inner, indent); err != nil {
				return err
			}
		}

	case *ast.ReportStatement:
		return g.genReport(b, s, indent)

	case *ast.AlertStatement:
		g.useOS = true
		msg, err := g.genExpr(s.Message)
		if err != nil {
			return err
		}
		b.WriteString(fmt.Sprintf("%ssetOutput(&_outputMu, _output, \"alert\", %s)\n", prefix, msg))
		b.WriteString(fmt.Sprintf("%sfmt.Fprintf(os.Stderr, \"ALERT: %%s\\n\", formatValue(%s))\n", prefix, msg))

	case *ast.ImportStatement:
		if strings.HasPrefix(s.Path, "go ") || strings.HasPrefix(s.Path, "go:") {
			return fmt.Errorf("import %q: third-party Go imports are not supported yet", s.Path)
		}
		// Standard SDK imports are declarative; nothing to compile.

	case *ast.PrivilegeStatement:
		// Metadata enforced by opsctl before deployment.

	case *ast.LogStatement:
		msg, err := g.genExpr(s.Message)
		if err != nil {
			return err
		}
		b.WriteString(fmt.Sprintf("%sfmt.Println(formatValue(%s))\n", prefix, msg))

	case *ast.MetricStatement:
		name, err := g.genExpr(s.Name)
		if err != nil {
			return err
		}
		value, err := g.genExpr(s.Value)
		if err != nil {
			return err
		}
		b.WriteString(fmt.Sprintf("%ssetOutput(&_outputMu, _output, \"metric:\"+opsStr(%s), %s)\n", prefix, name, value))

	case *ast.EnsureStatement:
		return g.genEnsure(b, s, indent)

	case *ast.FnStatement:
		// Already collected

	case *ast.BlockStatement:
		for _, inner := range s.Statements {
			if err := g.genStatementTo(b, inner, indent); err != nil {
				return err
			}
		}

	case *ast.ParallelStatement:
		return g.genParallel(b, s, indent)

	default:
		return fmt.Errorf("unsupported statement type: %T", stmt)
	}

	return nil
}

func (g *CodeGenerator) genIf(b *strings.Builder, s *ast.IfStatement, indent int) error {
	prefix := strings.Repeat("\t", indent)
	cond, err := g.genExpr(s.Condition)
	if err != nil {
		return err
	}
	b.WriteString(fmt.Sprintf("%sif isTruthy(%s) {\n", prefix, cond))
	for _, stmt := range s.Body.Statements {
		if err := g.genStatementTo(b, stmt, indent+1); err != nil {
			return err
		}
	}

	if s.ElseClause != nil {
		switch e := s.ElseClause.(type) {
		case *ast.BlockStatement:
			b.WriteString(fmt.Sprintf("%s} else {\n", prefix))
			for _, stmt := range e.Statements {
				if err := g.genStatementTo(b, stmt, indent+1); err != nil {
					return err
				}
			}
			b.WriteString(fmt.Sprintf("%s}\n", prefix))
		case *ast.IfStatement:
			b.WriteString(fmt.Sprintf("%s} else ", prefix))
			return g.genIf(b, e, indent)
		}
	} else {
		b.WriteString(fmt.Sprintf("%s}\n", prefix))
	}
	return nil
}

func (g *CodeGenerator) genFor(b *strings.Builder, s *ast.ForStatement, indent int) error {
	prefix := strings.Repeat("\t", indent)

	// C-style for loop. The loop variable must be interface{} so that the
	// post statement (`i = i + 1`, an interface{} expression) compiles.
	// Go's for-init only accepts short declarations, so a let-initializer
	// is emitted as `i := interface{}(<expr>)`.
	initStr := ""
	if s.Init != nil {
		if let, ok := s.Init.(*ast.LetStatement); ok {
			expr, err := g.genExpr(let.Value)
			if err != nil {
				return err
			}
			initStr = fmt.Sprintf("%s := interface{}(%s)", sanitizeName(let.Name.Name), expr)
		} else {
			var tmp strings.Builder
			if err := g.genStatementTo(&tmp, s.Init, 0); err != nil {
				return err
			}
			initStr = strings.TrimSpace(tmp.String())
		}
	}

	condStr := "true"
	if s.Condition != nil {
		cond, err := g.genExpr(s.Condition)
		if err != nil {
			return err
		}
		condStr = fmt.Sprintf("isTruthy(%s)", cond)
	}

	postStr := ""
	if s.Post != nil {
		var tmp strings.Builder
		if err := g.genStatementTo(&tmp, s.Post, 0); err != nil {
			return err
		}
		postStr = strings.TrimSpace(tmp.String())
	}

	b.WriteString(fmt.Sprintf("%sfor %s; %s; %s {\n", prefix, initStr, condStr, postStr))
	for _, stmt := range s.Body.Statements {
		if err := g.genStatementTo(b, stmt, indent+1); err != nil {
			return err
		}
	}
	b.WriteString(fmt.Sprintf("%s}\n", prefix))
	return nil
}

func (g *CodeGenerator) genForIn(b *strings.Builder, s *ast.ForInStatement, indent int) error {
	prefix := strings.Repeat("\t", indent)
	iterExpr, err := g.genExpr(s.Iterable)
	if err != nil {
		return err
	}
	varName := sanitizeName(s.Var.Name)

	// Emit a type-switch runtime dispatch over list/dict/string.
	b.WriteString(fmt.Sprintf("%sfor _, _item := range func() []interface{} {\n", prefix))
	b.WriteString(fmt.Sprintf("%s\t_v := %s\n", prefix, iterExpr))
	b.WriteString(fmt.Sprintf("%s\tswitch _c := _v.(type) {\n", prefix))
	b.WriteString(fmt.Sprintf("%s\tcase []interface{}:\n", prefix))
	b.WriteString(fmt.Sprintf("%s\t\treturn _c\n", prefix))
	b.WriteString(fmt.Sprintf("%s\tcase map[string]interface{}:\n", prefix))
	b.WriteString(fmt.Sprintf("%s\t\t_keys := make([]string, 0, len(_c))\n", prefix))
	b.WriteString(fmt.Sprintf("%s\t\tfor _k := range _c { _keys = append(_keys, _k) }\n", prefix))
	b.WriteString(fmt.Sprintf("%s\t\tsort.Strings(_keys)\n", prefix))
	b.WriteString(fmt.Sprintf("%s\t\t_out := make([]interface{}, len(_keys))\n", prefix))
	b.WriteString(fmt.Sprintf("%s\t\tfor _i, _k := range _keys { _out[_i] = _k }\n", prefix))
	b.WriteString(fmt.Sprintf("%s\t\treturn _out\n", prefix))
	b.WriteString(fmt.Sprintf("%s\tcase string:\n", prefix))
	b.WriteString(fmt.Sprintf("%s\t\t_out := make([]interface{}, 0, len(_c))\n", prefix))
	b.WriteString(fmt.Sprintf("%s\t\tfor _, _r := range _c { _out = append(_out, string(_r)) }\n", prefix))
	b.WriteString(fmt.Sprintf("%s\t\treturn _out\n", prefix))
	b.WriteString(fmt.Sprintf("%s\tdefault:\n", prefix))
	b.WriteString(fmt.Sprintf("%s\t\topsFatal(fmt.Errorf(\"for-in requires a list, dict, or string, got %%T\", _v))\n", prefix))
	b.WriteString(fmt.Sprintf("%s\t\treturn nil\n", prefix))
	b.WriteString(fmt.Sprintf("%s\t}\n", prefix))
	b.WriteString(fmt.Sprintf("%s}() {\n", prefix))
	b.WriteString(fmt.Sprintf("%s\t%s := _item\n", prefix, varName))
	b.WriteString(fmt.Sprintf("%s\t_ = %s\n", prefix, varName))
	for _, stmt := range s.Body.Statements {
		if err := g.genStatementTo(b, stmt, indent+1); err != nil {
			return err
		}
	}
	b.WriteString(fmt.Sprintf("%s}\n", prefix))
	return nil
}

func (g *CodeGenerator) genBlockRescue(b *strings.Builder, s *ast.BlockRescueStatement, indent int) error {
	prefix := strings.Repeat("\t", indent)

	if s.Rescue != nil {
		// Wrap Body in a func that recovers panics. When there is a rescue
		// clause, body errors are converted to Go panics internally and
		// caught here so the rescue body can inspect _error.
		b.WriteString(fmt.Sprintf("%sfunc() {\n", prefix))
		b.WriteString(fmt.Sprintf("%s\tdefer func() {\n", prefix))
		b.WriteString(fmt.Sprintf("%s\t\tif _r := recover(); _r != nil {\n", prefix))
		b.WriteString(fmt.Sprintf("%s\t\t\t_error := fmt.Sprintf(\"%%v\", _r)\n", prefix))
		b.WriteString(fmt.Sprintf("%s\t\t\t_ = _error\n", prefix))
		for _, stmt := range s.Rescue.Statements {
			if err := g.genStatementTo(b, stmt, indent+3); err != nil {
				return err
			}
		}
		b.WriteString(fmt.Sprintf("%s\t\t}\n", prefix))
		b.WriteString(fmt.Sprintf("%s\t}()\n", prefix))
	}

	// Block body.
	if s.Body != nil {
		for _, stmt := range s.Body.Statements {
			if err := g.genStatementTo(b, stmt, indent+1); err != nil {
				return err
			}
		}
	}

	if s.Rescue != nil {
		b.WriteString(fmt.Sprintf("%s}()\n", prefix))
	}

	// Always clause (runs unconditionally after body/rescue).
	if s.Always != nil {
		for _, stmt := range s.Always.Statements {
			if err := g.genStatementTo(b, stmt, indent); err != nil {
				return err
			}
		}
	}
	return nil
}

func (g *CodeGenerator) genWhile(b *strings.Builder, s *ast.WhileStatement, indent int) error {
	prefix := strings.Repeat("\t", indent)
	cond, err := g.genExpr(s.Condition)
	if err != nil {
		return err
	}
	b.WriteString(fmt.Sprintf("%sfor isTruthy(%s) {\n", prefix, cond))
	for _, stmt := range s.Body.Statements {
		if err := g.genStatementTo(b, stmt, indent+1); err != nil {
			return err
		}
	}
	b.WriteString(fmt.Sprintf("%s}\n", prefix))
	return nil
}

// genParallel compiles a parallel block: statements run concurrently, let
// declarations are captured per goroutine and merged back in source order
// after Wait (matching interpreter semantics: deterministic merge, no
// shared-map writes while goroutines run). Assignments inside parallel are
// rejected - they would race on shared variables.
func (g *CodeGenerator) genParallel(b *strings.Builder, s *ast.ParallelStatement, indent int) error {
	prefix := strings.Repeat("\t", indent)
	stmts := s.Body.Statements
	if len(stmts) == 0 {
		return nil
	}

	// Validate statement kinds first: only let / expression / report / log
	// statements can run isolated inside a goroutine.
	for _, stmt := range stmts {
		switch stmt.(type) {
		case *ast.LetStatement, *ast.ExpressionStatement, *ast.ReportStatement, *ast.LogStatement:
		default:
			return fmt.Errorf("parallel blocks in AOT mode support let, calls, report and log statements; %T would need shared-variable mutation", stmt)
		}
	}

	b.WriteString(fmt.Sprintf("%s{\n", prefix))
	b.WriteString(fmt.Sprintf("%s\tvar _pWg sync.WaitGroup\n", prefix))
	b.WriteString(fmt.Sprintf("%s\t_pWg.Add(%d)\n", prefix, len(stmts)))
	b.WriteString(fmt.Sprintf("%s\t_pRes := make([]map[string]interface{}, %d)\n", prefix, len(stmts)))

	mergeLines := []string{}
	for i, stmt := range stmts {
		switch st := stmt.(type) {
		case *ast.LetStatement:
			expr, err := g.genExpr(st.Value)
			if err != nil {
				return err
			}
			b.WriteString(fmt.Sprintf("%s\tgo func(idx int) {\n", prefix))
			b.WriteString(fmt.Sprintf("%s\t\tdefer _pWg.Done()\n", prefix))
			b.WriteString(fmt.Sprintf("%s\t\t_pRes[idx] = map[string]interface{}{%q: %s}\n", prefix, st.Name.Name, expr))
			b.WriteString(fmt.Sprintf("%s\t}(%d)\n", prefix, i))
			mergeLines = append(mergeLines,
				fmt.Sprintf("%s\t%s = _pRes[%d][%q]\n", prefix, sanitizeName(st.Name.Name), i, st.Name.Name))
		default:
			b.WriteString(fmt.Sprintf("%s\tgo func() {\n", prefix))
			b.WriteString(fmt.Sprintf("%s\t\tdefer _pWg.Done()\n", prefix))
			if err := g.genStatementTo(b, stmt, indent+2); err != nil {
				return err
			}
			b.WriteString(fmt.Sprintf("%s\t}()\n", prefix))
		}
	}

	b.WriteString(fmt.Sprintf("%s\t_pWg.Wait()\n", prefix))
	for _, line := range mergeLines {
		b.WriteString(line)
	}
	b.WriteString(fmt.Sprintf("%s}\n", prefix))
	return nil
}

// genEnsure implements the check -> apply -> verify (-> notify) contract
// with the same semantics as the interpreter.
func (g *CodeGenerator) genEnsure(b *strings.Builder, s *ast.EnsureStatement, indent int) error {
	prefix := strings.Repeat("\t", indent)

	cond, err := g.genExpr(s.Condition)
	if err != nil {
		return err
	}

	b.WriteString(fmt.Sprintf("%sif !isTruthy(%s) {\n", prefix, cond))
	bodyPrefix := strings.Repeat("\t", indent+1)
	for _, stmt := range s.Body.Statements {
		if err := g.genStatementTo(b, stmt, indent+1); err != nil {
			return err
		}
	}
	// VERIFY
	b.WriteString(fmt.Sprintf("%sif !isTruthy(%s) {\n", bodyPrefix, cond))
	b.WriteString(fmt.Sprintf("%s\topsFatal(fmt.Errorf(\"ensure: condition still false after applying actions\"))\n", bodyPrefix))
	b.WriteString(fmt.Sprintf("%s}\n", bodyPrefix))
	// NOTIFY (optional)
	if s.Notify != nil {
		notify, err := g.genExpr(s.Notify)
		if err != nil {
			return err
		}
		b.WriteString(fmt.Sprintf("%s_ = %s\n", bodyPrefix, notify))
	}
	b.WriteString(fmt.Sprintf("%s}\n", prefix))
	return nil
}

func (g *CodeGenerator) genReport(b *strings.Builder, s *ast.ReportStatement, indent int) error {
	prefix := strings.Repeat("\t", indent)
	for _, field := range s.Fields {
		val, err := g.genExpr(field.Value)
		if err != nil {
			return err
		}
		b.WriteString(fmt.Sprintf("%ssetOutput(&_outputMu, _output, %q, %s)\n", prefix, field.Key, val))
	}
	return nil
}

// genExpr generates a Go expression string from an AST expression.
func (g *CodeGenerator) genExpr(expr ast.Expression) (string, error) {
	switch e := expr.(type) {
	case *ast.IntegerLiteral:
		return fmt.Sprintf("int64(%d)", e.Value), nil

	case *ast.FloatLiteral:
		return fmt.Sprintf("float64(%g)", e.Value), nil

	case *ast.StringLiteral:
		return fmt.Sprintf("%q", e.Value), nil

	case *ast.BoolLiteral:
		if e.Value {
			return "true", nil
		}
		return "false", nil

	case *ast.NilLiteral:
		return "nil", nil

	case *ast.Identifier:
		return sanitizeName(e.Name), nil

	case *ast.BinaryExpression:
		return g.genBinary(e)

	case *ast.UnaryExpression:
		right, err := g.genExpr(e.Right)
		if err != nil {
			return "", err
		}
		if e.Op == "!" {
			return fmt.Sprintf("!isTruthy(%s)", right), nil
		}
		return fmt.Sprintf("(%s%s)", e.Op, right), nil

	case *ast.CallExpression:
		return g.genCall(e)

	case *ast.IndexExpression:
		left, err := g.genExpr(e.Left)
		if err != nil {
			return "", err
		}
		idx, err := g.genExpr(e.Index)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("func() interface{} { l := opsNormalize(%s); i := %s; if v, ok := l.([]interface{}); ok { idx := int(toFloat(i)); if idx >= 0 && idx < len(v) { return v[idx] }; return nil }; return opsToMapDeep(l)[opsStr(i)] }()", left, idx), nil

	case *ast.MemberExpression:
		obj, err := g.genExpr(e.Object)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("func() interface{} { return opsToMapDeep(%s)[%q] }()", obj, e.Member.Name), nil

	case *ast.ListLiteral:
		return g.genList(e)

	case *ast.DictLiteral:
		return g.genDict(e)

	case *ast.IfExpression:
		return g.genIfExpr(e)

	default:
		return "", fmt.Errorf("unsupported expression type: %T", expr)
	}
}

func (g *CodeGenerator) genBinary(e *ast.BinaryExpression) (string, error) {
	left, err := g.genExpr(e.Left)
	if err != nil {
		return "", err
	}
	right, err := g.genExpr(e.Right)
	if err != nil {
		return "", err
	}

	switch e.Op {
	case "+":
		return fmt.Sprintf("func() interface{} { var l, r interface{} = %s, %s; if ls, ok := l.(string); ok { return ls + formatValue(r) }; if _, ok := r.(string); ok { return formatValue(l) + formatValue(r) }; return toFloat(l) + toFloat(r) }()", left, right), nil
	case "-":
		return fmt.Sprintf("func() interface{} { l, r := toFloat(%s), toFloat(%s); if l == float64(int64(l)) && r == float64(int64(r)) { return int64(l) - int64(r) }; return l - r }()", left, right), nil
	case "*":
		return fmt.Sprintf("func() interface{} { l, r := toFloat(%s), toFloat(%s); if l == float64(int64(l)) && r == float64(int64(r)) { return int64(l) * int64(r) }; return l * r }()", left, right), nil
	case "/":
		return fmt.Sprintf("func() interface{} { l, r := toFloat(%s), toFloat(%s); if r == 0 { return nil }; if l == float64(int64(l)) && r == float64(int64(r)) { return int64(l) / int64(r) }; return l / r }()", left, right), nil
	case "%":
		return fmt.Sprintf("func() interface{} { l, r := toInt(%s), toInt(%s); if r == 0 { return nil }; return l %% r }()", left, right), nil
	case "==":
		return fmt.Sprintf("func() interface{} { return opsEqual(%s, %s) }()", left, right), nil
	case "!=":
		return fmt.Sprintf("func() interface{} { return !opsEqual(%s, %s) }()", left, right), nil
	case "<":
		return fmt.Sprintf("func() interface{} { return toFloat(%s) < toFloat(%s) }()", left, right), nil
	case ">":
		return fmt.Sprintf("func() interface{} { return toFloat(%s) > toFloat(%s) }()", left, right), nil
	case "<=":
		return fmt.Sprintf("func() interface{} { return toFloat(%s) <= toFloat(%s) }()", left, right), nil
	case ">=":
		return fmt.Sprintf("func() interface{} { return toFloat(%s) >= toFloat(%s) }()", left, right), nil
	case "&&":
		return fmt.Sprintf("func() interface{} { if !isTruthy(%s) { return false }; return isTruthy(%s) }()", left, right), nil
	case "||":
		return fmt.Sprintf("func() interface{} { if isTruthy(%s) { return true }; return isTruthy(%s) }()", left, right), nil
	default:
		return "", fmt.Errorf("unsupported binary operator: %s", e.Op)
	}
}

func (g *CodeGenerator) genCall(e *ast.CallExpression) (string, error) {
	fnName := g.resolveFuncName(e.Function)

	// Check builtins
	switch fnName {
	case "print":
		args, err := g.genArgs(e.Args)
		if err != nil {
			return "", err
		}
		if len(args) == 0 {
			return "func() interface{} { fmt.Println(); return nil }()", nil
		}
		parts := make([]string, len(args))
		for i, a := range args {
			parts[i] = fmt.Sprintf("formatValue(%s)", a)
		}
		return fmt.Sprintf("func() interface{} { fmt.Println(%s); return nil }()", strings.Join(parts, ", \" \", ")), nil

	case "len":
		if len(e.Args) != 1 {
			return "", fmt.Errorf("len() takes exactly 1 argument")
		}
		arg, err := g.genExpr(e.Args[0])
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("opsLen(%s)", arg), nil

	case "str":
		if len(e.Args) != 1 {
			return "", fmt.Errorf("str() takes exactly 1 argument")
		}
		arg, err := g.genExpr(e.Args[0])
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("formatValue(%s)", arg), nil

	case "int":
		if len(e.Args) != 1 {
			return "", fmt.Errorf("int() takes exactly 1 argument")
		}
		arg, err := g.genExpr(e.Args[0])
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("toInt(%s)", arg), nil

	case "float":
		if len(e.Args) != 1 {
			return "", fmt.Errorf("float() takes exactly 1 argument")
		}
		arg, err := g.genExpr(e.Args[0])
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("toFloat(%s)", arg), nil

	case "type":
		if len(e.Args) != 1 {
			return "", fmt.Errorf("type() takes exactly 1 argument")
		}
		arg, err := g.genExpr(e.Args[0])
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("typeName(%s)", arg), nil
	}

	// Check user-defined functions
	for _, fn := range g.userFuncs {
		if fn.name == fnName {
			args, err := g.genArgs(e.Args)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("%s(%s)", fn.name, strings.Join(args, ", ")), nil
		}
	}

	// Check SDK mapping
	if sdk, ok := sdkMapping[fnName]; ok {
		g.usedSDK[sdk.pkg] = true
		g.useOS = true // opsFatal writes to stderr and exits
		alias := pkgImportAlias[sdk.pkg]
		args, err := g.genArgs(e.Args)
		if err != nil {
			return "", err
		}

		// process.exec is variadic in the DSL: (command, arg1, arg2, ...).
		if fnName == "process.exec" && len(args) >= 1 {
			listArgs := "[]interface{}{}"
			if len(args) > 1 {
				listArgs = fmt.Sprintf("[]interface{}{%s}", strings.Join(args[1:], ", "))
			}
			args = []string{args[0], listArgs}
		}

		// Reject argument-count mismatches at generation time when the
		// signature is fixed (process.exec handled above).
		maxArgs := len(sdk.params)
		if fnName != "process.exec" && len(args) > maxArgs {
			return "", fmt.Errorf("%s() takes at most %d argument(s), got %d", fnName, maxArgs, len(e.Args))
		}

		call := sdk.generateSDKCall(alias, args)
		if sdk.noErr {
			return fmt.Sprintf("func() interface{} { return %s }()", call), nil
		}
		// A runtime SDK error aborts the binary with a non-zero exit code.
		// Turning errors into strings used to let failed deploys "succeed".
		return fmt.Sprintf("func() interface{} { v, err := %s; if err != nil { opsFatal(err) }; return v }()", call), nil
	}

	return "", fmt.Errorf("unknown function %q (not a builtin, user function, or SDK call)", fnName)
}

func (g *CodeGenerator) genList(e *ast.ListLiteral) (string, error) {
	if len(e.Elements) == 0 {
		return "[]interface{}{}", nil
	}
	elems, err := g.genArgs(e.Elements)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("[]interface{}{%s}", strings.Join(elems, ", ")), nil
}

func (g *CodeGenerator) genDict(e *ast.DictLiteral) (string, error) {
	if len(e.Keys) == 0 {
		return "map[string]interface{}{}", nil
	}
	var pairs []string
	for i := range e.Keys {
		key, err := g.genExpr(e.Keys[i])
		if err != nil {
			return "", err
		}
		val, err := g.genExpr(e.Values[i])
		if err != nil {
			return "", err
		}
		pairs = append(pairs, fmt.Sprintf("fmt.Sprintf(\"%%v\", %s): %s", key, val))
	}
	return fmt.Sprintf("map[string]interface{}{%s}", strings.Join(pairs, ", ")), nil
}

func (g *CodeGenerator) genIfExpr(e *ast.IfExpression) (string, error) {
	thenExpr, err := g.genExpr(e.Then)
	if err != nil {
		return "", err
	}
	elseExpr, err := g.genExpr(e.Else)
	if err != nil {
		return "", err
	}
	if e.Condition != nil {
		cond, err := g.genExpr(e.Condition)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("func() interface{} { if isTruthy(%s) { return %s }; return %s }()", cond, thenExpr, elseExpr), nil
	}
	return thenExpr, nil
}

func (g *CodeGenerator) genArgs(args []ast.Expression) ([]string, error) {
	result := make([]string, len(args))
	for i, arg := range args {
		s, err := g.genExpr(arg)
		if err != nil {
			return nil, err
		}
		result[i] = s
	}
	return result, nil
}

// resolveFuncName builds a dotted name from a call's function expression.
func (g *CodeGenerator) resolveFuncName(expr ast.Expression) string {
	return resolveCallName(expr)
}

// sanitizeName escapes Go reserved words and invalid identifier characters
// that might appear in OpsLang variable names.
func sanitizeName(name string) string {
	goReserved := map[string]bool{
		"break": true, "case": true, "chan": true, "const": true, "continue": true,
		"default": true, "defer": true, "else": true, "fallthrough": true, "for": true,
		"func": true, "go": true, "goto": true, "if": true, "import": true,
		"interface": true, "map": true, "package": true, "range": true, "return": true,
		"select": true, "struct": true, "switch": true, "type": true, "var": true,
	}
	if goReserved[name] {
		return "_" + name
	}
	var result strings.Builder
	for i, ch := range name {
		if i == 0 && !unicode.IsLetter(ch) && ch != '_' {
			result.WriteRune('_')
		} else if !unicode.IsLetter(ch) && !unicode.IsDigit(ch) && ch != '_' {
			result.WriteRune('_')
		} else {
			result.WriteRune(ch)
		}
	}
	return result.String()
}
