// Package opsspec defines the canonical registry of every OpsLang atomic
// operation: its dotted DSL name, its positional argument names, and where
// it can run. This table is the single source of truth shared by the three
// execution engines (interpreter, runner registry, AOT code generator).
// A cross-engine consistency test enforces that all engines agree with it.
package opsspec

// Availability describes which engines expose a function.
type Availability int

const (
	// All means the function is available in the interpreter (controller),
	// the remote runner registry, and the AOT code generator.
	All Availability = iota
	// ControllerOnly means the function only makes sense on the controller
	// machine (e.g. it fans out over SSH itself); the remote runner and the
	// AOT code generator must not expose it.
	ControllerOnly
)

// Func describes one atomic operation.
type Func struct {
	// Name is the canonical dotted DSL name, e.g. "sys.cpu.usage".
	// All engines must resolve exactly this name; historical aliases are
	// tolerated at lookup time but never generated.
	Name string
	// Args lists positional argument names in call order, e.g. ["path"].
	// Engines use these names when building instruction argument maps.
	Args []string
	// Avail restricts which engines expose the function.
	Avail Availability
	// Mutating marks functions that change system or remote state (files,
	// processes, services, packages, or the state of a remote HTTP
	// endpoint). Mutating functions require at least admin privilege;
	// non-mutating ones are available to every privilege level. This flag
	// is the single source of truth shared by the interpreter, the
	// instruction generator, the AOT compiler and the runner — engines
	// must not keep their own mutation lists.
	Mutating bool
}

// Funcs is the canonical table. Keep sorted by category then name.
var Funcs = []Func{
	// ── cron ──────────────────────────────────────────────────────────
	{Name: "cron.add", Args: []string{"user", "entry"}, Mutating: true},
	{Name: "cron.list", Args: []string{"user"}},
	{Name: "cron.remove", Args: []string{"user", "line_match"}, Mutating: true},

	// ── archive ───────────────────────────────────────────────────────
	{Name: "archive.create", Args: []string{"dest", "sources"}, Mutating: true},
	{Name: "archive.extract", Args: []string{"src", "dest"}, Mutating: true},

	// ── apt ─────────────────────────────────────────────────────────
	{Name: "apt.install", Args: []string{"name", "version", "update_cache"}, Mutating: true},
	{Name: "apt.remove", Args: []string{"name", "purge"}, Mutating: true},
	{Name: "apt.upgrade", Args: []string{"name"}, Mutating: true},
	{Name: "apt.update_cache", Mutating: true},
	{Name: "apt.full_upgrade", Mutating: true},
	{Name: "apt.dist_upgrade", Mutating: true},
	{Name: "apt.autoremove", Mutating: true},
	{Name: "apt.clean", Mutating: true},
	{Name: "apt.info", Args: []string{"name"}},
	{Name: "apt.list"},
	{Name: "apt.policy", Args: []string{"name"}},
	{Name: "apt.mark_auto", Args: []string{"name"}, Mutating: true},
	{Name: "apt.mark_manual", Args: []string{"name"}, Mutating: true},

	// ── dnf ─────────────────────────────────────────────────────────
	{Name: "dnf.install", Args: []string{"name", "version"}, Mutating: true},
	{Name: "dnf.remove", Args: []string{"name"}, Mutating: true},
	{Name: "dnf.update", Args: []string{"name"}, Mutating: true},
	{Name: "dnf.info", Args: []string{"name"}},
	{Name: "dnf.list"},
	{Name: "dnf.search", Args: []string{"name"}},
	{Name: "dnf.clean", Mutating: true},
	{Name: "dnf.repolist"},
	{Name: "dnf.grouplist"},
	{Name: "dnf.groupinstall", Args: []string{"name"}, Mutating: true},
	{Name: "dnf.groupremove", Args: []string{"name"}, Mutating: true},
	{Name: "dnf.history", Args: []string{"count"}},
	{Name: "dnf.check_update"},
	{Name: "dnf.modulelist"},
	{Name: "dnf.module_enable", Args: []string{"spec"}, Mutating: true},

	// ── apk ─────────────────────────────────────────────────────────
	{Name: "apk.install", Args: []string{"name", "version"}, Mutating: true},
	{Name: "apk.remove", Args: []string{"name", "purge"}, Mutating: true},
	{Name: "apk.update", Mutating: true},
	{Name: "apk.upgrade", Args: []string{"name"}, Mutating: true},
	{Name: "apk.info", Args: []string{"name"}},
	{Name: "apk.list"},
	{Name: "apk.search", Args: []string{"name"}},
	{Name: "apk.cache", Mutating: true},
	{Name: "apk.upgrade_available"},
	{Name: "apk.repository"},

	// ── sysvinit ──────────────────────────────────────────────────────
	{Name: "sysvinit.status", Args: []string{"name"}},
	{Name: "sysvinit.start", Args: []string{"name"}, Mutating: true},
	{Name: "sysvinit.stop", Args: []string{"name"}, Mutating: true},
	{Name: "sysvinit.restart", Args: []string{"name"}, Mutating: true},
	{Name: "sysvinit.reload", Args: []string{"name"}, Mutating: true},
	{Name: "sysvinit.enable", Args: []string{"name", "runlevels"}, Mutating: true},
	{Name: "sysvinit.disable", Args: []string{"name"}, Mutating: true},
	{Name: "sysvinit.list"},

	// ── dpkg_selections ───────────────────────────────────────────────
	{Name: "dpkg_selections.set", Args: []string{"name", "state"}, Mutating: true},
	{Name: "dpkg_selections.get", Args: []string{"name"}},
	{Name: "dpkg_selections.list"},
	{Name: "dpkg_selections.hold", Args: []string{"name"}, Mutating: true},
	{Name: "dpkg_selections.unhold", Args: []string{"name"}, Mutating: true},

	// ── homebrew ──────────────────────────────────────────────────────
	{Name: "homebrew.install", Args: []string{"name", "cask"}, Mutating: true},
	{Name: "homebrew.remove", Args: []string{"name", "cask"}, Mutating: true},
	{Name: "homebrew.upgrade", Args: []string{"name"}, Mutating: true},
	{Name: "homebrew.update", Mutating: true},
	{Name: "homebrew.info", Args: []string{"name"}},
	{Name: "homebrew.list"},
	{Name: "homebrew.list_casks"},
	{Name: "homebrew.outdated"},
	{Name: "homebrew.clean", Mutating: true},
	{Name: "homebrew.tap", Args: []string{"name"}, Mutating: true},
	{Name: "homebrew.untap", Args: []string{"name"}, Mutating: true},
	{Name: "homebrew.list_taps"},
	{Name: "homebrew.doctor"},

	// ── apt_repo ──────────────────────────────────────────────────────
	{Name: "apt_repo.list"},
	{Name: "apt_repo.exists", Args: []string{"uri"}},
	{Name: "apt_repo.add", Args: []string{"uri", "dist", "components"}, Mutating: true},
	{Name: "apt_repo.remove", Args: []string{"uri"}, Mutating: true},
	{Name: "apt_repo.update", Mutating: true},

	// ── disk ──────────────────────────────────────────────────────────
	{Name: "disk.filesystem", Args: []string{"device", "fs_type"}, Mutating: true},
	{Name: "disk.part_list", Args: []string{"device"}},

	// ── docker ────────────────────────────────────────────────────────
	{Name: "docker.container_list", Args: []string{"all"}},
	{Name: "docker.container_exists", Args: []string{"name"}},
	{Name: "docker.container_run", Args: []string{"name", "image", "opts"}, Mutating: true},
	{Name: "docker.container_stop", Args: []string{"name"}, Mutating: true},
	{Name: "docker.container_remove", Args: []string{"name", "force"}, Mutating: true},
	{Name: "docker.image_list"},
	{Name: "docker.image_pull", Args: []string{"image"}, Mutating: true},
	{Name: "docker.image_remove", Args: []string{"image", "force"}, Mutating: true},

	// ── file ──────────────────────────────────────────────────────────
	{Name: "file.append", Args: []string{"path", "content"}, Mutating: true},
	{Name: "file.blockinfile", Args: []string{"path", "marker", "content", "present", "insert_after", "insert_before"}, Mutating: true},
	{Name: "file.checksum", Args: []string{"path", "algo"}},
	{Name: "file.chmod", Args: []string{"path", "mode"}, Mutating: true},
	{Name: "file.collect", Args: []string{"source", "targets", "options"}, Avail: ControllerOnly, Mutating: true},
	{Name: "file.copy", Args: []string{"src", "dst"}, Mutating: true},
	{Name: "file.delete", Args: []string{"path"}, Mutating: true},
	// distribute/collect fan out from the controller over SSH; a remote
	// runner executing them would need controller credentials.
	{Name: "file.distribute", Args: []string{"source", "targets", "options"}, Avail: ControllerOnly, Mutating: true},
	{Name: "file.exists", Args: []string{"path"}},
	{Name: "file.find", Args: []string{"paths", "patterns", "regex", "file_type", "max_depth", "age", "size"}},
	{Name: "file.ini_get", Args: []string{"path", "section", "key"}},
	{Name: "file.ini_set", Args: []string{"path", "section", "key", "value"}, Mutating: true},
	{Name: "file.lineinfile", Args: []string{"path", "line", "present", "regexp"}, Mutating: true},
	{Name: "file.list", Args: []string{"dir"}},
	{Name: "file.mkdir", Args: []string{"path"}, Mutating: true},
	{Name: "file.move", Args: []string{"src", "dst"}, Mutating: true},
	{Name: "file.read", Args: []string{"path"}},
	{Name: "file.replace", Args: []string{"path", "pattern", "replacement", "after", "before"}, Mutating: true},
	{Name: "file.stat", Args: []string{"path"}},
	// file.template only READS the template and returns the rendered text;
	// it never writes a file, so it is not mutating.
	{Name: "file.template", Args: []string{"path", "vars"}},
	{Name: "file.write", Args: []string{"path", "content"}, Mutating: true},

	// ── firewall ──────────────────────────────────────────────────────
	{Name: "firewall.rule", Args: []string{"action", "protocol", "port", "source"}, Mutating: true},

	// ── firewalld ─────────────────────────────────────────────────────
	{Name: "firewalld.get"},
	{Name: "firewalld.start", Mutating: true},
	{Name: "firewalld.stop", Mutating: true},
	{Name: "firewalld.restart", Mutating: true},
	{Name: "firewalld.enable", Mutating: true},
	{Name: "firewalld.disable", Mutating: true},
	{Name: "firewalld.list_zones"},
	{Name: "firewalld.reload", Mutating: true},

	// ── git ───────────────────────────────────────────────────────────
	{Name: "git.clone", Args: []string{"url", "dest", "opts"}, Mutating: true},
	{Name: "git.pull", Args: []string{"repo_path", "remote", "branch"}, Mutating: true},

	// ── group ─────────────────────────────────────────────────────────
	{Name: "group.add", Args: []string{"name", "opts"}, Mutating: true},
	{Name: "group.exists", Args: []string{"name"}},
	{Name: "group.info", Args: []string{"name"}},
	{Name: "group.list"},
	{Name: "group.remove", Args: []string{"name"}, Mutating: true},

	// ── hosts ─────────────────────────────────────────────────────────
	{Name: "hosts.list"},
	{Name: "hosts.exists", Args: []string{"hostname"}},
	{Name: "hosts.add", Args: []string{"ip", "hostnames"}, Mutating: true},
	{Name: "hosts.remove", Args: []string{"hostnames"}, Mutating: true},

	// ── json ──────────────────────────────────────────────────────────
	{Name: "json.decode", Args: []string{"input"}},
	{Name: "json.encode", Args: []string{"value"}},

	// ── known_hosts ───────────────────────────────────────────────────
	{Name: "known_hosts.list"},
	{Name: "known_hosts.check", Args: []string{"host"}},
	{Name: "known_hosts.add", Args: []string{"host"}, Mutating: true},
	{Name: "known_hosts.remove", Args: []string{"host"}, Mutating: true},

	// ── kernel ────────────────────────────────────────────────────────
	{Name: "kernel.module_list"},
	{Name: "kernel.module_load", Args: []string{"name"}, Mutating: true},
	{Name: "kernel.module_unload", Args: []string{"name"}, Mutating: true},

	// ── limits ────────────────────────────────────────────────────────
	{Name: "limits.list"},
	{Name: "limits.get", Args: []string{"domain"}},
	{Name: "limits.set", Args: []string{"domain", "type", "item", "value"}, Mutating: true},
	{Name: "limits.remove", Args: []string{"domain"}, Mutating: true},

	// ── locale ────────────────────────────────────────────────────────
	{Name: "locale.get"},
	{Name: "locale.available"},
	{Name: "locale.set", Args: []string{"locale"}, Mutating: true},

	// ── logrotate ─────────────────────────────────────────────────────
	{Name: "logrotate.list"},
	{Name: "logrotate.get", Args: []string{"name"}},
	{Name: "logrotate.set", Args: []string{"name", "pattern", "frequency", "rotate", "compress", "post_rotate"}, Mutating: true},
	{Name: "logrotate.remove", Args: []string{"name"}, Mutating: true},

	// ── lvg ───────────────────────────────────────────────────────────
	{Name: "lvg.create", Args: []string{"name", "pvs"}, Mutating: true},
	{Name: "lvg.remove", Args: []string{"name"}, Mutating: true},
	{Name: "lvg.extend", Args: []string{"name", "pvs"}, Mutating: true},
	{Name: "lvg.reduce", Args: []string{"name", "pvs"}, Mutating: true},
	{Name: "lvg.activate", Args: []string{"name"}, Mutating: true},
	{Name: "lvg.deactivate", Args: []string{"name"}, Mutating: true},
	{Name: "lvg.list"},
	{Name: "lvg.get", Args: []string{"name"}},

	// ── net ───────────────────────────────────────────────────────────
	{Name: "net.dns_lookup", Args: []string{"host"}},
	{Name: "net.download", Args: []string{"url", "dest", "checksum_algo", "checksum_expected"}, Mutating: true},
	{Name: "net.http_get", Args: []string{"url"}},
	// net.http_post is classified as mutating even though it only changes
	// REMOTE state: a POST submits data (deploys, webhooks, form
	// submissions) and is the writing counterpart to the read-only GET.
	{Name: "net.http_post", Args: []string{"url", "body"}, Mutating: true},
	{Name: "net.interfaces"},
	{Name: "net.tcp_check", Args: []string{"host", "port"}},
	{Name: "net.wait_for", Args: []string{"host", "port", "timeout"}},
	{Name: "net.wait_for_connection", Args: []string{"host", "port", "timeout"}},

	// ── ntp ───────────────────────────────────────────────────────────
	{Name: "ntp.get"},
	{Name: "ntp.set", Args: []string{"server"}, Mutating: true},

	// ── pkg ───────────────────────────────────────────────────────────
	{Name: "pkg.info", Args: []string{"name"}},
	{Name: "pkg.install", Args: []string{"name"}, Mutating: true},
	{Name: "pkg.list"},
	{Name: "pkg.remove", Args: []string{"name"}, Mutating: true},

	// ── pip ───────────────────────────────────────────────────────────
	{Name: "pip.list"},
	{Name: "pip.exists", Args: []string{"name"}},
	{Name: "pip.install", Args: []string{"name", "version"}, Mutating: true},
	{Name: "pip.uninstall", Args: []string{"name"}, Mutating: true},

	// ── process ───────────────────────────────────────────────────────
	{Name: "process.exec", Args: []string{"command", "args"}, Mutating: true},
	{Name: "process.find_by_name", Args: []string{"name"}},
	{Name: "process.find_by_port", Args: []string{"port"}},
	{Name: "process.kill", Args: []string{"pid", "signal"}, Mutating: true},
	{Name: "process.list"},

	// ── resolv ────────────────────────────────────────────────────────
	{Name: "resolv.get"},
	{Name: "resolv.set", Args: []string{"nameservers", "search", "options", "domain"}, Mutating: true},
	{Name: "resolv.add_nameserver", Args: []string{"nameserver"}, Mutating: true},
	{Name: "resolv.remove_nameserver", Args: []string{"nameserver"}, Mutating: true},

	// ── service ───────────────────────────────────────────────────────
	{Name: "service.disable", Args: []string{"name"}, Mutating: true},
	{Name: "service.enable", Args: []string{"name"}, Mutating: true},
	{Name: "service.restart", Args: []string{"name"}, Mutating: true},
	{Name: "service.start", Args: []string{"name"}, Mutating: true},
	{Name: "service.status", Args: []string{"name"}},
	{Name: "service.stop", Args: []string{"name"}, Mutating: true},

	// ── snap ──────────────────────────────────────────────────────────
	{Name: "snap.install", Args: []string{"name", "channel", "classic"}, Mutating: true},
	{Name: "snap.remove", Args: []string{"name"}, Mutating: true},
	{Name: "snap.refresh", Args: []string{"name", "channel"}, Mutating: true},
	{Name: "snap.list"},
	{Name: "snap.get", Args: []string{"name"}},
	{Name: "snap.enable", Args: []string{"name"}, Mutating: true},
	{Name: "snap.disable", Args: []string{"name"}, Mutating: true},
	{Name: "snap.switch", Args: []string{"name", "channel"}, Mutating: true},
	{Name: "snap.changes"},

	// ── flatpak ──────────────────────────────────────────────────────
	{Name: "flatpak.install", Args: []string{"name", "from", "user"}, Mutating: true},
	{Name: "flatpak.remove", Args: []string{"name", "user"}, Mutating: true},
	{Name: "flatpak.update", Args: []string{"name", "user"}, Mutating: true},
	{Name: "flatpak.list", Args: []string{"user"}},
	{Name: "flatpak.info", Args: []string{"name", "user"}},
	{Name: "flatpak.run", Args: []string{"name", "args", "user"}, Mutating: true},
	{Name: "flatpak.repair", Args: []string{"user"}, Mutating: true},

	// ── zfs ────────────────────────────────────────────────────────────
	{Name: "zfs.create", Args: []string{"name", "properties"}, Mutating: true},
	{Name: "zfs.destroy", Args: []string{"name", "recursive"}, Mutating: true},
	{Name: "zfs.set", Args: []string{"name", "property", "value"}, Mutating: true},
	{Name: "zfs.get", Args: []string{"name", "property"}},
	{Name: "zfs.list"},
	{Name: "zfs.exists", Args: []string{"name"}},
	{Name: "zfs.list_pools"},
	{Name: "zfs.get_pool_status", Args: []string{"name"}},
	{Name: "zfs.snapshot", Args: []string{"name", "snapshot_name"}, Mutating: true},
	{Name: "zfs.destroy_snapshot", Args: []string{"name", "snapshot_name"}, Mutating: true},

	// ── nmcli ──────────────────────────────────────────────────────────
	{Name: "nmcli.add", Args: []string{"name", "type", "settings"}, Mutating: true},
	{Name: "nmcli.modify", Args: []string{"name", "settings"}, Mutating: true},
	{Name: "nmcli.delete", Args: []string{"name"}, Mutating: true},
	{Name: "nmcli.up", Args: []string{"name"}, Mutating: true},
	{Name: "nmcli.down", Args: []string{"name"}, Mutating: true},
	{Name: "nmcli.list"},
	{Name: "nmcli.show", Args: []string{"name"}},
	{Name: "nmcli.list_devices"},
	{Name: "nmcli.reload", Mutating: true},
	{Name: "nmcli.get_general_status"},

	// ── crypttab ───────────────────────────────────────────────────────
	{Name: "crypttab.add", Args: []string{"name", "device", "key_file", "options"}, Mutating: true},
	{Name: "crypttab.remove", Args: []string{"name"}, Mutating: true},
	{Name: "crypttab.modify", Args: []string{"name", "device", "key_file", "options"}, Mutating: true},
	{Name: "crypttab.get", Args: []string{"name"}},
	{Name: "crypttab.list"},
	{Name: "crypttab.exists", Args: []string{"name"}},
	{Name: "crypttab.validate"},
	{Name: "crypttab.backup", Args: []string{"backup_dir"}, Mutating: true},

	// ── sysfs ────────────────────────────────────────────────────────────
	{Name: "sysfs.read", Args: []string{"path"}},
	{Name: "sysfs.write", Args: []string{"path", "value"}, Mutating: true},
	{Name: "sysfs.exists", Args: []string{"path"}},
	{Name: "sysfs.get", Args: []string{"path"}},
	{Name: "sysfs.list", Args: []string{"dir_path"}},
	{Name: "sysfs.set_device_power", Args: []string{"device_path", "state"}, Mutating: true},
	{Name: "sysfs.get_device_power", Args: []string{"device_path"}},
	{Name: "sysfs.set_kernel_parameter", Args: []string{"param", "value"}, Mutating: true},
	{Name: "sysfs.get_kernel_parameter", Args: []string{"param"}},

	// ── pamd ──────────────────────────────────────────────────────────
	{Name: "pamd.get", Args: []string{"service"}},
	{Name: "pamd.list"},
	{Name: "pamd.add_rule", Args: []string{"service", "type", "control", "module", "args"}, Mutating: true},
	{Name: "pamd.remove_rule", Args: []string{"service", "type", "module"}, Mutating: true},
	{Name: "pamd.modify_rule", Args: []string{"service", "type", "module", "new_control", "new_args"}, Mutating: true},
	{Name: "pamd.validate", Args: []string{"service"}},
	{Name: "pamd.backup", Args: []string{"service", "backup_dir"}},

	// ── getent ────────────────────────────────────────────────────────
	{Name: "getent.passwd"},
	{Name: "getent.lookup_user", Args: []string{"key"}},
	{Name: "getent.groups"},
	{Name: "getent.lookup_group", Args: []string{"key"}},
	{Name: "getent.services"},
	{Name: "getent.lookup_service", Args: []string{"key"}},
	{Name: "getent.protocols"},
	{Name: "getent.lookup_protocol", Args: []string{"key"}},
	{Name: "getent.shells"},

	// ── haproxy ───────────────────────────────────────────────────────
	{Name: "haproxy.get_status"},
	{Name: "haproxy.list_backends", Args: []string{"socket"}},
	{Name: "haproxy.enable_backend", Args: []string{"backend", "server", "socket"}, Mutating: true},
	{Name: "haproxy.disable_backend", Args: []string{"backend", "server", "socket"}, Mutating: true},
	{Name: "haproxy.validate_config", Args: []string{"config_file"}},
	{Name: "haproxy.reload", Args: []string{"config_file"}, Mutating: true},
	{Name: "haproxy.restart", Mutating: true},
	{Name: "haproxy.version"},

	// ── openssl_cert ────────────────────────────────────────────────────
	{Name: "openssl_cert.create_csr", Args: []string{"key_path", "csr_path", "subject", "key_bits"}, Mutating: true},
	{Name: "openssl_cert.generate_self_signed", Args: []string{"cert_path", "key_path", "subject", "days", "key_bits"}, Mutating: true},
	{Name: "openssl_cert.inspect", Args: []string{"cert_path"}},
	{Name: "openssl_cert.verify", Args: []string{"cert_path", "ca_path"}},
	{Name: "openssl_cert.check_expiry", Args: []string{"cert_path"}},
	{Name: "openssl_cert.convert_format", Args: []string{"input_path", "output_path", "output_format"}, Mutating: true},

	// ── redis ───────────────────────────────────────────────────────────
	{Name: "redis.ping", Args: []string{"host", "port", "auth"}},
	{Name: "redis.get", Args: []string{"key", "host", "port", "auth"}},
	{Name: "redis.set", Args: []string{"key", "value", "host", "port", "auth", "expiry_sec"}, Mutating: true},
	{Name: "redis.del", Args: []string{"keys", "host", "port", "auth"}, Mutating: true},
	{Name: "redis.keys", Args: []string{"pattern", "host", "port", "auth"}},
	{Name: "redis.info", Args: []string{"host", "port", "auth"}},
	{Name: "redis.flush_db", Mutating: true},

	// ── gem ─────────────────────────────────────────────────────────────
	{Name: "gem.install", Args: []string{"name", "version", "user_install"}, Mutating: true},
	{Name: "gem.uninstall", Args: []string{"name", "force"}, Mutating: true},
	{Name: "gem.update", Args: []string{"name"}, Mutating: true},
	{Name: "gem.info", Args: []string{"name"}},
	{Name: "gem.list"},
	{Name: "gem.version"},

	// ── rabbitmq ──────────────────────────────────────────────────────
	{Name: "rabbitmq.add_vhost", Args: []string{"name"}, Mutating: true},
	{Name: "rabbitmq.delete_vhost", Args: []string{"name"}, Mutating: true},
	{Name: "rabbitmq.list_vhosts"},
	{Name: "rabbitmq.add_user", Args: []string{"name", "password", "tags"}, Mutating: true},
	{Name: "rabbitmq.delete_user", Args: []string{"name"}, Mutating: true},
	{Name: "rabbitmq.set_user_tags", Args: []string{"name", "tags"}, Mutating: true},
	{Name: "rabbitmq.list_users"},
	{Name: "rabbitmq.set_permission", Args: []string{"user", "vhost", "configure", "write", "read"}, Mutating: true},
	{Name: "rabbitmq.clear_permission", Args: []string{"user", "vhost"}, Mutating: true},
	{Name: "rabbitmq.set_policy", Args: []string{"name", "vhost", "pattern", "definition", "apply_to"}, Mutating: true},
	{Name: "rabbitmq.delete_policy", Args: []string{"name", "vhost"}, Mutating: true},
	{Name: "rabbitmq.declare_queue", Args: []string{"name", "vhost", "queue_type", "durable", "auto_delete"}, Mutating: true},
	{Name: "rabbitmq.delete_queue", Args: []string{"name", "vhost"}, Mutating: true},
	{Name: "rabbitmq.declare_exchange", Args: []string{"name", "vhost", "type", "durable", "auto_delete"}, Mutating: true},
	{Name: "rabbitmq.delete_exchange", Args: []string{"name", "vhost"}, Mutating: true},
	{Name: "rabbitmq.bind_queue", Args: []string{"queue", "exchange", "vhost", "routing_key"}, Mutating: true},
	{Name: "rabbitmq.unbind_queue", Args: []string{"queue", "exchange", "vhost", "routing_key"}, Mutating: true},
	{Name: "rabbitmq.get_status"},

	// ── consul ────────────────────────────────────────────────────────
	{Name: "consul.kv_get", Args: []string{"key", "addr"}},
	{Name: "consul.kv_put", Args: []string{"key", "value", "addr"}, Mutating: true},
	{Name: "consul.kv_delete", Args: []string{"key", "addr"}, Mutating: true},
	{Name: "consul.kv_list", Args: []string{"prefix", "addr"}},
	{Name: "consul.service_register", Args: []string{"name", "id", "addr", "port", "consul_addr"}, Mutating: true},
	{Name: "consul.service_deregister", Args: []string{"id", "consul_addr"}, Mutating: true},
	{Name: "consul.members", Args: []string{"addr"}},
	{Name: "consul.info", Args: []string{"addr"}},
	{Name: "consul.health_check", Args: []string{"service", "addr"}},
	{Name: "consul.version"},

	// ── memcached ─────────────────────────────────────────────────────
	{Name: "memcached.get", Args: []string{"key", "host", "port"}},
	{Name: "memcached.set", Args: []string{"key", "value", "host", "port", "expiry"}, Mutating: true},
	{Name: "memcached.delete", Args: []string{"key", "host", "port"}, Mutating: true},
	{Name: "memcached.flush_all", Args: []string{"host", "port"}, Mutating: true},
	{Name: "memcached.stats", Args: []string{"host", "port"}},
	{Name: "memcached.version", Args: []string{"host", "port"}},

	// ── selinux ───────────────────────────────────────────────────────
	{Name: "selinux.get"},
	{Name: "selinux.set", Args: []string{"mode"}, Mutating: true},

	// ── ssh ───────────────────────────────────────────────────────────
	{Name: "ssh.authorized_key_add", Args: []string{"user", "key", "exclusive"}, Mutating: true},
	{Name: "ssh.authorized_key_list", Args: []string{"user"}},
	{Name: "ssh.authorized_key_remove", Args: []string{"user", "key"}, Mutating: true},

	// ── sys ───────────────────────────────────────────────────────────
	{Name: "sys.cpu.count"},
	{Name: "sys.cpu.info"},
	{Name: "sys.cpu.usage"},
	{Name: "sys.disk.partitions"},
	{Name: "sys.disk.usage", Args: []string{"path"}},
	{Name: "sys.ethtool", Args: []string{"iface"}},
	{Name: "sys.hostname"},
	{Name: "sys.hostname_set", Args: []string{"name"}, Mutating: true},
	{Name: "sys.ip_route"},
	{Name: "sys.list_mounts"},
	{Name: "sys.load"},
	{Name: "sys.lsusb"},
	{Name: "sys.memory.info"},
	{Name: "sys.mount", Args: []string{"device", "mountpoint", "fs_type", "opts"}, Mutating: true},
	{Name: "sys.net.interfaces"},
	{Name: "sys.os"},
	{Name: "sys.reboot", Mutating: true},
	{Name: "sys.timezone_get"},
	{Name: "sys.timezone_set", Args: []string{"timezone"}, Mutating: true},
	{Name: "sys.unmount", Args: []string{"mountpoint"}, Mutating: true},
	{Name: "sys.uptime"},
	{Name: "sys.users"},

	// ── svn ───────────────────────────────────────────────────────────────
	{Name: "svn.checkout", Args: []string{"url", "dest", "revision", "force"}, Mutating: true},
	{Name: "svn.cleanup", Args: []string{"dest"}, Mutating: true},
	{Name: "svn.export", Args: []string{"url", "dest", "revision", "force"}, Mutating: true},
	{Name: "svn.info", Args: []string{"dest"}},
	{Name: "svn.revert", Args: []string{"dest", "recursive"}, Mutating: true},
	{Name: "svn.status", Args: []string{"dest"}},
	{Name: "svn.update", Args: []string{"dest", "revision"}, Mutating: true},

	// ── sysctl ────────────────────────────────────────────────────────
	{Name: "sysctl.get", Args: []string{"name"}},
	{Name: "sysctl.list"},
	{Name: "sysctl.set", Args: []string{"name", "value"}, Mutating: true},

	// ── time ──────────────────────────────────────────────────────────
	{Name: "time.diff", Args: []string{"t1", "t2"}},
	{Name: "time.format", Args: []string{"ts", "layout"}},
	{Name: "time.now"},
	{Name: "time.parse", Args: []string{"layout", "value"}},
	{Name: "time.since", Args: []string{"ts"}},
	{Name: "time.sleep", Args: []string{"ms"}},

	// ── user ──────────────────────────────────────────────────────────
	{Name: "user.add", Args: []string{"username", "opts"}, Mutating: true},
	{Name: "user.exists", Args: []string{"username"}},
	{Name: "user.info", Args: []string{"username"}},
	{Name: "user.list"},
	{Name: "user.modify", Args: []string{"username", "opts"}, Mutating: true},
	{Name: "user.remove", Args: []string{"username", "remove_home"}, Mutating: true},

	// ── yaml ──────────────────────────────────────────────────────────
	{Name: "yaml.decode", Args: []string{"input"}},
	{Name: "yaml.encode", Args: []string{"value"}},

	// ── yum_repo ──────────────────────────────────────────────────────
	{Name: "yum_repo.list"},
	{Name: "yum_repo.exists", Args: []string{"id"}},
	{Name: "yum_repo.add", Args: []string{"id", "name", "base_url", "gpg_check", "gpg_key"}, Mutating: true},
	{Name: "yum_repo.remove", Args: []string{"id"}, Mutating: true},

	// ── zypper ─────────────────────────────────────────────────────────────
	{Name: "zypper.clean", Mutating: true},
	{Name: "zypper.dist_upgrade", Mutating: true},
	{Name: "zypper.info", Args: []string{"name"}},
	{Name: "zypper.install", Args: []string{"name", "version"}, Mutating: true},
	{Name: "zypper.list"},
	{Name: "zypper.patch", Mutating: true},
	{Name: "zypper.pattern_install", Args: []string{"name"}, Mutating: true},
	{Name: "zypper.pattern_remove", Args: []string{"name"}, Mutating: true},
	{Name: "zypper.refresh", Mutating: true},
	{Name: "zypper.remove", Args: []string{"name"}, Mutating: true},
	{Name: "zypper.repo_add", Args: []string{"name", "url"}, Mutating: true},
	{Name: "zypper.repo_list"},
	{Name: "zypper.repo_remove", Args: []string{"name"}, Mutating: true},
	{Name: "zypper.search", Args: []string{"name"}},
	{Name: "zypper.update", Args: []string{"name"}, Mutating: true},

	// ── ufw ───────────────────────────────────────────────────────────
	{Name: "ufw.status"},
	{Name: "ufw.list"},
	{Name: "ufw.enable", Mutating: true},
	{Name: "ufw.disable", Mutating: true},
	{Name: "ufw.allow", Args: []string{"port", "proto"}, Mutating: true},
	{Name: "ufw.deny", Args: []string{"port", "proto"}, Mutating: true},
	{Name: "ufw.delete", Args: []string{"number"}, Mutating: true},
	{Name: "ufw.reset", Mutating: true},
	{Name: "ufw.reload", Mutating: true},

	// ── ini_file ──────────────────────────────────────────────────────
	{Name: "ini_file.sections", Args: []string{"path"}},
	{Name: "ini_file.get", Args: []string{"path", "section", "key"}},
	{Name: "ini_file.set", Args: []string{"path", "section", "key", "value"}, Mutating: true},
	{Name: "ini_file.remove", Args: []string{"path", "section", "key"}, Mutating: true},
	{Name: "ini_file.remove_section", Args: []string{"path", "section"}, Mutating: true},

	// ── mount ─────────────────────────────────────────────────────────
	{Name: "mount.list"},
	{Name: "mount.mount", Args: []string{"device", "mountpoint", "fstype", "options"}, Mutating: true},
	{Name: "mount.umount", Args: []string{"mountpoint"}, Mutating: true},
	{Name: "mount.fstab"},
	{Name: "mount.add_fstab", Args: []string{"device", "mountpoint", "fstype", "options"}, Mutating: true},
	{Name: "mount.remove_fstab", Args: []string{"target"}, Mutating: true},

	// ── hostname ──────────────────────────────────────────────────────
	{Name: "hostname.get"},
	{Name: "hostname.set", Args: []string{"hostname"}, Mutating: true},
	{Name: "hostname.set_fqdn", Args: []string{"fqdn"}, Mutating: true},

	// ── timezone ──────────────────────────────────────────────────────
	{Name: "timezone.get"},
	{Name: "timezone.set", Args: []string{"timezone"}, Mutating: true},
	{Name: "timezone.list"},

	// ── iptables ──────────────────────────────────────────────────────
	{Name: "iptables.list", Args: []string{"chain"}},
	{Name: "iptables.flush", Args: []string{"table"}, Mutating: true},
	{Name: "iptables.add_rule", Args: []string{"chain", "rule_spec"}, Mutating: true},
	{Name: "iptables.delete_rule", Args: []string{"chain", "number"}, Mutating: true},
	{Name: "iptables.save"},
	{Name: "iptables.list_chains"},

	// ── npm ───────────────────────────────────────────────────────────
	{Name: "npm.list", Args: []string{"global"}},
	{Name: "npm.install", Args: []string{"name", "global"}, Mutating: true},
	{Name: "npm.uninstall", Args: []string{"name", "global"}, Mutating: true},
	{Name: "npm.outdated", Args: []string{"global"}},

	// ── mysql ─────────────────────────────────────────────────────────
	{Name: "mysql.databases"},
	{Name: "mysql.create_database", Args: []string{"name"}, Mutating: true},
	{Name: "mysql.drop_database", Args: []string{"name"}, Mutating: true},
	{Name: "mysql.users"},
	{Name: "mysql.create_user", Args: []string{"user", "host", "password"}, Mutating: true},
	{Name: "mysql.drop_user", Args: []string{"user", "host"}, Mutating: true},
	{Name: "mysql.grant", Args: []string{"privileges", "database", "user", "host"}, Mutating: true},

	// ── nginx ─────────────────────────────────────────────────────────
	{Name: "nginx.config_test"},
	{Name: "nginx.reload", Mutating: true},
	{Name: "nginx.sites_list"},
	{Name: "nginx.site_enable", Args: []string{"name"}, Mutating: true},
	{Name: "nginx.site_disable", Args: []string{"name"}, Mutating: true},

	// ── modprobe ──────────────────────────────────────────────────────
	{Name: "modprobe.list"},
	{Name: "modprobe.load", Args: []string{"name"}, Mutating: true},
	{Name: "modprobe.unload", Args: []string{"name"}, Mutating: true},
	{Name: "modprobe.is_loaded", Args: []string{"name"}},

	// ── alternatives ──────────────────────────────────────────────────
	{Name: "alternatives.list", Args: []string{"name"}},
	{Name: "alternatives.display", Args: []string{"name"}},
	{Name: "alternatives.set", Args: []string{"name", "path"}, Mutating: true},
	{Name: "alternatives.install", Args: []string{"name", "link", "path", "priority"}, Mutating: true},
	{Name: "alternatives.remove", Args: []string{"name", "path"}, Mutating: true},

	// ── blockdev ──────────────────────────────────────────────────────
	{Name: "blockdev.list"},
	{Name: "blockdev.info", Args: []string{"device"}},
	{Name: "blockdev.flush_buffers", Args: []string{"device"}, Mutating: true},
	{Name: "blockdev.set_readahead", Args: []string{"device", "value"}, Mutating: true},

	// ── at ────────────────────────────────────────────────────────────
	{Name: "at.list"},
	{Name: "at.schedule", Args: []string{"command", "time_spec"}, Mutating: true},
	{Name: "at.remove", Args: []string{"job_id"}, Mutating: true},

	// ── postgresql ─────────────────────────────────────────────────────
	{Name: "postgresql.databases"},
	{Name: "postgresql.create_database", Args: []string{"name"}, Mutating: true},
	{Name: "postgresql.drop_database", Args: []string{"name"}, Mutating: true},
	{Name: "postgresql.users"},
	{Name: "postgresql.create_user", Args: []string{"user", "password"}, Mutating: true},
	{Name: "postgresql.drop_user", Args: []string{"user"}, Mutating: true},
	{Name: "postgresql.grant", Args: []string{"privileges", "database", "user"}, Mutating: true},

	// ── apache2 ────────────────────────────────────────────────────────
	{Name: "apache2.config_test"},
	{Name: "apache2.reload", Mutating: true},
	{Name: "apache2.sites_list"},
	{Name: "apache2.site_enable", Args: []string{"name"}, Mutating: true},
	{Name: "apache2.site_disable", Args: []string{"name"}, Mutating: true},
	{Name: "apache2.modules_list"},
	{Name: "apache2.module_enable", Args: []string{"name"}, Mutating: true},
	{Name: "apache2.module_disable", Args: []string{"name"}, Mutating: true},

	// ── filesystem ─────────────────────────────────────────────────────
	{Name: "filesystem.mkfs", Args: []string{"device", "fstype", "label"}, Mutating: true},
	{Name: "filesystem.resize_ext4", Args: []string{"device"}, Mutating: true},
	{Name: "filesystem.resize_xfs", Args: []string{"mountpoint"}, Mutating: true},
	{Name: "filesystem.check", Args: []string{"device"}},

	// ── parted ─────────────────────────────────────────────────────────
	{Name: "parted.list", Args: []string{"device"}},
	{Name: "parted.mklabel", Args: []string{"device", "label_type"}, Mutating: true},
	{Name: "parted.mkpart", Args: []string{"device", "part_type", "fstype", "start", "end"}, Mutating: true},
	{Name: "parted.rm", Args: []string{"device", "number"}, Mutating: true},

	// ── acl ────────────────────────────────────────────────────────────
	{Name: "acl.get", Args: []string{"path"}},
	{Name: "acl.set", Args: []string{"path", "entry", "recursive"}, Mutating: true},
	{Name: "acl.remove", Args: []string{"path", "entry", "recursive"}, Mutating: true},
	{Name: "acl.remove_all", Args: []string{"path", "recursive"}, Mutating: true},

	// ── wait_for ───────────────────────────────────────────────────────
	{Name: "wait_for.port", Args: []string{"host", "port", "timeout_ms"}},
	{Name: "wait_for.file", Args: []string{"path", "timeout_ms"}},
	{Name: "wait_for.url", Args: []string{"url", "timeout_ms"}},

	// ── lvol ───────────────────────────────────────────────────────────
	{Name: "lvol.list"},
	{Name: "lvol.vg_list"},
	{Name: "lvol.create", Args: []string{"name", "vg", "size"}, Mutating: true},
	{Name: "lvol.remove", Args: []string{"name", "vg"}, Mutating: true},
	{Name: "lvol.resize", Args: []string{"name", "vg", "size"}, Mutating: true},

	// ── synchronize ────────────────────────────────────────────────────
	{Name: "synchronize.sync", Args: []string{"source", "dest", "delete", "compress"}, Mutating: true},

	// ── fetch ──────────────────────────────────────────────────────────
	{Name: "fetch.file", Args: []string{"source", "dest"}, Mutating: true},
	{Name: "fetch.url", Args: []string{"url", "dest"}, Mutating: true},

	// ── seboolean ──────────────────────────────────────────────────────
	{Name: "seboolean.list"},
	{Name: "seboolean.get", Args: []string{"name"}},
	{Name: "seboolean.set", Args: []string{"name", "state", "persistent"}, Mutating: true},
	// ── uri ──────────────────────────────────────────────────────────────
	{Name: "uri.do", Args: []string{"url", "method", "headers", "body", "timeout_ms"}},
	{Name: "uri.get", Args: []string{"url"}},
	{Name: "uri.post", Args: []string{"url", "body"}, Mutating: true},
	{Name: "uri.put", Args: []string{"url", "body"}, Mutating: true},
	{Name: "uri.delete", Args: []string{"url"}, Mutating: true},
	{Name: "uri.download", Args: []string{"url", "dest"}, Mutating: true},

	// ── lineinfile ───────────────────────────────────────────────────────
	{Name: "lineinfile.present", Args: []string{"path", "line", "regexp", "create"}, Mutating: true},
	{Name: "lineinfile.absent", Args: []string{"path", "regexp"}, Mutating: true},

	// ── replace ──────────────────────────────────────────────────────────
	{Name: "replace.replace", Args: []string{"path", "pattern", "replacement", "regexp_mode"}, Mutating: true},

	// ── xml ──────────────────────────────────────────────────────────────
	{Name: "xml.get_element", Args: []string{"path", "element"}},
	{Name: "xml.set_element", Args: []string{"path", "element", "value"}, Mutating: true},

	// ── systemd ─────────────────────────────────────────────────────────────
	{Name: "systemd.is_active", Args: []string{"unit"}},
	{Name: "systemd.is_enabled", Args: []string{"unit"}},
	{Name: "systemd.enable", Args: []string{"unit"}, Mutating: true},
	{Name: "systemd.disable", Args: []string{"unit"}, Mutating: true},
	{Name: "systemd.start", Args: []string{"unit"}, Mutating: true},
	{Name: "systemd.stop", Args: []string{"unit"}, Mutating: true},
	{Name: "systemd.restart", Args: []string{"unit"}, Mutating: true},
	{Name: "systemd.reload", Args: []string{"unit"}, Mutating: true},
	{Name: "systemd.daemon_reload", Args: []string{}, Mutating: true},
	{Name: "systemd.mask", Args: []string{"unit"}, Mutating: true},
	{Name: "systemd.unmask", Args: []string{"unit"}, Mutating: true},
	{Name: "systemd.show", Args: []string{"unit"}},
	{Name: "systemd.list", Args: []string{"unit_type"}},

	// ── patch ───────────────────────────────────────────────────────────────
	{Name: "patch.apply", Args: []string{"patch_content", "reverse"}, Mutating: true},
	{Name: "patch.dry_run", Args: []string{"patch_content"}},

	// ── xattr ───────────────────────────────────────────────────────────────
	{Name: "xattr.get", Args: []string{"path", "name"}},
	{Name: "xattr.set", Args: []string{"path", "name", "value"}, Mutating: true},
	{Name: "xattr.remove", Args: []string{"path", "name"}, Mutating: true},
	{Name: "xattr.list", Args: []string{"path"}},

	// ── firewalld_zone ──────────────────────────────────────────────────────
	{Name: "firewalld_zone.get_default", Args: []string{}},
	{Name: "firewalld_zone.set_default", Args: []string{"zone"}, Mutating: true},
	{Name: "firewalld_zone.add_zone", Args: []string{"zone"}, Mutating: true},
	{Name: "firewalld_zone.remove_zone", Args: []string{"zone"}, Mutating: true},
	{Name: "firewalld_zone.add_service", Args: []string{"zone", "service"}, Mutating: true},
	{Name: "firewalld_zone.remove_service", Args: []string{"zone", "service"}, Mutating: true},
	{Name: "firewalld_zone.add_port", Args: []string{"zone", "port_protocol"}, Mutating: true},
	{Name: "firewalld_zone.remove_port", Args: []string{"zone", "port_protocol"}, Mutating: true},
	{Name: "firewalld_zone.add_rich_rule", Args: []string{"zone", "rule"}, Mutating: true},
	{Name: "firewalld_zone.remove_rich_rule", Args: []string{"zone", "rule"}, Mutating: true},
	{Name: "firewalld_zone.info", Args: []string{"zone"}},
	{Name: "firewalld_zone.list_zones", Args: []string{}},

	// ── get_url ─────────────────────────────────────────────────────────────
	{Name: "get_url.download", Args: []string{"url", "dest", "checksum", "force"}, Mutating: true},

	// ── sys utilities ───────────────────────────────────────────────────────
	{Name: "sys.uuid", Args: []string{}},
	{Name: "sys.random_password", Args: []string{"length", "use_special", "use_numbers", "use_uppercase"}},
	{Name: "sys.mac_address", Args: []string{"interface"}},
	{Name: "sys.mac_addresses", Args: []string{}},
	{Name: "sys.dmidecode", Args: []string{}},
	{Name: "sys.lspci", Args: []string{}},
	{Name: "sys.lsblk", Args: []string{}},

	// ── modprobe boot ─────────────────────────────────────────────────────────
	{Name: "modprobe.set_boot", Args: []string{"name", "present"}, Mutating: true},

	// ── seport ─────────────────────────────────────────────────────────────────
	{Name: "seport.add", Args: []string{"seport_type", "protocol", "port"}, Mutating: true},
	{Name: "seport.remove", Args: []string{"protocol", "port"}, Mutating: true},
	{Name: "seport.list", Args: []string{}},
	{Name: "seport.get", Args: []string{"protocol", "port"}},

	// ── sefcontext ─────────────────────────────────────────────────────────────
	{Name: "sefcontext.add", Args: []string{"filespec", "se_type"}, Mutating: true},
	{Name: "sefcontext.modify", Args: []string{"filespec", "se_type"}, Mutating: true},
	{Name: "sefcontext.remove", Args: []string{"filespec"}, Mutating: true},
	{Name: "sefcontext.list", Args: []string{}},
	{Name: "sefcontext.apply", Args: []string{"filespec", "recursive"}, Mutating: true},

	// ── composer ─────────────────────────────────────────────────────────────
	{Name: "composer.install", Args: []string{"dir", "no_dev"}, Mutating: true},
	{Name: "composer.update", Args: []string{"dir", "no_dev"}, Mutating: true},
	{Name: "composer.require", Args: []string{"dir", "package", "version"}, Mutating: true},
	{Name: "composer.remove", Args: []string{"dir", "package"}, Mutating: true},
	{Name: "composer.create_project", Args: []string{"dir", "package", "version"}, Mutating: true},
	{Name: "composer.global_install", Args: []string{"package", "version"}, Mutating: true},
	{Name: "composer.version"},

	// ── cargo ────────────────────────────────────────────────────────────────
	{Name: "cargo.install", Args: []string{"package", "version", "force"}, Mutating: true},
	{Name: "cargo.uninstall", Args: []string{"package"}, Mutating: true},
	{Name: "cargo.update", Args: []string{"package"}, Mutating: true},
	{Name: "cargo.list"},
	{Name: "cargo.build", Args: []string{"dir", "release"}},
	{Name: "cargo.test", Args: []string{"dir"}},
	{Name: "cargo.version"},

	// ── rpmkey ───────────────────────────────────────────────────────────────
	{Name: "rpmkey.import", Args: []string{"key_path"}, Mutating: true},
	{Name: "rpmkey.list"},
	{Name: "rpmkey.remove", Args: []string{"key_id"}, Mutating: true},

	// ── aptkey ───────────────────────────────────────────────────────────────
	{Name: "aptkey.add", Args: []string{"url", "keyring"}, Mutating: true},
	{Name: "aptkey.add_from_key", Args: []string{"path", "keyring"}, Mutating: true},
	{Name: "aptkey.remove", Args: []string{"key_id", "keyring"}, Mutating: true},
	{Name: "aptkey.list"},

	// ── dmidecode ────────────────────────────────────────────────────────────
	{Name: "dmidecode.system"},
	{Name: "dmidecode.bios"},
	{Name: "dmidecode.chassis"},
	{Name: "dmidecode.processor"},
	{Name: "dmidecode.keyword", Args: []string{"keyword"}},

	// ── tuned ────────────────────────────────────────────────────────────────
	{Name: "tuned.set", Args: []string{"profile"}, Mutating: true},
	{Name: "tuned.status"},
	{Name: "tuned.list"},
	{Name: "tuned.off", Mutating: true},
	{Name: "tuned.profile"},
	{Name: "tuned.verify"},

	// ── supervisor ───────────────────────────────────────────────────────────
	{Name: "supervisor.start", Args: []string{"name"}, Mutating: true},
	{Name: "supervisor.stop", Args: []string{"name"}, Mutating: true},
	{Name: "supervisor.restart", Args: []string{"name"}, Mutating: true},
	{Name: "supervisor.reload", Mutating: true},
	{Name: "supervisor.status"},
	{Name: "supervisor.clear_log", Args: []string{"name"}, Mutating: true},
	{Name: "supervisor.reread", Mutating: true},
	{Name: "supervisor.update", Args: []string{"name"}, Mutating: true},

	// ── smartctl ───────────────────────────────────────────────────────────
	{Name: "smartctl.device", Args: []string{"device"}},
	{Name: "smartctl.health", Args: []string{"device"}},
	{Name: "smartctl.attributes", Args: []string{"device"}},
	{Name: "smartctl.list"},
	{Name: "smartctl.json", Args: []string{"device"}},

	// ── virsh ──────────────────────────────────────────────────────────────
	{Name: "virsh.start", Args: []string{"domain"}, Mutating: true},
	{Name: "virsh.stop", Args: []string{"domain"}, Mutating: true},
	{Name: "virsh.reboot", Args: []string{"domain"}, Mutating: true},
	{Name: "virsh.shutdown", Args: []string{"domain"}, Mutating: true},
	{Name: "virsh.suspend", Args: []string{"domain"}, Mutating: true},
	{Name: "virsh.resume", Args: []string{"domain"}, Mutating: true},
	{Name: "virsh.list"},
	{Name: "virsh.info", Args: []string{"domain"}},
	{Name: "virsh.version"},

	// ── ethtool ────────────────────────────────────────────────────────────
	{Name: "ethtool.show", Args: []string{"interface"}},
	{Name: "ethtool.set_speed", Args: []string{"interface", "speed"}, Mutating: true},
	{Name: "ethtool.set_duplex", Args: []string{"interface", "duplex"}, Mutating: true},
	{Name: "ethtool.set_autoneg", Args: []string{"interface", "autoneg"}, Mutating: true},
	{Name: "ethtool.set_pause", Args: []string{"interface", "rx", "tx"}, Mutating: true},
	{Name: "ethtool.set_offload", Args: []string{"interface", "feature", "value"}, Mutating: true},

	// ── systemd_analyze ────────────────────────────────────────────────────
	{Name: "systemd_analyze.time"},
	{Name: "systemd_analyze.blame"},
	{Name: "systemd_analyze.critical_chain"},
	{Name: "systemd_analyze.security", Args: []string{"unit"}},
	{Name: "systemd_analyze.verify", Args: []string{"unit"}},

	// ── nvme ───────────────────────────────────────────────────────────────
	{Name: "nvme.list"},
	{Name: "nvme.smart_log", Args: []string{"device"}},
	{Name: "nvme.firmware_log", Args: []string{"device"}},
	{Name: "nvme.error_log", Args: []string{"device"}},
	{Name: "nvme.version"},

	// ── lshw ───────────────────────────────────────────────────────────────
	{Name: "lshw.short"},
	{Name: "lshw.class", Args: []string{"class"}},
	{Name: "lshw.json"},
	{Name: "lshw.system"},
	{Name: "lshw.memory"},
	{Name: "lshw.disk"},
	{Name: "lshw.network"},

	// ── ipaddr ─────────────────────────────────────────────────────────────
	{Name: "ipaddr.list"},
	{Name: "ipaddr.list_interface", Args: []string{"interface"}},
	{Name: "ipaddr.add", Args: []string{"address", "interface"}, Mutating: true},
	{Name: "ipaddr.delete", Args: []string{"address", "interface"}, Mutating: true},
	{Name: "ipaddr.flush", Args: []string{"interface"}, Mutating: true},
	{Name: "ipaddr.links"},
	{Name: "ipaddr.link_up", Args: []string{"interface"}, Mutating: true},
	{Name: "ipaddr.link_down", Args: []string{"interface"}, Mutating: true},

	// ── udevadm ───────────────────────────────────────────────────────────
	{Name: "udevadm.control", Args: []string{"action"}, Mutating: true},
	{Name: "udevadm.trigger", Args: []string{"subsystem"}, Mutating: true},
	{Name: "udevadm.settle", Args: []string{"timeout"}},
	{Name: "udevadm.info", Args: []string{"query", "device"}},
	{Name: "udevadm.monitor"},

	// ── modinfo ───────────────────────────────────────────────────────────
	{Name: "modinfo.info", Args: []string{"module"}},
	{Name: "modinfo.list"},
	{Name: "modinfo.version"},

	// ── dconf ─────────────────────────────────────────────────────────────
	{Name: "dconf.read", Args: []string{"key"}},
	{Name: "dconf.write", Args: []string{"key", "value"}, Mutating: true},
	{Name: "dconf.list", Args: []string{"dir"}},
	{Name: "dconf.reset", Args: []string{"key"}, Mutating: true},

	// ── locale_gen ────────────────────────────────────────────────────────
	{Name: "locale_gen.generate", Args: []string{"locale"}, Mutating: true},
	{Name: "locale_gen.list"},
	{Name: "locale_gen.remove", Args: []string{"locale"}, Mutating: true},

	// ── pam_limits ────────────────────────────────────────────────────────
	{Name: "pam_limits.set", Args: []string{"domain", "type", "item", "value"}, Mutating: true},
	{Name: "pam_limits.list"},

	// ── motd ──────────────────────────────────────────────────────────────
	{Name: "motd.read"},
	{Name: "motd.write", Args: []string{"content"}, Mutating: true},

	// ── issue ─────────────────────────────────────────────────────────────
	{Name: "issue.read"},
	{Name: "issue.write", Args: []string{"content"}, Mutating: true},

	// ── authorized_key ──────────────────────────────────────────────────────
	{Name: "authorized_key.manage", Args: []string{"username", "key", "state", "path"}, Mutating: true},
	{Name: "authorized_key.list", Args: []string{"username", "path"}},
	{Name: "authorized_key.check", Args: []string{"username", "key", "path"}},

	// ── blockinfile ─────────────────────────────────────────────────────────
	{Name: "blockinfile.manage", Args: []string{"path", "block", "state", "marker", "insert_after", "insert_before"}, Mutating: true},
	{Name: "blockinfile.read", Args: []string{"path", "marker"}},

	// ── debconf ─────────────────────────────────────────────────────────────
	{Name: "debconf.set", Args: []string{"package", "name", "vtype", "value"}, Mutating: true},
	{Name: "debconf.get", Args: []string{"package", "name"}},
	{Name: "debconf.list", Args: []string{"package"}},

	// ── reboot ──────────────────────────────────────────────────────────────
	{Name: "reboot.request", Args: []string{"msg", "delay"}, Mutating: true},
	{Name: "reboot.dry_run", Args: []string{"msg", "delay"}},
	{Name: "reboot.check"},

	// ── swap ────────────────────────────────────────────────────────────────
	{Name: "swap.info"},
	{Name: "swap.create", Args: []string{"path", "size_mb"}, Mutating: true},
	{Name: "swap.enable", Args: []string{"device"}, Mutating: true},
	{Name: "swap.disable", Args: []string{"device"}, Mutating: true},

	// ── raw ─────────────────────────────────────────────────────────────────
	{Name: "raw.execute", Args: []string{"command", "timeout"}, Mutating: true},
	{Name: "raw.execute_with_env", Args: []string{"command", "timeout", "env"}, Mutating: true},

	// ── expect ──────────────────────────────────────────────────────────────
	{Name: "expect.run", Args: []string{"command", "responses", "timeout"}, Mutating: true},
	{Name: "expect.run_simple", Args: []string{"command", "prompt", "response", "timeout"}, Mutating: true},

	// ── slurp ───────────────────────────────────────────────────────────────
	{Name: "slurp.encode", Args: []string{"path"}},
	{Name: "slurp.decode", Args: []string{"encoded", "dest_path"}, Mutating: true},

	// ── wait_for_connection ─────────────────────────────────────────────────
	{Name: "wait_for_connection.wait", Args: []string{"host", "port", "timeout", "delay"}},
	{Name: "wait_for_connection.check_once", Args: []string{"host", "port"}},

	// ── firewalld_rich_rule ─────────────────────────────────────────────────
	{Name: "firewalld_rich_rule.add", Args: []string{"zone", "rule"}, Mutating: true},
	{Name: "firewalld_rich_rule.remove", Args: []string{"zone", "rule"}, Mutating: true},
	{Name: "firewalld_rich_rule.list", Args: []string{"zone"}},
	{Name: "firewalld_rich_rule.exists", Args: []string{"zone", "rule"}},

	// ── firewalld_ipset ─────────────────────────────────────────────────────
	{Name: "firewalld_ipset.create", Args: []string{"name", "type"}, Mutating: true},
	{Name: "firewalld_ipset.delete", Args: []string{"name"}, Mutating: true},
	{Name: "firewalld_ipset.add_entry", Args: []string{"name", "entry"}, Mutating: true},
	{Name: "firewalld_ipset.remove_entry", Args: []string{"name", "entry"}, Mutating: true},
	{Name: "firewalld_ipset.list"},
	{Name: "firewalld_ipset.info", Args: []string{"name"}},

	// ── pause ───────────────────────────────────────────────────────────────
	{Name: "pause.seconds", Args: []string{"duration"}},
	{Name: "pause.prompt", Args: []string{"message"}},
	{Name: "pause.prompt_with_default", Args: []string{"message", "default"}},

	// ── meta ────────────────────────────────────────────────────────────────
	{Name: "meta.end_host"},
	{Name: "meta.end_play"},
	{Name: "meta.clear_host_errors"},
	{Name: "meta.refresh_inventory"},
	{Name: "meta.flush_handlers"},
	{Name: "meta.reset_connection"},
	{Name: "meta.noop"},
	{Name: "meta.fail", Args: []string{"message"}},
	{Name: "meta.assert", Args: []string{"condition", "message"}},
	{Name: "meta.debug", Args: []string{"message", "vars"}},

	// ── uri_ext ─────────────────────────────────────────────────────────────
	{Name: "uri_ext.patch", Args: []string{"url", "body", "headers", "timeout"}, Mutating: true},
	{Name: "uri_ext.delete", Args: []string{"url", "headers", "timeout"}, Mutating: true},
	{Name: "uri_ext.head", Args: []string{"url", "headers", "timeout"}},
	{Name: "uri_ext.options", Args: []string{"url", "headers", "timeout"}},

	// ── hwclock ─────────────────────────────────────────────────────────────
	{Name: "hwclock.get"},
	{Name: "hwclock.set", Mutating: true},
	{Name: "hwclock.hctosys", Mutating: true},
	{Name: "hwclock.set_time", Args: []string{"time"}, Mutating: true},

	// ── mdadm ───────────────────────────────────────────────────────────────
	{Name: "mdadm.create", Args: []string{"device", "level", "devices"}, Mutating: true},
	{Name: "mdadm.destroy", Args: []string{"device"}, Mutating: true},
	{Name: "mdadm.detail", Args: []string{"device"}},
	{Name: "mdadm.scan"},
	{Name: "mdadm.add", Args: []string{"device", "member"}, Mutating: true},
	{Name: "mdadm.remove", Args: []string{"device", "member"}, Mutating: true},

	// ── open_iscsi ──────────────────────────────────────────────────────────
	{Name: "open_iscsi.discover", Args: []string{"portal", "port"}},
	{Name: "open_iscsi.login", Args: []string{"target", "portal"}, Mutating: true},
	{Name: "open_iscsi.logout", Args: []string{"target", "portal"}, Mutating: true},
	{Name: "open_iscsi.list_sessions"},
	{Name: "open_iscsi.list_nodes"},
	{Name: "open_iscsi.set_startup", Args: []string{"target", "portal", "startup"}, Mutating: true},

	// ── rfkill ──────────────────────────────────────────────────────────────
	{Name: "rfkill.list"},
	{Name: "rfkill.block", Args: []string{"device"}, Mutating: true},
	{Name: "rfkill.unblock", Args: []string{"device"}, Mutating: true},
	{Name: "rfkill.block_all", Args: []string{"type"}, Mutating: true},
	{Name: "rfkill.unblock_all", Args: []string{"type"}, Mutating: true},

	// ── multipath ───────────────────────────────────────────────────────────
	{Name: "multipath.reconfigure", Mutating: true},
	{Name: "multipath.list_paths"},
	{Name: "multipath.list_maps"},
	{Name: "multipath.add_map", Args: []string{"device"}, Mutating: true},
	{Name: "multipath.remove_map", Args: []string{"device"}, Mutating: true},
	{Name: "multipath.flush", Mutating: true},

	// ── dmsetup ─────────────────────────────────────────────────────────────
	{Name: "dmsetup.create", Args: []string{"name", "table"}, Mutating: true},
	{Name: "dmsetup.remove", Args: []string{"name"}, Mutating: true},
	{Name: "dmsetup.remove_all", Mutating: true},
	{Name: "dmsetup.list"},
	{Name: "dmsetup.info", Args: []string{"name"}},
	{Name: "dmsetup.suspend", Args: []string{"name"}, Mutating: true},
	{Name: "dmsetup.resume", Args: []string{"name"}, Mutating: true},

	// ── lvm_enhanced ────────────────────────────────────────────────────────
	{Name: "lvm_enhanced.pv_create", Args: []string{"device"}, Mutating: true},
	{Name: "lvm_enhanced.pv_remove", Args: []string{"device", "force"}, Mutating: true},
	{Name: "lvm_enhanced.pv_list"},
	{Name: "lvm_enhanced.vg_create", Args: []string{"name", "devices"}, Mutating: true},
	{Name: "lvm_enhanced.vg_remove", Args: []string{"name", "force"}, Mutating: true},
	{Name: "lvm_enhanced.vg_extend", Args: []string{"vg_name", "device"}, Mutating: true},
	{Name: "lvm_enhanced.vg_list"},
	{Name: "lvm_enhanced.lv_extend", Args: []string{"lv_path", "size"}, Mutating: true},
	{Name: "lvm_enhanced.lv_extend_all", Args: []string{"lv_path"}, Mutating: true},
	{Name: "lvm_enhanced.lv_list"},

	// ── pacman ────────────────────────────────────────────────────────────
	{Name: "pacman.clean", Mutating: true},
	{Name: "pacman.info", Args: []string{"name"}},
	{Name: "pacman.install", Args: []string{"name"}, Mutating: true},
	{Name: "pacman.install_file", Args: []string{"path"}, Mutating: true},
	{Name: "pacman.list"},
	{Name: "pacman.remove", Args: []string{"name", "cascade"}, Mutating: true},
	{Name: "pacman.remove_orphans", Mutating: true},
	{Name: "pacman.search", Args: []string{"name"}},
	{Name: "pacman.update", Args: []string{"name"}, Mutating: true},
	{Name: "pacman.update_database", Mutating: true},
	{Name: "pacman.upgrade", Mutating: true},

	// ── puppet ──────────────────────────────────────────────────────────────
	{Name: "puppet.run", Args: []string{"environment", "tags"}, Mutating: true},
	{Name: "puppet.run_noop", Args: []string{"environment", "tags"}},
	{Name: "puppet.status"},
	{Name: "puppet.disable", Args: []string{"message"}, Mutating: true},
	{Name: "puppet.enable", Mutating: true},
	{Name: "puppet.fact", Args: []string{"name"}},
	{Name: "puppet.module_list"},
	{Name: "puppet.module_install", Args: []string{"name", "version"}, Mutating: true},

	// ── new functions (not in earlier batches) ──
	{Name: "pip.freeze", Args: []string{"executable"}},
	{Name: "pip.install_requirements", Args: []string{"requirements", "executable"}, Mutating: true},
	{Name: "flatpak.add_remote", Args: []string{"name", "url"}, Mutating: true},
	{Name: "yarn.install", Args: []string{"name", "version", "global"}, Mutating: true},
	{Name: "yarn.remove", Args: []string{"name", "global"}, Mutating: true},
	{Name: "yarn.global", Args: []string{"directory"}, Mutating: true},
	{Name: "yarn.list", Args: []string{"global"}},
	{Name: "htpasswd.set", Args: []string{"path", "username", "password", "create"}, Mutating: true},
	{Name: "htpasswd.remove", Args: []string{"path", "username"}, Mutating: true},
	{Name: "htpasswd.info", Args: []string{"path"}},
	{Name: "htpasswd.hash_sha1", Args: []string{"password"}},
	{Name: "sudoers.set", Args: []string{"name", "user", "commands", "nopasswd", "sudoers_dir"}, Mutating: true},
	{Name: "sudoers.remove", Args: []string{"name", "sudoers_dir"}, Mutating: true},
	{Name: "sudoers.info", Args: []string{"name", "sudoers_dir"}},
	{Name: "monit.start", Args: []string{"service"}, Mutating: true},
	{Name: "monit.stop", Args: []string{"service"}, Mutating: true},
	{Name: "monit.monitor", Args: []string{"service"}, Mutating: true},
	{Name: "monit.unmonitor", Args: []string{"service"}, Mutating: true},
	{Name: "monit.restart", Args: []string{"service"}, Mutating: true},
	{Name: "monit.status"},
	{Name: "monit.reload", Mutating: true},

	// ── kubernetes ─────────────────────────────────────────────────────────
	{Name: "kubernetes.apply", Args: []string{"manifest", "namespace", "dry_run"}, Mutating: true},
	{Name: "kubernetes.delete", Args: []string{"manifest", "namespace"}, Mutating: true},
	{Name: "kubernetes.get", Args: []string{"resource_type", "name", "namespace"}},
	{Name: "kubernetes.list", Args: []string{"resource_type", "namespace", "labels"}},
	{Name: "kubernetes.create_namespace", Args: []string{"name"}, Mutating: true},
	{Name: "kubernetes.delete_namespace", Args: []string{"name"}, Mutating: true},
	{Name: "kubernetes.get_pods", Args: []string{"namespace", "labels"}},
	{Name: "kubernetes.get_services", Args: []string{"namespace"}},
	{Name: "kubernetes.get_deployments", Args: []string{"namespace"}},
	{Name: "kubernetes.scale", Args: []string{"deployment", "replicas", "namespace"}, Mutating: true},
	{Name: "kubernetes.rollout_status", Args: []string{"deployment", "namespace"}},
	{Name: "kubernetes.exec", Args: []string{"pod", "command", "namespace", "container"}, Mutating: true},
	{Name: "kubernetes.logs", Args: []string{"pod", "namespace", "container", "tail"}},
	{Name: "kubernetes.wait_ready", Args: []string{"resource_type", "name", "namespace", "timeout"}},

	// ── portage ─────────────────────────────────────────────────────
	{Name: "portage.install", Args: []string{"name", "version"}, Mutating: true},
	{Name: "portage.remove", Args: []string{"name"}, Mutating: true},
	{Name: "portage.update", Args: []string{"name", "deep"}, Mutating: true},
	{Name: "portage.sync", Mutating: true},
	{Name: "portage.info", Args: []string{"name"}},
	{Name: "portage.list"},
	{Name: "portage.search", Args: []string{"name"}},
	{Name: "portage.depclean", Mutating: true},
	{Name: "portage.metadata", Args: []string{"name"}},

	// ── pkgng ───────────────────────────────────────────────────────
	{Name: "pkgng.install", Args: []string{"name", "version"}, Mutating: true},
	{Name: "pkgng.remove", Args: []string{"name"}, Mutating: true},
	{Name: "pkgng.update", Mutating: true},
	{Name: "pkgng.upgrade", Args: []string{"name"}, Mutating: true},
	{Name: "pkgng.autoclean", Mutating: true},
	{Name: "pkgng.info", Args: []string{"name"}},
	{Name: "pkgng.list"},
	{Name: "pkgng.search", Args: []string{"name"}},
	{Name: "pkgng.stats"},

	// ── podman ──────────────────────────────────────────────────────
	{Name: "podman.run", Args: []string{"image", "name", "command"}, Mutating: true},
	{Name: "podman.stop", Args: []string{"name", "timeout"}, Mutating: true},
	{Name: "podman.start", Args: []string{"name"}, Mutating: true},
	{Name: "podman.remove", Args: []string{"name", "force"}, Mutating: true},
	{Name: "podman.list_containers", Args: []string{"all"}},
	{Name: "podman.inspect", Args: []string{"name"}},
	{Name: "podman.pull", Args: []string{"image"}, Mutating: true},
	{Name: "podman.list_images"},
	{Name: "podman.remove_image", Args: []string{"image_id", "force"}, Mutating: true},
	{Name: "podman.create_pod", Args: []string{"name"}, Mutating: true},
	{Name: "podman.stop_pod", Args: []string{"name"}, Mutating: true},
	{Name: "podman.remove_pod", Args: []string{"name", "force"}, Mutating: true},
	{Name: "podman.list_pods"},

	// ── nftables ────────────────────────────────────────────────────
	{Name: "nftables.add_table", Args: []string{"family", "name"}, Mutating: true},
	{Name: "nftables.delete_table", Args: []string{"family", "name"}, Mutating: true},
	{Name: "nftables.list_tables"},
	{Name: "nftables.add_chain", Args: []string{"family", "table", "name", "type", "hook", "priority"}, Mutating: true},
	{Name: "nftables.delete_chain", Args: []string{"family", "table", "name"}, Mutating: true},
	{Name: "nftables.add_rule", Args: []string{"family", "table", "chain", "expression"}, Mutating: true},
	{Name: "nftables.delete_rule", Args: []string{"family", "table", "chain", "handle"}, Mutating: true},
	{Name: "nftables.flush_chain", Args: []string{"family", "table", "chain"}, Mutating: true},
	{Name: "nftables.flush_table", Args: []string{"family", "table"}, Mutating: true},
	{Name: "nftables.flush_ruleset", Mutating: true},
	{Name: "nftables.list_ruleset"},
	{Name: "nftables.add_set", Args: []string{"family", "table", "name", "type", "flags"}, Mutating: true},
	{Name: "nftables.delete_set", Args: []string{"family", "table", "name"}, Mutating: true},
	{Name: "nftables.add_element", Args: []string{"family", "table", "set", "element"}, Mutating: true},
	{Name: "nftables.delete_element", Args: []string{"family", "table", "set", "element"}, Mutating: true},
	{Name: "nftables.export", Args: []string{"format"}},

	// ── mongodb ──────────────────────────────────────────────────────────────
	{Name: "mongodb.create_database", Args: []string{"host", "port", "name"}, Mutating: true},
	{Name: "mongodb.drop_database", Args: []string{"host", "port", "name"}, Mutating: true},
	{Name: "mongodb.list_databases", Args: []string{"host", "port"}},
	{Name: "mongodb.create_user", Args: []string{"host", "port", "database", "user", "password", "roles"}, Mutating: true},
	{Name: "mongodb.drop_user", Args: []string{"host", "port", "database", "user"}, Mutating: true},
	{Name: "mongodb.list_users", Args: []string{"host", "port", "database"}},
	{Name: "mongodb.create_collection", Args: []string{"host", "port", "database", "collection"}, Mutating: true},
	{Name: "mongodb.drop_collection", Args: []string{"host", "port", "database", "collection"}, Mutating: true},
	{Name: "mongodb.list_collections", Args: []string{"host", "port", "database"}},
	{Name: "mongodb.create_index", Args: []string{"host", "port", "database", "collection", "keys", "unique", "name"}, Mutating: true},
	{Name: "mongodb.drop_index", Args: []string{"host", "port", "database", "collection", "index_name"}, Mutating: true},
	{Name: "mongodb.list_indexes", Args: []string{"host", "port", "database", "collection"}},
	{Name: "mongodb.server_status", Args: []string{"host", "port"}},
	{Name: "mongodb.replica_set_status", Args: []string{"host", "port"}},

	// ── tomcat ──────────────────────────────────────────────────────────────
	{Name: "tomcat.start", Args: []string{"catalina_home"}, Mutating: true},
	{Name: "tomcat.stop", Args: []string{"catalina_home"}, Mutating: true},
	{Name: "tomcat.restart", Args: []string{"catalina_home"}, Mutating: true},
	{Name: "tomcat.status", Args: []string{"catalina_home"}},
	{Name: "tomcat.deploy", Args: []string{"catalina_home", "war_path", "context_path"}, Mutating: true},
	{Name: "tomcat.undeploy", Args: []string{"catalina_home", "context_path"}, Mutating: true},
	{Name: "tomcat.list_apps", Args: []string{"catalina_home"}},
	{Name: "tomcat.reload", Args: []string{"catalina_home", "context_path"}, Mutating: true},
	{Name: "tomcat.version", Args: []string{"catalina_home"}},

	// ── java_cert ────────────────────────────────────────────────────────────
	{Name: "java_cert.import", Args: []string{"keystore_path", "password", "alias", "cert_path", "cert_type"}, Mutating: true},
	{Name: "java_cert.remove", Args: []string{"keystore_path", "password", "alias"}, Mutating: true},
	{Name: "java_cert.list", Args: []string{"keystore_path", "password"}},
	{Name: "java_cert.exists", Args: []string{"keystore_path", "password", "alias"}},
	{Name: "java_cert.export", Args: []string{"keystore_path", "password", "alias", "output_path", "cert_type"}, Mutating: true},
	{Name: "java_cert.info", Args: []string{"keystore_path", "password"}},
	{Name: "java_cert.import_chain", Args: []string{"keystore_path", "password", "p12_path", "p12_password"}, Mutating: true},
	{Name: "java_cert.change_password", Args: []string{"keystore_path", "old_password", "new_password"}, Mutating: true},

	// ── maven_artifact ───────────────────────────────────────────────────────
	{Name: "maven_artifact.download", Args: []string{"repo_url", "group_id", "artifact_id", "version", "dest", "extension"}, Mutating: true},
	{Name: "maven_artifact.resolve", Args: []string{"repo_url", "group_id", "artifact_id", "version", "extension"}},
	{Name: "maven_artifact.deploy", Args: []string{"repo_url", "group_id", "artifact_id", "version", "src_path", "extension"}, Mutating: true},
	{Name: "maven_artifact.get_latest_version", Args: []string{"repo_url", "group_id", "artifact_id"}},
	{Name: "maven_artifact.checksum", Args: []string{"file_path"}},
}

// BuiltinOps are runner instruction ops that are not SDK calls.
var BuiltinOps = []string{"log", "alert", "set", "report", "binary.exec"}

// byName indexes Funcs by canonical name.
var byName = func() map[string]Func {
	m := make(map[string]Func, len(Funcs))
	for _, f := range Funcs {
		m[f.Name] = f
	}
	return m
}()

// Lookup returns the spec for a canonical function name.
func Lookup(name string) (Func, bool) {
	f, ok := byName[name]
	return f, ok
}

// Mutating reports whether an operation (SDK call or builtin VM op, with
// historical aliases resolved) changes system or remote state. known is
// false for names outside the canonical table and BuiltinOps — callers
// must skip privilege enforcement for them (they are custom builtins, not
// OpsLang operations). binary.exec counts as mutating: it runs an
// arbitrary binary, which can change anything.
func Mutating(op string) (isMutating, known bool) {
	if canonical, isAlias := Aliases[op]; isAlias {
		op = canonical
	}
	if f, ok := byName[op]; ok {
		return f.Mutating, true
	}
	switch op {
	case "binary.exec":
		return true, true
	case "log", "alert", "set", "report":
		return false, true // output/bookkeeping ops only
	}
	return false, false
}

// ArgNames returns the positional argument names for an op (SDK call or
// builtin). The second return is false when the op is unknown, in which
// case callers must reject the call rather than guess.
func ArgNames(op string) ([]string, bool) {
	if f, ok := byName[op]; ok {
		return f.Args, true
	}
	for _, b := range BuiltinOps {
		if op == b {
			switch op {
			case "log", "alert":
				return []string{"message"}, true
			case "set":
				return []string{"value"}, true
			case "binary.exec":
				return []string{"path", "args"}, true
			}
		}
	}
	return nil, false
}

// Names returns all canonical function names, optionally filtered by
// availability.
func Names(avail *Availability) []string {
	out := make([]string, 0, len(Funcs))
	for _, f := range Funcs {
		if avail != nil && f.Avail != *avail {
			continue
		}
		out = append(out, f.Name)
	}
	return out
}

// Aliases maps historical runner-registry names to canonical names. Engines
// accept these at lookup time for backward compatibility with existing
// instruction packages, but generators only emit canonical names.
var Aliases = map[string]string{
	"sys.load.avg":         "sys.load",
	"sys.host.info":        "sys.os",
	"net.http.get":         "net.http_get",
	"net.http.post":        "net.http_post",
	"net.tcp.ping":         "net.tcp_check",
	"net.dns.resolve":      "net.dns_lookup",
	"process.find.by_name": "process.find_by_name",
	"process.find.by_port": "process.find_by_port",
	"file.info":            "file.stat",
	"pkg.search":           "pkg.info",
}
