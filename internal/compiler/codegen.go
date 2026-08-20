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

	// apt
	"apt.install":       {pkg: "apt", goName: "Install", args: true, params: []string{"s", "s", "b"}},
	"apt.remove":        {pkg: "apt", goName: "Remove", args: true, params: []string{"s", "b"}},
	"apt.upgrade":       {pkg: "apt", goName: "Upgrade", args: true, params: []string{"s"}},
	"apt.update_cache":  {pkg: "apt", goName: "UpdateCache"},
	"apt.full_upgrade":  {pkg: "apt", goName: "FullUpgrade"},
	"apt.dist_upgrade":  {pkg: "apt", goName: "DistUpgrade"},
	"apt.autoremove":    {pkg: "apt", goName: "Autoremove"},
	"apt.clean":         {pkg: "apt", goName: "Clean"},
	"apt.info":          {pkg: "apt", goName: "Info", args: true, params: []string{"s"}},
	"apt.list":          {pkg: "apt", goName: "List"},
	"apt.policy":        {pkg: "apt", goName: "Policy", args: true, params: []string{"s"}},
	"apt.mark_auto":     {pkg: "apt", goName: "MarkAuto", args: true, params: []string{"s"}},
	"apt.mark_manual":   {pkg: "apt", goName: "MarkManual", args: true, params: []string{"s"}},

	// dnf
	"dnf.install":       {pkg: "dnf", goName: "Install", args: true, params: []string{"s", "s"}},
	"dnf.remove":        {pkg: "dnf", goName: "Remove", args: true, params: []string{"s"}},
	"dnf.update":        {pkg: "dnf", goName: "Update", args: true, params: []string{"s"}},
	"dnf.info":          {pkg: "dnf", goName: "Info", args: true, params: []string{"s"}},
	"dnf.list":          {pkg: "dnf", goName: "List"},
	"dnf.search":        {pkg: "dnf", goName: "Search", args: true, params: []string{"s"}},
	"dnf.clean":         {pkg: "dnf", goName: "Clean"},
	"dnf.repolist":      {pkg: "dnf", goName: "RepoList"},
	"dnf.grouplist":     {pkg: "dnf", goName: "GroupList"},
	"dnf.groupinstall":  {pkg: "dnf", goName: "GroupInstall", args: true, params: []string{"s"}},
	"dnf.groupremove":   {pkg: "dnf", goName: "GroupRemove", args: true, params: []string{"s"}},
	"dnf.history":       {pkg: "dnf", goName: "History", args: true, params: []string{"i"}},
	"dnf.check_update":  {pkg: "dnf", goName: "CheckUpdate"},
	"dnf.modulelist":    {pkg: "dnf", goName: "ModuleList"},
	"dnf.module_enable": {pkg: "dnf", goName: "ModuleEnable", args: true, params: []string{"s"}},

	// apk
	"apk.install":           {pkg: "apk", goName: "Install", args: true, params: []string{"s", "s"}},
	"apk.remove":            {pkg: "apk", goName: "Remove", args: true, params: []string{"s", "b"}},
	"apk.update":            {pkg: "apk", goName: "Update"},
	"apk.upgrade":           {pkg: "apk", goName: "Upgrade", args: true, params: []string{"s"}},
	"apk.info":              {pkg: "apk", goName: "Info", args: true, params: []string{"s"}},
	"apk.list":              {pkg: "apk", goName: "List"},
	"apk.search":            {pkg: "apk", goName: "Search", args: true, params: []string{"s"}},
	"apk.cache":             {pkg: "apk", goName: "Cache"},
	"apk.upgrade_available": {pkg: "apk", goName: "UpgradeAvailable"},
	"apk.repository":        {pkg: "apk", goName: "Repository"},

	// sysvinit
	"sysvinit.status":  {pkg: "sysvinit", goName: "Status", args: true, params: []string{"s"}},
	"sysvinit.start":   {pkg: "sysvinit", goName: "Start", args: true, params: []string{"s"}},
	"sysvinit.stop":    {pkg: "sysvinit", goName: "Stop", args: true, params: []string{"s"}},
	"sysvinit.restart": {pkg: "sysvinit", goName: "Restart", args: true, params: []string{"s"}},
	"sysvinit.reload":  {pkg: "sysvinit", goName: "Reload", args: true, params: []string{"s"}},
	"sysvinit.enable":  {pkg: "sysvinit", goName: "Enable", args: true, params: []string{"s", "s"}},
	"sysvinit.disable": {pkg: "sysvinit", goName: "Disable", args: true, params: []string{"s"}},
	"sysvinit.list":    {pkg: "sysvinit", goName: "List"},

	// dpkg_selections
	"dpkg_selections.set":   {pkg: "dpkg_selections", goName: "SetSelection", args: true, params: []string{"s", "s"}},
	"dpkg_selections.get":   {pkg: "dpkg_selections", goName: "GetSelection", args: true, params: []string{"s"}},
	"dpkg_selections.list":  {pkg: "dpkg_selections", goName: "ListSelections"},
	"dpkg_selections.hold":  {pkg: "dpkg_selections", goName: "Hold", args: true, params: []string{"s"}},
	"dpkg_selections.unhold": {pkg: "dpkg_selections", goName: "Unhold", args: true, params: []string{"s"}},

	// homebrew
	"homebrew.install":   {pkg: "homebrew", goName: "Install", args: true, params: []string{"s", "b"}},
	"homebrew.remove":    {pkg: "homebrew", goName: "Remove", args: true, params: []string{"s", "b"}},
	"homebrew.upgrade":   {pkg: "homebrew", goName: "Upgrade", args: true, params: []string{"s"}},
	"homebrew.update":    {pkg: "homebrew", goName: "Update"},
	"homebrew.info":      {pkg: "homebrew", goName: "Info", args: true, params: []string{"s"}},
	"homebrew.list":      {pkg: "homebrew", goName: "List"},
	"homebrew.list_casks": {pkg: "homebrew", goName: "ListCasks"},
	"homebrew.outdated":  {pkg: "homebrew", goName: "Outdated"},
	"homebrew.clean":     {pkg: "homebrew", goName: "Clean"},
	"homebrew.tap":       {pkg: "homebrew", goName: "Tap", args: true, params: []string{"s"}},
	"homebrew.untap":     {pkg: "homebrew", goName: "Untap", args: true, params: []string{"s"}},
	"homebrew.list_taps": {pkg: "homebrew", goName: "ListTaps"},
	"homebrew.doctor":    {pkg: "homebrew", goName: "Doctor"},

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

	// lvg
	"lvg.create":    {pkg: "lvg", goName: "Create", args: true, params: []string{"s", "l"}},
	"lvg.remove":    {pkg: "lvg", goName: "Remove", args: true, params: []string{"s"}},
	"lvg.extend":    {pkg: "lvg", goName: "Extend", args: true, params: []string{"s", "l"}},
	"lvg.reduce":    {pkg: "lvg", goName: "Reduce", args: true, params: []string{"s", "l"}},
	"lvg.activate":  {pkg: "lvg", goName: "Activate", args: true, params: []string{"s"}},
	"lvg.deactivate": {pkg: "lvg", goName: "Deactivate", args: true, params: []string{"s"}},
	"lvg.list":      {pkg: "lvg", goName: "List"},
	"lvg.get":       {pkg: "lvg", goName: "Get", args: true, params: []string{"s"}},

	// snap
	"snap.install":  {pkg: "snap", goName: "Install", args: true, params: []string{"s", "s", "b"}},
	"snap.remove":   {pkg: "snap", goName: "Remove", args: true, params: []string{"s"}},
	"snap.refresh":  {pkg: "snap", goName: "Refresh", args: true, params: []string{"s", "s"}},
	"snap.list":     {pkg: "snap", goName: "List"},
	"snap.get":      {pkg: "snap", goName: "Get", args: true, params: []string{"s"}},
	"snap.enable":   {pkg: "snap", goName: "Enable", args: true, params: []string{"s"}},
	"snap.disable":  {pkg: "snap", goName: "Disable", args: true, params: []string{"s"}},
	"snap.switch":   {pkg: "snap", goName: "Switch", args: true, params: []string{"s", "s"}},
	"snap.changes":  {pkg: "snap", goName: "Changes"},

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

	// ufw
	"ufw.status":  {pkg: "ufw", goName: "Status"},
	"ufw.list":    {pkg: "ufw", goName: "List"},
	"ufw.enable":  {pkg: "ufw", goName: "Enable"},
	"ufw.disable": {pkg: "ufw", goName: "Disable"},
	"ufw.allow":   {pkg: "ufw", goName: "Allow", args: true, params: []string{"s", "s"}},
	"ufw.deny":    {pkg: "ufw", goName: "Deny", args: true, params: []string{"s", "s"}},
	"ufw.delete":  {pkg: "ufw", goName: "Delete", args: true, params: []string{"i"}},
	"ufw.reset":   {pkg: "ufw", goName: "Reset"},
	"ufw.reload":  {pkg: "ufw", goName: "Reload"},

	// ini_file
	"ini_file.sections":        {pkg: "ini_file", goName: "Sections", args: true, params: []string{"s"}},
	"ini_file.get":             {pkg: "ini_file", goName: "Get", args: true, params: []string{"s", "s", "s"}},
	"ini_file.set":             {pkg: "ini_file", goName: "Set", args: true, params: []string{"s", "s", "s", "s"}},
	"ini_file.remove":          {pkg: "ini_file", goName: "Remove", args: true, params: []string{"s", "s", "s"}},
	"ini_file.remove_section":  {pkg: "ini_file", goName: "RemoveSection", args: true, params: []string{"s", "s"}},

	// mount
	"mount.list":        {pkg: "mount", goName: "List"},
	"mount.mount":       {pkg: "mount", goName: "Mount", args: true, params: []string{"s", "s", "s", "s"}},
	"mount.umount":      {pkg: "mount", goName: "Unmount", args: true, params: []string{"s"}},
	"mount.fstab":       {pkg: "mount", goName: "Fstab"},
	"mount.add_fstab":   {pkg: "mount", goName: "AddFstab", args: true, params: []string{"s", "s", "s", "s"}},
	"mount.remove_fstab": {pkg: "mount", goName: "RemoveFstab", args: true, params: []string{"s"}},

	// hostname
	"hostname.get":      {pkg: "hostname", goName: "Get"},
	"hostname.set":      {pkg: "hostname", goName: "Set", args: true, params: []string{"s"}},
	"hostname.set_fqdn": {pkg: "hostname", goName: "SetFQDN", args: true, params: []string{"s"}},

	// timezone
	"timezone.get":  {pkg: "timezone", goName: "Get"},
	"timezone.set":  {pkg: "timezone", goName: "Set", args: true, params: []string{"s"}},
	"timezone.list": {pkg: "timezone", goName: "List"},

	// iptables
	"iptables.list":        {pkg: "iptables", goName: "List", args: true, params: []string{"s"}},
	"iptables.flush":       {pkg: "iptables", goName: "Flush", args: true, params: []string{"s"}},
	"iptables.add_rule":    {pkg: "iptables", goName: "AddRule", args: true, params: []string{"s", "s"}},
	"iptables.delete_rule": {pkg: "iptables", goName: "DeleteRule", args: true, params: []string{"s", "i"}},
	"iptables.save":        {pkg: "iptables", goName: "Save"},
	"iptables.list_chains": {pkg: "iptables", goName: "ListChains"},

	// npm
	"npm.list":      {pkg: "npm", goName: "List", args: true, params: []string{"b"}},
	"npm.install":   {pkg: "npm", goName: "Install", args: true, params: []string{"s", "b"}},
	"npm.uninstall": {pkg: "npm", goName: "Uninstall", args: true, params: []string{"s", "b"}},
	"npm.outdated":  {pkg: "npm", goName: "Outdated", args: true, params: []string{"b"}},

	// mysql
	"mysql.databases":     {pkg: "mysql", goName: "Databases"},
	"mysql.create_database": {pkg: "mysql", goName: "CreateDatabase", args: true, params: []string{"s"}},
	"mysql.drop_database": {pkg: "mysql", goName: "DropDatabase", args: true, params: []string{"s"}},
	"mysql.users":         {pkg: "mysql", goName: "Users"},
	"mysql.create_user":   {pkg: "mysql", goName: "CreateUser", args: true, params: []string{"s", "s", "s"}},
	"mysql.drop_user":     {pkg: "mysql", goName: "DropUser", args: true, params: []string{"s", "s"}},
	"mysql.grant":         {pkg: "mysql", goName: "Grant", args: true, params: []string{"s", "s", "s", "s"}},

	// nginx
	"nginx.config_test":  {pkg: "nginx", goName: "ConfigTest"},
	"nginx.reload":       {pkg: "nginx", goName: "Reload"},
	"nginx.sites_list":   {pkg: "nginx", goName: "SitesList"},
	"nginx.site_enable":  {pkg: "nginx", goName: "SiteEnable", args: true, params: []string{"s"}},
	"nginx.site_disable": {pkg: "nginx", goName: "SiteDisable", args: true, params: []string{"s"}},

	// modprobe
	"modprobe.list":      {pkg: "modprobe", goName: "List"},
	"modprobe.load":      {pkg: "modprobe", goName: "Load", args: true, params: []string{"s"}},
	"modprobe.unload":    {pkg: "modprobe", goName: "Unload", args: true, params: []string{"s"}},
	"modprobe.is_loaded": {pkg: "modprobe", goName: "IsLoaded", args: true, params: []string{"s"}},

	// alternatives
	"alternatives.list":     {pkg: "alternatives", goName: "List", args: true, params: []string{"s"}},
	"alternatives.display":  {pkg: "alternatives", goName: "Display", args: true, params: []string{"s"}},
	"alternatives.set":      {pkg: "alternatives", goName: "Set", args: true, params: []string{"s", "s"}},
	"alternatives.install":  {pkg: "alternatives", goName: "Install", args: true, params: []string{"s", "s", "s", "i"}},
	"alternatives.remove":   {pkg: "alternatives", goName: "Remove", args: true, params: []string{"s", "s"}},

	// blockdev
	"blockdev.list":          {pkg: "blockdev", goName: "List"},
	"blockdev.info":          {pkg: "blockdev", goName: "Info", args: true, params: []string{"s"}},
	"blockdev.flush_buffers": {pkg: "blockdev", goName: "FlushBuffers", args: true, params: []string{"s"}},
	"blockdev.set_readahead": {pkg: "blockdev", goName: "SetReadahead", args: true, params: []string{"s", "i"}},

	// at
	"at.list":     {pkg: "at", goName: "List"},
	"at.schedule": {pkg: "at", goName: "Schedule", args: true, params: []string{"s", "s"}},
	"at.remove":   {pkg: "at", goName: "Remove", args: true, params: []string{"s"}},

	// ── postgresql ─────────────────────────────────────────────────────
	"postgresql.databases":      {pkg: "postgresql", goName: "Databases"},
	"postgresql.create_database": {pkg: "postgresql", goName: "CreateDatabase", args: true, params: []string{"s"}},
	"postgresql.drop_database":  {pkg: "postgresql", goName: "DropDatabase", args: true, params: []string{"s"}},
	"postgresql.users":          {pkg: "postgresql", goName: "Users"},
	"postgresql.create_user":    {pkg: "postgresql", goName: "CreateUser", args: true, params: []string{"s", "s"}},
	"postgresql.drop_user":      {pkg: "postgresql", goName: "DropUser", args: true, params: []string{"s"}},
	"postgresql.grant":          {pkg: "postgresql", goName: "Grant", args: true, params: []string{"s", "s", "s"}},

	// ── apache2 ────────────────────────────────────────────────────────
	"apache2.config_test":  {pkg: "apache2", goName: "ConfigTest"},
	"apache2.reload":       {pkg: "apache2", goName: "Reload"},
	"apache2.sites_list":   {pkg: "apache2", goName: "SitesList"},
	"apache2.site_enable":  {pkg: "apache2", goName: "SiteEnable", args: true, params: []string{"s"}},
	"apache2.site_disable": {pkg: "apache2", goName: "SiteDisable", args: true, params: []string{"s"}},
	"apache2.modules_list": {pkg: "apache2", goName: "ModulesList"},
	"apache2.module_enable": {pkg: "apache2", goName: "ModuleEnable", args: true, params: []string{"s"}},
	"apache2.module_disable": {pkg: "apache2", goName: "ModuleDisable", args: true, params: []string{"s"}},

	// ── filesystem ─────────────────────────────────────────────────────
	"filesystem.mkfs":        {pkg: "filesystem", goName: "Mkfs", args: true, params: []string{"s", "s", "s"}},
	"filesystem.resize_ext4": {pkg: "filesystem", goName: "ResizeExt4", args: true, params: []string{"s"}},
	"filesystem.resize_xfs":  {pkg: "filesystem", goName: "ResizeXFS", args: true, params: []string{"s"}},
	"filesystem.check":       {pkg: "filesystem", goName: "Check", args: true, params: []string{"s"}},

	// ── parted ─────────────────────────────────────────────────────────
	"parted.list":   {pkg: "parted", goName: "List", args: true, params: []string{"s"}},
	"parted.mklabel": {pkg: "parted", goName: "MkLabel", args: true, params: []string{"s", "s"}},
	"parted.mkpart":  {pkg: "parted", goName: "MkPart", args: true, params: []string{"s", "s", "s", "s", "s"}},
	"parted.rm":      {pkg: "parted", goName: "Rm", args: true, params: []string{"s", "i"}},

	// ── acl ────────────────────────────────────────────────────────────
	"acl.get":       {pkg: "acl", goName: "Get", args: true, params: []string{"s"}},
	"acl.set":       {pkg: "acl", goName: "Set", args: true, params: []string{"s", "s", "b"}},
	"acl.remove":    {pkg: "acl", goName: "Remove", args: true, params: []string{"s", "s", "b"}},
	"acl.remove_all": {pkg: "acl", goName: "RemoveAll", args: true, params: []string{"s", "b"}},

	// ── wait_for ───────────────────────────────────────────────────────
	"wait_for.port": {pkg: "wait_for", goName: "Port", args: true, params: []string{"s", "i", "i"}},
	"wait_for.file": {pkg: "wait_for", goName: "File", args: true, params: []string{"s", "i"}},
	"wait_for.url":  {pkg: "wait_for", goName: "URL", args: true, params: []string{"s", "i"}},

	// ── lvol ───────────────────────────────────────────────────────────
	"lvol.list":    {pkg: "lvol", goName: "List"},
	"lvol.vg_list": {pkg: "lvol", goName: "VGList"},
	"lvol.create":  {pkg: "lvol", goName: "Create", args: true, params: []string{"s", "s", "s"}},
	"lvol.remove":  {pkg: "lvol", goName: "Remove", args: true, params: []string{"s", "s"}},
	"lvol.resize":  {pkg: "lvol", goName: "Resize", args: true, params: []string{"s", "s", "s"}},

	// ── synchronize ────────────────────────────────────────────────────
	"synchronize.sync": {pkg: "synchronize", goName: "Sync", args: true, params: []string{"s", "s", "b", "b"}},

	// ── fetch ──────────────────────────────────────────────────────────
	"fetch.file": {pkg: "fetch", goName: "File", args: true, params: []string{"s", "s"}},
	"fetch.url":  {pkg: "fetch", goName: "URL", args: true, params: []string{"s", "s"}},

	// ── seboolean ──────────────────────────────────────────────────────
	"seboolean.list": {pkg: "seboolean", goName: "List"},
	"seboolean.get":  {pkg: "seboolean", goName: "Get", args: true, params: []string{"s"}},
	"seboolean.set":  {pkg: "seboolean", goName: "Set", args: true, params: []string{"s", "b", "b"}},

	// ── uri ──────────────────────────────────────────────────────────────
	"uri.do":       {pkg: "uri", goName: "Do", args: true, params: []string{"s", "s", "ms", "s", "i"}},
	"uri.get":      {pkg: "uri", goName: "Get", args: true, params: []string{"s"}},
	"uri.post":     {pkg: "uri", goName: "Post", args: true, params: []string{"s", "a"}},
	"uri.put":      {pkg: "uri", goName: "Put", args: true, params: []string{"s", "a"}},
	"uri.delete":   {pkg: "uri", goName: "Delete", args: true, params: []string{"s"}},
	"uri.download": {pkg: "uri", goName: "Download", args: true, params: []string{"s", "s"}},

	// ── lineinfile ───────────────────────────────────────────────────────
	"lineinfile.present": {pkg: "lineinfile", goName: "Ensure", args: true, params: []string{"s", "s", "s", "b"}},
	"lineinfile.absent": {pkg: "lineinfile", goName: "Absent", args: true, params: []string{"s", "s"}},

	// ── replace ──────────────────────────────────────────────────────────
	"replace.replace": {pkg: "replace", goName: "Replace", args: true, params: []string{"s", "s", "s", "b"}},

	// ── xml ──────────────────────────────────────────────────────────────
	"xml.get_element": {pkg: "xml", goName: "GetElement", args: true, params: []string{"s", "s"}},
	"xml.set_element": {pkg: "xml", goName: "SetElement", args: true, params: []string{"s", "s", "s"}},

	// ── systemd ─────────────────────────────────────────────────────────────
	"systemd.is_active":     {pkg: "systemd", goName: "IsActive", args: true, params: []string{"s"}},
	"systemd.is_enabled":    {pkg: "systemd", goName: "IsEnabled", args: true, params: []string{"s"}},
	"systemd.enable":        {pkg: "systemd", goName: "Enable", args: true, params: []string{"s"}},
	"systemd.disable":       {pkg: "systemd", goName: "Disable", args: true, params: []string{"s"}},
	"systemd.start":         {pkg: "systemd", goName: "Start", args: true, params: []string{"s"}},
	"systemd.stop":          {pkg: "systemd", goName: "Stop", args: true, params: []string{"s"}},
	"systemd.restart":       {pkg: "systemd", goName: "Restart", args: true, params: []string{"s"}},
	"systemd.reload":        {pkg: "systemd", goName: "Reload", args: true, params: []string{"s"}},
	"systemd.daemon_reload": {pkg: "systemd", goName: "DaemonReload", args: true, params: []string{}},
	"systemd.mask":          {pkg: "systemd", goName: "Mask", args: true, params: []string{"s"}},
	"systemd.unmask":        {pkg: "systemd", goName: "Unmask", args: true, params: []string{"s"}},
	"systemd.show":          {pkg: "systemd", goName: "Show", args: true, params: []string{"s"}},
	"systemd.list":          {pkg: "systemd", goName: "List", args: true, params: []string{"s"}},

	// ── patch ───────────────────────────────────────────────────────────────
	"patch.apply":   {pkg: "patch", goName: "Apply", args: true, params: []string{"s", "b"}},
	"patch.dry_run": {pkg: "patch", goName: "DryRun", args: true, params: []string{"s"}},

	// ── xattr ───────────────────────────────────────────────────────────────
	"xattr.get":    {pkg: "xattr", goName: "Get", args: true, params: []string{"s", "s"}},
	"xattr.set":    {pkg: "xattr", goName: "Set", args: true, params: []string{"s", "s", "s"}},
	"xattr.remove": {pkg: "xattr", goName: "Remove", args: true, params: []string{"s", "s"}},
	"xattr.list":   {pkg: "xattr", goName: "List", args: true, params: []string{"s"}},

	// ── firewalld_zone ──────────────────────────────────────────────────────
	"firewalld_zone.get_default":      {pkg: "firewalld_zone", goName: "GetDefaultZone", args: true, params: []string{}},
	"firewalld_zone.set_default":      {pkg: "firewalld_zone", goName: "SetDefaultZone", args: true, params: []string{"s"}},
	"firewalld_zone.add_zone":         {pkg: "firewalld_zone", goName: "AddZone", args: true, params: []string{"s"}},
	"firewalld_zone.remove_zone":      {pkg: "firewalld_zone", goName: "RemoveZone", args: true, params: []string{"s"}},
	"firewalld_zone.add_service":      {pkg: "firewalld_zone", goName: "AddService", args: true, params: []string{"s", "s"}},
	"firewalld_zone.remove_service":   {pkg: "firewalld_zone", goName: "RemoveService", args: true, params: []string{"s", "s"}},
	"firewalld_zone.add_port":         {pkg: "firewalld_zone", goName: "AddPort", args: true, params: []string{"s", "s"}},
	"firewalld_zone.remove_port":      {pkg: "firewalld_zone", goName: "RemovePort", args: true, params: []string{"s", "s"}},
	"firewalld_zone.add_rich_rule":   {pkg: "firewalld_zone", goName: "AddRichRule", args: true, params: []string{"s", "s"}},
	"firewalld_zone.remove_rich_rule": {pkg: "firewalld_zone", goName: "RemoveRichRule", args: true, params: []string{"s", "s"}},
	"firewalld_zone.info":             {pkg: "firewalld_zone", goName: "Info", args: true, params: []string{"s"}},
	"firewalld_zone.list_zones":       {pkg: "firewalld_zone", goName: "ListZones", args: true, params: []string{}},

	// ── get_url ─────────────────────────────────────────────────────────────
	"get_url.download": {pkg: "get_url", goName: "Download", args: true, params: []string{"s", "s", "s", "b"}},

	// ── sys utilities ───────────────────────────────────────────────────────
	"sys.uuid":            {pkg: "sys", goName: "UUID", args: true, params: []string{}},
	"sys.random_password": {pkg: "sys", goName: "RandomPassword", args: true, params: []string{"i", "b", "b", "b"}},
	"sys.mac_address":     {pkg: "sys", goName: "MACAddress", args: true, params: []string{"s"}},
	"sys.mac_addresses":   {pkg: "sys", goName: "MACAddresses", args: true, params: []string{}},
	"sys.dmidecode":       {pkg: "sys", goName: "Dmidecode", args: true, params: []string{}},
	"sys.lspci":           {pkg: "sys", goName: "LsPci", args: true, params: []string{}},
	"sys.lsblk":           {pkg: "sys", goName: "LsBlk", args: true, params: []string{}},
	"sys.lsusb":           {pkg: "sys", goName: "LsUsb", args: true, params: []string{}},
	"sys.ip_route":        {pkg: "sys", goName: "IpRoute", args: true, params: []string{}},
	"sys.ethtool":         {pkg: "sys", goName: "Ethtool", args: true, params: []string{"s"}},

	// ── modprobe boot ─────────────────────────────────────────────────────────
	"modprobe.set_boot": {pkg: "modprobe", goName: "SetBoot", args: true, params: []string{"s", "b"}},

	// ── seport ─────────────────────────────────────────────────────────────────
	"seport.add":    {pkg: "seport", goName: "Add", args: true, params: []string{"s", "s", "s"}},
	"seport.remove": {pkg: "seport", goName: "Remove", args: true, params: []string{"s", "s"}},
	"seport.list":   {pkg: "seport", goName: "List", args: true, params: []string{}},
	"seport.get":    {pkg: "seport", goName: "Get", args: true, params: []string{"s", "s"}},

	// ── sefcontext ─────────────────────────────────────────────────────────────
	"sefcontext.add":    {pkg: "sefcontext", goName: "Add", args: true, params: []string{"s", "s"}},
	"sefcontext.modify": {pkg: "sefcontext", goName: "Modify", args: true, params: []string{"s", "s"}},
	"sefcontext.remove": {pkg: "sefcontext", goName: "Remove", args: true, params: []string{"s"}},
	"sefcontext.list":   {pkg: "sefcontext", goName: "List", args: true, params: []string{}},
	"sefcontext.apply":  {pkg: "sefcontext", goName: "Apply", args: true, params: []string{"s", "b"}},

	// ── flatpak ─────────────────────────────────────────────────────────────
	"flatpak.install": {pkg: "flatpak", goName: "Install", args: true, params: []string{"s", "s", "b"}},
	"flatpak.remove":  {pkg: "flatpak", goName: "Remove", args: true, params: []string{"s", "b"}},
	"flatpak.update":  {pkg: "flatpak", goName: "Update", args: true, params: []string{"s", "b"}},
	"flatpak.list":    {pkg: "flatpak", goName: "List", args: true, params: []string{"b"}},
	"flatpak.info":    {pkg: "flatpak", goName: "Info", args: true, params: []string{"s", "b"}},
	"flatpak.run":     {pkg: "flatpak", goName: "Run", args: true, params: []string{"s", "l", "b"}},
	"flatpak.repair":  {pkg: "flatpak", goName: "Repair", args: true, params: []string{"b"}},

	// ── zfs ─────────────────────────────────────────────────────────────
	"zfs.create":           {pkg: "zfs", goName: "Create", args: true, params: []string{"s", "ms"}},
	"zfs.destroy":          {pkg: "zfs", goName: "Destroy", args: true, params: []string{"s", "b"}},
	"zfs.set":              {pkg: "zfs", goName: "Set", args: true, params: []string{"s", "s", "s"}},
	"zfs.get":              {pkg: "zfs", goName: "Get", args: true, params: []string{"s", "s"}},
	"zfs.list":             {pkg: "zfs", goName: "List"},
	"zfs.exists":           {pkg: "zfs", goName: "Exists", args: true, params: []string{"s"}},
	"zfs.list_pools":       {pkg: "zfs", goName: "ListPools"},
	"zfs.get_pool_status":  {pkg: "zfs", goName: "GetPoolStatus", args: true, params: []string{"s"}},
	"zfs.snapshot":         {pkg: "zfs", goName: "Snapshot", args: true, params: []string{"s", "s"}},
	"zfs.destroy_snapshot": {pkg: "zfs", goName: "DestroySnapshot", args: true, params: []string{"s", "s"}},

	// ── nmcli ─────────────────────────────────────────────────────────────
	"nmcli.add":                {pkg: "nmcli", goName: "Add", args: true, params: []string{"s", "s", "ms"}},
	"nmcli.modify":             {pkg: "nmcli", goName: "Modify", args: true, params: []string{"s", "ms"}},
	"nmcli.delete":             {pkg: "nmcli", goName: "Delete", args: true, params: []string{"s"}},
	"nmcli.up":                 {pkg: "nmcli", goName: "Up", args: true, params: []string{"s"}},
	"nmcli.down":               {pkg: "nmcli", goName: "Down", args: true, params: []string{"s"}},
	"nmcli.list":               {pkg: "nmcli", goName: "List"},
	"nmcli.show":               {pkg: "nmcli", goName: "Show", args: true, params: []string{"s"}},
	"nmcli.list_devices":       {pkg: "nmcli", goName: "ListDevices"},
	"nmcli.reload":             {pkg: "nmcli", goName: "Reload"},
	"nmcli.get_general_status": {pkg: "nmcli", goName: "GetGeneralStatus"},

	// ── crypttab ──────────────────────────────────────────────────────────
	"crypttab.add":      {pkg: "crypttab", goName: "Add", args: true, params: []string{"s", "s", "s", "s"}},
	"crypttab.remove":   {pkg: "crypttab", goName: "Remove", args: true, params: []string{"s"}},
	"crypttab.modify":   {pkg: "crypttab", goName: "Modify", args: true, params: []string{"s", "s", "s", "s"}},
	"crypttab.get":      {pkg: "crypttab", goName: "Get", args: true, params: []string{"s"}},
	"crypttab.list":     {pkg: "crypttab", goName: "List"},
	"crypttab.exists":   {pkg: "crypttab", goName: "Exists", args: true, params: []string{"s"}},
	"crypttab.validate": {pkg: "crypttab", goName: "Validate"},
	"crypttab.backup":   {pkg: "crypttab", goName: "Backup", args: true, params: []string{"s"}},

	// ── sysfs ─────────────────────────────────────────────────────────────
	"sysfs.read":                {pkg: "sysfs", goName: "Read", args: true, params: []string{"s"}},
	"sysfs.write":               {pkg: "sysfs", goName: "Write", args: true, params: []string{"s", "s"}},
	"sysfs.exists":              {pkg: "sysfs", goName: "Exists", args: true, params: []string{"s"}},
	"sysfs.get":                 {pkg: "sysfs", goName: "Get", args: true, params: []string{"s"}},
	"sysfs.list":                {pkg: "sysfs", goName: "List", args: true, params: []string{"s"}},
	"sysfs.set_device_power":    {pkg: "sysfs", goName: "SetDevicePower", args: true, params: []string{"s", "s"}},
	"sysfs.get_device_power":    {pkg: "sysfs", goName: "GetDevicePower", args: true, params: []string{"s"}},
	"sysfs.set_kernel_parameter": {pkg: "sysfs", goName: "SetKernelParameter", args: true, params: []string{"s", "s"}},
	"sysfs.get_kernel_parameter": {pkg: "sysfs", goName: "GetKernelParameter", args: true, params: []string{"s"}},

	// pamd
	"pamd.get":          {pkg: "pamd", goName: "Get", args: true, params: []string{"s"}},
	"pamd.list":         {pkg: "pamd", goName: "List", args: true},
	"pamd.add_rule":     {pkg: "pamd", goName: "AddRule", args: true, params: []string{"s", "s", "s", "s", "s"}},
	"pamd.remove_rule":  {pkg: "pamd", goName: "RemoveRule", args: true, params: []string{"s", "s", "s"}},
	"pamd.modify_rule":  {pkg: "pamd", goName: "ModifyRule", args: true, params: []string{"s", "s", "s", "s", "s"}},
	"pamd.validate":     {pkg: "pamd", goName: "Validate", args: true, params: []string{"s"}},
	"pamd.backup":       {pkg: "pamd", goName: "Backup", args: true, params: []string{"s", "s"}},

	// getent
	"getent.passwd":           {pkg: "getent", goName: "GetPasswd", args: true},
	"getent.lookup_user":      {pkg: "getent", goName: "LookupUser", args: true, params: []string{"s"}},
	"getent.groups":           {pkg: "getent", goName: "GetGroups", args: true},
	"getent.lookup_group":     {pkg: "getent", goName: "LookupGroup", args: true, params: []string{"s"}},
	"getent.services":         {pkg: "getent", goName: "GetServices", args: true},
	"getent.lookup_service":   {pkg: "getent", goName: "LookupService", args: true, params: []string{"s"}},
	"getent.protocols":        {pkg: "getent", goName: "GetProtocols", args: true},
	"getent.lookup_protocol":  {pkg: "getent", goName: "LookupProtocol", args: true, params: []string{"s"}},
	"getent.shells":           {pkg: "getent", goName: "Shells", args: true},

	// haproxy
	"haproxy.get_status":      {pkg: "haproxy", goName: "GetStatus", args: true},
	"haproxy.list_backends":   {pkg: "haproxy", goName: "ListBackends", args: true, params: []string{"s"}},
	"haproxy.enable_backend":  {pkg: "haproxy", goName: "EnableBackend", args: true, params: []string{"s", "s", "s"}},
	"haproxy.disable_backend": {pkg: "haproxy", goName: "DisableBackend", args: true, params: []string{"s", "s", "s"}},
	"haproxy.validate_config": {pkg: "haproxy", goName: "ValidateConfig", args: true, params: []string{"s"}},
	"haproxy.reload":          {pkg: "haproxy", goName: "Reload", args: true, params: []string{"s"}},
	"haproxy.restart":         {pkg: "haproxy", goName: "Restart", args: true},
	"haproxy.version":         {pkg: "haproxy", goName: "Version", args: true},

	// openssl_cert
	"openssl_cert.create_csr":          {pkg: "openssl_cert", goName: "CreateCSR", args: true, params: []string{"s", "s", "s", "i"}},
	"openssl_cert.generate_self_signed": {pkg: "openssl_cert", goName: "GenerateSelfSigned", args: true, params: []string{"s", "s", "s", "i", "i"}},
	"openssl_cert.inspect":             {pkg: "openssl_cert", goName: "Inspect", args: true, params: []string{"s"}},
	"openssl_cert.verify":              {pkg: "openssl_cert", goName: "Verify", args: true, params: []string{"s", "s"}},
	"openssl_cert.check_expiry":        {pkg: "openssl_cert", goName: "CheckExpiry", args: true, params: []string{"s"}},
	"openssl_cert.convert_format":      {pkg: "openssl_cert", goName: "ConvertFormat", args: true, params: []string{"s", "s", "s"}},

	// redis
	"redis.ping":    {pkg: "redis", goName: "Ping", args: true, params: []string{"s", "i", "s"}},
	"redis.get":     {pkg: "redis", goName: "Get", args: true, params: []string{"s", "s", "i", "s"}},
	"redis.set":     {pkg: "redis", goName: "Set", args: true, params: []string{"s", "s", "s", "i", "s", "i"}},
	"redis.del":     {pkg: "redis", goName: "Del", args: true, params: []string{"l", "s", "i", "s"}},
	"redis.keys":    {pkg: "redis", goName: "Keys", args: true, params: []string{"s", "s", "i", "s"}},
	"redis.info":    {pkg: "redis", goName: "Info", args: true, params: []string{"s", "i", "s"}},
	"redis.flush_db": {pkg: "redis", goName: "FlushDB", args: true, params: []string{"s", "i", "s"}},

	// gem
	"gem.install":   {pkg: "gem", goName: "Install", args: true, params: []string{"s", "s", "b"}},
	"gem.uninstall": {pkg: "gem", goName: "Uninstall", args: true, params: []string{"s", "b"}},
	"gem.update":    {pkg: "gem", goName: "Update", args: true, params: []string{"s"}},
	"gem.info":      {pkg: "gem", goName: "Info", args: true, params: []string{"s"}},
	"gem.list":      {pkg: "gem", goName: "List", args: true},
	"gem.version":   {pkg: "gem", goName: "Version", args: true},

	// rabbitmq
	"rabbitmq.add_vhost":       {pkg: "rabbitmq", goName: "AddVhost", args: true, params: []string{"s"}},
	"rabbitmq.delete_vhost":    {pkg: "rabbitmq", goName: "DeleteVhost", args: true, params: []string{"s"}},
	"rabbitmq.list_vhosts":     {pkg: "rabbitmq", goName: "ListVhosts", args: true},
	"rabbitmq.add_user":        {pkg: "rabbitmq", goName: "AddUser", args: true, params: []string{"s", "s", "s"}},
	"rabbitmq.delete_user":     {pkg: "rabbitmq", goName: "DeleteUser", args: true, params: []string{"s"}},
	"rabbitmq.set_user_tags":   {pkg: "rabbitmq", goName: "SetUserTags", args: true, params: []string{"s", "s"}},
	"rabbitmq.list_users":      {pkg: "rabbitmq", goName: "ListUsers", args: true},
	"rabbitmq.set_permission":  {pkg: "rabbitmq", goName: "SetPermission", args: true, params: []string{"s", "s", "s", "s", "s"}},
	"rabbitmq.clear_permission": {pkg: "rabbitmq", goName: "ClearPermission", args: true, params: []string{"s", "s"}},
	"rabbitmq.set_policy":      {pkg: "rabbitmq", goName: "SetPolicy", args: true, params: []string{"s", "s", "s", "s", "s"}},
	"rabbitmq.delete_policy":   {pkg: "rabbitmq", goName: "DeletePolicy", args: true, params: []string{"s", "s"}},
	"rabbitmq.declare_queue":   {pkg: "rabbitmq", goName: "DeclareQueue", args: true, params: []string{"s", "s", "s", "b", "b"}},
	"rabbitmq.delete_queue":    {pkg: "rabbitmq", goName: "DeleteQueue", args: true, params: []string{"s", "s"}},
	"rabbitmq.declare_exchange": {pkg: "rabbitmq", goName: "DeclareExchange", args: true, params: []string{"s", "s", "s", "b", "b"}},
	"rabbitmq.delete_exchange": {pkg: "rabbitmq", goName: "DeleteExchange", args: true, params: []string{"s", "s"}},
	"rabbitmq.bind_queue":      {pkg: "rabbitmq", goName: "BindQueue", args: true, params: []string{"s", "s", "s", "s"}},
	"rabbitmq.unbind_queue":    {pkg: "rabbitmq", goName: "UnbindQueue", args: true, params: []string{"s", "s", "s", "s"}},
	"rabbitmq.get_status":      {pkg: "rabbitmq", goName: "GetStatus", args: true},

	// consul
	"consul.kv_get":           {pkg: "consul", goName: "KVGet", args: true, params: []string{"s", "s"}},
	"consul.kv_put":           {pkg: "consul", goName: "KVPut", args: true, params: []string{"s", "s", "s"}},
	"consul.kv_delete":        {pkg: "consul", goName: "KVDelete", args: true, params: []string{"s", "s"}},
	"consul.kv_list":          {pkg: "consul", goName: "KVList", args: true, params: []string{"s", "s"}},
	"consul.service_register": {pkg: "consul", goName: "ServiceRegister", args: true, params: []string{"s", "s", "s", "s", "s"}},
	"consul.service_deregister": {pkg: "consul", goName: "ServiceDeregister", args: true, params: []string{"s", "s"}},
	"consul.members":          {pkg: "consul", goName: "Members", args: true, params: []string{"s"}},
	"consul.info":             {pkg: "consul", goName: "Info", args: true, params: []string{"s"}},
	"consul.health_check":     {pkg: "consul", goName: "HealthCheck", args: true, params: []string{"s", "s"}},
	"consul.version":          {pkg: "consul", goName: "Version", args: true},

	// memcached
	"memcached.get":        {pkg: "memcached", goName: "Get", args: true, params: []string{"s", "s", "i"}},
	"memcached.set":        {pkg: "memcached", goName: "Set", args: true, params: []string{"s", "s", "s", "i", "i"}},
	"memcached.delete":     {pkg: "memcached", goName: "Delete", args: true, params: []string{"s", "s", "i"}},
	"memcached.flush_all":  {pkg: "memcached", goName: "FlushAll", args: true, params: []string{"s", "i"}},
	"memcached.stats":      {pkg: "memcached", goName: "Stats", args: true, params: []string{"s", "i"}},
	"memcached.version":    {pkg: "memcached", goName: "Version", args: true, params: []string{"s", "i"}},

	// composer
	"composer.install":       {pkg: "composer", goName: "Install", args: true, params: []string{"s", "b"}},
	"composer.update":        {pkg: "composer", goName: "Update", args: true, params: []string{"s", "b"}},
	"composer.require":       {pkg: "composer", goName: "Require", args: true, params: []string{"s", "s", "s"}},
	"composer.remove":        {pkg: "composer", goName: "Remove", args: true, params: []string{"s", "s"}},
	"composer.create_project": {pkg: "composer", goName: "CreateProject", args: true, params: []string{"s", "s", "s"}},
	"composer.global_install": {pkg: "composer", goName: "GlobalInstall", args: true, params: []string{"s", "s"}},
	"composer.version":       {pkg: "composer", goName: "Version", args: true},

	// cargo
	"cargo.install":    {pkg: "cargo", goName: "Install", args: true, params: []string{"s", "s", "b"}},
	"cargo.uninstall":  {pkg: "cargo", goName: "Uninstall", args: true, params: []string{"s"}},
	"cargo.update":     {pkg: "cargo", goName: "Update", args: true, params: []string{"s"}},
	"cargo.list":       {pkg: "cargo", goName: "List", args: true},
	"cargo.build":      {pkg: "cargo", goName: "Build", args: true, params: []string{"s", "b"}},
	"cargo.test":       {pkg: "cargo", goName: "Test", args: true, params: []string{"s"}},
	"cargo.version":    {pkg: "cargo", goName: "Version", args: true},

	// rpmkey
	"rpmkey.import": {pkg: "rpmkey", goName: "Import", args: true, params: []string{"s"}},
	"rpmkey.list":   {pkg: "rpmkey", goName: "List", args: true},
	"rpmkey.remove": {pkg: "rpmkey", goName: "Remove", args: true, params: []string{"s"}},

	// aptkey
	"aptkey.add":          {pkg: "aptkey", goName: "Add", args: true, params: []string{"s", "s"}},
	"aptkey.add_from_key": {pkg: "aptkey", goName: "AddFromKey", args: true, params: []string{"s", "s"}},
	"aptkey.remove":       {pkg: "aptkey", goName: "Remove", args: true, params: []string{"s", "s"}},
	"aptkey.list":         {pkg: "aptkey", goName: "List", args: true},

	// dmidecode
	"dmidecode.system":    {pkg: "dmidecode", goName: "System", args: true},
	"dmidecode.bios":      {pkg: "dmidecode", goName: "BIOS", args: true},
	"dmidecode.chassis":   {pkg: "dmidecode", goName: "Chassis", args: true},
	"dmidecode.processor": {pkg: "dmidecode", goName: "Processor", args: true},
	"dmidecode.keyword":   {pkg: "dmidecode", goName: "Keyword", args: true, params: []string{"s"}},

	// tuned
	"tuned.set":    {pkg: "tuned", goName: "Set", args: true, params: []string{"s"}},
	"tuned.status": {pkg: "tuned", goName: "Status", args: true},
	"tuned.list":   {pkg: "tuned", goName: "List", args: true},
	"tuned.off":    {pkg: "tuned", goName: "Off", args: true},
	"tuned.profile": {pkg: "tuned", goName: "Profile", args: true},
	"tuned.verify": {pkg: "tuned", goName: "Verify", args: true},

	// supervisor
	"supervisor.start":      {pkg: "supervisor", goName: "Start", args: true, params: []string{"s"}},
	"supervisor.stop":       {pkg: "supervisor", goName: "Stop", args: true, params: []string{"s"}},
	"supervisor.restart":    {pkg: "supervisor", goName: "Restart", args: true, params: []string{"s"}},
	"supervisor.reload":     {pkg: "supervisor", goName: "Reload", args: true},
	"supervisor.status":     {pkg: "supervisor", goName: "Status", args: true},
	"supervisor.clear_log":  {pkg: "supervisor", goName: "ClearLog", args: true, params: []string{"s"}},
	"supervisor.reread":     {pkg: "supervisor", goName: "Reread", args: true},
	"supervisor.update":     {pkg: "supervisor", goName: "Update", args: true, params: []string{"s"}},

	// smartctl
	"smartctl.device":     {pkg: "smartctl", goName: "Device", args: true, params: []string{"s"}},
	"smartctl.health":     {pkg: "smartctl", goName: "Health", args: true, params: []string{"s"}},
	"smartctl.attributes": {pkg: "smartctl", goName: "Attributes", args: true, params: []string{"s"}},
	"smartctl.list":       {pkg: "smartctl", goName: "List"},
	"smartctl.json":       {pkg: "smartctl", goName: "JSON", args: true, params: []string{"s"}},

	// virsh
	"virsh.start":   {pkg: "virsh", goName: "Start", args: true, params: []string{"s"}},
	"virsh.stop":    {pkg: "virsh", goName: "Stop", args: true, params: []string{"s"}},
	"virsh.reboot":  {pkg: "virsh", goName: "Reboot", args: true, params: []string{"s"}},
	"virsh.shutdown": {pkg: "virsh", goName: "Shutdown", args: true, params: []string{"s"}},
	"virsh.suspend": {pkg: "virsh", goName: "Suspend", args: true, params: []string{"s"}},
	"virsh.resume":  {pkg: "virsh", goName: "Resume", args: true, params: []string{"s"}},
	"virsh.list":    {pkg: "virsh", goName: "List"},
	"virsh.info":    {pkg: "virsh", goName: "Info", args: true, params: []string{"s"}},
	"virsh.version": {pkg: "virsh", goName: "Version"},

	// ethtool
	"ethtool.show":        {pkg: "ethtool", goName: "Show", args: true, params: []string{"s"}},
	"ethtool.set_speed":   {pkg: "ethtool", goName: "SetSpeed", args: true, params: []string{"s", "s"}},
	"ethtool.set_duplex":  {pkg: "ethtool", goName: "SetDuplex", args: true, params: []string{"s", "s"}},
	"ethtool.set_autoneg": {pkg: "ethtool", goName: "SetAutoneg", args: true, params: []string{"s", "s"}},
	"ethtool.set_pause":   {pkg: "ethtool", goName: "SetPause", args: true, params: []string{"s", "s", "s"}},
	"ethtool.set_offload": {pkg: "ethtool", goName: "SetOffload", args: true, params: []string{"s", "s", "s"}},

	// systemd_analyze
	"systemd_analyze.time":          {pkg: "systemd_analyze", goName: "Time"},
	"systemd_analyze.blame":         {pkg: "systemd_analyze", goName: "Blame"},
	"systemd_analyze.critical_chain": {pkg: "systemd_analyze", goName: "CriticalChain"},
	"systemd_analyze.security":      {pkg: "systemd_analyze", goName: "Security", args: true, params: []string{"s"}},
	"systemd_analyze.verify":        {pkg: "systemd_analyze", goName: "Verify", args: true, params: []string{"s"}},

	// nvme
	"nvme.list":         {pkg: "nvme", goName: "List"},
	"nvme.smart_log":    {pkg: "nvme", goName: "SmartLog", args: true, params: []string{"s"}},
	"nvme.firmware_log": {pkg: "nvme", goName: "FirmwareLog", args: true, params: []string{"s"}},
	"nvme.error_log":    {pkg: "nvme", goName: "ErrorLog", args: true, params: []string{"s"}},
	"nvme.version":      {pkg: "nvme", goName: "Version"},

	// lshw
	"lshw.short":   {pkg: "lshw", goName: "Short"},
	"lshw.class":   {pkg: "lshw", goName: "Class", args: true, params: []string{"s"}},
	"lshw.json":    {pkg: "lshw", goName: "JSON"},
	"lshw.system":  {pkg: "lshw", goName: "System"},
	"lshw.memory":  {pkg: "lshw", goName: "Memory"},
	"lshw.disk":    {pkg: "lshw", goName: "Disk"},
	"lshw.network": {pkg: "lshw", goName: "Network"},

	// ipaddr
	"ipaddr.list":           {pkg: "ipaddr", goName: "List"},
	"ipaddr.list_interface": {pkg: "ipaddr", goName: "ListInterface", args: true, params: []string{"s"}},
	"ipaddr.add":            {pkg: "ipaddr", goName: "Add", args: true, params: []string{"s", "s"}},
	"ipaddr.delete":         {pkg: "ipaddr", goName: "Delete", args: true, params: []string{"s", "s"}},
	"ipaddr.flush":          {pkg: "ipaddr", goName: "Flush", args: true, params: []string{"s"}},
	"ipaddr.links":          {pkg: "ipaddr", goName: "Links"},
	"ipaddr.link_up":        {pkg: "ipaddr", goName: "LinkUp", args: true, params: []string{"s"}},
	"ipaddr.link_down":      {pkg: "ipaddr", goName: "LinkDown", args: true, params: []string{"s"}},

	// udevadm
	"udevadm.control": {pkg: "udevadm", goName: "Control", args: true, params: []string{"s"}},
	"udevadm.trigger": {pkg: "udevadm", goName: "Trigger", args: true, params: []string{"s"}},
	"udevadm.settle":  {pkg: "udevadm", goName: "Settle", args: true, params: []string{"i"}},
	"udevadm.info":    {pkg: "udevadm", goName: "Info", args: true, params: []string{"s", "s"}},
	"udevadm.monitor": {pkg: "udevadm", goName: "Monitor"},

	// modinfo
	"modinfo.info":    {pkg: "modinfo", goName: "Info", args: true, params: []string{"s"}},
	"modinfo.list":    {pkg: "modinfo", goName: "List"},
	"modinfo.version": {pkg: "modinfo", goName: "Version"},

	// dconf
	"dconf.read":  {pkg: "dconf", goName: "Read", args: true, params: []string{"s"}},
	"dconf.write": {pkg: "dconf", goName: "Write", args: true, params: []string{"s", "s"}},
	"dconf.list":  {pkg: "dconf", goName: "List", args: true, params: []string{"s"}},
	"dconf.reset": {pkg: "dconf", goName: "Reset", args: true, params: []string{"s"}},

	// locale_gen
	"locale_gen.generate": {pkg: "locale_gen", goName: "Generate", args: true, params: []string{"s"}},
	"locale_gen.list":     {pkg: "locale_gen", goName: "List"},
	"locale_gen.remove":   {pkg: "locale_gen", goName: "Remove", args: true, params: []string{"s"}},

	// pam_limits
	"pam_limits.set":  {pkg: "pam_limits", goName: "Set", args: true, params: []string{"s", "s", "s", "s"}},
	"pam_limits.list": {pkg: "pam_limits", goName: "List"},

	// motd
	"motd.read":  {pkg: "motd", goName: "Read"},
	"motd.write": {pkg: "motd", goName: "Write", args: true, params: []string{"s"}},

	// issue
	"issue.read":  {pkg: "issue", goName: "Read"},
	"issue.write": {pkg: "issue", goName: "Write", args: true, params: []string{"s"}},

	// authorized_key
	"authorized_key.manage": {pkg: "authorized_key", goName: "Manage", args: true, params: []string{"s", "s", "s", "s"}},
	"authorized_key.list":   {pkg: "authorized_key", goName: "List", args: true, params: []string{"s", "s"}},
	"authorized_key.check":  {pkg: "authorized_key", goName: "Check", args: true, params: []string{"s", "s", "s"}},

	// blockinfile
	"blockinfile.manage": {pkg: "blockinfile", goName: "Manage", args: true, params: []string{"s", "s", "s", "s", "s", "s"}},
	"blockinfile.read":   {pkg: "blockinfile", goName: "Read", args: true, params: []string{"s", "s"}},

	// debconf
	"debconf.set":  {pkg: "debconf", goName: "Set", args: true, params: []string{"s", "s", "s", "s"}},
	"debconf.get":  {pkg: "debconf", goName: "Get", args: true, params: []string{"s", "s"}},
	"debconf.list": {pkg: "debconf", goName: "List", args: true, params: []string{"s"}},

	// reboot
	"reboot.request": {pkg: "reboot", goName: "Request", args: true, params: []string{"s", "i"}},
	"reboot.dry_run": {pkg: "reboot", goName: "DryRun", args: true, params: []string{"s", "i"}},
	"reboot.check":   {pkg: "reboot", goName: "Check"},

	// swap
	"swap.info":    {pkg: "swap", goName: "Info"},
	"swap.create":  {pkg: "swap", goName: "Create", args: true, params: []string{"s", "i"}},
	"swap.enable":  {pkg: "swap", goName: "Enable", args: true, params: []string{"s"}},
	"swap.disable": {pkg: "swap", goName: "Disable", args: true, params: []string{"s"}},

	// raw
	"raw.execute":        {pkg: "raw", goName: "Execute", args: true, params: []string{"s", "i"}},
	"raw.execute_with_env": {pkg: "raw", goName: "ExecuteWithEnv", args: true, params: []string{"s", "i", "ms"}},

	// expect
	"expect.run":        {pkg: "expect", goName: "Run", args: true, params: []string{"s", "ms", "i"}},
	"expect.run_simple": {pkg: "expect", goName: "RunSimple", args: true, params: []string{"s", "s", "s", "i"}},

	// slurp
	"slurp.encode": {pkg: "slurp", goName: "Encode", args: true, params: []string{"s"}},
	"slurp.decode": {pkg: "slurp", goName: "Decode", args: true, params: []string{"s", "s"}},

	// wait_for_connection
	"wait_for_connection.wait":       {pkg: "wait_for_connection", goName: "Wait", args: true, params: []string{"s", "i", "i", "i"}},
	"wait_for_connection.check_once": {pkg: "wait_for_connection", goName: "CheckOnce", args: true, params: []string{"s", "i"}},

	// firewalld_rich_rule
	"firewalld_rich_rule.add":     {pkg: "firewalld_rich_rule", goName: "Add", args: true, params: []string{"s", "s"}},
	"firewalld_rich_rule.remove":  {pkg: "firewalld_rich_rule", goName: "Remove", args: true, params: []string{"s", "s"}},
	"firewalld_rich_rule.list":    {pkg: "firewalld_rich_rule", goName: "List", args: true, params: []string{"s"}},
	"firewalld_rich_rule.exists":  {pkg: "firewalld_rich_rule", goName: "Exists", args: true, params: []string{"s", "s"}},

	// firewalld_ipset
	"firewalld_ipset.create":       {pkg: "firewalld_ipset", goName: "Create", args: true, params: []string{"s", "s"}},
	"firewalld_ipset.delete":       {pkg: "firewalld_ipset", goName: "Delete", args: true, params: []string{"s"}},
	"firewalld_ipset.add_entry":    {pkg: "firewalld_ipset", goName: "AddEntry", args: true, params: []string{"s", "s"}},
	"firewalld_ipset.remove_entry": {pkg: "firewalld_ipset", goName: "RemoveEntry", args: true, params: []string{"s", "s"}},
	"firewalld_ipset.list":         {pkg: "firewalld_ipset", goName: "List"},
	"firewalld_ipset.info":         {pkg: "firewalld_ipset", goName: "Info", args: true, params: []string{"s"}},

	// pause
	"pause.seconds":            {pkg: "pause", goName: "Seconds", args: true, params: []string{"i"}},
	"pause.prompt":             {pkg: "pause", goName: "Prompt", args: true, params: []string{"s"}},
	"pause.prompt_with_default": {pkg: "pause", goName: "PromptWithDefault", args: true, params: []string{"s", "s"}},

	// meta
	"meta.end_host":         {pkg: "meta", goName: "EndHost"},
	"meta.end_play":         {pkg: "meta", goName: "EndPlay"},
	"meta.clear_host_errors": {pkg: "meta", goName: "ClearHostErrors"},
	"meta.refresh_inventory": {pkg: "meta", goName: "RefreshInventory"},
	"meta.flush_handlers":   {pkg: "meta", goName: "FlushHandlers"},
	"meta.reset_connection": {pkg: "meta", goName: "ResetConnection"},
	"meta.noop":             {pkg: "meta", goName: "Noop"},
	"meta.fail":             {pkg: "meta", goName: "Fail", args: true, params: []string{"s"}},
	"meta.assert":           {pkg: "meta", goName: "Assert", args: true, params: []string{"b", "s"}},
	"meta.debug":            {pkg: "meta", goName: "Debug", args: true, params: []string{"s", "ms"}},

	// uri_ext
	"uri_ext.patch":   {pkg: "uri_ext", goName: "Patch", args: true, params: []string{"s", "s", "ms", "i"}},
	"uri_ext.delete":  {pkg: "uri_ext", goName: "Delete", args: true, params: []string{"s", "ms", "i"}},
	"uri_ext.head":    {pkg: "uri_ext", goName: "Head", args: true, params: []string{"s", "ms", "i"}},
	"uri_ext.options": {pkg: "uri_ext", goName: "Options", args: true, params: []string{"s", "ms", "i"}},

	// hwclock
	"hwclock.get":      {pkg: "hwclock", goName: "Get"},
	"hwclock.set":      {pkg: "hwclock", goName: "Set"},
	"hwclock.hctosys":  {pkg: "hwclock", goName: "HCToSys"},
	"hwclock.set_time": {pkg: "hwclock", goName: "SetTime", args: true, params: []string{"s"}},

	// mdadm
	"mdadm.create":  {pkg: "mdadm", goName: "Create", args: true, params: []string{"s", "s", "l"}},
	"mdadm.destroy": {pkg: "mdadm", goName: "Destroy", args: true, params: []string{"s"}},
	"mdadm.detail":  {pkg: "mdadm", goName: "Detail", args: true, params: []string{"s"}},
	"mdadm.scan":    {pkg: "mdadm", goName: "Scan"},
	"mdadm.add":     {pkg: "mdadm", goName: "Add", args: true, params: []string{"s", "s"}},
	"mdadm.remove":  {pkg: "mdadm", goName: "Remove", args: true, params: []string{"s", "s"}},

	// open_iscsi
	"open_iscsi.discover":   {pkg: "open_iscsi", goName: "Discover", args: true, params: []string{"s", "i"}},
	"open_iscsi.login":      {pkg: "open_iscsi", goName: "Login", args: true, params: []string{"s", "s"}},
	"open_iscsi.logout":     {pkg: "open_iscsi", goName: "Logout", args: true, params: []string{"s", "s"}},
	"open_iscsi.list_sessions": {pkg: "open_iscsi", goName: "ListSessions"},
	"open_iscsi.list_nodes":    {pkg: "open_iscsi", goName: "ListNodes"},
	"open_iscsi.set_startup":   {pkg: "open_iscsi", goName: "SetStartup", args: true, params: []string{"s", "s", "s"}},

	// rfkill
	"rfkill.list":       {pkg: "rfkill", goName: "List"},
	"rfkill.block":      {pkg: "rfkill", goName: "Block", args: true, params: []string{"s"}},
	"rfkill.unblock":    {pkg: "rfkill", goName: "Unblock", args: true, params: []string{"s"}},
	"rfkill.block_all":  {pkg: "rfkill", goName: "BlockAll", args: true, params: []string{"s"}},
	"rfkill.unblock_all": {pkg: "rfkill", goName: "UnblockAll", args: true, params: []string{"s"}},

	// multipath
	"multipath.reconfigure": {pkg: "multipath", goName: "Reconfigure"},
	"multipath.list_paths":  {pkg: "multipath", goName: "ListPaths"},
	"multipath.list_maps":   {pkg: "multipath", goName: "ListMaps"},
	"multipath.add_map":     {pkg: "multipath", goName: "AddMap", args: true, params: []string{"s"}},
	"multipath.remove_map":  {pkg: "multipath", goName: "RemoveMap", args: true, params: []string{"s"}},
	"multipath.flush":       {pkg: "multipath", goName: "Flush"},

	// dmsetup
	"dmsetup.create":     {pkg: "dmsetup", goName: "Create", args: true, params: []string{"s", "s"}},
	"dmsetup.remove":     {pkg: "dmsetup", goName: "Remove", args: true, params: []string{"s"}},
	"dmsetup.remove_all": {pkg: "dmsetup", goName: "RemoveAll"},
	"dmsetup.list":       {pkg: "dmsetup", goName: "List"},
	"dmsetup.info":       {pkg: "dmsetup", goName: "Info", args: true, params: []string{"s"}},
	"dmsetup.suspend":    {pkg: "dmsetup", goName: "Suspend", args: true, params: []string{"s"}},
	"dmsetup.resume":     {pkg: "dmsetup", goName: "Resume", args: true, params: []string{"s"}},

	// lvm_enhanced
	"lvm_enhanced.pv_create":     {pkg: "lvm_enhanced", goName: "PVCreate", args: true, params: []string{"s"}},
	"lvm_enhanced.pv_remove":     {pkg: "lvm_enhanced", goName: "PVRemove", args: true, params: []string{"s", "b"}},
	"lvm_enhanced.pv_list":       {pkg: "lvm_enhanced", goName: "PVList"},
	"lvm_enhanced.vg_create":     {pkg: "lvm_enhanced", goName: "VGCreate", args: true, params: []string{"s", "l"}},
	"lvm_enhanced.vg_remove":     {pkg: "lvm_enhanced", goName: "VGRemove", args: true, params: []string{"s", "b"}},
	"lvm_enhanced.vg_extend":     {pkg: "lvm_enhanced", goName: "VGExtend", args: true, params: []string{"s", "s"}},
	"lvm_enhanced.vg_list":       {pkg: "lvm_enhanced", goName: "VGList"},
	"lvm_enhanced.lv_extend":     {pkg: "lvm_enhanced", goName: "LVExtend", args: true, params: []string{"s", "s"}},
	"lvm_enhanced.lv_extend_all": {pkg: "lvm_enhanced", goName: "LVExtendAll", args: true, params: []string{"s"}},
	"lvm_enhanced.lv_list":       {pkg: "lvm_enhanced", goName: "LVList"},

	// puppet
	"puppet.run":            {pkg: "puppet", goName: "Run", args: true, params: []string{"s", "l"}},
	"puppet.run_noop":       {pkg: "puppet", goName: "RunNoop", args: true, params: []string{"s", "l"}},
	"puppet.status":         {pkg: "puppet", goName: "Status"},
	"puppet.disable":        {pkg: "puppet", goName: "Disable", args: true, params: []string{"s"}},
	"puppet.enable":         {pkg: "puppet", goName: "Enable"},
	"puppet.fact":           {pkg: "puppet", goName: "Fact", args: true, params: []string{"s"}},
	"puppet.module_list":    {pkg: "puppet", goName: "ModuleList"},
	"puppet.module_install": {pkg: "puppet", goName: "ModuleInstall", args: true, params: []string{"s", "s"}},

	// new functions
	"pip.freeze":              {pkg: "pip", goName: "Freeze", args: true, params: []string{"s"}},
	"pip.install_requirements": {pkg: "pip", goName: "InstallRequirements", args: true, params: []string{"s", "s"}},
	"flatpak.add_remote":      {pkg: "flatpak", goName: "AddRemote", args: true, params: []string{"s", "s"}},
	"yarn.install":            {pkg: "yarn", goName: "Install", args: true, params: []string{"s", "s", "b"}},
	"yarn.remove":             {pkg: "yarn", goName: "Remove", args: true, params: []string{"s", "b"}},
	"yarn.global":             {pkg: "yarn", goName: "Global", args: true, params: []string{"s"}},
	"yarn.list":               {pkg: "yarn", goName: "List", args: true, params: []string{"b"}},
	"htpasswd.set":            {pkg: "htpasswd", goName: "Set", args: true, params: []string{"s", "s", "s", "b"}},
	"htpasswd.remove":         {pkg: "htpasswd", goName: "Remove", args: true, params: []string{"s", "s"}},
	"htpasswd.info":           {pkg: "htpasswd", goName: "Info", args: true, params: []string{"s"}},
	"htpasswd.hash_sha1":      {pkg: "htpasswd", goName: "HashSHA1", args: true, params: []string{"s"}},
	"sudoers.set":             {pkg: "sudoers", goName: "Set", args: true, params: []string{"s", "s", "s", "b", "s"}},
	"sudoers.remove":          {pkg: "sudoers", goName: "Remove", args: true, params: []string{"s", "s"}},
	"sudoers.info":            {pkg: "sudoers", goName: "Info", args: true, params: []string{"s", "s"}},
	"monit.start":             {pkg: "monit", goName: "Start", args: true, params: []string{"s"}},
	"monit.stop":              {pkg: "monit", goName: "Stop", args: true, params: []string{"s"}},
	"monit.monitor":           {pkg: "monit", goName: "Monitor", args: true, params: []string{"s"}},
	"monit.unmonitor":         {pkg: "monit", goName: "Unmonitor", args: true, params: []string{"s"}},
	"monit.restart":           {pkg: "monit", goName: "Restart", args: true, params: []string{"s"}},
	"monit.status":            {pkg: "monit", goName: "Status"},
	"monit.reload":            {pkg: "monit", goName: "Reload"},

	// svn
	"svn.checkout": {pkg: "svn", goName: "Checkout", args: true, params: []string{"s", "s", "s", "b"}},
	"svn.update":   {pkg: "svn", goName: "Update", args: true, params: []string{"s", "s"}},
	"svn.export":   {pkg: "svn", goName: "Export", args: true, params: []string{"s", "s", "s", "b"}},
	"svn.status":   {pkg: "svn", goName: "Status", args: true, params: []string{"s"}},
	"svn.info":     {pkg: "svn", goName: "Info", args: true, params: []string{"s"}},
	"svn.cleanup":  {pkg: "svn", goName: "Cleanup", args: true, params: []string{"s"}},
	"svn.revert":   {pkg: "svn", goName: "Revert", args: true, params: []string{"s", "b"}},

	// zypper
	"zypper.install":         {pkg: "zypper", goName: "Install", args: true, params: []string{"s", "s"}},
	"zypper.remove":          {pkg: "zypper", goName: "Remove", args: true, params: []string{"s"}},
	"zypper.update":          {pkg: "zypper", goName: "Update", args: true, params: []string{"s"}},
	"zypper.dist_upgrade":    {pkg: "zypper", goName: "DistUpgrade"},
	"zypper.info":            {pkg: "zypper", goName: "Info", args: true, params: []string{"s"}},
	"zypper.list":            {pkg: "zypper", goName: "List"},
	"zypper.clean":           {pkg: "zypper", goName: "Clean"},
	"zypper.repo_list":       {pkg: "zypper", goName: "RepoList"},
	"zypper.repo_add":        {pkg: "zypper", goName: "RepoAdd", args: true, params: []string{"s", "s"}},
	"zypper.repo_remove":     {pkg: "zypper", goName: "RepoRemove", args: true, params: []string{"s"}},
	"zypper.refresh":         {pkg: "zypper", goName: "Refresh"},
	"zypper.search":          {pkg: "zypper", goName: "Search", args: true, params: []string{"s"}},
	"zypper.patch":           {pkg: "zypper", goName: "Patch"},
	"zypper.pattern_install": {pkg: "zypper", goName: "PatternInstall", args: true, params: []string{"s"}},
	"zypper.pattern_remove":  {pkg: "zypper", goName: "PatternRemove", args: true, params: []string{"s"}},

	// pacman
	"pacman.install":         {pkg: "pacman", goName: "Install", args: true, params: []string{"s"}},
	"pacman.remove":          {pkg: "pacman", goName: "Remove", args: true, params: []string{"s", "b"}},
	"pacman.update":          {pkg: "pacman", goName: "Update", args: true, params: []string{"s"}},
	"pacman.upgrade":         {pkg: "pacman", goName: "Upgrade"},
	"pacman.info":            {pkg: "pacman", goName: "Info", args: true, params: []string{"s"}},
	"pacman.list":            {pkg: "pacman", goName: "List"},
	"pacman.search":          {pkg: "pacman", goName: "Search", args: true, params: []string{"s"}},
	"pacman.clean":           {pkg: "pacman", goName: "Clean"},
	"pacman.install_file":    {pkg: "pacman", goName: "InstallFile", args: true, params: []string{"s"}},
	"pacman.remove_orphans":  {pkg: "pacman", goName: "RemoveOrphans"},
	"pacman.update_database": {pkg: "pacman", goName: "UpdateDatabase"},

	// kubernetes
	"kubernetes.apply":           {pkg: "kubernetes", goName: "Apply", args: true, params: []string{"s", "s", "b"}},
	"kubernetes.delete":          {pkg: "kubernetes", goName: "Delete", args: true, params: []string{"s", "s"}},
	"kubernetes.get":             {pkg: "kubernetes", goName: "Get", args: true, params: []string{"s", "s", "s"}},
	"kubernetes.list":            {pkg: "kubernetes", goName: "List", args: true, params: []string{"s", "s", "s"}},
	"kubernetes.create_namespace": {pkg: "kubernetes", goName: "CreateNamespace", args: true, params: []string{"s"}},
	"kubernetes.delete_namespace": {pkg: "kubernetes", goName: "DeleteNamespace", args: true, params: []string{"s"}},
	"kubernetes.get_pods":        {pkg: "kubernetes", goName: "GetPods", args: true, params: []string{"s", "s"}},
	"kubernetes.get_services":    {pkg: "kubernetes", goName: "GetServices", args: true, params: []string{"s"}},
	"kubernetes.get_deployments": {pkg: "kubernetes", goName: "GetDeployments", args: true, params: []string{"s"}},
	"kubernetes.scale":           {pkg: "kubernetes", goName: "Scale", args: true, params: []string{"s", "s", "s"}},
	"kubernetes.rollout_status":  {pkg: "kubernetes", goName: "RolloutStatus", args: true, params: []string{"s", "s"}},
	"kubernetes.exec":            {pkg: "kubernetes", goName: "Exec", args: true, params: []string{"s", "s", "s", "s"}},
	"kubernetes.logs":            {pkg: "kubernetes", goName: "Logs", args: true, params: []string{"s", "s", "s", "s"}},
	"kubernetes.wait_ready":      {pkg: "kubernetes", goName: "WaitReady", args: true, params: []string{"s", "s", "s", "s"}},

	"portage.install":   {pkg: "portage", goName: "Install", args: true, params: []string{"s", "s"}},
	"portage.remove":    {pkg: "portage", goName: "Remove", args: true, params: []string{"s"}},
	"portage.update":    {pkg: "portage", goName: "Update", args: true, params: []string{"s", "b"}},
	"portage.sync":      {pkg: "portage", goName: "Sync", args: false},
	"portage.info":      {pkg: "portage", goName: "Info", args: true, params: []string{"s"}},
	"portage.list":      {pkg: "portage", goName: "List", args: false},
	"portage.search":    {pkg: "portage", goName: "Search", args: true, params: []string{"s"}},
	"portage.depclean":  {pkg: "portage", goName: "Depclean", args: false},
	"portage.metadata":  {pkg: "portage", goName: "Metadata", args: true, params: []string{"s"}},

	"pkgng.install":   {pkg: "pkgng", goName: "Install", args: true, params: []string{"s", "s"}},
	"pkgng.remove":    {pkg: "pkgng", goName: "Remove", args: true, params: []string{"s"}},
	"pkgng.update":    {pkg: "pkgng", goName: "Update", args: false},
	"pkgng.upgrade":   {pkg: "pkgng", goName: "Upgrade", args: true, params: []string{"s"}},
	"pkgng.autoclean": {pkg: "pkgng", goName: "Autoclean", args: false},
	"pkgng.info":      {pkg: "pkgng", goName: "Info", args: true, params: []string{"s"}},
	"pkgng.list":      {pkg: "pkgng", goName: "List", args: false},
	"pkgng.search":    {pkg: "pkgng", goName: "Search", args: true, params: []string{"s"}},
	"pkgng.stats":     {pkg: "pkgng", goName: "Stats", args: false},

	"podman.run":              {pkg: "podman", goName: "Run", args: true, params: []string{"s", "s", "s"}},
	"podman.stop":             {pkg: "podman", goName: "Stop", args: true, params: []string{"s", "i"}},
	"podman.start":            {pkg: "podman", goName: "Start", args: true, params: []string{"s"}},
	"podman.remove":           {pkg: "podman", goName: "Remove", args: true, params: []string{"s", "b"}},
	"podman.list_containers":  {pkg: "podman", goName: "ListContainers", args: true, params: []string{"b"}},
	"podman.inspect":          {pkg: "podman", goName: "Inspect", args: true, params: []string{"s"}},
	"podman.pull":             {pkg: "podman", goName: "Pull", args: true, params: []string{"s"}},
	"podman.list_images":      {pkg: "podman", goName: "ListImages", args: false},
	"podman.remove_image":     {pkg: "podman", goName: "RemoveImage", args: true, params: []string{"s", "b"}},
	"podman.create_pod":       {pkg: "podman", goName: "CreatePod", args: true, params: []string{"s"}},
	"podman.stop_pod":         {pkg: "podman", goName: "StopPod", args: true, params: []string{"s"}},
	"podman.remove_pod":       {pkg: "podman", goName: "RemovePod", args: true, params: []string{"s", "b"}},
	"podman.list_pods":        {pkg: "podman", goName: "ListPods", args: false},

	"nftables.add_table":      {pkg: "nftables", goName: "AddTable", args: true, params: []string{"s", "s"}},
	"nftables.delete_table":   {pkg: "nftables", goName: "DeleteTable", args: true, params: []string{"s", "s"}},
	"nftables.list_tables":    {pkg: "nftables", goName: "ListTables", args: false},
	"nftables.add_chain":      {pkg: "nftables", goName: "AddChain", args: true, params: []string{"s", "s", "s", "s", "s", "s"}},
	"nftables.delete_chain":   {pkg: "nftables", goName: "DeleteChain", args: true, params: []string{"s", "s", "s"}},
	"nftables.add_rule":       {pkg: "nftables", goName: "AddRule", args: true, params: []string{"s", "s", "s", "s"}},
	"nftables.delete_rule":    {pkg: "nftables", goName: "DeleteRule", args: true, params: []string{"s", "s", "s", "s"}},
	"nftables.flush_chain":    {pkg: "nftables", goName: "FlushChain", args: true, params: []string{"s", "s", "s"}},
	"nftables.flush_table":    {pkg: "nftables", goName: "FlushTable", args: true, params: []string{"s", "s"}},
	"nftables.flush_ruleset":  {pkg: "nftables", goName: "FlushRuleset", args: false},
	"nftables.list_ruleset":   {pkg: "nftables", goName: "ListRuleset", args: false},
	"nftables.add_set":        {pkg: "nftables", goName: "AddSet", args: true, params: []string{"s", "s", "s", "s", "s"}},
	"nftables.delete_set":     {pkg: "nftables", goName: "DeleteSet", args: true, params: []string{"s", "s", "s"}},
	"nftables.add_element":    {pkg: "nftables", goName: "AddElement", args: true, params: []string{"s", "s", "s", "s"}},
	"nftables.delete_element": {pkg: "nftables", goName: "DeleteElement", args: true, params: []string{"s", "s", "s", "s"}},
	"nftables.export":         {pkg: "nftables", goName: "Export", args: true, params: []string{"s"}},

	// mongodb
	"mongodb.create_database":     {pkg: "mongodb", goName: "CreateDatabase", args: true, params: []string{"s", "i", "s"}},
	"mongodb.drop_database":       {pkg: "mongodb", goName: "DropDatabase", args: true, params: []string{"s", "i", "s"}},
	"mongodb.list_databases":      {pkg: "mongodb", goName: "ListDatabases", args: true, params: []string{"s", "i"}},
	"mongodb.create_user":         {pkg: "mongodb", goName: "CreateUser", args: true, params: []string{"s", "i", "s", "s", "s", "s"}},
	"mongodb.drop_user":           {pkg: "mongodb", goName: "DropUser", args: true, params: []string{"s", "i", "s", "s"}},
	"mongodb.list_users":          {pkg: "mongodb", goName: "ListUsers", args: true, params: []string{"s", "i", "s"}},
	"mongodb.create_collection":   {pkg: "mongodb", goName: "CreateCollection", args: true, params: []string{"s", "i", "s", "s"}},
	"mongodb.drop_collection":     {pkg: "mongodb", goName: "DropCollection", args: true, params: []string{"s", "i", "s", "s"}},
	"mongodb.list_collections":    {pkg: "mongodb", goName: "ListCollections", args: true, params: []string{"s", "i", "s"}},
	"mongodb.create_index":        {pkg: "mongodb", goName: "CreateIndex", args: true, params: []string{"s", "i", "s", "s", "s", "b", "s"}},
	"mongodb.drop_index":          {pkg: "mongodb", goName: "DropIndex", args: true, params: []string{"s", "i", "s", "s", "s"}},
	"mongodb.list_indexes":        {pkg: "mongodb", goName: "ListIndexes", args: true, params: []string{"s", "i", "s", "s"}},
	"mongodb.server_status":       {pkg: "mongodb", goName: "ServerStatus", args: true, params: []string{"s", "i"}},
	"mongodb.replica_set_status":  {pkg: "mongodb", goName: "ReplicaSetStatus", args: true, params: []string{"s", "i"}},

	// tomcat
	"tomcat.start":     {pkg: "tomcat", goName: "Start", args: true, params: []string{"s"}},
	"tomcat.stop":      {pkg: "tomcat", goName: "Stop", args: true, params: []string{"s"}},
	"tomcat.restart":   {pkg: "tomcat", goName: "Restart", args: true, params: []string{"s"}},
	"tomcat.status":    {pkg: "tomcat", goName: "Status", args: true, params: []string{"s"}},
	"tomcat.deploy":    {pkg: "tomcat", goName: "Deploy", args: true, params: []string{"s", "s", "s"}},
	"tomcat.undeploy":  {pkg: "tomcat", goName: "Undeploy", args: true, params: []string{"s", "s"}},
	"tomcat.list_apps": {pkg: "tomcat", goName: "ListApps", args: true, params: []string{"s"}},
	"tomcat.reload":    {pkg: "tomcat", goName: "Reload", args: true, params: []string{"s", "s"}},
	"tomcat.version":   {pkg: "tomcat", goName: "Version", args: true, params: []string{"s"}},

	// java_cert
	"java_cert.import":          {pkg: "java_cert", goName: "Import", args: true, params: []string{"s", "s", "s", "s", "s"}},
	"java_cert.remove":          {pkg: "java_cert", goName: "Remove", args: true, params: []string{"s", "s", "s"}},
	"java_cert.list":            {pkg: "java_cert", goName: "List", args: true, params: []string{"s", "s"}},
	"java_cert.exists":          {pkg: "java_cert", goName: "Exists", args: true, params: []string{"s", "s", "s"}},
	"java_cert.export":          {pkg: "java_cert", goName: "Export", args: true, params: []string{"s", "s", "s", "s", "s"}},
	"java_cert.info":            {pkg: "java_cert", goName: "Info", args: true, params: []string{"s", "s"}},
	"java_cert.import_chain":    {pkg: "java_cert", goName: "ImportChain", args: true, params: []string{"s", "s", "s", "s"}},
	"java_cert.change_password": {pkg: "java_cert", goName: "ChangePassword", args: true, params: []string{"s", "s", "s"}},
	// maven_artifact
	"maven_artifact.download":          {pkg: "maven_artifact", goName: "Download", args: true, params: []string{"s", "s", "s", "s", "s", "s"}},
	"maven_artifact.resolve":           {pkg: "maven_artifact", goName: "Resolve", args: true, params: []string{"s", "s", "s", "s", "s"}},
	"maven_artifact.deploy":            {pkg: "maven_artifact", goName: "Deploy", args: true, params: []string{"s", "s", "s", "s", "s", "s"}},

	// ── ping ─────────────────────────────────────────────────────────────
	"ping.ping":     {pkg: "ping", goName: "Ping"},
	"ping.win_ping": {pkg: "ping", goName: "WinPing"},

	// ── find ─────────────────────────────────────────────────────────────
	"find.find": {pkg: "find", goName: "Find", args: true, params: []string{"l", "l", "s", "b", "i"}},

	// ── tempfile ─────────────────────────────────────────────────────────
	"tempfile.create_file": {pkg: "tempfile", goName: "CreateFile", args: true, params: []string{"s", "s", "s"}},
	"tempfile.create_dir":  {pkg: "tempfile", goName: "CreateDir", args: true, params: []string{"s", "s", "s"}},
	"tempfile.delete":      {pkg: "tempfile", goName: "Delete", args: true, params: []string{"s"}},

	// ── fail ─────────────────────────────────────────────────────────────
	"fail.fail": {pkg: "fail", goName: "Fail", args: true, params: []string{"s"}},

	// ── assert ───────────────────────────────────────────────────────────
	"assert.assert": {pkg: "assert", goName: "Assert", args: true, params: []string{"b", "s", "s"}},

	// ── debug ────────────────────────────────────────────────────────────
	"debug.debug":     {pkg: "debug", goName: "Debug", args: true, params: []string{"s"}},
	"debug.debug_var": {pkg: "debug", goName: "DebugVar", args: true, params: []string{"s", "s"}},

	// ── set_fact ─────────────────────────────────────────────────────────
	"set_fact.set":      {pkg: "set_fact", goName: "Set", args: true, params: []string{"s"}},
	"set_fact.get":      {pkg: "set_fact", goName: "Get", args: true, params: []string{"s"}},
	"set_fact.get_all":  {pkg: "set_fact", goName: "GetAll"},
	"set_fact.clear":    {pkg: "set_fact", goName: "Clear"},

	// ── unarchive ────────────────────────────────────────────────────────
	"unarchive.unarchive": {pkg: "unarchive", goName: "Unarchive", args: true, params: []string{"s", "s", "s", "s", "s", "s"}},

	// ── package_facts ────────────────────────────────────────────────────
	"package_facts.collect": {pkg: "package_facts", goName: "Collect", args: true, params: []string{"l"}},

	// ── service_facts ────────────────────────────────────────────────────
	"service_facts.collect": {pkg: "service_facts", goName: "Collect"},

	// ── command ──────────────────────────────────────────────────────────
	"command.run":   {pkg: "command", goName: "Run", args: true, params: []string{"l", "s", "s", "s", "i"}},
	"command.shell": {pkg: "command", goName: "Shell", args: true, params: []string{"l", "s", "s", "s", "i", "s"}},
	"maven_artifact.get_latest_version": {pkg: "maven_artifact", goName: "GetLatestVersion", args: true, params: []string{"s", "s", "s"}},
	"maven_artifact.checksum":          {pkg: "maven_artifact", goName: "Checksum", args: true, params: []string{"s"}},
	// docker_image
	"docker_image.pull":   {pkg: "docker_image", goName: "Pull", args: true, params: []string{"s", "s", "b"}},
	"docker_image.build":  {pkg: "docker_image", goName: "Build", args: true, params: []string{"s", "s", "s", "s"}},
	"docker_image.remove": {pkg: "docker_image", goName: "Remove", args: true, params: []string{"s", "s", "b"}},
	"docker_image.tag":    {pkg: "docker_image", goName: "Tag", args: true, params: []string{"s", "s"}},
	"docker_image.inspect": {pkg: "docker_image", goName: "Inspect", args: true, params: []string{"s"}},
	"docker_image.list":   {pkg: "docker_image", goName: "List", args: false, params: nil},
	"docker_image.push":   {pkg: "docker_image", goName: "Push", args: true, params: []string{"s", "s"}},
	// docker_container
	"docker_container.start":   {pkg: "docker_container", goName: "Start", args: true, params: []string{"s"}},
	"docker_container.stop":    {pkg: "docker_container", goName: "Stop", args: true, params: []string{"s", "i"}},
	"docker_container.remove":  {pkg: "docker_container", goName: "Remove", args: true, params: []string{"s", "b"}},
	"docker_container.restart": {pkg: "docker_container", goName: "Restart", args: true, params: []string{"s", "i"}},
	"docker_container.pause":   {pkg: "docker_container", goName: "Pause", args: true, params: []string{"s"}},
	"docker_container.unpause": {pkg: "docker_container", goName: "Unpause", args: true, params: []string{"s"}},
	"docker_container.inspect": {pkg: "docker_container", goName: "Inspect", args: true, params: []string{"s"}},
	"docker_container.list":    {pkg: "docker_container", goName: "List", args: true, params: []string{"b"}},
	"docker_container.logs":    {pkg: "docker_container", goName: "Logs", args: true, params: []string{"s", "s"}},

	// ── script ──────────────────────────────────────────────────────────
	"script.run": {pkg: "script", goName: "Run", args: true, params: []string{"s", "l", "s", "s", "s", "i", "s"}},

	// ── copy ────────────────────────────────────────────────────────────
	"copy.file":    {pkg: "copy", goName: "File", args: true, params: []string{"s", "s", "s", "s", "s", "b"}},
	"copy.content": {pkg: "copy", goName: "Content", args: true, params: []string{"s", "s", "s", "s", "s", "b"}},

	// ── cronvar ─────────────────────────────────────────────────────────
	"cronvar.present": {pkg: "cronvar", goName: "Present", args: true, params: []string{"s", "s", "s", "s", "s"}},
	"cronvar.absent":  {pkg: "cronvar", goName: "Absent", args: true, params: []string{"s", "s"}},
	"cronvar.get":     {pkg: "cronvar", goName: "Get", args: true, params: []string{"s", "s"}},

	// ── stat ────────────────────────────────────────────────────────────
	"stat.stat": {pkg: "stat", goName: "Stat", args: true, params: []string{"s", "b", "s"}},

	// ── add_host ────────────────────────────────────────────────────────
	"add_host.add":         {pkg: "add_host", goName: "Add", args: true, params: []string{"s", "l", "ms"}},
	"add_host.get_host":    {pkg: "add_host", goName: "GetHost", args: true, params: []string{"s"}},
	"add_host.get_group":   {pkg: "add_host", goName: "GetGroup", args: true, params: []string{"s"}},
	"add_host.list_hosts":  {pkg: "add_host", goName: "ListHosts"},
	"add_host.list_groups": {pkg: "add_host", goName: "ListGroups"},

	// ── set_stats ───────────────────────────────────────────────────────
	"set_stats.set":     {pkg: "set_stats", goName: "Set", args: true, params: []string{"ms"}},
	"set_stats.get":     {pkg: "set_stats", goName: "Get", args: true, params: []string{"s"}},
	"set_stats.get_all": {pkg: "set_stats", goName: "GetAll"},
	"set_stats.clear":   {pkg: "set_stats", goName: "Clear"},

	// ── include_vars ────────────────────────────────────────────────────
	"include_vars.load":     {pkg: "include_vars", goName: "Load", args: true, params: []string{"s"}},
	"include_vars.get":      {pkg: "include_vars", goName: "Get", args: true, params: []string{"s"}},
	"include_vars.get_all":  {pkg: "include_vars", goName: "GetAll"},

	// ── async_status ────────────────────────────────────────────────────
	"async_status.poll":     {pkg: "async_status", goName: "Poll", args: true, params: []string{"s", "s"}},
	"async_status.cleanup":  {pkg: "async_status", goName: "Cleanup", args: true, params: []string{"s", "s"}},
	"async_status.wait_for": {pkg: "async_status", goName: "WaitFor", args: true, params: []string{"s", "s", "i", "i"}},

	// ── package ─────────────────────────────────────────────────────────
	"package.install": {pkg: "package", goName: "Install", args: true, params: []string{"s"}},
	"package.remove":  {pkg: "package", goName: "Remove", args: true, params: []string{"s"}},
	"package.update":  {pkg: "package", goName: "Update", args: true, params: []string{"s"}},
	"package.info":    {pkg: "package", goName: "Info", args: true, params: []string{"s"}},

	// ── type_debug ──────────────────────────────────────────────────────
	"type_debug.debug": {pkg: "type_debug", goName: "Debug", args: true, params: []string{"a"}},

	// ── group_by ────────────────────────────────────────────────────────
	"group_by.group_by":   {pkg: "group_by", goName: "GroupBy", args: true, params: []string{"a", "s"}},
	"group_by.get_hosts":  {pkg: "group_by", goName: "GetHosts", args: true, params: []string{"s"}},
	"group_by.list_groups": {pkg: "group_by", goName: "ListGroups"},
	"group_by.clear":      {pkg: "group_by", goName: "Clear"},

	// ── normalize ───────────────────────────────────────────────────────
	"normalize.lower":      {pkg: "normalize", goName: "Lower", args: true, params: []string{"s"}},
	"normalize.upper":      {pkg: "normalize", goName: "Upper", args: true, params: []string{"s"}},
	"normalize.trim":       {pkg: "normalize", goName: "Trim", args: true, params: []string{"s"}},
	"normalize.slugify":    {pkg: "normalize", goName: "Slugify", args: true, params: []string{"s"}},
	"normalize.title":      {pkg: "normalize", goName: "Title", args: true, params: []string{"s"}},
	"normalize.camel_case": {pkg: "normalize", goName: "CamelCase", args: true, params: []string{"s"}},
	"normalize.snake_case": {pkg: "normalize", goName: "SnakeCase", args: true, params: []string{"s"}},
	"normalize.kebab_case": {pkg: "normalize", goName: "KebabCase", args: true, params: []string{"s"}},

	// ── validate_certs ──────────────────────────────────────────────────
	"validate_certs.validate":     {pkg: "validate_certs", goName: "Validate", args: true, params: []string{"s", "i", "d"}},
	"validate_certs.check_expiry": {pkg: "validate_certs", goName: "CheckExpiry", args: true, params: []string{"s", "i", "i", "d"}},

	// ── mail ────────────────────────────────────────────────────────────
	"mail.send":      {pkg: "mail", goName: "SendSimple", args: true, params: []string{"s", "i", "s", "as", "s", "s"}},
	"mail.send_html": {pkg: "mail", goName: "Send", args: true, params: []string{"s", "i", "s", "as", "s", "s"}},

	// ── webhook ─────────────────────────────────────────────────────────
	"webhook.send":    {pkg: "webhook", goName: "SendGeneric", args: true, params: []string{"s", "s", "mss", "a"}},
	"webhook.slack":   {pkg: "webhook", goName: "SendSlack", args: true, params: []string{"s", "s"}},
	"webhook.discord": {pkg: "webhook", goName: "SendDiscord", args: true, params: []string{"s", "s"}},
	"webhook.teams":   {pkg: "webhook", goName: "SendTeams", args: true, params: []string{"s", "s", "s"}},

	// ── openssl_privatekey ──────────────────────────────────────────────
	"openssl_privatekey.generate": {pkg: "openssl_privatekey", goName: "Generate", args: true, params: []string{"s", "s", "i"}},
	"openssl_privatekey.info":     {pkg: "openssl_privatekey", goName: "Info", args: true, params: []string{"s"}},
	"openssl_privatekey.delete":   {pkg: "openssl_privatekey", goName: "Delete", args: true, params: []string{"s"}},

	// ── ip_route ────────────────────────────────────────────────────────
	"ip_route.list":       {pkg: "ip_route", goName: "List"},
	"ip_route.list_table": {pkg: "ip_route", goName: "ListTable", args: true, params: []string{"s"}},
	"ip_route.add":        {pkg: "ip_route", goName: "Add", args: true, params: []string{"s", "s", "s", "i", "s"}},
	"ip_route.delete":     {pkg: "ip_route", goName: "Delete", args: true, params: []string{"s", "s"}},
	"ip_route.flush":      {pkg: "ip_route", goName: "Flush", args: true, params: []string{"s", "s"}},
	"ip_route.get":        {pkg: "ip_route", goName: "Get", args: true, params: []string{"s"}},
	// ip_link
	"ip_link.list":     {pkg: "ip_link", goName: "List"},
	"ip_link.get":      {pkg: "ip_link", goName: "Get", args: true, params: []string{"s"}},
	"ip_link.set_up":   {pkg: "ip_link", goName: "SetUp", args: true, params: []string{"s"}},
	"ip_link.set_down": {pkg: "ip_link", goName: "SetDown", args: true, params: []string{"s"}},
	"ip_link.set_mtu":  {pkg: "ip_link", goName: "SetMTU", args: true, params: []string{"s", "i"}},
	"ip_link.set_mac":  {pkg: "ip_link", goName: "SetMAC", args: true, params: []string{"s", "s"}},
	"ip_link.set_name": {pkg: "ip_link", goName: "SetName", args: true, params: []string{"s", "s"}},
	// ip_netns
	"ip_netns.list":   {pkg: "ip_netns", goName: "List"},
	"ip_netns.get":    {pkg: "ip_netns", goName: "Get", args: true, params: []string{"s"}},
	"ip_netns.add":    {pkg: "ip_netns", goName: "Add", args: true, params: []string{"s"}},
	"ip_netns.delete": {pkg: "ip_netns", goName: "Delete", args: true, params: []string{"s"}},
	"ip_netns.exec":   {pkg: "ip_netns", goName: "Exec", args: true, params: []string{"s", "s", "l"}},
	"ip_netns.pids":   {pkg: "ip_netns", goName: "Pids", args: true, params: []string{"s"}},
	// ip_neighbor
	"ip_neighbor.list":      {pkg: "ip_neighbor", goName: "List"},
	"ip_neighbor.list_dev":  {pkg: "ip_neighbor", goName: "ListDev", args: true, params: []string{"s"}},
	"ip_neighbor.add":       {pkg: "ip_neighbor", goName: "Add", args: true, params: []string{"s", "s", "s"}},
	"ip_neighbor.delete":    {pkg: "ip_neighbor", goName: "Delete", args: true, params: []string{"s", "s"}},
	"ip_neighbor.flush":     {pkg: "ip_neighbor", goName: "Flush", args: true, params: []string{"s"}},
	// openssl_csr
	"openssl_csr.generate": {pkg: "openssl_csr", goName: "Generate", args: true, params: []string{"s", "s", "s", "s", "s", "s", "s", "s", "s", "l", "b"}},
	"openssl_csr.info":     {pkg: "openssl_csr", goName: "Info", args: true, params: []string{"s"}},
	"openssl_csr.delete":   {pkg: "openssl_csr", goName: "Delete", args: true, params: []string{"s"}},
	// openssl_publickey
	"openssl_publickey.extract": {pkg: "openssl_publickey", goName: "Extract", args: true, params: []string{"s", "s", "b"}},
	"openssl_publickey.info":    {pkg: "openssl_publickey", goName: "Info", args: true, params: []string{"s"}},
	"openssl_publickey.delete":  {pkg: "openssl_publickey", goName: "Delete", args: true, params: []string{"s"}},
	// etcd
	"etcd.get":    {pkg: "etcd", goName: "Get", args: true, params: []string{"s", "l"}},
	"etcd.set":    {pkg: "etcd", goName: "Set", args: true, params: []string{"s", "s", "l"}},
	"etcd.delete": {pkg: "etcd", goName: "Delete", args: true, params: []string{"s", "l"}},
	"etcd.list":   {pkg: "etcd", goName: "List", args: true, params: []string{"s", "l"}},
	// zookeeper
	"zookeeper.get":    {pkg: "zookeeper", goName: "Get", args: true, params: []string{"s", "l"}},
	"zookeeper.set":    {pkg: "zookeeper", goName: "Set", args: true, params: []string{"s", "s", "l"}},
	"zookeeper.create": {pkg: "zookeeper", goName: "Create", args: true, params: []string{"s", "s", "b", "l"}},
	"zookeeper.delete": {pkg: "zookeeper", goName: "Delete", args: true, params: []string{"s", "l"}},
	"zookeeper.list":   {pkg: "zookeeper", goName: "List", args: true, params: []string{"s", "l"}},
	"zookeeper.exists": {pkg: "zookeeper", goName: "Exists", args: true, params: []string{"s", "l"}},
	// vault
	"vault.read":   {pkg: "vault", goName: "Read", args: true, params: []string{"s", "s", "s"}},
	"vault.write":  {pkg: "vault", goName: "Write", args: true, params: []string{"s", "s", "s", "m"}},
	"vault.delete": {pkg: "vault", goName: "Delete", args: true, params: []string{"s", "s", "s"}},
	"vault.list":   {pkg: "vault", goName: "List", args: true, params: []string{"s", "s", "s"}},
	// git_config
	"git_config.get":   {pkg: "git_config", goName: "Get", args: true, params: []string{"s", "s"}},
	"git_config.set":   {pkg: "git_config", goName: "Set", args: true, params: []string{"s", "s", "s"}},
	"git_config.unset": {pkg: "git_config", goName: "Unset", args: true, params: []string{"s", "s"}},
	"git_config.list":  {pkg: "git_config", goName: "List", args: true, params: []string{"s"}},

	// sshd_config
	"sshd_config.get":   {pkg: "sshd_config", goName: "Get", args: true, params: []string{"s"}},
	"sshd_config.set":   {pkg: "sshd_config", goName: "Set", args: true, params: []string{"s", "s"}},
	"sshd_config.absent": {pkg: "sshd_config", goName: "Absent", args: true, params: []string{"s"}},

	// docker_network
	"docker_network.inspect": {pkg: "docker_network", goName: "Inspect", args: true, params: []string{"s"}},
	"docker_network.create":  {pkg: "docker_network", goName: "Create", args: true, params: []string{"s", "s"}},
	"docker_network.remove":  {pkg: "docker_network", goName: "Remove", args: true, params: []string{"s"}},
	"docker_network.list":    {pkg: "docker_network", goName: "List", args: true},

	// docker_volume
	"docker_volume.inspect": {pkg: "docker_volume", goName: "Inspect", args: true, params: []string{"s"}},
	"docker_volume.create":  {pkg: "docker_volume", goName: "Create", args: true, params: []string{"s", "s"}},
	"docker_volume.remove":  {pkg: "docker_volume", goName: "Remove", args: true, params: []string{"s"}},
	"docker_volume.list":    {pkg: "docker_volume", goName: "List", args: true},

	// journald
	"journald.get": {pkg: "journald", goName: "Get", args: true, params: []string{"s"}},
	"journald.set": {pkg: "journald", goName: "Set", args: true, params: []string{"s", "s"}},

	// nfs_exports
	"nfs_exports.present": {pkg: "nfs_exports", goName: "Present", args: true, params: []string{"s", "s", "s"}},
	"nfs_exports.absent":  {pkg: "nfs_exports", goName: "Absent", args: true, params: []string{"s"}},
	"nfs_exports.list":    {pkg: "nfs_exports", goName: "List", args: true},
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
	"iptables":  "opsiptables",
	"npm":       "opsnpm",
	"mysql":     "opsmysql",
	"nginx":     "opsnginx",
	"modprobe":  "opsmodprobe",
	"alternatives": "opsalternatives",
	"blockdev":     "opsblockdev",
	"at":           "opsat",
	"postgresql":   "opspostgresql",
	"apache2":      "opsapache2",
	"filesystem":   "opsfilesystem",
	"parted":       "opsparted",
	"acl":          "opsacl",
	"wait_for":     "opswaitfor",
	"lvol":         "opslvol",
	"synchronize":  "opssync",
	"fetch":        "opsfetch",
	"seboolean":    "opssebool",
	"uri":          "opsuri",
	"lineinfile":   "opslineinfile",
	"replace":      "opsreplace",
	"xml":            "opsxml",
	"systemd":        "opssystemd",
	"patch":          "opspatch",
	"xattr":          "opsxattr",
	"firewalld_zone": "opsfirewalldzone",
	"get_url":        "opsgeturl",
	"seport":         "opsseport",
	"sefcontext":     "opssefcontext",
	"lvg":            "opslvg",
	"snap":           "opssnap",
	"flatpak":        "opsflatpak",
	"zfs":            "opszfs",
	"nmcli":          "opsnmcli",
	"crypttab":       "opscrypttab",
	"sysfs":          "opssysfs",
	"pamd":           "opspamd",
	"getent":         "opsgetent",
	"haproxy":        "opshaproxy",
	"openssl_cert":   "opsopenssl",
	"redis":          "opsredis",
	"gem":            "opsgem",
	"rabbitmq":       "opsrabbitmq",
	"consul":         "opsconsul",
	"memcached":      "opsmemcached",
	"composer":       "opscomposer",
	"cargo":          "opscargo",
	"rpmkey":         "opsrpmkey",
	"aptkey":         "opsaptkey",
	"dmidecode":      "opsdmidecode",
	"tuned":          "opstuned",
	"supervisor":     "opssupervisor",
	"smartctl":       "opssmartctl",
	"virsh":          "opsvirsh",
	"ethtool":        "opsethtool",
	"systemd_analyze": "opssystemd_analyze",
	"nvme":           "opsnvme",
	"lshw":           "opslshw",
	"ipaddr":         "opsipaddr",
	"udevadm":        "opsudevadm",
	"modinfo":        "opsmodinfo",
	"dconf":          "opsdconf",
	"locale_gen":     "opslocale_gen",
	"pam_limits":     "opspam_limits",
	"motd":           "opsmotd",
	"issue":          "opsissue",
	"authorized_key": "opsauthorized_key",
	"blockinfile":    "opsblockinfile",
	"debconf":        "opsdebconf",
	"reboot":         "opsreboot",
	"swap":           "opsswap",
	"raw":            "opsraw",
	"expect":         "opsexpect",
	"slurp":          "opsslurp",
	"wait_for_connection": "opswait_for_connection",
	"firewalld_rich_rule": "opsfirewalld_rich_rule",
	"firewalld_ipset": "opsfirewalld_ipset",
	"pause":          "opspause",
	"meta":           "opsmeta",
	"uri_ext":        "opsuri_ext",
	"hwclock":        "opshwclock",
	"mdadm":          "opsmdadm",
	"open_iscsi":     "opsopen_iscsi",
	"rfkill":         "opsrfkill",
	"multipath":      "opsmultipath",
	"dmsetup":        "opsdmsetup",
	"lvm_enhanced":   "opslvm_enhanced",
	"puppet":         "opspuppet",
	"yarn":           "opsyarn",
	"htpasswd":       "opshtpasswd",
	"sudoers":        "opssudoers",
	"monit":          "opsmonit",
	"apt":            "opsapt",
	"apt_repo":       "opsaptrepo",
	"apk":            "opsapk",
	"sysvinit":       "opssysvinit",
	"dpkg_selections": "opsdpkgsel",
	"homebrew":       "opsbrew",
	"dnf":            "opsdnf",
	"kubernetes":     "opsk8s",
	"svn":            "opssvn",
	"zypper":         "opszypper",
	"pacman":         "opspacman",
	"portage":        "opsportage",
	"pkgng":          "opspkgng",
	"podman":         "opspodman",
	"nftables":       "opsnftables",
	"mongodb":        "opsmongodb",
	"tomcat":         "opstomcat",
	"java_cert":      "opsjavacert",
	"maven_artifact": "opsmaven",
	"docker_image":   "opsdockerimage",
	"docker_container": "opsdockercontainer",
	"ping":           "opsping",
	"find":           "opsfind",
	"tempfile":       "opstempfile",
	"fail":           "opsfail",
	"assert":         "opsassert",
	"debug":          "opsdebug",
	"set_fact":       "opssetfact",
	"unarchive":      "opsunarchive",
	"package_facts":  "opspackagefacts",
	"service_facts":  "opsservicefacts",
	"command":        "opscommand",
	"script":         "opsscript",
	"copy":           "opscopy",
	"cronvar":        "opscronvar",
	"stat":           "opsstat",
	"add_host":       "opsaddhost",
	"set_stats":      "opssetstats",
	"include_vars":   "opsincludevars",
	"async_status":   "opsasyncstatus",
	"package":        "opspackage",
	"type_debug":     "opstypedebug",
	"group_by":       "opsgroupby",
	"normalize":      "opsnormalize",
	"validate_certs": "opsvalidatecerts",
	"mail":           "ops",
	"webhook":        "opswebhook",
	"openssl_privatekey": "opsopensslprivatekey",
	"ip_route":         "opsiproute",
	"ip_link":          "opsiplink",
	"ip_netns":         "opsipnetns",
	"ip_neighbor":      "opsipneighbor",
	"openssl_csr":      "opsopensslcsr",
	"openssl_publickey": "opsopensslpublickey",
	"etcd":             "opsetcd",
	"zookeeper":        "opszookeeper",
	"vault":            "opsvault",
	"git_config":       "opsgitconfig",
	"sshd_config":      "opssshdconfig",
	"docker_network":   "opsdockernet",
	"docker_volume":    "opsdockervol",
	"journald":         "opsjournald",
	"nfs_exports":      "opsnfsexports",
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
	"iptables":  "github.com/opslang/opslang/pkg/ops-core-sdk/iptables",
	"npm":       "github.com/opslang/opslang/pkg/ops-core-sdk/npm",
	"mysql":     "github.com/opslang/opslang/pkg/ops-core-sdk/mysql",
	"nginx":     "github.com/opslang/opslang/pkg/ops-core-sdk/nginx",
	"modprobe":  "github.com/opslang/opslang/pkg/ops-core-sdk/modprobe",
	"alternatives": "github.com/opslang/opslang/pkg/ops-core-sdk/alternatives",
	"blockdev":     "github.com/opslang/opslang/pkg/ops-core-sdk/blockdev",
	"at":           "github.com/opslang/opslang/pkg/ops-core-sdk/at",
	"postgresql":   "github.com/opslang/opslang/pkg/ops-core-sdk/postgresql",
	"apache2":      "github.com/opslang/opslang/pkg/ops-core-sdk/apache2",
	"filesystem":   "github.com/opslang/opslang/pkg/ops-core-sdk/filesystem",
	"parted":       "github.com/opslang/opslang/pkg/ops-core-sdk/parted",
	"acl":          "github.com/opslang/opslang/pkg/ops-core-sdk/acl",
	"wait_for":     "github.com/opslang/opslang/pkg/ops-core-sdk/wait_for",
	"lvol":         "github.com/opslang/opslang/pkg/ops-core-sdk/lvol",
	"synchronize":  "github.com/opslang/opslang/pkg/ops-core-sdk/synchronize",
	"fetch":        "github.com/opslang/opslang/pkg/ops-core-sdk/fetch",
	"seboolean":    "github.com/opslang/opslang/pkg/ops-core-sdk/seboolean",
	"uri":          "github.com/opslang/opslang/pkg/ops-core-sdk/uri",
	"lineinfile":   "github.com/opslang/opslang/pkg/ops-core-sdk/lineinfile",
	"replace":      "github.com/opslang/opslang/pkg/ops-core-sdk/replace",
	"xml":          "github.com/opslang/opslang/pkg/ops-core-sdk/xml",
	"systemd":      "github.com/opslang/opslang/pkg/ops-core-sdk/systemd",
	"patch":        "github.com/opslang/opslang/pkg/ops-core-sdk/patch",
	"xattr":        "github.com/opslang/opslang/pkg/ops-core-sdk/xattr",
	"firewalld_zone": "github.com/opslang/opslang/pkg/ops-core-sdk/firewalld_zone",
	"get_url":        "github.com/opslang/opslang/pkg/ops-core-sdk/get_url",
	"seport":         "github.com/opslang/opslang/pkg/ops-core-sdk/seport",
	"sefcontext":     "github.com/opslang/opslang/pkg/ops-core-sdk/sefcontext",
	"lvg":            "github.com/opslang/opslang/pkg/ops-core-sdk/lvg",
	"snap":           "github.com/opslang/opslang/pkg/ops-core-sdk/snap",
	"flatpak":        "github.com/opslang/opslang/pkg/ops-core-sdk/flatpak",
	"zfs":            "github.com/opslang/opslang/pkg/ops-core-sdk/zfs",
	"nmcli":          "github.com/opslang/opslang/pkg/ops-core-sdk/nmcli",
	"crypttab":       "github.com/opslang/opslang/pkg/ops-core-sdk/crypttab",
	"sysfs":          "github.com/opslang/opslang/pkg/ops-core-sdk/sysfs",
	"pamd":           "github.com/opslang/opslang/pkg/ops-core-sdk/pamd",
	"getent":         "github.com/opslang/opslang/pkg/ops-core-sdk/getent",
	"haproxy":        "github.com/opslang/opslang/pkg/ops-core-sdk/haproxy",
	"openssl_cert":   "github.com/opslang/opslang/pkg/ops-core-sdk/openssl_cert",
	"redis":          "github.com/opslang/opslang/pkg/ops-core-sdk/redis",
	"gem":            "github.com/opslang/opslang/pkg/ops-core-sdk/gem",
	"rabbitmq":       "github.com/opslang/opslang/pkg/ops-core-sdk/rabbitmq",
	"consul":         "github.com/opslang/opslang/pkg/ops-core-sdk/consul",
	"memcached":      "github.com/opslang/opslang/pkg/ops-core-sdk/memcached",
	"composer":       "github.com/opslang/opslang/pkg/ops-core-sdk/composer",
	"cargo":          "github.com/opslang/opslang/pkg/ops-core-sdk/cargo",
	"rpmkey":         "github.com/opslang/opslang/pkg/ops-core-sdk/rpmkey",
	"aptkey":         "github.com/opslang/opslang/pkg/ops-core-sdk/aptkey",
	"dmidecode":      "github.com/opslang/opslang/pkg/ops-core-sdk/dmidecode",
	"tuned":          "github.com/opslang/opslang/pkg/ops-core-sdk/tuned",
	"supervisor":     "github.com/opslang/opslang/pkg/ops-core-sdk/supervisor",
	"smartctl":       "github.com/opslang/opslang/pkg/ops-core-sdk/smartctl",
	"virsh":          "github.com/opslang/opslang/pkg/ops-core-sdk/virsh",
	"ethtool":        "github.com/opslang/opslang/pkg/ops-core-sdk/ethtool",
	"systemd_analyze": "github.com/opslang/opslang/pkg/ops-core-sdk/systemd_analyze",
	"nvme":           "github.com/opslang/opslang/pkg/ops-core-sdk/nvme",
	"lshw":           "github.com/opslang/opslang/pkg/ops-core-sdk/lshw",
	"ipaddr":         "github.com/opslang/opslang/pkg/ops-core-sdk/ipaddr",
	"udevadm":        "github.com/opslang/opslang/pkg/ops-core-sdk/udevadm",
	"modinfo":        "github.com/opslang/opslang/pkg/ops-core-sdk/modinfo",
	"dconf":          "github.com/opslang/opslang/pkg/ops-core-sdk/dconf",
	"locale_gen":     "github.com/opslang/opslang/pkg/ops-core-sdk/locale_gen",
	"pam_limits":     "github.com/opslang/opslang/pkg/ops-core-sdk/pam_limits",
	"motd":           "github.com/opslang/opslang/pkg/ops-core-sdk/motd",
	"issue":          "github.com/opslang/opslang/pkg/ops-core-sdk/issue",
	"authorized_key": "github.com/opslang/opslang/pkg/ops-core-sdk/authorized_key",
	"blockinfile":    "github.com/opslang/opslang/pkg/ops-core-sdk/blockinfile",
	"debconf":        "github.com/opslang/opslang/pkg/ops-core-sdk/debconf",
	"reboot":         "github.com/opslang/opslang/pkg/ops-core-sdk/reboot",
	"swap":           "github.com/opslang/opslang/pkg/ops-core-sdk/swap",
	"raw":            "github.com/opslang/opslang/pkg/ops-core-sdk/raw",
	"expect":         "github.com/opslang/opslang/pkg/ops-core-sdk/expect",
	"slurp":          "github.com/opslang/opslang/pkg/ops-core-sdk/slurp",
	"wait_for_connection": "github.com/opslang/opslang/pkg/ops-core-sdk/wait_for_connection",
	"firewalld_rich_rule": "github.com/opslang/opslang/pkg/ops-core-sdk/firewalld_rich_rule",
	"firewalld_ipset": "github.com/opslang/opslang/pkg/ops-core-sdk/firewalld_ipset",
	"pause":          "github.com/opslang/opslang/pkg/ops-core-sdk/pause",
	"meta":           "github.com/opslang/opslang/pkg/ops-core-sdk/meta",
	"uri_ext":        "github.com/opslang/opslang/pkg/ops-core-sdk/uri_ext",
	"hwclock":        "github.com/opslang/opslang/pkg/ops-core-sdk/hwclock",
	"mdadm":          "github.com/opslang/opslang/pkg/ops-core-sdk/mdadm",
	"open_iscsi":     "github.com/opslang/opslang/pkg/ops-core-sdk/open_iscsi",
	"rfkill":         "github.com/opslang/opslang/pkg/ops-core-sdk/rfkill",
	"multipath":      "github.com/opslang/opslang/pkg/ops-core-sdk/multipath",
	"dmsetup":        "github.com/opslang/opslang/pkg/ops-core-sdk/dmsetup",
	"lvm_enhanced":   "github.com/opslang/opslang/pkg/ops-core-sdk/lvm_enhanced",
	"puppet":         "github.com/opslang/opslang/pkg/ops-core-sdk/puppet",
	"yarn":           "github.com/opslang/opslang/pkg/ops-core-sdk/yarn",
	"htpasswd":       "github.com/opslang/opslang/pkg/ops-core-sdk/htpasswd",
	"sudoers":        "github.com/opslang/opslang/pkg/ops-core-sdk/sudoers",
	"monit":          "github.com/opslang/opslang/pkg/ops-core-sdk/monit",
	"apt":            "github.com/opslang/opslang/pkg/ops-core-sdk/apt",
	"apt_repo":       "github.com/opslang/opslang/pkg/ops-core-sdk/apt_repo",
	"apk":            "github.com/opslang/opslang/pkg/ops-core-sdk/apk",
	"sysvinit":       "github.com/opslang/opslang/pkg/ops-core-sdk/sysvinit",
	"dpkg_selections": "github.com/opslang/opslang/pkg/ops-core-sdk/dpkg_selections",
	"homebrew":       "github.com/opslang/opslang/pkg/ops-core-sdk/homebrew",
	"dnf":            "github.com/opslang/opslang/pkg/ops-core-sdk/dnf",
	"kubernetes":     "github.com/opslang/opslang/pkg/ops-core-sdk/kubernetes",
	"svn":            "github.com/opslang/opslang/pkg/ops-core-sdk/svn",
	"zypper":         "github.com/opslang/opslang/pkg/ops-core-sdk/zypper",
	"pacman":         "github.com/opslang/opslang/pkg/ops-core-sdk/pacman",
	"portage":        "github.com/opslang/opslang/pkg/ops-core-sdk/portage",
	"pkgng":          "github.com/opslang/opslang/pkg/ops-core-sdk/pkgng",
	"podman":         "github.com/opslang/opslang/pkg/ops-core-sdk/podman",
	"nftables":       "github.com/opslang/opslang/pkg/ops-core-sdk/nftables",
	"mongodb":        "github.com/opslang/opslang/pkg/ops-core-sdk/mongodb",
	"tomcat":         "github.com/opslang/opslang/pkg/ops-core-sdk/tomcat",
	"java_cert":      "github.com/opslang/opslang/pkg/ops-core-sdk/java_cert",
	"maven_artifact": "github.com/opslang/opslang/pkg/ops-core-sdk/maven_artifact",
	"docker_image":   "github.com/opslang/opslang/pkg/ops-core-sdk/docker_image",
	"docker_container": "github.com/opslang/opslang/pkg/ops-core-sdk/docker_container",
	"ping":           "github.com/opslang/opslang/pkg/ops-core-sdk/ping",
	"find":           "github.com/opslang/opslang/pkg/ops-core-sdk/find",
	"tempfile":       "github.com/opslang/opslang/pkg/ops-core-sdk/tempfile",
	"fail":           "github.com/opslang/opslang/pkg/ops-core-sdk/fail",
	"assert":         "github.com/opslang/opslang/pkg/ops-core-sdk/assert",
	"debug":          "github.com/opslang/opslang/pkg/ops-core-sdk/debug",
	"set_fact":       "github.com/opslang/opslang/pkg/ops-core-sdk/set_fact",
	"unarchive":      "github.com/opslang/opslang/pkg/ops-core-sdk/unarchive",
	"package_facts":  "github.com/opslang/opslang/pkg/ops-core-sdk/package_facts",
	"service_facts":  "github.com/opslang/opslang/pkg/ops-core-sdk/service_facts",
	"command":        "github.com/opslang/opslang/pkg/ops-core-sdk/command",
	"script":         "github.com/opslang/opslang/pkg/ops-core-sdk/script",
	"copy":           "github.com/opslang/opslang/pkg/ops-core-sdk/copy",
	"cronvar":        "github.com/opslang/opslang/pkg/ops-core-sdk/cronvar",
	"stat":           "github.com/opslang/opslang/pkg/ops-core-sdk/stat",
	"add_host":       "github.com/opslang/opslang/pkg/ops-core-sdk/add_host",
	"set_stats":      "github.com/opslang/opslang/pkg/ops-core-sdk/set_stats",
	"include_vars":   "github.com/opslang/opslang/pkg/ops-core-sdk/include_vars",
	"async_status":   "github.com/opslang/opslang/pkg/ops-core-sdk/async_status",
	"package":        "github.com/opslang/opslang/pkg/ops-core-sdk/package",
	"type_debug":     "github.com/opslang/opslang/pkg/ops-core-sdk/type_debug",
	"group_by":       "github.com/opslang/opslang/pkg/ops-core-sdk/group_by",
	"normalize":      "github.com/opslang/opslang/pkg/ops-core-sdk/normalize",
	"validate_certs": "github.com/opslang/opslang/pkg/ops-core-sdk/validate_certs",
	"mail":           "github.com/opslang/opslang/pkg/ops-core-sdk/mail",
	"webhook":        "github.com/opslang/opslang/pkg/ops-core-sdk/webhook",
	"openssl_privatekey": "github.com/opslang/opslang/pkg/ops-core-sdk/openssl_privatekey",
	"ip_route":         "github.com/opslang/opslang/pkg/ops-core-sdk/ip_route",
	"ip_link":          "github.com/opslang/opslang/pkg/ops-core-sdk/ip_link",
	"ip_netns":         "github.com/opslang/opslang/pkg/ops-core-sdk/ip_netns",
	"ip_neighbor":      "github.com/opslang/opslang/pkg/ops-core-sdk/ip_neighbor",
	"openssl_csr":      "github.com/opslang/opslang/pkg/ops-core-sdk/openssl_csr",
	"openssl_publickey": "github.com/opslang/opslang/pkg/ops-core-sdk/openssl_publickey",
	"etcd":             "github.com/opslang/opslang/pkg/ops-core-sdk/etcd",
	"zookeeper":        "github.com/opslang/opslang/pkg/ops-core-sdk/zookeeper",
	"vault":            "github.com/opslang/opslang/pkg/ops-core-sdk/vault",
	"git_config":       "github.com/opslang/opslang/pkg/ops-core-sdk/git_config",
	"sshd_config":      "github.com/opslang/opslang/pkg/ops-core-sdk/sshd_config",
	"docker_network":   "github.com/opslang/opslang/pkg/ops-core-sdk/docker_network",
	"docker_volume":    "github.com/opslang/opslang/pkg/ops-core-sdk/docker_volume",
	"journald":         "github.com/opslang/opslang/pkg/ops-core-sdk/journald",
	"nfs_exports":      "github.com/opslang/opslang/pkg/ops-core-sdk/nfs_exports",
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
