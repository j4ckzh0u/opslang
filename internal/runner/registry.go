package runner

import (
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/opslang/opslang/internal/ast"
	"github.com/opslang/opslang/internal/opsspec"
	sdkapt "github.com/opslang/opslang/pkg/ops-core-sdk/apt"
	sdkaptrepo "github.com/opslang/opslang/pkg/ops-core-sdk/apt_repo"
	sdkapk "github.com/opslang/opslang/pkg/ops-core-sdk/apk"
	sdksysvinit "github.com/opslang/opslang/pkg/ops-core-sdk/sysvinit"
	sdkdpkgsel "github.com/opslang/opslang/pkg/ops-core-sdk/dpkg_selections"
	sdkbrew "github.com/opslang/opslang/pkg/ops-core-sdk/homebrew"
	sdkarchive "github.com/opslang/opslang/pkg/ops-core-sdk/archive"
	opscron "github.com/opslang/opslang/pkg/ops-core-sdk/cron"
	sdkdisk "github.com/opslang/opslang/pkg/ops-core-sdk/disk"
	sdkdnf "github.com/opslang/opslang/pkg/ops-core-sdk/dnf"
	sdkdocker "github.com/opslang/opslang/pkg/ops-core-sdk/docker"
	"github.com/opslang/opslang/pkg/ops-core-sdk/file"
	sdkfirewalld "github.com/opslang/opslang/pkg/ops-core-sdk/firewalld"
	opsgit "github.com/opslang/opslang/pkg/ops-core-sdk/git"
	opsgrp "github.com/opslang/opslang/pkg/ops-core-sdk/group"
	opshosts "github.com/opslang/opslang/pkg/ops-core-sdk/hosts"
	opsjson "github.com/opslang/opslang/pkg/ops-core-sdk/json"
	sdkkernel "github.com/opslang/opslang/pkg/ops-core-sdk/kernel"
	sdkknownhosts "github.com/opslang/opslang/pkg/ops-core-sdk/known_hosts"
	sdklimits "github.com/opslang/opslang/pkg/ops-core-sdk/limits"
	sdklocale "github.com/opslang/opslang/pkg/ops-core-sdk/locale"
	sdklogrotate "github.com/opslang/opslang/pkg/ops-core-sdk/logrotate"
	sdklvg "github.com/opslang/opslang/pkg/ops-core-sdk/lvg"
	opsnet "github.com/opslang/opslang/pkg/ops-core-sdk/net"
	sdkntp "github.com/opslang/opslang/pkg/ops-core-sdk/ntp"
	sdkpip "github.com/opslang/opslang/pkg/ops-core-sdk/pip"
	opspkg "github.com/opslang/opslang/pkg/ops-core-sdk/pkg"
	"github.com/opslang/opslang/pkg/ops-core-sdk/process"
	sdkresolv "github.com/opslang/opslang/pkg/ops-core-sdk/resolv"
	sdksnap "github.com/opslang/opslang/pkg/ops-core-sdk/snap"
	sdkselinux "github.com/opslang/opslang/pkg/ops-core-sdk/selinux"
	"github.com/opslang/opslang/pkg/ops-core-sdk/service"
	sdkssh "github.com/opslang/opslang/pkg/ops-core-sdk/ssh"
	"github.com/opslang/opslang/pkg/ops-core-sdk/sys"
	sdksysctl "github.com/opslang/opslang/pkg/ops-core-sdk/sysctl"
	optime "github.com/opslang/opslang/pkg/ops-core-sdk/time"
	opsuser "github.com/opslang/opslang/pkg/ops-core-sdk/user"
	opsyaml "github.com/opslang/opslang/pkg/ops-core-sdk/yaml"
	sdkyumrepo "github.com/opslang/opslang/pkg/ops-core-sdk/yum_repo"
	sdkufw "github.com/opslang/opslang/pkg/ops-core-sdk/ufw"
	sdkinifile "github.com/opslang/opslang/pkg/ops-core-sdk/ini_file"
	sdkmount "github.com/opslang/opslang/pkg/ops-core-sdk/mount"
	sdkhostname "github.com/opslang/opslang/pkg/ops-core-sdk/hostname"
	sdkiptables "github.com/opslang/opslang/pkg/ops-core-sdk/iptables"
	sdknpm "github.com/opslang/opslang/pkg/ops-core-sdk/npm"
	sdkmysql "github.com/opslang/opslang/pkg/ops-core-sdk/mysql"
	sdknginx "github.com/opslang/opslang/pkg/ops-core-sdk/nginx"
	sdkmodprobe "github.com/opslang/opslang/pkg/ops-core-sdk/modprobe"
	sdkalternatives "github.com/opslang/opslang/pkg/ops-core-sdk/alternatives"
	sdkblockdev "github.com/opslang/opslang/pkg/ops-core-sdk/blockdev"
	sdkat "github.com/opslang/opslang/pkg/ops-core-sdk/at"
	sdkpostgresql "github.com/opslang/opslang/pkg/ops-core-sdk/postgresql"
	sdkapache2 "github.com/opslang/opslang/pkg/ops-core-sdk/apache2"
	sdkfilesystem "github.com/opslang/opslang/pkg/ops-core-sdk/filesystem"
	sdkparted "github.com/opslang/opslang/pkg/ops-core-sdk/parted"
	sdkacl "github.com/opslang/opslang/pkg/ops-core-sdk/acl"
	sdkwaitfor "github.com/opslang/opslang/pkg/ops-core-sdk/wait_for"
	sdklvol "github.com/opslang/opslang/pkg/ops-core-sdk/lvol"
	sdksync "github.com/opslang/opslang/pkg/ops-core-sdk/synchronize"
	sdkfetch "github.com/opslang/opslang/pkg/ops-core-sdk/fetch"
	sdksebool "github.com/opslang/opslang/pkg/ops-core-sdk/seboolean"
	sdktimezone "github.com/opslang/opslang/pkg/ops-core-sdk/timezone"
	sdkuri "github.com/opslang/opslang/pkg/ops-core-sdk/uri"
	sdklineinfile "github.com/opslang/opslang/pkg/ops-core-sdk/lineinfile"
	sdkreplace "github.com/opslang/opslang/pkg/ops-core-sdk/replace"
	sdkxml "github.com/opslang/opslang/pkg/ops-core-sdk/xml"
	sdksystemd "github.com/opslang/opslang/pkg/ops-core-sdk/systemd"
	sdkpatch "github.com/opslang/opslang/pkg/ops-core-sdk/patch"
	sdkxattr "github.com/opslang/opslang/pkg/ops-core-sdk/xattr"
	sdkfirewalldzone "github.com/opslang/opslang/pkg/ops-core-sdk/firewalld_zone"
	sdkgeturl "github.com/opslang/opslang/pkg/ops-core-sdk/get_url"
	sdkseport "github.com/opslang/opslang/pkg/ops-core-sdk/seport"
	sdksefcontext "github.com/opslang/opslang/pkg/ops-core-sdk/sefcontext"
	sdkflatpak "github.com/opslang/opslang/pkg/ops-core-sdk/flatpak"
	sdkzfs "github.com/opslang/opslang/pkg/ops-core-sdk/zfs"
	sdknmcli "github.com/opslang/opslang/pkg/ops-core-sdk/nmcli"
	sdkcrypttab "github.com/opslang/opslang/pkg/ops-core-sdk/crypttab"
	sdksysfs "github.com/opslang/opslang/pkg/ops-core-sdk/sysfs"
	sdkpamd "github.com/opslang/opslang/pkg/ops-core-sdk/pamd"
	sdkgetent "github.com/opslang/opslang/pkg/ops-core-sdk/getent"
	sdkhaproxy "github.com/opslang/opslang/pkg/ops-core-sdk/haproxy"
	sdkopenssl "github.com/opslang/opslang/pkg/ops-core-sdk/openssl_cert"
	sdkredis "github.com/opslang/opslang/pkg/ops-core-sdk/redis"
	sdkgem "github.com/opslang/opslang/pkg/ops-core-sdk/gem"
	sdkrabbitmq "github.com/opslang/opslang/pkg/ops-core-sdk/rabbitmq"
	sdkconsul "github.com/opslang/opslang/pkg/ops-core-sdk/consul"
	sdkmemcached "github.com/opslang/opslang/pkg/ops-core-sdk/memcached"
	sdkcomposer "github.com/opslang/opslang/pkg/ops-core-sdk/composer"
	sdkcargo "github.com/opslang/opslang/pkg/ops-core-sdk/cargo"
	sdkrpmkey "github.com/opslang/opslang/pkg/ops-core-sdk/rpmkey"
	sdkaptkey "github.com/opslang/opslang/pkg/ops-core-sdk/aptkey"
	sdkdmidecode "github.com/opslang/opslang/pkg/ops-core-sdk/dmidecode"
	sdktuned "github.com/opslang/opslang/pkg/ops-core-sdk/tuned"
	sdksupervisor "github.com/opslang/opslang/pkg/ops-core-sdk/supervisor"
	sdksmartctl "github.com/opslang/opslang/pkg/ops-core-sdk/smartctl"
	sdkvirsh "github.com/opslang/opslang/pkg/ops-core-sdk/virsh"
	sdkethtool "github.com/opslang/opslang/pkg/ops-core-sdk/ethtool"
	sdksystemd_analyze "github.com/opslang/opslang/pkg/ops-core-sdk/systemd_analyze"
	sdknvme "github.com/opslang/opslang/pkg/ops-core-sdk/nvme"
	sdkslshw "github.com/opslang/opslang/pkg/ops-core-sdk/lshw"
	sdkipaddr "github.com/opslang/opslang/pkg/ops-core-sdk/ipaddr"
	sdkudevadm "github.com/opslang/opslang/pkg/ops-core-sdk/udevadm"
	sdkmodinfo "github.com/opslang/opslang/pkg/ops-core-sdk/modinfo"
	sdkdconf "github.com/opslang/opslang/pkg/ops-core-sdk/dconf"
	sdklocale_gen "github.com/opslang/opslang/pkg/ops-core-sdk/locale_gen"
	sdkpam_limits "github.com/opslang/opslang/pkg/ops-core-sdk/pam_limits"
	sdkmotd "github.com/opslang/opslang/pkg/ops-core-sdk/motd"
	sdkissue "github.com/opslang/opslang/pkg/ops-core-sdk/issue"
	sdkauthorized_key "github.com/opslang/opslang/pkg/ops-core-sdk/authorized_key"
	sdkblockinfile "github.com/opslang/opslang/pkg/ops-core-sdk/blockinfile"
	sdkdebconf "github.com/opslang/opslang/pkg/ops-core-sdk/debconf"
	sdkreboot "github.com/opslang/opslang/pkg/ops-core-sdk/reboot"
	sdkswap "github.com/opslang/opslang/pkg/ops-core-sdk/swap"
	sdkraw "github.com/opslang/opslang/pkg/ops-core-sdk/raw"
	sdkexpect "github.com/opslang/opslang/pkg/ops-core-sdk/expect"
	sdkslurp "github.com/opslang/opslang/pkg/ops-core-sdk/slurp"
	sdkwait_for_connection "github.com/opslang/opslang/pkg/ops-core-sdk/wait_for_connection"
	sdkfirewalld_rich_rule "github.com/opslang/opslang/pkg/ops-core-sdk/firewalld_rich_rule"
	sdkfirewalld_ipset "github.com/opslang/opslang/pkg/ops-core-sdk/firewalld_ipset"
	sdkpause "github.com/opslang/opslang/pkg/ops-core-sdk/pause"
	sdkmeta "github.com/opslang/opslang/pkg/ops-core-sdk/meta"
	sdkuri_ext "github.com/opslang/opslang/pkg/ops-core-sdk/uri_ext"
	sdkhwclock "github.com/opslang/opslang/pkg/ops-core-sdk/hwclock"
	sdkmdadm "github.com/opslang/opslang/pkg/ops-core-sdk/mdadm"
	sdkopen_iscsi "github.com/opslang/opslang/pkg/ops-core-sdk/open_iscsi"
	sdkrfkill "github.com/opslang/opslang/pkg/ops-core-sdk/rfkill"
	sdkmultipath "github.com/opslang/opslang/pkg/ops-core-sdk/multipath"
	sdkdmsetup "github.com/opslang/opslang/pkg/ops-core-sdk/dmsetup"
	sdklvm_enhanced "github.com/opslang/opslang/pkg/ops-core-sdk/lvm_enhanced"
	sdkpuppet "github.com/opslang/opslang/pkg/ops-core-sdk/puppet"
	sdkyarn "github.com/opslang/opslang/pkg/ops-core-sdk/yarn"
	sdkhtpasswd "github.com/opslang/opslang/pkg/ops-core-sdk/htpasswd"
	sdksudoers "github.com/opslang/opslang/pkg/ops-core-sdk/sudoers"
	sdkmonit "github.com/opslang/opslang/pkg/ops-core-sdk/monit"
	sdkk8s "github.com/opslang/opslang/pkg/ops-core-sdk/kubernetes"
	sdksvn "github.com/opslang/opslang/pkg/ops-core-sdk/svn"
	sdkzypper "github.com/opslang/opslang/pkg/ops-core-sdk/zypper"
	sdkpacman "github.com/opslang/opslang/pkg/ops-core-sdk/pacman"
	sdkportage "github.com/opslang/opslang/pkg/ops-core-sdk/portage"
	sdkpkgng "github.com/opslang/opslang/pkg/ops-core-sdk/pkgng"
	sdkpodman "github.com/opslang/opslang/pkg/ops-core-sdk/podman"
	sdknftables "github.com/opslang/opslang/pkg/ops-core-sdk/nftables"
)

// Registry holds all registered operations and provides lookup and execution.
// Operation names follow the canonical table in internal/opsspec; historical
// aliases are resolved transparently at lookup time.
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

// Register adds an operation to the registry.
func (r *Registry) Register(name string, fn OperationFunc) {
	r.ops[name] = fn
}

// Get retrieves an operation from the registry, resolving canonical aliases.
func (r *Registry) Get(name string) (OperationFunc, bool) {
	if fn, ok := r.ops[name]; ok {
		return fn, true
	}
	if canonical, ok := opsspec.Aliases[name]; ok {
		fn, ok := r.ops[canonical]
		return fn, ok
	}
	return nil, false
}

// Has reports whether an operation (or its alias) is registered.
func (r *Registry) Has(name string) bool {
	_, ok := r.Get(name)
	return ok
}

// ListOperations returns the names of all registered operations.
func (r *Registry) ListOperations() []string {
	names := make([]string, 0, len(r.ops))
	for name := range r.ops {
		names = append(names, name)
	}
	return names
}

// registerAll registers all standard library operations grouped by SDK package.
func (r *Registry) registerAll() {
	r.registerSysOps()
	r.registerFileOps()
	r.registerNetOps()
	r.registerProcessOps()
	r.registerServiceOps()
	r.registerPkgOps()
	r.registerTimeOps()
	r.registerJSONOps()
	r.registerYAMLOps()
	r.registerGitOps()
	r.registerBuiltinOps()
	r.registerPlatformOps()
	r.registerSelinuxOps()
	r.registerExtensions()
}

// ============================================================
// sys operations
// ============================================================

func (r *Registry) registerSysOps() {
	r.Register("sys.cpu.usage", func(_ map[string]interface{}) (interface{}, error) {
		return sys.GetCPUUsage()
	})
	r.Register("sys.cpu.count", func(_ map[string]interface{}) (interface{}, error) {
		return sys.GetCPUCount()
	})
	r.Register("sys.cpu.info", func(_ map[string]interface{}) (interface{}, error) {
		return sys.GetCPUInfo()
	})
	r.Register("sys.memory.info", func(_ map[string]interface{}) (interface{}, error) {
		return sys.GetMemoryInfo()
	})
	r.Register("sys.disk.usage", func(args map[string]interface{}) (interface{}, error) {
		path, err := argString(args, "path")
		if err != nil {
			return nil, fmt.Errorf("sys.disk.usage: %w", err)
		}
		return sys.GetDiskUsage(path)
	})
	r.Register("sys.disk.partitions", func(_ map[string]interface{}) (interface{}, error) {
		return sys.GetDiskPartitions()
	})
	r.Register("sys.os", func(_ map[string]interface{}) (interface{}, error) {
		return sys.GetHostInfo()
	})
	r.Register("sys.load", func(_ map[string]interface{}) (interface{}, error) {
		return sys.GetLoadAvg()
	})
	r.Register("sys.net.interfaces", func(_ map[string]interface{}) (interface{}, error) {
		return sys.GetNetInterfaces()
	})
	r.Register("sys.users", func(_ map[string]interface{}) (interface{}, error) {
		return sys.Users()
	})
	r.Register("sys.uptime", func(_ map[string]interface{}) (interface{}, error) {
		return sys.Uptime()
	})
	r.Register("sys.hostname", func(_ map[string]interface{}) (interface{}, error) {
		return sys.Hostname()
	})
}

// ============================================================
// file operations
// ============================================================

func (r *Registry) registerFileOps() {
	r.Register("file.read", func(args map[string]interface{}) (interface{}, error) {
		path, err := argString(args, "path")
		if err != nil {
			return nil, fmt.Errorf("file.read: %w", err)
		}
		return file.Read(path)
	})
	r.Register("file.write", func(args map[string]interface{}) (interface{}, error) {
		path, err := argString(args, "path")
		if err != nil {
			return nil, fmt.Errorf("file.write: %w", err)
		}
		content, err := argString(args, "content")
		if err != nil {
			return nil, fmt.Errorf("file.write: %w", err)
		}
		return file.Write(path, content)
	})
	r.Register("file.append", func(args map[string]interface{}) (interface{}, error) {
		path, err := argString(args, "path")
		if err != nil {
			return nil, fmt.Errorf("file.append: %w", err)
		}
		content, err := argString(args, "content")
		if err != nil {
			return nil, fmt.Errorf("file.append: %w", err)
		}
		return file.Append(path, content)
	})
	r.Register("file.exists", func(args map[string]interface{}) (interface{}, error) {
		path, err := argString(args, "path")
		if err != nil {
			return nil, fmt.Errorf("file.exists: %w", err)
		}
		return file.Exists(path)
	})
	r.Register("file.copy", func(args map[string]interface{}) (interface{}, error) {
		src, err := argString(args, "src")
		if err != nil {
			return nil, fmt.Errorf("file.copy: %w", err)
		}
		dst, err := argString(args, "dst")
		if err != nil {
			return nil, fmt.Errorf("file.copy: %w", err)
		}
		return file.Copy(src, dst)
	})
	r.Register("file.move", func(args map[string]interface{}) (interface{}, error) {
		src, err := argString(args, "src")
		if err != nil {
			return nil, fmt.Errorf("file.move: %w", err)
		}
		dst, err := argString(args, "dst")
		if err != nil {
			return nil, fmt.Errorf("file.move: %w", err)
		}
		return file.Move(src, dst)
	})
	r.Register("file.delete", func(args map[string]interface{}) (interface{}, error) {
		path, err := argString(args, "path")
		if err != nil {
			return nil, fmt.Errorf("file.delete: %w", err)
		}
		return file.Delete(path)
	})
	r.Register("file.stat", func(args map[string]interface{}) (interface{}, error) {
		path, err := argString(args, "path")
		if err != nil {
			return nil, fmt.Errorf("file.stat: %w", err)
		}
		return file.Stat(path)
	})
	r.Register("file.list", func(args map[string]interface{}) (interface{}, error) {
		dir, err := argString(args, "dir")
		if err != nil {
			return nil, fmt.Errorf("file.list: %w", err)
		}
		return file.List(dir)
	})
	r.Register("file.mkdir", func(args map[string]interface{}) (interface{}, error) {
		path, err := argString(args, "path")
		if err != nil {
			return nil, fmt.Errorf("file.mkdir: %w", err)
		}
		return file.Mkdir(path)
	})
	r.Register("file.chmod", func(args map[string]interface{}) (interface{}, error) {
		path, err := argString(args, "path")
		if err != nil {
			return nil, fmt.Errorf("file.chmod: %w", err)
		}
		modeStr, err := argString(args, "mode")
		if err != nil {
			return nil, fmt.Errorf("file.chmod: %w", err)
		}
		var mode uint64
		if _, err := fmt.Sscanf(modeStr, "%o", &mode); err != nil {
			return nil, fmt.Errorf("file.chmod: mode must be an octal string like \"0755\", got %q", modeStr)
		}
		return file.Chmod(path, uint32(mode))
	})
	r.Register("file.template", func(args map[string]interface{}) (interface{}, error) {
		path, err := argString(args, "path")
		if err != nil {
			return nil, fmt.Errorf("file.template: %w", err)
		}
		vars, _ := args["vars"].(map[string]interface{})
		return file.Template(path, vars)
	})
	r.Register("file.checksum", func(args map[string]interface{}) (interface{}, error) {
		path, err := argString(args, "path")
		if err != nil {
			return nil, fmt.Errorf("file.checksum: %w", err)
		}
		algo := getStringArg(args, "algo", "sha256")
		return file.Checksum(path, algo)
	})
	r.Register("file.lineinfile", func(args map[string]interface{}) (interface{}, error) {
		path, err := argString(args, "path")
		if err != nil {
			return nil, fmt.Errorf("file.lineinfile: %w", err)
		}
		line, err := argString(args, "line")
		if err != nil {
			return nil, fmt.Errorf("file.lineinfile: %w", err)
		}
		present, err := argBool(args, "present")
		if err != nil {
			return nil, fmt.Errorf("file.lineinfile: %w", err)
		}
		rx := getStringArg(args, "regexp", "")
		return file.LineInFile(path, line, present, rx)
	})
}

// ============================================================
// net operations
// ============================================================

func (r *Registry) registerNetOps() {
	r.Register("net.http_get", func(args map[string]interface{}) (interface{}, error) {
		url, err := argString(args, "url")
		if err != nil {
			return nil, fmt.Errorf("net.http_get: %w", err)
		}
		return opsnet.HTTPGet(url)
	})
	r.Register("net.http_post", func(args map[string]interface{}) (interface{}, error) {
		url, err := argString(args, "url")
		if err != nil {
			return nil, fmt.Errorf("net.http_post: %w", err)
		}
		body, err := argString(args, "body")
		if err != nil {
			return nil, fmt.Errorf("net.http_post: %w", err)
		}
		return opsnet.HTTPPost(url, body)
	})
	r.Register("net.tcp_check", func(args map[string]interface{}) (interface{}, error) {
		host, err := argString(args, "host")
		if err != nil {
			return nil, fmt.Errorf("net.tcp_check: %w", err)
		}
		port, err := argInt(args, "port")
		if err != nil {
			return nil, fmt.Errorf("net.tcp_check: %w", err)
		}
		return opsnet.TCPConnect(host, port)
	})
	r.Register("net.dns_lookup", func(args map[string]interface{}) (interface{}, error) {
		host, err := argString(args, "host")
		if err != nil {
			return nil, fmt.Errorf("net.dns_lookup: %w", err)
		}
		return opsnet.DNSLookup(host)
	})
	r.Register("net.interfaces", func(_ map[string]interface{}) (interface{}, error) {
		return opsnet.Interfaces()
	})
	r.Register("net.wait_for", func(args map[string]interface{}) (interface{}, error) {
		host, err := argString(args, "host")
		if err != nil {
			return nil, fmt.Errorf("net.wait_for: %w", err)
		}
		port, err := argInt(args, "port")
		if err != nil {
			return nil, fmt.Errorf("net.wait_for: %w", err)
		}
		timeout, err := argInt(args, "timeout")
		if err != nil {
			return nil, fmt.Errorf("net.wait_for: %w", err)
		}
		return opsnet.WaitFor(host, port, timeout)
	})
}

// ============================================================
// process operations
// ============================================================

func (r *Registry) registerProcessOps() {
	r.Register("process.list", func(_ map[string]interface{}) (interface{}, error) {
		return process.List()
	})
	r.Register("process.find_by_name", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("process.find_by_name: %w", err)
		}
		return process.FindByName(name)
	})
	r.Register("process.find_by_port", func(args map[string]interface{}) (interface{}, error) {
		port, err := argInt(args, "port")
		if err != nil {
			return nil, fmt.Errorf("process.find_by_port: %w", err)
		}
		return process.FindByPort(port)
	})
	r.Register("process.kill", func(args map[string]interface{}) (interface{}, error) {
		pid, err := argInt(args, "pid")
		if err != nil {
			return nil, fmt.Errorf("process.kill: %w", err)
		}
		signal := getStringArg(args, "signal", "TERM")
		return process.Kill(pid, signal)
	})
	r.Register("process.exec", func(args map[string]interface{}) (interface{}, error) {
		command, err := argString(args, "command")
		if err != nil {
			return nil, fmt.Errorf("process.exec: %w", err)
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
	})
}

// ============================================================
// service operations
// ============================================================

func (r *Registry) registerServiceOps() {
	r.Register("service.status", serviceOp("service.status", service.Status))
	r.Register("service.start", serviceOp("service.start", service.Start))
	r.Register("service.stop", serviceOp("service.stop", service.Stop))
	r.Register("service.restart", serviceOp("service.restart", service.Restart))
	r.Register("service.enable", serviceOp("service.enable", service.Enable))
	r.Register("service.disable", serviceOp("service.disable", service.Disable))

	// snap operations
	r.Register("snap.install", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		channel, _ := argString(args, "channel")
		if channel == "" {
			channel = "stable"
		}
		classic, _ := args["classic"].(bool)
		return sdksnap.Install(name, channel, classic)
	})
	r.Register("snap.remove", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		return sdksnap.Remove(name)
	})
	r.Register("snap.refresh", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		channel, _ := argString(args, "channel")
		return sdksnap.Refresh(name, channel)
	})
	r.Register("snap.list", func(_ map[string]interface{}) (interface{}, error) {
		return sdksnap.List()
	})
	r.Register("snap.get", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		return sdksnap.Get(name)
	})
	r.Register("snap.enable", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		return sdksnap.Enable(name)
	})
	r.Register("snap.disable", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		return sdksnap.Disable(name)
	})
	r.Register("snap.switch", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		channel, _ := argString(args, "channel")
		return sdksnap.Switch(name, channel)
	})
	r.Register("snap.changes", func(_ map[string]interface{}) (interface{}, error) {
		return sdksnap.Changes()
	})
}

// serviceOp adapts a service SDK function taking just a name.
func serviceOp[T any](opName string, fn func(string) (T, error)) OperationFunc {
	return func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("%s: %w", opName, err)
		}
		return fn(name)
	}
}

// ============================================================
// selinux operations
// ============================================================

func (r *Registry) registerSelinuxOps() {
	r.Register("selinux.get", func(args map[string]interface{}) (interface{}, error) {
		return sdkselinux.Get()
	})
	r.Register("selinux.set", func(args map[string]interface{}) (interface{}, error) {
		mode, err := argString(args, "mode")
		if err != nil {
			return nil, fmt.Errorf("selinux.set: %w", err)
		}
		return sdkselinux.Set(mode)
	})
}

// ============================================================
// pkg operations
// ============================================================

func (r *Registry) registerPkgOps() {
	r.Register("pkg.install", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("pkg.install: %w", err)
		}
		r, _ := opspkg.Install(name)
		return r, nil
	})
	r.Register("pkg.remove", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("pkg.remove: %w", err)
		}
		r, _ := opspkg.Remove(name)
		return r, nil
	})
	r.Register("pkg.info", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("pkg.info: %w", err)
		}
		return opspkg.Info(name)
	})
	r.Register("pkg.list", func(_ map[string]interface{}) (interface{}, error) {
		return opspkg.List()
	})
}

// ============================================================
// time operations
// ============================================================

func (r *Registry) registerTimeOps() {
	r.Register("time.now", func(_ map[string]interface{}) (interface{}, error) {
		return optime.Now(), nil
	})
	r.Register("time.format", func(args map[string]interface{}) (interface{}, error) {
		ts, err := argInt64(args, "ts")
		if err != nil {
			return nil, fmt.Errorf("time.format: %w", err)
		}
		layout := getStringArg(args, "layout", "2006-01-02 15:04:05")
		return optime.Format(ts, layout)
	})
	r.Register("time.parse", func(args map[string]interface{}) (interface{}, error) {
		layout, err := argString(args, "layout")
		if err != nil {
			return nil, fmt.Errorf("time.parse: %w", err)
		}
		value, err := argString(args, "value")
		if err != nil {
			return nil, fmt.Errorf("time.parse: %w", err)
		}
		return optime.Parse(layout, value)
	})
	r.Register("time.diff", func(args map[string]interface{}) (interface{}, error) {
		t1, err := argInt64(args, "t1")
		if err != nil {
			return nil, fmt.Errorf("time.diff: %w", err)
		}
		t2, err := argInt64(args, "t2")
		if err != nil {
			return nil, fmt.Errorf("time.diff: %w", err)
		}
		return optime.Diff(t1, t2), nil
	})
	r.Register("time.since", func(args map[string]interface{}) (interface{}, error) {
		ts, err := argInt64(args, "ts")
		if err != nil {
			return nil, fmt.Errorf("time.since: %w", err)
		}
		return optime.Since(ts), nil
	})
	r.Register("time.sleep", func(args map[string]interface{}) (interface{}, error) {
		ms, err := argInt(args, "ms")
		if err != nil {
			return nil, fmt.Errorf("time.sleep: %w", err)
		}
		return optime.Sleep(ms)
	})
}

// ============================================================
// json operations
// ============================================================

func (r *Registry) registerJSONOps() {
	r.Register("json.encode", func(args map[string]interface{}) (interface{}, error) {
		data, ok := args["data"]
		if !ok {
			data, ok = args["value"]
			if !ok {
				return nil, fmt.Errorf("json.encode: argument \"value\" is required")
			}
		}
		return opsjson.Encode(data)
	})
	r.Register("json.decode", func(args map[string]interface{}) (interface{}, error) {
		input, err := argString(args, "input")
		if err != nil {
			return nil, fmt.Errorf("json.decode: %w", err)
		}
		return opsjson.Decode(input)
	})
}

// ============================================================
// yaml operations
// ============================================================

func (r *Registry) registerYAMLOps() {
	r.Register("yaml.encode", func(args map[string]interface{}) (interface{}, error) {
		data, ok := args["data"]
		if !ok {
			data, ok = args["value"]
			if !ok {
				return nil, fmt.Errorf("yaml.encode: argument \"value\" is required")
			}
		}
		return opsyaml.Encode(data)
	})
	r.Register("yaml.decode", func(args map[string]interface{}) (interface{}, error) {
		input, err := argString(args, "input")
		if err != nil {
			return nil, fmt.Errorf("yaml.decode: %w", err)
		}
		return opsyaml.Decode(input)
	})
}

// ============================================================
// git operations
// ============================================================

func (r *Registry) registerGitOps() {
	r.Register("git.clone", func(args map[string]interface{}) (interface{}, error) {
		url, err := argString(args, "url")
		if err != nil {
			return nil, fmt.Errorf("git.clone: %w", err)
		}
		dest, err := argString(args, "dest")
		if err != nil {
			return nil, fmt.Errorf("git.clone: %w", err)
		}
		var opts map[string]string
		if raw, ok := args["opts"]; ok && raw != nil {
			if m, ok := raw.(map[string]interface{}); ok {
				opts = make(map[string]string, len(m))
				for k, v := range m {
					opts[k] = fmt.Sprintf("%v", v)
				}
			}
		}
		return opsgit.Clone(url, dest, opts)
	})
	r.Register("git.pull", func(args map[string]interface{}) (interface{}, error) {
		repoPath, err := argString(args, "repo_path")
		if err != nil {
			return nil, fmt.Errorf("git.pull: %w", err)
		}
		remote := getStringArg(args, "remote", "")
		branch := getStringArg(args, "branch", "")
		return opsgit.Pull(repoPath, remote, branch)
	})
}

// ============================================================
// platform operations (user, group, cron, sysctl, mount, firewall, etc.)
// ============================================================

func (r *Registry) registerPlatformOps() {
	// ── user.* ─────────────────────────────────────────────────────────
	r.Register("user.info", func(args map[string]interface{}) (interface{}, error) {
		username, err := argString(args, "username")
		if err != nil {
			return nil, fmt.Errorf("user.info: %w", err)
		}
		return opsuser.Info(username)
	})
	r.Register("user.list", func(_ map[string]interface{}) (interface{}, error) {
		return opsuser.List()
	})
	r.Register("user.add", func(args map[string]interface{}) (interface{}, error) {
		username, err := argString(args, "username")
		if err != nil {
			return nil, fmt.Errorf("user.add: %w", err)
		}
		opts := toStringMapArg(args, "opts")
		return opsuser.Add(username, opts)
	})
	r.Register("user.remove", func(args map[string]interface{}) (interface{}, error) {
		username, err := argString(args, "username")
		if err != nil {
			return nil, fmt.Errorf("user.remove: %w", err)
		}
		removeHome, _ := args["remove_home"].(bool)
		return opsuser.Remove(username, removeHome)
	})
	r.Register("user.modify", func(args map[string]interface{}) (interface{}, error) {
		username, err := argString(args, "username")
		if err != nil {
			return nil, fmt.Errorf("user.modify: %w", err)
		}
		opts := toStringMapArg(args, "opts")
		return opsuser.Modify(username, opts)
	})
	r.Register("user.exists", func(args map[string]interface{}) (interface{}, error) {
		username, err := argString(args, "username")
		if err != nil {
			return nil, fmt.Errorf("user.exists: %w", err)
		}
		return opsuser.Exists(username)
	})

	// ── group.* ────────────────────────────────────────────────────────
	r.Register("group.info", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("group.info: %w", err)
		}
		return opsgrp.Info(name)
	})
	r.Register("group.list", func(_ map[string]interface{}) (interface{}, error) {
		return opsgrp.List()
	})
	r.Register("group.add", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("group.add: %w", err)
		}
		opts := toStringMapArg(args, "opts")
		return opsgrp.Add(name, opts)
	})
	r.Register("group.remove", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("group.remove: %w", err)
		}
		return opsgrp.Remove(name)
	})
	r.Register("group.exists", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("group.exists: %w", err)
		}
		return opsgrp.Exists(name)
	})

	// ── cron.* ─────────────────────────────────────────────────────────
	r.Register("cron.list", func(args map[string]interface{}) (interface{}, error) {
		user, err := argString(args, "user")
		if err != nil {
			return nil, fmt.Errorf("cron.list: %w", err)
		}
		return opscron.List(user)
	})
	r.Register("cron.add", func(args map[string]interface{}) (interface{}, error) {
		user, err := argString(args, "user")
		if err != nil {
			return nil, fmt.Errorf("cron.add: %w", err)
		}
		entryMap, _ := args["entry"].(map[string]interface{})
		entry := opscron.CronEntry{
			Minute:     mapStrArg(entryMap, "minute", "*"),
			Hour:       mapStrArg(entryMap, "hour", "*"),
			DayOfMonth: mapStrArg(entryMap, "day_of_month", "*"),
			Month:      mapStrArg(entryMap, "month", "*"),
			DayOfWeek:  mapStrArg(entryMap, "day_of_week", "*"),
			Command:    mapStrArg(entryMap, "command", ""),
		}
		return opscron.Add(user, entry)
	})
	r.Register("cron.remove", func(args map[string]interface{}) (interface{}, error) {
		user, err := argString(args, "user")
		if err != nil {
			return nil, fmt.Errorf("cron.remove: %w", err)
		}
		lineMatch, err := argString(args, "line_match")
		if err != nil {
			return nil, fmt.Errorf("cron.remove: %w", err)
		}
		return opscron.Remove(user, lineMatch)
	})

	// ── sysctl.* ───────────────────────────────────────────────────────
	r.Register("sysctl.get", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("sysctl.get: %w", err)
		}
		return sdksysctl.Get(name)
	})
	r.Register("sysctl.set", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("sysctl.set: %w", err)
		}
		value, err := argString(args, "value")
		if err != nil {
			return nil, fmt.Errorf("sysctl.set: %w", err)
		}
		return sdksysctl.Set(name, value)
	})
	r.Register("sysctl.list", func(_ map[string]interface{}) (interface{}, error) {
		return sdksysctl.List()
	})

	// ── sys.mount / sys.unmount / sys.list_mounts ──────────────────────
	r.Register("sys.mount", func(args map[string]interface{}) (interface{}, error) {
		device, err := argString(args, "device")
		if err != nil {
			return nil, fmt.Errorf("sys.mount: %w", err)
		}
		mountpoint, err := argString(args, "mountpoint")
		if err != nil {
			return nil, fmt.Errorf("sys.mount: %w", err)
		}
		fsType := getStringArg(args, "fs_type", "")
		opts := toStringMapArg(args, "opts")
		return sys.Mount(device, mountpoint, fsType, opts)
	})
	r.Register("sys.unmount", func(args map[string]interface{}) (interface{}, error) {
		mountpoint, err := argString(args, "mountpoint")
		if err != nil {
			return nil, fmt.Errorf("sys.unmount: %w", err)
		}
		return sys.Unmount(mountpoint)
	})
	r.Register("sys.list_mounts", func(_ map[string]interface{}) (interface{}, error) {
		return sys.ListMounts()
	})

	// ── sys.hostname_set ───────────────────────────────────────────────
	r.Register("sys.hostname_set", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("sys.hostname_set: %w", err)
		}
		return sys.HostnameSet(name)
	})

	// ── firewall.rule ──────────────────────────────────────────────────
	r.Register("firewall.rule", func(args map[string]interface{}) (interface{}, error) {
		action, err := argString(args, "action")
		if err != nil {
			return nil, fmt.Errorf("firewall.rule: %w", err)
		}
		protocol, err := argString(args, "protocol")
		if err != nil {
			return nil, fmt.Errorf("firewall.rule: %w", err)
		}
		port, _ := argInt(args, "port")
		source := getStringArg(args, "source", "")
		return sys.FirewallRule(action, protocol, port, source)
	})

	// ── firewalld ────────────────────────────────────────────────────────
	r.Register("firewalld.get", func(args map[string]interface{}) (interface{}, error) {
		return sdkfirewalld.Get()
	})
	r.Register("firewalld.start", func(args map[string]interface{}) (interface{}, error) {
		return sdkfirewalld.Start()
	})
	r.Register("firewalld.stop", func(args map[string]interface{}) (interface{}, error) {
		return sdkfirewalld.Stop()
	})
	r.Register("firewalld.restart", func(args map[string]interface{}) (interface{}, error) {
		return sdkfirewalld.Restart()
	})
	r.Register("firewalld.enable", func(args map[string]interface{}) (interface{}, error) {
		return sdkfirewalld.Enable()
	})
	r.Register("firewalld.disable", func(args map[string]interface{}) (interface{}, error) {
		return sdkfirewalld.Disable()
	})
	r.Register("firewalld.list_zones", func(args map[string]interface{}) (interface{}, error) {
		return sdkfirewalld.ListZones()
	})
	r.Register("firewalld.reload", func(args map[string]interface{}) (interface{}, error) {
		return sdkfirewalld.Reload()
	})
}

// ============================================================
// extension operations (Ansible-aligned capabilities)
// ============================================================

func (r *Registry) registerExtensions() {
	// ── file.find ────────────────────────────────────────────────────────
	r.Register("file.find", func(args map[string]interface{}) (interface{}, error) {
		opts := file.FindOptions{}
		if paths, ok := args["paths"].([]interface{}); ok {
			for _, p := range paths {
				if s, ok := p.(string); ok {
					opts.Paths = append(opts.Paths, s)
				}
			}
		} else if p, ok := args["paths"].(string); ok {
			opts.Paths = []string{p}
		}
		if pats, ok := args["patterns"].([]interface{}); ok {
			for _, p := range pats {
				if s, ok := p.(string); ok {
					opts.Patterns = append(opts.Patterns, s)
				}
			}
		} else if p, ok := args["patterns"].(string); ok {
			if p != "" {
				opts.Patterns = []string{p}
			}
		}
		if rx, ok := args["regex"].(string); ok {
			opts.Regex = rx
		}
		if ft, ok := args["file_type"].(string); ok {
			opts.FileType = ft
		}
		if md, ok := args["max_depth"].(float64); ok {
			opts.MaxDepth = int(md)
		}
		if age, ok := args["age"].(float64); ok {
			opts.Age = int64(age)
		}
		if sz, ok := args["size"].(float64); ok {
			opts.Size = int64(sz)
		}
		return file.Find(opts)
	})

	// ── file.replace ─────────────────────────────────────────────────────
	r.Register("file.replace", func(args map[string]interface{}) (interface{}, error) {
		path, err := argString(args, "path")
		if err != nil {
			return nil, fmt.Errorf("file.replace: %w", err)
		}
		pattern, err := argString(args, "pattern")
		if err != nil {
			return nil, fmt.Errorf("file.replace: %w", err)
		}
		replacement, err := argString(args, "replacement")
		if err != nil {
			return nil, fmt.Errorf("file.replace: %w", err)
		}
		after := getStringArg(args, "after", "")
		before := getStringArg(args, "before", "")
		return file.Replace(path, pattern, replacement, after, before)
	})

	// ── file.blockinfile ─────────────────────────────────────────────────
	r.Register("file.blockinfile", func(args map[string]interface{}) (interface{}, error) {
		path, err := argString(args, "path")
		if err != nil {
			return nil, fmt.Errorf("file.blockinfile: %w", err)
		}
		marker, err := argString(args, "marker")
		if err != nil {
			return nil, fmt.Errorf("file.blockinfile: %w", err)
		}
		content, err := argString(args, "content")
		if err != nil {
			return nil, fmt.Errorf("file.blockinfile: %w", err)
		}
		present := true
		if v, ok := args["present"].(bool); ok {
			present = v
		}
		insertAfter := getStringArg(args, "insert_after", "")
		insertBefore := getStringArg(args, "insert_before", "")
		return file.BlockInFile(path, marker, content, present, insertAfter, insertBefore)
	})

	// ── file.ini_get ─────────────────────────────────────────────────────
	r.Register("file.ini_get", func(args map[string]interface{}) (interface{}, error) {
		path, err := argString(args, "path")
		if err != nil {
			return nil, fmt.Errorf("file.ini_get: %w", err)
		}
		section, err := argString(args, "section")
		if err != nil {
			return nil, fmt.Errorf("file.ini_get: %w", err)
		}
		key, err := argString(args, "key")
		if err != nil {
			return nil, fmt.Errorf("file.ini_get: %w", err)
		}
		return file.IniGet(path, section, key)
	})

	// ── file.ini_set ─────────────────────────────────────────────────────
	r.Register("file.ini_set", func(args map[string]interface{}) (interface{}, error) {
		path, err := argString(args, "path")
		if err != nil {
			return nil, fmt.Errorf("file.ini_set: %w", err)
		}
		section, err := argString(args, "section")
		if err != nil {
			return nil, fmt.Errorf("file.ini_set: %w", err)
		}
		key, err := argString(args, "key")
		if err != nil {
			return nil, fmt.Errorf("file.ini_set: %w", err)
		}
		value, err := argString(args, "value")
		if err != nil {
			return nil, fmt.Errorf("file.ini_set: %w", err)
		}
		return file.IniSet(path, section, key, value)
	})

	// ── archive.create ───────────────────────────────────────────────────
	r.Register("archive.create", func(args map[string]interface{}) (interface{}, error) {
		dest, err := argString(args, "dest")
		if err != nil {
			return nil, fmt.Errorf("archive.create: %w", err)
		}
		var sources []string
		switch v := args["sources"].(type) {
		case string:
			sources = []string{v}
		case []interface{}:
			for _, s := range v {
				if str, ok := s.(string); ok {
					sources = append(sources, str)
				}
			}
		}
		return sdkarchive.Create(dest, sources)
	})

	// ── archive.extract ──────────────────────────────────────────────────
	r.Register("archive.extract", func(args map[string]interface{}) (interface{}, error) {
		src, err := argString(args, "src")
		if err != nil {
			return nil, fmt.Errorf("archive.extract: %w", err)
		}
		dest, err := argString(args, "dest")
		if err != nil {
			return nil, fmt.Errorf("archive.extract: %w", err)
		}
		return sdkarchive.Extract(src, dest)
	})

	// ── net.download ─────────────────────────────────────────────────────
	r.Register("net.download", func(args map[string]interface{}) (interface{}, error) {
		url, err := argString(args, "url")
		if err != nil {
			return nil, fmt.Errorf("net.download: %w", err)
		}
		dest, err := argString(args, "dest")
		if err != nil {
			return nil, fmt.Errorf("net.download: %w", err)
		}
		algo := getStringArg(args, "checksum_algo", "")
		expected := getStringArg(args, "checksum_expected", "")
		return opsnet.Download(url, dest, algo, expected)
	})

	// ── net.wait_for_connection ──────────────────────────────────────────
	r.Register("net.wait_for_connection", func(args map[string]interface{}) (interface{}, error) {
		host, err := argString(args, "host")
		if err != nil {
			return nil, fmt.Errorf("net.wait_for_connection: %w", err)
		}
		port, err := argInt(args, "port")
		if err != nil {
			return nil, fmt.Errorf("net.wait_for_connection: %w", err)
		}
		timeout := 30
		if t, ok := args["timeout"].(float64); ok {
			timeout = int(t)
		}
		return opsnet.WaitForConnection(host, port, timeout)
	})

	// ── sys.timezone_get ─────────────────────────────────────────────────
	r.Register("sys.timezone_get", func(_ map[string]interface{}) (interface{}, error) {
		return sys.TimezoneGet()
	})

	// ── sys.timezone_set ─────────────────────────────────────────────────
	r.Register("sys.timezone_set", func(args map[string]interface{}) (interface{}, error) {
		tz, err := argString(args, "timezone")
		if err != nil {
			return nil, fmt.Errorf("sys.timezone_set: %w", err)
		}
		return sys.TimezoneSet(tz)
	})

	// ── sys.reboot ───────────────────────────────────────────────────────
	r.Register("sys.reboot", func(_ map[string]interface{}) (interface{}, error) {
		return sys.Reboot()
	})

	// ── ssh.authorized_key_add ───────────────────────────────────────────
	r.Register("ssh.authorized_key_add", func(args map[string]interface{}) (interface{}, error) {
		user, err := argString(args, "user")
		if err != nil {
			return nil, fmt.Errorf("ssh.authorized_key_add: %w", err)
		}
		key, err := argString(args, "key")
		if err != nil {
			return nil, fmt.Errorf("ssh.authorized_key_add: %w", err)
		}
		exclusive, _ := args["exclusive"].(bool)
		return sdkssh.AuthorizedKeyAdd(user, key, exclusive)
	})

	// ── ssh.authorized_key_remove ────────────────────────────────────────
	r.Register("ssh.authorized_key_remove", func(args map[string]interface{}) (interface{}, error) {
		user, err := argString(args, "user")
		if err != nil {
			return nil, fmt.Errorf("ssh.authorized_key_remove: %w", err)
		}
		key, err := argString(args, "key")
		if err != nil {
			return nil, fmt.Errorf("ssh.authorized_key_remove: %w", err)
		}
		return sdkssh.AuthorizedKeyRemove(user, key)
	})

	// ── ssh.authorized_key_list ──────────────────────────────────────────
	r.Register("ssh.authorized_key_list", func(args map[string]interface{}) (interface{}, error) {
		user, err := argString(args, "user")
		if err != nil {
			return nil, fmt.Errorf("ssh.authorized_key_list: %w", err)
		}
		return sdkssh.AuthorizedKeyList(user)
	})

	// ── kernel.module_list ───────────────────────────────────────────────
	r.Register("kernel.module_list", func(_ map[string]interface{}) (interface{}, error) {
		return sdkkernel.ModuleList()
	})

	// ── kernel.module_load ───────────────────────────────────────────────
	r.Register("kernel.module_load", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("kernel.module_load: %w", err)
		}
		return sdkkernel.ModuleLoad(name)
	})

	// ── kernel.module_unload ─────────────────────────────────────────────
	r.Register("kernel.module_unload", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("kernel.module_unload: %w", err)
		}
		return sdkkernel.ModuleUnload(name)
	})

	// ── disk.filesystem ──────────────────────────────────────────────────
	r.Register("disk.filesystem", func(args map[string]interface{}) (interface{}, error) {
		device, err := argString(args, "device")
		if err != nil {
			return nil, fmt.Errorf("disk.filesystem: %w", err)
		}
		fsType := getStringArg(args, "fs_type", "ext4")
		return sdkdisk.FilesystemCreate(device, fsType)
	})

	// ── disk.part_list ───────────────────────────────────────────────────
	r.Register("disk.part_list", func(args map[string]interface{}) (interface{}, error) {
		device, err := argString(args, "device")
		if err != nil {
			return nil, fmt.Errorf("disk.part_list: %w", err)
		}
		return sdkdisk.PartList(device)
	})

	// ── docker.container_list ─────────────────────────────────────────────
	r.Register("docker.container_list", func(args map[string]interface{}) (interface{}, error) {
		all := getBoolArg(args, "all", false)
		return sdkdocker.ContainerList(all)
	})

	// ── docker.container_exists ───────────────────────────────────────────
	r.Register("docker.container_exists", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("docker.container_exists: %w", err)
		}
		return sdkdocker.ContainerExists(name)
	})

	// ── docker.container_run ──────────────────────────────────────────────
	r.Register("docker.container_run", func(args map[string]interface{}) (interface{}, error) {
		name := getStringArg(args, "name", "")
		image, err := argString(args, "image")
		if err != nil {
			return nil, fmt.Errorf("docker.container_run: %w", err)
		}
		opts := toStringMapArg(args, "opts")
		return sdkdocker.ContainerRun(name, image, opts)
	})

	// ── docker.container_stop ─────────────────────────────────────────────
	r.Register("docker.container_stop", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("docker.container_stop: %w", err)
		}
		return sdkdocker.ContainerStop(name)
	})

	// ── docker.container_remove ───────────────────────────────────────────
	r.Register("docker.container_remove", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("docker.container_remove: %w", err)
		}
		force := getBoolArg(args, "force", false)
		return sdkdocker.ContainerRemove(name, force)
	})

	// ── docker.image_list ─────────────────────────────────────────────────
	r.Register("docker.image_list", func(args map[string]interface{}) (interface{}, error) {
		return sdkdocker.ImageList()
	})

	// ── docker.image_pull ─────────────────────────────────────────────────
	r.Register("docker.image_pull", func(args map[string]interface{}) (interface{}, error) {
		image, err := argString(args, "image")
		if err != nil {
			return nil, fmt.Errorf("docker.image_pull: %w", err)
		}
		return sdkdocker.ImagePull(image)
	})

	// ── docker.image_remove ───────────────────────────────────────────────
	r.Register("docker.image_remove", func(args map[string]interface{}) (interface{}, error) {
		image, err := argString(args, "image")
		if err != nil {
			return nil, fmt.Errorf("docker.image_remove: %w", err)
		}
		force := getBoolArg(args, "force", false)
		return sdkdocker.ImageRemove(image, force)
	})

	// ── hosts.list ────────────────────────────────────────────────────────
	r.Register("hosts.list", func(args map[string]interface{}) (interface{}, error) {
		return opshosts.List()
	})

	// ── hosts.exists ──────────────────────────────────────────────────────
	r.Register("hosts.exists", func(args map[string]interface{}) (interface{}, error) {
		hostname, err := argString(args, "hostname")
		if err != nil {
			return nil, fmt.Errorf("hosts.exists: %w", err)
		}
		return opshosts.Exists(hostname)
	})

	// ── hosts.add ─────────────────────────────────────────────────────────
	r.Register("hosts.add", func(args map[string]interface{}) (interface{}, error) {
		ip, err := argString(args, "ip")
		if err != nil {
			return nil, fmt.Errorf("hosts.add: %w", err)
		}
		hostnamesRaw, ok := args["hostnames"].([]interface{})
		if !ok {
			return nil, fmt.Errorf("hosts.add: hostnames must be an array")
		}
		hostnames := make([]string, len(hostnamesRaw))
		for i, v := range hostnamesRaw {
			if s, ok := v.(string); ok {
				hostnames[i] = s
			}
		}
		return opshosts.Add(ip, hostnames)
	})

	// ── hosts.remove ──────────────────────────────────────────────────────
	r.Register("hosts.remove", func(args map[string]interface{}) (interface{}, error) {
		hostnamesRaw, ok := args["hostnames"].([]interface{})
		if !ok {
			return nil, fmt.Errorf("hosts.remove: hostnames must be an array")
		}
		hostnames := make([]string, len(hostnamesRaw))
		for i, v := range hostnamesRaw {
			if s, ok := v.(string); ok {
				hostnames[i] = s
			}
		}
		return opshosts.Remove(hostnames)
	})

	// ── locale.get ────────────────────────────────────────────────────────
	r.Register("locale.get", func(args map[string]interface{}) (interface{}, error) {
		return sdklocale.Get()
	})

	// ── locale.available ──────────────────────────────────────────────────
	r.Register("locale.available", func(args map[string]interface{}) (interface{}, error) {
		return sdklocale.Available()
	})

	// ── locale.set ────────────────────────────────────────────────────────
	r.Register("locale.set", func(args map[string]interface{}) (interface{}, error) {
		locale, err := argString(args, "locale")
		if err != nil {
			return nil, fmt.Errorf("locale.set: %w", err)
		}
		return sdklocale.Set(locale)
	})

	// ── pip.list ──────────────────────────────────────────────────────────
	r.Register("pip.list", func(args map[string]interface{}) (interface{}, error) {
		return sdkpip.List()
	})

	// ── pip.exists ────────────────────────────────────────────────────────
	r.Register("pip.exists", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("pip.exists: %w", err)
		}
		return sdkpip.Exists(name)
	})

	// ── pip.install ───────────────────────────────────────────────────────
	r.Register("pip.install", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("pip.install: %w", err)
		}
		version := getStringArg(args, "version", "")
		return sdkpip.Install(name, version)
	})

	// ── pip.uninstall ─────────────────────────────────────────────────────
	r.Register("pip.uninstall", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("pip.uninstall: %w", err)
		}
		return sdkpip.Uninstall(name)
	})

	// ── apt.* ─────────────────────────────────────────────────────────────
	r.Register("apt.install", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("apt.install: %w", err)
		}
		version, _ := args["version"].(string)
		updateCache, _ := args["update_cache"].(bool)
		return sdkapt.Install(name, version, updateCache)
	})
	r.Register("apt.remove", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("apt.remove: %w", err)
		}
		purge, _ := args["purge"].(bool)
		return sdkapt.Remove(name, purge)
	})
	r.Register("apt.upgrade", func(args map[string]interface{}) (interface{}, error) {
		name, _ := args["name"].(string)
		return sdkapt.Upgrade(name)
	})
	r.Register("apt.update_cache", func(_ map[string]interface{}) (interface{}, error) {
		return sdkapt.UpdateCache()
	})
	r.Register("apt.full_upgrade", func(_ map[string]interface{}) (interface{}, error) {
		return sdkapt.FullUpgrade()
	})
	r.Register("apt.dist_upgrade", func(_ map[string]interface{}) (interface{}, error) {
		return sdkapt.DistUpgrade()
	})
	r.Register("apt.autoremove", func(_ map[string]interface{}) (interface{}, error) {
		return sdkapt.Autoremove()
	})
	r.Register("apt.clean", func(_ map[string]interface{}) (interface{}, error) {
		return sdkapt.Clean()
	})
	r.Register("apt.info", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("apt.info: %w", err)
		}
		return sdkapt.Info(name)
	})
	r.Register("apt.list", func(_ map[string]interface{}) (interface{}, error) {
		return sdkapt.List()
	})
	r.Register("apt.policy", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("apt.policy: %w", err)
		}
		return sdkapt.Policy(name)
	})
	r.Register("apt.mark_auto", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("apt.mark_auto: %w", err)
		}
		return sdkapt.MarkAuto(name)
	})
	r.Register("apt.mark_manual", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("apt.mark_manual: %w", err)
		}
		return sdkapt.MarkManual(name)
	})

	// ── dnf.* ─────────────────────────────────────────────────────────
	r.Register("dnf.install", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("dnf.install: %w", err)
		}
		version, _ := args["version"].(string)
		return sdkdnf.Install(name, version)
	})
	r.Register("dnf.remove", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("dnf.remove: %w", err)
		}
		return sdkdnf.Remove(name)
	})
	r.Register("dnf.update", func(args map[string]interface{}) (interface{}, error) {
		name, _ := args["name"].(string)
		return sdkdnf.Update(name)
	})
	r.Register("dnf.info", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("dnf.info: %w", err)
		}
		return sdkdnf.Info(name)
	})
	r.Register("dnf.list", func(_ map[string]interface{}) (interface{}, error) {
		return sdkdnf.List()
	})
	r.Register("dnf.search", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("dnf.search: %w", err)
		}
		return sdkdnf.Search(name)
	})
	r.Register("dnf.clean", func(_ map[string]interface{}) (interface{}, error) {
		return sdkdnf.Clean()
	})
	r.Register("dnf.repolist", func(_ map[string]interface{}) (interface{}, error) {
		return sdkdnf.RepoList()
	})
	r.Register("dnf.grouplist", func(_ map[string]interface{}) (interface{}, error) {
		return sdkdnf.GroupList()
	})
	r.Register("dnf.groupinstall", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("dnf.groupinstall: %w", err)
		}
		return sdkdnf.GroupInstall(name)
	})
	r.Register("dnf.groupremove", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("dnf.groupremove: %w", err)
		}
		return sdkdnf.GroupRemove(name)
	})
	r.Register("dnf.history", func(args map[string]interface{}) (interface{}, error) {
		count := 10
		if v, ok := args["count"].(float64); ok {
			count = int(v)
		}
		return sdkdnf.History(count)
	})
	r.Register("dnf.check_update", func(_ map[string]interface{}) (interface{}, error) {
		return sdkdnf.CheckUpdate()
	})
	r.Register("dnf.modulelist", func(_ map[string]interface{}) (interface{}, error) {
		return sdkdnf.ModuleList()
	})
	r.Register("dnf.module_enable", func(args map[string]interface{}) (interface{}, error) {
		spec, err := argString(args, "spec")
		if err != nil {
			return nil, fmt.Errorf("dnf.module_enable: %w", err)
		}
		return sdkdnf.ModuleEnable(spec)
	})

	// ── apk.* ─────────────────────────────────────────────────────────
	r.Register("apk.install", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("apk.install: %w", err)
		}
		version, _ := args["version"].(string)
		return sdkapk.Install(name, version)
	})
	r.Register("apk.remove", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("apk.remove: %w", err)
		}
		purge, _ := args["purge"].(bool)
		return sdkapk.Remove(name, purge)
	})
	r.Register("apk.update", func(_ map[string]interface{}) (interface{}, error) {
		return sdkapk.Update()
	})
	r.Register("apk.upgrade", func(args map[string]interface{}) (interface{}, error) {
		name, _ := args["name"].(string)
		return sdkapk.Upgrade(name)
	})
	r.Register("apk.info", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("apk.info: %w", err)
		}
		return sdkapk.Info(name)
	})
	r.Register("apk.list", func(_ map[string]interface{}) (interface{}, error) {
		return sdkapk.List()
	})
	r.Register("apk.search", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("apk.search: %w", err)
		}
		return sdkapk.Search(name)
	})
	r.Register("apk.cache", func(_ map[string]interface{}) (interface{}, error) {
		return sdkapk.Cache()
	})
	r.Register("apk.upgrade_available", func(_ map[string]interface{}) (interface{}, error) {
		return sdkapk.UpgradeAvailable()
	})
	r.Register("apk.repository", func(_ map[string]interface{}) (interface{}, error) {
		return sdkapk.Repository()
	})

	// ── sysvinit.* ────────────────────────────────────────────────────
	r.Register("sysvinit.status", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("sysvinit.status: %w", err)
		}
		return sdksysvinit.Status(name)
	})
	r.Register("sysvinit.start", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("sysvinit.start: %w", err)
		}
		return sdksysvinit.Start(name)
	})
	r.Register("sysvinit.stop", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("sysvinit.stop: %w", err)
		}
		return sdksysvinit.Stop(name)
	})
	r.Register("sysvinit.restart", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("sysvinit.restart: %w", err)
		}
		return sdksysvinit.Restart(name)
	})
	r.Register("sysvinit.reload", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("sysvinit.reload: %w", err)
		}
		return sdksysvinit.Reload(name)
	})
	r.Register("sysvinit.enable", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("sysvinit.enable: %w", err)
		}
		runlevels, _ := args["runlevels"].(string)
		return sdksysvinit.Enable(name, runlevels)
	})
	r.Register("sysvinit.disable", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("sysvinit.disable: %w", err)
		}
		return sdksysvinit.Disable(name)
	})
	r.Register("sysvinit.list", func(_ map[string]interface{}) (interface{}, error) {
		return sdksysvinit.List()
	})

	// ── dpkg_selections.* ─────────────────────────────────────────────
	r.Register("dpkg_selections.set", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("dpkg_selections.set: %w", err)
		}
		state, err := argString(args, "state")
		if err != nil {
			return nil, fmt.Errorf("dpkg_selections.set: %w", err)
		}
		return sdkdpkgsel.SetSelection(name, state)
	})
	r.Register("dpkg_selections.get", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("dpkg_selections.get: %w", err)
		}
		return sdkdpkgsel.GetSelection(name)
	})
	r.Register("dpkg_selections.list", func(_ map[string]interface{}) (interface{}, error) {
		return sdkdpkgsel.ListSelections()
	})
	r.Register("dpkg_selections.hold", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("dpkg_selections.hold: %w", err)
		}
		return sdkdpkgsel.Hold(name)
	})
	r.Register("dpkg_selections.unhold", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("dpkg_selections.unhold: %w", err)
		}
		return sdkdpkgsel.Unhold(name)
	})

	// ── homebrew.* ────────────────────────────────────────────────────
	r.Register("homebrew.install", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("homebrew.install: %w", err)
		}
		cask, _ := args["cask"].(bool)
		return sdkbrew.Install(name, cask)
	})
	r.Register("homebrew.remove", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("homebrew.remove: %w", err)
		}
		cask, _ := args["cask"].(bool)
		return sdkbrew.Remove(name, cask)
	})
	r.Register("homebrew.upgrade", func(args map[string]interface{}) (interface{}, error) {
		name, _ := args["name"].(string)
		return sdkbrew.Upgrade(name)
	})
	r.Register("homebrew.update", func(_ map[string]interface{}) (interface{}, error) {
		return sdkbrew.Update()
	})
	r.Register("homebrew.info", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("homebrew.info: %w", err)
		}
		return sdkbrew.Info(name)
	})
	r.Register("homebrew.list", func(_ map[string]interface{}) (interface{}, error) {
		return sdkbrew.List()
	})
	r.Register("homebrew.list_casks", func(_ map[string]interface{}) (interface{}, error) {
		return sdkbrew.ListCasks()
	})
	r.Register("homebrew.outdated", func(_ map[string]interface{}) (interface{}, error) {
		return sdkbrew.Outdated()
	})
	r.Register("homebrew.clean", func(_ map[string]interface{}) (interface{}, error) {
		return sdkbrew.Clean()
	})
	r.Register("homebrew.tap", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("homebrew.tap: %w", err)
		}
		return sdkbrew.Tap(name)
	})
	r.Register("homebrew.untap", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("homebrew.untap: %w", err)
		}
		return sdkbrew.Untap(name)
	})
	r.Register("homebrew.list_taps", func(_ map[string]interface{}) (interface{}, error) {
		return sdkbrew.ListTaps()
	})
	r.Register("homebrew.doctor", func(_ map[string]interface{}) (interface{}, error) {
		return sdkbrew.Doctor()
	})

	// ── apt_repo.* ──────────────────────────────────────────────────────
	r.Register("apt_repo.list", func(_ map[string]interface{}) (interface{}, error) {
		return sdkaptrepo.List()
	})
	r.Register("apt_repo.exists", func(args map[string]interface{}) (interface{}, error) {
		uri, err := argString(args, "uri")
		if err != nil {
			return nil, fmt.Errorf("apt_repo.exists: %w", err)
		}
		return sdkaptrepo.Exists(uri)
	})
	r.Register("apt_repo.add", func(args map[string]interface{}) (interface{}, error) {
		uri, err := argString(args, "uri")
		if err != nil {
			return nil, fmt.Errorf("apt_repo.add: %w", err)
		}
		dist, _ := args["dist"].(string)
		comps, _ := args["components"].(string)
		return sdkaptrepo.Add(uri, dist, comps)
	})
	r.Register("apt_repo.remove", func(args map[string]interface{}) (interface{}, error) {
		uri, err := argString(args, "uri")
		if err != nil {
			return nil, fmt.Errorf("apt_repo.remove: %w", err)
		}
		return sdkaptrepo.Remove(uri)
	})
	r.Register("apt_repo.update", func(_ map[string]interface{}) (interface{}, error) {
		return sdkaptrepo.Update()
	})

	// ── logrotate.* ─────────────────────────────────────────────────────
	r.Register("logrotate.list", func(_ map[string]interface{}) (interface{}, error) {
		return sdklogrotate.List()
	})
	r.Register("logrotate.get", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("logrotate.get: %w", err)
		}
		return sdklogrotate.Get(name)
	})
	r.Register("logrotate.set", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("logrotate.set: %w", err)
		}
		pattern, _ := args["pattern"].(string)
		freq, _ := args["frequency"].(string)
		rotate := 7
		if v, ok := args["rotate"].(float64); ok {
			rotate = int(v)
		}
		compress := getBoolArg(args, "compress", false)
		postRotate, _ := args["post_rotate"].(string)
		return sdklogrotate.Set(name, pattern, freq, rotate, compress, postRotate)
	})
	r.Register("logrotate.remove", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("logrotate.remove: %w", err)
		}
		return sdklogrotate.Remove(name)
	})

	// ── lvg.* ─────────────────────────────────────────────────────────────
	r.Register("lvg.create", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		pvsRaw, _ := args["pvs"].([]interface{})
		pvs := make([]string, len(pvsRaw))
		for i, pv := range pvsRaw {
			pvs[i], _ = pv.(string)
		}
		return sdklvg.Create(name, pvs)
	})
	r.Register("lvg.remove", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		return sdklvg.Remove(name)
	})
	r.Register("lvg.extend", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		pvsRaw, _ := args["pvs"].([]interface{})
		pvs := make([]string, len(pvsRaw))
		for i, pv := range pvsRaw {
			pvs[i], _ = pv.(string)
		}
		return sdklvg.Extend(name, pvs)
	})
	r.Register("lvg.reduce", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		pvsRaw, _ := args["pvs"].([]interface{})
		pvs := make([]string, len(pvsRaw))
		for i, pv := range pvsRaw {
			pvs[i], _ = pv.(string)
		}
		return sdklvg.Reduce(name, pvs)
	})
	r.Register("lvg.activate", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		return sdklvg.Activate(name)
	})
	r.Register("lvg.deactivate", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		return sdklvg.Deactivate(name)
	})
	r.Register("lvg.list", func(_ map[string]interface{}) (interface{}, error) {
		return sdklvg.List()
	})
	r.Register("lvg.get", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		return sdklvg.Get(name)
	})

	// ── resolv.* ────────────────────────────────────────────────────────
	r.Register("resolv.get", func(_ map[string]interface{}) (interface{}, error) {
		return sdkresolv.Get()
	})
	r.Register("resolv.set", func(args map[string]interface{}) (interface{}, error) {
		var nameservers, search, options []string
		if ns, ok := args["nameservers"].([]interface{}); ok {
			for _, v := range ns {
				if s, ok := v.(string); ok {
					nameservers = append(nameservers, s)
				}
			}
		}
		if s, ok := args["search"].([]interface{}); ok {
			for _, v := range s {
				if str, ok := v.(string); ok {
					search = append(search, str)
				}
			}
		}
		if o, ok := args["options"].([]interface{}); ok {
			for _, v := range o {
				if str, ok := v.(string); ok {
					options = append(options, str)
				}
			}
		}
		domain, _ := args["domain"].(string)
		return sdkresolv.Set(nameservers, search, options, domain)
	})
	r.Register("resolv.add_nameserver", func(args map[string]interface{}) (interface{}, error) {
		ns, err := argString(args, "nameserver")
		if err != nil {
			return nil, fmt.Errorf("resolv.add_nameserver: %w", err)
		}
		return sdkresolv.AddNameserver(ns)
	})
	r.Register("resolv.remove_nameserver", func(args map[string]interface{}) (interface{}, error) {
		ns, err := argString(args, "nameserver")
		if err != nil {
			return nil, fmt.Errorf("resolv.remove_nameserver: %w", err)
		}
		return sdkresolv.RemoveNameserver(ns)
	})

	// ── yum_repo.* ──────────────────────────────────────────────────────
	r.Register("yum_repo.list", func(_ map[string]interface{}) (interface{}, error) {
		return sdkyumrepo.List()
	})
	r.Register("yum_repo.exists", func(args map[string]interface{}) (interface{}, error) {
		id, err := argString(args, "id")
		if err != nil {
			return nil, fmt.Errorf("yum_repo.exists: %w", err)
		}
		return sdkyumrepo.Exists(id)
	})
	r.Register("yum_repo.add", func(args map[string]interface{}) (interface{}, error) {
		id, err := argString(args, "id")
		if err != nil {
			return nil, fmt.Errorf("yum_repo.add: %w", err)
		}
		name, _ := args["name"].(string)
		baseURL, _ := args["base_url"].(string)
		gpgCheck := getBoolArg(args, "gpg_check", false)
		gpgKey, _ := args["gpg_key"].(string)
		return sdkyumrepo.Add(id, name, baseURL, gpgCheck, gpgKey)
	})
	r.Register("yum_repo.remove", func(args map[string]interface{}) (interface{}, error) {
		id, err := argString(args, "id")
		if err != nil {
			return nil, fmt.Errorf("yum_repo.remove: %w", err)
		}
		return sdkyumrepo.Remove(id)
	})

	// ── known_hosts ──────────────────────────────────────────────────────
	r.Register("known_hosts.list", func(args map[string]interface{}) (interface{}, error) {
		return sdkknownhosts.List()
	})
	r.Register("known_hosts.check", func(args map[string]interface{}) (interface{}, error) {
		host, err := argString(args, "host")
		if err != nil {
			return nil, fmt.Errorf("known_hosts.check: %w", err)
		}
		return sdkknownhosts.Check(host)
	})
	r.Register("known_hosts.add", func(args map[string]interface{}) (interface{}, error) {
		host, err := argString(args, "host")
		if err != nil {
			return nil, fmt.Errorf("known_hosts.add: %w", err)
		}
		return sdkknownhosts.Add(host)
	})
	r.Register("known_hosts.remove", func(args map[string]interface{}) (interface{}, error) {
		host, err := argString(args, "host")
		if err != nil {
			return nil, fmt.Errorf("known_hosts.remove: %w", err)
		}
		return sdkknownhosts.Remove(host)
	})

	// ── limits ───────────────────────────────────────────────────────────
	r.Register("limits.list", func(args map[string]interface{}) (interface{}, error) {
		return sdklimits.List()
	})
	r.Register("limits.get", func(args map[string]interface{}) (interface{}, error) {
		domain, err := argString(args, "domain")
		if err != nil {
			return nil, fmt.Errorf("limits.get: %w", err)
		}
		return sdklimits.Get(domain)
	})
	r.Register("limits.set", func(args map[string]interface{}) (interface{}, error) {
		domain, _ := argString(args, "domain")
		typ, _ := argString(args, "type")
		item, _ := argString(args, "item")
		value, _ := argString(args, "value")
		return sdklimits.Set(domain, typ, item, value)
	})
	r.Register("limits.remove", func(args map[string]interface{}) (interface{}, error) {
		domain, err := argString(args, "domain")
		if err != nil {
			return nil, fmt.Errorf("limits.remove: %w", err)
		}
		return sdklimits.Remove(domain)
	})

	// ── ntp ──────────────────────────────────────────────────────────────
	r.Register("ntp.get", func(args map[string]interface{}) (interface{}, error) {
		return sdkntp.Get()
	})
	r.Register("ntp.set", func(args map[string]interface{}) (interface{}, error) {
		server, err := argString(args, "server")
		if err != nil {
			return nil, fmt.Errorf("ntp.set: %w", err)
		}
		return sdkntp.Set(server)
	})

	// ── ufw ──────────────────────────────────────────────────────────────
	r.Register("ufw.status", func(args map[string]interface{}) (interface{}, error) {
		return sdkufw.Status()
	})
	r.Register("ufw.list", func(args map[string]interface{}) (interface{}, error) {
		return sdkufw.List()
	})
	r.Register("ufw.enable", func(args map[string]interface{}) (interface{}, error) {
		return sdkufw.Enable()
	})
	r.Register("ufw.disable", func(args map[string]interface{}) (interface{}, error) {
		return sdkufw.Disable()
	})
	r.Register("ufw.allow", func(args map[string]interface{}) (interface{}, error) {
		port, err := argString(args, "port")
		if err != nil {
			return nil, fmt.Errorf("ufw.allow: %w", err)
		}
		proto, _ := argString(args, "proto")
		if proto == "" {
			proto = "tcp"
		}
		return sdkufw.Allow(port, proto)
	})
	r.Register("ufw.deny", func(args map[string]interface{}) (interface{}, error) {
		port, err := argString(args, "port")
		if err != nil {
			return nil, fmt.Errorf("ufw.deny: %w", err)
		}
		proto, _ := argString(args, "proto")
		if proto == "" {
			proto = "tcp"
		}
		return sdkufw.Deny(port, proto)
	})
	r.Register("ufw.delete", func(args map[string]interface{}) (interface{}, error) {
		num, err := argInt(args, "number")
		if err != nil {
			return nil, fmt.Errorf("ufw.delete: %w", err)
		}
		return sdkufw.Delete(num)
	})
	r.Register("ufw.reset", func(args map[string]interface{}) (interface{}, error) {
		return sdkufw.Reset()
	})
	r.Register("ufw.reload", func(args map[string]interface{}) (interface{}, error) {
		return sdkufw.Reload()
	})

	// ── ini_file ─────────────────────────────────────────────────────────
	r.Register("ini_file.sections", func(args map[string]interface{}) (interface{}, error) {
		path, err := argString(args, "path")
		if err != nil {
			return nil, fmt.Errorf("ini_file.sections: %w", err)
		}
		return sdkinifile.Sections(path)
	})
	r.Register("ini_file.get", func(args map[string]interface{}) (interface{}, error) {
		path, err := argString(args, "path")
		if err != nil {
			return nil, fmt.Errorf("ini_file.get: %w", err)
		}
		section, err := argString(args, "section")
		if err != nil {
			return nil, fmt.Errorf("ini_file.get: %w", err)
		}
		key, err := argString(args, "key")
		if err != nil {
			return nil, fmt.Errorf("ini_file.get: %w", err)
		}
		return sdkinifile.Get(path, section, key)
	})
	r.Register("ini_file.set", func(args map[string]interface{}) (interface{}, error) {
		path, err := argString(args, "path")
		if err != nil {
			return nil, fmt.Errorf("ini_file.set: %w", err)
		}
		section, err := argString(args, "section")
		if err != nil {
			return nil, fmt.Errorf("ini_file.set: %w", err)
		}
		key, err := argString(args, "key")
		if err != nil {
			return nil, fmt.Errorf("ini_file.set: %w", err)
		}
		value, err := argString(args, "value")
		if err != nil {
			return nil, fmt.Errorf("ini_file.set: %w", err)
		}
		return sdkinifile.Set(path, section, key, value)
	})
	r.Register("ini_file.remove", func(args map[string]interface{}) (interface{}, error) {
		path, err := argString(args, "path")
		if err != nil {
			return nil, fmt.Errorf("ini_file.remove: %w", err)
		}
		section, err := argString(args, "section")
		if err != nil {
			return nil, fmt.Errorf("ini_file.remove: %w", err)
		}
		key, err := argString(args, "key")
		if err != nil {
			return nil, fmt.Errorf("ini_file.remove: %w", err)
		}
		return sdkinifile.Remove(path, section, key)
	})
	r.Register("ini_file.remove_section", func(args map[string]interface{}) (interface{}, error) {
		path, err := argString(args, "path")
		if err != nil {
			return nil, fmt.Errorf("ini_file.remove_section: %w", err)
		}
		section, err := argString(args, "section")
		if err != nil {
			return nil, fmt.Errorf("ini_file.remove_section: %w", err)
		}
		return sdkinifile.RemoveSection(path, section)
	})

	// ── mount ────────────────────────────────────────────────────────────
	r.Register("mount.list", func(args map[string]interface{}) (interface{}, error) {
		return sdkmount.List()
	})
	r.Register("mount.mount", func(args map[string]interface{}) (interface{}, error) {
		device, err := argString(args, "device")
		if err != nil {
			return nil, fmt.Errorf("mount.mount: %w", err)
		}
		mountpoint, err := argString(args, "mountpoint")
		if err != nil {
			return nil, fmt.Errorf("mount.mount: %w", err)
		}
		fstype, _ := argString(args, "fstype")
		options, _ := argString(args, "options")
		return sdkmount.Mount(device, mountpoint, fstype, options)
	})
	r.Register("mount.umount", func(args map[string]interface{}) (interface{}, error) {
		mountpoint, err := argString(args, "mountpoint")
		if err != nil {
			return nil, fmt.Errorf("mount.umount: %w", err)
		}
		return sdkmount.Unmount(mountpoint)
	})
	r.Register("mount.fstab", func(args map[string]interface{}) (interface{}, error) {
		return sdkmount.Fstab()
	})
	r.Register("mount.add_fstab", func(args map[string]interface{}) (interface{}, error) {
		device, err := argString(args, "device")
		if err != nil {
			return nil, fmt.Errorf("mount.add_fstab: %w", err)
		}
		mountpoint, err := argString(args, "mountpoint")
		if err != nil {
			return nil, fmt.Errorf("mount.add_fstab: %w", err)
		}
		fstype, err := argString(args, "fstype")
		if err != nil {
			return nil, fmt.Errorf("mount.add_fstab: %w", err)
		}
		options, _ := argString(args, "options")
		return sdkmount.AddFstab(device, mountpoint, fstype, options)
	})
	r.Register("mount.remove_fstab", func(args map[string]interface{}) (interface{}, error) {
		target, err := argString(args, "target")
		if err != nil {
			return nil, fmt.Errorf("mount.remove_fstab: %w", err)
		}
		return sdkmount.RemoveFstab(target)
	})

	// ── hostname ─────────────────────────────────────────────────────────
	r.Register("hostname.get", func(args map[string]interface{}) (interface{}, error) {
		return sdkhostname.Get()
	})
	r.Register("hostname.set", func(args map[string]interface{}) (interface{}, error) {
		hostname, err := argString(args, "hostname")
		if err != nil {
			return nil, fmt.Errorf("hostname.set: %w", err)
		}
		return sdkhostname.Set(hostname)
	})
	r.Register("hostname.set_fqdn", func(args map[string]interface{}) (interface{}, error) {
		fqdn, err := argString(args, "fqdn")
		if err != nil {
			return nil, fmt.Errorf("hostname.set_fqdn: %w", err)
		}
		return sdkhostname.SetFQDN(fqdn)
	})

	// ── timezone ─────────────────────────────────────────────────────────
	r.Register("timezone.get", func(args map[string]interface{}) (interface{}, error) {
		return sdktimezone.Get()
	})
	r.Register("timezone.set", func(args map[string]interface{}) (interface{}, error) {
		timezone, err := argString(args, "timezone")
		if err != nil {
			return nil, fmt.Errorf("timezone.set: %w", err)
		}
		return sdktimezone.Set(timezone)
	})
	r.Register("timezone.list", func(args map[string]interface{}) (interface{}, error) {
		return sdktimezone.List()
	})

	// ── iptables ──────────────────────────────────────────────────────
	r.Register("iptables.list", func(args map[string]interface{}) (interface{}, error) {
		chain := getStringArg(args, "chain", "")
		return sdkiptables.List(chain)
	})
	r.Register("iptables.flush", func(args map[string]interface{}) (interface{}, error) {
		table := getStringArg(args, "table", "")
		return sdkiptables.Flush(table)
	})
	r.Register("iptables.add_rule", func(args map[string]interface{}) (interface{}, error) {
		chain, err := argString(args, "chain")
		if err != nil {
			return nil, fmt.Errorf("iptables.add_rule: %w", err)
		}
		ruleSpec, err := argString(args, "rule_spec")
		if err != nil {
			return nil, fmt.Errorf("iptables.add_rule: %w", err)
		}
		return sdkiptables.AddRule(chain, ruleSpec)
	})
	r.Register("iptables.delete_rule", func(args map[string]interface{}) (interface{}, error) {
		chain, err := argString(args, "chain")
		if err != nil {
			return nil, fmt.Errorf("iptables.delete_rule: %w", err)
		}
		num, err := argInt(args, "number")
		if err != nil {
			return nil, fmt.Errorf("iptables.delete_rule: %w", err)
		}
		return sdkiptables.DeleteRule(chain, num)
	})
	r.Register("iptables.save", func(args map[string]interface{}) (interface{}, error) {
		return sdkiptables.Save()
	})
	r.Register("iptables.list_chains", func(args map[string]interface{}) (interface{}, error) {
		return sdkiptables.ListChains()
	})

	// ── npm ───────────────────────────────────────────────────────────
	r.Register("npm.list", func(args map[string]interface{}) (interface{}, error) {
		global, _ := argBool(args, "global")
		return sdknpm.List(global)
	})
	r.Register("npm.install", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("npm.install: %w", err)
		}
		global, _ := argBool(args, "global")
		return sdknpm.Install(name, global)
	})
	r.Register("npm.uninstall", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("npm.uninstall: %w", err)
		}
		global, _ := argBool(args, "global")
		return sdknpm.Uninstall(name, global)
	})
	r.Register("npm.outdated", func(args map[string]interface{}) (interface{}, error) {
		global, _ := argBool(args, "global")
		return sdknpm.Outdated(global)
	})

	// ── mysql ─────────────────────────────────────────────────────────
	r.Register("mysql.databases", func(args map[string]interface{}) (interface{}, error) {
		return sdkmysql.Databases()
	})
	r.Register("mysql.create_database", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("mysql.create_database: %w", err)
		}
		return sdkmysql.CreateDatabase(name)
	})
	r.Register("mysql.drop_database", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("mysql.drop_database: %w", err)
		}
		return sdkmysql.DropDatabase(name)
	})
	r.Register("mysql.users", func(args map[string]interface{}) (interface{}, error) {
		return sdkmysql.Users()
	})
	r.Register("mysql.create_user", func(args map[string]interface{}) (interface{}, error) {
		user, err := argString(args, "user")
		if err != nil {
			return nil, fmt.Errorf("mysql.create_user: %w", err)
		}
		host, err := argString(args, "host")
		if err != nil {
			return nil, fmt.Errorf("mysql.create_user: %w", err)
		}
		password, err := argString(args, "password")
		if err != nil {
			return nil, fmt.Errorf("mysql.create_user: %w", err)
		}
		return sdkmysql.CreateUser(user, host, password)
	})
	r.Register("mysql.drop_user", func(args map[string]interface{}) (interface{}, error) {
		user, err := argString(args, "user")
		if err != nil {
			return nil, fmt.Errorf("mysql.drop_user: %w", err)
		}
		host, err := argString(args, "host")
		if err != nil {
			return nil, fmt.Errorf("mysql.drop_user: %w", err)
		}
		return sdkmysql.DropUser(user, host)
	})
	r.Register("mysql.grant", func(args map[string]interface{}) (interface{}, error) {
		privileges, err := argString(args, "privileges")
		if err != nil {
			return nil, fmt.Errorf("mysql.grant: %w", err)
		}
		database, err := argString(args, "database")
		if err != nil {
			return nil, fmt.Errorf("mysql.grant: %w", err)
		}
		user, err := argString(args, "user")
		if err != nil {
			return nil, fmt.Errorf("mysql.grant: %w", err)
		}
		host, err := argString(args, "host")
		if err != nil {
			return nil, fmt.Errorf("mysql.grant: %w", err)
		}
		return sdkmysql.Grant(privileges, database, user, host)
	})

	// ── nginx ─────────────────────────────────────────────────────────
	r.Register("nginx.config_test", func(args map[string]interface{}) (interface{}, error) {
		return sdknginx.ConfigTest()
	})
	r.Register("nginx.reload", func(args map[string]interface{}) (interface{}, error) {
		return sdknginx.Reload()
	})
	r.Register("nginx.sites_list", func(args map[string]interface{}) (interface{}, error) {
		return sdknginx.SitesList()
	})
	r.Register("nginx.site_enable", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("nginx.site_enable: %w", err)
		}
		return sdknginx.SiteEnable(name)
	})
	r.Register("nginx.site_disable", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("nginx.site_disable: %w", err)
		}
		return sdknginx.SiteDisable(name)
	})

	// ── modprobe ──────────────────────────────────────────────────────
	r.Register("modprobe.list", func(args map[string]interface{}) (interface{}, error) {
		return sdkmodprobe.List()
	})
	r.Register("modprobe.load", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("modprobe.load: %w", err)
		}
		return sdkmodprobe.Load(name)
	})
	r.Register("modprobe.unload", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("modprobe.unload: %w", err)
		}
		return sdkmodprobe.Unload(name)
	})
	r.Register("modprobe.is_loaded", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("modprobe.is_loaded: %w", err)
		}
		return sdkmodprobe.IsLoaded(name)
	})

	// ── alternatives ──────────────────────────────────────────────────
	r.Register("alternatives.list", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("alternatives.list: %w", err)
		}
		return sdkalternatives.List(name)
	})
	r.Register("alternatives.display", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("alternatives.display: %w", err)
		}
		return sdkalternatives.Display(name)
	})
	r.Register("alternatives.set", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("alternatives.set: %w", err)
		}
		path, err := argString(args, "path")
		if err != nil {
			return nil, fmt.Errorf("alternatives.set: %w", err)
		}
		return sdkalternatives.Set(name, path)
	})
	r.Register("alternatives.install", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("alternatives.install: %w", err)
		}
		link, err := argString(args, "link")
		if err != nil {
			return nil, fmt.Errorf("alternatives.install: %w", err)
		}
		path, err := argString(args, "path")
		if err != nil {
			return nil, fmt.Errorf("alternatives.install: %w", err)
		}
		priority, err := argInt(args, "priority")
		if err != nil {
			return nil, fmt.Errorf("alternatives.install: %w", err)
		}
		return sdkalternatives.Install(name, link, path, priority)
	})
	r.Register("alternatives.remove", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("alternatives.remove: %w", err)
		}
		path, err := argString(args, "path")
		if err != nil {
			return nil, fmt.Errorf("alternatives.remove: %w", err)
		}
		return sdkalternatives.Remove(name, path)
	})

	// ── blockdev ──────────────────────────────────────────────────────
	r.Register("blockdev.list", func(args map[string]interface{}) (interface{}, error) {
		return sdkblockdev.List()
	})
	r.Register("blockdev.info", func(args map[string]interface{}) (interface{}, error) {
		device, err := argString(args, "device")
		if err != nil {
			return nil, fmt.Errorf("blockdev.info: %w", err)
		}
		return sdkblockdev.Info(device)
	})
	r.Register("blockdev.flush_buffers", func(args map[string]interface{}) (interface{}, error) {
		device, err := argString(args, "device")
		if err != nil {
			return nil, fmt.Errorf("blockdev.flush_buffers: %w", err)
		}
		return sdkblockdev.FlushBuffers(device)
	})
	r.Register("blockdev.set_readahead", func(args map[string]interface{}) (interface{}, error) {
		device, err := argString(args, "device")
		if err != nil {
			return nil, fmt.Errorf("blockdev.set_readahead: %w", err)
		}
		value, err := argInt(args, "value")
		if err != nil {
			return nil, fmt.Errorf("blockdev.set_readahead: %w", err)
		}
		return sdkblockdev.SetReadahead(device, value)
	})

	// ── at ────────────────────────────────────────────────────────────
	r.Register("at.list", func(args map[string]interface{}) (interface{}, error) {
		return sdkat.List()
	})
	r.Register("at.schedule", func(args map[string]interface{}) (interface{}, error) {
		command, err := argString(args, "command")
		if err != nil {
			return nil, fmt.Errorf("at.schedule: %w", err)
		}
		timeSpec, err := argString(args, "time_spec")
		if err != nil {
			return nil, fmt.Errorf("at.schedule: %w", err)
		}
		return sdkat.Schedule(command, timeSpec)
	})
	r.Register("at.remove", func(args map[string]interface{}) (interface{}, error) {
		jobID, err := argString(args, "job_id")
		if err != nil {
			return nil, fmt.Errorf("at.remove: %w", err)
		}
		return sdkat.Remove(jobID)
	})

	// ── postgresql ─────────────────────────────────────────────────────
	r.Register("postgresql.databases", func(args map[string]interface{}) (interface{}, error) {
		return sdkpostgresql.Databases()
	})
	r.Register("postgresql.create_database", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("postgresql.create_database: %w", err)
		}
		return sdkpostgresql.CreateDatabase(name)
	})
	r.Register("postgresql.drop_database", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("postgresql.drop_database: %w", err)
		}
		return sdkpostgresql.DropDatabase(name)
	})
	r.Register("postgresql.users", func(args map[string]interface{}) (interface{}, error) {
		return sdkpostgresql.Users()
	})
	r.Register("postgresql.create_user", func(args map[string]interface{}) (interface{}, error) {
		user, err := argString(args, "user")
		if err != nil {
			return nil, fmt.Errorf("postgresql.create_user: %w", err)
		}
		password, err := argString(args, "password")
		if err != nil {
			return nil, fmt.Errorf("postgresql.create_user: %w", err)
		}
		return sdkpostgresql.CreateUser(user, password)
	})
	r.Register("postgresql.drop_user", func(args map[string]interface{}) (interface{}, error) {
		user, err := argString(args, "user")
		if err != nil {
			return nil, fmt.Errorf("postgresql.drop_user: %w", err)
		}
		return sdkpostgresql.DropUser(user)
	})
	r.Register("postgresql.grant", func(args map[string]interface{}) (interface{}, error) {
		privileges, err := argString(args, "privileges")
		if err != nil {
			return nil, fmt.Errorf("postgresql.grant: %w", err)
		}
		database, err := argString(args, "database")
		if err != nil {
			return nil, fmt.Errorf("postgresql.grant: %w", err)
		}
		user, err := argString(args, "user")
		if err != nil {
			return nil, fmt.Errorf("postgresql.grant: %w", err)
		}
		return sdkpostgresql.Grant(privileges, database, user)
	})

	// ── apache2 ────────────────────────────────────────────────────────
	r.Register("apache2.config_test", func(args map[string]interface{}) (interface{}, error) {
		return sdkapache2.ConfigTest()
	})
	r.Register("apache2.reload", func(args map[string]interface{}) (interface{}, error) {
		return sdkapache2.Reload()
	})
	r.Register("apache2.sites_list", func(args map[string]interface{}) (interface{}, error) {
		return sdkapache2.SitesList()
	})
	r.Register("apache2.site_enable", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("apache2.site_enable: %w", err)
		}
		return sdkapache2.SiteEnable(name)
	})
	r.Register("apache2.site_disable", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("apache2.site_disable: %w", err)
		}
		return sdkapache2.SiteDisable(name)
	})
	r.Register("apache2.modules_list", func(args map[string]interface{}) (interface{}, error) {
		return sdkapache2.ModulesList()
	})
	r.Register("apache2.module_enable", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("apache2.module_enable: %w", err)
		}
		return sdkapache2.ModuleEnable(name)
	})
	r.Register("apache2.module_disable", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("apache2.module_disable: %w", err)
		}
		return sdkapache2.ModuleDisable(name)
	})

	// ── filesystem ─────────────────────────────────────────────────────
	r.Register("filesystem.mkfs", func(args map[string]interface{}) (interface{}, error) {
		device, err := argString(args, "device")
		if err != nil {
			return nil, fmt.Errorf("filesystem.mkfs: %w", err)
		}
		fsType, err := argString(args, "fstype")
		if err != nil {
			return nil, fmt.Errorf("filesystem.mkfs: %w", err)
		}
		label, _ := args["label"].(string)
		return sdkfilesystem.Mkfs(device, fsType, label)
	})
	r.Register("filesystem.resize_ext4", func(args map[string]interface{}) (interface{}, error) {
		device, err := argString(args, "device")
		if err != nil {
			return nil, fmt.Errorf("filesystem.resize_ext4: %w", err)
		}
		return sdkfilesystem.ResizeExt4(device)
	})
	r.Register("filesystem.resize_xfs", func(args map[string]interface{}) (interface{}, error) {
		mountpoint, err := argString(args, "mountpoint")
		if err != nil {
			return nil, fmt.Errorf("filesystem.resize_xfs: %w", err)
		}
		return sdkfilesystem.ResizeXFS(mountpoint)
	})
	r.Register("filesystem.check", func(args map[string]interface{}) (interface{}, error) {
		device, err := argString(args, "device")
		if err != nil {
			return nil, fmt.Errorf("filesystem.check: %w", err)
		}
		return sdkfilesystem.Check(device)
	})

	// ── parted ─────────────────────────────────────────────────────────
	r.Register("parted.list", func(args map[string]interface{}) (interface{}, error) {
		device, err := argString(args, "device")
		if err != nil {
			return nil, fmt.Errorf("parted.list: %w", err)
		}
		return sdkparted.List(device)
	})
	r.Register("parted.mklabel", func(args map[string]interface{}) (interface{}, error) {
		device, err := argString(args, "device")
		if err != nil {
			return nil, fmt.Errorf("parted.mklabel: %w", err)
		}
		labelType, _ := args["label_type"].(string)
		if labelType == "" {
			labelType = "gpt"
		}
		return sdkparted.MkLabel(device, labelType)
	})
	r.Register("parted.mkpart", func(args map[string]interface{}) (interface{}, error) {
		device, err := argString(args, "device")
		if err != nil {
			return nil, fmt.Errorf("parted.mkpart: %w", err)
		}
		partType, _ := args["part_type"].(string)
		fsType, _ := args["fstype"].(string)
		start, err := argString(args, "start")
		if err != nil {
			return nil, fmt.Errorf("parted.mkpart: %w", err)
		}
		end, err := argString(args, "end")
		if err != nil {
			return nil, fmt.Errorf("parted.mkpart: %w", err)
		}
		return sdkparted.MkPart(device, partType, fsType, start, end)
	})
	r.Register("parted.rm", func(args map[string]interface{}) (interface{}, error) {
		device, err := argString(args, "device")
		if err != nil {
			return nil, fmt.Errorf("parted.rm: %w", err)
		}
		number, err := argInt(args, "number")
		if err != nil {
			return nil, fmt.Errorf("parted.rm: %w", err)
		}
		return sdkparted.Rm(device, number)
	})

	// ── acl ────────────────────────────────────────────────────────────
	r.Register("acl.get", func(args map[string]interface{}) (interface{}, error) {
		path, err := argString(args, "path")
		if err != nil {
			return nil, fmt.Errorf("acl.get: %w", err)
		}
		return sdkacl.Get(path)
	})
	r.Register("acl.set", func(args map[string]interface{}) (interface{}, error) {
		path, err := argString(args, "path")
		if err != nil {
			return nil, fmt.Errorf("acl.set: %w", err)
		}
		entry, err := argString(args, "entry")
		if err != nil {
			return nil, fmt.Errorf("acl.set: %w", err)
		}
		recursive, _ := argBool(args, "recursive")
		return sdkacl.Set(path, entry, recursive)
	})
	r.Register("acl.remove", func(args map[string]interface{}) (interface{}, error) {
		path, err := argString(args, "path")
		if err != nil {
			return nil, fmt.Errorf("acl.remove: %w", err)
		}
		entry, err := argString(args, "entry")
		if err != nil {
			return nil, fmt.Errorf("acl.remove: %w", err)
		}
		recursive, _ := argBool(args, "recursive")
		return sdkacl.Remove(path, entry, recursive)
	})
	r.Register("acl.remove_all", func(args map[string]interface{}) (interface{}, error) {
		path, err := argString(args, "path")
		if err != nil {
			return nil, fmt.Errorf("acl.remove_all: %w", err)
		}
		recursive, _ := argBool(args, "recursive")
		return sdkacl.RemoveAll(path, recursive)
	})

	// ── wait_for ───────────────────────────────────────────────────────
	r.Register("wait_for.port", func(args map[string]interface{}) (interface{}, error) {
		host, err := argString(args, "host")
		if err != nil {
			return nil, fmt.Errorf("wait_for.port: %w", err)
		}
		port, err := argInt(args, "port")
		if err != nil {
			return nil, fmt.Errorf("wait_for.port: %w", err)
		}
		timeoutMs := 30000
		if v, ok := args["timeout_ms"]; ok {
			if t, e := argInt(map[string]interface{}{"timeout_ms": v}, "timeout_ms"); e == nil {
				timeoutMs = t
			}
		}
		return sdkwaitfor.Port(host, port, timeoutMs)
	})
	r.Register("wait_for.file", func(args map[string]interface{}) (interface{}, error) {
		path, err := argString(args, "path")
		if err != nil {
			return nil, fmt.Errorf("wait_for.file: %w", err)
		}
		timeoutMs := 30000
		if v, ok := args["timeout_ms"]; ok {
			if t, e := argInt(map[string]interface{}{"timeout_ms": v}, "timeout_ms"); e == nil {
				timeoutMs = t
			}
		}
		return sdkwaitfor.File(path, timeoutMs)
	})
	r.Register("wait_for.url", func(args map[string]interface{}) (interface{}, error) {
		url, err := argString(args, "url")
		if err != nil {
			return nil, fmt.Errorf("wait_for.url: %w", err)
		}
		timeoutMs := 30000
		if v, ok := args["timeout_ms"]; ok {
			if t, e := argInt(map[string]interface{}{"timeout_ms": v}, "timeout_ms"); e == nil {
				timeoutMs = t
			}
		}
		return sdkwaitfor.URL(url, timeoutMs)
	})

	// ── lvol ───────────────────────────────────────────────────────────
	r.Register("lvol.list", func(args map[string]interface{}) (interface{}, error) {
		return sdklvol.List()
	})
	r.Register("lvol.vg_list", func(args map[string]interface{}) (interface{}, error) {
		return sdklvol.VGList()
	})
	r.Register("lvol.create", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("lvol.create: %w", err)
		}
		vg, err := argString(args, "vg")
		if err != nil {
			return nil, fmt.Errorf("lvol.create: %w", err)
		}
		size, err := argString(args, "size")
		if err != nil {
			return nil, fmt.Errorf("lvol.create: %w", err)
		}
		return sdklvol.Create(name, vg, size)
	})
	r.Register("lvol.remove", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("lvol.remove: %w", err)
		}
		vg, err := argString(args, "vg")
		if err != nil {
			return nil, fmt.Errorf("lvol.remove: %w", err)
		}
		return sdklvol.Remove(name, vg)
	})
	r.Register("lvol.resize", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("lvol.resize: %w", err)
		}
		vg, err := argString(args, "vg")
		if err != nil {
			return nil, fmt.Errorf("lvol.resize: %w", err)
		}
		size, err := argString(args, "size")
		if err != nil {
			return nil, fmt.Errorf("lvol.resize: %w", err)
		}
		return sdklvol.Resize(name, vg, size)
	})

	// ── synchronize ────────────────────────────────────────────────────
	r.Register("synchronize.sync", func(args map[string]interface{}) (interface{}, error) {
		source, err := argString(args, "source")
		if err != nil {
			return nil, fmt.Errorf("synchronize.sync: %w", err)
		}
		dest, err := argString(args, "dest")
		if err != nil {
			return nil, fmt.Errorf("synchronize.sync: %w", err)
		}
		del, _ := argBool(args, "delete")
		compress, _ := argBool(args, "compress")
		return sdksync.Sync(source, dest, del, compress)
	})

	// ── fetch ──────────────────────────────────────────────────────────
	r.Register("fetch.file", func(args map[string]interface{}) (interface{}, error) {
		source, err := argString(args, "source")
		if err != nil {
			return nil, fmt.Errorf("fetch.file: %w", err)
		}
		dest, err := argString(args, "dest")
		if err != nil {
			return nil, fmt.Errorf("fetch.file: %w", err)
		}
		return sdkfetch.File(source, dest)
	})
	r.Register("fetch.url", func(args map[string]interface{}) (interface{}, error) {
		url, err := argString(args, "url")
		if err != nil {
			return nil, fmt.Errorf("fetch.url: %w", err)
		}
		dest, err := argString(args, "dest")
		if err != nil {
			return nil, fmt.Errorf("fetch.url: %w", err)
		}
		return sdkfetch.URL(url, dest)
	})

	// ── seboolean ──────────────────────────────────────────────────────
	r.Register("seboolean.list", func(args map[string]interface{}) (interface{}, error) {
		return sdksebool.List()
	})
	r.Register("seboolean.get", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("seboolean.get: %w", err)
		}
		return sdksebool.Get(name)
	})
	r.Register("seboolean.set", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("seboolean.set: %w", err)
		}
		state, err := argBool(args, "state")
		if err != nil {
			return nil, fmt.Errorf("seboolean.set: %w", err)
		}
		persistent, _ := argBool(args, "persistent")
		return sdksebool.Set(name, state, persistent)
	})

	// ── uri ──────────────────────────────────────────────────────────────
	r.Register("uri.do", func(args map[string]interface{}) (interface{}, error) {
		url, err := argString(args, "url")
		if err != nil {
			return nil, fmt.Errorf("uri.do: %w", err)
		}
		method := getStringArg(args, "method", "GET")
		headers := toStringMapArg(args, "headers")
		body := getStringArg(args, "body", "")
		timeoutMs, _ := argInt(args, "timeout_ms")
		if timeoutMs <= 0 {
			timeoutMs = 30000
		}
		return sdkuri.Do(url, method, headers, body, timeoutMs)
	})
	r.Register("uri.get", func(args map[string]interface{}) (interface{}, error) {
		url, err := argString(args, "url")
		if err != nil {
			return nil, fmt.Errorf("uri.get: %w", err)
		}
		return sdkuri.Get(url)
	})
	r.Register("uri.post", func(args map[string]interface{}) (interface{}, error) {
		url, err := argString(args, "url")
		if err != nil {
			return nil, fmt.Errorf("uri.post: %w", err)
		}
		return sdkuri.Post(url, args["body"])
	})
	r.Register("uri.put", func(args map[string]interface{}) (interface{}, error) {
		url, err := argString(args, "url")
		if err != nil {
			return nil, fmt.Errorf("uri.put: %w", err)
		}
		return sdkuri.Put(url, args["body"])
	})
	r.Register("uri.delete", func(args map[string]interface{}) (interface{}, error) {
		url, err := argString(args, "url")
		if err != nil {
			return nil, fmt.Errorf("uri.delete: %w", err)
		}
		return sdkuri.Delete(url)
	})
	r.Register("uri.download", func(args map[string]interface{}) (interface{}, error) {
		url, err := argString(args, "url")
		if err != nil {
			return nil, fmt.Errorf("uri.download: %w", err)
		}
		dest, err := argString(args, "dest")
		if err != nil {
			return nil, fmt.Errorf("uri.download: %w", err)
		}
		return sdkuri.Download(url, dest)
	})

	// ── lineinfile ───────────────────────────────────────────────────────
	r.Register("lineinfile.present", func(args map[string]interface{}) (interface{}, error) {
		path, err := argString(args, "path")
		if err != nil {
			return nil, fmt.Errorf("lineinfile.ensure: %w", err)
		}
		line, err := argString(args, "line")
		if err != nil {
			return nil, fmt.Errorf("lineinfile.ensure: %w", err)
		}
		re := getStringArg(args, "regexp", "")
		create, _ := argBool(args, "create")
		return sdklineinfile.Ensure(path, line, re, create)
	})
	r.Register("lineinfile.absent", func(args map[string]interface{}) (interface{}, error) {
		path, err := argString(args, "path")
		if err != nil {
			return nil, fmt.Errorf("lineinfile.absent: %w", err)
		}
		re, err := argString(args, "regexp")
		if err != nil {
			return nil, fmt.Errorf("lineinfile.absent: %w", err)
		}
		return sdklineinfile.Absent(path, re)
	})

	// ── replace ──────────────────────────────────────────────────────────
	r.Register("replace.replace", func(args map[string]interface{}) (interface{}, error) {
		path, err := argString(args, "path")
		if err != nil {
			return nil, fmt.Errorf("replace.replace: %w", err)
		}
		pattern, err := argString(args, "pattern")
		if err != nil {
			return nil, fmt.Errorf("replace.replace: %w", err)
		}
		replacement, err := argString(args, "replacement")
		if err != nil {
			return nil, fmt.Errorf("replace.replace: %w", err)
		}
		regexpMode, _ := argBool(args, "regexp_mode")
		return sdkreplace.Replace(path, pattern, replacement, regexpMode)
	})

	// ── xml ──────────────────────────────────────────────────────────────
	r.Register("xml.get_element", func(args map[string]interface{}) (interface{}, error) {
		path, err := argString(args, "path")
		if err != nil {
			return nil, fmt.Errorf("xml.get_element: %w", err)
		}
		element, err := argString(args, "element")
		if err != nil {
			return nil, fmt.Errorf("xml.get_element: %w", err)
		}
		return sdkxml.GetElement(path, element)
	})
	r.Register("xml.set_element", func(args map[string]interface{}) (interface{}, error) {
		path, err := argString(args, "path")
		if err != nil {
			return nil, fmt.Errorf("xml.set_element: %w", err)
		}
		element, err := argString(args, "element")
		if err != nil {
			return nil, fmt.Errorf("xml.set_element: %w", err)
		}
		value, err := argString(args, "value")
		if err != nil {
			return nil, fmt.Errorf("xml.set_element: %w", err)
		}
		return sdkxml.SetElement(path, element, value)
	})

	// ── systemd ─────────────────────────────────────────────────────────────
	r.Register("systemd.is_active", func(args map[string]interface{}) (interface{}, error) {
		unit, err := argString(args, "unit")
		if err != nil {
			return nil, fmt.Errorf("systemd.is_active: %w", err)
		}
		return sdksystemd.IsActive(unit)
	})
	r.Register("systemd.is_enabled", func(args map[string]interface{}) (interface{}, error) {
		unit, err := argString(args, "unit")
		if err != nil {
			return nil, fmt.Errorf("systemd.is_enabled: %w", err)
		}
		return sdksystemd.IsEnabled(unit)
	})
	r.Register("systemd.enable", func(args map[string]interface{}) (interface{}, error) {
		unit, err := argString(args, "unit")
		if err != nil {
			return nil, fmt.Errorf("systemd.enable: %w", err)
		}
		return sdksystemd.Enable(unit)
	})
	r.Register("systemd.disable", func(args map[string]interface{}) (interface{}, error) {
		unit, err := argString(args, "unit")
		if err != nil {
			return nil, fmt.Errorf("systemd.disable: %w", err)
		}
		return sdksystemd.Disable(unit)
	})
	r.Register("systemd.start", func(args map[string]interface{}) (interface{}, error) {
		unit, err := argString(args, "unit")
		if err != nil {
			return nil, fmt.Errorf("systemd.start: %w", err)
		}
		return sdksystemd.Start(unit)
	})
	r.Register("systemd.stop", func(args map[string]interface{}) (interface{}, error) {
		unit, err := argString(args, "unit")
		if err != nil {
			return nil, fmt.Errorf("systemd.stop: %w", err)
		}
		return sdksystemd.Stop(unit)
	})
	r.Register("systemd.restart", func(args map[string]interface{}) (interface{}, error) {
		unit, err := argString(args, "unit")
		if err != nil {
			return nil, fmt.Errorf("systemd.restart: %w", err)
		}
		return sdksystemd.Restart(unit)
	})
	r.Register("systemd.reload", func(args map[string]interface{}) (interface{}, error) {
		unit, err := argString(args, "unit")
		if err != nil {
			return nil, fmt.Errorf("systemd.reload: %w", err)
		}
		return sdksystemd.Reload(unit)
	})
	r.Register("systemd.daemon_reload", func(args map[string]interface{}) (interface{}, error) {
		return sdksystemd.DaemonReload()
	})
	r.Register("systemd.mask", func(args map[string]interface{}) (interface{}, error) {
		unit, err := argString(args, "unit")
		if err != nil {
			return nil, fmt.Errorf("systemd.mask: %w", err)
		}
		return sdksystemd.Mask(unit)
	})
	r.Register("systemd.unmask", func(args map[string]interface{}) (interface{}, error) {
		unit, err := argString(args, "unit")
		if err != nil {
			return nil, fmt.Errorf("systemd.unmask: %w", err)
		}
		return sdksystemd.Unmask(unit)
	})
	r.Register("systemd.show", func(args map[string]interface{}) (interface{}, error) {
		unit, err := argString(args, "unit")
		if err != nil {
			return nil, fmt.Errorf("systemd.show: %w", err)
		}
		return sdksystemd.Show(unit)
	})
	r.Register("systemd.list", func(args map[string]interface{}) (interface{}, error) {
		unitType := getStringArg(args, "unit_type", "")
		return sdksystemd.List(unitType)
	})

	// ── patch ───────────────────────────────────────────────────────────────
	r.Register("patch.apply", func(args map[string]interface{}) (interface{}, error) {
		patchContent, err := argString(args, "patch_content")
		if err != nil {
			return nil, fmt.Errorf("patch.apply: %w", err)
		}
		reverse, _ := argBool(args, "reverse")
		return sdkpatch.Apply(patchContent, reverse)
	})
	r.Register("patch.dry_run", func(args map[string]interface{}) (interface{}, error) {
		patchContent, err := argString(args, "patch_content")
		if err != nil {
			return nil, fmt.Errorf("patch.dry_run: %w", err)
		}
		return sdkpatch.DryRun(patchContent)
	})

	// ── xattr ───────────────────────────────────────────────────────────────
	r.Register("xattr.get", func(args map[string]interface{}) (interface{}, error) {
		path, err := argString(args, "path")
		if err != nil {
			return nil, fmt.Errorf("xattr.get: %w", err)
		}
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("xattr.get: %w", err)
		}
		return sdkxattr.Get(path, name)
	})
	r.Register("xattr.set", func(args map[string]interface{}) (interface{}, error) {
		path, err := argString(args, "path")
		if err != nil {
			return nil, fmt.Errorf("xattr.set: %w", err)
		}
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("xattr.set: %w", err)
		}
		value, err := argString(args, "value")
		if err != nil {
			return nil, fmt.Errorf("xattr.set: %w", err)
		}
		return sdkxattr.Set(path, name, value)
	})
	r.Register("xattr.remove", func(args map[string]interface{}) (interface{}, error) {
		path, err := argString(args, "path")
		if err != nil {
			return nil, fmt.Errorf("xattr.remove: %w", err)
		}
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("xattr.remove: %w", err)
		}
		return sdkxattr.Remove(path, name)
	})
	r.Register("xattr.list", func(args map[string]interface{}) (interface{}, error) {
		path, err := argString(args, "path")
		if err != nil {
			return nil, fmt.Errorf("xattr.list: %w", err)
		}
		return sdkxattr.List(path)
	})

	// ── firewalld_zone ──────────────────────────────────────────────────────
	r.Register("firewalld_zone.get_default", func(args map[string]interface{}) (interface{}, error) {
		return sdkfirewalldzone.GetDefaultZone()
	})
	r.Register("firewalld_zone.set_default", func(args map[string]interface{}) (interface{}, error) {
		zone, err := argString(args, "zone")
		if err != nil {
			return nil, fmt.Errorf("firewalld_zone.set_default: %w", err)
		}
		return sdkfirewalldzone.SetDefaultZone(zone)
	})
	r.Register("firewalld_zone.add_zone", func(args map[string]interface{}) (interface{}, error) {
		zone, err := argString(args, "zone")
		if err != nil {
			return nil, fmt.Errorf("firewalld_zone.add_zone: %w", err)
		}
		return sdkfirewalldzone.AddZone(zone)
	})
	r.Register("firewalld_zone.remove_zone", func(args map[string]interface{}) (interface{}, error) {
		zone, err := argString(args, "zone")
		if err != nil {
			return nil, fmt.Errorf("firewalld_zone.remove_zone: %w", err)
		}
		return sdkfirewalldzone.RemoveZone(zone)
	})
	r.Register("firewalld_zone.add_service", func(args map[string]interface{}) (interface{}, error) {
		zone, err := argString(args, "zone")
		if err != nil {
			return nil, fmt.Errorf("firewalld_zone.add_service: %w", err)
		}
		svc, err := argString(args, "service")
		if err != nil {
			return nil, fmt.Errorf("firewalld_zone.add_service: %w", err)
		}
		return sdkfirewalldzone.AddService(zone, svc)
	})
	r.Register("firewalld_zone.remove_service", func(args map[string]interface{}) (interface{}, error) {
		zone, err := argString(args, "zone")
		if err != nil {
			return nil, fmt.Errorf("firewalld_zone.remove_service: %w", err)
		}
		svc, err := argString(args, "service")
		if err != nil {
			return nil, fmt.Errorf("firewalld_zone.remove_service: %w", err)
		}
		return sdkfirewalldzone.RemoveService(zone, svc)
	})
	r.Register("firewalld_zone.add_port", func(args map[string]interface{}) (interface{}, error) {
		zone, err := argString(args, "zone")
		if err != nil {
			return nil, fmt.Errorf("firewalld_zone.add_port: %w", err)
		}
		pp, err := argString(args, "port_protocol")
		if err != nil {
			return nil, fmt.Errorf("firewalld_zone.add_port: %w", err)
		}
		return sdkfirewalldzone.AddPort(zone, pp)
	})
	r.Register("firewalld_zone.remove_port", func(args map[string]interface{}) (interface{}, error) {
		zone, err := argString(args, "zone")
		if err != nil {
			return nil, fmt.Errorf("firewalld_zone.remove_port: %w", err)
		}
		pp, err := argString(args, "port_protocol")
		if err != nil {
			return nil, fmt.Errorf("firewalld_zone.remove_port: %w", err)
		}
		return sdkfirewalldzone.RemovePort(zone, pp)
	})
	r.Register("firewalld_zone.add_rich_rule", func(args map[string]interface{}) (interface{}, error) {
		zone, err := argString(args, "zone")
		if err != nil {
			return nil, fmt.Errorf("firewalld_zone.add_rich_rule: %w", err)
		}
		rule, err := argString(args, "rule")
		if err != nil {
			return nil, fmt.Errorf("firewalld_zone.add_rich_rule: %w", err)
		}
		return sdkfirewalldzone.AddRichRule(zone, rule)
	})
	r.Register("firewalld_zone.remove_rich_rule", func(args map[string]interface{}) (interface{}, error) {
		zone, err := argString(args, "zone")
		if err != nil {
			return nil, fmt.Errorf("firewalld_zone.remove_rich_rule: %w", err)
		}
		rule, err := argString(args, "rule")
		if err != nil {
			return nil, fmt.Errorf("firewalld_zone.remove_rich_rule: %w", err)
		}
		return sdkfirewalldzone.RemoveRichRule(zone, rule)
	})
	r.Register("firewalld_zone.info", func(args map[string]interface{}) (interface{}, error) {
		zone, err := argString(args, "zone")
		if err != nil {
			return nil, fmt.Errorf("firewalld_zone.info: %w", err)
		}
		return sdkfirewalldzone.Info(zone)
	})
	r.Register("firewalld_zone.list_zones", func(args map[string]interface{}) (interface{}, error) {
		return sdkfirewalldzone.ListZones()
	})

	// ── get_url ─────────────────────────────────────────────────────────────
	r.Register("get_url.download", func(args map[string]interface{}) (interface{}, error) {
		url, err := argString(args, "url")
		if err != nil {
			return nil, fmt.Errorf("get_url.download: %w", err)
		}
		dest, err := argString(args, "dest")
		if err != nil {
			return nil, fmt.Errorf("get_url.download: %w", err)
		}
		checksum := getStringArg(args, "checksum", "")
		force, _ := argBool(args, "force")
		return sdkgeturl.Download(url, dest, checksum, force)
	})

	// ── sys utilities ───────────────────────────────────────────────────────
	r.Register("sys.uuid", func(args map[string]interface{}) (interface{}, error) {
		return sys.UUID()
	})
	r.Register("sys.random_password", func(args map[string]interface{}) (interface{}, error) {
		length := 16
		if v, ok := args["length"]; ok {
			switch n := v.(type) {
			case int:
				length = n
			case float64:
				length = int(n)
			default:
				if parsed, err := argInt(args, "length"); err == nil {
					length = parsed
				}
			}
		}
		useSpecial := getBoolArg(args, "use_special", true)
		useNumbers := getBoolArg(args, "use_numbers", true)
		useUppercase := getBoolArg(args, "use_uppercase", true)
		return sys.RandomPassword(length, useSpecial, useNumbers, useUppercase)
	})

	// ── sys mac_address ──────────────────────────────────────────────────────
	r.Register("sys.mac_address", func(args map[string]interface{}) (interface{}, error) {
		iface, _ := args["interface"].(string)
		return sys.MACAddress(iface)
	})
	r.Register("sys.mac_addresses", func(args map[string]interface{}) (interface{}, error) {
		return sys.MACAddresses()
	})
	r.Register("sys.dmidecode", func(args map[string]interface{}) (interface{}, error) {
		return sys.Dmidecode()
	})
	r.Register("sys.lspci", func(args map[string]interface{}) (interface{}, error) {
		return sys.LsPci()
	})
	r.Register("sys.lsblk", func(args map[string]interface{}) (interface{}, error) {
		return sys.LsBlk()
	})
	r.Register("sys.lsusb", func(args map[string]interface{}) (interface{}, error) {
		return sys.LsUsb()
	})
	r.Register("sys.ip_route", func(args map[string]interface{}) (interface{}, error) {
		return sys.IpRoute()
	})
	r.Register("sys.ethtool", func(args map[string]interface{}) (interface{}, error) {
		iface, _ := argString(args, "iface")
		return sys.Ethtool(iface)
	})

	// ── modprobe.set_boot ──────────────────────────────────────────────────────
	r.Register("modprobe.set_boot", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		present := true
		if v, ok := args["present"]; ok {
			if b, ok := v.(bool); ok {
				present = b
			}
		}
		return sdkmodprobe.SetBoot(name, present)
	})

	// ── seport ─────────────────────────────────────────────────────────────────
	r.Register("seport.add", func(args map[string]interface{}) (interface{}, error) {
		seportType, _ := argString(args, "seport_type")
		proto, _ := argString(args, "protocol")
		port, _ := argString(args, "port")
		return sdkseport.Add(seportType, proto, port)
	})
	r.Register("seport.remove", func(args map[string]interface{}) (interface{}, error) {
		proto, _ := argString(args, "protocol")
		port, _ := argString(args, "port")
		return sdkseport.Remove(proto, port)
	})
	r.Register("seport.list", func(args map[string]interface{}) (interface{}, error) {
		return sdkseport.List()
	})
	r.Register("seport.get", func(args map[string]interface{}) (interface{}, error) {
		proto, _ := argString(args, "protocol")
		port, _ := argString(args, "port")
		return sdkseport.Get(proto, port)
	})

	// ── sefcontext ─────────────────────────────────────────────────────────────
	r.Register("sefcontext.add", func(args map[string]interface{}) (interface{}, error) {
		filespec, _ := argString(args, "filespec")
		seType, _ := argString(args, "se_type")
		return sdksefcontext.Add(filespec, seType)
	})
	r.Register("sefcontext.modify", func(args map[string]interface{}) (interface{}, error) {
		filespec, _ := argString(args, "filespec")
		seType, _ := argString(args, "se_type")
		return sdksefcontext.Modify(filespec, seType)
	})
	r.Register("sefcontext.remove", func(args map[string]interface{}) (interface{}, error) {
		filespec, _ := argString(args, "filespec")
		return sdksefcontext.Remove(filespec)
	})
	r.Register("sefcontext.list", func(args map[string]interface{}) (interface{}, error) {
		return sdksefcontext.List()
	})
	r.Register("sefcontext.apply", func(args map[string]interface{}) (interface{}, error) {
		recursive := false
		if v, ok := args["recursive"]; ok {
			if b, ok := v.(bool); ok {
				recursive = b
			}
		}
		filespec, _ := argString(args, "filespec")
		return sdksefcontext.Apply(filespec, recursive)
	})

	// ── flatpak ─────────────────────────────────────────────────────────────
	r.Register("flatpak.install", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		from, _ := argString(args, "from")
		user, _ := argBool(args, "user")
		return sdkflatpak.Install(name, from, user)
	})
	r.Register("flatpak.remove", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		user, _ := argBool(args, "user")
		return sdkflatpak.Remove(name, user)
	})
	r.Register("flatpak.update", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		user, _ := argBool(args, "user")
		return sdkflatpak.Update(name, user)
	})
	r.Register("flatpak.list", func(args map[string]interface{}) (interface{}, error) {
		user, _ := argBool(args, "user")
		return sdkflatpak.List(user)
	})
	r.Register("flatpak.info", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		user, _ := argBool(args, "user")
		return sdkflatpak.Info(name, user)
	})
	r.Register("flatpak.run", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		var runArgs []string
		if argsList, ok := args["args"]; ok && argsList != nil {
			if list, ok := argsList.([]interface{}); ok {
				for _, item := range list {
					if s, ok := item.(string); ok {
						runArgs = append(runArgs, s)
					}
				}
			}
		}
		user, _ := argBool(args, "user")
		return sdkflatpak.Run(name, runArgs, user)
	})
	r.Register("flatpak.repair", func(args map[string]interface{}) (interface{}, error) {
		user, _ := argBool(args, "user")
		return sdkflatpak.Repair(user)
	})

	// ── zfs ─────────────────────────────────────────────────────────────
	r.Register("zfs.create", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		var props map[string]string
		if propsRaw, ok := args["properties"]; ok && propsRaw != nil {
			if m, ok := propsRaw.(map[string]interface{}); ok {
				props = make(map[string]string)
				for k, v := range m {
					if s, ok := v.(string); ok {
						props[k] = s
					}
				}
			}
		}
		return sdkzfs.Create(name, props)
	})
	r.Register("zfs.destroy", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		recursive, _ := argBool(args, "recursive")
		return sdkzfs.Destroy(name, recursive)
	})
	r.Register("zfs.set", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		property, _ := argString(args, "property")
		value, _ := argString(args, "value")
		return sdkzfs.Set(name, property, value)
	})
	r.Register("zfs.get", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		property, _ := argString(args, "property")
		return sdkzfs.Get(name, property)
	})
	r.Register("zfs.list", func(args map[string]interface{}) (interface{}, error) {
		return sdkzfs.List()
	})
	r.Register("zfs.exists", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		exists, err := sdkzfs.Exists(name)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"exists": exists}, nil
	})
	r.Register("zfs.list_pools", func(args map[string]interface{}) (interface{}, error) {
		return sdkzfs.ListPools()
	})
	r.Register("zfs.get_pool_status", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		return sdkzfs.GetPoolStatus(name)
	})
	r.Register("zfs.snapshot", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		snapName, _ := argString(args, "snapshot_name")
		return sdkzfs.Snapshot(name, snapName)
	})
	r.Register("zfs.destroy_snapshot", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		snapName, _ := argString(args, "snapshot_name")
		return sdkzfs.DestroySnapshot(name, snapName)
	})

	// ── nmcli ─────────────────────────────────────────────────────────────
	r.Register("nmcli.add", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		connType, _ := argString(args, "type")
		var settings map[string]string
		if settingsRaw, ok := args["settings"]; ok && settingsRaw != nil {
			if m, ok := settingsRaw.(map[string]interface{}); ok {
				settings = make(map[string]string)
				for k, v := range m {
					if s, ok := v.(string); ok {
						settings[k] = s
					}
				}
			}
		}
		return sdknmcli.Add(name, connType, settings)
	})
	r.Register("nmcli.modify", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		var settings map[string]string
		if settingsRaw, ok := args["settings"]; ok && settingsRaw != nil {
			if m, ok := settingsRaw.(map[string]interface{}); ok {
				settings = make(map[string]string)
				for k, v := range m {
					if s, ok := v.(string); ok {
						settings[k] = s
					}
				}
			}
		}
		return sdknmcli.Modify(name, settings)
	})
	r.Register("nmcli.delete", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		return sdknmcli.Delete(name)
	})
	r.Register("nmcli.up", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		return sdknmcli.Up(name)
	})
	r.Register("nmcli.down", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		return sdknmcli.Down(name)
	})
	r.Register("nmcli.list", func(args map[string]interface{}) (interface{}, error) {
		return sdknmcli.List()
	})
	r.Register("nmcli.show", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		return sdknmcli.Show(name)
	})
	r.Register("nmcli.list_devices", func(args map[string]interface{}) (interface{}, error) {
		return sdknmcli.ListDevices()
	})
	r.Register("nmcli.reload", func(args map[string]interface{}) (interface{}, error) {
		return sdknmcli.Reload()
	})
	r.Register("nmcli.get_general_status", func(args map[string]interface{}) (interface{}, error) {
		return sdknmcli.GetGeneralStatus()
	})

	// ── crypttab ──────────────────────────────────────────────────────────
	r.Register("crypttab.add", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		device, _ := argString(args, "device")
		keyFile, _ := argString(args, "key_file")
		options, _ := argString(args, "options")
		return sdkcrypttab.Add(name, device, keyFile, options)
	})
	r.Register("crypttab.remove", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		return sdkcrypttab.Remove(name)
	})
	r.Register("crypttab.modify", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		device, _ := argString(args, "device")
		keyFile, _ := argString(args, "key_file")
		options, _ := argString(args, "options")
		return sdkcrypttab.Modify(name, device, keyFile, options)
	})
	r.Register("crypttab.get", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		return sdkcrypttab.Get(name)
	})
	r.Register("crypttab.list", func(args map[string]interface{}) (interface{}, error) {
		return sdkcrypttab.List()
	})
	r.Register("crypttab.exists", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		exists, err := sdkcrypttab.Exists(name)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"exists": exists}, nil
	})
	r.Register("crypttab.validate", func(args map[string]interface{}) (interface{}, error) {
		return sdkcrypttab.Validate()
	})
	r.Register("crypttab.backup", func(args map[string]interface{}) (interface{}, error) {
		backupDir, _ := argString(args, "backup_dir")
		return sdkcrypttab.Backup(backupDir)
	})

	// ── sysfs ─────────────────────────────────────────────────────────────
	r.Register("sysfs.read", func(args map[string]interface{}) (interface{}, error) {
		path, _ := argString(args, "path")
		value, err := sdksysfs.Read(path)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"value": value}, nil
	})
	r.Register("sysfs.write", func(args map[string]interface{}) (interface{}, error) {
		path, _ := argString(args, "path")
		value, _ := argString(args, "value")
		return sdksysfs.Write(path, value)
	})
	r.Register("sysfs.exists", func(args map[string]interface{}) (interface{}, error) {
		path, _ := argString(args, "path")
		exists, err := sdksysfs.Exists(path)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"exists": exists}, nil
	})
	r.Register("sysfs.get", func(args map[string]interface{}) (interface{}, error) {
		path, _ := argString(args, "path")
		return sdksysfs.Get(path)
	})
	r.Register("sysfs.list", func(args map[string]interface{}) (interface{}, error) {
		dirPath, _ := argString(args, "dir_path")
		return sdksysfs.List(dirPath)
	})
	r.Register("sysfs.set_device_power", func(args map[string]interface{}) (interface{}, error) {
		devicePath, _ := argString(args, "device_path")
		state, _ := argString(args, "state")
		return sdksysfs.SetDevicePower(devicePath, state)
	})
	r.Register("sysfs.get_device_power", func(args map[string]interface{}) (interface{}, error) {
		devicePath, _ := argString(args, "device_path")
		state, err := sdksysfs.GetDevicePower(devicePath)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"state": state}, nil
	})
	r.Register("sysfs.set_kernel_parameter", func(args map[string]interface{}) (interface{}, error) {
		param, _ := argString(args, "param")
		value, _ := argString(args, "value")
		return sdksysfs.SetKernelParameter(param, value)
	})
	r.Register("sysfs.get_kernel_parameter", func(args map[string]interface{}) (interface{}, error) {
		param, _ := argString(args, "param")
		value, err := sdksysfs.GetKernelParameter(param)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"value": value}, nil
	})

	// ── pamd.* ──────────────────────────────────────────────────────────
	r.Register("pamd.get", func(args map[string]interface{}) (interface{}, error) {
		service, _ := argString(args, "service")
		return sdkpamd.Get(service)
	})
	r.Register("pamd.list", func(args map[string]interface{}) (interface{}, error) {
		return sdkpamd.List()
	})
	r.Register("pamd.add_rule", func(args map[string]interface{}) (interface{}, error) {
		service, _ := argString(args, "service")
		rtype, _ := argString(args, "type")
		control, _ := argString(args, "control")
		module, _ := argString(args, "module")
		a, _ := argString(args, "args")
		return sdkpamd.AddRule(service, rtype, control, module, a)
	})
	r.Register("pamd.remove_rule", func(args map[string]interface{}) (interface{}, error) {
		service, _ := argString(args, "service")
		rtype, _ := argString(args, "type")
		module, _ := argString(args, "module")
		return sdkpamd.RemoveRule(service, rtype, module)
	})
	r.Register("pamd.modify_rule", func(args map[string]interface{}) (interface{}, error) {
		service, _ := argString(args, "service")
		rtype, _ := argString(args, "type")
		module, _ := argString(args, "module")
		nc, _ := argString(args, "new_control")
		na, _ := argString(args, "new_args")
		return sdkpamd.ModifyRule(service, rtype, module, nc, na)
	})
	r.Register("pamd.validate", func(args map[string]interface{}) (interface{}, error) {
		service, _ := argString(args, "service")
		return sdkpamd.Validate(service)
	})
	r.Register("pamd.backup", func(args map[string]interface{}) (interface{}, error) {
		service, _ := argString(args, "service")
		dir, _ := argString(args, "backup_dir")
		return sdkpamd.Backup(service, dir)
	})

	// ── getent.* ────────────────────────────────────────────────────────
	r.Register("getent.passwd", func(args map[string]interface{}) (interface{}, error) {
		return sdkgetent.GetPasswd()
	})
	r.Register("getent.lookup_user", func(args map[string]interface{}) (interface{}, error) {
		key, _ := argString(args, "key")
		return sdkgetent.LookupUser(key)
	})
	r.Register("getent.groups", func(args map[string]interface{}) (interface{}, error) {
		return sdkgetent.GetGroups()
	})
	r.Register("getent.lookup_group", func(args map[string]interface{}) (interface{}, error) {
		key, _ := argString(args, "key")
		return sdkgetent.LookupGroup(key)
	})
	r.Register("getent.services", func(args map[string]interface{}) (interface{}, error) {
		return sdkgetent.GetServices()
	})
	r.Register("getent.lookup_service", func(args map[string]interface{}) (interface{}, error) {
		key, _ := argString(args, "key")
		return sdkgetent.LookupService(key)
	})
	r.Register("getent.protocols", func(args map[string]interface{}) (interface{}, error) {
		return sdkgetent.GetProtocols()
	})
	r.Register("getent.lookup_protocol", func(args map[string]interface{}) (interface{}, error) {
		key, _ := argString(args, "key")
		return sdkgetent.LookupProtocol(key)
	})
	r.Register("getent.shells", func(args map[string]interface{}) (interface{}, error) {
		return sdkgetent.Shells()
	})

	// ── haproxy.* ───────────────────────────────────────────────────────
	r.Register("haproxy.get_status", func(args map[string]interface{}) (interface{}, error) {
		return sdkhaproxy.GetStatus()
	})
	r.Register("haproxy.list_backends", func(args map[string]interface{}) (interface{}, error) {
		socket, _ := argString(args, "socket")
		return sdkhaproxy.ListBackends(socket)
	})
	r.Register("haproxy.enable_backend", func(args map[string]interface{}) (interface{}, error) {
		backend, _ := argString(args, "backend")
		server, _ := argString(args, "server")
		socket, _ := argString(args, "socket")
		return sdkhaproxy.EnableBackend(backend, server, socket)
	})
	r.Register("haproxy.disable_backend", func(args map[string]interface{}) (interface{}, error) {
		backend, _ := argString(args, "backend")
		server, _ := argString(args, "server")
		socket, _ := argString(args, "socket")
		return sdkhaproxy.DisableBackend(backend, server, socket)
	})
	r.Register("haproxy.validate_config", func(args map[string]interface{}) (interface{}, error) {
		configFile, _ := argString(args, "config_file")
		return sdkhaproxy.ValidateConfig(configFile)
	})
	r.Register("haproxy.reload", func(args map[string]interface{}) (interface{}, error) {
		configFile, _ := argString(args, "config_file")
		return sdkhaproxy.Reload(configFile)
	})
	r.Register("haproxy.restart", func(args map[string]interface{}) (interface{}, error) {
		return sdkhaproxy.Restart()
	})
	r.Register("haproxy.version", func(args map[string]interface{}) (interface{}, error) {
		return sdkhaproxy.Version()
	})

	// ── openssl_cert.* ──────────────────────────────────────────────────
	r.Register("openssl_cert.create_csr", func(args map[string]interface{}) (interface{}, error) {
		kp, _ := argString(args, "key_path")
		cp, _ := argString(args, "csr_path")
		subj, _ := argString(args, "subject")
		bits, _ := argInt(args, "key_bits")
		if bits <= 0 { bits = 2048 }
		return sdkopenssl.CreateCSR(kp, cp, subj, bits)
	})
	r.Register("openssl_cert.generate_self_signed", func(args map[string]interface{}) (interface{}, error) {
		cp, _ := argString(args, "cert_path")
		kp, _ := argString(args, "key_path")
		subj, _ := argString(args, "subject")
		days, _ := argInt(args, "days")
		bits, _ := argInt(args, "key_bits")
		return sdkopenssl.GenerateSelfSigned(cp, kp, subj, days, bits)
	})
	r.Register("openssl_cert.inspect", func(args map[string]interface{}) (interface{}, error) {
		cp, _ := argString(args, "cert_path")
		return sdkopenssl.Inspect(cp)
	})
	r.Register("openssl_cert.verify", func(args map[string]interface{}) (interface{}, error) {
		cp, _ := argString(args, "cert_path")
		ca, _ := argString(args, "ca_path")
		return sdkopenssl.Verify(cp, ca)
	})
	r.Register("openssl_cert.check_expiry", func(args map[string]interface{}) (interface{}, error) {
		cp, _ := argString(args, "cert_path")
		return sdkopenssl.CheckExpiry(cp)
	})
	r.Register("openssl_cert.convert_format", func(args map[string]interface{}) (interface{}, error) {
		ip, _ := argString(args, "input_path")
		op, _ := argString(args, "output_path")
		of, _ := argString(args, "output_format")
		return sdkopenssl.ConvertFormat(ip, op, of)
	})

	// ── redis.* ─────────────────────────────────────────────────────────
	r.Register("redis.ping", func(args map[string]interface{}) (interface{}, error) {
		h, _ := argString(args, "host")
		p, _ := argInt(args, "port")
		a, _ := argString(args, "auth")
		return sdkredis.Ping(h, p, a)
	})
	r.Register("redis.get", func(args map[string]interface{}) (interface{}, error) {
		key, _ := argString(args, "key")
		h, _ := argString(args, "host")
		p, _ := argInt(args, "port")
		a, _ := argString(args, "auth")
		return sdkredis.Get(key, h, p, a)
	})
	r.Register("redis.set", func(args map[string]interface{}) (interface{}, error) {
		key, _ := argString(args, "key")
		val, _ := argString(args, "value")
		h, _ := argString(args, "host")
		p, _ := argInt(args, "port")
		a, _ := argString(args, "auth")
		exp, _ := argInt(args, "expiry_sec")
		return sdkredis.Set(key, val, h, p, a, exp)
	})
	r.Register("redis.del", func(args map[string]interface{}) (interface{}, error) {
		var keys []string
		if raw, ok := args["keys"]; ok && raw != nil {
			if arr, ok := raw.([]interface{}); ok {
				for _, v := range arr {
					if s, ok := v.(string); ok {
						keys = append(keys, s)
					}
				}
			}
		}
		h, _ := argString(args, "host")
		p, _ := argInt(args, "port")
		a, _ := argString(args, "auth")
		return sdkredis.Del(keys, h, p, a)
	})
	r.Register("redis.keys", func(args map[string]interface{}) (interface{}, error) {
		pat, _ := argString(args, "pattern")
		h, _ := argString(args, "host")
		p, _ := argInt(args, "port")
		a, _ := argString(args, "auth")
		return sdkredis.Keys(pat, h, p, a)
	})
	r.Register("redis.info", func(args map[string]interface{}) (interface{}, error) {
		h, _ := argString(args, "host")
		p, _ := argInt(args, "port")
		a, _ := argString(args, "auth")
		return sdkredis.Info(h, p, a)
	})
	r.Register("redis.flush_db", func(args map[string]interface{}) (interface{}, error) {
		h, _ := argString(args, "host")
		p, _ := argInt(args, "port")
		a, _ := argString(args, "auth")
		return sdkredis.FlushDB(h, p, a)
	})

	// ── gem.* ───────────────────────────────────────────────────────────
	r.Register("gem.install", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		ver, _ := argString(args, "version")
		user, _ := argBool(args, "user_install")
		return sdkgem.Install(name, ver, user)
	})
	r.Register("gem.uninstall", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		force, _ := argBool(args, "force")
		return sdkgem.Uninstall(name, force)
	})
	r.Register("gem.update", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		return sdkgem.Update(name)
	})
	r.Register("gem.info", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		return sdkgem.Info(name)
	})
	r.Register("gem.list", func(args map[string]interface{}) (interface{}, error) {
		return sdkgem.List()
	})
	r.Register("gem.version", func(args map[string]interface{}) (interface{}, error) {
		return sdkgem.Version()
	})

	// ── rabbitmq ────────────────────────────────────────────────────────
	r.Register("rabbitmq.add_vhost", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		return sdkrabbitmq.AddVhost(name), nil
	})
	r.Register("rabbitmq.delete_vhost", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		return sdkrabbitmq.DeleteVhost(name), nil
	})
	r.Register("rabbitmq.list_vhosts", func(args map[string]interface{}) (interface{}, error) {
		return sdkrabbitmq.ListVhosts()
	})
	r.Register("rabbitmq.add_user", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		pass, _ := argString(args, "password")
		tags, _ := argString(args, "tags")
		return sdkrabbitmq.AddUser(name, pass, tags), nil
	})
	r.Register("rabbitmq.delete_user", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		return sdkrabbitmq.DeleteUser(name), nil
	})
	r.Register("rabbitmq.set_user_tags", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		tags, _ := argString(args, "tags")
		return sdkrabbitmq.SetUserTags(name, tags), nil
	})
	r.Register("rabbitmq.list_users", func(args map[string]interface{}) (interface{}, error) {
		return sdkrabbitmq.ListUsers()
	})
	r.Register("rabbitmq.set_permission", func(args map[string]interface{}) (interface{}, error) {
		user, _ := argString(args, "user")
		vhost, _ := argString(args, "vhost")
		configure, _ := argString(args, "configure")
		write, _ := argString(args, "write")
		read, _ := argString(args, "read")
		return sdkrabbitmq.SetPermission(user, vhost, configure, write, read), nil
	})
	r.Register("rabbitmq.clear_permission", func(args map[string]interface{}) (interface{}, error) {
		user, _ := argString(args, "user")
		vhost, _ := argString(args, "vhost")
		return sdkrabbitmq.ClearPermission(user, vhost), nil
	})
	r.Register("rabbitmq.set_policy", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		vhost, _ := argString(args, "vhost")
		pattern, _ := argString(args, "pattern")
		definition, _ := argString(args, "definition")
		applyTo, _ := argString(args, "apply_to")
		return sdkrabbitmq.SetPolicy(name, vhost, pattern, definition, applyTo), nil
	})
	r.Register("rabbitmq.delete_policy", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		vhost, _ := argString(args, "vhost")
		return sdkrabbitmq.DeletePolicy(name, vhost), nil
	})
	r.Register("rabbitmq.declare_queue", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		vhost, _ := argString(args, "vhost")
		queueType, _ := argString(args, "queue_type")
		durable, _ := argBool(args, "durable")
		autoDelete, _ := argBool(args, "auto_delete")
		return sdkrabbitmq.DeclareQueue(name, vhost, queueType, durable, autoDelete), nil
	})
	r.Register("rabbitmq.delete_queue", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		vhost, _ := argString(args, "vhost")
		return sdkrabbitmq.DeleteQueue(name, vhost), nil
	})
	r.Register("rabbitmq.declare_exchange", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		vhost, _ := argString(args, "vhost")
		exType, _ := argString(args, "type")
		durable, _ := argBool(args, "durable")
		autoDelete, _ := argBool(args, "auto_delete")
		return sdkrabbitmq.DeclareExchange(name, vhost, exType, durable, autoDelete), nil
	})
	r.Register("rabbitmq.delete_exchange", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		vhost, _ := argString(args, "vhost")
		return sdkrabbitmq.DeleteExchange(name, vhost), nil
	})
	r.Register("rabbitmq.bind_queue", func(args map[string]interface{}) (interface{}, error) {
		queue, _ := argString(args, "queue")
		exchange, _ := argString(args, "exchange")
		vhost, _ := argString(args, "vhost")
		routingKey, _ := argString(args, "routing_key")
		return sdkrabbitmq.BindQueue(queue, exchange, vhost, routingKey), nil
	})
	r.Register("rabbitmq.unbind_queue", func(args map[string]interface{}) (interface{}, error) {
		queue, _ := argString(args, "queue")
		exchange, _ := argString(args, "exchange")
		vhost, _ := argString(args, "vhost")
		routingKey, _ := argString(args, "routing_key")
		return sdkrabbitmq.UnbindQueue(queue, exchange, vhost, routingKey), nil
	})
	r.Register("rabbitmq.get_status", func(args map[string]interface{}) (interface{}, error) {
		return sdkrabbitmq.GetStatus(), nil
	})

	// ── consul ──────────────────────────────────────────────────────────
	r.Register("consul.kv_get", func(args map[string]interface{}) (interface{}, error) {
		key, _ := argString(args, "key")
		addr, _ := argString(args, "addr")
		return sdkconsul.KVGet(key, addr), nil
	})
	r.Register("consul.kv_put", func(args map[string]interface{}) (interface{}, error) {
		key, _ := argString(args, "key")
		value, _ := argString(args, "value")
		addr, _ := argString(args, "addr")
		return sdkconsul.KVPut(key, value, addr), nil
	})
	r.Register("consul.kv_delete", func(args map[string]interface{}) (interface{}, error) {
		key, _ := argString(args, "key")
		addr, _ := argString(args, "addr")
		return sdkconsul.KVDelete(key, addr), nil
	})
	r.Register("consul.kv_list", func(args map[string]interface{}) (interface{}, error) {
		prefix, _ := argString(args, "prefix")
		addr, _ := argString(args, "addr")
		return sdkconsul.KVList(prefix, addr)
	})
	r.Register("consul.service_register", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		id, _ := argString(args, "id")
		addr, _ := argString(args, "addr")
		port, _ := argString(args, "port")
		consulAddr, _ := argString(args, "consul_addr")
		return sdkconsul.ServiceRegister(name, id, addr, port, consulAddr), nil
	})
	r.Register("consul.service_deregister", func(args map[string]interface{}) (interface{}, error) {
		id, _ := argString(args, "id")
		consulAddr, _ := argString(args, "consul_addr")
		return sdkconsul.ServiceDeregister(id, consulAddr), nil
	})
	r.Register("consul.members", func(args map[string]interface{}) (interface{}, error) {
		addr, _ := argString(args, "addr")
		return sdkconsul.Members(addr), nil
	})
	r.Register("consul.info", func(args map[string]interface{}) (interface{}, error) {
		addr, _ := argString(args, "addr")
		return sdkconsul.Info(addr), nil
	})
	r.Register("consul.health_check", func(args map[string]interface{}) (interface{}, error) {
		service, _ := argString(args, "service")
		addr, _ := argString(args, "addr")
		return sdkconsul.HealthCheck(service, addr), nil
	})
	r.Register("consul.version", func(args map[string]interface{}) (interface{}, error) {
		return sdkconsul.Version()
	})

	// ── memcached ───────────────────────────────────────────────────────
	r.Register("memcached.get", func(args map[string]interface{}) (interface{}, error) {
		key, _ := argString(args, "key")
		host, _ := argString(args, "host")
		port, _ := argInt(args, "port")
		return sdkmemcached.Get(key, host, port), nil
	})
	r.Register("memcached.set", func(args map[string]interface{}) (interface{}, error) {
		key, _ := argString(args, "key")
		value, _ := argString(args, "value")
		host, _ := argString(args, "host")
		port, _ := argInt(args, "port")
		expiry, _ := argInt(args, "expiry")
		return sdkmemcached.Set(key, value, host, port, expiry), nil
	})
	r.Register("memcached.delete", func(args map[string]interface{}) (interface{}, error) {
		key, _ := argString(args, "key")
		host, _ := argString(args, "host")
		port, _ := argInt(args, "port")
		return sdkmemcached.Delete(key, host, port), nil
	})
	r.Register("memcached.flush_all", func(args map[string]interface{}) (interface{}, error) {
		host, _ := argString(args, "host")
		port, _ := argInt(args, "port")
		return sdkmemcached.FlushAll(host, port), nil
	})
	r.Register("memcached.stats", func(args map[string]interface{}) (interface{}, error) {
		host, _ := argString(args, "host")
		port, _ := argInt(args, "port")
		return sdkmemcached.Stats(host, port), nil
	})
	r.Register("memcached.version", func(args map[string]interface{}) (interface{}, error) {
		host, _ := argString(args, "host")
		port, _ := argInt(args, "port")
		return sdkmemcached.Version(host, port), nil
	})

	// ── composer ───────────────────────────────────────────────────────
	r.Register("composer.install", func(args map[string]interface{}) (interface{}, error) {
		dir, _ := argString(args, "dir")
		noDev, _ := argBool(args, "no_dev")
		return sdkcomposer.Install(dir, noDev), nil
	})
	r.Register("composer.update", func(args map[string]interface{}) (interface{}, error) {
		dir, _ := argString(args, "dir")
		noDev, _ := argBool(args, "no_dev")
		return sdkcomposer.Update(dir, noDev), nil
	})
	r.Register("composer.require", func(args map[string]interface{}) (interface{}, error) {
		dir, _ := argString(args, "dir")
		pkg, _ := argString(args, "package")
		ver, _ := argString(args, "version")
		return sdkcomposer.Require(dir, pkg, ver), nil
	})
	r.Register("composer.remove", func(args map[string]interface{}) (interface{}, error) {
		dir, _ := argString(args, "dir")
		pkg, _ := argString(args, "package")
		return sdkcomposer.Remove(dir, pkg), nil
	})
	r.Register("composer.create_project", func(args map[string]interface{}) (interface{}, error) {
		dir, _ := argString(args, "dir")
		pkg, _ := argString(args, "package")
		ver, _ := argString(args, "version")
		return sdkcomposer.CreateProject(dir, pkg, ver), nil
	})
	r.Register("composer.global_install", func(args map[string]interface{}) (interface{}, error) {
		pkg, _ := argString(args, "package")
		ver, _ := argString(args, "version")
		return sdkcomposer.GlobalInstall(pkg, ver), nil
	})
	r.Register("composer.version", func(args map[string]interface{}) (interface{}, error) {
		return sdkcomposer.Version(), nil
	})

	// ── cargo ─────────────────────────────────────────────────────────
	r.Register("cargo.install", func(args map[string]interface{}) (interface{}, error) {
		pkg, _ := argString(args, "package")
		ver, _ := argString(args, "version")
		force, _ := argBool(args, "force")
		return sdkcargo.Install(pkg, ver, force), nil
	})
	r.Register("cargo.uninstall", func(args map[string]interface{}) (interface{}, error) {
		pkg, _ := argString(args, "package")
		return sdkcargo.Uninstall(pkg), nil
	})
	r.Register("cargo.update", func(args map[string]interface{}) (interface{}, error) {
		pkg, _ := argString(args, "package")
		return sdkcargo.Update(pkg), nil
	})
	r.Register("cargo.list", func(args map[string]interface{}) (interface{}, error) {
		return sdkcargo.List()
	})
	r.Register("cargo.build", func(args map[string]interface{}) (interface{}, error) {
		dir, _ := argString(args, "dir")
		release, _ := argBool(args, "release")
		return sdkcargo.Build(dir, release), nil
	})
	r.Register("cargo.test", func(args map[string]interface{}) (interface{}, error) {
		dir, _ := argString(args, "dir")
		return sdkcargo.Test(dir), nil
	})
	r.Register("cargo.version", func(args map[string]interface{}) (interface{}, error) {
		return sdkcargo.Version(), nil
	})

	// ── rpmkey ────────────────────────────────────────────────────────
	r.Register("rpmkey.import", func(args map[string]interface{}) (interface{}, error) {
		keyPath, _ := argString(args, "key_path")
		return sdkrpmkey.Import(keyPath), nil
	})
	r.Register("rpmkey.list", func(args map[string]interface{}) (interface{}, error) {
		return sdkrpmkey.List(), nil
	})
	r.Register("rpmkey.remove", func(args map[string]interface{}) (interface{}, error) {
		keyID, _ := argString(args, "key_id")
		return sdkrpmkey.Remove(keyID), nil
	})

	// ── aptkey ────────────────────────────────────────────────────────
	r.Register("aptkey.add", func(args map[string]interface{}) (interface{}, error) {
		url, _ := argString(args, "url")
		keyring, _ := argString(args, "keyring")
		return sdkaptkey.Add(url, keyring), nil
	})
	r.Register("aptkey.add_from_key", func(args map[string]interface{}) (interface{}, error) {
		path, _ := argString(args, "path")
		keyring, _ := argString(args, "keyring")
		return sdkaptkey.AddFromKey(path, keyring), nil
	})
	r.Register("aptkey.remove", func(args map[string]interface{}) (interface{}, error) {
		keyID, _ := argString(args, "key_id")
		keyring, _ := argString(args, "keyring")
		return sdkaptkey.Remove(keyID, keyring), nil
	})
	r.Register("aptkey.list", func(args map[string]interface{}) (interface{}, error) {
		return sdkaptkey.List(), nil
	})

	// ── dmidecode ─────────────────────────────────────────────────────
	r.Register("dmidecode.system", func(args map[string]interface{}) (interface{}, error) {
		return sdkdmidecode.System(), nil
	})
	r.Register("dmidecode.bios", func(args map[string]interface{}) (interface{}, error) {
		return sdkdmidecode.BIOS(), nil
	})
	r.Register("dmidecode.chassis", func(args map[string]interface{}) (interface{}, error) {
		return sdkdmidecode.Chassis(), nil
	})
	r.Register("dmidecode.processor", func(args map[string]interface{}) (interface{}, error) {
		return sdkdmidecode.Processor(), nil
	})
	r.Register("dmidecode.keyword", func(args map[string]interface{}) (interface{}, error) {
		keyword, _ := argString(args, "keyword")
		v, err := sdkdmidecode.Keyword(keyword)
		return map[string]interface{}{"value": v}, err
	})

	// ── tuned ─────────────────────────────────────────────────────────
	r.Register("tuned.set", func(args map[string]interface{}) (interface{}, error) {
		profile, _ := argString(args, "profile")
		return sdktuned.Set(profile), nil
	})
	r.Register("tuned.status", func(args map[string]interface{}) (interface{}, error) {
		return sdktuned.Status(), nil
	})
	r.Register("tuned.list", func(args map[string]interface{}) (interface{}, error) {
		return sdktuned.List(), nil
	})
	r.Register("tuned.off", func(args map[string]interface{}) (interface{}, error) {
		return sdktuned.Off(), nil
	})
	r.Register("tuned.profile", func(args map[string]interface{}) (interface{}, error) {
		p, err := sdktuned.Profile()
		return map[string]interface{}{"profile": p}, err
	})
	r.Register("tuned.verify", func(args map[string]interface{}) (interface{}, error) {
		return sdktuned.Verify(), nil
	})

	// ── supervisor ────────────────────────────────────────────────────
	r.Register("supervisor.start", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		return sdksupervisor.Start(name), nil
	})
	r.Register("supervisor.stop", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		return sdksupervisor.Stop(name), nil
	})
	r.Register("supervisor.restart", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		return sdksupervisor.Restart(name), nil
	})
	r.Register("supervisor.reload", func(args map[string]interface{}) (interface{}, error) {
		return sdksupervisor.Reload(), nil
	})
	r.Register("supervisor.status", func(args map[string]interface{}) (interface{}, error) {
		return sdksupervisor.Status(), nil
	})
	r.Register("supervisor.clear_log", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		return sdksupervisor.ClearLog(name), nil
	})
	r.Register("supervisor.reread", func(args map[string]interface{}) (interface{}, error) {
		return sdksupervisor.Reread(), nil
	})
	r.Register("supervisor.update", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		return sdksupervisor.Update(name), nil
	})

	// ── pip.freeze / pip.install_requirements ──────────────────────────
	r.Register("pip.freeze", func(args map[string]interface{}) (interface{}, error) {
		return sdkpip.Freeze()
	})
	r.Register("pip.install_requirements", func(args map[string]interface{}) (interface{}, error) {
		req, _ := argString(args, "requirements")
		return sdkpip.InstallRequirements(req)
	})

	// ── flatpak.add_remote ─────────────────────────────────────────────
	r.Register("flatpak.add_remote", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		url, _ := argString(args, "url")
		return sdkflatpak.AddRemote(name, url)
	})

	// ── yarn ───────────────────────────────────────────────────────────
	r.Register("yarn.install", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		ver, _ := argString(args, "version")
		global, _ := argBool(args, "global")
		return sdkyarn.Install(name, ver, global), nil
	})
	r.Register("yarn.remove", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		global, _ := argBool(args, "global")
		return sdkyarn.Remove(name, global), nil
	})
	r.Register("yarn.global", func(args map[string]interface{}) (interface{}, error) {
		dir, _ := argString(args, "directory")
		return sdkyarn.Global(dir), nil
	})
	r.Register("yarn.list", func(args map[string]interface{}) (interface{}, error) {
		global, _ := argBool(args, "global")
		return sdkyarn.List(global), nil
	})

	// ── htpasswd ───────────────────────────────────────────────────────
	r.Register("htpasswd.set", func(args map[string]interface{}) (interface{}, error) {
		path, _ := argString(args, "path")
		username, _ := argString(args, "username")
		password, _ := argString(args, "password")
		create, _ := argBool(args, "create")
		return sdkhtpasswd.Set(path, username, password, create), nil
	})
	r.Register("htpasswd.remove", func(args map[string]interface{}) (interface{}, error) {
		path, _ := argString(args, "path")
		username, _ := argString(args, "username")
		return sdkhtpasswd.Remove(path, username), nil
	})
	r.Register("htpasswd.info", func(args map[string]interface{}) (interface{}, error) {
		path, _ := argString(args, "path")
		return sdkhtpasswd.Info(path), nil
	})
	r.Register("htpasswd.hash_sha1", func(args map[string]interface{}) (interface{}, error) {
		password, _ := argString(args, "password")
		return sdkhtpasswd.HashSHA1(password), nil
	})

	// ── sudoers ────────────────────────────────────────────────────────
	r.Register("sudoers.set", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		user, _ := argString(args, "user")
		commands, _ := argString(args, "commands")
		nopasswd, _ := argBool(args, "nopasswd")
		dir, _ := argString(args, "sudoers_dir")
		return sdksudoers.Set(name, user, commands, nopasswd, dir), nil
	})
	r.Register("sudoers.remove", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		dir, _ := argString(args, "sudoers_dir")
		return sdksudoers.Remove(name, dir), nil
	})
	r.Register("sudoers.info", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		dir, _ := argString(args, "sudoers_dir")
		return sdksudoers.Info(name, dir), nil
	})

	// ── monit ──────────────────────────────────────────────────────────
	r.Register("monit.start", func(args map[string]interface{}) (interface{}, error) {
		svc, _ := argString(args, "service")
		return sdkmonit.Start(svc), nil
	})
	r.Register("monit.stop", func(args map[string]interface{}) (interface{}, error) {
		svc, _ := argString(args, "service")
		return sdkmonit.Stop(svc), nil
	})
	r.Register("monit.monitor", func(args map[string]interface{}) (interface{}, error) {
		svc, _ := argString(args, "service")
		return sdkmonit.Monitor(svc), nil
	})
	r.Register("monit.unmonitor", func(args map[string]interface{}) (interface{}, error) {
		svc, _ := argString(args, "service")
		return sdkmonit.Unmonitor(svc), nil
	})
	r.Register("monit.restart", func(args map[string]interface{}) (interface{}, error) {
		svc, _ := argString(args, "service")
		return sdkmonit.Restart(svc), nil
	})
	r.Register("monit.status", func(args map[string]interface{}) (interface{}, error) {
		return sdkmonit.Status(), nil
	})
	r.Register("monit.reload", func(args map[string]interface{}) (interface{}, error) {
		return sdkmonit.Reload(), nil
	})

	// ── smartctl ──────────────────────────────────────────────────────────
	r.Register("smartctl.device", func(args map[string]interface{}) (interface{}, error) {
		device, _ := argString(args, "device")
		return sdksmartctl.Device(device), nil
	})
	r.Register("smartctl.health", func(args map[string]interface{}) (interface{}, error) {
		device, _ := argString(args, "device")
		return sdksmartctl.Health(device), nil
	})
	r.Register("smartctl.attributes", func(args map[string]interface{}) (interface{}, error) {
		device, _ := argString(args, "device")
		return sdksmartctl.Attributes(device), nil
	})
	r.Register("smartctl.list", func(args map[string]interface{}) (interface{}, error) {
		return sdksmartctl.List(), nil
	})
	r.Register("smartctl.json", func(args map[string]interface{}) (interface{}, error) {
		device, _ := argString(args, "device")
		v, err := sdksmartctl.JSON(device)
		return map[string]interface{}{"output": v}, err
	})

	// ── virsh ─────────────────────────────────────────────────────────────
	r.Register("virsh.start", func(args map[string]interface{}) (interface{}, error) {
		domain, _ := argString(args, "domain")
		return sdkvirsh.Start(domain), nil
	})
	r.Register("virsh.stop", func(args map[string]interface{}) (interface{}, error) {
		domain, _ := argString(args, "domain")
		return sdkvirsh.Stop(domain), nil
	})
	r.Register("virsh.reboot", func(args map[string]interface{}) (interface{}, error) {
		domain, _ := argString(args, "domain")
		return sdkvirsh.Reboot(domain), nil
	})
	r.Register("virsh.shutdown", func(args map[string]interface{}) (interface{}, error) {
		domain, _ := argString(args, "domain")
		return sdkvirsh.Shutdown(domain), nil
	})
	r.Register("virsh.suspend", func(args map[string]interface{}) (interface{}, error) {
		domain, _ := argString(args, "domain")
		return sdkvirsh.Suspend(domain), nil
	})
	r.Register("virsh.resume", func(args map[string]interface{}) (interface{}, error) {
		domain, _ := argString(args, "domain")
		return sdkvirsh.Resume(domain), nil
	})
	r.Register("virsh.list", func(args map[string]interface{}) (interface{}, error) {
		return sdkvirsh.List(), nil
	})
	r.Register("virsh.info", func(args map[string]interface{}) (interface{}, error) {
		domain, _ := argString(args, "domain")
		return sdkvirsh.Info(domain)
	})
	r.Register("virsh.version", func(args map[string]interface{}) (interface{}, error) {
		v, err := sdkvirsh.Version()
		return map[string]interface{}{"version": v}, err
	})

	// ── ethtool ───────────────────────────────────────────────────────────
	r.Register("ethtool.show", func(args map[string]interface{}) (interface{}, error) {
		iface, _ := argString(args, "interface")
		return sdkethtool.Show(iface), nil
	})
	r.Register("ethtool.set_speed", func(args map[string]interface{}) (interface{}, error) {
		iface, _ := argString(args, "interface")
		speed, _ := argString(args, "speed")
		return sdkethtool.SetSpeed(iface, speed), nil
	})
	r.Register("ethtool.set_duplex", func(args map[string]interface{}) (interface{}, error) {
		iface, _ := argString(args, "interface")
		duplex, _ := argString(args, "duplex")
		return sdkethtool.SetDuplex(iface, duplex), nil
	})
	r.Register("ethtool.set_autoneg", func(args map[string]interface{}) (interface{}, error) {
		iface, _ := argString(args, "interface")
		autoneg, _ := argString(args, "autoneg")
		return sdkethtool.SetAutoneg(iface, autoneg), nil
	})
	r.Register("ethtool.set_pause", func(args map[string]interface{}) (interface{}, error) {
		iface, _ := argString(args, "interface")
		rx, _ := argString(args, "rx")
		tx, _ := argString(args, "tx")
		return sdkethtool.SetPause(iface, rx, tx), nil
	})
	r.Register("ethtool.set_offload", func(args map[string]interface{}) (interface{}, error) {
		iface, _ := argString(args, "interface")
		feature, _ := argString(args, "feature")
		value, _ := argString(args, "value")
		return sdkethtool.SetOffload(iface, feature, value), nil
	})

	// ── systemd_analyze ───────────────────────────────────────────────────
	r.Register("systemd_analyze.time", func(args map[string]interface{}) (interface{}, error) {
		return sdksystemd_analyze.Time(), nil
	})
	r.Register("systemd_analyze.blame", func(args map[string]interface{}) (interface{}, error) {
		return sdksystemd_analyze.Blame(), nil
	})
	r.Register("systemd_analyze.critical_chain", func(args map[string]interface{}) (interface{}, error) {
		return sdksystemd_analyze.CriticalChain(), nil
	})
	r.Register("systemd_analyze.security", func(args map[string]interface{}) (interface{}, error) {
		unit, _ := argString(args, "unit")
		v, err := sdksystemd_analyze.Security(unit)
		return map[string]interface{}{"output": v}, err
	})
	r.Register("systemd_analyze.verify", func(args map[string]interface{}) (interface{}, error) {
		unit, _ := argString(args, "unit")
		v, err := sdksystemd_analyze.Verify(unit)
		return map[string]interface{}{"output": v}, err
	})

	// ── nvme ──────────────────────────────────────────────────────────────
	r.Register("nvme.list", func(args map[string]interface{}) (interface{}, error) {
		return sdknvme.List(), nil
	})
	r.Register("nvme.smart_log", func(args map[string]interface{}) (interface{}, error) {
		device, _ := argString(args, "device")
		v, err := sdknvme.SmartLog(device)
		return map[string]interface{}{"output": v}, err
	})
	r.Register("nvme.firmware_log", func(args map[string]interface{}) (interface{}, error) {
		device, _ := argString(args, "device")
		v, err := sdknvme.FirmwareLog(device)
		return map[string]interface{}{"output": v}, err
	})
	r.Register("nvme.error_log", func(args map[string]interface{}) (interface{}, error) {
		device, _ := argString(args, "device")
		v, err := sdknvme.ErrorLog(device)
		return map[string]interface{}{"output": v}, err
	})
	r.Register("nvme.version", func(args map[string]interface{}) (interface{}, error) {
		v, err := sdknvme.Version()
		return map[string]interface{}{"version": v}, err
	})

	// ── lshw ──────────────────────────────────────────────────────────────
	r.Register("lshw.short", func(args map[string]interface{}) (interface{}, error) {
		return sdkslshw.Short(), nil
	})
	r.Register("lshw.class", func(args map[string]interface{}) (interface{}, error) {
		class, _ := argString(args, "class")
		v, err := sdkslshw.Class(class)
		return map[string]interface{}{"output": v}, err
	})
	r.Register("lshw.json", func(args map[string]interface{}) (interface{}, error) {
		v, err := sdkslshw.JSON()
		return map[string]interface{}{"output": v}, err
	})
	r.Register("lshw.system", func(args map[string]interface{}) (interface{}, error) {
		v, err := sdkslshw.System()
		return map[string]interface{}{"output": v}, err
	})
	r.Register("lshw.memory", func(args map[string]interface{}) (interface{}, error) {
		v, err := sdkslshw.Memory()
		return map[string]interface{}{"output": v}, err
	})
	r.Register("lshw.disk", func(args map[string]interface{}) (interface{}, error) {
		v, err := sdkslshw.Disk()
		return map[string]interface{}{"output": v}, err
	})
	r.Register("lshw.network", func(args map[string]interface{}) (interface{}, error) {
		v, err := sdkslshw.Network()
		return map[string]interface{}{"output": v}, err
	})

	// ── ipaddr ────────────────────────────────────────────────────────────
	r.Register("ipaddr.list", func(args map[string]interface{}) (interface{}, error) {
		return sdkipaddr.List(), nil
	})
	r.Register("ipaddr.list_interface", func(args map[string]interface{}) (interface{}, error) {
		iface, _ := argString(args, "interface")
		return sdkipaddr.ListInterface(iface), nil
	})
	r.Register("ipaddr.add", func(args map[string]interface{}) (interface{}, error) {
		addr, _ := argString(args, "address")
		iface, _ := argString(args, "interface")
		return sdkipaddr.Add(addr, iface), nil
	})
	r.Register("ipaddr.delete", func(args map[string]interface{}) (interface{}, error) {
		addr, _ := argString(args, "address")
		iface, _ := argString(args, "interface")
		return sdkipaddr.Delete(addr, iface), nil
	})
	r.Register("ipaddr.flush", func(args map[string]interface{}) (interface{}, error) {
		iface, _ := argString(args, "interface")
		return sdkipaddr.Flush(iface), nil
	})
	r.Register("ipaddr.links", func(args map[string]interface{}) (interface{}, error) {
		return sdkipaddr.Links(), nil
	})
	r.Register("ipaddr.link_up", func(args map[string]interface{}) (interface{}, error) {
		iface, _ := argString(args, "interface")
		return sdkipaddr.LinkUp(iface), nil
	})
	r.Register("ipaddr.link_down", func(args map[string]interface{}) (interface{}, error) {
		iface, _ := argString(args, "interface")
		return sdkipaddr.LinkDown(iface), nil
	})

	// ── udevadm ───────────────────────────────────────────────────────────
	r.Register("udevadm.control", func(args map[string]interface{}) (interface{}, error) {
		action, _ := argString(args, "action")
		return sdkudevadm.Control(action), nil
	})
	r.Register("udevadm.trigger", func(args map[string]interface{}) (interface{}, error) {
		subsystem, _ := argString(args, "subsystem")
		return sdkudevadm.Trigger(subsystem), nil
	})
	r.Register("udevadm.settle", func(args map[string]interface{}) (interface{}, error) {
		timeout, _ := argInt(args, "timeout")
		return sdkudevadm.Settle(timeout), nil
	})
	r.Register("udevadm.info", func(args map[string]interface{}) (interface{}, error) {
		query, _ := argString(args, "query")
		device, _ := argString(args, "device")
		return sdkudevadm.Info(query, device), nil
	})
	r.Register("udevadm.monitor", func(args map[string]interface{}) (interface{}, error) {
		return sdkudevadm.Monitor(), nil
	})

	// ── modinfo ───────────────────────────────────────────────────────────
	r.Register("modinfo.info", func(args map[string]interface{}) (interface{}, error) {
		module, _ := argString(args, "module")
		return sdkmodinfo.Info(module), nil
	})
	r.Register("modinfo.list", func(args map[string]interface{}) (interface{}, error) {
		return sdkmodinfo.List(), nil
	})
	r.Register("modinfo.version", func(args map[string]interface{}) (interface{}, error) {
		return sdkmodinfo.Version(), nil
	})

	// ── dconf ─────────────────────────────────────────────────────────────
	r.Register("dconf.read", func(args map[string]interface{}) (interface{}, error) {
		key, _ := argString(args, "key")
		return sdkdconf.Read(key), nil
	})
	r.Register("dconf.write", func(args map[string]interface{}) (interface{}, error) {
		key, _ := argString(args, "key")
		value, _ := argString(args, "value")
		return sdkdconf.Write(key, value), nil
	})
	r.Register("dconf.list", func(args map[string]interface{}) (interface{}, error) {
		dir, _ := argString(args, "dir")
		return sdkdconf.List(dir), nil
	})
	r.Register("dconf.reset", func(args map[string]interface{}) (interface{}, error) {
		key, _ := argString(args, "key")
		return sdkdconf.Reset(key), nil
	})

	// ── locale_gen ────────────────────────────────────────────────────────
	r.Register("locale_gen.generate", func(args map[string]interface{}) (interface{}, error) {
		locale, _ := argString(args, "locale")
		return sdklocale_gen.Generate(locale), nil
	})
	r.Register("locale_gen.list", func(args map[string]interface{}) (interface{}, error) {
		return sdklocale_gen.List(), nil
	})
	r.Register("locale_gen.remove", func(args map[string]interface{}) (interface{}, error) {
		locale, _ := argString(args, "locale")
		return sdklocale_gen.Remove(locale), nil
	})

	// ── pam_limits ────────────────────────────────────────────────────────
	r.Register("pam_limits.set", func(args map[string]interface{}) (interface{}, error) {
		domain, _ := argString(args, "domain")
		limitType, _ := argString(args, "type")
		item, _ := argString(args, "item")
		value, _ := argString(args, "value")
		return sdkpam_limits.Set(domain, limitType, item, value), nil
	})
	r.Register("pam_limits.list", func(args map[string]interface{}) (interface{}, error) {
		return sdkpam_limits.List(), nil
	})

	// ── motd ──────────────────────────────────────────────────────────────
	r.Register("motd.read", func(args map[string]interface{}) (interface{}, error) {
		return sdkmotd.Read(), nil
	})
	r.Register("motd.write", func(args map[string]interface{}) (interface{}, error) {
		content, _ := argString(args, "content")
		return sdkmotd.Write(content), nil
	})

	// ── issue ─────────────────────────────────────────────────────────────
	r.Register("issue.read", func(args map[string]interface{}) (interface{}, error) {
		return sdkissue.Read(), nil
	})
	r.Register("issue.write", func(args map[string]interface{}) (interface{}, error) {
		content, _ := argString(args, "content")
		return sdkissue.Write(content), nil
	})

	// ── authorized_key ──────────────────────────────────────────────────────
	r.Register("authorized_key.manage", func(args map[string]interface{}) (interface{}, error) {
		username, _ := argString(args, "username")
		key, _ := argString(args, "key")
		state, _ := argString(args, "state")
		path, _ := argString(args, "path")
		return sdkauthorized_key.Manage(username, key, state, path), nil
	})
	r.Register("authorized_key.list", func(args map[string]interface{}) (interface{}, error) {
		username, _ := argString(args, "username")
		path, _ := argString(args, "path")
		return sdkauthorized_key.List(username, path), nil
	})
	r.Register("authorized_key.check", func(args map[string]interface{}) (interface{}, error) {
		username, _ := argString(args, "username")
		key, _ := argString(args, "key")
		path, _ := argString(args, "path")
		return sdkauthorized_key.Check(username, key, path), nil
	})

	// ── blockinfile ─────────────────────────────────────────────────────────
	r.Register("blockinfile.manage", func(args map[string]interface{}) (interface{}, error) {
		path, _ := argString(args, "path")
		block, _ := argString(args, "block")
		state, _ := argString(args, "state")
		marker, _ := argString(args, "marker")
		insertAfter, _ := argString(args, "insert_after")
		insertBefore, _ := argString(args, "insert_before")
		return sdkblockinfile.Manage(path, block, state, marker, insertAfter, insertBefore), nil
	})
	r.Register("blockinfile.read", func(args map[string]interface{}) (interface{}, error) {
		path, _ := argString(args, "path")
		marker, _ := argString(args, "marker")
		content, found, err := sdkblockinfile.Read(path, marker)
		errStr := ""
		if err != nil {
			errStr = err.Error()
		}
		return map[string]interface{}{"content": content, "found": found, "error": errStr}, nil
	})

	// ── debconf ─────────────────────────────────────────────────────────────
	r.Register("debconf.set", func(args map[string]interface{}) (interface{}, error) {
		pkg, _ := argString(args, "package")
		name, _ := argString(args, "name")
		vtype, _ := argString(args, "vtype")
		value, _ := argString(args, "value")
		return sdkdebconf.Set(pkg, name, vtype, value), nil
	})
	r.Register("debconf.get", func(args map[string]interface{}) (interface{}, error) {
		pkg, _ := argString(args, "package")
		name, _ := argString(args, "name")
		return sdkdebconf.Get(pkg, name), nil
	})
	r.Register("debconf.list", func(args map[string]interface{}) (interface{}, error) {
		pkg, _ := argString(args, "package")
		return sdkdebconf.List(pkg), nil
	})

	// ── reboot ──────────────────────────────────────────────────────────────
	r.Register("reboot.request", func(args map[string]interface{}) (interface{}, error) {
		msg, _ := argString(args, "msg")
		delay, _ := argInt(args, "delay")
		return sdkreboot.Request(msg, delay), nil
	})
	r.Register("reboot.dry_run", func(args map[string]interface{}) (interface{}, error) {
		msg, _ := argString(args, "msg")
		delay, _ := argInt(args, "delay")
		return sdkreboot.DryRun(msg, delay), nil
	})
	r.Register("reboot.check", func(args map[string]interface{}) (interface{}, error) {
		return sdkreboot.Check(), nil
	})

	// ── swap ────────────────────────────────────────────────────────────────
	r.Register("swap.info", func(args map[string]interface{}) (interface{}, error) {
		return sdkswap.Info(), nil
	})
	r.Register("swap.create", func(args map[string]interface{}) (interface{}, error) {
		path, _ := argString(args, "path")
		sizeMB, _ := argInt(args, "size_mb")
		return sdkswap.Create(path, sizeMB), nil
	})
	r.Register("swap.enable", func(args map[string]interface{}) (interface{}, error) {
		device, _ := argString(args, "device")
		return sdkswap.Enable(device), nil
	})
	r.Register("swap.disable", func(args map[string]interface{}) (interface{}, error) {
		device, _ := argString(args, "device")
		return sdkswap.Disable(device), nil
	})

	// ── raw ─────────────────────────────────────────────────────────────────
	r.Register("raw.execute", func(args map[string]interface{}) (interface{}, error) {
		command, _ := argString(args, "command")
		timeout, _ := argInt(args, "timeout")
		return sdkraw.Execute(command, timeout), nil
	})
	r.Register("raw.execute_with_env", func(args map[string]interface{}) (interface{}, error) {
		command, _ := argString(args, "command")
		timeout, _ := argInt(args, "timeout")
		env := toStringMapArg(args, "env")
		return sdkraw.ExecuteWithEnv(command, timeout, env), nil
	})

	// ── expect ──────────────────────────────────────────────────────────────
	r.Register("expect.run", func(args map[string]interface{}) (interface{}, error) {
		command, _ := argString(args, "command")
		responses := toStringMapArg(args, "responses")
		timeout, _ := argInt(args, "timeout")
		return sdkexpect.Run(command, responses, timeout), nil
	})
	r.Register("expect.run_simple", func(args map[string]interface{}) (interface{}, error) {
		command, _ := argString(args, "command")
		prompt, _ := argString(args, "prompt")
		response, _ := argString(args, "response")
		timeout, _ := argInt(args, "timeout")
		return sdkexpect.RunSimple(command, prompt, response, timeout), nil
	})

	// ── slurp ───────────────────────────────────────────────────────────────
	r.Register("slurp.encode", func(args map[string]interface{}) (interface{}, error) {
		path, _ := argString(args, "path")
		return sdkslurp.Encode(path), nil
	})
	r.Register("slurp.decode", func(args map[string]interface{}) (interface{}, error) {
		encoded, _ := argString(args, "encoded")
		destPath, _ := argString(args, "dest_path")
		return sdkslurp.Decode(encoded, destPath), nil
	})

	// ── wait_for_connection ─────────────────────────────────────────────────
	r.Register("wait_for_connection.wait", func(args map[string]interface{}) (interface{}, error) {
		host, _ := argString(args, "host")
		port, _ := argInt(args, "port")
		timeout, _ := argInt(args, "timeout")
		delay, _ := argInt(args, "delay")
		return sdkwait_for_connection.Wait(host, port, timeout, delay), nil
	})
	r.Register("wait_for_connection.check_once", func(args map[string]interface{}) (interface{}, error) {
		host, _ := argString(args, "host")
		port, _ := argInt(args, "port")
		return sdkwait_for_connection.CheckOnce(host, port), nil
	})

	// ── firewalld_rich_rule ─────────────────────────────────────────────────
	r.Register("firewalld_rich_rule.add", func(args map[string]interface{}) (interface{}, error) {
		zone, _ := argString(args, "zone")
		rule, _ := argString(args, "rule")
		return sdkfirewalld_rich_rule.Add(zone, rule), nil
	})
	r.Register("firewalld_rich_rule.remove", func(args map[string]interface{}) (interface{}, error) {
		zone, _ := argString(args, "zone")
		rule, _ := argString(args, "rule")
		return sdkfirewalld_rich_rule.Remove(zone, rule), nil
	})
	r.Register("firewalld_rich_rule.list", func(args map[string]interface{}) (interface{}, error) {
		zone, _ := argString(args, "zone")
		return sdkfirewalld_rich_rule.List(zone), nil
	})
	r.Register("firewalld_rich_rule.exists", func(args map[string]interface{}) (interface{}, error) {
		zone, _ := argString(args, "zone")
		rule, _ := argString(args, "rule")
		return sdkfirewalld_rich_rule.Exists(zone, rule), nil
	})

	// ── firewalld_ipset ─────────────────────────────────────────────────────
	r.Register("firewalld_ipset.create", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		setType, _ := argString(args, "type")
		return sdkfirewalld_ipset.Create(name, setType), nil
	})
	r.Register("firewalld_ipset.delete", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		return sdkfirewalld_ipset.Delete(name), nil
	})
	r.Register("firewalld_ipset.add_entry", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		entry, _ := argString(args, "entry")
		return sdkfirewalld_ipset.AddEntry(name, entry), nil
	})
	r.Register("firewalld_ipset.remove_entry", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		entry, _ := argString(args, "entry")
		return sdkfirewalld_ipset.RemoveEntry(name, entry), nil
	})
	r.Register("firewalld_ipset.list", func(args map[string]interface{}) (interface{}, error) {
		return sdkfirewalld_ipset.List(), nil
	})
	r.Register("firewalld_ipset.info", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		return sdkfirewalld_ipset.Info(name), nil
	})

	// ── pause ───────────────────────────────────────────────────────────────
	r.Register("pause.seconds", func(args map[string]interface{}) (interface{}, error) {
		duration, _ := argInt(args, "duration")
		return sdkpause.Seconds(duration), nil
	})
	r.Register("pause.prompt", func(args map[string]interface{}) (interface{}, error) {
		message, _ := argString(args, "message")
		return sdkpause.Prompt(message), nil
	})
	r.Register("pause.prompt_with_default", func(args map[string]interface{}) (interface{}, error) {
		message, _ := argString(args, "message")
		defaultVal, _ := argString(args, "default")
		return sdkpause.PromptWithDefault(message, defaultVal), nil
	})

	// ── meta ────────────────────────────────────────────────────────────────
	r.Register("meta.end_host", func(args map[string]interface{}) (interface{}, error) {
		return sdkmeta.EndHost(), nil
	})
	r.Register("meta.end_play", func(args map[string]interface{}) (interface{}, error) {
		return sdkmeta.EndPlay(), nil
	})
	r.Register("meta.clear_host_errors", func(args map[string]interface{}) (interface{}, error) {
		return sdkmeta.ClearHostErrors(), nil
	})
	r.Register("meta.refresh_inventory", func(args map[string]interface{}) (interface{}, error) {
		return sdkmeta.RefreshInventory(), nil
	})
	r.Register("meta.flush_handlers", func(args map[string]interface{}) (interface{}, error) {
		return sdkmeta.FlushHandlers(), nil
	})
	r.Register("meta.reset_connection", func(args map[string]interface{}) (interface{}, error) {
		return sdkmeta.ResetConnection(), nil
	})
	r.Register("meta.noop", func(args map[string]interface{}) (interface{}, error) {
		return sdkmeta.Noop(), nil
	})
	r.Register("meta.fail", func(args map[string]interface{}) (interface{}, error) {
		message, _ := argString(args, "message")
		return sdkmeta.Fail(message), nil
	})
	r.Register("meta.assert", func(args map[string]interface{}) (interface{}, error) {
		condition, _ := args["condition"].(bool)
		message, _ := argString(args, "message")
		return sdkmeta.Assert(condition, message), nil
	})
	r.Register("meta.debug", func(args map[string]interface{}) (interface{}, error) {
		message, _ := argString(args, "message")
		return sdkmeta.Debug(message, nil), nil
	})

	// ── uri_ext ─────────────────────────────────────────────────────────────
	r.Register("uri_ext.patch", func(args map[string]interface{}) (interface{}, error) {
		url, _ := argString(args, "url")
		body, _ := args["body"].([]byte)
		headers := toStringMapArg(args, "headers")
		timeout, _ := argInt(args, "timeout")
		return sdkuri_ext.Patch(url, body, headers, timeout), nil
	})
	r.Register("uri_ext.delete", func(args map[string]interface{}) (interface{}, error) {
		url, _ := argString(args, "url")
		headers := toStringMapArg(args, "headers")
		timeout, _ := argInt(args, "timeout")
		return sdkuri_ext.Delete(url, headers, timeout), nil
	})
	r.Register("uri_ext.head", func(args map[string]interface{}) (interface{}, error) {
		url, _ := argString(args, "url")
		headers := toStringMapArg(args, "headers")
		timeout, _ := argInt(args, "timeout")
		return sdkuri_ext.Head(url, headers, timeout), nil
	})
	r.Register("uri_ext.options", func(args map[string]interface{}) (interface{}, error) {
		url, _ := argString(args, "url")
		headers := toStringMapArg(args, "headers")
		timeout, _ := argInt(args, "timeout")
		return sdkuri_ext.Options(url, headers, timeout), nil
	})

	// ── hwclock ─────────────────────────────────────────────────────────────
	r.Register("hwclock.get", func(args map[string]interface{}) (interface{}, error) {
		return sdkhwclock.Get(), nil
	})
	r.Register("hwclock.set", func(args map[string]interface{}) (interface{}, error) {
		return sdkhwclock.Set(), nil
	})
	r.Register("hwclock.hctosys", func(args map[string]interface{}) (interface{}, error) {
		return sdkhwclock.HCToSys(), nil
	})
	r.Register("hwclock.set_time", func(args map[string]interface{}) (interface{}, error) {
		timeStr, _ := argString(args, "time")
		return sdkhwclock.SetTime(timeStr), nil
	})

	// ── mdadm ───────────────────────────────────────────────────────────────
	r.Register("mdadm.create", func(args map[string]interface{}) (interface{}, error) {
		device, _ := argString(args, "device")
		level, _ := argString(args, "level")
		devices := argStringSlice(args, "devices")
		return sdkmdadm.Create(device, level, devices), nil
	})
	r.Register("mdadm.destroy", func(args map[string]interface{}) (interface{}, error) {
		device, _ := argString(args, "device")
		return sdkmdadm.Destroy(device), nil
	})
	r.Register("mdadm.detail", func(args map[string]interface{}) (interface{}, error) {
		device, _ := argString(args, "device")
		return sdkmdadm.Detail(device), nil
	})
	r.Register("mdadm.scan", func(args map[string]interface{}) (interface{}, error) {
		return sdkmdadm.Scan(), nil
	})
	r.Register("mdadm.add", func(args map[string]interface{}) (interface{}, error) {
		device, _ := argString(args, "device")
		member, _ := argString(args, "member")
		return sdkmdadm.Add(device, member), nil
	})
	r.Register("mdadm.remove", func(args map[string]interface{}) (interface{}, error) {
		device, _ := argString(args, "device")
		member, _ := argString(args, "member")
		return sdkmdadm.Remove(device, member), nil
	})

	// ── open_iscsi ──────────────────────────────────────────────────────────
	r.Register("open_iscsi.discover", func(args map[string]interface{}) (interface{}, error) {
		portal, _ := argString(args, "portal")
		port, _ := argInt(args, "port")
		return sdkopen_iscsi.Discover(portal, port), nil
	})
	r.Register("open_iscsi.login", func(args map[string]interface{}) (interface{}, error) {
		target, _ := argString(args, "target")
		portal, _ := argString(args, "portal")
		return sdkopen_iscsi.Login(target, portal), nil
	})
	r.Register("open_iscsi.logout", func(args map[string]interface{}) (interface{}, error) {
		target, _ := argString(args, "target")
		portal, _ := argString(args, "portal")
		return sdkopen_iscsi.Logout(target, portal), nil
	})
	r.Register("open_iscsi.list_sessions", func(args map[string]interface{}) (interface{}, error) {
		return sdkopen_iscsi.ListSessions(), nil
	})
	r.Register("open_iscsi.list_nodes", func(args map[string]interface{}) (interface{}, error) {
		return sdkopen_iscsi.ListNodes(), nil
	})
	r.Register("open_iscsi.set_startup", func(args map[string]interface{}) (interface{}, error) {
		target, _ := argString(args, "target")
		portal, _ := argString(args, "portal")
		startup, _ := argString(args, "startup")
		return sdkopen_iscsi.SetStartup(target, portal, startup), nil
	})

	// ── rfkill ──────────────────────────────────────────────────────────────
	r.Register("rfkill.list", func(args map[string]interface{}) (interface{}, error) {
		return sdkrfkill.List(), nil
	})
	r.Register("rfkill.block", func(args map[string]interface{}) (interface{}, error) {
		device, _ := argString(args, "device")
		return sdkrfkill.Block(device), nil
	})
	r.Register("rfkill.unblock", func(args map[string]interface{}) (interface{}, error) {
		device, _ := argString(args, "device")
		return sdkrfkill.Unblock(device), nil
	})
	r.Register("rfkill.block_all", func(args map[string]interface{}) (interface{}, error) {
		deviceType, _ := argString(args, "type")
		return sdkrfkill.BlockAll(deviceType), nil
	})
	r.Register("rfkill.unblock_all", func(args map[string]interface{}) (interface{}, error) {
		deviceType, _ := argString(args, "type")
		return sdkrfkill.UnblockAll(deviceType), nil
	})

	// ── multipath ───────────────────────────────────────────────────────────
	r.Register("multipath.reconfigure", func(args map[string]interface{}) (interface{}, error) {
		return sdkmultipath.Reconfigure(), nil
	})
	r.Register("multipath.list_paths", func(args map[string]interface{}) (interface{}, error) {
		return sdkmultipath.ListPaths(), nil
	})
	r.Register("multipath.list_maps", func(args map[string]interface{}) (interface{}, error) {
		return sdkmultipath.ListMaps(), nil
	})
	r.Register("multipath.add_map", func(args map[string]interface{}) (interface{}, error) {
		device, _ := argString(args, "device")
		return sdkmultipath.AddMap(device), nil
	})
	r.Register("multipath.remove_map", func(args map[string]interface{}) (interface{}, error) {
		device, _ := argString(args, "device")
		return sdkmultipath.RemoveMap(device), nil
	})
	r.Register("multipath.flush", func(args map[string]interface{}) (interface{}, error) {
		return sdkmultipath.Flush(), nil
	})

	// ── dmsetup ─────────────────────────────────────────────────────────────
	r.Register("dmsetup.create", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		table, _ := argString(args, "table")
		return sdkdmsetup.Create(name, table), nil
	})
	r.Register("dmsetup.remove", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		return sdkdmsetup.Remove(name), nil
	})
	r.Register("dmsetup.remove_all", func(args map[string]interface{}) (interface{}, error) {
		return sdkdmsetup.RemoveAll(), nil
	})
	r.Register("dmsetup.list", func(args map[string]interface{}) (interface{}, error) {
		return sdkdmsetup.List(), nil
	})
	r.Register("dmsetup.info", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		return sdkdmsetup.Info(name), nil
	})
	r.Register("dmsetup.suspend", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		return sdkdmsetup.Suspend(name), nil
	})
	r.Register("dmsetup.resume", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		return sdkdmsetup.Resume(name), nil
	})

	// ── lvm_enhanced ────────────────────────────────────────────────────────
	r.Register("lvm_enhanced.pv_create", func(args map[string]interface{}) (interface{}, error) {
		device, _ := argString(args, "device")
		return sdklvm_enhanced.PVCreate(device), nil
	})
	r.Register("lvm_enhanced.pv_remove", func(args map[string]interface{}) (interface{}, error) {
		device, _ := argString(args, "device")
		force, _ := argBool(args, "force")
		return sdklvm_enhanced.PVRemove(device, force), nil
	})
	r.Register("lvm_enhanced.pv_list", func(args map[string]interface{}) (interface{}, error) {
		return sdklvm_enhanced.PVList(), nil
	})
	r.Register("lvm_enhanced.vg_create", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		devices := argStringSlice(args, "devices")
		return sdklvm_enhanced.VGCreate(name, devices), nil
	})
	r.Register("lvm_enhanced.vg_remove", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		force, _ := argBool(args, "force")
		return sdklvm_enhanced.VGRemove(name, force), nil
	})
	r.Register("lvm_enhanced.vg_extend", func(args map[string]interface{}) (interface{}, error) {
		vgName, _ := argString(args, "vg_name")
		device, _ := argString(args, "device")
		return sdklvm_enhanced.VGExtend(vgName, device), nil
	})
	r.Register("lvm_enhanced.vg_list", func(args map[string]interface{}) (interface{}, error) {
		return sdklvm_enhanced.VGList(), nil
	})
	r.Register("lvm_enhanced.lv_extend", func(args map[string]interface{}) (interface{}, error) {
		lvPath, _ := argString(args, "lv_path")
		size, _ := argString(args, "size")
		return sdklvm_enhanced.LVExtend(lvPath, size), nil
	})
	r.Register("lvm_enhanced.lv_extend_all", func(args map[string]interface{}) (interface{}, error) {
		lvPath, _ := argString(args, "lv_path")
		return sdklvm_enhanced.LVExtendAll(lvPath), nil
	})
	r.Register("lvm_enhanced.lv_list", func(args map[string]interface{}) (interface{}, error) {
		return sdklvm_enhanced.LVList(), nil
	})

	// ── puppet ──────────────────────────────────────────────────────────────
	r.Register("puppet.run", func(args map[string]interface{}) (interface{}, error) {
		environment, _ := argString(args, "environment")
		tags := argStringSlice(args, "tags")
		return sdkpuppet.Run(environment, tags), nil
	})
	r.Register("puppet.run_noop", func(args map[string]interface{}) (interface{}, error) {
		environment, _ := argString(args, "environment")
		tags := argStringSlice(args, "tags")
		return sdkpuppet.RunNoop(environment, tags), nil
	})
	r.Register("puppet.status", func(args map[string]interface{}) (interface{}, error) {
		return sdkpuppet.Status(), nil
	})
	r.Register("puppet.disable", func(args map[string]interface{}) (interface{}, error) {
		message, _ := argString(args, "message")
		return sdkpuppet.Disable(message), nil
	})
	r.Register("puppet.enable", func(args map[string]interface{}) (interface{}, error) {
		return sdkpuppet.Enable(), nil
	})
	r.Register("puppet.fact", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		return sdkpuppet.Fact(name), nil
	})
	r.Register("puppet.module_list", func(args map[string]interface{}) (interface{}, error) {
		return sdkpuppet.ModuleList(), nil
	})
	r.Register("puppet.module_install", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		version, _ := argString(args, "version")
		return sdkpuppet.ModuleInstall(name, version), nil
	})

	// ── svn ───────────────────────────────────────────────────────────────
	r.Register("svn.checkout", func(args map[string]interface{}) (interface{}, error) {
		url, _ := argString(args, "url")
		dest, _ := argString(args, "dest")
		revision, _ := argString(args, "revision")
		force, _ := argBool(args, "force")
		return sdksvn.Checkout(url, dest, revision, force)
	})
	r.Register("svn.update", func(args map[string]interface{}) (interface{}, error) {
		dest, _ := argString(args, "dest")
		revision, _ := argString(args, "revision")
		return sdksvn.Update(dest, revision)
	})
	r.Register("svn.export", func(args map[string]interface{}) (interface{}, error) {
		url, _ := argString(args, "url")
		dest, _ := argString(args, "dest")
		revision, _ := argString(args, "revision")
		force, _ := argBool(args, "force")
		return sdksvn.Export(url, dest, revision, force)
	})
	r.Register("svn.status", func(args map[string]interface{}) (interface{}, error) {
		dest, _ := argString(args, "dest")
		return sdksvn.Status(dest)
	})
	r.Register("svn.info", func(args map[string]interface{}) (interface{}, error) {
		dest, _ := argString(args, "dest")
		return sdksvn.Info(dest)
	})
	r.Register("svn.cleanup", func(args map[string]interface{}) (interface{}, error) {
		dest, _ := argString(args, "dest")
		return sdksvn.Cleanup(dest)
	})
	r.Register("svn.revert", func(args map[string]interface{}) (interface{}, error) {
		dest, _ := argString(args, "dest")
		recursive, _ := argBool(args, "recursive")
		return sdksvn.Revert(dest, recursive)
	})

	// ── zypper ────────────────────────────────────────────────────────────
	r.Register("zypper.install", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		version, _ := argString(args, "version")
		return sdkzypper.Install(name, version)
	})
	r.Register("zypper.remove", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		return sdkzypper.Remove(name)
	})
	r.Register("zypper.update", func(args map[string]interface{}) (interface{}, error) {
		name := getStringArg(args, "name", "")
		return sdkzypper.Update(name)
	})
	r.Register("zypper.dist_upgrade", func(_ map[string]interface{}) (interface{}, error) {
		return sdkzypper.DistUpgrade()
	})
	r.Register("zypper.info", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		return sdkzypper.Info(name)
	})
	r.Register("zypper.list", func(_ map[string]interface{}) (interface{}, error) {
		return sdkzypper.List()
	})
	r.Register("zypper.clean", func(_ map[string]interface{}) (interface{}, error) {
		return sdkzypper.Clean()
	})
	r.Register("zypper.repo_list", func(_ map[string]interface{}) (interface{}, error) {
		return sdkzypper.RepoList()
	})
	r.Register("zypper.repo_add", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		url, _ := argString(args, "url")
		return sdkzypper.RepoAdd(name, url)
	})
	r.Register("zypper.repo_remove", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		return sdkzypper.RepoRemove(name)
	})
	r.Register("zypper.refresh", func(_ map[string]interface{}) (interface{}, error) {
		return sdkzypper.Refresh()
	})
	r.Register("zypper.search", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		return sdkzypper.Search(name)
	})
	r.Register("zypper.patch", func(_ map[string]interface{}) (interface{}, error) {
		return sdkzypper.Patch()
	})
	r.Register("zypper.pattern_install", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		return sdkzypper.PatternInstall(name)
	})
	r.Register("zypper.pattern_remove", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		return sdkzypper.PatternRemove(name)
	})

	// ── pacman ────────────────────────────────────────────────────────────
	r.Register("pacman.install", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		return sdkpacman.Install(name)
	})
	r.Register("pacman.remove", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		cascade, _ := argBool(args, "cascade")
		return sdkpacman.Remove(name, cascade)
	})
	r.Register("pacman.update", func(args map[string]interface{}) (interface{}, error) {
		name := getStringArg(args, "name", "")
		return sdkpacman.Update(name)
	})
	r.Register("pacman.upgrade", func(_ map[string]interface{}) (interface{}, error) {
		return sdkpacman.Upgrade()
	})
	r.Register("pacman.info", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		return sdkpacman.Info(name)
	})
	r.Register("pacman.list", func(_ map[string]interface{}) (interface{}, error) {
		return sdkpacman.List()
	})
	r.Register("pacman.search", func(args map[string]interface{}) (interface{}, error) {
		name, _ := argString(args, "name")
		return sdkpacman.Search(name)
	})
	r.Register("pacman.clean", func(_ map[string]interface{}) (interface{}, error) {
		return sdkpacman.Clean()
	})
	r.Register("pacman.install_file", func(args map[string]interface{}) (interface{}, error) {
		path, _ := argString(args, "path")
		return sdkpacman.InstallFile(path)
	})
	r.Register("pacman.remove_orphans", func(_ map[string]interface{}) (interface{}, error) {
		return sdkpacman.RemoveOrphans()
	})
	r.Register("pacman.update_database", func(_ map[string]interface{}) (interface{}, error) {
		return sdkpacman.UpdateDatabase()
	})

	// ── kubernetes.* ──────────────────────────────────────────────────
	r.Register("kubernetes.apply", func(args map[string]interface{}) (interface{}, error) {
		manifest, err := argString(args, "manifest")
		if err != nil {
			return nil, fmt.Errorf("kubernetes.apply: %w", err)
		}
		namespace, _ := args["namespace"].(string)
		dryRun, _ := args["dry_run"].(bool)
		return sdkk8s.Apply(manifest, namespace, dryRun)
	})
	r.Register("kubernetes.delete", func(args map[string]interface{}) (interface{}, error) {
		manifest, err := argString(args, "manifest")
		if err != nil {
			return nil, fmt.Errorf("kubernetes.delete: %w", err)
		}
		namespace, _ := args["namespace"].(string)
		return sdkk8s.Delete(manifest, namespace)
	})
	r.Register("kubernetes.get", func(args map[string]interface{}) (interface{}, error) {
		rt, err := argString(args, "resource_type")
		if err != nil {
			return nil, fmt.Errorf("kubernetes.get: %w", err)
		}
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("kubernetes.get: %w", err)
		}
		namespace, _ := args["namespace"].(string)
		return sdkk8s.Get(rt, name, namespace)
	})
	r.Register("kubernetes.list", func(args map[string]interface{}) (interface{}, error) {
		rt, err := argString(args, "resource_type")
		if err != nil {
			return nil, fmt.Errorf("kubernetes.list: %w", err)
		}
		namespace, _ := args["namespace"].(string)
		labels, _ := args["labels"].(string)
		return sdkk8s.List(rt, namespace, labels)
	})
	r.Register("kubernetes.create_namespace", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("kubernetes.create_namespace: %w", err)
		}
		return sdkk8s.CreateNamespace(name)
	})
	r.Register("kubernetes.delete_namespace", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("kubernetes.delete_namespace: %w", err)
		}
		return sdkk8s.DeleteNamespace(name)
	})
	r.Register("kubernetes.get_pods", func(args map[string]interface{}) (interface{}, error) {
		namespace, _ := args["namespace"].(string)
		labels, _ := args["labels"].(string)
		return sdkk8s.GetPods(namespace, labels)
	})
	r.Register("kubernetes.get_services", func(args map[string]interface{}) (interface{}, error) {
		namespace, _ := args["namespace"].(string)
		return sdkk8s.GetServices(namespace)
	})
	r.Register("kubernetes.get_deployments", func(args map[string]interface{}) (interface{}, error) {
		namespace, _ := args["namespace"].(string)
		return sdkk8s.GetDeployments(namespace)
	})
	r.Register("kubernetes.scale", func(args map[string]interface{}) (interface{}, error) {
		dep, err := argString(args, "deployment")
		if err != nil {
			return nil, fmt.Errorf("kubernetes.scale: %w", err)
		}
		replicas := 1
		if v, ok := args["replicas"].(int); ok {
			replicas = v
		} else if v, ok := args["replicas"].(float64); ok {
			replicas = int(v)
		}
		namespace, _ := args["namespace"].(string)
		return sdkk8s.Scale(dep, replicas, namespace)
	})
	r.Register("kubernetes.rollout_status", func(args map[string]interface{}) (interface{}, error) {
		dep, err := argString(args, "deployment")
		if err != nil {
			return nil, fmt.Errorf("kubernetes.rollout_status: %w", err)
		}
		namespace, _ := args["namespace"].(string)
		return sdkk8s.RolloutStatus(dep, namespace)
	})
	r.Register("kubernetes.exec", func(args map[string]interface{}) (interface{}, error) {
		pod, err := argString(args, "pod")
		if err != nil {
			return nil, fmt.Errorf("kubernetes.exec: %w", err)
		}
		command, err := argString(args, "command")
		if err != nil {
			return nil, fmt.Errorf("kubernetes.exec: %w", err)
		}
		namespace, _ := args["namespace"].(string)
		container, _ := args["container"].(string)
		return sdkk8s.Exec(pod, command, namespace, container)
	})
	r.Register("kubernetes.logs", func(args map[string]interface{}) (interface{}, error) {
		pod, err := argString(args, "pod")
		if err != nil {
			return nil, fmt.Errorf("kubernetes.logs: %w", err)
		}
		namespace, _ := args["namespace"].(string)
		container, _ := args["container"].(string)
		tail := 0
		if v, ok := args["tail"].(int); ok {
			tail = v
		} else if v, ok := args["tail"].(float64); ok {
			tail = int(v)
		}
		return sdkk8s.Logs(pod, namespace, container, tail)
	})
	r.Register("kubernetes.wait_ready", func(args map[string]interface{}) (interface{}, error) {
		rt, err := argString(args, "resource_type")
		if err != nil {
			return nil, fmt.Errorf("kubernetes.wait_ready: %w", err)
		}
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("kubernetes.wait_ready: %w", err)
		}
		namespace, _ := args["namespace"].(string)
		timeout := 300
		if v, ok := args["timeout"].(int); ok {
			timeout = v
		} else if v, ok := args["timeout"].(float64); ok {
			timeout = int(v)
		}
		return sdkk8s.WaitReady(rt, name, namespace, timeout)
	})

	// ── portage ─────────────────────────────────────────────────────
	r.Register("portage.install", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("portage.install: %w", err)
		}
		version, _ := args["version"].(string)
		return sdkportage.Install(name, version)
	})
	r.Register("portage.remove", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("portage.remove: %w", err)
		}
		return sdkportage.Remove(name)
	})
	r.Register("portage.update", func(args map[string]interface{}) (interface{}, error) {
		name, _ := args["name"].(string)
		deep, _ := args["deep"].(bool)
		return sdkportage.Update(name, deep)
	})
	r.Register("portage.sync", func(args map[string]interface{}) (interface{}, error) {
		return sdkportage.Sync()
	})
	r.Register("portage.info", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("portage.info: %w", err)
		}
		return sdkportage.Info(name)
	})
	r.Register("portage.list", func(args map[string]interface{}) (interface{}, error) {
		return sdkportage.List()
	})
	r.Register("portage.search", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("portage.search: %w", err)
		}
		return sdkportage.Search(name)
	})
	r.Register("portage.depclean", func(args map[string]interface{}) (interface{}, error) {
		return sdkportage.Depclean()
	})
	r.Register("portage.metadata", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("portage.metadata: %w", err)
		}
		return sdkportage.Metadata(name)
	})

	// ── pkgng ───────────────────────────────────────────────────────
	r.Register("pkgng.install", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("pkgng.install: %w", err)
		}
		version, _ := args["version"].(string)
		return sdkpkgng.Install(name, version)
	})
	r.Register("pkgng.remove", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("pkgng.remove: %w", err)
		}
		return sdkpkgng.Remove(name)
	})
	r.Register("pkgng.update", func(args map[string]interface{}) (interface{}, error) {
		return sdkpkgng.Update()
	})
	r.Register("pkgng.upgrade", func(args map[string]interface{}) (interface{}, error) {
		name, _ := args["name"].(string)
		return sdkpkgng.Upgrade(name)
	})
	r.Register("pkgng.autoclean", func(args map[string]interface{}) (interface{}, error) {
		return sdkpkgng.Autoclean()
	})
	r.Register("pkgng.info", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("pkgng.info: %w", err)
		}
		return sdkpkgng.Info(name)
	})
	r.Register("pkgng.list", func(args map[string]interface{}) (interface{}, error) {
		return sdkpkgng.List()
	})
	r.Register("pkgng.search", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("pkgng.search: %w", err)
		}
		return sdkpkgng.Search(name)
	})
	r.Register("pkgng.stats", func(args map[string]interface{}) (interface{}, error) {
		return sdkpkgng.Stats()
	})

	// ── podman ──────────────────────────────────────────────────────
	r.Register("podman.run", func(args map[string]interface{}) (interface{}, error) {
		image, err := argString(args, "image")
		if err != nil {
			return nil, fmt.Errorf("podman.run: %w", err)
		}
		name, _ := args["name"].(string)
		command, _ := args["command"].(string)
		return sdkpodman.Run(image, name, command)
	})
	r.Register("podman.stop", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("podman.stop: %w", err)
		}
		timeout := 0
		if v, ok := args["timeout"].(int); ok {
			timeout = v
		} else if v, ok := args["timeout"].(float64); ok {
			timeout = int(v)
		}
		return sdkpodman.Stop(name, timeout)
	})
	r.Register("podman.start", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("podman.start: %w", err)
		}
		return sdkpodman.Start(name)
	})
	r.Register("podman.remove", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("podman.remove: %w", err)
		}
		force, _ := args["force"].(bool)
		return sdkpodman.Remove(name, force)
	})
	r.Register("podman.list_containers", func(args map[string]interface{}) (interface{}, error) {
		all, _ := args["all"].(bool)
		return sdkpodman.ListContainers(all)
	})
	r.Register("podman.inspect", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("podman.inspect: %w", err)
		}
		return sdkpodman.Inspect(name)
	})
	r.Register("podman.pull", func(args map[string]interface{}) (interface{}, error) {
		image, err := argString(args, "image")
		if err != nil {
			return nil, fmt.Errorf("podman.pull: %w", err)
		}
		return sdkpodman.Pull(image)
	})
	r.Register("podman.list_images", func(args map[string]interface{}) (interface{}, error) {
		return sdkpodman.ListImages()
	})
	r.Register("podman.remove_image", func(args map[string]interface{}) (interface{}, error) {
		imageID, err := argString(args, "image_id")
		if err != nil {
			return nil, fmt.Errorf("podman.remove_image: %w", err)
		}
		force, _ := args["force"].(bool)
		return sdkpodman.RemoveImage(imageID, force)
	})
	r.Register("podman.create_pod", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("podman.create_pod: %w", err)
		}
		return sdkpodman.CreatePod(name)
	})
	r.Register("podman.stop_pod", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("podman.stop_pod: %w", err)
		}
		return sdkpodman.StopPod(name)
	})
	r.Register("podman.remove_pod", func(args map[string]interface{}) (interface{}, error) {
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("podman.remove_pod: %w", err)
		}
		force, _ := args["force"].(bool)
		return sdkpodman.RemovePod(name, force)
	})
	r.Register("podman.list_pods", func(args map[string]interface{}) (interface{}, error) {
		return sdkpodman.ListPods()
	})

	// ── nftables ────────────────────────────────────────────────────
	r.Register("nftables.add_table", func(args map[string]interface{}) (interface{}, error) {
		family, err := argString(args, "family")
		if err != nil {
			return nil, fmt.Errorf("nftables.add_table: %w", err)
		}
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("nftables.add_table: %w", err)
		}
		return sdknftables.AddTable(family, name)
	})
	r.Register("nftables.delete_table", func(args map[string]interface{}) (interface{}, error) {
		family, err := argString(args, "family")
		if err != nil {
			return nil, fmt.Errorf("nftables.delete_table: %w", err)
		}
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("nftables.delete_table: %w", err)
		}
		return sdknftables.DeleteTable(family, name)
	})
	r.Register("nftables.list_tables", func(args map[string]interface{}) (interface{}, error) {
		return sdknftables.ListTables()
	})
	r.Register("nftables.add_chain", func(args map[string]interface{}) (interface{}, error) {
		family, err := argString(args, "family")
		if err != nil {
			return nil, fmt.Errorf("nftables.add_chain: %w", err)
		}
		table, err := argString(args, "table")
		if err != nil {
			return nil, fmt.Errorf("nftables.add_chain: %w", err)
		}
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("nftables.add_chain: %w", err)
		}
		chainType, _ := args["type"].(string)
		hook, _ := args["hook"].(string)
		priority, _ := args["priority"].(string)
		return sdknftables.AddChain(family, table, name, chainType, hook, priority)
	})
	r.Register("nftables.delete_chain", func(args map[string]interface{}) (interface{}, error) {
		family, err := argString(args, "family")
		if err != nil {
			return nil, fmt.Errorf("nftables.delete_chain: %w", err)
		}
		table, err := argString(args, "table")
		if err != nil {
			return nil, fmt.Errorf("nftables.delete_chain: %w", err)
		}
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("nftables.delete_chain: %w", err)
		}
		return sdknftables.DeleteChain(family, table, name)
	})
	r.Register("nftables.add_rule", func(args map[string]interface{}) (interface{}, error) {
		family, err := argString(args, "family")
		if err != nil {
			return nil, fmt.Errorf("nftables.add_rule: %w", err)
		}
		table, err := argString(args, "table")
		if err != nil {
			return nil, fmt.Errorf("nftables.add_rule: %w", err)
		}
		chain, err := argString(args, "chain")
		if err != nil {
			return nil, fmt.Errorf("nftables.add_rule: %w", err)
		}
		expr, err := argString(args, "expression")
		if err != nil {
			return nil, fmt.Errorf("nftables.add_rule: %w", err)
		}
		return sdknftables.AddRule(family, table, chain, expr)
	})
	r.Register("nftables.delete_rule", func(args map[string]interface{}) (interface{}, error) {
		family, err := argString(args, "family")
		if err != nil {
			return nil, fmt.Errorf("nftables.delete_rule: %w", err)
		}
		table, err := argString(args, "table")
		if err != nil {
			return nil, fmt.Errorf("nftables.delete_rule: %w", err)
		}
		chain, err := argString(args, "chain")
		if err != nil {
			return nil, fmt.Errorf("nftables.delete_rule: %w", err)
		}
		handle, err := argString(args, "handle")
		if err != nil {
			return nil, fmt.Errorf("nftables.delete_rule: %w", err)
		}
		return sdknftables.DeleteRule(family, table, chain, handle)
	})
	r.Register("nftables.flush_chain", func(args map[string]interface{}) (interface{}, error) {
		family, err := argString(args, "family")
		if err != nil {
			return nil, fmt.Errorf("nftables.flush_chain: %w", err)
		}
		table, err := argString(args, "table")
		if err != nil {
			return nil, fmt.Errorf("nftables.flush_chain: %w", err)
		}
		chain, err := argString(args, "chain")
		if err != nil {
			return nil, fmt.Errorf("nftables.flush_chain: %w", err)
		}
		return sdknftables.FlushChain(family, table, chain)
	})
	r.Register("nftables.flush_table", func(args map[string]interface{}) (interface{}, error) {
		family, err := argString(args, "family")
		if err != nil {
			return nil, fmt.Errorf("nftables.flush_table: %w", err)
		}
		table, err := argString(args, "table")
		if err != nil {
			return nil, fmt.Errorf("nftables.flush_table: %w", err)
		}
		return sdknftables.FlushTable(family, table)
	})
	r.Register("nftables.flush_ruleset", func(args map[string]interface{}) (interface{}, error) {
		return sdknftables.FlushRuleset()
	})
	r.Register("nftables.list_ruleset", func(args map[string]interface{}) (interface{}, error) {
		return sdknftables.ListRuleset()
	})
	r.Register("nftables.add_set", func(args map[string]interface{}) (interface{}, error) {
		family, err := argString(args, "family")
		if err != nil {
			return nil, fmt.Errorf("nftables.add_set: %w", err)
		}
		table, err := argString(args, "table")
		if err != nil {
			return nil, fmt.Errorf("nftables.add_set: %w", err)
		}
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("nftables.add_set: %w", err)
		}
		setType, err := argString(args, "type")
		if err != nil {
			return nil, fmt.Errorf("nftables.add_set: %w", err)
		}
		flags, _ := args["flags"].(string)
		return sdknftables.AddSet(family, table, name, setType, flags)
	})
	r.Register("nftables.delete_set", func(args map[string]interface{}) (interface{}, error) {
		family, err := argString(args, "family")
		if err != nil {
			return nil, fmt.Errorf("nftables.delete_set: %w", err)
		}
		table, err := argString(args, "table")
		if err != nil {
			return nil, fmt.Errorf("nftables.delete_set: %w", err)
		}
		name, err := argString(args, "name")
		if err != nil {
			return nil, fmt.Errorf("nftables.delete_set: %w", err)
		}
		return sdknftables.DeleteSet(family, table, name)
	})
	r.Register("nftables.add_element", func(args map[string]interface{}) (interface{}, error) {
		family, err := argString(args, "family")
		if err != nil {
			return nil, fmt.Errorf("nftables.add_element: %w", err)
		}
		table, err := argString(args, "table")
		if err != nil {
			return nil, fmt.Errorf("nftables.add_element: %w", err)
		}
		set, err := argString(args, "set")
		if err != nil {
			return nil, fmt.Errorf("nftables.add_element: %w", err)
		}
		element, err := argString(args, "element")
		if err != nil {
			return nil, fmt.Errorf("nftables.add_element: %w", err)
		}
		return sdknftables.AddElement(family, table, set, element)
	})
	r.Register("nftables.delete_element", func(args map[string]interface{}) (interface{}, error) {
		family, err := argString(args, "family")
		if err != nil {
			return nil, fmt.Errorf("nftables.delete_element: %w", err)
		}
		table, err := argString(args, "table")
		if err != nil {
			return nil, fmt.Errorf("nftables.delete_element: %w", err)
		}
		set, err := argString(args, "set")
		if err != nil {
			return nil, fmt.Errorf("nftables.delete_element: %w", err)
		}
		element, err := argString(args, "element")
		if err != nil {
			return nil, fmt.Errorf("nftables.delete_element: %w", err)
		}
		return sdknftables.DeleteElement(family, table, set, element)
	})
	r.Register("nftables.export", func(args map[string]interface{}) (interface{}, error) {
		format, _ := args["format"].(string)
		if format == "" {
			format = "json"
		}
		return sdknftables.Export(format)
	})
}

// toStringMapArg extracts a map[string]string from args[key].
func toStringMapArg(args map[string]interface{}, key string) map[string]string {
	result := make(map[string]string)
	raw, ok := args[key]
	if !ok || raw == nil {
		return result
	}
	m, ok := raw.(map[string]interface{})
	if !ok {
		return result
	}
	for k, v := range m {
		result[k] = fmt.Sprintf("%v", v)
	}
	return result
}

// mapStrArg extracts a string from a map with a default.
func mapStrArg(m map[string]interface{}, key string, def string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
		return fmt.Sprintf("%v", v)
	}
	return def
}

// ============================================================
// built-in operations (log, alert, set, report, binary.exec)
// ============================================================

func (r *Registry) registerBuiltinOps() {
	// report: collects named values into the output data map. The executor
	// resolves "$name" variable references before it runs.
	r.Register("report", func(args map[string]interface{}) (interface{}, error) {
		result := make(map[string]interface{}, len(args))
		for key, value := range args {
			result[key] = value
		}
		return result, nil
	})

	// set: stores a value; the executor handles variable assignment.
	r.Register("set", func(args map[string]interface{}) (interface{}, error) {
		v, ok := args["value"]
		if !ok {
			return nil, fmt.Errorf("set: argument \"value\" is required")
		}
		return v, nil
	})

	r.Register("log", func(args map[string]interface{}) (interface{}, error) {
		return getStringArg(args, "message", ""), nil
	})

	r.Register("alert", func(args map[string]interface{}) (interface{}, error) {
		msg, err := argString(args, "message")
		if err != nil {
			return nil, fmt.Errorf("alert: %w", err)
		}
		return msg, nil
	})

	// binary.exec: executes a compiled OpsLang binary (AOT deploy mode) and
	// returns its parsed JSON output. A non-zero exit or startup failure is
	// an ERROR, not a result — deployment must not report success when the
	// program failed.
	r.Register("binary.exec", func(args map[string]interface{}) (interface{}, error) {
		path, err := argString(args, "path")
		if err != nil {
			return nil, fmt.Errorf("binary.exec: %w", err)
		}

		var execArgs []string
		if a, ok := args["args"].([]interface{}); ok {
			for _, v := range a {
				if s, ok := v.(string); ok {
					execArgs = append(execArgs, s)
				}
			}
		}

		cmd := exec.Command(path, execArgs...)
		output, err := cmd.Output()
		if err != nil {
			stderr := ""
			if exitErr, ok := err.(*exec.ExitError); ok {
				stderr = string(exitErr.Stderr)
			}
			return nil, fmt.Errorf("binary.exec: %s failed: %v%s", path, err, stderr)
		}

		// The AOT binary prints its report as JSON on stdout.
		var result interface{}
		if jsonErr := json.Unmarshal(output, &result); jsonErr == nil {
			return result, nil
		}
		return map[string]interface{}{
			"output": string(output),
		}, nil
	})
}

// ============================================================
// Argument helper functions
// ============================================================

// getStringArg returns the string value of an optional arg, or defaultVal.
func getStringArg(args map[string]interface{}, key string, defaultVal string) string {
	if v, ok := args[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return defaultVal
}

// getBoolArg returns the bool value of an optional arg, or defaultVal.
func getBoolArg(args map[string]interface{}, key string, defaultVal bool) bool {
	if v, ok := args[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return defaultVal
}

// argString returns a required string arg or a descriptive error. Silent
// empty-string fallbacks masked broken instruction packages in the past;
// missing required arguments must fail loudly.
func argString(args map[string]interface{}, key string) (string, error) {
	v, ok := args[key]
	if !ok {
		return "", fmt.Errorf("missing required argument %q", key)
	}
	switch s := v.(type) {
	case string:
		return s, nil
	default:
		return "", fmt.Errorf("argument %q must be a string, got %T", key, v)
	}
}

// argInt returns a required integer arg, accepting the numeric types that
// JSON unmarshalling can produce (float64) as well as native ints.
func argInt(args map[string]interface{}, key string) (int, error) {
	v, ok := args[key]
	if !ok {
		return 0, fmt.Errorf("missing required argument %q", key)
	}
	switch n := v.(type) {
	case float64:
		return int(n), nil
	case int:
		return n, nil
	case int64:
		return int(n), nil
	default:
		return 0, fmt.Errorf("argument %q must be a number, got %T", key, v)
	}
}

// argInt64 is argInt with an int64 result.
func argInt64(args map[string]interface{}, key string) (int64, error) {
	v, ok := args[key]
	if !ok {
		return 0, fmt.Errorf("missing required argument %q", key)
	}
	switch n := v.(type) {
	case float64:
		return int64(n), nil
	case int:
		return int64(n), nil
	case int64:
		return n, nil
	default:
		return 0, fmt.Errorf("argument %q must be a number, got %T", key, v)
	}
}

// argBool returns a boolean arg, defaulting to false if missing.
func argBool(args map[string]interface{}, key string) (bool, error) {
	v, ok := args[key]
	if !ok {
		return false, nil
	}
	switch b := v.(type) {
	case bool:
		return b, nil
	case float64:
		return b != 0, nil
	default:
		return false, fmt.Errorf("argument %q must be a boolean, got %T", key, v)
	}
}

// argStringSlice extracts a []string from args[key].
func argStringSlice(args map[string]interface{}, key string) []string {
	v, ok := args[key]
	if !ok || v == nil {
		return nil
	}
	switch s := v.(type) {
	case []string:
		return s
	case []interface{}:
		result := make([]string, 0, len(s))
		for _, item := range s {
			if str, ok := item.(string); ok {
				result = append(result, str)
			}
		}
		return result
	case string:
		return []string{s}
	}
	return nil
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
	// The privilege field is optional (legacy packages omit it), but when
	// present it must be a declared level so the runner's second check
	// cannot be silently downgraded by a typo.
	switch pkg.Privilege {
	case "", string(ast.PrivilegeReadOnly), string(ast.PrivilegeAdmin), string(ast.PrivilegeRoot):
	default:
		return fmt.Errorf("invalid privilege %q (expected read_only, admin, or root)", pkg.Privilege)
	}
	registry := NewRegistry()
	for i, inst := range pkg.Instructions {
		if inst.Op == "" {
			return fmt.Errorf("instruction %d: op is required", i)
		}
		if !registry.Has(inst.Op) {
			return fmt.Errorf("instruction %d: unknown operation %q", i, inst.Op)
		}
	}
	return nil
}
