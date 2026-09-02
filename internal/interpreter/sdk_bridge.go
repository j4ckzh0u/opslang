// Package interpreter - SDK bridge: registers ops-core-sdk functions as interpreter builtins.
package interpreter

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/j4ckzh0u/opslang/internal/opsspec"
	sdkacl "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/acl"
	sdkaddhost "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/add_host"
	sdkalternatives "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/alternatives"
	sdkapache2 "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/apache2"
	sdkapache2mod "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/apache2_module"
	sdkapk "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/apk"
	sdkapt "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/apt"
	sdkaptrepo "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/apt_repo"
	sdkaptkey "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/aptkey"
	sdkarchive "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/archive"
	sdkassert "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/assert"
	sdkasyncstatus "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/async_status"
	sdkat "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/at"
	sdkauthorized_key "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/authorized_key"
	sdkblockdev "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/blockdev"
	sdkblockinfile "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/blockinfile"
	sdkbtrfs "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/btrfs"
	sdkcapture "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/capture"
	sdkcargo "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/cargo"
	sdkcausal "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/causal"
	sdkcertbot "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/certbot"
	sdkcloudinit "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/cloud_init"
	sdkcommand "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/command"
	sdkcomposer "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/composer"
	sdkconsul "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/consul"
	sdkcopy "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/copy"
	sdkcron "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/cron"
	sdkcronvar "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/cronvar"
	sdkcrypttab "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/crypttab"
	sdkdconf "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/dconf"
	sdkdebconf "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/debconf"
	sdkdebug "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/debug"
	sdkdisk "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/disk"
	sdkdmidecode "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/dmidecode"
	sdkdmsetup "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/dmsetup"
	sdkdnf "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/dnf"
	sdkdnsmasq "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/dnsmasq"
	sdkdocker "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/docker"
	sdkcompose "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/docker_compose"
	sdkdockercontainer "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/docker_container"
	sdkdockerimage "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/docker_image"
	sdkdockernet "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/docker_network"
	sdkdockervol "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/docker_volume"
	sdkdpkgsel "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/dpkg_selections"
	sdketcd "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/etcd"
	sdkethtool "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/ethtool"
	sdkexpect "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/expect"
	sdkfail "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/fail"
	sdkfail2ban "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/fail2ban"
	sdkfetch "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/fetch"
	sdkfile "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/file"
	sdkfilesystem "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/filesystem"
	sdkfind "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/find"
	sdkfirewalld "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/firewalld"
	sdkfirewalld_ipset "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/firewalld_ipset"
	sdkfirewalld_rich_rule "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/firewalld_rich_rule"
	sdkfirewalldzone "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/firewalld_zone"
	sdkflatpak "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/flatpak"
	sdkgem "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/gem"
	sdkgeturl "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/get_url"
	sdkgetent "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/getent"
	sdkgit "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/git"
	sdkgitconfig "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/git_config"
	sdkgluster "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/gluster"
	sdkgroup "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/group"
	sdkgroupby "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/group_by"
	sdkhaproxy "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/haproxy"
	sdkbrew "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/homebrew"
	sdkhostname "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/hostname"
	sdkhosts "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/hosts"
	sdkhtpasswd "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/htpasswd"
	sdkhwclock "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/hwclock"
	sdkincludevars "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/include_vars"
	sdkinifile "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/ini_file"
	sdkiplink "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/ip_link"
	sdkipneighbor "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/ip_neighbor"
	sdkipnetns "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/ip_netns"
	sdkiproute "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/ip_route"
	sdkipaddr "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/ipaddr"
	sdkiptables "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/iptables"
	sdkissue "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/issue"
	sdkjavacert "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/java_cert"
	sdkjournald "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/journald"
	sdkjson "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/json"
	sdkkernel "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/kernel"
	sdkknownhosts "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/known_hosts"
	sdkk8s "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/kubernetes"
	sdklimits "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/limits"
	sdklineinfile "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/lineinfile"
	sdklocale "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/locale"
	sdklocale_gen "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/locale_gen"
	sdklogrotate "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/logrotate"
	sdklsb "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/lsb_release"
	sdkslshw "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/lshw"
	sdklvg "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/lvg"
	sdklvm_enhanced "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/lvm_enhanced"
	sdklvol "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/lvol"
	sdkmail "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/mail"
	sdkmaven "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/maven_artifact"
	sdkmdadm "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/mdadm"
	sdkmemcached "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/memcached"
	sdkmeta "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/meta"
	sdkmodinfo "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/modinfo"
	sdkmodprobe "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/modprobe"
	sdkmongodb "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/mongodb"
	sdkmonit "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/monit"
	sdkmotd "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/motd"
	sdkmount "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/mount"
	sdkmultipath "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/multipath"
	sdkmysql "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/mysql"
	sdknet "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/net"
	sdknfsexports "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/nfs_exports"
	sdknftables "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/nftables"
	sdknginx "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/nginx"
	sdknmcli "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/nmcli"
	sdknomad "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/nomad"
	sdknormalize "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/normalize"
	sdknpm "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/npm"
	sdkntp "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/ntp"
	sdknvme "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/nvme"
	sdkopen_iscsi "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/open_iscsi"
	sdkopenssl "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/openssl_cert"
	sdkopensslcsr "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/openssl_csr"
	sdkopensslprivatekey "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/openssl_privatekey"
	sdkopensslpublickey "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/openssl_publickey"
	sdkopenvpn "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/openvpn"
	sdkpackagemgr "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/package"
	sdkpackagefacts "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/package_facts"
	sdkpacman "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/pacman"
	sdkpam_limits "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/pam_limits"
	sdkpamd "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/pamd"
	sdkparted "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/parted"
	sdkpatch "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/patch"
	sdkpause "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/pause"
	sdkping "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/ping"
	sdkpip "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/pip"
	sdkpipx "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/pipx"
	opspkg "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/pkg"
	sdkpkgng "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/pkgng"
	sdkpodman "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/podman"
	sdkportage "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/portage"
	sdkpostfix "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/postfix"
	sdkpostgresql "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/postgresql"
	sdkprocess "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/process"
	sdkpuppet "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/puppet"
	sdkrabbitmq "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/rabbitmq"
	sdkraw "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/raw"
	sdkreboot "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/reboot"
	sdkredis "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/redis"
	sdkreplace "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/replace"
	sdkresolv "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/resolv"
	sdkrfkill "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/rfkill"
	sdkrpmkey "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/rpmkey"
	sdkrunit "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/runit"
	sdkscript "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/script"
	sdksebool "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/seboolean"
	sefcontext "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/sefcontext"
	sdkselinux "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/selinux"
	seport "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/seport"
	sdkservice "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/service"
	sdkservicefacts "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/service_facts"
	sdksetfact "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/set_fact"
	sdksetstats "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/set_stats"
	sdkslurp "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/slurp"
	sdksmartctl "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/smartctl"
	sdksmartnotify "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/smartctl_notify"
	sdksnap "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/snap"
	sdkssh "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/ssh"
	sdksshconfig "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/ssh_config"
	sdksshdconfig "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/sshd_config"
	sdkstat "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/stat"
	sdksudoers "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/sudoers"
	sdksupervisor "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/supervisor"
	sdksvn "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/svn"
	sdkswap "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/swap"
	sdksync "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/synchronize"
	sdksys "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/sys"
	sdksyspersist "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/sys_persist"
	sdksysctl "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/sysctl"
	sdksysfs "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/sysfs"
	sdksystemd "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/systemd"
	sdksystemd_analyze "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/systemd_analyze"
	sdksysvinit "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/sysvinit"
	sdktempfile "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/tempfile"
	sdktime "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/time"
	sdktimezone "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/timezone"
	sdktomcat "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/tomcat"
	sdktuned "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/tuned"
	sdktypedebug "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/type_debug"
	sdkudevadm "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/udevadm"
	sdkufw "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/ufw"
	sdkunarchive "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/unarchive"
	sdkuri "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/uri"
	sdkuri_ext "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/uri_ext"
	sdkuser "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/user"
	sdkvalidatecerts "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/validate_certs"
	sdkvault "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/vault"
	sdkvirsh "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/virsh"
	sdkwaitfor "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/wait_for"
	sdkwait_for_connection "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/wait_for_connection"
	sdkwebhook "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/webhook"
	sdkwireguard "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/wireguard"
	sdkxattr "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/xattr"
	sdkxml "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/xml"
	sdkyaml "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/yaml"
	sdkyarn "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/yarn"
	sdkyumrepo "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/yum_repo"
	sdkzfs "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/zfs"
	sdkzookeeper "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/zookeeper"
	sdkzypper "github.com/j4ckzh0u/opslang/pkg/ops-core-sdk/zypper"
)

// SDKBuiltinNames returns every SDK function name registered by
// RegisterSDKBuiltins. Used by cross-engine consistency tests.
func SDKBuiltinNames() []string {
	interp := New(nil)
	RegisterSDKBuiltins(interp)
	names := make([]string, 0, len(interp.builtins))
	for name := range interp.builtins {
		names = append(names, name)
	}
	return names
}

// structToMap converts any struct to map[string]interface{} via JSON roundtrip.
// This is needed because the interpreter's member access only supports map[string]interface{}.
func structToMap(v interface{}) (interface{}, error) {
	if v == nil {
		return nil, nil
	}
	data, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("structToMap marshal: %w", err)
	}
	var result interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("structToMap unmarshal: %w", err)
	}
	return result, nil
}

// RegisterSDKBuiltins registers all ops-core-sdk functions into the interpreter.
func RegisterSDKBuiltins(interp *Interpreter) {
	// ── sys.* ──────────────────────────────────────────────────────────
	interp.builtins["sys.hostname"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdksys.Hostname()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	interp.builtins["sys.cpu.usage"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdksys.GetCPUUsage()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	interp.builtins["sys.cpu.info"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdksys.GetCPUInfo()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	interp.builtins["sys.cpu.count"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdksys.GetCPUCount()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	interp.builtins["sys.memory.info"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdksys.GetMemoryInfo()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	interp.builtins["sys.disk.usage"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("sys.disk.usage() requires 1 argument (path)")
		}
		path, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("sys.disk.usage(): argument must be string")
		}
		r, err := sdksys.GetDiskUsage(path)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	interp.builtins["sys.disk.partitions"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdksys.GetDiskPartitions()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	interp.builtins["sys.load"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdksys.GetLoadAvg()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	interp.builtins["sys.os"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdksys.GetHostInfo()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	interp.builtins["sys.uptime"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdksys.Uptime()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	interp.builtins["sys.users"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdksys.Users()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	interp.builtins["sys.net.interfaces"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdksys.GetNetInterfaces()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	interp.builtins["sys.net.all_interfaces"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdksys.GetAllNetInterfaces()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	interp.builtins["sys.net.primary_ip"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdksys.GetPrimaryIP()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["sys.net.rate"] = func(args ...interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("sys.net.rate() requires 1 argument (seconds)")
		}
		secondsValue, err := toInt(args[0])
		if err != nil {
			return nil, fmt.Errorf("sys.net.rate(): seconds must be a number")
		}
		seconds, ok := secondsValue.(int64)
		if !ok {
			return nil, fmt.Errorf("sys.net.rate(): seconds must be a number")
		}
		r, err := sdksys.GetNetRate(int(seconds))
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	interp.builtins["sys.virt"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdksys.GetVirtInfo()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── file.* ─────────────────────────────────────────────────────────
	interp.builtins["file.read"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("file.read() requires 1 argument (path)")
		}
		path, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("file.read(): argument must be string")
		}
		r, err := sdkfile.Read(path)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	interp.builtins["file.write"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("file.write() requires 2 arguments (path, content)")
		}
		path, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("file.write(): path must be string")
		}
		content, ok := args[1].(string)
		if !ok {
			return nil, fmt.Errorf("file.write(): content must be string")
		}
		r, err := sdkfile.Write(path, content)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	interp.builtins["file.exists"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("file.exists() requires 1 argument (path)")
		}
		path, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("file.exists(): argument must be string")
		}
		r, err := sdkfile.Exists(path)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	interp.builtins["file.ensure"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("file.ensure() requires at least 2 arguments (path, state)")
		}
		path, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("file.ensure(): path must be string")
		}
		state, ok := args[1].(string)
		if !ok {
			return nil, fmt.Errorf("file.ensure(): state must be string (directory|file|touch|absent)")
		}
		mode := ""
		if len(args) > 2 {
			mode, ok = args[2].(string)
			if !ok {
				return nil, fmt.Errorf("file.ensure(): mode must be octal string like \"0755\"")
			}
		}
		r, err := sdkfile.Ensure(path, state, mode)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	interp.builtins["file.copy"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("file.copy() requires 2 arguments (src, dst)")
		}
		src, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("file.copy(): src must be string")
		}
		dst, ok := args[1].(string)
		if !ok {
			return nil, fmt.Errorf("file.copy(): dst must be string")
		}
		r, err := sdkfile.Copy(src, dst)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	interp.builtins["file.move"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("file.move() requires 2 arguments (src, dst)")
		}
		src, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("file.move(): src must be string")
		}
		dst, ok := args[1].(string)
		if !ok {
			return nil, fmt.Errorf("file.move(): dst must be string")
		}
		r, err := sdkfile.Move(src, dst)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	interp.builtins["file.delete"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("file.delete() requires 1 argument (path)")
		}
		path, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("file.delete(): argument must be string")
		}
		r, err := sdkfile.Delete(path)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	interp.builtins["file.stat"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("file.stat() requires 1 argument (path)")
		}
		path, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("file.stat(): argument must be string")
		}
		r, err := sdkfile.Stat(path)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	interp.builtins["file.list"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("file.list() requires 1 argument (dir)")
		}
		dir, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("file.list(): argument must be string")
		}
		r, err := sdkfile.List(dir)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	interp.builtins["file.mkdir"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("file.mkdir() requires 1 argument (path)")
		}
		path, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("file.mkdir(): argument must be string")
		}
		r, err := sdkfile.Mkdir(path)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	interp.builtins["file.checksum"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("file.checksum() requires 2 arguments (path, algo)")
		}
		path, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("file.checksum(): path must be string")
		}
		algo, ok := args[1].(string)
		if !ok {
			return nil, fmt.Errorf("file.checksum(): algo must be string")
		}
		r, err := sdkfile.Checksum(path, algo)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	interp.builtins["file.distribute"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("file.distribute() requires at least 2 arguments (source, targets)")
		}
		source, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("file.distribute(): source must be string")
		}
		targetsRaw, ok := args[1].([]interface{})
		if !ok {
			return nil, fmt.Errorf("file.distribute(): targets must be a list")
		}
		var targets []sdkfile.DistributeTarget
		for i, item := range targetsRaw {
			m, ok := item.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("file.distribute(): target %d must be a dict", i)
			}
			t := sdkfile.DistributeTarget{}
			if h, ok := m["host"].(string); ok {
				t.Host = h
			}
			if p, ok := m["port"].(float64); ok {
				t.Port = int(p)
			}
			if u, ok := m["user"].(string); ok {
				t.User = u
			}
			if d, ok := m["dest"].(string); ok {
				t.Dest = d
			}
			targets = append(targets, t)
		}

		opts := sdkfile.DistributeOptions{}
		if len(args) >= 3 {
			if optsMap, ok := args[2].(map[string]interface{}); ok {
				if v, ok := optsMap["checksum"].(bool); ok {
					opts.Checksum = v
				}
				if v, ok := optsMap["mode"].(string); ok {
					opts.Mode = v
				}
				if v, ok := optsMap["parallel"].(float64); ok {
					opts.Parallel = int(v)
				}
				if v, ok := optsMap["retries"].(float64); ok {
					opts.Retries = int(v)
				}
			}
		}

		r, err := sdkfile.Distribute(source, targets, opts)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	interp.builtins["file.collect"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("file.collect() requires at least 2 arguments (source, targets)")
		}
		source, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("file.collect(): source must be string")
		}
		targetsRaw, ok := args[1].([]interface{})
		if !ok {
			return nil, fmt.Errorf("file.collect(): targets must be a list")
		}
		var targets []sdkfile.CollectTarget
		for i, item := range targetsRaw {
			m, ok := item.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("file.collect(): target %d must be a dict", i)
			}
			t := sdkfile.CollectTarget{}
			if h, ok := m["host"].(string); ok {
				t.Host = h
			}
			if p, ok := m["port"].(float64); ok {
				t.Port = int(p)
			}
			if u, ok := m["user"].(string); ok {
				t.User = u
			}
			if s, ok := m["source"].(string); ok {
				t.Source = s
			}
			targets = append(targets, t)
		}

		opts := sdkfile.CollectOptions{}
		if len(args) >= 3 {
			if optsMap, ok := args[2].(map[string]interface{}); ok {
				if v, ok := optsMap["dest_dir"].(string); ok {
					opts.DestDir = v
				}
				if v, ok := optsMap["parallel"].(float64); ok {
					opts.Parallel = int(v)
				}
				if v, ok := optsMap["retries"].(float64); ok {
					opts.Retries = int(v)
				}
			}
		}

		r, err := sdkfile.Collect(source, targets, opts)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── net.* ──────────────────────────────────────────────────────────
	interp.builtins["net.http_get"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("net.http_get() requires 1 argument (url)")
		}
		url, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("net.http_get(): argument must be string")
		}
		r, err := sdknet.HTTPGet(url)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	interp.builtins["net.http_post"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("net.http_post() requires 2 arguments (url, body)")
		}
		url, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("net.http_post(): url must be string")
		}
		body, ok := args[1].(string)
		if !ok {
			return nil, fmt.Errorf("net.http_post(): body must be string")
		}
		r, err := sdknet.HTTPPost(url, body)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	interp.builtins["net.capture"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 || len(args) > 4 {
			return nil, fmt.Errorf("net.capture() requires 1..4 arguments (iface, seconds, max_packets, pcap_path)")
		}
		iface, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("net.capture(): iface must be string")
		}
		seconds := 5
		maxPkts := 200
		pcapPath := ""
		if len(args) >= 2 {
			if v, ok := args[1].(int64); ok {
				seconds = int(v)
			} else {
				return nil, fmt.Errorf("net.capture(): seconds must be int")
			}
		}
		if len(args) >= 3 {
			if v, ok := args[2].(int64); ok {
				maxPkts = int(v)
			} else {
				return nil, fmt.Errorf("net.capture(): max_packets must be int")
			}
		}
		if len(args) >= 4 {
			v, ok := args[3].(string)
			if !ok {
				return nil, fmt.Errorf("net.capture(): pcap_path must be string")
			}
			pcapPath = v
		}
		// The interpreter executes in-process on the machine running opsctl:
		// a "local:" pcap path is simply an ordinary local file here. Remote
		// contexts (runner) get the embed-and-transfer semantics instead.
		pcapPath = strings.TrimPrefix(pcapPath, sdkcapture.PcapLocalPrefix)
		r, err := sdkcapture.Capture(sdkcapture.Options{
			Iface:    iface,
			Seconds:  seconds,
			MaxPkts:  maxPkts,
			PcapPath: pcapPath,
		})
		if err != nil {
			return nil, err
		}
		return structToMap(*r)
	}

	interp.builtins["net.tcp_check"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("net.tcp_check() requires 2 arguments (host, port)")
		}
		host, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("net.tcp_check(): host must be string")
		}
		portF, err := toFloat(args[1])
		if err != nil {
			return nil, fmt.Errorf("net.tcp_check(): port must be number")
		}
		r, err := sdknet.TCPConnect(host, int(portF))
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	interp.builtins["net.dns_lookup"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("net.dns_lookup() requires 1 argument (domain)")
		}
		domain, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("net.dns_lookup(): argument must be string")
		}
		r, err := sdknet.DNSLookup(domain)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	interp.builtins["net.connections"] = func(args ...interface{}) (interface{}, error) {
		kind := ""
		if len(args) > 0 {
			k, ok := args[0].(string)
			if !ok {
				return nil, fmt.Errorf("net.connections(): kind must be string")
			}
			kind = k
		}
		r, err := sdknet.Connections(kind)
		if err != nil {
			return nil, err
		}
		converted, err := structToMap(r)
		if err != nil {
			return nil, err
		}
		return converted, nil
	}

	interp.builtins["net.interfaces"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdknet.Interfaces()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── process.* ──────────────────────────────────────────────────────
	interp.builtins["process.list"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkprocess.List()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["process.java_apps"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkprocess.JavaApps()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["causal.find"] = func(args ...interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("causal.find() requires 1 argument (name)")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("causal.find(): name must be a string")
		}
		traces, err := sdkcausal.Find(name)
		if err != nil {
			return nil, err
		}
		return structToMap(traces)
	}
	interp.builtins["causal.trace_pid"] = func(args ...interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("causal.trace_pid() requires 1 argument (pid)")
		}
		pidValue, err := toInt(args[0])
		if err != nil {
			return nil, fmt.Errorf("causal.trace_pid(): pid must be a number")
		}
		pid, ok := pidValue.(int64)
		if !ok {
			return nil, fmt.Errorf("causal.trace_pid(): pid must be a number")
		}
		trace, err := sdkcausal.TracePID(int(pid))
		if err != nil {
			return nil, err
		}
		return structToMap(trace)
	}
	interp.builtins["causal.trace_port"] = func(args ...interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("causal.trace_port() requires 1 argument (port)")
		}
		portValue, err := toInt(args[0])
		if err != nil {
			return nil, fmt.Errorf("causal.trace_port(): port must be a number")
		}
		port, ok := portValue.(int64)
		if !ok {
			return nil, fmt.Errorf("causal.trace_port(): port must be a number")
		}
		traces, err := sdkcausal.TracePort(int(port))
		if err != nil {
			return nil, err
		}
		return structToMap(traces)
	}
	interp.builtins["causal.trace_file"] = func(args ...interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("causal.trace_file() requires 1 argument (path)")
		}
		path, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("causal.trace_file(): path must be a string")
		}
		traces, err := sdkcausal.TraceFile(path)
		if err != nil {
			return nil, err
		}
		return structToMap(traces)
	}
	interp.builtins["causal.trace_container"] = func(args ...interface{}) (interface{}, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("causal.trace_container() requires 1 argument (id)")
		}
		id, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("causal.trace_container(): id must be a string")
		}
		traces, err := sdkcausal.TraceContainer(id)
		if err != nil {
			return nil, err
		}
		return structToMap(traces)
	}

	interp.builtins["process.find_by_name"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("process.find_by_name() requires 1 argument (name)")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("process.find_by_name(): argument must be string")
		}
		r, err := sdkprocess.FindByName(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	interp.builtins["process.find_by_port"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("process.find_by_port() requires 1 argument (port)")
		}
		portF, err := toFloat(args[0])
		if err != nil {
			return nil, fmt.Errorf("process.find_by_port(): argument must be number")
		}
		r, err := sdkprocess.FindByPort(int(portF))
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	interp.builtins["process.exec"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("process.exec() requires at least 1 argument (command)")
		}
		cmd, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("process.exec(): command must be string")
		}
		var cmdArgs []string
		for _, a := range args[1:] {
			cmdArgs = append(cmdArgs, fmt.Sprintf("%v", a))
		}
		r, err := sdkprocess.Exec(cmd, cmdArgs)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── service.* ──────────────────────────────────────────────────────
	interp.builtins["service.ensure"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("service.ensure() requires 2 arguments (name, state)")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("service.ensure(): name must be string")
		}
		state, ok := args[1].(string)
		if !ok {
			return nil, fmt.Errorf("service.ensure(): state must be string (started|stopped|restarted|reloaded)")
		}
		r, err := sdkservice.Ensure(name, state)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	interp.builtins["service.ensure_enabled"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("service.ensure_enabled() requires 2 arguments (name, enabled)")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("service.ensure_enabled(): name must be string")
		}
		enabled, ok := args[1].(bool)
		if !ok {
			return nil, fmt.Errorf("service.ensure_enabled(): enabled must be bool")
		}
		r, err := sdkservice.EnsureEnabled(name, enabled)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	interp.builtins["service.status"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("service.status() requires 1 argument (name)")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("service.status(): argument must be string")
		}
		r, err := sdkservice.Status(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	interp.builtins["service.start"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("service.start() requires 1 argument (name)")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("service.start(): argument must be string")
		}
		r, err := sdkservice.Start(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	interp.builtins["service.stop"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("service.stop() requires 1 argument (name)")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("service.stop(): argument must be string")
		}
		r, err := sdkservice.Stop(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	interp.builtins["service.restart"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("service.restart() requires 1 argument (name)")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("service.restart(): argument must be string")
		}
		r, err := sdkservice.Restart(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	interp.builtins["service.enable"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("service.enable() requires 1 argument (name)")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("service.enable(): argument must be string")
		}
		r, err := sdkservice.Enable(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── snap.* ──────────────────────────────────────────────────────────────
	interp.builtins["snap.install"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("snap.install() requires at least 1 argument (name)")
		}
		name, _ := args[0].(string)
		channel := "stable"
		if len(args) > 1 {
			channel, _ = args[1].(string)
		}
		classic := false
		if len(args) > 2 {
			classic, _ = args[2].(bool)
		}
		r, err := sdksnap.Install(name, channel, classic)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["snap.remove"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("snap.remove() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		r, err := sdksnap.Remove(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["snap.refresh"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("snap.refresh() requires at least 1 argument (name)")
		}
		name, _ := args[0].(string)
		channel := ""
		if len(args) > 1 {
			channel, _ = args[1].(string)
		}
		r, err := sdksnap.Refresh(name, channel)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["snap.list"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdksnap.List()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["snap.get"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("snap.get() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		r, err := sdksnap.Get(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["snap.enable"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("snap.enable() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		r, err := sdksnap.Enable(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["snap.disable"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("snap.disable() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		r, err := sdksnap.Disable(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["snap.switch"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("snap.switch() requires 2 arguments (name, channel)")
		}
		name, _ := args[0].(string)
		channel, _ := args[1].(string)
		r, err := sdksnap.Switch(name, channel)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["snap.changes"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdksnap.Changes()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── flatpak.* ────────────────────────────────────────────────────────
	interp.builtins["flatpak.install"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("flatpak.install() requires 1-3 arguments (name, from, user)")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("flatpak.install(): first argument must be string")
		}
		from := ""
		if len(args) > 1 {
			from, _ = args[1].(string)
		}
		user := false
		if len(args) > 2 {
			user, _ = args[2].(bool)
		}
		r, err := sdkflatpak.Install(name, from, user)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["flatpak.remove"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("flatpak.remove() requires 1-2 arguments (name, user)")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("flatpak.remove(): first argument must be string")
		}
		user := false
		if len(args) > 1 {
			user, _ = args[1].(bool)
		}
		r, err := sdkflatpak.Remove(name, user)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["flatpak.update"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("flatpak.update() requires 1-2 arguments (name, user)")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("flatpak.update(): first argument must be string")
		}
		user := false
		if len(args) > 1 {
			user, _ = args[1].(bool)
		}
		r, err := sdkflatpak.Update(name, user)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["flatpak.list"] = func(args ...interface{}) (interface{}, error) {
		user := false
		if len(args) > 0 {
			user, _ = args[0].(bool)
		}
		r, err := sdkflatpak.List(user)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["flatpak.info"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("flatpak.info() requires 1-2 arguments (name, user)")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("flatpak.info(): first argument must be string")
		}
		user := false
		if len(args) > 1 {
			user, _ = args[1].(bool)
		}
		r, err := sdkflatpak.Info(name, user)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["flatpak.run"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("flatpak.run() requires 1-3 arguments (name, args, user)")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("flatpak.run(): first argument must be string")
		}
		var runArgs []string
		if len(args) > 1 && args[1] != nil {
			if argList, ok := args[1].([]interface{}); ok {
				for _, a := range argList {
					if s, ok := a.(string); ok {
						runArgs = append(runArgs, s)
					}
				}
			}
		}
		user := false
		if len(args) > 2 {
			user, _ = args[2].(bool)
		}
		r, err := sdkflatpak.Run(name, runArgs, user)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["flatpak.repair"] = func(args ...interface{}) (interface{}, error) {
		user := false
		if len(args) > 0 {
			user, _ = args[0].(bool)
		}
		r, err := sdkflatpak.Repair(user)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── zfs.* ────────────────────────────────────────────────────────
	interp.builtins["zfs.create"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("zfs.create() requires 1-2 arguments (name, properties)")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("zfs.create(): first argument must be string")
		}
		var props map[string]string
		if len(args) > 1 && args[1] != nil {
			if m, ok := args[1].(map[string]interface{}); ok {
				props = make(map[string]string)
				for k, v := range m {
					if s, ok := v.(string); ok {
						props[k] = s
					}
				}
			}
		}
		r, err := sdkzfs.Create(name, props)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["zfs.destroy"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("zfs.destroy() requires 1-2 arguments (name, recursive)")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("zfs.destroy(): first argument must be string")
		}
		recursive := false
		if len(args) > 1 {
			recursive, _ = args[1].(bool)
		}
		r, err := sdkzfs.Destroy(name, recursive)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["zfs.set"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 3 {
			return nil, fmt.Errorf("zfs.set() requires 3 arguments (name, property, value)")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("zfs.set(): first argument must be string")
		}
		property, ok := args[1].(string)
		if !ok {
			return nil, fmt.Errorf("zfs.set(): second argument must be string")
		}
		value, ok := args[2].(string)
		if !ok {
			return nil, fmt.Errorf("zfs.set(): third argument must be string")
		}
		r, err := sdkzfs.Set(name, property, value)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["zfs.get"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("zfs.get() requires 1-2 arguments (name, property)")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("zfs.get(): first argument must be string")
		}
		property := ""
		if len(args) > 1 {
			property, _ = args[1].(string)
		}
		r, err := sdkzfs.Get(name, property)
		if err != nil {
			return nil, err
		}
		return r, nil
	}
	interp.builtins["zfs.list"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkzfs.List()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["zfs.exists"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("zfs.exists() requires 1 argument (name)")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("zfs.exists(): first argument must be string")
		}
		r, err := sdkzfs.Exists(name)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"exists": r}, nil
	}
	interp.builtins["zfs.list_pools"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkzfs.ListPools()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["zfs.get_pool_status"] = func(args ...interface{}) (interface{}, error) {
		name := ""
		if len(args) > 0 {
			name, _ = args[0].(string)
		}
		r, err := sdkzfs.GetPoolStatus(name)
		if err != nil {
			return nil, err
		}
		return r, nil
	}
	interp.builtins["zfs.snapshot"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("zfs.snapshot() requires 2 arguments (name, snapshot_name)")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("zfs.snapshot(): first argument must be string")
		}
		snapName, ok := args[1].(string)
		if !ok {
			return nil, fmt.Errorf("zfs.snapshot(): second argument must be string")
		}
		r, err := sdkzfs.Snapshot(name, snapName)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["zfs.destroy_snapshot"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("zfs.destroy_snapshot() requires 2 arguments (name, snapshot_name)")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("zfs.destroy_snapshot(): first argument must be string")
		}
		snapName, ok := args[1].(string)
		if !ok {
			return nil, fmt.Errorf("zfs.destroy_snapshot(): second argument must be string")
		}
		r, err := sdkzfs.DestroySnapshot(name, snapName)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── nmcli.* ────────────────────────────────────────────────────────
	interp.builtins["nmcli.add"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("nmcli.add() requires 2-3 arguments (name, type, settings)")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("nmcli.add(): first argument must be string")
		}
		connType, ok := args[1].(string)
		if !ok {
			return nil, fmt.Errorf("nmcli.add(): second argument must be string")
		}
		var settings map[string]string
		if len(args) > 2 && args[2] != nil {
			if m, ok := args[2].(map[string]interface{}); ok {
				settings = make(map[string]string)
				for k, v := range m {
					if s, ok := v.(string); ok {
						settings[k] = s
					}
				}
			}
		}
		r, err := sdknmcli.Add(name, connType, settings)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["nmcli.modify"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("nmcli.modify() requires 2 arguments (name, settings)")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("nmcli.modify(): first argument must be string")
		}
		var settings map[string]string
		if args[1] != nil {
			if m, ok := args[1].(map[string]interface{}); ok {
				settings = make(map[string]string)
				for k, v := range m {
					if s, ok := v.(string); ok {
						settings[k] = s
					}
				}
			}
		}
		r, err := sdknmcli.Modify(name, settings)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["nmcli.delete"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("nmcli.delete() requires 1 argument (name)")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("nmcli.delete(): first argument must be string")
		}
		r, err := sdknmcli.Delete(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["nmcli.up"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("nmcli.up() requires 1 argument (name)")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("nmcli.up(): first argument must be string")
		}
		r, err := sdknmcli.Up(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["nmcli.down"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("nmcli.down() requires 1 argument (name)")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("nmcli.down(): first argument must be string")
		}
		r, err := sdknmcli.Down(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["nmcli.list"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdknmcli.List()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["nmcli.show"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("nmcli.show() requires 1 argument (name)")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("nmcli.show(): first argument must be string")
		}
		r, err := sdknmcli.Show(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["nmcli.list_devices"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdknmcli.ListDevices()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["nmcli.reload"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdknmcli.Reload()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["nmcli.get_general_status"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdknmcli.GetGeneralStatus()
		if err != nil {
			return nil, err
		}
		return r, nil
	}

	// ── crypttab.* ────────────────────────────────────────────────────────
	interp.builtins["crypttab.add"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("crypttab.add() requires 2-4 arguments (name, device, key_file, options)")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("crypttab.add(): first argument must be string")
		}
		device, ok := args[1].(string)
		if !ok {
			return nil, fmt.Errorf("crypttab.add(): second argument must be string")
		}
		keyFile := ""
		if len(args) > 2 {
			keyFile, _ = args[2].(string)
		}
		options := ""
		if len(args) > 3 {
			options, _ = args[3].(string)
		}
		r, err := sdkcrypttab.Add(name, device, keyFile, options)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["crypttab.remove"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("crypttab.remove() requires 1 argument (name)")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("crypttab.remove(): first argument must be string")
		}
		r, err := sdkcrypttab.Remove(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["crypttab.modify"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("crypttab.modify() requires 2-4 arguments (name, device, key_file, options)")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("crypttab.modify(): first argument must be string")
		}
		device := ""
		if len(args) > 1 {
			device, _ = args[1].(string)
		}
		keyFile := ""
		if len(args) > 2 {
			keyFile, _ = args[2].(string)
		}
		options := ""
		if len(args) > 3 {
			options, _ = args[3].(string)
		}
		r, err := sdkcrypttab.Modify(name, device, keyFile, options)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["crypttab.get"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("crypttab.get() requires 1 argument (name)")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("crypttab.get(): first argument must be string")
		}
		r, err := sdkcrypttab.Get(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["crypttab.list"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkcrypttab.List()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["crypttab.exists"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("crypttab.exists() requires 1 argument (name)")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("crypttab.exists(): first argument must be string")
		}
		r, err := sdkcrypttab.Exists(name)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"exists": r}, nil
	}
	interp.builtins["crypttab.validate"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkcrypttab.Validate()
		if err != nil {
			return nil, err
		}
		return r, nil
	}
	interp.builtins["crypttab.backup"] = func(args ...interface{}) (interface{}, error) {
		backupDir := ""
		if len(args) > 0 {
			backupDir, _ = args[0].(string)
		}
		r, err := sdkcrypttab.Backup(backupDir)
		if err != nil {
			return nil, err
		}
		return r, nil
	}

	// ── sysfs.* ────────────────────────────────────────────────────────
	interp.builtins["sysfs.read"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("sysfs.read() requires 1 argument (path)")
		}
		path, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("sysfs.read(): first argument must be string")
		}
		r, err := sdksysfs.Read(path)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"value": r}, nil
	}
	interp.builtins["sysfs.write"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("sysfs.write() requires 2 arguments (path, value)")
		}
		path, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("sysfs.write(): first argument must be string")
		}
		value, ok := args[1].(string)
		if !ok {
			return nil, fmt.Errorf("sysfs.write(): second argument must be string")
		}
		r, err := sdksysfs.Write(path, value)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["sysfs.exists"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("sysfs.exists() requires 1 argument (path)")
		}
		path, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("sysfs.exists(): first argument must be string")
		}
		r, err := sdksysfs.Exists(path)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"exists": r}, nil
	}
	interp.builtins["sysfs.get"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("sysfs.get() requires 1 argument (path)")
		}
		path, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("sysfs.get(): first argument must be string")
		}
		r, err := sdksysfs.Get(path)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["sysfs.list"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("sysfs.list() requires 1 argument (dir_path)")
		}
		dirPath, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("sysfs.list(): first argument must be string")
		}
		r, err := sdksysfs.List(dirPath)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["sysfs.set_device_power"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("sysfs.set_device_power() requires 2 arguments (device_path, state)")
		}
		devicePath, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("sysfs.set_device_power(): first argument must be string")
		}
		state, ok := args[1].(string)
		if !ok {
			return nil, fmt.Errorf("sysfs.set_device_power(): second argument must be string")
		}
		r, err := sdksysfs.SetDevicePower(devicePath, state)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["sysfs.get_device_power"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("sysfs.get_device_power() requires 1 argument (device_path)")
		}
		devicePath, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("sysfs.get_device_power(): first argument must be string")
		}
		r, err := sdksysfs.GetDevicePower(devicePath)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"state": r}, nil
	}
	interp.builtins["sysfs.set_kernel_parameter"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("sysfs.set_kernel_parameter() requires 2 arguments (param, value)")
		}
		param, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("sysfs.set_kernel_parameter(): first argument must be string")
		}
		value, ok := args[1].(string)
		if !ok {
			return nil, fmt.Errorf("sysfs.set_kernel_parameter(): second argument must be string")
		}
		r, err := sdksysfs.SetKernelParameter(param, value)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["sysfs.get_kernel_parameter"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("sysfs.get_kernel_parameter() requires 1 argument (param)")
		}
		param, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("sysfs.get_kernel_parameter(): first argument must be string")
		}
		r, err := sdksysfs.GetKernelParameter(param)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"value": r}, nil
	}

	// ── pamd.* ──────────────────────────────────────────────────────────
	interp.builtins["pamd.get"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("pamd.get() requires 1 argument (service)")
		}
		service, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("pamd.get(): first argument must be string")
		}
		return sdkpamd.Get(service)
	}
	interp.builtins["pamd.list"] = func(args ...interface{}) (interface{}, error) {
		return sdkpamd.List()
	}
	interp.builtins["pamd.add_rule"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 5 {
			return nil, fmt.Errorf("pamd.add_rule() requires 5 arguments")
		}
		svc := args[0].(string)
		rt := args[1].(string)
		ctrl := args[2].(string)
		mod := args[3].(string)
		a := args[4].(string)
		return sdkpamd.AddRule(svc, rt, ctrl, mod, a)
	}
	interp.builtins["pamd.remove_rule"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 3 {
			return nil, fmt.Errorf("pamd.remove_rule() requires 3 arguments")
		}
		return sdkpamd.RemoveRule(args[0].(string), args[1].(string), args[2].(string))
	}
	interp.builtins["pamd.modify_rule"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 5 {
			return nil, fmt.Errorf("pamd.modify_rule() requires 5 arguments")
		}
		return sdkpamd.ModifyRule(args[0].(string), args[1].(string), args[2].(string), args[3].(string), args[4].(string))
	}
	interp.builtins["pamd.validate"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("pamd.validate() requires 1 argument")
		}
		return sdkpamd.Validate(args[0].(string))
	}
	interp.builtins["pamd.backup"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("pamd.backup() requires 2 arguments")
		}
		return sdkpamd.Backup(args[0].(string), args[1].(string))
	}

	// ── getent.* ────────────────────────────────────────────────────────
	interp.builtins["getent.passwd"] = func(args ...interface{}) (interface{}, error) {
		return sdkgetent.GetPasswd()
	}
	interp.builtins["getent.lookup_user"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("getent.lookup_user() requires 1 argument")
		}
		return sdkgetent.LookupUser(args[0].(string))
	}
	interp.builtins["getent.groups"] = func(args ...interface{}) (interface{}, error) {
		return sdkgetent.GetGroups()
	}
	interp.builtins["getent.lookup_group"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("getent.lookup_group() requires 1 argument")
		}
		return sdkgetent.LookupGroup(args[0].(string))
	}
	interp.builtins["getent.services"] = func(args ...interface{}) (interface{}, error) {
		return sdkgetent.GetServices()
	}
	interp.builtins["getent.lookup_service"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("getent.lookup_service() requires 1 argument")
		}
		return sdkgetent.LookupService(args[0].(string))
	}
	interp.builtins["getent.protocols"] = func(args ...interface{}) (interface{}, error) {
		return sdkgetent.GetProtocols()
	}
	interp.builtins["getent.lookup_protocol"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("getent.lookup_protocol() requires 1 argument")
		}
		return sdkgetent.LookupProtocol(args[0].(string))
	}
	interp.builtins["getent.shells"] = func(args ...interface{}) (interface{}, error) {
		return sdkgetent.Shells()
	}

	// ── haproxy.* ───────────────────────────────────────────────────────
	interp.builtins["haproxy.get_status"] = func(args ...interface{}) (interface{}, error) {
		return sdkhaproxy.GetStatus()
	}
	interp.builtins["haproxy.list_backends"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("haproxy.list_backends() requires 1 argument")
		}
		return sdkhaproxy.ListBackends(args[0].(string))
	}
	interp.builtins["haproxy.enable_backend"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 3 {
			return nil, fmt.Errorf("haproxy.enable_backend() requires 3 arguments")
		}
		return sdkhaproxy.EnableBackend(args[0].(string), args[1].(string), args[2].(string))
	}
	interp.builtins["haproxy.disable_backend"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 3 {
			return nil, fmt.Errorf("haproxy.disable_backend() requires 3 arguments")
		}
		return sdkhaproxy.DisableBackend(args[0].(string), args[1].(string), args[2].(string))
	}
	interp.builtins["haproxy.validate_config"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("haproxy.validate_config() requires 1 argument")
		}
		return sdkhaproxy.ValidateConfig(args[0].(string))
	}
	interp.builtins["haproxy.reload"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("haproxy.reload() requires 1 argument")
		}
		return sdkhaproxy.Reload(args[0].(string))
	}
	interp.builtins["haproxy.restart"] = func(args ...interface{}) (interface{}, error) {
		return sdkhaproxy.Restart()
	}
	interp.builtins["haproxy.version"] = func(args ...interface{}) (interface{}, error) {
		return sdkhaproxy.Version()
	}

	// ── openssl_cert.* ──────────────────────────────────────────────────
	interp.builtins["openssl_cert.create_csr"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 4 {
			return nil, fmt.Errorf("openssl_cert.create_csr() requires 4 arguments")
		}
		kp, _ := args[0].(string)
		cp, _ := args[1].(string)
		subj, _ := args[2].(string)
		bitsF, _ := toFloat(args[3])
		return sdkopenssl.CreateCSR(kp, cp, subj, int(bitsF))
	}
	interp.builtins["openssl_cert.generate_self_signed"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 5 {
			return nil, fmt.Errorf("openssl_cert.generate_self_signed() requires 5 arguments")
		}
		cp, _ := args[0].(string)
		kp, _ := args[1].(string)
		subj, _ := args[2].(string)
		daysF, _ := toFloat(args[3])
		bitsF, _ := toFloat(args[4])
		return sdkopenssl.GenerateSelfSigned(cp, kp, subj, int(daysF), int(bitsF))
	}
	interp.builtins["openssl_cert.inspect"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("openssl_cert.inspect() requires 1 argument")
		}
		cp, _ := args[0].(string)
		return sdkopenssl.Inspect(cp)
	}
	interp.builtins["openssl_cert.verify"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("openssl_cert.verify() requires 2 arguments")
		}
		cp, _ := args[0].(string)
		ca, _ := args[1].(string)
		return sdkopenssl.Verify(cp, ca)
	}
	interp.builtins["openssl_cert.check_expiry"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("openssl_cert.check_expiry() requires 1 argument")
		}
		cp, _ := args[0].(string)
		return sdkopenssl.CheckExpiry(cp)
	}
	interp.builtins["openssl_cert.convert_format"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 3 {
			return nil, fmt.Errorf("openssl_cert.convert_format() requires 3 arguments")
		}
		ip, _ := args[0].(string)
		op, _ := args[1].(string)
		of, _ := args[2].(string)
		return sdkopenssl.ConvertFormat(ip, op, of)
	}

	// ── redis.* ─────────────────────────────────────────────────────────
	interp.builtins["redis.ping"] = func(args ...interface{}) (interface{}, error) {
		h, p, a := "", 0, ""
		if len(args) > 0 {
			h, _ = args[0].(string)
		}
		if len(args) > 1 {
			pf, _ := toFloat(args[1])
			p = int(pf)
		}
		if len(args) > 2 {
			a, _ = args[2].(string)
		}
		return sdkredis.Ping(h, p, a)
	}
	interp.builtins["redis.get"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("redis.get() requires key")
		}
		key, _ := args[0].(string)
		h, p, a := "", 0, ""
		if len(args) > 1 {
			h, _ = args[1].(string)
		}
		if len(args) > 2 {
			pf, _ := toFloat(args[2])
			p = int(pf)
		}
		if len(args) > 3 {
			a, _ = args[3].(string)
		}
		return sdkredis.Get(key, h, p, a)
	}
	interp.builtins["redis.set"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("redis.set() requires key and value")
		}
		key, _ := args[0].(string)
		val, _ := args[1].(string)
		h, p, a, exp := "", 0, "", 0
		if len(args) > 2 {
			h, _ = args[2].(string)
		}
		if len(args) > 3 {
			pf, _ := toFloat(args[3])
			p = int(pf)
		}
		if len(args) > 4 {
			a, _ = args[4].(string)
		}
		if len(args) > 5 {
			ef, _ := toFloat(args[5])
			exp = int(ef)
		}
		return sdkredis.Set(key, val, h, p, a, exp)
	}
	interp.builtins["redis.del"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("redis.del() requires keys")
		}
		// Convert []interface{} to []string
		var keys []string
		if arr, ok := args[0].([]interface{}); ok {
			for _, v := range arr {
				if s, ok := v.(string); ok {
					keys = append(keys, s)
				}
			}
		}
		h, p, a := "", 0, ""
		if len(args) > 1 {
			h, _ = args[1].(string)
		}
		if len(args) > 2 {
			pf, _ := toFloat(args[2])
			p = int(pf)
		}
		if len(args) > 3 {
			a, _ = args[3].(string)
		}
		return sdkredis.Del(keys, h, p, a)
	}
	interp.builtins["redis.keys"] = func(args ...interface{}) (interface{}, error) {
		pat, h, p, a := "*", "", 0, ""
		if len(args) > 0 {
			pat, _ = args[0].(string)
		}
		if len(args) > 1 {
			h, _ = args[1].(string)
		}
		if len(args) > 2 {
			pf, _ := toFloat(args[2])
			p = int(pf)
		}
		if len(args) > 3 {
			a, _ = args[3].(string)
		}
		return sdkredis.Keys(pat, h, p, a)
	}
	interp.builtins["redis.info"] = func(args ...interface{}) (interface{}, error) {
		h, p, a := "", 0, ""
		if len(args) > 0 {
			h, _ = args[0].(string)
		}
		if len(args) > 1 {
			pf, _ := toFloat(args[1])
			p = int(pf)
		}
		if len(args) > 2 {
			a, _ = args[2].(string)
		}
		return sdkredis.Info(h, p, a)
	}
	interp.builtins["redis.flush_db"] = func(args ...interface{}) (interface{}, error) {
		h, p, a := "", 0, ""
		if len(args) > 0 {
			h, _ = args[0].(string)
		}
		if len(args) > 1 {
			pf, _ := toFloat(args[1])
			p = int(pf)
		}
		if len(args) > 2 {
			a, _ = args[2].(string)
		}
		return sdkredis.FlushDB(h, p, a)
	}

	// ── gem.* ───────────────────────────────────────────────────────────
	interp.builtins["gem.install"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("gem.install() requires name")
		}
		name, _ := args[0].(string)
		v, u := "", false
		if len(args) > 1 {
			v, _ = args[1].(string)
		}
		if len(args) > 2 {
			u = opsBool(args[2])
		}
		r, err := sdkgem.Install(name, v, u)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["gem.uninstall"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("gem.uninstall() requires name")
		}
		name, _ := args[0].(string)
		f := false
		if len(args) > 1 {
			f = opsBool(args[1])
		}
		r, err := sdkgem.Uninstall(name, f)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["gem.update"] = func(args ...interface{}) (interface{}, error) {
		n := ""
		if len(args) > 0 {
			n, _ = args[0].(string)
		}
		r, err := sdkgem.Update(n)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["gem.info"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("gem.info() requires name")
		}
		r, err := sdkgem.Info(args[0].(string))
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["gem.list"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkgem.List()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["gem.version"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkgem.Version()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── rabbitmq.* ──────────────────────────────────────────────────────
	interp.builtins["rabbitmq.add_vhost"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("rabbitmq.add_vhost() requires name")
		}
		name, _ := args[0].(string)
		r := sdkrabbitmq.AddVhost(name)
		return r, nil
	}
	interp.builtins["rabbitmq.delete_vhost"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("rabbitmq.delete_vhost() requires name")
		}
		name, _ := args[0].(string)
		r := sdkrabbitmq.DeleteVhost(name)
		return r, nil
	}
	interp.builtins["rabbitmq.list_vhosts"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkrabbitmq.ListVhosts()
		return r, err
	}
	interp.builtins["rabbitmq.add_user"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 3 {
			return nil, fmt.Errorf("rabbitmq.add_user() requires name, password, tags")
		}
		name, _ := args[0].(string)
		pass, _ := args[1].(string)
		tags, _ := args[2].(string)
		r := sdkrabbitmq.AddUser(name, pass, tags)
		return r, nil
	}
	interp.builtins["rabbitmq.delete_user"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("rabbitmq.delete_user() requires name")
		}
		name, _ := args[0].(string)
		r := sdkrabbitmq.DeleteUser(name)
		return r, nil
	}
	interp.builtins["rabbitmq.set_user_tags"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("rabbitmq.set_user_tags() requires name, tags")
		}
		name, _ := args[0].(string)
		tags, _ := args[1].(string)
		r := sdkrabbitmq.SetUserTags(name, tags)
		return r, nil
	}
	interp.builtins["rabbitmq.list_users"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkrabbitmq.ListUsers()
		return r, err
	}
	interp.builtins["rabbitmq.set_permission"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 5 {
			return nil, fmt.Errorf("rabbitmq.set_permission() requires user, vhost, configure, write, read")
		}
		user, _ := args[0].(string)
		vhost, _ := args[1].(string)
		configure, _ := args[2].(string)
		write, _ := args[3].(string)
		read, _ := args[4].(string)
		r := sdkrabbitmq.SetPermission(user, vhost, configure, write, read)
		return r, nil
	}
	interp.builtins["rabbitmq.clear_permission"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("rabbitmq.clear_permission() requires user, vhost")
		}
		user, _ := args[0].(string)
		vhost, _ := args[1].(string)
		r := sdkrabbitmq.ClearPermission(user, vhost)
		return r, nil
	}
	interp.builtins["rabbitmq.set_policy"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 5 {
			return nil, fmt.Errorf("rabbitmq.set_policy() requires name, vhost, pattern, definition, apply_to")
		}
		name, _ := args[0].(string)
		vhost, _ := args[1].(string)
		pattern, _ := args[2].(string)
		definition, _ := args[3].(string)
		applyTo, _ := args[4].(string)
		r := sdkrabbitmq.SetPolicy(name, vhost, pattern, definition, applyTo)
		return r, nil
	}
	interp.builtins["rabbitmq.delete_policy"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("rabbitmq.delete_policy() requires name, vhost")
		}
		name, _ := args[0].(string)
		vhost, _ := args[1].(string)
		r := sdkrabbitmq.DeletePolicy(name, vhost)
		return r, nil
	}
	interp.builtins["rabbitmq.declare_queue"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 5 {
			return nil, fmt.Errorf("rabbitmq.declare_queue() requires name, vhost, queue_type, durable, auto_delete")
		}
		name, _ := args[0].(string)
		vhost, _ := args[1].(string)
		queueType, _ := args[2].(string)
		durable := opsBool(args[3])
		autoDelete := opsBool(args[4])
		r := sdkrabbitmq.DeclareQueue(name, vhost, queueType, durable, autoDelete)
		return r, nil
	}
	interp.builtins["rabbitmq.delete_queue"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("rabbitmq.delete_queue() requires name, vhost")
		}
		name, _ := args[0].(string)
		vhost, _ := args[1].(string)
		r := sdkrabbitmq.DeleteQueue(name, vhost)
		return r, nil
	}
	interp.builtins["rabbitmq.declare_exchange"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 5 {
			return nil, fmt.Errorf("rabbitmq.declare_exchange() requires name, vhost, type, durable, auto_delete")
		}
		name, _ := args[0].(string)
		vhost, _ := args[1].(string)
		exType, _ := args[2].(string)
		durable := opsBool(args[3])
		autoDelete := opsBool(args[4])
		r := sdkrabbitmq.DeclareExchange(name, vhost, exType, durable, autoDelete)
		return r, nil
	}
	interp.builtins["rabbitmq.delete_exchange"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("rabbitmq.delete_exchange() requires name, vhost")
		}
		name, _ := args[0].(string)
		vhost, _ := args[1].(string)
		r := sdkrabbitmq.DeleteExchange(name, vhost)
		return r, nil
	}
	interp.builtins["rabbitmq.bind_queue"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 4 {
			return nil, fmt.Errorf("rabbitmq.bind_queue() requires queue, exchange, vhost, routing_key")
		}
		queue, _ := args[0].(string)
		exchange, _ := args[1].(string)
		vhost, _ := args[2].(string)
		routingKey, _ := args[3].(string)
		r := sdkrabbitmq.BindQueue(queue, exchange, vhost, routingKey)
		return r, nil
	}
	interp.builtins["rabbitmq.unbind_queue"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 4 {
			return nil, fmt.Errorf("rabbitmq.unbind_queue() requires queue, exchange, vhost, routing_key")
		}
		queue, _ := args[0].(string)
		exchange, _ := args[1].(string)
		vhost, _ := args[2].(string)
		routingKey, _ := args[3].(string)
		r := sdkrabbitmq.UnbindQueue(queue, exchange, vhost, routingKey)
		return r, nil
	}
	interp.builtins["rabbitmq.get_status"] = func(args ...interface{}) (interface{}, error) {
		r := sdkrabbitmq.GetStatus()
		return r, nil
	}

	// ── consul.* ────────────────────────────────────────────────────────
	interp.builtins["consul.kv_get"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("consul.kv_get() requires key, addr")
		}
		key, _ := args[0].(string)
		addr, _ := args[1].(string)
		r := sdkconsul.KVGet(key, addr)
		return r, nil
	}
	interp.builtins["consul.kv_put"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 3 {
			return nil, fmt.Errorf("consul.kv_put() requires key, value, addr")
		}
		key, _ := args[0].(string)
		value, _ := args[1].(string)
		addr, _ := args[2].(string)
		r := sdkconsul.KVPut(key, value, addr)
		return r, nil
	}
	interp.builtins["consul.kv_delete"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("consul.kv_delete() requires key, addr")
		}
		key, _ := args[0].(string)
		addr, _ := args[1].(string)
		r := sdkconsul.KVDelete(key, addr)
		return r, nil
	}
	interp.builtins["consul.kv_list"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("consul.kv_list() requires prefix, addr")
		}
		prefix, _ := args[0].(string)
		addr, _ := args[1].(string)
		r, err := sdkconsul.KVList(prefix, addr)
		return r, err
	}
	interp.builtins["consul.service_register"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 5 {
			return nil, fmt.Errorf("consul.service_register() requires name, id, addr, port, consul_addr")
		}
		name, _ := args[0].(string)
		id, _ := args[1].(string)
		addr, _ := args[2].(string)
		port, _ := args[3].(string)
		consulAddr, _ := args[4].(string)
		r := sdkconsul.ServiceRegister(name, id, addr, port, consulAddr)
		return r, nil
	}
	interp.builtins["consul.service_deregister"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("consul.service_deregister() requires id, consul_addr")
		}
		id, _ := args[0].(string)
		consulAddr, _ := args[1].(string)
		r := sdkconsul.ServiceDeregister(id, consulAddr)
		return r, nil
	}
	interp.builtins["consul.members"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("consul.members() requires addr")
		}
		addr, _ := args[0].(string)
		r := sdkconsul.Members(addr)
		return r, nil
	}
	interp.builtins["consul.info"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("consul.info() requires addr")
		}
		addr, _ := args[0].(string)
		r := sdkconsul.Info(addr)
		return r, nil
	}
	interp.builtins["consul.health_check"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("consul.health_check() requires service, addr")
		}
		service, _ := args[0].(string)
		addr, _ := args[1].(string)
		r := sdkconsul.HealthCheck(service, addr)
		return r, nil
	}
	interp.builtins["consul.version"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkconsul.Version()
		return r, err
	}

	// ── memcached.* ─────────────────────────────────────────────────────
	interp.builtins["memcached.get"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 3 {
			return nil, fmt.Errorf("memcached.get() requires key, host, port")
		}
		key, _ := args[0].(string)
		host, _ := args[1].(string)
		pf, _ := toFloat(args[2])
		r := sdkmemcached.Get(key, host, int(pf))
		return r, nil
	}
	interp.builtins["memcached.set"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 5 {
			return nil, fmt.Errorf("memcached.set() requires key, value, host, port, expiry")
		}
		key, _ := args[0].(string)
		value, _ := args[1].(string)
		host, _ := args[2].(string)
		pf, _ := toFloat(args[3])
		ef, _ := toFloat(args[4])
		r := sdkmemcached.Set(key, value, host, int(pf), int(ef))
		return r, nil
	}
	interp.builtins["memcached.delete"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 3 {
			return nil, fmt.Errorf("memcached.delete() requires key, host, port")
		}
		key, _ := args[0].(string)
		host, _ := args[1].(string)
		pf, _ := toFloat(args[2])
		r := sdkmemcached.Delete(key, host, int(pf))
		return r, nil
	}
	interp.builtins["memcached.flush_all"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("memcached.flush_all() requires host, port")
		}
		host, _ := args[0].(string)
		pf, _ := toFloat(args[1])
		r := sdkmemcached.FlushAll(host, int(pf))
		return r, nil
	}
	interp.builtins["memcached.stats"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("memcached.stats() requires host, port")
		}
		host, _ := args[0].(string)
		pf, _ := toFloat(args[1])
		r := sdkmemcached.Stats(host, int(pf))
		return r, nil
	}
	interp.builtins["memcached.version"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("memcached.version() requires host, port")
		}
		host, _ := args[0].(string)
		pf, _ := toFloat(args[1])
		r := sdkmemcached.Version(host, int(pf))
		return r, nil
	}

	// ── selinux.* ────────────────────────────────────────────────────────
	interp.builtins["selinux.get"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkselinux.Get()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["selinux.set"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("selinux.set() requires 1 argument (mode)")
		}
		mode, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("selinux.set(): argument must be string")
		}
		r, err := sdkselinux.Set(mode)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── time.* ─────────────────────────────────────────────────────────
	interp.builtins["time.now"] = func(args ...interface{}) (interface{}, error) {
		r := sdktime.Now()
		return structToMap(r)
	}

	interp.builtins["time.format"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("time.format() requires 2 arguments (unix, layout)")
		}
		unixF, err := toFloat(args[0])
		if err != nil {
			return nil, fmt.Errorf("time.format(): unix must be number")
		}
		layout, strOk := args[1].(string)
		if !strOk {
			return nil, fmt.Errorf("time.format(): layout must be string")
		}
		r, err := sdktime.Format(int64(unixF), layout)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	interp.builtins["time.since"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("time.since() requires 1 argument (unix)")
		}
		unixF, err := toFloat(args[0])
		if err != nil {
			return nil, fmt.Errorf("time.since(): argument must be number")
		}
		r := sdktime.Since(int64(unixF))
		return structToMap(r)
	}

	interp.builtins["time.sleep"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("time.sleep() requires 1 argument (ms)")
		}
		msF, err := toFloat(args[0])
		if err != nil {
			return nil, fmt.Errorf("time.sleep(): argument must be number")
		}
		r, err := sdktime.Sleep(int(msF))
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── json.* ─────────────────────────────────────────────────────────
	interp.builtins["json.encode"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("json.encode() requires 1 argument")
		}
		r, err := sdkjson.Encode(args[0])
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	interp.builtins["json.decode"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("json.decode() requires 1 argument (string)")
		}
		input, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("json.decode(): argument must be string")
		}
		r, err := sdkjson.Decode(input)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── known_hosts.* ────────────────────────────────────────────────────
	interp.builtins["known_hosts.list"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkknownhosts.List()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["known_hosts.check"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("known_hosts.check() requires 1 argument (host)")
		}
		host, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("known_hosts.check(): argument must be string")
		}
		r, err := sdkknownhosts.Check(host)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["known_hosts.add"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("known_hosts.add() requires 1 argument (host)")
		}
		host, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("known_hosts.add(): argument must be string")
		}
		r, err := sdkknownhosts.Add(host)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["known_hosts.remove"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("known_hosts.remove() requires 1 argument (host)")
		}
		host, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("known_hosts.remove(): argument must be string")
		}
		r, err := sdkknownhosts.Remove(host)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── yaml.* ─────────────────────────────────────────────────────────
	interp.builtins["yaml.encode"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("yaml.encode() requires 1 argument")
		}
		r, err := sdkyaml.Encode(args[0])
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	interp.builtins["yaml.decode"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("yaml.decode() requires 1 argument (string)")
		}
		input, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("yaml.decode(): argument must be string")
		}
		r, err := sdkyaml.Decode(input)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── file.* (additions) ────────────────────────────────────────────
	interp.builtins["file.append"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("file.append() requires 2 arguments (path, content)")
		}
		path, _ := args[0].(string)
		content, _ := args[1].(string)
		if path == "" {
			return nil, fmt.Errorf("file.append(): path must be string")
		}
		r, err := sdkfile.Append(path, content)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	interp.builtins["file.chmod"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("file.chmod() requires 2 arguments (path, mode)")
		}
		path, _ := args[0].(string)
		modeStr, _ := args[1].(string)
		var mode uint64
		if _, err := fmt.Sscanf(modeStr, "%o", &mode); err != nil {
			return nil, fmt.Errorf("file.chmod(): mode must be an octal string like \"0755\"")
		}
		r, err := sdkfile.Chmod(path, uint32(mode))
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	interp.builtins["file.template"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("file.template() requires at least 1 argument (path)")
		}
		path, _ := args[0].(string)
		vars := map[string]interface{}{}
		if len(args) >= 2 {
			m, ok := args[1].(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("file.template(): vars must be a dict")
			}
			vars = m
		}
		r, err := sdkfile.Template(path, vars)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── process.kill ─────────────────────────────────────────────────
	interp.builtins["process.kill"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("process.kill() requires at least 1 argument (pid)")
		}
		pidF, err := toFloat(args[0])
		if err != nil {
			return nil, fmt.Errorf("process.kill(): pid must be number")
		}
		signal := "TERM"
		if len(args) >= 2 {
			if s, ok := args[1].(string); ok && s != "" {
				signal = s
			}
		}
		r, err := sdkprocess.Kill(int(pidF), signal)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── service.disable ──────────────────────────────────────────────
	interp.builtins["service.disable"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("service.disable() requires 1 argument (name)")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("service.disable(): argument must be string")
		}
		r, err := sdkservice.Disable(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── pkg.* ────────────────────────────────────────────────────────
	interp.builtins["pkg.install"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("pkg.install() requires 1 argument (name)")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("pkg.install(): argument must be string")
		}
		r, _ := opspkg.Install(name)
		return structToMap(r)
	}
	interp.builtins["pkg.ensure"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("pkg.ensure() requires 1 argument (name)")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("pkg.ensure(): argument must be string")
		}
		r, err := opspkg.Ensure(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	interp.builtins["pkg.owner"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("pkg.owner() requires 1 argument (path)")
		}
		path, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("pkg.owner(): argument must be string")
		}
		r, err := opspkg.Owner(path)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	interp.builtins["pkg.remove"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("pkg.remove() requires 1 argument (name)")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("pkg.remove(): argument must be string")
		}
		r, _ := opspkg.Remove(name)
		return structToMap(r)
	}

	interp.builtins["pkg.info"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("pkg.info() requires 1 argument (name)")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("pkg.info(): argument must be string")
		}
		r, err := opspkg.Info(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	interp.builtins["pkg.list"] = func(args ...interface{}) (interface{}, error) {
		r, err := opspkg.List()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── time.parse / time.diff ───────────────────────────────────────
	interp.builtins["time.parse"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("time.parse() requires 2 arguments (layout, value)")
		}
		layout, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("time.parse(): layout must be string")
		}
		value, ok := args[1].(string)
		if !ok {
			return nil, fmt.Errorf("time.parse(): value must be string")
		}
		r, err := sdktime.Parse(layout, value)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	interp.builtins["time.diff"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("time.diff() requires 2 arguments (t1, t2)")
		}
		t1F, err := toFloat(args[0])
		if err != nil {
			return nil, fmt.Errorf("time.diff(): t1 must be number")
		}
		t2F, err := toFloat(args[1])
		if err != nil {
			return nil, fmt.Errorf("time.diff(): t2 must be number")
		}
		r := sdktime.Diff(int64(t1F), int64(t2F))
		return structToMap(r)
	}

	// ── user.* ─────────────────────────────────────────────────────────
	interp.builtins["user.info"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("user.info() requires 1 argument (username)")
		}
		username, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("user.info(): username must be string")
		}
		r, err := sdkuser.Info(username)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["user.list"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkuser.List()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["user.add"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("user.add() requires at least 1 argument (username)")
		}
		username, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("user.add(): username must be string")
		}
		opts := toStringMap(args, 1)
		r, err := sdkuser.Add(username, opts)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["user.ensure"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("user.ensure() requires at least 1 argument (username)")
		}
		username, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("user.ensure(): username must be string")
		}
		opts := toStringMap(args, 1)
		r, err := sdkuser.Ensure(username, opts)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["user.absent"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("user.absent() requires at least 1 argument (username)")
		}
		username, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("user.absent(): username must be string")
		}
		removeHome := false
		if len(args) > 1 {
			removeHome, _ = args[1].(bool)
		}
		r, err := sdkuser.Absent(username, removeHome)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["user.remove"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("user.remove() requires at least 1 argument (username)")
		}
		username, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("user.remove(): username must be string")
		}
		removeHome := false
		if len(args) > 1 {
			removeHome, _ = args[1].(bool)
		}
		r, err := sdkuser.Remove(username, removeHome)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["user.modify"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("user.modify() requires at least 1 argument (username)")
		}
		username, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("user.modify(): username must be string")
		}
		opts := toStringMap(args, 1)
		r, err := sdkuser.Modify(username, opts)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["user.exists"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("user.exists() requires 1 argument (username)")
		}
		username, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("user.exists(): username must be string")
		}
		r, err := sdkuser.Exists(username)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── group.* ────────────────────────────────────────────────────────
	interp.builtins["group.info"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("group.info() requires 1 argument (name)")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("group.info(): name must be string")
		}
		r, err := sdkgroup.Info(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["group.list"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkgroup.List()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["group.add"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("group.add() requires at least 1 argument (name)")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("group.add(): name must be string")
		}
		opts := toStringMap(args, 1)
		r, err := sdkgroup.Add(name, opts)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["group.ensure"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("group.ensure() requires at least 1 argument (name)")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("group.ensure(): name must be string")
		}
		opts := toStringMap(args, 1)
		r, err := sdkgroup.Ensure(name, opts)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["group.absent"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("group.absent() requires 1 argument (name)")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("group.absent(): name must be string")
		}
		r, err := sdkgroup.Absent(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["group.remove"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("group.remove() requires 1 argument (name)")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("group.remove(): name must be string")
		}
		r, err := sdkgroup.Remove(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["group.exists"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("group.exists() requires 1 argument (name)")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("group.exists(): name must be string")
		}
		r, err := sdkgroup.Exists(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── cron.* ─────────────────────────────────────────────────────────
	interp.builtins["cron.list"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("cron.list() requires 1 argument (user)")
		}
		user, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("cron.list(): user must be string")
		}
		r, err := sdkcron.List(user)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["cron.add"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("cron.add() requires 2 arguments (user, entry)")
		}
		user, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("cron.add(): user must be string")
		}
		entryMap, ok := args[1].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("cron.add(): entry must be a dict")
		}
		entry := sdkcron.CronEntry{
			Minute:     mapStr(entryMap, "minute", "*"),
			Hour:       mapStr(entryMap, "hour", "*"),
			DayOfMonth: mapStr(entryMap, "day_of_month", "*"),
			Month:      mapStr(entryMap, "month", "*"),
			DayOfWeek:  mapStr(entryMap, "day_of_week", "*"),
			Command:    mapStr(entryMap, "command", ""),
		}
		r, err := sdkcron.Add(user, entry)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["cron.remove"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("cron.remove() requires 2 arguments (user, line_match)")
		}
		user, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("cron.remove(): user must be string")
		}
		lineMatch, ok := args[1].(string)
		if !ok {
			return nil, fmt.Errorf("cron.remove(): line_match must be string")
		}
		r, err := sdkcron.Remove(user, lineMatch)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── sysctl.* ───────────────────────────────────────────────────────
	interp.builtins["sysctl.get"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("sysctl.get() requires 1 argument (name)")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("sysctl.get(): name must be string")
		}
		r, err := sdksysctl.Get(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["sysctl.set"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("sysctl.set() requires 2 arguments (name, value)")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("sysctl.set(): name must be string")
		}
		value, ok := args[1].(string)
		if !ok {
			return nil, fmt.Errorf("sysctl.set(): value must be string")
		}
		r, err := sdksysctl.Set(name, value)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["sysctl.list"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdksysctl.List()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── git.* ──────────────────────────────────────────────────────────
	interp.builtins["git.clone"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("git.clone() requires at least 2 arguments (url, dest)")
		}
		url, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("git.clone(): url must be string")
		}
		dest, ok := args[1].(string)
		if !ok {
			return nil, fmt.Errorf("git.clone(): dest must be string")
		}
		opts := toStringMap(args, 2)
		r, err := sdkgit.Clone(url, dest, opts)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["git.pull"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("git.pull() requires at least 1 argument (repo_path)")
		}
		repoPath, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("git.pull(): repo_path must be string")
		}
		remote := getStringArgBridge(args, 1, "origin")
		branch := getStringArgBridge(args, 2, "")
		r, err := sdkgit.Pull(repoPath, remote, branch)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── file.lineinfile ────────────────────────────────────────────────
	interp.builtins["file.lineinfile"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("file.lineinfile() requires at least 2 arguments (path, line)")
		}
		path, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("file.lineinfile(): path must be string")
		}
		line, ok := args[1].(string)
		if !ok {
			return nil, fmt.Errorf("file.lineinfile(): line must be string")
		}
		present := true
		if len(args) > 2 {
			present, _ = args[2].(bool)
		}
		rx := ""
		if len(args) > 3 {
			rx, _ = args[3].(string)
		}
		r, err := sdkfile.LineInFile(path, line, present, rx)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── net.wait_for ───────────────────────────────────────────────────
	interp.builtins["net.wait_for"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("net.wait_for() requires at least 2 arguments (host, port)")
		}
		host, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("net.wait_for(): host must be string")
		}
		portF, err := toFloat(args[1])
		if err != nil {
			return nil, fmt.Errorf("net.wait_for(): port must be number")
		}
		timeout := 30
		if len(args) > 2 {
			tF, err := toFloat(args[2])
			if err != nil {
				return nil, fmt.Errorf("net.wait_for(): timeout must be number")
			}
			timeout = int(tF)
		}
		r, err := sdknet.WaitFor(host, int(portF), timeout)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── sys.mount / sys.unmount / sys.list_mounts ──────────────────────
	interp.builtins["sys.mount"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 3 {
			return nil, fmt.Errorf("sys.mount() requires at least 3 arguments (device, mountpoint, fs_type)")
		}
		device, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("sys.mount(): device must be string")
		}
		mountpoint, ok := args[1].(string)
		if !ok {
			return nil, fmt.Errorf("sys.mount(): mountpoint must be string")
		}
		fsType, ok := args[2].(string)
		if !ok {
			return nil, fmt.Errorf("sys.mount(): fs_type must be string")
		}
		opts := toStringMap(args, 3)
		r, err := sdksys.Mount(device, mountpoint, fsType, opts)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["sys.unmount"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("sys.unmount() requires 1 argument (mountpoint)")
		}
		mountpoint, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("sys.unmount(): mountpoint must be string")
		}
		r, err := sdksys.Unmount(mountpoint)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["sys.list_mounts"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdksys.ListMounts()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── sys.hostname_set ───────────────────────────────────────────────
	interp.builtins["sys.hostname_set"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("sys.hostname_set() requires 1 argument (name)")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("sys.hostname_set(): name must be string")
		}
		r, err := sdksys.HostnameSet(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── firewall.rule ──────────────────────────────────────────────────
	interp.builtins["firewall.rule"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("firewall.rule() requires at least 2 arguments (action, protocol)")
		}
		action, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("firewall.rule(): action must be string")
		}
		protocol, ok := args[1].(string)
		if !ok {
			return nil, fmt.Errorf("firewall.rule(): protocol must be string")
		}
		port := 0
		if len(args) > 2 {
			pF, err := toFloat(args[2])
			if err != nil {
				return nil, fmt.Errorf("firewall.rule(): port must be number")
			}
			port = int(pF)
		}
		source := ""
		if len(args) > 3 {
			source, _ = args[3].(string)
		}
		r, err := sdksys.FirewallRule(action, protocol, port, source)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── firewalld ────────────────────────────────────────────────────────
	interp.builtins["firewalld.get"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkfirewalld.Get()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["firewalld.start"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkfirewalld.Start()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["firewalld.stop"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkfirewalld.Stop()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["firewalld.restart"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkfirewalld.Restart()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["firewalld.enable"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkfirewalld.Enable()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["firewalld.disable"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkfirewalld.Disable()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["firewalld.list_zones"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkfirewalld.ListZones()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["firewalld.reload"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkfirewalld.Reload()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── file.find ────────────────────────────────────────────────────────
	interp.builtins["file.find"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("file.find() requires at least 1 argument (paths)")
		}
		opts := sdkfile.FindOptions{}
		// paths: string or []string
		switch v := args[0].(type) {
		case string:
			opts.Paths = []string{v}
		case []interface{}:
			for _, p := range v {
				if s, ok := p.(string); ok {
					opts.Paths = append(opts.Paths, s)
				}
			}
		case []string:
			opts.Paths = v
		}
		// patterns: optional string or []string
		if len(args) > 1 {
			switch v := args[1].(type) {
			case string:
				if v != "" {
					opts.Patterns = []string{v}
				}
			case []interface{}:
				for _, p := range v {
					if s, ok := p.(string); ok {
						opts.Patterns = append(opts.Patterns, s)
					}
				}
			case []string:
				opts.Patterns = v
			}
		}
		// regex
		if len(args) > 2 {
			if s, ok := args[2].(string); ok {
				opts.Regex = s
			}
		}
		// file_type
		if len(args) > 3 {
			if s, ok := args[3].(string); ok {
				opts.FileType = s
			}
		}
		// max_depth
		if len(args) > 4 {
			if f, err := toFloat(args[4]); err == nil {
				opts.MaxDepth = int(f)
			}
		}
		// age
		if len(args) > 5 {
			if f, err := toFloat(args[5]); err == nil {
				opts.Age = int64(f)
			}
		}
		// size
		if len(args) > 6 {
			if f, err := toFloat(args[6]); err == nil {
				opts.Size = int64(f)
			}
		}
		r, err := sdkfile.Find(opts)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── file.replace ─────────────────────────────────────────────────────
	interp.builtins["file.replace"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 3 {
			return nil, fmt.Errorf("file.replace() requires at least 3 arguments (path, pattern, replacement)")
		}
		path, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("file.replace(): path must be string")
		}
		pattern, ok := args[1].(string)
		if !ok {
			return nil, fmt.Errorf("file.replace(): pattern must be string")
		}
		replacement, ok := args[2].(string)
		if !ok {
			return nil, fmt.Errorf("file.replace(): replacement must be string")
		}
		after := getStringArgBridge(args, 3, "")
		before := getStringArgBridge(args, 4, "")
		r, err := sdkfile.Replace(path, pattern, replacement, after, before)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── file.blockinfile ─────────────────────────────────────────────────
	interp.builtins["file.blockinfile"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 3 {
			return nil, fmt.Errorf("file.blockinfile() requires at least 3 arguments (path, marker, content)")
		}
		path, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("file.blockinfile(): path must be string")
		}
		marker, ok := args[1].(string)
		if !ok {
			return nil, fmt.Errorf("file.blockinfile(): marker must be string")
		}
		content, ok := args[2].(string)
		if !ok {
			return nil, fmt.Errorf("file.blockinfile(): content must be string")
		}
		present := true
		if len(args) > 3 {
			present = opsBool(args[3])
		}
		insertAfter := getStringArgBridge(args, 4, "")
		insertBefore := getStringArgBridge(args, 5, "")
		r, err := sdkfile.BlockInFile(path, marker, content, present, insertAfter, insertBefore)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── file.ini_get ─────────────────────────────────────────────────────
	interp.builtins["file.ini_get"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 3 {
			return nil, fmt.Errorf("file.ini_get() requires 3 arguments (path, section, key)")
		}
		path, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("file.ini_get(): path must be string")
		}
		section, ok := args[1].(string)
		if !ok {
			return nil, fmt.Errorf("file.ini_get(): section must be string")
		}
		key, ok := args[2].(string)
		if !ok {
			return nil, fmt.Errorf("file.ini_get(): key must be string")
		}
		r, err := sdkfile.IniGet(path, section, key)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── file.ini_set ─────────────────────────────────────────────────────
	interp.builtins["file.ini_set"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 4 {
			return nil, fmt.Errorf("file.ini_set() requires 4 arguments (path, section, key, value)")
		}
		path, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("file.ini_set(): path must be string")
		}
		section, ok := args[1].(string)
		if !ok {
			return nil, fmt.Errorf("file.ini_set(): section must be string")
		}
		key, ok := args[2].(string)
		if !ok {
			return nil, fmt.Errorf("file.ini_set(): key must be string")
		}
		value, ok := args[3].(string)
		if !ok {
			return nil, fmt.Errorf("file.ini_set(): value must be string")
		}
		r, err := sdkfile.IniSet(path, section, key, value)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── archive.create ───────────────────────────────────────────────────
	interp.builtins["archive.create"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("archive.create() requires 2 arguments (dest, sources)")
		}
		dest, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("archive.create(): dest must be string")
		}
		var sources []string
		switch v := args[1].(type) {
		case string:
			sources = []string{v}
		case []interface{}:
			for _, s := range v {
				if str, ok := s.(string); ok {
					sources = append(sources, str)
				}
			}
		case []string:
			sources = v
		default:
			return nil, fmt.Errorf("archive.create(): sources must be string or list")
		}
		r, err := sdkarchive.Create(dest, sources)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── archive.extract ──────────────────────────────────────────────────
	interp.builtins["archive.extract"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("archive.extract() requires 2 arguments (src, dest)")
		}
		src, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("archive.extract(): src must be string")
		}
		dest, ok := args[1].(string)
		if !ok {
			return nil, fmt.Errorf("archive.extract(): dest must be string")
		}
		r, err := sdkarchive.Extract(src, dest)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── net.download ─────────────────────────────────────────────────────
	interp.builtins["net.download"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("net.download() requires at least 2 arguments (url, dest)")
		}
		url, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("net.download(): url must be string")
		}
		dest, ok := args[1].(string)
		if !ok {
			return nil, fmt.Errorf("net.download(): dest must be string")
		}
		algo := getStringArgBridge(args, 2, "")
		expected := getStringArgBridge(args, 3, "")
		r, err := sdknet.Download(url, dest, algo, expected)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── net.wait_for_connection ──────────────────────────────────────────
	interp.builtins["net.wait_for_connection"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("net.wait_for_connection() requires at least 2 arguments (host, port)")
		}
		host, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("net.wait_for_connection(): host must be string")
		}
		portF, err := toFloat(args[1])
		if err != nil {
			return nil, fmt.Errorf("net.wait_for_connection(): port must be number")
		}
		port := int(portF)
		timeout := 30
		if len(args) > 2 {
			tF, err := toFloat(args[2])
			if err != nil {
				return nil, fmt.Errorf("net.wait_for_connection(): timeout must be number")
			}
			timeout = int(tF)
		}
		r, err := sdknet.WaitForConnection(host, port, timeout)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── ntp.* ────────────────────────────────────────────────────────────
	interp.builtins["ntp.get"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkntp.Get()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["ntp.set"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("ntp.set() requires 1 argument (server)")
		}
		server, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("ntp.set(): argument must be string")
		}
		r, err := sdkntp.Set(server)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── sys.timezone_get ─────────────────────────────────────────────────
	interp.builtins["sys.timezone_get"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdksys.TimezoneGet()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── sys.timezone_set ─────────────────────────────────────────────────
	interp.builtins["sys.timezone_set"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("sys.timezone_set() requires 1 argument (timezone)")
		}
		tz, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("sys.timezone_set(): timezone must be string")
		}
		r, err := sdksys.TimezoneSet(tz)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── sys.reboot ───────────────────────────────────────────────────────
	interp.builtins["sys.reboot"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdksys.Reboot()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── ssh.authorized_key_add ───────────────────────────────────────────
	interp.builtins["ssh.authorized_key_add"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("ssh.authorized_key_add() requires at least 2 arguments (user, key)")
		}
		user, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("ssh.authorized_key_add(): user must be string")
		}
		key, ok := args[1].(string)
		if !ok {
			return nil, fmt.Errorf("ssh.authorized_key_add(): key must be string")
		}
		exclusive := false
		if len(args) > 2 {
			exclusive = opsBool(args[2])
		}
		r, err := sdkssh.AuthorizedKeyAdd(user, key, exclusive)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── ssh.authorized_key_remove ────────────────────────────────────────
	interp.builtins["ssh.authorized_key_remove"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("ssh.authorized_key_remove() requires 2 arguments (user, key)")
		}
		user, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("ssh.authorized_key_remove(): user must be string")
		}
		key, ok := args[1].(string)
		if !ok {
			return nil, fmt.Errorf("ssh.authorized_key_remove(): key must be string")
		}
		r, err := sdkssh.AuthorizedKeyRemove(user, key)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── ssh.authorized_key_list ──────────────────────────────────────────
	interp.builtins["ssh.authorized_key_list"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("ssh.authorized_key_list() requires 1 argument (user)")
		}
		user, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("ssh.authorized_key_list(): user must be string")
		}
		r, err := sdkssh.AuthorizedKeyList(user)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── kernel.module_list ───────────────────────────────────────────────
	interp.builtins["kernel.module_list"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkkernel.ModuleList()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── kernel.module_load ───────────────────────────────────────────────
	interp.builtins["kernel.module_load"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("kernel.module_load() requires 1 argument (name)")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("kernel.module_load(): name must be string")
		}
		r, err := sdkkernel.ModuleLoad(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── kernel.module_unload ─────────────────────────────────────────────
	interp.builtins["kernel.module_unload"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("kernel.module_unload() requires 1 argument (name)")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("kernel.module_unload(): name must be string")
		}
		r, err := sdkkernel.ModuleUnload(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── limits.* ─────────────────────────────────────────────────────────
	interp.builtins["limits.list"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdklimits.List()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["limits.get"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("limits.get() requires 1 argument (domain)")
		}
		domain, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("limits.get(): argument must be string")
		}
		r, err := sdklimits.Get(domain)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["limits.set"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 4 {
			return nil, fmt.Errorf("limits.set() requires 4 arguments (domain, type, item, value)")
		}
		domain, _ := args[0].(string)
		typ, _ := args[1].(string)
		item, _ := args[2].(string)
		value, _ := args[3].(string)
		r, err := sdklimits.Set(domain, typ, item, value)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["limits.remove"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("limits.remove() requires 1 argument (domain)")
		}
		domain, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("limits.remove(): argument must be string")
		}
		r, err := sdklimits.Remove(domain)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── disk.filesystem ──────────────────────────────────────────────────
	interp.builtins["disk.filesystem"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("disk.filesystem() requires at least 1 argument (device)")
		}
		device, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("disk.filesystem(): device must be string")
		}
		fsType := getStringArgBridge(args, 1, "ext4")
		r, err := sdkdisk.FilesystemCreate(device, fsType)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── disk.part_list ───────────────────────────────────────────────────
	interp.builtins["disk.part_list"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("disk.part_list() requires 1 argument (device)")
		}
		device, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("disk.part_list(): device must be string")
		}
		r, err := sdkdisk.PartList(device)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── docker.* ────────────────────────────────────────────────────────
	interp.builtins["docker.container_list"] = func(args ...interface{}) (interface{}, error) {
		all := false
		if len(args) > 0 {
			all, _ = args[0].(bool)
		}
		r, err := sdkdocker.ContainerList(all)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["docker.container_exists"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("docker.container_exists() requires 1 argument (name)")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("docker.container_exists(): name must be string")
		}
		r, err := sdkdocker.ContainerExists(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["docker.container_run"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("docker.container_run() requires at least 2 arguments (name, image)")
		}
		name, _ := args[0].(string)
		image, ok := args[1].(string)
		if !ok {
			return nil, fmt.Errorf("docker.container_run(): image must be string")
		}
		opts := toStringMap(args, 2)
		r, err := sdkdocker.ContainerRun(name, image, opts)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["docker.container_stop"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("docker.container_stop() requires 1 argument (name)")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("docker.container_stop(): name must be string")
		}
		r, err := sdkdocker.ContainerStop(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["docker.container_remove"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("docker.container_remove() requires at least 1 argument (name)")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("docker.container_remove(): name must be string")
		}
		force := false
		if len(args) > 1 {
			force, _ = args[1].(bool)
		}
		r, err := sdkdocker.ContainerRemove(name, force)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["docker.image_list"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkdocker.ImageList()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["docker.image_pull"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("docker.image_pull() requires 1 argument (image)")
		}
		image, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("docker.image_pull(): image must be string")
		}
		r, err := sdkdocker.ImagePull(image)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["docker.image_remove"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("docker.image_remove() requires at least 1 argument (image)")
		}
		image, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("docker.image_remove(): image must be string")
		}
		force := false
		if len(args) > 1 {
			force, _ = args[1].(bool)
		}
		r, err := sdkdocker.ImageRemove(image, force)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── hosts.* ─────────────────────────────────────────────────────────
	interp.builtins["hosts.list"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkhosts.List()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["hosts.exists"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("hosts.exists() requires 1 argument (hostname)")
		}
		hostname, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("hosts.exists(): hostname must be string")
		}
		r, err := sdkhosts.Exists(hostname)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["hosts.add"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("hosts.add() requires 2 arguments (ip, hostnames)")
		}
		ip, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("hosts.add(): ip must be string")
		}
		hostnamesRaw, ok := args[1].([]interface{})
		if !ok {
			return nil, fmt.Errorf("hosts.add(): hostnames must be array")
		}
		hostnames := make([]string, len(hostnamesRaw))
		for i, h := range hostnamesRaw {
			hostnames[i], _ = h.(string)
		}
		r, err := sdkhosts.Add(ip, hostnames)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["hosts.remove"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("hosts.remove() requires 1 argument (hostnames)")
		}
		hostnamesRaw, ok := args[0].([]interface{})
		if !ok {
			return nil, fmt.Errorf("hosts.remove(): hostnames must be array")
		}
		hostnames := make([]string, len(hostnamesRaw))
		for i, h := range hostnamesRaw {
			hostnames[i], _ = h.(string)
		}
		r, err := sdkhosts.Remove(hostnames)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── locale.* ────────────────────────────────────────────────────────
	interp.builtins["locale.get"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdklocale.Get()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["locale.available"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdklocale.Available()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["locale.set"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("locale.set() requires 1 argument (locale)")
		}
		locale, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("locale.set(): locale must be string")
		}
		r, err := sdklocale.Set(locale)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── pip.* ───────────────────────────────────────────────────────────
	interp.builtins["pip.list"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkpip.List()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["pip.exists"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("pip.exists() requires 1 argument (name)")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("pip.exists(): name must be string")
		}
		r, err := sdkpip.Exists(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["pip.install"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("pip.install() requires at least 1 argument (name)")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("pip.install(): name must be string")
		}
		version := ""
		if len(args) > 1 {
			version, _ = args[1].(string)
		}
		r, err := sdkpip.Install(name, version)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["pip.uninstall"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("pip.uninstall() requires 1 argument (name)")
		}
		name, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("pip.uninstall(): name must be string")
		}
		r, err := sdkpip.Uninstall(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── apt.* ──────────────────────────────────────────────────────────
	interp.builtins["apt.install"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("apt.install() requires at least 1 argument (name)")
		}
		name, _ := args[0].(string)
		version, _ := args[1].(string)
		updateCache, _ := args[2].(bool)
		r, err := sdkapt.Install(name, version, updateCache)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["apt.remove"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("apt.remove() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		purge, _ := args[1].(bool)
		r, err := sdkapt.Remove(name, purge)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["apt.upgrade"] = func(args ...interface{}) (interface{}, error) {
		name, _ := args[0].(string)
		r, err := sdkapt.Upgrade(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["apt.update_cache"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkapt.UpdateCache()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["apt.full_upgrade"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkapt.FullUpgrade()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["apt.dist_upgrade"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkapt.DistUpgrade()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["apt.autoremove"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkapt.Autoremove()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["apt.clean"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkapt.Clean()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["apt.info"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("apt.info() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		r, err := sdkapt.Info(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["apt.list"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkapt.List()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["apt.policy"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("apt.policy() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		r, err := sdkapt.Policy(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["apt.mark_auto"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("apt.mark_auto() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		r, err := sdkapt.MarkAuto(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["apt.mark_manual"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("apt.mark_manual() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		r, err := sdkapt.MarkManual(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── dnf.* ─────────────────────────────────────────────────────────
	interp.builtins["dnf.install"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("dnf.install() requires at least 1 argument (name)")
		}
		name, _ := args[0].(string)
		version := ""
		if len(args) >= 2 {
			version, _ = args[1].(string)
		}
		r, err := sdkdnf.Install(name, version)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["dnf.remove"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("dnf.remove() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		r, err := sdkdnf.Remove(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["dnf.update"] = func(args ...interface{}) (interface{}, error) {
		name := ""
		if len(args) >= 1 {
			name, _ = args[0].(string)
		}
		r, err := sdkdnf.Update(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["dnf.info"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("dnf.info() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		r, err := sdkdnf.Info(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["dnf.list"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkdnf.List()
		if err != nil {
			return nil, err
		}
		return r, nil
	}
	interp.builtins["dnf.search"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("dnf.search() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		r, err := sdkdnf.Search(name)
		if err != nil {
			return nil, err
		}
		return r, nil
	}
	interp.builtins["dnf.clean"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkdnf.Clean()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["dnf.repolist"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkdnf.RepoList()
		if err != nil {
			return nil, err
		}
		return r, nil
	}
	interp.builtins["dnf.grouplist"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkdnf.GroupList()
		if err != nil {
			return nil, err
		}
		return r, nil
	}
	interp.builtins["dnf.groupinstall"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("dnf.groupinstall() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		r, err := sdkdnf.GroupInstall(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["dnf.groupremove"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("dnf.groupremove() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		r, err := sdkdnf.GroupRemove(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["dnf.history"] = func(args ...interface{}) (interface{}, error) {
		count := 10
		if len(args) >= 1 {
			if v, ok := args[0].(int); ok {
				count = v
			}
		}
		r, err := sdkdnf.History(count)
		if err != nil {
			return nil, err
		}
		return r, nil
	}
	interp.builtins["dnf.check_update"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkdnf.CheckUpdate()
		if err != nil {
			return nil, err
		}
		return r, nil
	}
	interp.builtins["dnf.modulelist"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkdnf.ModuleList()
		if err != nil {
			return nil, err
		}
		return r, nil
	}
	interp.builtins["dnf.module_enable"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("dnf.module_enable() requires 1 argument (spec)")
		}
		spec, _ := args[0].(string)
		r, err := sdkdnf.ModuleEnable(spec)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── apk.* ─────────────────────────────────────────────────────────
	interp.builtins["apk.install"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("apk.install() requires at least 1 argument (name)")
		}
		name, _ := args[0].(string)
		version := ""
		if len(args) >= 2 {
			version, _ = args[1].(string)
		}
		r, err := sdkapk.Install(name, version)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["apk.remove"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("apk.remove() requires at least 1 argument (name)")
		}
		name, _ := args[0].(string)
		purge := false
		if len(args) >= 2 {
			purge, _ = args[1].(bool)
		}
		r, err := sdkapk.Remove(name, purge)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["apk.update"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkapk.Update()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["apk.upgrade"] = func(args ...interface{}) (interface{}, error) {
		name := ""
		if len(args) >= 1 {
			name, _ = args[0].(string)
		}
		r, err := sdkapk.Upgrade(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["apk.info"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("apk.info() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		r, err := sdkapk.Info(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["apk.list"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkapk.List()
		if err != nil {
			return nil, err
		}
		return r, nil
	}
	interp.builtins["apk.search"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("apk.search() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		r, err := sdkapk.Search(name)
		if err != nil {
			return nil, err
		}
		return r, nil
	}
	interp.builtins["apk.cache"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkapk.Cache()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["apk.upgrade_available"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkapk.UpgradeAvailable()
		if err != nil {
			return nil, err
		}
		return r, nil
	}
	interp.builtins["apk.repository"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkapk.Repository()
		if err != nil {
			return nil, err
		}
		return r, nil
	}

	// ── sysvinit.* ────────────────────────────────────────────────────
	interp.builtins["sysvinit.status"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("sysvinit.status() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		r, err := sdksysvinit.Status(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["sysvinit.start"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("sysvinit.start() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		r, err := sdksysvinit.Start(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["sysvinit.stop"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("sysvinit.stop() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		r, err := sdksysvinit.Stop(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["sysvinit.restart"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("sysvinit.restart() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		r, err := sdksysvinit.Restart(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["sysvinit.reload"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("sysvinit.reload() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		r, err := sdksysvinit.Reload(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["sysvinit.enable"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("sysvinit.enable() requires at least 1 argument (name)")
		}
		name, _ := args[0].(string)
		runlevels := ""
		if len(args) >= 2 {
			runlevels, _ = args[1].(string)
		}
		r, err := sdksysvinit.Enable(name, runlevels)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["sysvinit.disable"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("sysvinit.disable() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		r, err := sdksysvinit.Disable(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["sysvinit.list"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdksysvinit.List()
		if err != nil {
			return nil, err
		}
		return r, nil
	}

	// ── runit.* ─────────────────────────────────────────────────────
	interp.builtins["runit.status"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("runit.status() requires 1 argument (service)")
		}
		svc, _ := args[0].(string)
		r, err := sdkrunit.Status(svc)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["runit.start"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("runit.start() requires 1 argument (service)")
		}
		svc, _ := args[0].(string)
		r, err := sdkrunit.Start(svc)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["runit.stop"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("runit.stop() requires 1 argument (service)")
		}
		svc, _ := args[0].(string)
		r, err := sdkrunit.Stop(svc)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["runit.restart"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("runit.restart() requires 1 argument (service)")
		}
		svc, _ := args[0].(string)
		r, err := sdkrunit.Restart(svc)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["runit.reload"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("runit.reload() requires 1 argument (service)")
		}
		svc, _ := args[0].(string)
		r, err := sdkrunit.Reload(svc)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["runit.enable"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("runit.enable() requires at least 1 argument (service)")
		}
		svc, _ := args[0].(string)
		svcDir := ""
		if len(args) >= 2 {
			svcDir, _ = args[1].(string)
		}
		r, err := sdkrunit.Enable(svc, svcDir)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["runit.disable"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("runit.disable() requires 1 argument (service)")
		}
		svc, _ := args[0].(string)
		r, err := sdkrunit.Disable(svc)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["runit.list"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkrunit.List()
		if err != nil {
			return nil, err
		}
		return r, nil
	}

	// ── fail2ban.* ────────────────────────────────────────────────────
	interp.builtins["fail2ban.get"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkfail2ban.Get()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["fail2ban.jail_status"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("fail2ban.jail_status() requires 1 argument (jail)")
		}
		jail, _ := args[0].(string)
		r, err := sdkfail2ban.JailStatus(jail)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["fail2ban.start"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkfail2ban.Start()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["fail2ban.stop"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkfail2ban.Stop()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["fail2ban.reload"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkfail2ban.Reload()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["fail2ban.ban_ip"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("fail2ban.ban_ip() requires 2 arguments (jail, ip)")
		}
		jail, _ := args[0].(string)
		ip, _ := args[1].(string)
		r, err := sdkfail2ban.BanIP(jail, ip)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["fail2ban.unban_ip"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("fail2ban.unban_ip() requires 2 arguments (jail, ip)")
		}
		jail, _ := args[0].(string)
		ip, _ := args[1].(string)
		r, err := sdkfail2ban.UnbanIP(jail, ip)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── lsb_release.* ─────────────────────────────────────────────────
	interp.builtins["lsb_release.get"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdklsb.Get()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── docker_compose.* ──────────────────────────────────────────────
	interp.builtins["docker_compose.up"] = func(args ...interface{}) (interface{}, error) {
		dir := ""
		if len(args) >= 1 {
			dir, _ = args[0].(string)
		}
		r, err := sdkcompose.Up(dir)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["docker_compose.down"] = func(args ...interface{}) (interface{}, error) {
		dir := ""
		if len(args) >= 1 {
			dir, _ = args[0].(string)
		}
		r, err := sdkcompose.Down(dir)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["docker_compose.restart"] = func(args ...interface{}) (interface{}, error) {
		dir := ""
		if len(args) >= 1 {
			dir, _ = args[0].(string)
		}
		r, err := sdkcompose.Restart(dir)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["docker_compose.pull"] = func(args ...interface{}) (interface{}, error) {
		dir := ""
		if len(args) >= 1 {
			dir, _ = args[0].(string)
		}
		r, err := sdkcompose.Pull(dir)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["docker_compose.status"] = func(args ...interface{}) (interface{}, error) {
		dir := ""
		if len(args) >= 1 {
			dir, _ = args[0].(string)
		}
		r, err := sdkcompose.Status(dir)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["docker_compose.build"] = func(args ...interface{}) (interface{}, error) {
		dir := ""
		if len(args) >= 1 {
			dir, _ = args[0].(string)
		}
		r, err := sdkcompose.Build(dir)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["docker_compose.logs"] = func(args ...interface{}) (interface{}, error) {
		dir := ""
		tail := 100
		if len(args) >= 1 {
			dir, _ = args[0].(string)
		}
		if len(args) >= 2 {
			if v, ok := args[1].(int); ok {
				tail = v
			}
		}
		output, err := sdkcompose.Logs(dir, tail)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"logs": output}, nil
	}

	// ── cloud_init.* ──────────────────────────────────────────────────
	interp.builtins["cloud_init.status"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkcloudinit.Status()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["cloud_init.modules"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkcloudinit.Modules()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["cloud_init.clean"] = func(args ...interface{}) (interface{}, error) {
		removeLogs := false
		if len(args) >= 1 {
			removeLogs, _ = args[0].(bool)
		}
		r, err := sdkcloudinit.Clean(removeLogs)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["cloud_init.init"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkcloudinit.Init()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── sys_persist.* ─────────────────────────────────────────────────
	interp.builtins["sys_persist.set"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("sys_persist.set() requires 2 arguments (name, value)")
		}
		name, _ := args[0].(string)
		value, _ := args[1].(string)
		r, err := sdksyspersist.Set(name, value)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["sys_persist.get"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("sys_persist.get() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		r, err := sdksyspersist.Get(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["sys_persist.remove"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("sys_persist.remove() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		r, err := sdksyspersist.Remove(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["sys_persist.list"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdksyspersist.List()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── wireguard.* ───────────────────────────────────────────────────
	interp.builtins["wireguard.show"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkwireguard.Show()
		if err != nil {
			return nil, err
		}
		return r, nil
	}
	interp.builtins["wireguard.up"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("wireguard.up() requires at least 1 argument (interface)")
		}
		iface, _ := args[0].(string)
		cfg := ""
		if len(args) >= 2 {
			cfg, _ = args[1].(string)
		}
		r, err := sdkwireguard.Up(iface, cfg)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["wireguard.down"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("wireguard.down() requires 1 argument (interface)")
		}
		iface, _ := args[0].(string)
		r, err := sdkwireguard.Down(iface)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["wireguard.add_peer"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("wireguard.add_peer() requires at least 2 arguments (interface, public_key)")
		}
		iface, _ := args[0].(string)
		pubkey, _ := args[1].(string)
		allowedIPs, endpoint := "", ""
		if len(args) >= 3 {
			allowedIPs, _ = args[2].(string)
		}
		if len(args) >= 4 {
			endpoint, _ = args[3].(string)
		}
		r, err := sdkwireguard.AddPeer(iface, pubkey, allowedIPs, endpoint)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["wireguard.remove_peer"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("wireguard.remove_peer() requires 2 arguments (interface, public_key)")
		}
		iface, _ := args[0].(string)
		pubkey, _ := args[1].(string)
		r, err := sdkwireguard.RemovePeer(iface, pubkey)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["wireguard.genkey"] = func(args ...interface{}) (interface{}, error) {
		key, err := sdkwireguard.GenKey()
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"private_key": key}, nil
	}
	interp.builtins["wireguard.genpsk"] = func(args ...interface{}) (interface{}, error) {
		key, err := sdkwireguard.GenPSK()
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"preshared_key": key}, nil
	}
	interp.builtins["wireguard.pubkey"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("wireguard.pubkey() requires 1 argument (private_key)")
		}
		privkey, _ := args[0].(string)
		key, err := sdkwireguard.PubKey(privkey)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"public_key": key}, nil
	}

	// ── smartctl_notify.* ─────────────────────────────────────────────
	interp.builtins["smartctl_notify.check"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("smartctl_notify.check() requires 1 argument (device)")
		}
		device, _ := args[0].(string)
		r, err := sdksmartnotify.Check(device)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["smartctl_notify.list_devices"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdksmartnotify.ListDevices()
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"devices": r}, nil
	}
	interp.builtins["smartctl_notify.short_test"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("smartctl_notify.short_test() requires 1 argument (device)")
		}
		device, _ := args[0].(string)
		r, err := sdksmartnotify.ShortTest(device)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["smartctl_notify.long_test"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("smartctl_notify.long_test() requires 1 argument (device)")
		}
		device, _ := args[0].(string)
		r, err := sdksmartnotify.LongTest(device)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── dpkg_selections.* ─────────────────────────────────────────────
	interp.builtins["dpkg_selections.set"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("dpkg_selections.set() requires 2 arguments (name, state)")
		}
		name, _ := args[0].(string)
		state, _ := args[1].(string)
		r, err := sdkdpkgsel.SetSelection(name, state)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["dpkg_selections.get"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("dpkg_selections.get() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		r, err := sdkdpkgsel.GetSelection(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["dpkg_selections.list"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkdpkgsel.ListSelections()
		if err != nil {
			return nil, err
		}
		return r, nil
	}
	interp.builtins["dpkg_selections.hold"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("dpkg_selections.hold() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		r, err := sdkdpkgsel.Hold(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["dpkg_selections.unhold"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("dpkg_selections.unhold() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		r, err := sdkdpkgsel.Unhold(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── homebrew.* ────────────────────────────────────────────────────
	interp.builtins["homebrew.install"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("homebrew.install() requires at least 1 argument (name)")
		}
		name, _ := args[0].(string)
		cask := false
		if len(args) >= 2 {
			cask, _ = args[1].(bool)
		}
		r, err := sdkbrew.Install(name, cask)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["homebrew.remove"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("homebrew.remove() requires at least 1 argument (name)")
		}
		name, _ := args[0].(string)
		cask := false
		if len(args) >= 2 {
			cask, _ = args[1].(bool)
		}
		r, err := sdkbrew.Remove(name, cask)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["homebrew.upgrade"] = func(args ...interface{}) (interface{}, error) {
		name := ""
		if len(args) >= 1 {
			name, _ = args[0].(string)
		}
		r, err := sdkbrew.Upgrade(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["homebrew.update"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkbrew.Update()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["homebrew.info"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("homebrew.info() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		r, err := sdkbrew.Info(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["homebrew.list"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkbrew.List()
		if err != nil {
			return nil, err
		}
		return r, nil
	}
	interp.builtins["homebrew.list_casks"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkbrew.ListCasks()
		if err != nil {
			return nil, err
		}
		return r, nil
	}
	interp.builtins["homebrew.outdated"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkbrew.Outdated()
		if err != nil {
			return nil, err
		}
		return r, nil
	}
	interp.builtins["homebrew.clean"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkbrew.Clean()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["homebrew.tap"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("homebrew.tap() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		r, err := sdkbrew.Tap(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["homebrew.untap"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("homebrew.untap() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		r, err := sdkbrew.Untap(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["homebrew.list_taps"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkbrew.ListTaps()
		if err != nil {
			return nil, err
		}
		return r, nil
	}
	interp.builtins["homebrew.doctor"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkbrew.Doctor()
		if err != nil {
			return nil, err
		}
		return r, nil
	}

	// ── apt_repo.* ──────────────────────────────────────────────────────
	interp.builtins["apt_repo.list"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkaptrepo.List()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["apt_repo.exists"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("apt_repo.exists() requires 1 argument (uri)")
		}
		uri, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("apt_repo.exists(): uri must be string")
		}
		r, err := sdkaptrepo.Exists(uri)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["apt_repo.add"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 3 {
			return nil, fmt.Errorf("apt_repo.add() requires 3 arguments (uri, dist, components)")
		}
		uri, _ := args[0].(string)
		dist, _ := args[1].(string)
		comps, _ := args[2].(string)
		r, err := sdkaptrepo.Add(uri, dist, comps)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["apt_repo.remove"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("apt_repo.remove() requires 1 argument (uri)")
		}
		uri, _ := args[0].(string)
		r, err := sdkaptrepo.Remove(uri)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["apt_repo.update"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkaptrepo.Update()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── logrotate.* ─────────────────────────────────────────────────────
	interp.builtins["logrotate.list"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdklogrotate.List()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["logrotate.get"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("logrotate.get() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		r, err := sdklogrotate.Get(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["logrotate.set"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 4 {
			return nil, fmt.Errorf("logrotate.set() requires at least 4 arguments (name, pattern, frequency, rotate)")
		}
		name, _ := args[0].(string)
		pattern, _ := args[1].(string)
		freq, _ := args[2].(string)
		rotate := int(opsFloat(args, 3))
		compress := false
		if len(args) > 4 {
			compress = opsBool(args[4])
		}
		postRotate := ""
		if len(args) > 5 {
			postRotate, _ = args[5].(string)
		}
		r, err := sdklogrotate.Set(name, pattern, freq, rotate, compress, postRotate)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["logrotate.remove"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("logrotate.remove() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		r, err := sdklogrotate.Remove(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── lvg.* ─────────────────────────────────────────────────────────────
	interp.builtins["lvg.create"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("lvg.create() requires at least 2 arguments (name, pvs)")
		}
		name, _ := args[0].(string)
		pvsRaw, _ := args[1].([]interface{})
		pvs := make([]string, len(pvsRaw))
		for i, pv := range pvsRaw {
			pvs[i], _ = pv.(string)
		}
		r, err := sdklvg.Create(name, pvs)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["lvg.remove"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("lvg.remove() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		r, err := sdklvg.Remove(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["lvg.extend"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("lvg.extend() requires at least 2 arguments (name, pvs)")
		}
		name, _ := args[0].(string)
		pvsRaw, _ := args[1].([]interface{})
		pvs := make([]string, len(pvsRaw))
		for i, pv := range pvsRaw {
			pvs[i], _ = pv.(string)
		}
		r, err := sdklvg.Extend(name, pvs)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["lvg.reduce"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("lvg.reduce() requires at least 2 arguments (name, pvs)")
		}
		name, _ := args[0].(string)
		pvsRaw, _ := args[1].([]interface{})
		pvs := make([]string, len(pvsRaw))
		for i, pv := range pvsRaw {
			pvs[i], _ = pv.(string)
		}
		r, err := sdklvg.Reduce(name, pvs)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["lvg.activate"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("lvg.activate() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		r, err := sdklvg.Activate(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["lvg.deactivate"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("lvg.deactivate() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		r, err := sdklvg.Deactivate(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["lvg.list"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdklvg.List()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["lvg.get"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("lvg.get() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		r, err := sdklvg.Get(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── resolv.* ────────────────────────────────────────────────────────
	interp.builtins["resolv.get"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkresolv.Get()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["resolv.set"] = func(args ...interface{}) (interface{}, error) {
		var nameservers, search, options []string
		domain := ""
		if len(args) > 0 {
			if l, ok := args[0].([]interface{}); ok {
				for _, v := range l {
					if s, ok := v.(string); ok {
						nameservers = append(nameservers, s)
					}
				}
			}
		}
		if len(args) > 1 {
			if l, ok := args[1].([]interface{}); ok {
				for _, v := range l {
					if s, ok := v.(string); ok {
						search = append(search, s)
					}
				}
			}
		}
		if len(args) > 2 {
			if l, ok := args[2].([]interface{}); ok {
				for _, v := range l {
					if s, ok := v.(string); ok {
						options = append(options, s)
					}
				}
			}
		}
		if len(args) > 3 {
			domain, _ = args[3].(string)
		}
		r, err := sdkresolv.Set(nameservers, search, options, domain)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["resolv.add_nameserver"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("resolv.add_nameserver() requires 1 argument (nameserver)")
		}
		ns, _ := args[0].(string)
		r, err := sdkresolv.AddNameserver(ns)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["resolv.remove_nameserver"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("resolv.remove_nameserver() requires 1 argument (nameserver)")
		}
		ns, _ := args[0].(string)
		r, err := sdkresolv.RemoveNameserver(ns)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── yum_repo.* ──────────────────────────────────────────────────────
	interp.builtins["yum_repo.list"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkyumrepo.List()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["yum_repo.exists"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("yum_repo.exists() requires 1 argument (id)")
		}
		id, _ := args[0].(string)
		r, err := sdkyumrepo.Exists(id)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["yum_repo.add"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 3 {
			return nil, fmt.Errorf("yum_repo.add() requires at least 3 arguments (id, name, base_url)")
		}
		id, _ := args[0].(string)
		name, _ := args[1].(string)
		baseURL, _ := args[2].(string)
		gpgCheck := false
		if len(args) > 3 {
			gpgCheck = opsBool(args[3])
		}
		gpgKey := ""
		if len(args) > 4 {
			gpgKey, _ = args[4].(string)
		}
		r, err := sdkyumrepo.Add(id, name, baseURL, gpgCheck, gpgKey)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["yum_repo.remove"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("yum_repo.remove() requires 1 argument (id)")
		}
		id, _ := args[0].(string)
		r, err := sdkyumrepo.Remove(id)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ufw
	interp.builtins["ufw.status"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkufw.Status()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["ufw.list"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkufw.List()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["ufw.enable"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkufw.Enable()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["ufw.disable"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkufw.Disable()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["ufw.allow"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("ufw.allow() requires at least 1 argument (port)")
		}
		port, _ := args[0].(string)
		proto := getStringArgBridge(args, 1, "tcp")
		r, err := sdkufw.Allow(port, proto)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["ufw.deny"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("ufw.deny() requires at least 1 argument (port)")
		}
		port, _ := args[0].(string)
		proto := getStringArgBridge(args, 1, "tcp")
		r, err := sdkufw.Deny(port, proto)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["ufw.delete"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("ufw.delete() requires 1 argument (number)")
		}
		numFloat, ok := args[0].(float64)
		if !ok {
			return nil, fmt.Errorf("ufw.delete() number must be an integer")
		}
		num := int(numFloat)
		r, err := sdkufw.Delete(num)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["ufw.reset"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkufw.Reset()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["ufw.reload"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkufw.Reload()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ini_file
	interp.builtins["ini_file.sections"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("ini_file.sections() requires 1 argument (path)")
		}
		path, _ := args[0].(string)
		r, err := sdkinifile.Sections(path)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["ini_file.get"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 3 {
			return nil, fmt.Errorf("ini_file.get() requires 3 arguments (path, section, key)")
		}
		path, _ := args[0].(string)
		section, _ := args[1].(string)
		key, _ := args[2].(string)
		r, err := sdkinifile.Get(path, section, key)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["ini_file.set"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 4 {
			return nil, fmt.Errorf("ini_file.set() requires 4 arguments (path, section, key, value)")
		}
		path, _ := args[0].(string)
		section, _ := args[1].(string)
		key, _ := args[2].(string)
		value, _ := args[3].(string)
		r, err := sdkinifile.Set(path, section, key, value)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["ini_file.remove"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 3 {
			return nil, fmt.Errorf("ini_file.remove() requires 3 arguments (path, section, key)")
		}
		path, _ := args[0].(string)
		section, _ := args[1].(string)
		key, _ := args[2].(string)
		r, err := sdkinifile.Remove(path, section, key)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["ini_file.remove_section"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("ini_file.remove_section() requires 2 arguments (path, section)")
		}
		path, _ := args[0].(string)
		section, _ := args[1].(string)
		r, err := sdkinifile.RemoveSection(path, section)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// mount
	interp.builtins["mount.list"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkmount.List()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["mount.mount"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("mount.mount() requires at least 2 arguments (device, mountpoint)")
		}
		device, _ := args[0].(string)
		mountpoint, _ := args[1].(string)
		fstype := getStringArgBridge(args, 2, "")
		options := getStringArgBridge(args, 3, "")
		r, err := sdkmount.Mount(device, mountpoint, fstype, options)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["mount.umount"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("mount.umount() requires 1 argument (mountpoint)")
		}
		mountpoint, _ := args[0].(string)
		r, err := sdkmount.Unmount(mountpoint)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["mount.fstab"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkmount.Fstab()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["mount.add_fstab"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 3 {
			return nil, fmt.Errorf("mount.add_fstab() requires at least 3 arguments (device, mountpoint, fstype)")
		}
		device, _ := args[0].(string)
		mountpoint, _ := args[1].(string)
		fstype, _ := args[2].(string)
		options := getStringArgBridge(args, 3, "")
		r, err := sdkmount.AddFstab(device, mountpoint, fstype, options)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["mount.remove_fstab"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("mount.remove_fstab() requires 1 argument (target)")
		}
		target, _ := args[0].(string)
		r, err := sdkmount.RemoveFstab(target)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// hostname
	interp.builtins["hostname.get"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkhostname.Get()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["hostname.set"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("hostname.set() requires 1 argument (hostname)")
		}
		hostname, _ := args[0].(string)
		r, err := sdkhostname.Set(hostname)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["hostname.set_fqdn"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("hostname.set_fqdn() requires 1 argument (fqdn)")
		}
		fqdn, _ := args[0].(string)
		r, err := sdkhostname.SetFQDN(fqdn)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// timezone
	interp.builtins["timezone.get"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdktimezone.Get()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["timezone.set"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("timezone.set() requires 1 argument (timezone)")
		}
		timezone, _ := args[0].(string)
		r, err := sdktimezone.Set(timezone)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["timezone.list"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdktimezone.List()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── iptables ──────────────────────────────────────────────────────
	interp.builtins["iptables.list"] = func(args ...interface{}) (interface{}, error) {
		chain := ""
		if len(args) > 0 {
			chain, _ = args[0].(string)
		}
		r, err := sdkiptables.List(chain)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["iptables.flush"] = func(args ...interface{}) (interface{}, error) {
		table := ""
		if len(args) > 0 {
			table, _ = args[0].(string)
		}
		r, err := sdkiptables.Flush(table)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["iptables.add_rule"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("iptables.add_rule() requires 2 arguments (chain, rule_spec)")
		}
		chain, _ := args[0].(string)
		ruleSpec, _ := args[1].(string)
		r, err := sdkiptables.AddRule(chain, ruleSpec)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["iptables.delete_rule"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("iptables.delete_rule() requires 2 arguments (chain, number)")
		}
		chain, _ := args[0].(string)
		num := int(0)
		if n, ok := args[1].(float64); ok {
			num = int(n)
		}
		r, err := sdkiptables.DeleteRule(chain, num)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["iptables.save"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkiptables.Save()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["iptables.list_chains"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkiptables.ListChains()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── npm ───────────────────────────────────────────────────────────
	interp.builtins["npm.list"] = func(args ...interface{}) (interface{}, error) {
		global := false
		if len(args) > 0 {
			global, _ = args[0].(bool)
		}
		r, err := sdknpm.List(global)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["npm.install"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("npm.install() requires at least 1 argument (name)")
		}
		name, _ := args[0].(string)
		global := false
		if len(args) > 1 {
			global, _ = args[1].(bool)
		}
		r, err := sdknpm.Install(name, global)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["npm.uninstall"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("npm.uninstall() requires at least 1 argument (name)")
		}
		name, _ := args[0].(string)
		global := false
		if len(args) > 1 {
			global, _ = args[1].(bool)
		}
		r, err := sdknpm.Uninstall(name, global)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["npm.outdated"] = func(args ...interface{}) (interface{}, error) {
		global := false
		if len(args) > 0 {
			global, _ = args[0].(bool)
		}
		r, err := sdknpm.Outdated(global)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── mysql ─────────────────────────────────────────────────────────
	interp.builtins["mysql.databases"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkmysql.Databases()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["mysql.create_database"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("mysql.create_database() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		r, err := sdkmysql.CreateDatabase(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["mysql.drop_database"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("mysql.drop_database() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		r, err := sdkmysql.DropDatabase(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["mysql.users"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkmysql.Users()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["mysql.create_user"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 3 {
			return nil, fmt.Errorf("mysql.create_user() requires 3 arguments (user, host, password)")
		}
		user, _ := args[0].(string)
		host, _ := args[1].(string)
		password, _ := args[2].(string)
		r, err := sdkmysql.CreateUser(user, host, password)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["mysql.drop_user"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("mysql.drop_user() requires 2 arguments (user, host)")
		}
		user, _ := args[0].(string)
		host, _ := args[1].(string)
		r, err := sdkmysql.DropUser(user, host)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["mysql.grant"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 4 {
			return nil, fmt.Errorf("mysql.grant() requires 4 arguments (privileges, database, user, host)")
		}
		privileges, _ := args[0].(string)
		database, _ := args[1].(string)
		user, _ := args[2].(string)
		host, _ := args[3].(string)
		r, err := sdkmysql.Grant(privileges, database, user, host)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── nginx ─────────────────────────────────────────────────────────
	interp.builtins["nginx.config_test"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdknginx.ConfigTest()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["nginx.reload"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdknginx.Reload()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["nginx.sites_list"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdknginx.SitesList()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["nginx.site_enable"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("nginx.site_enable() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		r, err := sdknginx.SiteEnable(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["nginx.site_disable"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("nginx.site_disable() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		r, err := sdknginx.SiteDisable(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── modprobe ──────────────────────────────────────────────────────
	interp.builtins["modprobe.list"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkmodprobe.List()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["modprobe.load"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("modprobe.load() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		r, err := sdkmodprobe.Load(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["modprobe.unload"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("modprobe.unload() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		r, err := sdkmodprobe.Unload(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["modprobe.is_loaded"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("modprobe.is_loaded() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		r, err := sdkmodprobe.IsLoaded(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── alternatives ──────────────────────────────────────────────────
	interp.builtins["alternatives.list"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("alternatives.list() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		r, err := sdkalternatives.List(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["alternatives.display"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("alternatives.display() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		r, err := sdkalternatives.Display(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["alternatives.set"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("alternatives.set() requires 2 arguments (name, path)")
		}
		name, _ := args[0].(string)
		path, _ := args[1].(string)
		r, err := sdkalternatives.Set(name, path)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["alternatives.install"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 4 {
			return nil, fmt.Errorf("alternatives.install() requires 4 arguments (name, link, path, priority)")
		}
		name, _ := args[0].(string)
		link, _ := args[1].(string)
		path, _ := args[2].(string)
		priority := int(0)
		if p, ok := args[3].(float64); ok {
			priority = int(p)
		}
		r, err := sdkalternatives.Install(name, link, path, priority)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["alternatives.remove"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("alternatives.remove() requires 2 arguments (name, path)")
		}
		name, _ := args[0].(string)
		path, _ := args[1].(string)
		r, err := sdkalternatives.Remove(name, path)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── blockdev ──────────────────────────────────────────────────────
	interp.builtins["blockdev.list"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkblockdev.List()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["blockdev.info"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("blockdev.info() requires 1 argument (device)")
		}
		device, _ := args[0].(string)
		r, err := sdkblockdev.Info(device)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["blockdev.flush_buffers"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("blockdev.flush_buffers() requires 1 argument (device)")
		}
		device, _ := args[0].(string)
		r, err := sdkblockdev.FlushBuffers(device)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["blockdev.set_readahead"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("blockdev.set_readahead() requires 2 arguments (device, value)")
		}
		device, _ := args[0].(string)
		value := int(0)
		if v, ok := args[1].(float64); ok {
			value = int(v)
		}
		r, err := sdkblockdev.SetReadahead(device, value)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── at ────────────────────────────────────────────────────────────
	interp.builtins["at.list"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkat.List()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["at.schedule"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("at.schedule() requires 2 arguments (command, time_spec)")
		}
		command, _ := args[0].(string)
		timeSpec, _ := args[1].(string)
		r, err := sdkat.Schedule(command, timeSpec)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["at.remove"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("at.remove() requires 1 argument (job_id)")
		}
		jobID, _ := args[0].(string)
		r, err := sdkat.Remove(jobID)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── postgresql ─────────────────────────────────────────────────────
	interp.builtins["postgresql.databases"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkpostgresql.Databases()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["postgresql.create_database"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("postgresql.create_database() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		r, err := sdkpostgresql.CreateDatabase(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["postgresql.drop_database"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("postgresql.drop_database() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		r, err := sdkpostgresql.DropDatabase(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["postgresql.users"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkpostgresql.Users()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["postgresql.create_user"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("postgresql.create_user() requires 2 arguments (user, password)")
		}
		user, _ := args[0].(string)
		password, _ := args[1].(string)
		r, err := sdkpostgresql.CreateUser(user, password)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["postgresql.drop_user"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("postgresql.drop_user() requires 1 argument (user)")
		}
		user, _ := args[0].(string)
		r, err := sdkpostgresql.DropUser(user)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["postgresql.grant"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 3 {
			return nil, fmt.Errorf("postgresql.grant() requires 3 arguments (privileges, database, user)")
		}
		privileges, _ := args[0].(string)
		database, _ := args[1].(string)
		user, _ := args[2].(string)
		r, err := sdkpostgresql.Grant(privileges, database, user)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── apache2 ────────────────────────────────────────────────────────
	interp.builtins["apache2.config_test"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkapache2.ConfigTest()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["apache2.reload"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkapache2.Reload()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["apache2.sites_list"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkapache2.SitesList()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["apache2.site_enable"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("apache2.site_enable() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		r, err := sdkapache2.SiteEnable(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["apache2.site_disable"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("apache2.site_disable() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		r, err := sdkapache2.SiteDisable(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["apache2.modules_list"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkapache2.ModulesList()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["apache2.module_enable"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("apache2.module_enable() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		r, err := sdkapache2.ModuleEnable(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["apache2.module_disable"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("apache2.module_disable() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		r, err := sdkapache2.ModuleDisable(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── filesystem ─────────────────────────────────────────────────────
	interp.builtins["filesystem.mkfs"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("filesystem.mkfs() requires at least 2 arguments (device, fstype)")
		}
		device, _ := args[0].(string)
		fsType, _ := args[1].(string)
		label := getStringArgBridge(args, 2, "")
		r, err := sdkfilesystem.Mkfs(device, fsType, label)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["filesystem.resize_ext4"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("filesystem.resize_ext4() requires 1 argument (device)")
		}
		device, _ := args[0].(string)
		r, err := sdkfilesystem.ResizeExt4(device)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["filesystem.resize_xfs"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("filesystem.resize_xfs() requires 1 argument (mountpoint)")
		}
		mountpoint, _ := args[0].(string)
		r, err := sdkfilesystem.ResizeXFS(mountpoint)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["filesystem.check"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("filesystem.check() requires 1 argument (device)")
		}
		device, _ := args[0].(string)
		r, err := sdkfilesystem.Check(device)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── parted ─────────────────────────────────────────────────────────
	interp.builtins["parted.list"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("parted.list() requires 1 argument (device)")
		}
		device, _ := args[0].(string)
		r, err := sdkparted.List(device)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["parted.mklabel"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("parted.mklabel() requires at least 1 argument (device)")
		}
		device, _ := args[0].(string)
		labelType := getStringArgBridge(args, 1, "gpt")
		r, err := sdkparted.MkLabel(device, labelType)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["parted.mkpart"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 5 {
			return nil, fmt.Errorf("parted.mkpart() requires 5 arguments (device, part_type, fstype, start, end)")
		}
		device, _ := args[0].(string)
		partType, _ := args[1].(string)
		fsType, _ := args[2].(string)
		start, _ := args[3].(string)
		end, _ := args[4].(string)
		r, err := sdkparted.MkPart(device, partType, fsType, start, end)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["parted.rm"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("parted.rm() requires 2 arguments (device, number)")
		}
		device, _ := args[0].(string)
		number := int(opsFloat(args, 1))
		r, err := sdkparted.Rm(device, number)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── acl ────────────────────────────────────────────────────────────
	interp.builtins["acl.get"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("acl.get() requires 1 argument (path)")
		}
		path, _ := args[0].(string)
		r, err := sdkacl.Get(path)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["acl.set"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("acl.set() requires at least 2 arguments (path, entry)")
		}
		path, _ := args[0].(string)
		entry, _ := args[1].(string)
		recursive := false
		if len(args) > 2 {
			recursive, _ = args[2].(bool)
		}
		r, err := sdkacl.Set(path, entry, recursive)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["acl.remove"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("acl.remove() requires at least 2 arguments (path, entry)")
		}
		path, _ := args[0].(string)
		entry, _ := args[1].(string)
		recursive := false
		if len(args) > 2 {
			recursive, _ = args[2].(bool)
		}
		r, err := sdkacl.Remove(path, entry, recursive)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["acl.remove_all"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("acl.remove_all() requires at least 1 argument (path)")
		}
		path, _ := args[0].(string)
		recursive := false
		if len(args) > 1 {
			recursive, _ = args[1].(bool)
		}
		r, err := sdkacl.RemoveAll(path, recursive)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── wait_for ───────────────────────────────────────────────────────
	interp.builtins["wait_for.port"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("wait_for.port() requires at least 2 arguments (host, port)")
		}
		host, _ := args[0].(string)
		port := int(opsFloat(args, 1))
		timeoutMs := 30000
		if len(args) > 2 {
			timeoutMs = int(opsFloat(args, 2))
		}
		r, err := sdkwaitfor.Port(host, port, timeoutMs)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["wait_for.file"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("wait_for.file() requires at least 1 argument (path)")
		}
		path, _ := args[0].(string)
		timeoutMs := 30000
		if len(args) > 1 {
			timeoutMs = int(opsFloat(args, 1))
		}
		r, err := sdkwaitfor.File(path, timeoutMs)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["wait_for.url"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("wait_for.url() requires at least 1 argument (url)")
		}
		url, _ := args[0].(string)
		timeoutMs := 30000
		if len(args) > 1 {
			timeoutMs = int(opsFloat(args, 1))
		}
		r, err := sdkwaitfor.URL(url, timeoutMs)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── lvol ───────────────────────────────────────────────────────────
	interp.builtins["lvol.list"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdklvol.List()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["lvol.vg_list"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdklvol.VGList()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["lvol.create"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 3 {
			return nil, fmt.Errorf("lvol.create() requires 3 arguments (name, vg, size)")
		}
		name, _ := args[0].(string)
		vg, _ := args[1].(string)
		size, _ := args[2].(string)
		r, err := sdklvol.Create(name, vg, size)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["lvol.remove"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("lvol.remove() requires 2 arguments (name, vg)")
		}
		name, _ := args[0].(string)
		vg, _ := args[1].(string)
		r, err := sdklvol.Remove(name, vg)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["lvol.resize"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 3 {
			return nil, fmt.Errorf("lvol.resize() requires 3 arguments (name, vg, size)")
		}
		name, _ := args[0].(string)
		vg, _ := args[1].(string)
		size, _ := args[2].(string)
		r, err := sdklvol.Resize(name, vg, size)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── synchronize ────────────────────────────────────────────────────
	interp.builtins["synchronize.sync"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("synchronize.sync() requires at least 2 arguments (source, dest)")
		}
		source, _ := args[0].(string)
		dest, _ := args[1].(string)
		del := false
		if len(args) > 2 {
			del, _ = args[2].(bool)
		}
		compress := false
		if len(args) > 3 {
			compress, _ = args[3].(bool)
		}
		r, err := sdksync.Sync(source, dest, del, compress)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── fetch ──────────────────────────────────────────────────────────
	interp.builtins["fetch.file"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("fetch.file() requires 2 arguments (source, dest)")
		}
		source, _ := args[0].(string)
		dest, _ := args[1].(string)
		r, err := sdkfetch.File(source, dest)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["fetch.url"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("fetch.url() requires 2 arguments (url, dest)")
		}
		url, _ := args[0].(string)
		dest, _ := args[1].(string)
		r, err := sdkfetch.URL(url, dest)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── seboolean ──────────────────────────────────────────────────────
	interp.builtins["seboolean.list"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdksebool.List()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["seboolean.get"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("seboolean.get() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		r, err := sdksebool.Get(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["seboolean.set"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("seboolean.set() requires at least 2 arguments (name, state)")
		}
		name, _ := args[0].(string)
		state, _ := args[1].(bool)
		persistent := false
		if len(args) > 2 {
			persistent, _ = args[2].(bool)
		}
		r, err := sdksebool.Set(name, state, persistent)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── uri ─────────────────────────────────────────────────────────────
	interp.builtins["uri.do"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("uri.do() requires at least 1 argument (url)")
		}
		url, _ := args[0].(string)
		method := "GET"
		if len(args) > 1 {
			method, _ = args[1].(string)
		}
		headers := toStringMap(args, 2)
		body := ""
		if len(args) > 3 {
			body, _ = args[3].(string)
		}
		timeoutMs := 30000
		if len(args) > 4 {
			if f, ok := args[4].(float64); ok {
				timeoutMs = int(f)
			}
		}
		r, err := sdkuri.Do(url, method, headers, body, timeoutMs)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["uri.get"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("uri.get() requires 1 argument (url)")
		}
		url, _ := args[0].(string)
		r, err := sdkuri.Get(url)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["uri.post"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("uri.post() requires 2 arguments (url, body)")
		}
		url, _ := args[0].(string)
		r, err := sdkuri.Post(url, args[1])
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["uri.put"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("uri.put() requires 2 arguments (url, body)")
		}
		url, _ := args[0].(string)
		r, err := sdkuri.Put(url, args[1])
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["uri.delete"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("uri.delete() requires 1 argument (url)")
		}
		url, _ := args[0].(string)
		r, err := sdkuri.Delete(url)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["uri.download"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("uri.download() requires 2 arguments (url, dest)")
		}
		url, _ := args[0].(string)
		dest, _ := args[1].(string)
		r, err := sdkuri.Download(url, dest)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── lineinfile ──────────────────────────────────────────────────────
	interp.builtins["lineinfile.present"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("lineinfile.ensure() requires at least 2 arguments (path, line)")
		}
		path, _ := args[0].(string)
		line, _ := args[1].(string)
		re := ""
		if len(args) > 2 {
			re, _ = args[2].(string)
		}
		create := false
		if len(args) > 3 {
			create, _ = args[3].(bool)
		}
		r, err := sdklineinfile.Ensure(path, line, re, create)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["lineinfile.absent"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("lineinfile.absent() requires 2 arguments (path, regexp)")
		}
		path, _ := args[0].(string)
		re, _ := args[1].(string)
		r, err := sdklineinfile.Absent(path, re)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── replace ─────────────────────────────────────────────────────────
	interp.builtins["replace.replace"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 3 {
			return nil, fmt.Errorf("replace.replace() requires at least 3 arguments (path, pattern, replacement)")
		}
		path, _ := args[0].(string)
		pattern, _ := args[1].(string)
		replacement, _ := args[2].(string)
		regexpMode := false
		if len(args) > 3 {
			regexpMode, _ = args[3].(bool)
		}
		r, err := sdkreplace.Replace(path, pattern, replacement, regexpMode)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── xml ─────────────────────────────────────────────────────────────
	interp.builtins["xml.get_element"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("xml.get_element() requires 2 arguments (path, element)")
		}
		path, _ := args[0].(string)
		element, _ := args[1].(string)
		r, err := sdkxml.GetElement(path, element)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["xml.set_element"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 3 {
			return nil, fmt.Errorf("xml.set_element() requires 3 arguments (path, element, value)")
		}
		path, _ := args[0].(string)
		element, _ := args[1].(string)
		value, _ := args[2].(string)
		r, err := sdkxml.SetElement(path, element, value)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── systemd ─────────────────────────────────────────────────────────────
	interp.builtins["systemd.is_active"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("systemd.is_active() requires 1 argument (unit)")
		}
		unit, _ := args[0].(string)
		r, err := sdksystemd.IsActive(unit)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["systemd.is_enabled"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("systemd.is_enabled() requires 1 argument (unit)")
		}
		unit, _ := args[0].(string)
		r, err := sdksystemd.IsEnabled(unit)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["systemd.enable"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("systemd.enable() requires 1 argument (unit)")
		}
		unit, _ := args[0].(string)
		r, err := sdksystemd.Enable(unit)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["systemd.disable"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("systemd.disable() requires 1 argument (unit)")
		}
		unit, _ := args[0].(string)
		r, err := sdksystemd.Disable(unit)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["systemd.start"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("systemd.start() requires 1 argument (unit)")
		}
		unit, _ := args[0].(string)
		r, err := sdksystemd.Start(unit)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["systemd.stop"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("systemd.stop() requires 1 argument (unit)")
		}
		unit, _ := args[0].(string)
		r, err := sdksystemd.Stop(unit)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["systemd.restart"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("systemd.restart() requires 1 argument (unit)")
		}
		unit, _ := args[0].(string)
		r, err := sdksystemd.Restart(unit)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["systemd.reload"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("systemd.reload() requires 1 argument (unit)")
		}
		unit, _ := args[0].(string)
		r, err := sdksystemd.Reload(unit)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["systemd.daemon_reload"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdksystemd.DaemonReload()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["systemd.mask"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("systemd.mask() requires 1 argument (unit)")
		}
		unit, _ := args[0].(string)
		r, err := sdksystemd.Mask(unit)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["systemd.unmask"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("systemd.unmask() requires 1 argument (unit)")
		}
		unit, _ := args[0].(string)
		r, err := sdksystemd.Unmask(unit)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["systemd.show"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("systemd.show() requires 1 argument (unit)")
		}
		unit, _ := args[0].(string)
		r, err := sdksystemd.Show(unit)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["systemd.list"] = func(args ...interface{}) (interface{}, error) {
		unitType := ""
		if len(args) > 0 {
			unitType, _ = args[0].(string)
		}
		r, err := sdksystemd.List(unitType)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── patch ───────────────────────────────────────────────────────────────
	interp.builtins["patch.apply"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("patch.apply() requires at least 1 argument (patch_content)")
		}
		patchContent, _ := args[0].(string)
		reverse := false
		if len(args) > 1 {
			reverse, _ = args[1].(bool)
		}
		r, err := sdkpatch.Apply(patchContent, reverse)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["patch.dry_run"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("patch.dry_run() requires 1 argument (patch_content)")
		}
		patchContent, _ := args[0].(string)
		r, err := sdkpatch.DryRun(patchContent)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── xattr ───────────────────────────────────────────────────────────────
	interp.builtins["xattr.get"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("xattr.get() requires 2 arguments (path, name)")
		}
		path, _ := args[0].(string)
		name, _ := args[1].(string)
		r, err := sdkxattr.Get(path, name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["xattr.set"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 3 {
			return nil, fmt.Errorf("xattr.set() requires 3 arguments (path, name, value)")
		}
		path, _ := args[0].(string)
		name, _ := args[1].(string)
		value, _ := args[2].(string)
		r, err := sdkxattr.Set(path, name, value)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["xattr.remove"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("xattr.remove() requires 2 arguments (path, name)")
		}
		path, _ := args[0].(string)
		name, _ := args[1].(string)
		r, err := sdkxattr.Remove(path, name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["xattr.list"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("xattr.list() requires 1 argument (path)")
		}
		path, _ := args[0].(string)
		r, err := sdkxattr.List(path)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── firewalld_zone ──────────────────────────────────────────────────────
	interp.builtins["firewalld_zone.get_default"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkfirewalldzone.GetDefaultZone()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["firewalld_zone.set_default"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("firewalld_zone.set_default() requires 1 argument (zone)")
		}
		zone, _ := args[0].(string)
		r, err := sdkfirewalldzone.SetDefaultZone(zone)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["firewalld_zone.add_zone"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("firewalld_zone.add_zone() requires 1 argument (zone)")
		}
		zone, _ := args[0].(string)
		r, err := sdkfirewalldzone.AddZone(zone)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["firewalld_zone.remove_zone"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("firewalld_zone.remove_zone() requires 1 argument (zone)")
		}
		zone, _ := args[0].(string)
		r, err := sdkfirewalldzone.RemoveZone(zone)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["firewalld_zone.add_service"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("firewalld_zone.add_service() requires 2 arguments (zone, service)")
		}
		zone, _ := args[0].(string)
		svc, _ := args[1].(string)
		r, err := sdkfirewalldzone.AddService(zone, svc)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["firewalld_zone.remove_service"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("firewalld_zone.remove_service() requires 2 arguments (zone, service)")
		}
		zone, _ := args[0].(string)
		svc, _ := args[1].(string)
		r, err := sdkfirewalldzone.RemoveService(zone, svc)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["firewalld_zone.add_port"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("firewalld_zone.add_port() requires 2 arguments (zone, port_protocol)")
		}
		zone, _ := args[0].(string)
		pp, _ := args[1].(string)
		r, err := sdkfirewalldzone.AddPort(zone, pp)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["firewalld_zone.remove_port"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("firewalld_zone.remove_port() requires 2 arguments (zone, port_protocol)")
		}
		zone, _ := args[0].(string)
		pp, _ := args[1].(string)
		r, err := sdkfirewalldzone.RemovePort(zone, pp)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["firewalld_zone.add_rich_rule"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("firewalld_zone.add_rich_rule() requires 2 arguments (zone, rule)")
		}
		zone, _ := args[0].(string)
		rule, _ := args[1].(string)
		r, err := sdkfirewalldzone.AddRichRule(zone, rule)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["firewalld_zone.remove_rich_rule"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("firewalld_zone.remove_rich_rule() requires 2 arguments (zone, rule)")
		}
		zone, _ := args[0].(string)
		rule, _ := args[1].(string)
		r, err := sdkfirewalldzone.RemoveRichRule(zone, rule)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["firewalld_zone.info"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("firewalld_zone.info() requires 1 argument (zone)")
		}
		zone, _ := args[0].(string)
		r, err := sdkfirewalldzone.Info(zone)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["firewalld_zone.list_zones"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkfirewalldzone.ListZones()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── get_url ─────────────────────────────────────────────────────────────
	interp.builtins["get_url.download"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("get_url.download() requires at least 2 arguments (url, dest)")
		}
		url, _ := args[0].(string)
		dest, _ := args[1].(string)
		checksum := ""
		if len(args) > 2 {
			checksum, _ = args[2].(string)
		}
		force := false
		if len(args) > 3 {
			force, _ = args[3].(bool)
		}
		r, err := sdkgeturl.Download(url, dest, checksum, force)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── sys utilities ───────────────────────────────────────────────────────
	interp.builtins["sys.uuid"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdksys.UUID()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["sys.random_password"] = func(args ...interface{}) (interface{}, error) {
		length := 16
		if len(args) > 0 {
			switch v := args[0].(type) {
			case int:
				length = v
			case float64:
				length = int(v)
			}
		}
		useSpecial := true
		if len(args) > 1 {
			useSpecial, _ = args[1].(bool)
		}
		useNumbers := true
		if len(args) > 2 {
			useNumbers, _ = args[2].(bool)
		}
		useUppercase := true
		if len(args) > 3 {
			useUppercase, _ = args[3].(bool)
		}
		r, err := sdksys.RandomPassword(length, useSpecial, useNumbers, useUppercase)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── sys mac_address ──────────────────────────────────────────────────────
	interp.builtins["sys.mac_address"] = func(args ...interface{}) (interface{}, error) {
		iface := ""
		if len(args) > 0 {
			iface, _ = args[0].(string)
		}
		r, err := sdksys.MACAddress(iface)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["sys.mac_addresses"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdksys.MACAddresses()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["sys.dmidecode"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdksys.Dmidecode()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["sys.lspci"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdksys.LsPci()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["sys.lsblk"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdksys.LsBlk()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["sys.lsusb"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdksys.LsUsb()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["sys.ip_route"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdksys.IpRoute()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["sys.ethtool"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("sys.ethtool() requires 1 argument (iface)")
		}
		iface, _ := args[0].(string)
		r, err := sdksys.Ethtool(iface)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── modprobe.set_boot ──────────────────────────────────────────────────────
	interp.builtins["modprobe.set_boot"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("modprobe.set_boot() requires at least 1 argument (name)")
		}
		name, _ := args[0].(string)
		present := true
		if len(args) > 1 {
			present, _ = args[1].(bool)
		}
		r, err := sdkmodprobe.SetBoot(name, present)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── seport ─────────────────────────────────────────────────────────────────
	interp.builtins["seport.add"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 3 {
			return nil, fmt.Errorf("seport.add() requires 3 arguments (seport_type, protocol, port)")
		}
		seportType, _ := args[0].(string)
		proto, _ := args[1].(string)
		port, _ := args[2].(string)
		r, err := seport.Add(seportType, proto, port)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["seport.remove"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("seport.remove() requires 2 arguments (protocol, port)")
		}
		proto, _ := args[0].(string)
		port, _ := args[1].(string)
		r, err := seport.Remove(proto, port)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["seport.list"] = func(args ...interface{}) (interface{}, error) {
		r, err := seport.List()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["seport.get"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("seport.get() requires 2 arguments (protocol, port)")
		}
		proto, _ := args[0].(string)
		port, _ := args[1].(string)
		r, err := seport.Get(proto, port)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── sefcontext ─────────────────────────────────────────────────────────────
	interp.builtins["sefcontext.add"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("sefcontext.add() requires 2 arguments (filespec, se_type)")
		}
		filespec, _ := args[0].(string)
		seType, _ := args[1].(string)
		r, err := sefcontext.Add(filespec, seType)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["sefcontext.modify"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("sefcontext.modify() requires 2 arguments (filespec, se_type)")
		}
		filespec, _ := args[0].(string)
		seType, _ := args[1].(string)
		r, err := sefcontext.Modify(filespec, seType)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["sefcontext.remove"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("sefcontext.remove() requires 1 argument (filespec)")
		}
		filespec, _ := args[0].(string)
		r, err := sefcontext.Remove(filespec)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["sefcontext.list"] = func(args ...interface{}) (interface{}, error) {
		r, err := sefcontext.List()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["sefcontext.apply"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("sefcontext.apply() requires at least 1 argument (filespec)")
		}
		filespec, _ := args[0].(string)
		recursive := false
		if len(args) > 1 {
			recursive, _ = args[1].(bool)
		}
		r, err := sefcontext.Apply(filespec, recursive)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── composer ──────────────────────────────────────────────────────────
	interp.builtins["composer.install"] = func(args ...interface{}) (interface{}, error) {
		dir := ""
		noDev := false
		if len(args) > 0 {
			if s, ok := args[0].(string); ok {
				dir = s
			}
		}
		if len(args) > 1 {
			if b, ok := args[1].(bool); ok {
				noDev = b
			}
		}
		return sdkcomposer.Install(dir, noDev), nil
	}
	interp.builtins["composer.update"] = func(args ...interface{}) (interface{}, error) {
		dir := ""
		noDev := false
		if len(args) > 0 {
			if s, ok := args[0].(string); ok {
				dir = s
			}
		}
		if len(args) > 1 {
			if b, ok := args[1].(bool); ok {
				noDev = b
			}
		}
		return sdkcomposer.Update(dir, noDev), nil
	}
	interp.builtins["composer.require"] = func(args ...interface{}) (interface{}, error) {
		dir, pkg, ver := "", "", ""
		if len(args) > 0 {
			if s, ok := args[0].(string); ok {
				dir = s
			}
		}
		if len(args) > 1 {
			if s, ok := args[1].(string); ok {
				pkg = s
			}
		}
		if len(args) > 2 {
			if s, ok := args[2].(string); ok {
				ver = s
			}
		}
		return sdkcomposer.Require(dir, pkg, ver), nil
	}
	interp.builtins["composer.remove"] = func(args ...interface{}) (interface{}, error) {
		dir, pkg := "", ""
		if len(args) > 0 {
			if s, ok := args[0].(string); ok {
				dir = s
			}
		}
		if len(args) > 1 {
			if s, ok := args[1].(string); ok {
				pkg = s
			}
		}
		return sdkcomposer.Remove(dir, pkg), nil
	}
	interp.builtins["composer.create_project"] = func(args ...interface{}) (interface{}, error) {
		dir, pkg, ver := "", "", ""
		if len(args) > 0 {
			if s, ok := args[0].(string); ok {
				dir = s
			}
		}
		if len(args) > 1 {
			if s, ok := args[1].(string); ok {
				pkg = s
			}
		}
		if len(args) > 2 {
			if s, ok := args[2].(string); ok {
				ver = s
			}
		}
		return sdkcomposer.CreateProject(dir, pkg, ver), nil
	}
	interp.builtins["composer.global_install"] = func(args ...interface{}) (interface{}, error) {
		pkg, ver := "", ""
		if len(args) > 0 {
			if s, ok := args[0].(string); ok {
				pkg = s
			}
		}
		if len(args) > 1 {
			if s, ok := args[1].(string); ok {
				ver = s
			}
		}
		return sdkcomposer.GlobalInstall(pkg, ver), nil
	}
	interp.builtins["composer.version"] = func(args ...interface{}) (interface{}, error) {
		return sdkcomposer.Version(), nil
	}

	// ── cargo ─────────────────────────────────────────────────────────────
	interp.builtins["cargo.install"] = func(args ...interface{}) (interface{}, error) {
		pkg, ver := "", ""
		force := false
		if len(args) > 0 {
			if s, ok := args[0].(string); ok {
				pkg = s
			}
		}
		if len(args) > 1 {
			if s, ok := args[1].(string); ok {
				ver = s
			}
		}
		if len(args) > 2 {
			if b, ok := args[2].(bool); ok {
				force = b
			}
		}
		return sdkcargo.Install(pkg, ver, force), nil
	}
	interp.builtins["cargo.uninstall"] = func(args ...interface{}) (interface{}, error) {
		pkg := ""
		if len(args) > 0 {
			if s, ok := args[0].(string); ok {
				pkg = s
			}
		}
		return sdkcargo.Uninstall(pkg), nil
	}
	interp.builtins["cargo.update"] = func(args ...interface{}) (interface{}, error) {
		pkg := ""
		if len(args) > 0 {
			if s, ok := args[0].(string); ok {
				pkg = s
			}
		}
		return sdkcargo.Update(pkg), nil
	}
	interp.builtins["cargo.list"] = func(args ...interface{}) (interface{}, error) {
		return sdkcargo.List()
	}
	interp.builtins["cargo.build"] = func(args ...interface{}) (interface{}, error) {
		dir := ""
		release := false
		if len(args) > 0 {
			if s, ok := args[0].(string); ok {
				dir = s
			}
		}
		if len(args) > 1 {
			if b, ok := args[1].(bool); ok {
				release = b
			}
		}
		return sdkcargo.Build(dir, release), nil
	}
	interp.builtins["cargo.test"] = func(args ...interface{}) (interface{}, error) {
		dir := ""
		if len(args) > 0 {
			if s, ok := args[0].(string); ok {
				dir = s
			}
		}
		return sdkcargo.Test(dir), nil
	}
	interp.builtins["cargo.version"] = func(args ...interface{}) (interface{}, error) {
		return sdkcargo.Version(), nil
	}

	// ── rpmkey ────────────────────────────────────────────────────────────
	interp.builtins["rpmkey.import"] = func(args ...interface{}) (interface{}, error) {
		keyPath := ""
		if len(args) > 0 {
			if s, ok := args[0].(string); ok {
				keyPath = s
			}
		}
		return sdkrpmkey.Import(keyPath), nil
	}
	interp.builtins["rpmkey.list"] = func(args ...interface{}) (interface{}, error) {
		return sdkrpmkey.List(), nil
	}
	interp.builtins["rpmkey.remove"] = func(args ...interface{}) (interface{}, error) {
		keyID := ""
		if len(args) > 0 {
			if s, ok := args[0].(string); ok {
				keyID = s
			}
		}
		return sdkrpmkey.Remove(keyID), nil
	}

	// ── aptkey ────────────────────────────────────────────────────────────
	interp.builtins["aptkey.add"] = func(args ...interface{}) (interface{}, error) {
		url, keyring := "", ""
		if len(args) > 0 {
			if s, ok := args[0].(string); ok {
				url = s
			}
		}
		if len(args) > 1 {
			if s, ok := args[1].(string); ok {
				keyring = s
			}
		}
		return sdkaptkey.Add(url, keyring), nil
	}
	interp.builtins["aptkey.add_from_key"] = func(args ...interface{}) (interface{}, error) {
		path, keyring := "", ""
		if len(args) > 0 {
			if s, ok := args[0].(string); ok {
				path = s
			}
		}
		if len(args) > 1 {
			if s, ok := args[1].(string); ok {
				keyring = s
			}
		}
		return sdkaptkey.AddFromKey(path, keyring), nil
	}
	interp.builtins["aptkey.remove"] = func(args ...interface{}) (interface{}, error) {
		keyID, keyring := "", ""
		if len(args) > 0 {
			if s, ok := args[0].(string); ok {
				keyID = s
			}
		}
		if len(args) > 1 {
			if s, ok := args[1].(string); ok {
				keyring = s
			}
		}
		return sdkaptkey.Remove(keyID, keyring), nil
	}
	interp.builtins["aptkey.list"] = func(args ...interface{}) (interface{}, error) {
		return sdkaptkey.List(), nil
	}

	// ── dmidecode ─────────────────────────────────────────────────────────
	interp.builtins["dmidecode.system"] = func(args ...interface{}) (interface{}, error) {
		return sdkdmidecode.System(), nil
	}
	interp.builtins["dmidecode.bios"] = func(args ...interface{}) (interface{}, error) {
		return sdkdmidecode.BIOS(), nil
	}
	interp.builtins["dmidecode.chassis"] = func(args ...interface{}) (interface{}, error) {
		return sdkdmidecode.Chassis(), nil
	}
	interp.builtins["dmidecode.processor"] = func(args ...interface{}) (interface{}, error) {
		return sdkdmidecode.Processor(), nil
	}
	interp.builtins["dmidecode.keyword"] = func(args ...interface{}) (interface{}, error) {
		keyword := ""
		if len(args) > 0 {
			if s, ok := args[0].(string); ok {
				keyword = s
			}
		}
		v, err := sdkdmidecode.Keyword(keyword)
		return map[string]interface{}{"value": v}, err
	}

	// ── tuned ─────────────────────────────────────────────────────────────
	interp.builtins["tuned.set"] = func(args ...interface{}) (interface{}, error) {
		profile := ""
		if len(args) > 0 {
			if s, ok := args[0].(string); ok {
				profile = s
			}
		}
		return sdktuned.Set(profile), nil
	}
	interp.builtins["tuned.status"] = func(args ...interface{}) (interface{}, error) {
		return sdktuned.Status(), nil
	}
	interp.builtins["tuned.list"] = func(args ...interface{}) (interface{}, error) {
		return sdktuned.List(), nil
	}
	interp.builtins["tuned.off"] = func(args ...interface{}) (interface{}, error) {
		return sdktuned.Off(), nil
	}
	interp.builtins["tuned.profile"] = func(args ...interface{}) (interface{}, error) {
		p, err := sdktuned.Profile()
		return map[string]interface{}{"profile": p}, err
	}
	interp.builtins["tuned.verify"] = func(args ...interface{}) (interface{}, error) {
		return sdktuned.Verify(), nil
	}

	// ── supervisor ────────────────────────────────────────────────────────
	interp.builtins["supervisor.start"] = func(args ...interface{}) (interface{}, error) {
		name := ""
		if len(args) > 0 {
			if s, ok := args[0].(string); ok {
				name = s
			}
		}
		return sdksupervisor.Start(name), nil
	}
	interp.builtins["supervisor.stop"] = func(args ...interface{}) (interface{}, error) {
		name := ""
		if len(args) > 0 {
			if s, ok := args[0].(string); ok {
				name = s
			}
		}
		return sdksupervisor.Stop(name), nil
	}
	interp.builtins["supervisor.restart"] = func(args ...interface{}) (interface{}, error) {
		name := ""
		if len(args) > 0 {
			if s, ok := args[0].(string); ok {
				name = s
			}
		}
		return sdksupervisor.Restart(name), nil
	}
	interp.builtins["supervisor.reload"] = func(args ...interface{}) (interface{}, error) {
		return sdksupervisor.Reload(), nil
	}
	interp.builtins["supervisor.status"] = func(args ...interface{}) (interface{}, error) {
		return sdksupervisor.Status(), nil
	}
	interp.builtins["supervisor.clear_log"] = func(args ...interface{}) (interface{}, error) {
		name := ""
		if len(args) > 0 {
			if s, ok := args[0].(string); ok {
				name = s
			}
		}
		return sdksupervisor.ClearLog(name), nil
	}
	interp.builtins["supervisor.reread"] = func(args ...interface{}) (interface{}, error) {
		return sdksupervisor.Reread(), nil
	}
	interp.builtins["supervisor.update"] = func(args ...interface{}) (interface{}, error) {
		name := ""
		if len(args) > 0 {
			if s, ok := args[0].(string); ok {
				name = s
			}
		}
		return sdksupervisor.Update(name), nil
	}

	// ── pip.freeze / pip.install_requirements ──────────────────────────
	interp.builtins["pip.freeze"] = func(args ...interface{}) (interface{}, error) {
		return sdkpip.Freeze()
	}
	interp.builtins["pip.install_requirements"] = func(args ...interface{}) (interface{}, error) {
		req := ""
		if len(args) > 0 {
			if s, ok := args[0].(string); ok {
				req = s
			}
		}
		return sdkpip.InstallRequirements(req)
	}

	// ── flatpak.add_remote ─────────────────────────────────────────────
	interp.builtins["flatpak.add_remote"] = func(args ...interface{}) (interface{}, error) {
		name, url := "", ""
		if len(args) > 0 {
			if s, ok := args[0].(string); ok {
				name = s
			}
		}
		if len(args) > 1 {
			if s, ok := args[1].(string); ok {
				url = s
			}
		}
		return sdkflatpak.AddRemote(name, url)
	}

	// ── yarn ───────────────────────────────────────────────────────────
	interp.builtins["yarn.install"] = func(args ...interface{}) (interface{}, error) {
		name, ver := "", ""
		global := false
		if len(args) > 0 {
			if s, ok := args[0].(string); ok {
				name = s
			}
		}
		if len(args) > 1 {
			if s, ok := args[1].(string); ok {
				ver = s
			}
		}
		if len(args) > 2 {
			global, _ = args[2].(bool)
		}
		return sdkyarn.Install(name, ver, global), nil
	}
	interp.builtins["yarn.remove"] = func(args ...interface{}) (interface{}, error) {
		name := ""
		global := false
		if len(args) > 0 {
			if s, ok := args[0].(string); ok {
				name = s
			}
		}
		if len(args) > 1 {
			global, _ = args[1].(bool)
		}
		return sdkyarn.Remove(name, global), nil
	}
	interp.builtins["yarn.global"] = func(args ...interface{}) (interface{}, error) {
		dir := ""
		if len(args) > 0 {
			if s, ok := args[0].(string); ok {
				dir = s
			}
		}
		return sdkyarn.Global(dir), nil
	}
	interp.builtins["yarn.list"] = func(args ...interface{}) (interface{}, error) {
		global := false
		if len(args) > 0 {
			global, _ = args[0].(bool)
		}
		return sdkyarn.List(global), nil
	}

	// ── htpasswd ───────────────────────────────────────────────────────
	interp.builtins["htpasswd.set"] = func(args ...interface{}) (interface{}, error) {
		path, username, password := "", "", ""
		create := false
		if len(args) > 0 {
			if s, ok := args[0].(string); ok {
				path = s
			}
		}
		if len(args) > 1 {
			if s, ok := args[1].(string); ok {
				username = s
			}
		}
		if len(args) > 2 {
			if s, ok := args[2].(string); ok {
				password = s
			}
		}
		if len(args) > 3 {
			create, _ = args[3].(bool)
		}
		return sdkhtpasswd.Set(path, username, password, create), nil
	}
	interp.builtins["htpasswd.remove"] = func(args ...interface{}) (interface{}, error) {
		path, username := "", ""
		if len(args) > 0 {
			if s, ok := args[0].(string); ok {
				path = s
			}
		}
		if len(args) > 1 {
			if s, ok := args[1].(string); ok {
				username = s
			}
		}
		return sdkhtpasswd.Remove(path, username), nil
	}
	interp.builtins["htpasswd.info"] = func(args ...interface{}) (interface{}, error) {
		path := ""
		if len(args) > 0 {
			if s, ok := args[0].(string); ok {
				path = s
			}
		}
		return sdkhtpasswd.Info(path), nil
	}
	interp.builtins["htpasswd.hash_sha1"] = func(args ...interface{}) (interface{}, error) {
		password := ""
		if len(args) > 0 {
			if s, ok := args[0].(string); ok {
				password = s
			}
		}
		return sdkhtpasswd.HashSHA1(password), nil
	}

	// ── sudoers ────────────────────────────────────────────────────────
	interp.builtins["sudoers.set"] = func(args ...interface{}) (interface{}, error) {
		name, user, commands := "", "", ""
		nopasswd := false
		dir := ""
		if len(args) > 0 {
			if s, ok := args[0].(string); ok {
				name = s
			}
		}
		if len(args) > 1 {
			if s, ok := args[1].(string); ok {
				user = s
			}
		}
		if len(args) > 2 {
			if s, ok := args[2].(string); ok {
				commands = s
			}
		}
		if len(args) > 3 {
			nopasswd, _ = args[3].(bool)
		}
		if len(args) > 4 {
			if s, ok := args[4].(string); ok {
				dir = s
			}
		}
		return sdksudoers.Set(name, user, commands, nopasswd, dir), nil
	}
	interp.builtins["sudoers.remove"] = func(args ...interface{}) (interface{}, error) {
		name, dir := "", ""
		if len(args) > 0 {
			if s, ok := args[0].(string); ok {
				name = s
			}
		}
		if len(args) > 1 {
			if s, ok := args[1].(string); ok {
				dir = s
			}
		}
		return sdksudoers.Remove(name, dir), nil
	}
	interp.builtins["sudoers.info"] = func(args ...interface{}) (interface{}, error) {
		name, dir := "", ""
		if len(args) > 0 {
			if s, ok := args[0].(string); ok {
				name = s
			}
		}
		if len(args) > 1 {
			if s, ok := args[1].(string); ok {
				dir = s
			}
		}
		return sdksudoers.Info(name, dir), nil
	}

	// ── monit ──────────────────────────────────────────────────────────
	interp.builtins["monit.start"] = func(args ...interface{}) (interface{}, error) {
		svc := ""
		if len(args) > 0 {
			if s, ok := args[0].(string); ok {
				svc = s
			}
		}
		return sdkmonit.Start(svc), nil
	}
	interp.builtins["monit.stop"] = func(args ...interface{}) (interface{}, error) {
		svc := ""
		if len(args) > 0 {
			if s, ok := args[0].(string); ok {
				svc = s
			}
		}
		return sdkmonit.Stop(svc), nil
	}
	interp.builtins["monit.monitor"] = func(args ...interface{}) (interface{}, error) {
		svc := ""
		if len(args) > 0 {
			if s, ok := args[0].(string); ok {
				svc = s
			}
		}
		return sdkmonit.Monitor(svc), nil
	}
	interp.builtins["monit.unmonitor"] = func(args ...interface{}) (interface{}, error) {
		svc := ""
		if len(args) > 0 {
			if s, ok := args[0].(string); ok {
				svc = s
			}
		}
		return sdkmonit.Unmonitor(svc), nil
	}
	interp.builtins["monit.restart"] = func(args ...interface{}) (interface{}, error) {
		svc := ""
		if len(args) > 0 {
			if s, ok := args[0].(string); ok {
				svc = s
			}
		}
		return sdkmonit.Restart(svc), nil
	}
	interp.builtins["monit.status"] = func(args ...interface{}) (interface{}, error) {
		return sdkmonit.Status(), nil
	}
	interp.builtins["monit.reload"] = func(args ...interface{}) (interface{}, error) {
		return sdkmonit.Reload(), nil
	}

	// ── smartctl ──────────────────────────────────────────────────────────
	interp.builtins["smartctl.device"] = func(args ...interface{}) (interface{}, error) {
		device := ""
		if len(args) > 0 {
			if s, ok := args[0].(string); ok {
				device = s
			}
		}
		return sdksmartctl.Device(device), nil
	}
	interp.builtins["smartctl.health"] = func(args ...interface{}) (interface{}, error) {
		device := ""
		if len(args) > 0 {
			if s, ok := args[0].(string); ok {
				device = s
			}
		}
		return sdksmartctl.Health(device), nil
	}
	interp.builtins["smartctl.attributes"] = func(args ...interface{}) (interface{}, error) {
		device := ""
		if len(args) > 0 {
			if s, ok := args[0].(string); ok {
				device = s
			}
		}
		return sdksmartctl.Attributes(device), nil
	}
	interp.builtins["smartctl.list"] = func(args ...interface{}) (interface{}, error) {
		return sdksmartctl.List(), nil
	}
	interp.builtins["smartctl.json"] = func(args ...interface{}) (interface{}, error) {
		device := ""
		if len(args) > 0 {
			if s, ok := args[0].(string); ok {
				device = s
			}
		}
		v, err := sdksmartctl.JSON(device)
		return map[string]interface{}{"output": v}, err
	}

	// ── virsh ─────────────────────────────────────────────────────────────
	interp.builtins["virsh.start"] = func(args ...interface{}) (interface{}, error) {
		domain := ""
		if len(args) > 0 {
			if s, ok := args[0].(string); ok {
				domain = s
			}
		}
		return sdkvirsh.Start(domain), nil
	}
	interp.builtins["virsh.stop"] = func(args ...interface{}) (interface{}, error) {
		domain := ""
		if len(args) > 0 {
			if s, ok := args[0].(string); ok {
				domain = s
			}
		}
		return sdkvirsh.Stop(domain), nil
	}
	interp.builtins["virsh.reboot"] = func(args ...interface{}) (interface{}, error) {
		domain := ""
		if len(args) > 0 {
			if s, ok := args[0].(string); ok {
				domain = s
			}
		}
		return sdkvirsh.Reboot(domain), nil
	}
	interp.builtins["virsh.shutdown"] = func(args ...interface{}) (interface{}, error) {
		domain := ""
		if len(args) > 0 {
			if s, ok := args[0].(string); ok {
				domain = s
			}
		}
		return sdkvirsh.Shutdown(domain), nil
	}
	interp.builtins["virsh.suspend"] = func(args ...interface{}) (interface{}, error) {
		domain := ""
		if len(args) > 0 {
			if s, ok := args[0].(string); ok {
				domain = s
			}
		}
		return sdkvirsh.Suspend(domain), nil
	}
	interp.builtins["virsh.resume"] = func(args ...interface{}) (interface{}, error) {
		domain := ""
		if len(args) > 0 {
			if s, ok := args[0].(string); ok {
				domain = s
			}
		}
		return sdkvirsh.Resume(domain), nil
	}
	interp.builtins["virsh.list"] = func(args ...interface{}) (interface{}, error) {
		return sdkvirsh.List(), nil
	}
	interp.builtins["virsh.info"] = func(args ...interface{}) (interface{}, error) {
		domain := ""
		if len(args) > 0 {
			if s, ok := args[0].(string); ok {
				domain = s
			}
		}
		return sdkvirsh.Info(domain)
	}
	interp.builtins["virsh.version"] = func(args ...interface{}) (interface{}, error) {
		v, err := sdkvirsh.Version()
		return map[string]interface{}{"version": v}, err
	}

	// ── ethtool ───────────────────────────────────────────────────────────
	interp.builtins["ethtool.show"] = func(args ...interface{}) (interface{}, error) {
		iface := ""
		if len(args) > 0 {
			if s, ok := args[0].(string); ok {
				iface = s
			}
		}
		return sdkethtool.Show(iface), nil
	}
	interp.builtins["ethtool.set_speed"] = func(args ...interface{}) (interface{}, error) {
		iface, speed := "", ""
		if len(args) > 0 {
			if s, ok := args[0].(string); ok {
				iface = s
			}
		}
		if len(args) > 1 {
			if s, ok := args[1].(string); ok {
				speed = s
			}
		}
		return sdkethtool.SetSpeed(iface, speed), nil
	}
	interp.builtins["ethtool.set_duplex"] = func(args ...interface{}) (interface{}, error) {
		iface, duplex := "", ""
		if len(args) > 0 {
			if s, ok := args[0].(string); ok {
				iface = s
			}
		}
		if len(args) > 1 {
			if s, ok := args[1].(string); ok {
				duplex = s
			}
		}
		return sdkethtool.SetDuplex(iface, duplex), nil
	}
	interp.builtins["ethtool.set_autoneg"] = func(args ...interface{}) (interface{}, error) {
		iface, autoneg := "", ""
		if len(args) > 0 {
			if s, ok := args[0].(string); ok {
				iface = s
			}
		}
		if len(args) > 1 {
			if s, ok := args[1].(string); ok {
				autoneg = s
			}
		}
		return sdkethtool.SetAutoneg(iface, autoneg), nil
	}
	interp.builtins["ethtool.set_pause"] = func(args ...interface{}) (interface{}, error) {
		iface, rx, tx := "", "", ""
		if len(args) > 0 {
			if s, ok := args[0].(string); ok {
				iface = s
			}
		}
		if len(args) > 1 {
			if s, ok := args[1].(string); ok {
				rx = s
			}
		}
		if len(args) > 2 {
			if s, ok := args[2].(string); ok {
				tx = s
			}
		}
		return sdkethtool.SetPause(iface, rx, tx), nil
	}
	interp.builtins["ethtool.set_offload"] = func(args ...interface{}) (interface{}, error) {
		iface, feature, value := "", "", ""
		if len(args) > 0 {
			if s, ok := args[0].(string); ok {
				iface = s
			}
		}
		if len(args) > 1 {
			if s, ok := args[1].(string); ok {
				feature = s
			}
		}
		if len(args) > 2 {
			if s, ok := args[2].(string); ok {
				value = s
			}
		}
		return sdkethtool.SetOffload(iface, feature, value), nil
	}

	// ── systemd_analyze ───────────────────────────────────────────────────
	interp.builtins["systemd_analyze.time"] = func(args ...interface{}) (interface{}, error) {
		return sdksystemd_analyze.Time(), nil
	}
	interp.builtins["systemd_analyze.blame"] = func(args ...interface{}) (interface{}, error) {
		return sdksystemd_analyze.Blame(), nil
	}
	interp.builtins["systemd_analyze.critical_chain"] = func(args ...interface{}) (interface{}, error) {
		return sdksystemd_analyze.CriticalChain(), nil
	}
	interp.builtins["systemd_analyze.security"] = func(args ...interface{}) (interface{}, error) {
		unit := ""
		if len(args) > 0 {
			if s, ok := args[0].(string); ok {
				unit = s
			}
		}
		v, err := sdksystemd_analyze.Security(unit)
		return map[string]interface{}{"output": v}, err
	}
	interp.builtins["systemd_analyze.verify"] = func(args ...interface{}) (interface{}, error) {
		unit := ""
		if len(args) > 0 {
			if s, ok := args[0].(string); ok {
				unit = s
			}
		}
		v, err := sdksystemd_analyze.Verify(unit)
		return map[string]interface{}{"output": v}, err
	}

	// ── nvme ──────────────────────────────────────────────────────────────
	interp.builtins["nvme.list"] = func(args ...interface{}) (interface{}, error) {
		return sdknvme.List(), nil
	}
	interp.builtins["nvme.smart_log"] = func(args ...interface{}) (interface{}, error) {
		device := ""
		if len(args) > 0 {
			if s, ok := args[0].(string); ok {
				device = s
			}
		}
		v, err := sdknvme.SmartLog(device)
		return map[string]interface{}{"output": v}, err
	}
	interp.builtins["nvme.firmware_log"] = func(args ...interface{}) (interface{}, error) {
		device := ""
		if len(args) > 0 {
			if s, ok := args[0].(string); ok {
				device = s
			}
		}
		v, err := sdknvme.FirmwareLog(device)
		return map[string]interface{}{"output": v}, err
	}
	interp.builtins["nvme.error_log"] = func(args ...interface{}) (interface{}, error) {
		device := ""
		if len(args) > 0 {
			if s, ok := args[0].(string); ok {
				device = s
			}
		}
		v, err := sdknvme.ErrorLog(device)
		return map[string]interface{}{"output": v}, err
	}
	interp.builtins["nvme.version"] = func(args ...interface{}) (interface{}, error) {
		v, err := sdknvme.Version()
		return map[string]interface{}{"version": v}, err
	}

	// ── lshw ──────────────────────────────────────────────────────────────
	interp.builtins["lshw.short"] = func(args ...interface{}) (interface{}, error) {
		return sdkslshw.Short(), nil
	}
	interp.builtins["lshw.class"] = func(args ...interface{}) (interface{}, error) {
		class := ""
		if len(args) > 0 {
			if s, ok := args[0].(string); ok {
				class = s
			}
		}
		v, err := sdkslshw.Class(class)
		return map[string]interface{}{"output": v}, err
	}
	interp.builtins["lshw.json"] = func(args ...interface{}) (interface{}, error) {
		v, err := sdkslshw.JSON()
		return map[string]interface{}{"output": v}, err
	}
	interp.builtins["lshw.system"] = func(args ...interface{}) (interface{}, error) {
		v, err := sdkslshw.System()
		return map[string]interface{}{"output": v}, err
	}
	interp.builtins["lshw.memory"] = func(args ...interface{}) (interface{}, error) {
		v, err := sdkslshw.Memory()
		return map[string]interface{}{"output": v}, err
	}
	interp.builtins["lshw.disk"] = func(args ...interface{}) (interface{}, error) {
		v, err := sdkslshw.Disk()
		return map[string]interface{}{"output": v}, err
	}
	interp.builtins["lshw.network"] = func(args ...interface{}) (interface{}, error) {
		v, err := sdkslshw.Network()
		return map[string]interface{}{"output": v}, err
	}

	// ── ipaddr ────────────────────────────────────────────────────────────
	interp.builtins["ipaddr.list"] = func(args ...interface{}) (interface{}, error) {
		return sdkipaddr.List(), nil
	}
	interp.builtins["ipaddr.list_interface"] = func(args ...interface{}) (interface{}, error) {
		iface := ""
		if len(args) > 0 {
			if s, ok := args[0].(string); ok {
				iface = s
			}
		}
		return sdkipaddr.ListInterface(iface), nil
	}
	interp.builtins["ipaddr.add"] = func(args ...interface{}) (interface{}, error) {
		addr, iface := "", ""
		if len(args) > 0 {
			if s, ok := args[0].(string); ok {
				addr = s
			}
		}
		if len(args) > 1 {
			if s, ok := args[1].(string); ok {
				iface = s
			}
		}
		return sdkipaddr.Add(addr, iface), nil
	}
	interp.builtins["ipaddr.delete"] = func(args ...interface{}) (interface{}, error) {
		addr, iface := "", ""
		if len(args) > 0 {
			if s, ok := args[0].(string); ok {
				addr = s
			}
		}
		if len(args) > 1 {
			if s, ok := args[1].(string); ok {
				iface = s
			}
		}
		return sdkipaddr.Delete(addr, iface), nil
	}
	interp.builtins["ipaddr.flush"] = func(args ...interface{}) (interface{}, error) {
		iface := ""
		if len(args) > 0 {
			if s, ok := args[0].(string); ok {
				iface = s
			}
		}
		return sdkipaddr.Flush(iface), nil
	}
	interp.builtins["ipaddr.links"] = func(args ...interface{}) (interface{}, error) {
		return sdkipaddr.Links(), nil
	}
	interp.builtins["ipaddr.link_up"] = func(args ...interface{}) (interface{}, error) {
		iface := ""
		if len(args) > 0 {
			if s, ok := args[0].(string); ok {
				iface = s
			}
		}
		return sdkipaddr.LinkUp(iface), nil
	}
	interp.builtins["ipaddr.link_down"] = func(args ...interface{}) (interface{}, error) {
		iface := ""
		if len(args) > 0 {
			if s, ok := args[0].(string); ok {
				iface = s
			}
		}
		return sdkipaddr.LinkDown(iface), nil
	}

	// ── udevadm ───────────────────────────────────────────────────────────
	interp.builtins["udevadm.control"] = func(args ...interface{}) (interface{}, error) {
		action := ""
		if len(args) > 0 {
			if s, ok := args[0].(string); ok {
				action = s
			}
		}
		return sdkudevadm.Control(action), nil
	}
	interp.builtins["udevadm.trigger"] = func(args ...interface{}) (interface{}, error) {
		subsystem := ""
		if len(args) > 0 {
			if s, ok := args[0].(string); ok {
				subsystem = s
			}
		}
		return sdkudevadm.Trigger(subsystem), nil
	}
	interp.builtins["udevadm.settle"] = func(args ...interface{}) (interface{}, error) {
		timeout := 120
		if len(args) > 0 {
			if i, ok := args[0].(int); ok {
				timeout = i
			}
		}
		return sdkudevadm.Settle(timeout), nil
	}
	interp.builtins["udevadm.info"] = func(args ...interface{}) (interface{}, error) {
		query, device := "", ""
		if len(args) > 0 {
			if s, ok := args[0].(string); ok {
				query = s
			}
		}
		if len(args) > 1 {
			if s, ok := args[1].(string); ok {
				device = s
			}
		}
		return sdkudevadm.Info(query, device), nil
	}
	interp.builtins["udevadm.monitor"] = func(args ...interface{}) (interface{}, error) {
		return sdkudevadm.Monitor(), nil
	}

	// ── modinfo ───────────────────────────────────────────────────────────
	interp.builtins["modinfo.info"] = func(args ...interface{}) (interface{}, error) {
		module := ""
		if len(args) > 0 {
			if s, ok := args[0].(string); ok {
				module = s
			}
		}
		return sdkmodinfo.Info(module), nil
	}
	interp.builtins["modinfo.list"] = func(args ...interface{}) (interface{}, error) {
		return sdkmodinfo.List(), nil
	}
	interp.builtins["modinfo.version"] = func(args ...interface{}) (interface{}, error) {
		return sdkmodinfo.Version(), nil
	}

	// ── dconf ─────────────────────────────────────────────────────────────
	interp.builtins["dconf.read"] = func(args ...interface{}) (interface{}, error) {
		key := ""
		if len(args) > 0 {
			if s, ok := args[0].(string); ok {
				key = s
			}
		}
		return sdkdconf.Read(key), nil
	}
	interp.builtins["dconf.write"] = func(args ...interface{}) (interface{}, error) {
		key, value := "", ""
		if len(args) > 0 {
			if s, ok := args[0].(string); ok {
				key = s
			}
		}
		if len(args) > 1 {
			if s, ok := args[1].(string); ok {
				value = s
			}
		}
		return sdkdconf.Write(key, value), nil
	}
	interp.builtins["dconf.list"] = func(args ...interface{}) (interface{}, error) {
		dir := ""
		if len(args) > 0 {
			if s, ok := args[0].(string); ok {
				dir = s
			}
		}
		return sdkdconf.List(dir), nil
	}
	interp.builtins["dconf.reset"] = func(args ...interface{}) (interface{}, error) {
		key := ""
		if len(args) > 0 {
			if s, ok := args[0].(string); ok {
				key = s
			}
		}
		return sdkdconf.Reset(key), nil
	}

	// ── locale_gen ────────────────────────────────────────────────────────
	interp.builtins["locale_gen.generate"] = func(args ...interface{}) (interface{}, error) {
		locale := ""
		if len(args) > 0 {
			if s, ok := args[0].(string); ok {
				locale = s
			}
		}
		return sdklocale_gen.Generate(locale), nil
	}
	interp.builtins["locale_gen.list"] = func(args ...interface{}) (interface{}, error) {
		return sdklocale_gen.List(), nil
	}
	interp.builtins["locale_gen.remove"] = func(args ...interface{}) (interface{}, error) {
		locale := ""
		if len(args) > 0 {
			if s, ok := args[0].(string); ok {
				locale = s
			}
		}
		return sdklocale_gen.Remove(locale), nil
	}

	// ── pam_limits ────────────────────────────────────────────────────────
	interp.builtins["pam_limits.set"] = func(args ...interface{}) (interface{}, error) {
		domain, limitType, item, value := "", "", "", ""
		if len(args) > 0 {
			if s, ok := args[0].(string); ok {
				domain = s
			}
		}
		if len(args) > 1 {
			if s, ok := args[1].(string); ok {
				limitType = s
			}
		}
		if len(args) > 2 {
			if s, ok := args[2].(string); ok {
				item = s
			}
		}
		if len(args) > 3 {
			if s, ok := args[3].(string); ok {
				value = s
			}
		}
		return sdkpam_limits.Set(domain, limitType, item, value), nil
	}
	interp.builtins["pam_limits.list"] = func(args ...interface{}) (interface{}, error) {
		return sdkpam_limits.List(), nil
	}

	// ── motd ──────────────────────────────────────────────────────────────
	interp.builtins["motd.read"] = func(args ...interface{}) (interface{}, error) {
		return sdkmotd.Read(), nil
	}
	interp.builtins["motd.write"] = func(args ...interface{}) (interface{}, error) {
		content := ""
		if len(args) > 0 {
			if s, ok := args[0].(string); ok {
				content = s
			}
		}
		return sdkmotd.Write(content), nil
	}

	// ── issue ─────────────────────────────────────────────────────────────
	interp.builtins["issue.read"] = func(args ...interface{}) (interface{}, error) {
		return sdkissue.Read(), nil
	}
	interp.builtins["issue.write"] = func(args ...interface{}) (interface{}, error) {
		content := ""
		if len(args) > 0 {
			if s, ok := args[0].(string); ok {
				content = s
			}
		}
		return sdkissue.Write(content), nil
	}

	// ── authorized_key ──────────────────────────────────────────────────────
	interp.builtins["authorized_key.manage"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 4 {
			return nil, fmt.Errorf("authorized_key.manage() requires 4 arguments")
		}
		username, _ := args[0].(string)
		key, _ := args[1].(string)
		state, _ := args[2].(string)
		path, _ := args[3].(string)
		return sdkauthorized_key.Manage(username, key, state, path), nil
	}
	interp.builtins["authorized_key.list"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("authorized_key.list() requires 2 arguments")
		}
		username, _ := args[0].(string)
		path, _ := args[1].(string)
		return sdkauthorized_key.List(username, path), nil
	}
	interp.builtins["authorized_key.check"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 3 {
			return nil, fmt.Errorf("authorized_key.check() requires 3 arguments")
		}
		username, _ := args[0].(string)
		key, _ := args[1].(string)
		path, _ := args[2].(string)
		return sdkauthorized_key.Check(username, key, path), nil
	}

	// ── blockinfile ─────────────────────────────────────────────────────────
	interp.builtins["blockinfile.manage"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 6 {
			return nil, fmt.Errorf("blockinfile.manage() requires 6 arguments")
		}
		path, _ := args[0].(string)
		block, _ := args[1].(string)
		state, _ := args[2].(string)
		marker, _ := args[3].(string)
		insertAfter, _ := args[4].(string)
		insertBefore, _ := args[5].(string)
		return sdkblockinfile.Manage(path, block, state, marker, insertAfter, insertBefore), nil
	}
	interp.builtins["blockinfile.read"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("blockinfile.read() requires 2 arguments")
		}
		path, _ := args[0].(string)
		marker, _ := args[1].(string)
		content, found, err := sdkblockinfile.Read(path, marker)
		errStr := ""
		if err != nil {
			errStr = err.Error()
		}
		return map[string]interface{}{"content": content, "found": found, "error": errStr}, nil
	}

	// ── debconf ─────────────────────────────────────────────────────────────
	interp.builtins["debconf.set"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 4 {
			return nil, fmt.Errorf("debconf.set() requires 4 arguments")
		}
		pkg, _ := args[0].(string)
		name, _ := args[1].(string)
		vtype, _ := args[2].(string)
		value, _ := args[3].(string)
		return sdkdebconf.Set(pkg, name, vtype, value), nil
	}
	interp.builtins["debconf.get"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("debconf.get() requires 2 arguments")
		}
		pkg, _ := args[0].(string)
		name, _ := args[1].(string)
		return sdkdebconf.Get(pkg, name), nil
	}
	interp.builtins["debconf.list"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("debconf.list() requires 1 argument")
		}
		pkg, _ := args[0].(string)
		return sdkdebconf.List(pkg), nil
	}

	// ── reboot ──────────────────────────────────────────────────────────────
	interp.builtins["reboot.request"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("reboot.request() requires 2 arguments")
		}
		msg, _ := args[0].(string)
		delay, _ := args[1].(int)
		return sdkreboot.Request(msg, delay), nil
	}
	interp.builtins["reboot.dry_run"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("reboot.dry_run() requires 2 arguments")
		}
		msg, _ := args[0].(string)
		delay, _ := args[1].(int)
		return sdkreboot.DryRun(msg, delay), nil
	}
	interp.builtins["reboot.check"] = func(args ...interface{}) (interface{}, error) {
		return sdkreboot.Check(), nil
	}

	// ── swap ────────────────────────────────────────────────────────────────
	interp.builtins["swap.info"] = func(args ...interface{}) (interface{}, error) {
		return sdkswap.Info(), nil
	}
	interp.builtins["swap.create"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("swap.create() requires 2 arguments")
		}
		path, _ := args[0].(string)
		sizeMB, _ := args[1].(int)
		return sdkswap.Create(path, sizeMB), nil
	}
	interp.builtins["swap.enable"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("swap.enable() requires 1 argument")
		}
		device, _ := args[0].(string)
		return sdkswap.Enable(device), nil
	}
	interp.builtins["swap.disable"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("swap.disable() requires 1 argument")
		}
		device, _ := args[0].(string)
		return sdkswap.Disable(device), nil
	}

	// ── raw ─────────────────────────────────────────────────────────────────
	interp.builtins["raw.execute"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("raw.execute() requires 2 arguments")
		}
		command, _ := args[0].(string)
		timeout, _ := args[1].(int)
		return sdkraw.Execute(command, timeout), nil
	}
	interp.builtins["raw.execute_with_env"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 3 {
			return nil, fmt.Errorf("raw.execute_with_env() requires 3 arguments")
		}
		command, _ := args[0].(string)
		timeout, _ := args[1].(int)
		env := toStringMap(args, 2)
		return sdkraw.ExecuteWithEnv(command, timeout, env), nil
	}

	// ── expect ──────────────────────────────────────────────────────────────
	interp.builtins["expect.run"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 3 {
			return nil, fmt.Errorf("expect.run() requires 3 arguments")
		}
		command, _ := args[0].(string)
		responses := toStringMap(args, 1)
		timeout, _ := args[2].(int)
		return sdkexpect.Run(command, responses, timeout), nil
	}
	interp.builtins["expect.run_simple"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 4 {
			return nil, fmt.Errorf("expect.run_simple() requires 4 arguments")
		}
		command, _ := args[0].(string)
		prompt, _ := args[1].(string)
		response, _ := args[2].(string)
		timeout, _ := args[3].(int)
		return sdkexpect.RunSimple(command, prompt, response, timeout), nil
	}

	// ── slurp ───────────────────────────────────────────────────────────────
	interp.builtins["slurp.encode"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("slurp.encode() requires 1 argument")
		}
		path, _ := args[0].(string)
		return sdkslurp.Encode(path), nil
	}
	interp.builtins["slurp.decode"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("slurp.decode() requires 2 arguments")
		}
		encoded, _ := args[0].(string)
		destPath, _ := args[1].(string)
		return sdkslurp.Decode(encoded, destPath), nil
	}

	// ── wait_for_connection ─────────────────────────────────────────────────
	interp.builtins["wait_for_connection.wait"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 4 {
			return nil, fmt.Errorf("wait_for_connection.wait() requires 4 arguments")
		}
		host, _ := args[0].(string)
		port, _ := args[1].(int)
		timeout, _ := args[2].(int)
		delay, _ := args[3].(int)
		return sdkwait_for_connection.Wait(host, port, timeout, delay), nil
	}
	interp.builtins["wait_for_connection.check_once"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("wait_for_connection.check_once() requires 2 arguments")
		}
		host, _ := args[0].(string)
		port, _ := args[1].(int)
		return sdkwait_for_connection.CheckOnce(host, port), nil
	}

	// ── firewalld_rich_rule ─────────────────────────────────────────────────
	interp.builtins["firewalld_rich_rule.add"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("firewalld_rich_rule.add() requires 2 arguments")
		}
		zone, _ := args[0].(string)
		rule, _ := args[1].(string)
		return sdkfirewalld_rich_rule.Add(zone, rule), nil
	}
	interp.builtins["firewalld_rich_rule.remove"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("firewalld_rich_rule.remove() requires 2 arguments")
		}
		zone, _ := args[0].(string)
		rule, _ := args[1].(string)
		return sdkfirewalld_rich_rule.Remove(zone, rule), nil
	}
	interp.builtins["firewalld_rich_rule.list"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("firewalld_rich_rule.list() requires 1 argument")
		}
		zone, _ := args[0].(string)
		return sdkfirewalld_rich_rule.List(zone), nil
	}
	interp.builtins["firewalld_rich_rule.exists"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("firewalld_rich_rule.exists() requires 2 arguments")
		}
		zone, _ := args[0].(string)
		rule, _ := args[1].(string)
		return sdkfirewalld_rich_rule.Exists(zone, rule), nil
	}

	// ── firewalld_ipset ─────────────────────────────────────────────────────
	interp.builtins["firewalld_ipset.create"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("firewalld_ipset.create() requires 2 arguments")
		}
		name, _ := args[0].(string)
		setType, _ := args[1].(string)
		return sdkfirewalld_ipset.Create(name, setType), nil
	}
	interp.builtins["firewalld_ipset.delete"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("firewalld_ipset.delete() requires 1 argument")
		}
		name, _ := args[0].(string)
		return sdkfirewalld_ipset.Delete(name), nil
	}
	interp.builtins["firewalld_ipset.add_entry"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("firewalld_ipset.add_entry() requires 2 arguments")
		}
		name, _ := args[0].(string)
		entry, _ := args[1].(string)
		return sdkfirewalld_ipset.AddEntry(name, entry), nil
	}
	interp.builtins["firewalld_ipset.remove_entry"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("firewalld_ipset.remove_entry() requires 2 arguments")
		}
		name, _ := args[0].(string)
		entry, _ := args[1].(string)
		return sdkfirewalld_ipset.RemoveEntry(name, entry), nil
	}
	interp.builtins["firewalld_ipset.list"] = func(args ...interface{}) (interface{}, error) {
		return sdkfirewalld_ipset.List(), nil
	}
	interp.builtins["firewalld_ipset.info"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("firewalld_ipset.info() requires 1 argument")
		}
		name, _ := args[0].(string)
		return sdkfirewalld_ipset.Info(name), nil
	}

	// ── pause ───────────────────────────────────────────────────────────────
	interp.builtins["pause.seconds"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("pause.seconds() requires 1 argument")
		}
		duration, _ := args[0].(int)
		return sdkpause.Seconds(duration), nil
	}
	interp.builtins["pause.prompt"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("pause.prompt() requires 1 argument")
		}
		message, _ := args[0].(string)
		return sdkpause.Prompt(message), nil
	}
	interp.builtins["pause.prompt_with_default"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("pause.prompt_with_default() requires 2 arguments")
		}
		message, _ := args[0].(string)
		defaultVal, _ := args[1].(string)
		return sdkpause.PromptWithDefault(message, defaultVal), nil
	}

	// ── meta ────────────────────────────────────────────────────────────────
	interp.builtins["meta.end_host"] = func(args ...interface{}) (interface{}, error) {
		return sdkmeta.EndHost(), nil
	}
	interp.builtins["meta.end_play"] = func(args ...interface{}) (interface{}, error) {
		return sdkmeta.EndPlay(), nil
	}
	interp.builtins["meta.clear_host_errors"] = func(args ...interface{}) (interface{}, error) {
		return sdkmeta.ClearHostErrors(), nil
	}
	interp.builtins["meta.refresh_inventory"] = func(args ...interface{}) (interface{}, error) {
		return sdkmeta.RefreshInventory(), nil
	}
	interp.builtins["meta.flush_handlers"] = func(args ...interface{}) (interface{}, error) {
		return sdkmeta.FlushHandlers(), nil
	}
	interp.builtins["meta.reset_connection"] = func(args ...interface{}) (interface{}, error) {
		return sdkmeta.ResetConnection(), nil
	}
	interp.builtins["meta.noop"] = func(args ...interface{}) (interface{}, error) {
		return sdkmeta.Noop(), nil
	}
	interp.builtins["meta.fail"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("meta.fail() requires 1 argument")
		}
		message, _ := args[0].(string)
		return sdkmeta.Fail(message), nil
	}
	interp.builtins["meta.assert"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("meta.assert() requires 2 arguments")
		}
		condition, _ := args[0].(bool)
		message, _ := args[1].(string)
		return sdkmeta.Assert(condition, message), nil
	}
	interp.builtins["meta.debug"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("meta.debug() requires 1 argument")
		}
		message, _ := args[0].(string)
		return sdkmeta.Debug(message, nil), nil
	}

	// ── uri_ext ─────────────────────────────────────────────────────────────
	interp.builtins["uri_ext.patch"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 4 {
			return nil, fmt.Errorf("uri_ext.patch() requires 4 arguments")
		}
		url, _ := args[0].(string)
		body, _ := args[1].([]byte)
		headers := toStringMap(args, 2)
		timeout, _ := args[3].(int)
		return sdkuri_ext.Patch(url, body, headers, timeout), nil
	}
	interp.builtins["uri_ext.delete"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 3 {
			return nil, fmt.Errorf("uri_ext.delete() requires 3 arguments")
		}
		url, _ := args[0].(string)
		headers := toStringMap(args, 1)
		timeout, _ := args[2].(int)
		return sdkuri_ext.Delete(url, headers, timeout), nil
	}
	interp.builtins["uri_ext.head"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 3 {
			return nil, fmt.Errorf("uri_ext.head() requires 3 arguments")
		}
		url, _ := args[0].(string)
		headers := toStringMap(args, 1)
		timeout, _ := args[2].(int)
		return sdkuri_ext.Head(url, headers, timeout), nil
	}
	interp.builtins["uri_ext.options"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 3 {
			return nil, fmt.Errorf("uri_ext.options() requires 3 arguments")
		}
		url, _ := args[0].(string)
		headers := toStringMap(args, 1)
		timeout, _ := args[2].(int)
		return sdkuri_ext.Options(url, headers, timeout), nil
	}

	// ── hwclock ─────────────────────────────────────────────────────────────
	interp.builtins["hwclock.get"] = func(args ...interface{}) (interface{}, error) {
		return sdkhwclock.Get(), nil
	}
	interp.builtins["hwclock.set"] = func(args ...interface{}) (interface{}, error) {
		return sdkhwclock.Set(), nil
	}
	interp.builtins["hwclock.hctosys"] = func(args ...interface{}) (interface{}, error) {
		return sdkhwclock.HCToSys(), nil
	}
	interp.builtins["hwclock.set_time"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("hwclock.set_time() requires 1 argument")
		}
		timeStr, _ := args[0].(string)
		return sdkhwclock.SetTime(timeStr), nil
	}

	// ── mdadm ───────────────────────────────────────────────────────────────
	interp.builtins["mdadm.create"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 3 {
			return nil, fmt.Errorf("mdadm.create() requires 3 arguments")
		}
		device, _ := args[0].(string)
		level, _ := args[1].(string)
		devices := args[2].([]interface{})
		devStrs := make([]string, len(devices))
		for i, d := range devices {
			devStrs[i], _ = d.(string)
		}
		return sdkmdadm.Create(device, level, devStrs), nil
	}
	interp.builtins["mdadm.destroy"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("mdadm.destroy() requires 1 argument")
		}
		device, _ := args[0].(string)
		return sdkmdadm.Destroy(device), nil
	}
	interp.builtins["mdadm.detail"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("mdadm.detail() requires 1 argument")
		}
		device, _ := args[0].(string)
		return sdkmdadm.Detail(device), nil
	}
	interp.builtins["mdadm.scan"] = func(args ...interface{}) (interface{}, error) {
		return sdkmdadm.Scan(), nil
	}
	interp.builtins["mdadm.add"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("mdadm.add() requires 2 arguments")
		}
		device, _ := args[0].(string)
		member, _ := args[1].(string)
		return sdkmdadm.Add(device, member), nil
	}
	interp.builtins["mdadm.remove"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("mdadm.remove() requires 2 arguments")
		}
		device, _ := args[0].(string)
		member, _ := args[1].(string)
		return sdkmdadm.Remove(device, member), nil
	}

	// ── open_iscsi ──────────────────────────────────────────────────────────
	interp.builtins["open_iscsi.discover"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("open_iscsi.discover() requires 2 arguments")
		}
		portal, _ := args[0].(string)
		port, _ := args[1].(int)
		return sdkopen_iscsi.Discover(portal, port), nil
	}
	interp.builtins["open_iscsi.login"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("open_iscsi.login() requires 2 arguments")
		}
		target, _ := args[0].(string)
		portal, _ := args[1].(string)
		return sdkopen_iscsi.Login(target, portal), nil
	}
	interp.builtins["open_iscsi.logout"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("open_iscsi.logout() requires 2 arguments")
		}
		target, _ := args[0].(string)
		portal, _ := args[1].(string)
		return sdkopen_iscsi.Logout(target, portal), nil
	}
	interp.builtins["open_iscsi.list_sessions"] = func(args ...interface{}) (interface{}, error) {
		return sdkopen_iscsi.ListSessions(), nil
	}
	interp.builtins["open_iscsi.list_nodes"] = func(args ...interface{}) (interface{}, error) {
		return sdkopen_iscsi.ListNodes(), nil
	}
	interp.builtins["open_iscsi.set_startup"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 3 {
			return nil, fmt.Errorf("open_iscsi.set_startup() requires 3 arguments")
		}
		target, _ := args[0].(string)
		portal, _ := args[1].(string)
		startup, _ := args[2].(string)
		return sdkopen_iscsi.SetStartup(target, portal, startup), nil
	}

	// ── rfkill ──────────────────────────────────────────────────────────────
	interp.builtins["rfkill.list"] = func(args ...interface{}) (interface{}, error) {
		return sdkrfkill.List(), nil
	}
	interp.builtins["rfkill.block"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("rfkill.block() requires 1 argument")
		}
		device, _ := args[0].(string)
		return sdkrfkill.Block(device), nil
	}
	interp.builtins["rfkill.unblock"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("rfkill.unblock() requires 1 argument")
		}
		device, _ := args[0].(string)
		return sdkrfkill.Unblock(device), nil
	}
	interp.builtins["rfkill.block_all"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("rfkill.block_all() requires 1 argument")
		}
		deviceType, _ := args[0].(string)
		return sdkrfkill.BlockAll(deviceType), nil
	}
	interp.builtins["rfkill.unblock_all"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("rfkill.unblock_all() requires 1 argument")
		}
		deviceType, _ := args[0].(string)
		return sdkrfkill.UnblockAll(deviceType), nil
	}

	// ── multipath ───────────────────────────────────────────────────────────
	interp.builtins["multipath.reconfigure"] = func(args ...interface{}) (interface{}, error) {
		return sdkmultipath.Reconfigure(), nil
	}
	interp.builtins["multipath.list_paths"] = func(args ...interface{}) (interface{}, error) {
		return sdkmultipath.ListPaths(), nil
	}
	interp.builtins["multipath.list_maps"] = func(args ...interface{}) (interface{}, error) {
		return sdkmultipath.ListMaps(), nil
	}
	interp.builtins["multipath.add_map"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("multipath.add_map() requires 1 argument")
		}
		device, _ := args[0].(string)
		return sdkmultipath.AddMap(device), nil
	}
	interp.builtins["multipath.remove_map"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("multipath.remove_map() requires 1 argument")
		}
		device, _ := args[0].(string)
		return sdkmultipath.RemoveMap(device), nil
	}
	interp.builtins["multipath.flush"] = func(args ...interface{}) (interface{}, error) {
		return sdkmultipath.Flush(), nil
	}

	// ── dmsetup ─────────────────────────────────────────────────────────────
	interp.builtins["dmsetup.create"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("dmsetup.create() requires 2 arguments")
		}
		name, _ := args[0].(string)
		table, _ := args[1].(string)
		return sdkdmsetup.Create(name, table), nil
	}
	interp.builtins["dmsetup.remove"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("dmsetup.remove() requires 1 argument")
		}
		name, _ := args[0].(string)
		return sdkdmsetup.Remove(name), nil
	}
	interp.builtins["dmsetup.remove_all"] = func(args ...interface{}) (interface{}, error) {
		return sdkdmsetup.RemoveAll(), nil
	}
	interp.builtins["dmsetup.list"] = func(args ...interface{}) (interface{}, error) {
		return sdkdmsetup.List(), nil
	}
	interp.builtins["dmsetup.info"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("dmsetup.info() requires 1 argument")
		}
		name, _ := args[0].(string)
		return sdkdmsetup.Info(name), nil
	}
	interp.builtins["dmsetup.suspend"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("dmsetup.suspend() requires 1 argument")
		}
		name, _ := args[0].(string)
		return sdkdmsetup.Suspend(name), nil
	}
	interp.builtins["dmsetup.resume"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("dmsetup.resume() requires 1 argument")
		}
		name, _ := args[0].(string)
		return sdkdmsetup.Resume(name), nil
	}

	// ── lvm_enhanced ────────────────────────────────────────────────────────
	interp.builtins["lvm_enhanced.pv_create"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("lvm_enhanced.pv_create() requires 1 argument")
		}
		device, _ := args[0].(string)
		return sdklvm_enhanced.PVCreate(device), nil
	}
	interp.builtins["lvm_enhanced.pv_remove"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("lvm_enhanced.pv_remove() requires 2 arguments")
		}
		device, _ := args[0].(string)
		force, _ := args[1].(bool)
		return sdklvm_enhanced.PVRemove(device, force), nil
	}
	interp.builtins["lvm_enhanced.pv_list"] = func(args ...interface{}) (interface{}, error) {
		return sdklvm_enhanced.PVList(), nil
	}
	interp.builtins["lvm_enhanced.vg_create"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("lvm_enhanced.vg_create() requires 2 arguments")
		}
		name, _ := args[0].(string)
		devicesRaw := args[1].([]interface{})
		devStrs := make([]string, len(devicesRaw))
		for i, d := range devicesRaw {
			devStrs[i], _ = d.(string)
		}
		return sdklvm_enhanced.VGCreate(name, devStrs), nil
	}
	interp.builtins["lvm_enhanced.vg_remove"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("lvm_enhanced.vg_remove() requires 2 arguments")
		}
		name, _ := args[0].(string)
		force, _ := args[1].(bool)
		return sdklvm_enhanced.VGRemove(name, force), nil
	}
	interp.builtins["lvm_enhanced.vg_extend"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("lvm_enhanced.vg_extend() requires 2 arguments")
		}
		vgName, _ := args[0].(string)
		device, _ := args[1].(string)
		return sdklvm_enhanced.VGExtend(vgName, device), nil
	}
	interp.builtins["lvm_enhanced.vg_list"] = func(args ...interface{}) (interface{}, error) {
		return sdklvm_enhanced.VGList(), nil
	}
	interp.builtins["lvm_enhanced.lv_extend"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("lvm_enhanced.lv_extend() requires 2 arguments")
		}
		lvPath, _ := args[0].(string)
		size, _ := args[1].(string)
		return sdklvm_enhanced.LVExtend(lvPath, size), nil
	}
	interp.builtins["lvm_enhanced.lv_extend_all"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("lvm_enhanced.lv_extend_all() requires 1 argument")
		}
		lvPath, _ := args[0].(string)
		return sdklvm_enhanced.LVExtendAll(lvPath), nil
	}
	interp.builtins["lvm_enhanced.lv_list"] = func(args ...interface{}) (interface{}, error) {
		return sdklvm_enhanced.LVList(), nil
	}

	// ── puppet ──────────────────────────────────────────────────────────────
	interp.builtins["puppet.run"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("puppet.run() requires 2 arguments")
		}
		environment, _ := args[0].(string)
		tagsRaw := args[1].([]interface{})
		tags := make([]string, len(tagsRaw))
		for i, t := range tagsRaw {
			tags[i], _ = t.(string)
		}
		return sdkpuppet.Run(environment, tags), nil
	}
	interp.builtins["puppet.run_noop"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("puppet.run_noop() requires 2 arguments")
		}
		environment, _ := args[0].(string)
		tagsRaw := args[1].([]interface{})
		tags := make([]string, len(tagsRaw))
		for i, t := range tagsRaw {
			tags[i], _ = t.(string)
		}
		return sdkpuppet.RunNoop(environment, tags), nil
	}
	interp.builtins["puppet.status"] = func(args ...interface{}) (interface{}, error) {
		return sdkpuppet.Status(), nil
	}
	interp.builtins["puppet.disable"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("puppet.disable() requires 1 argument")
		}
		message, _ := args[0].(string)
		return sdkpuppet.Disable(message), nil
	}
	interp.builtins["puppet.enable"] = func(args ...interface{}) (interface{}, error) {
		return sdkpuppet.Enable(), nil
	}
	interp.builtins["puppet.fact"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("puppet.fact() requires 1 argument")
		}
		name, _ := args[0].(string)
		return sdkpuppet.Fact(name), nil
	}
	interp.builtins["puppet.module_list"] = func(args ...interface{}) (interface{}, error) {
		return sdkpuppet.ModuleList(), nil
	}
	interp.builtins["puppet.module_install"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("puppet.module_install() requires 2 arguments")
		}
		name, _ := args[0].(string)
		version, _ := args[1].(string)
		return sdkpuppet.ModuleInstall(name, version), nil
	}

	// ── svn ───────────────────────────────────────────────────────────────
	interp.builtins["svn.checkout"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("svn.checkout() requires at least 2 arguments (url, dest)")
		}
		url, _ := args[0].(string)
		dest, _ := args[1].(string)
		revision := getStringArgBridge(args, 2, "")
		force := opsBool(args[3])
		r, err := sdksvn.Checkout(url, dest, revision, force)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["svn.update"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("svn.update() requires 1 argument (dest)")
		}
		dest, _ := args[0].(string)
		revision := getStringArgBridge(args, 1, "")
		r, err := sdksvn.Update(dest, revision)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["svn.export"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("svn.export() requires at least 2 arguments (url, dest)")
		}
		url, _ := args[0].(string)
		dest, _ := args[1].(string)
		revision := getStringArgBridge(args, 2, "")
		force := opsBool(args[3])
		r, err := sdksvn.Export(url, dest, revision, force)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["svn.status"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("svn.status() requires 1 argument (dest)")
		}
		dest, _ := args[0].(string)
		r, err := sdksvn.Status(dest)
		if err != nil {
			return nil, err
		}
		return r, nil
	}
	interp.builtins["svn.info"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("svn.info() requires 1 argument (dest)")
		}
		dest, _ := args[0].(string)
		r, err := sdksvn.Info(dest)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["svn.cleanup"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("svn.cleanup() requires 1 argument (dest)")
		}
		dest, _ := args[0].(string)
		r, err := sdksvn.Cleanup(dest)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["svn.revert"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("svn.revert() requires 1 argument (dest)")
		}
		dest, _ := args[0].(string)
		recursive := opsBool(args[1])
		r, err := sdksvn.Revert(dest, recursive)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── zypper ────────────────────────────────────────────────────────────
	interp.builtins["zypper.install"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("zypper.install() requires at least 1 argument (name)")
		}
		name, _ := args[0].(string)
		version := getStringArgBridge(args, 1, "")
		r, err := sdkzypper.Install(name, version)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["zypper.remove"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("zypper.remove() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		r, err := sdkzypper.Remove(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["zypper.update"] = func(args ...interface{}) (interface{}, error) {
		name := getStringArgBridge(args, 0, "")
		r, err := sdkzypper.Update(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["zypper.dist_upgrade"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkzypper.DistUpgrade()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["zypper.info"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("zypper.info() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		r, err := sdkzypper.Info(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["zypper.list"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkzypper.List()
		if err != nil {
			return nil, err
		}
		return r, nil
	}
	interp.builtins["zypper.clean"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkzypper.Clean()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["zypper.repo_list"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkzypper.RepoList()
		if err != nil {
			return nil, err
		}
		return r, nil
	}
	interp.builtins["zypper.repo_add"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("zypper.repo_add() requires 2 arguments (name, url)")
		}
		name, _ := args[0].(string)
		url, _ := args[1].(string)
		r, err := sdkzypper.RepoAdd(name, url)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["zypper.repo_remove"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("zypper.repo_remove() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		r, err := sdkzypper.RepoRemove(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["zypper.refresh"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkzypper.Refresh()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["zypper.search"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("zypper.search() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		r, err := sdkzypper.Search(name)
		if err != nil {
			return nil, err
		}
		return r, nil
	}
	interp.builtins["zypper.patch"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkzypper.Patch()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["zypper.pattern_install"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("zypper.pattern_install() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		r, err := sdkzypper.PatternInstall(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["zypper.pattern_remove"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("zypper.pattern_remove() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		r, err := sdkzypper.PatternRemove(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── pacman ────────────────────────────────────────────────────────────
	interp.builtins["pacman.install"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("pacman.install() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		r, err := sdkpacman.Install(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["pacman.remove"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("pacman.remove() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		cascade := opsBool(args[1])
		r, err := sdkpacman.Remove(name, cascade)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["pacman.update"] = func(args ...interface{}) (interface{}, error) {
		name := getStringArgBridge(args, 0, "")
		r, err := sdkpacman.Update(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["pacman.upgrade"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkpacman.Upgrade()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["pacman.info"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("pacman.info() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		r, err := sdkpacman.Info(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["pacman.list"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkpacman.List()
		if err != nil {
			return nil, err
		}
		return r, nil
	}
	interp.builtins["pacman.search"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("pacman.search() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		r, err := sdkpacman.Search(name)
		if err != nil {
			return nil, err
		}
		return r, nil
	}
	interp.builtins["pacman.clean"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkpacman.Clean()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["pacman.install_file"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("pacman.install_file() requires 1 argument (path)")
		}
		path, _ := args[0].(string)
		r, err := sdkpacman.InstallFile(path)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["pacman.remove_orphans"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkpacman.RemoveOrphans()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["pacman.update_database"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkpacman.UpdateDatabase()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── kubernetes.* ──────────────────────────────────────────────────────
	interp.builtins["kubernetes.apply"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("kubernetes.apply() requires at least 1 argument (manifest)")
		}
		manifest, _ := args[0].(string)
		namespace, _ := args[1].(string)
		dryRun, _ := args[2].(bool)
		r, err := sdkk8s.Apply(manifest, namespace, dryRun)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["kubernetes.delete"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("kubernetes.delete() requires at least 1 argument (manifest)")
		}
		manifest, _ := args[0].(string)
		namespace, _ := args[1].(string)
		r, err := sdkk8s.Delete(manifest, namespace)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["kubernetes.get"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("kubernetes.get() requires at least 2 arguments (resource_type, name)")
		}
		rt, _ := args[0].(string)
		name, _ := args[1].(string)
		ns, _ := args[2].(string)
		r, err := sdkk8s.Get(rt, name, ns)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["kubernetes.list"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("kubernetes.list() requires at least 1 argument (resource_type)")
		}
		rt, _ := args[0].(string)
		ns, _ := args[1].(string)
		labels, _ := args[2].(string)
		r, err := sdkk8s.List(rt, ns, labels)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["kubernetes.create_namespace"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("kubernetes.create_namespace() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		r, err := sdkk8s.CreateNamespace(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["kubernetes.delete_namespace"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("kubernetes.delete_namespace() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		r, err := sdkk8s.DeleteNamespace(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["kubernetes.get_pods"] = func(args ...interface{}) (interface{}, error) {
		ns, _ := args[0].(string)
		labels, _ := args[1].(string)
		r, err := sdkk8s.GetPods(ns, labels)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["kubernetes.get_services"] = func(args ...interface{}) (interface{}, error) {
		ns, _ := args[0].(string)
		r, err := sdkk8s.GetServices(ns)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["kubernetes.get_deployments"] = func(args ...interface{}) (interface{}, error) {
		ns, _ := args[0].(string)
		r, err := sdkk8s.GetDeployments(ns)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["kubernetes.scale"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("kubernetes.scale() requires at least 2 arguments (deployment, replicas)")
		}
		dep, _ := args[0].(string)
		replicas := 1
		if v, ok := args[1].(int); ok {
			replicas = v
		} else if v, ok := args[1].(float64); ok {
			replicas = int(v)
		}
		ns, _ := args[2].(string)
		r, err := sdkk8s.Scale(dep, replicas, ns)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["kubernetes.rollout_status"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("kubernetes.rollout_status() requires at least 1 argument (deployment)")
		}
		dep, _ := args[0].(string)
		ns, _ := args[1].(string)
		r, err := sdkk8s.RolloutStatus(dep, ns)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["kubernetes.exec"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("kubernetes.exec() requires at least 2 arguments (pod, command)")
		}
		pod, _ := args[0].(string)
		cmd, _ := args[1].(string)
		ns, _ := args[2].(string)
		container, _ := args[3].(string)
		r, err := sdkk8s.Exec(pod, cmd, ns, container)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["kubernetes.logs"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("kubernetes.logs() requires at least 1 argument (pod)")
		}
		pod, _ := args[0].(string)
		ns, _ := args[1].(string)
		container, _ := args[2].(string)
		tail := 0
		if v, ok := args[3].(int); ok {
			tail = v
		} else if v, ok := args[3].(float64); ok {
			tail = int(v)
		}
		r, err := sdkk8s.Logs(pod, ns, container, tail)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["kubernetes.wait_ready"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("kubernetes.wait_ready() requires at least 2 arguments (resource_type, name)")
		}
		rt, _ := args[0].(string)
		name, _ := args[1].(string)
		ns, _ := args[2].(string)
		timeout := 300
		if v, ok := args[3].(int); ok {
			timeout = v
		} else if v, ok := args[3].(float64); ok {
			timeout = int(v)
		}
		r, err := sdkk8s.WaitReady(rt, name, ns, timeout)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}

	// ── portage.* ───────────────────────────────────────────────────
	interp.builtins["portage.install"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("portage.install() requires at least 1 argument (name)")
		}
		name, _ := args[0].(string)
		version := ""
		if len(args) >= 2 {
			version, _ = args[1].(string)
		}
		r, err := sdkportage.Install(name, version)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["portage.remove"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("portage.remove() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		r, err := sdkportage.Remove(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["portage.update"] = func(args ...interface{}) (interface{}, error) {
		name := ""
		if len(args) >= 1 {
			name, _ = args[0].(string)
		}
		deep := false
		if len(args) >= 2 {
			deep, _ = args[1].(bool)
		}
		r, err := sdkportage.Update(name, deep)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["portage.sync"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkportage.Sync()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["portage.info"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("portage.info() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		r, err := sdkportage.Info(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["portage.list"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkportage.List()
		if err != nil {
			return nil, err
		}
		return r, nil
	}
	interp.builtins["portage.search"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("portage.search() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		r, err := sdkportage.Search(name)
		if err != nil {
			return nil, err
		}
		return r, nil
	}
	interp.builtins["portage.depclean"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkportage.Depclean()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["portage.metadata"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("portage.metadata() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		r, err := sdkportage.Metadata(name)
		if err != nil {
			return nil, err
		}
		return r, nil
	}

	// ── pkgng.* ─────────────────────────────────────────────────────
	interp.builtins["pkgng.install"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("pkgng.install() requires at least 1 argument (name)")
		}
		name, _ := args[0].(string)
		version := ""
		if len(args) >= 2 {
			version, _ = args[1].(string)
		}
		r, err := sdkpkgng.Install(name, version)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["pkgng.remove"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("pkgng.remove() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		r, err := sdkpkgng.Remove(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["pkgng.update"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkpkgng.Update()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["pkgng.upgrade"] = func(args ...interface{}) (interface{}, error) {
		name := ""
		if len(args) >= 1 {
			name, _ = args[0].(string)
		}
		r, err := sdkpkgng.Upgrade(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["pkgng.autoclean"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkpkgng.Autoclean()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["pkgng.info"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("pkgng.info() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		r, err := sdkpkgng.Info(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["pkgng.list"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkpkgng.List()
		if err != nil {
			return nil, err
		}
		return r, nil
	}
	interp.builtins["pkgng.search"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("pkgng.search() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		r, err := sdkpkgng.Search(name)
		if err != nil {
			return nil, err
		}
		return r, nil
	}
	interp.builtins["pkgng.stats"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkpkgng.Stats()
		if err != nil {
			return nil, err
		}
		return r, nil
	}

	// ── podman.* ────────────────────────────────────────────────────
	interp.builtins["podman.run"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("podman.run() requires at least 1 argument (image)")
		}
		image, _ := args[0].(string)
		name := ""
		if len(args) >= 2 {
			name, _ = args[1].(string)
		}
		command := ""
		if len(args) >= 3 {
			command, _ = args[2].(string)
		}
		r, err := sdkpodman.Run(image, name, command)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["podman.stop"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("podman.stop() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		timeout := 0
		if len(args) >= 2 {
			if v, ok := args[1].(int); ok {
				timeout = v
			} else if v, ok := args[1].(float64); ok {
				timeout = int(v)
			}
		}
		r, err := sdkpodman.Stop(name, timeout)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["podman.start"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("podman.start() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		r, err := sdkpodman.Start(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["podman.remove"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("podman.remove() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		force := false
		if len(args) >= 2 {
			force, _ = args[1].(bool)
		}
		r, err := sdkpodman.Remove(name, force)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["podman.list_containers"] = func(args ...interface{}) (interface{}, error) {
		all := false
		if len(args) >= 1 {
			all, _ = args[0].(bool)
		}
		r, err := sdkpodman.ListContainers(all)
		if err != nil {
			return nil, err
		}
		return r, nil
	}
	interp.builtins["podman.inspect"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("podman.inspect() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		r, err := sdkpodman.Inspect(name)
		if err != nil {
			return nil, err
		}
		return r, nil
	}
	interp.builtins["podman.pull"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("podman.pull() requires 1 argument (image)")
		}
		image, _ := args[0].(string)
		r, err := sdkpodman.Pull(image)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["podman.list_images"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkpodman.ListImages()
		if err != nil {
			return nil, err
		}
		return r, nil
	}
	interp.builtins["podman.remove_image"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("podman.remove_image() requires 1 argument (image_id)")
		}
		imageID, _ := args[0].(string)
		force := false
		if len(args) >= 2 {
			force, _ = args[1].(bool)
		}
		r, err := sdkpodman.RemoveImage(imageID, force)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["podman.create_pod"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("podman.create_pod() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		r, err := sdkpodman.CreatePod(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["podman.stop_pod"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("podman.stop_pod() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		r, err := sdkpodman.StopPod(name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["podman.remove_pod"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("podman.remove_pod() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		force := false
		if len(args) >= 2 {
			force, _ = args[1].(bool)
		}
		r, err := sdkpodman.RemovePod(name, force)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["podman.list_pods"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkpodman.ListPods()
		if err != nil {
			return nil, err
		}
		return r, nil
	}

	// ── nftables.* ──────────────────────────────────────────────────
	interp.builtins["nftables.add_table"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("nftables.add_table() requires 2 arguments (family, name)")
		}
		family, _ := args[0].(string)
		name, _ := args[1].(string)
		r, err := sdknftables.AddTable(family, name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["nftables.delete_table"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("nftables.delete_table() requires 2 arguments (family, name)")
		}
		family, _ := args[0].(string)
		name, _ := args[1].(string)
		r, err := sdknftables.DeleteTable(family, name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["nftables.list_tables"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdknftables.ListTables()
		if err != nil {
			return nil, err
		}
		return r, nil
	}
	interp.builtins["nftables.add_chain"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 3 {
			return nil, fmt.Errorf("nftables.add_chain() requires at least 3 arguments (family, table, name)")
		}
		family, _ := args[0].(string)
		table, _ := args[1].(string)
		name, _ := args[2].(string)
		chainType := ""
		if len(args) >= 4 {
			chainType, _ = args[3].(string)
		}
		hook := ""
		if len(args) >= 5 {
			hook, _ = args[4].(string)
		}
		priority := ""
		if len(args) >= 6 {
			priority, _ = args[5].(string)
		}
		r, err := sdknftables.AddChain(family, table, name, chainType, hook, priority)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["nftables.delete_chain"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 3 {
			return nil, fmt.Errorf("nftables.delete_chain() requires 3 arguments (family, table, name)")
		}
		family, _ := args[0].(string)
		table, _ := args[1].(string)
		name, _ := args[2].(string)
		r, err := sdknftables.DeleteChain(family, table, name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["nftables.add_rule"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 4 {
			return nil, fmt.Errorf("nftables.add_rule() requires 4 arguments (family, table, chain, expression)")
		}
		family, _ := args[0].(string)
		table, _ := args[1].(string)
		chain, _ := args[2].(string)
		expr, _ := args[3].(string)
		r, err := sdknftables.AddRule(family, table, chain, expr)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["nftables.delete_rule"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 4 {
			return nil, fmt.Errorf("nftables.delete_rule() requires 4 arguments (family, table, chain, handle)")
		}
		family, _ := args[0].(string)
		table, _ := args[1].(string)
		chain, _ := args[2].(string)
		handle, _ := args[3].(string)
		r, err := sdknftables.DeleteRule(family, table, chain, handle)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["nftables.flush_chain"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 3 {
			return nil, fmt.Errorf("nftables.flush_chain() requires 3 arguments (family, table, chain)")
		}
		family, _ := args[0].(string)
		table, _ := args[1].(string)
		chain, _ := args[2].(string)
		r, err := sdknftables.FlushChain(family, table, chain)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["nftables.flush_table"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("nftables.flush_table() requires 2 arguments (family, table)")
		}
		family, _ := args[0].(string)
		table, _ := args[1].(string)
		r, err := sdknftables.FlushTable(family, table)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["nftables.flush_ruleset"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdknftables.FlushRuleset()
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["nftables.list_ruleset"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdknftables.ListRuleset()
		if err != nil {
			return nil, err
		}
		return r, nil
	}
	interp.builtins["nftables.add_set"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 4 {
			return nil, fmt.Errorf("nftables.add_set() requires at least 4 arguments (family, table, name, type)")
		}
		family, _ := args[0].(string)
		table, _ := args[1].(string)
		name, _ := args[2].(string)
		setType, _ := args[3].(string)
		flags := ""
		if len(args) >= 5 {
			flags, _ = args[4].(string)
		}
		r, err := sdknftables.AddSet(family, table, name, setType, flags)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["nftables.delete_set"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 3 {
			return nil, fmt.Errorf("nftables.delete_set() requires 3 arguments (family, table, name)")
		}
		family, _ := args[0].(string)
		table, _ := args[1].(string)
		name, _ := args[2].(string)
		r, err := sdknftables.DeleteSet(family, table, name)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["nftables.add_element"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 4 {
			return nil, fmt.Errorf("nftables.add_element() requires 4 arguments (family, table, set, element)")
		}
		family, _ := args[0].(string)
		table, _ := args[1].(string)
		set, _ := args[2].(string)
		element, _ := args[3].(string)
		r, err := sdknftables.AddElement(family, table, set, element)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["nftables.delete_element"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 4 {
			return nil, fmt.Errorf("nftables.delete_element() requires 4 arguments (family, table, set, element)")
		}
		family, _ := args[0].(string)
		table, _ := args[1].(string)
		set, _ := args[2].(string)
		element, _ := args[3].(string)
		r, err := sdknftables.DeleteElement(family, table, set, element)
		if err != nil {
			return nil, err
		}
		return structToMap(r)
	}
	interp.builtins["nftables.export"] = func(args ...interface{}) (interface{}, error) {
		format := "json"
		if len(args) >= 1 {
			format, _ = args[0].(string)
		}
		r, err := sdknftables.Export(format)
		if err != nil {
			return nil, err
		}
		return r, nil
	}

	// ── mongodb.* ────────────────────────────────────────────────────────────
	interp.builtins["mongodb.create_database"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 3 {
			return nil, fmt.Errorf("mongodb.create_database() requires 3 arguments (host, port, name)")
		}
		host, _ := args[0].(string)
		port := 27017
		if p, ok := args[1].(int); ok {
			port = p
		}
		name, _ := args[2].(string)
		r, err := sdkmongodb.CreateDatabase(host, port, name)
		if err != nil {
			return nil, err
		}
		return r, nil
	}
	interp.builtins["mongodb.drop_database"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 3 {
			return nil, fmt.Errorf("mongodb.drop_database() requires 3 arguments (host, port, name)")
		}
		host, _ := args[0].(string)
		port := 27017
		if p, ok := args[1].(int); ok {
			port = p
		}
		name, _ := args[2].(string)
		r, err := sdkmongodb.DropDatabase(host, port, name)
		if err != nil {
			return nil, err
		}
		return r, nil
	}
	interp.builtins["mongodb.list_databases"] = func(args ...interface{}) (interface{}, error) {
		host := "localhost"
		port := 27017
		if len(args) >= 1 {
			host, _ = args[0].(string)
		}
		if len(args) >= 2 {
			if p, ok := args[1].(int); ok {
				port = p
			}
		}
		r, err := sdkmongodb.ListDatabases(host, port)
		if err != nil {
			return nil, err
		}
		return r, nil
	}
	interp.builtins["mongodb.create_user"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 6 {
			return nil, fmt.Errorf("mongodb.create_user() requires 6 arguments (host, port, database, user, password, roles)")
		}
		host, _ := args[0].(string)
		port := 27017
		if p, ok := args[1].(int); ok {
			port = p
		}
		database, _ := args[2].(string)
		user, _ := args[3].(string)
		password, _ := args[4].(string)
		roles, _ := args[5].(string)
		r, err := sdkmongodb.CreateUser(host, port, database, user, password, roles)
		if err != nil {
			return nil, err
		}
		return r, nil
	}
	interp.builtins["mongodb.drop_user"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 4 {
			return nil, fmt.Errorf("mongodb.drop_user() requires 4 arguments (host, port, database, user)")
		}
		host, _ := args[0].(string)
		port := 27017
		if p, ok := args[1].(int); ok {
			port = p
		}
		database, _ := args[2].(string)
		user, _ := args[3].(string)
		r, err := sdkmongodb.DropUser(host, port, database, user)
		if err != nil {
			return nil, err
		}
		return r, nil
	}
	interp.builtins["mongodb.list_users"] = func(args ...interface{}) (interface{}, error) {
		host := "localhost"
		port := 27017
		database := "admin"
		if len(args) >= 1 {
			host, _ = args[0].(string)
		}
		if len(args) >= 2 {
			if p, ok := args[1].(int); ok {
				port = p
			}
		}
		if len(args) >= 3 {
			database, _ = args[2].(string)
		}
		r, err := sdkmongodb.ListUsers(host, port, database)
		if err != nil {
			return nil, err
		}
		return r, nil
	}
	interp.builtins["mongodb.create_collection"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 4 {
			return nil, fmt.Errorf("mongodb.create_collection() requires 4 arguments (host, port, database, collection)")
		}
		host, _ := args[0].(string)
		port := 27017
		if p, ok := args[1].(int); ok {
			port = p
		}
		database, _ := args[2].(string)
		collection, _ := args[3].(string)
		r, err := sdkmongodb.CreateCollection(host, port, database, collection)
		if err != nil {
			return nil, err
		}
		return r, nil
	}
	interp.builtins["mongodb.drop_collection"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 4 {
			return nil, fmt.Errorf("mongodb.drop_collection() requires 4 arguments (host, port, database, collection)")
		}
		host, _ := args[0].(string)
		port := 27017
		if p, ok := args[1].(int); ok {
			port = p
		}
		database, _ := args[2].(string)
		collection, _ := args[3].(string)
		r, err := sdkmongodb.DropCollection(host, port, database, collection)
		if err != nil {
			return nil, err
		}
		return r, nil
	}
	interp.builtins["mongodb.list_collections"] = func(args ...interface{}) (interface{}, error) {
		host := "localhost"
		port := 27017
		database := "admin"
		if len(args) >= 1 {
			host, _ = args[0].(string)
		}
		if len(args) >= 2 {
			if p, ok := args[1].(int); ok {
				port = p
			}
		}
		if len(args) >= 3 {
			database, _ = args[2].(string)
		}
		r, err := sdkmongodb.ListCollections(host, port, database)
		if err != nil {
			return nil, err
		}
		return r, nil
	}
	interp.builtins["mongodb.create_index"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 5 {
			return nil, fmt.Errorf("mongodb.create_index() requires at least 5 arguments (host, port, database, collection, keys)")
		}
		host, _ := args[0].(string)
		port := 27017
		if p, ok := args[1].(int); ok {
			port = p
		}
		database, _ := args[2].(string)
		collection, _ := args[3].(string)
		keys, _ := args[4].(string)
		unique := false
		name := ""
		if len(args) >= 6 {
			unique, _ = args[5].(bool)
		}
		if len(args) >= 7 {
			name, _ = args[6].(string)
		}
		r, err := sdkmongodb.CreateIndex(host, port, database, collection, keys, unique, name)
		if err != nil {
			return nil, err
		}
		return r, nil
	}
	interp.builtins["mongodb.drop_index"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 5 {
			return nil, fmt.Errorf("mongodb.drop_index() requires 5 arguments (host, port, database, collection, index_name)")
		}
		host, _ := args[0].(string)
		port := 27017
		if p, ok := args[1].(int); ok {
			port = p
		}
		database, _ := args[2].(string)
		collection, _ := args[3].(string)
		indexName, _ := args[4].(string)
		r, err := sdkmongodb.DropIndex(host, port, database, collection, indexName)
		if err != nil {
			return nil, err
		}
		return r, nil
	}
	interp.builtins["mongodb.list_indexes"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 4 {
			return nil, fmt.Errorf("mongodb.list_indexes() requires 4 arguments (host, port, database, collection)")
		}
		host, _ := args[0].(string)
		port := 27017
		if p, ok := args[1].(int); ok {
			port = p
		}
		database, _ := args[2].(string)
		collection, _ := args[3].(string)
		r, err := sdkmongodb.ListIndexes(host, port, database, collection)
		if err != nil {
			return nil, err
		}
		return r, nil
	}
	interp.builtins["mongodb.server_status"] = func(args ...interface{}) (interface{}, error) {
		host := "localhost"
		port := 27017
		if len(args) >= 1 {
			host, _ = args[0].(string)
		}
		if len(args) >= 2 {
			if p, ok := args[1].(int); ok {
				port = p
			}
		}
		r, err := sdkmongodb.ServerStatus(host, port)
		if err != nil {
			return nil, err
		}
		return r, nil
	}
	interp.builtins["mongodb.replica_set_status"] = func(args ...interface{}) (interface{}, error) {
		host := "localhost"
		port := 27017
		if len(args) >= 1 {
			host, _ = args[0].(string)
		}
		if len(args) >= 2 {
			if p, ok := args[1].(int); ok {
				port = p
			}
		}
		r, err := sdkmongodb.ReplicaSetStatus(host, port)
		if err != nil {
			return nil, err
		}
		return r, nil
	}

	// ── tomcat.* ────────────────────────────────────────────────────────────
	interp.builtins["tomcat.start"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("tomcat.start() requires 1 argument (catalina_home)")
		}
		home, _ := args[0].(string)
		r, err := sdktomcat.Start(home)
		if err != nil {
			return nil, err
		}
		return r, nil
	}
	interp.builtins["tomcat.stop"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("tomcat.stop() requires 1 argument (catalina_home)")
		}
		home, _ := args[0].(string)
		r, err := sdktomcat.Stop(home)
		if err != nil {
			return nil, err
		}
		return r, nil
	}
	interp.builtins["tomcat.restart"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("tomcat.restart() requires 1 argument (catalina_home)")
		}
		home, _ := args[0].(string)
		r, err := sdktomcat.Restart(home)
		if err != nil {
			return nil, err
		}
		return r, nil
	}
	interp.builtins["tomcat.status"] = func(args ...interface{}) (interface{}, error) {
		home := ""
		if len(args) >= 1 {
			home, _ = args[0].(string)
		}
		r, err := sdktomcat.Status(home)
		if err != nil {
			return nil, err
		}
		return r, nil
	}
	interp.builtins["tomcat.deploy"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("tomcat.deploy() requires at least 2 arguments (catalina_home, war_path)")
		}
		home, _ := args[0].(string)
		warPath, _ := args[1].(string)
		contextPath := ""
		if len(args) >= 3 {
			contextPath, _ = args[2].(string)
		}
		r, err := sdktomcat.Deploy(home, warPath, contextPath)
		if err != nil {
			return nil, err
		}
		return r, nil
	}
	interp.builtins["tomcat.undeploy"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("tomcat.undeploy() requires 2 arguments (catalina_home, context_path)")
		}
		home, _ := args[0].(string)
		contextPath, _ := args[1].(string)
		r, err := sdktomcat.Undeploy(home, contextPath)
		if err != nil {
			return nil, err
		}
		return r, nil
	}
	interp.builtins["tomcat.list_apps"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("tomcat.list_apps() requires 1 argument (catalina_home)")
		}
		home, _ := args[0].(string)
		r, err := sdktomcat.ListApps(home)
		if err != nil {
			return nil, err
		}
		return r, nil
	}
	interp.builtins["tomcat.reload"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("tomcat.reload() requires 2 arguments (catalina_home, context_path)")
		}
		home, _ := args[0].(string)
		contextPath, _ := args[1].(string)
		r, err := sdktomcat.Reload(home, contextPath)
		if err != nil {
			return nil, err
		}
		return r, nil
	}
	interp.builtins["tomcat.version"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("tomcat.version() requires 1 argument (catalina_home)")
		}
		home, _ := args[0].(string)
		r, err := sdktomcat.Version(home)
		if err != nil {
			return nil, err
		}
		return r, nil
	}

	// ── java_cert.* ─────────────────────────────────────────────────────────
	interp.builtins["java_cert.import"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 4 {
			return nil, fmt.Errorf("java_cert.import() requires at least 4 arguments (keystore_path, password, alias, cert_path)")
		}
		ksPath, _ := args[0].(string)
		password, _ := args[1].(string)
		alias, _ := args[2].(string)
		certPath, _ := args[3].(string)
		certType := ""
		if len(args) >= 5 {
			certType, _ = args[4].(string)
		}
		r, err := sdkjavacert.Import(ksPath, password, alias, certPath, certType)
		if err != nil {
			return nil, err
		}
		return r, nil
	}
	interp.builtins["java_cert.remove"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 3 {
			return nil, fmt.Errorf("java_cert.remove() requires 3 arguments (keystore_path, password, alias)")
		}
		ksPath, _ := args[0].(string)
		password, _ := args[1].(string)
		alias, _ := args[2].(string)
		r, err := sdkjavacert.Remove(ksPath, password, alias)
		if err != nil {
			return nil, err
		}
		return r, nil
	}
	interp.builtins["java_cert.list"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("java_cert.list() requires 2 arguments (keystore_path, password)")
		}
		ksPath, _ := args[0].(string)
		password, _ := args[1].(string)
		r, err := sdkjavacert.List(ksPath, password)
		if err != nil {
			return nil, err
		}
		return r, nil
	}
	interp.builtins["java_cert.exists"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 3 {
			return nil, fmt.Errorf("java_cert.exists() requires 3 arguments (keystore_path, password, alias)")
		}
		ksPath, _ := args[0].(string)
		password, _ := args[1].(string)
		alias, _ := args[2].(string)
		r, err := sdkjavacert.Exists(ksPath, password, alias)
		if err != nil {
			return nil, err
		}
		return r, nil
	}
	interp.builtins["java_cert.export"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 4 {
			return nil, fmt.Errorf("java_cert.export() requires at least 4 arguments (keystore_path, password, alias, output_path)")
		}
		ksPath, _ := args[0].(string)
		password, _ := args[1].(string)
		alias, _ := args[2].(string)
		outputPath, _ := args[3].(string)
		certType := ""
		if len(args) >= 5 {
			certType, _ = args[4].(string)
		}
		r, err := sdkjavacert.Export(ksPath, password, alias, outputPath, certType)
		if err != nil {
			return nil, err
		}
		return r, nil
	}
	interp.builtins["java_cert.info"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("java_cert.info() requires 2 arguments (keystore_path, password)")
		}
		ksPath, _ := args[0].(string)
		password, _ := args[1].(string)
		r, err := sdkjavacert.Info(ksPath, password)
		if err != nil {
			return nil, err
		}
		return r, nil
	}
	interp.builtins["java_cert.import_chain"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 4 {
			return nil, fmt.Errorf("java_cert.import_chain() requires 4 arguments (keystore_path, password, p12_path, p12_password)")
		}
		ksPath, _ := args[0].(string)
		password, _ := args[1].(string)
		p12Path, _ := args[2].(string)
		p12Password, _ := args[3].(string)
		r, err := sdkjavacert.ImportChain(ksPath, password, p12Path, p12Password)
		if err != nil {
			return nil, err
		}
		return r, nil
	}
	interp.builtins["java_cert.change_password"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 3 {
			return nil, fmt.Errorf("java_cert.change_password() requires 3 arguments (keystore_path, old_password, new_password)")
		}
		ksPath, _ := args[0].(string)
		oldPass, _ := args[1].(string)
		newPass, _ := args[2].(string)
		r, err := sdkjavacert.ChangePassword(ksPath, oldPass, newPass)
		if err != nil {
			return nil, err
		}
		return r, nil
	}

	// ── maven_artifact.* ────────────────────────────────────────────────────
	interp.builtins["maven_artifact.download"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 5 {
			return nil, fmt.Errorf("maven_artifact.download() requires at least 5 arguments (repo_url, group_id, artifact_id, version, dest)")
		}
		repoURL, _ := args[0].(string)
		groupID, _ := args[1].(string)
		artifactID, _ := args[2].(string)
		version, _ := args[3].(string)
		dest, _ := args[4].(string)
		extension := ""
		if len(args) >= 6 {
			extension, _ = args[5].(string)
		}
		r, err := sdkmaven.Download(repoURL, groupID, artifactID, version, dest, extension)
		if err != nil {
			return nil, err
		}
		return r, nil
	}
	interp.builtins["maven_artifact.resolve"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 4 {
			return nil, fmt.Errorf("maven_artifact.resolve() requires at least 4 arguments (repo_url, group_id, artifact_id, version)")
		}
		repoURL, _ := args[0].(string)
		groupID, _ := args[1].(string)
		artifactID, _ := args[2].(string)
		version, _ := args[3].(string)
		extension := ""
		if len(args) >= 5 {
			extension, _ = args[4].(string)
		}
		r, err := sdkmaven.Resolve(repoURL, groupID, artifactID, version, extension)
		if err != nil {
			return nil, err
		}
		return r, nil
	}
	interp.builtins["maven_artifact.deploy"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 5 {
			return nil, fmt.Errorf("maven_artifact.deploy() requires at least 5 arguments (repo_url, group_id, artifact_id, version, src_path)")
		}
		repoURL, _ := args[0].(string)
		groupID, _ := args[1].(string)
		artifactID, _ := args[2].(string)
		version, _ := args[3].(string)
		srcPath, _ := args[4].(string)
		extension := ""
		if len(args) >= 6 {
			extension, _ = args[5].(string)
		}
		r, err := sdkmaven.Deploy(repoURL, groupID, artifactID, version, srcPath, extension)
		if err != nil {
			return nil, err
		}
		return r, nil
	}
	interp.builtins["maven_artifact.get_latest_version"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 3 {
			return nil, fmt.Errorf("maven_artifact.get_latest_version() requires 3 arguments (repo_url, group_id, artifact_id)")
		}
		repoURL, _ := args[0].(string)
		groupID, _ := args[1].(string)
		artifactID, _ := args[2].(string)
		r, err := sdkmaven.GetLatestVersion(repoURL, groupID, artifactID)
		if err != nil {
			return nil, err
		}
		return r, nil
	}
	interp.builtins["maven_artifact.checksum"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("maven_artifact.checksum() requires 1 argument (file_path)")
		}
		filePath, _ := args[0].(string)
		r, err := sdkmaven.Checksum(filePath)
		if err != nil {
			return nil, err
		}
		return r, nil
	}

	// ── docker_image.* ──────────────────────────────────────────────────
	interp.builtins["docker_image.pull"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("docker_image.pull() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		tag := ""
		force := false
		if len(args) >= 2 {
			tag, _ = args[1].(string)
		}
		if len(args) >= 3 {
			force, _ = args[2].(bool)
		}
		r, err := sdkdockerimage.Pull(name, tag, force)
		if err != nil {
			return nil, err
		}
		return r, nil
	}
	interp.builtins["docker_image.build"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("docker_image.build() requires 2 arguments (path, name)")
		}
		path, _ := args[0].(string)
		name, _ := args[1].(string)
		tag := ""
		dockerfile := ""
		if len(args) >= 3 {
			tag, _ = args[2].(string)
		}
		if len(args) >= 4 {
			dockerfile, _ = args[3].(string)
		}
		r, err := sdkdockerimage.Build(path, name, tag, dockerfile)
		if err != nil {
			return nil, err
		}
		return r, nil
	}
	interp.builtins["docker_image.remove"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("docker_image.remove() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		tag := ""
		force := false
		if len(args) >= 2 {
			tag, _ = args[1].(string)
		}
		if len(args) >= 3 {
			force, _ = args[2].(bool)
		}
		r, err := sdkdockerimage.Remove(name, tag, force)
		if err != nil {
			return nil, err
		}
		return r, nil
	}
	interp.builtins["docker_image.tag"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("docker_image.tag() requires 2 arguments (source, target)")
		}
		source, _ := args[0].(string)
		target, _ := args[1].(string)
		r, err := sdkdockerimage.Tag(source, target)
		if err != nil {
			return nil, err
		}
		return r, nil
	}
	interp.builtins["docker_image.inspect"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("docker_image.inspect() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		r, err := sdkdockerimage.Inspect(name)
		if err != nil {
			return nil, err
		}
		return r, nil
	}
	interp.builtins["docker_image.list"] = func(args ...interface{}) (interface{}, error) {
		r, err := sdkdockerimage.List()
		if err != nil {
			return nil, err
		}
		return r, nil
	}
	interp.builtins["docker_image.push"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("docker_image.push() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		tag := ""
		if len(args) >= 2 {
			tag, _ = args[1].(string)
		}
		r, err := sdkdockerimage.Push(name, tag)
		if err != nil {
			return nil, err
		}
		return r, nil
	}

	// ── docker_container.* ──────────────────────────────────────────────
	interp.builtins["docker_container.start"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("docker_container.start() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		r, err := sdkdockercontainer.Start(name)
		if err != nil {
			return nil, err
		}
		return r, nil
	}
	interp.builtins["docker_container.stop"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("docker_container.stop() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		timeout := 0
		if len(args) >= 2 {
			timeout, _ = args[1].(int)
		}
		r, err := sdkdockercontainer.Stop(name, timeout)
		if err != nil {
			return nil, err
		}
		return r, nil
	}
	interp.builtins["docker_container.remove"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("docker_container.remove() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		force := false
		if len(args) >= 2 {
			force, _ = args[1].(bool)
		}
		r, err := sdkdockercontainer.Remove(name, force)
		if err != nil {
			return nil, err
		}
		return r, nil
	}
	interp.builtins["docker_container.restart"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("docker_container.restart() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		timeout := 0
		if len(args) >= 2 {
			timeout, _ = args[1].(int)
		}
		r, err := sdkdockercontainer.Restart(name, timeout)
		if err != nil {
			return nil, err
		}
		return r, nil
	}
	interp.builtins["docker_container.pause"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("docker_container.pause() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		r, err := sdkdockercontainer.Pause(name)
		if err != nil {
			return nil, err
		}
		return r, nil
	}
	interp.builtins["docker_container.unpause"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("docker_container.unpause() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		r, err := sdkdockercontainer.Unpause(name)
		if err != nil {
			return nil, err
		}
		return r, nil
	}
	interp.builtins["docker_container.inspect"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("docker_container.inspect() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		r, err := sdkdockercontainer.Inspect(name)
		if err != nil {
			return nil, err
		}
		return r, nil
	}
	interp.builtins["docker_container.list"] = func(args ...interface{}) (interface{}, error) {
		all := false
		if len(args) >= 1 {
			all, _ = args[0].(bool)
		}
		r, err := sdkdockercontainer.List(all)
		if err != nil {
			return nil, err
		}
		return r, nil
	}
	interp.builtins["docker_container.logs"] = func(args ...interface{}) (interface{}, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("docker_container.logs() requires 1 argument (name)")
		}
		name, _ := args[0].(string)
		tail := "100"
		if len(args) >= 2 {
			tail, _ = args[1].(string)
		}
		r, err := sdkdockercontainer.Logs(name, tail)
		if err != nil {
			return nil, err
		}
		return r, nil
	}

	// ── ping.* ────────────────────────────────────────────────────────
	interp.builtins["ping.ping"] = func(args ...interface{}) (interface{}, error) {
		data := ""
		if len(args) >= 1 {
			data, _ = args[0].(string)
		}
		return sdkping.Ping(data), nil
	}
	interp.builtins["ping.win_ping"] = func(args ...interface{}) (interface{}, error) {
		data := ""
		if len(args) >= 1 {
			data, _ = args[0].(string)
		}
		return sdkping.WinPing(data), nil
	}

	// ── find.* ────────────────────────────────────────────────────────
	interp.builtins["find.find"] = func(args ...interface{}) (interface{}, error) {
		opts := sdkfind.FindOptions{}
		if len(args) >= 1 {
			if v, ok := args[0].([]interface{}); ok {
				for _, p := range v {
					if s, ok := p.(string); ok {
						opts.Paths = append(opts.Paths, s)
					}
				}
			}
		}
		if len(args) >= 2 {
			if v, ok := args[1].([]interface{}); ok {
				for _, p := range v {
					if s, ok := p.(string); ok {
						opts.Patterns = append(opts.Patterns, s)
					}
				}
			}
		}
		if len(args) >= 3 {
			opts.FileType, _ = args[2].(string)
		}
		if len(args) >= 4 {
			opts.Recurse, _ = args[3].(bool)
		}
		if len(args) >= 5 {
			if d, ok := args[4].(float64); ok {
				opts.Depth = int(d)
			}
		}
		return sdkfind.Find(opts), nil
	}

	// ── tempfile.* ────────────────────────────────────────────────────
	interp.builtins["tempfile.create_file"] = func(args ...interface{}) (interface{}, error) {
		prefix, suffix, path := "", "", ""
		if len(args) >= 1 {
			prefix, _ = args[0].(string)
		}
		if len(args) >= 2 {
			suffix, _ = args[1].(string)
		}
		if len(args) >= 3 {
			path, _ = args[2].(string)
		}
		return sdktempfile.CreateFile(prefix, suffix, path), nil
	}
	interp.builtins["tempfile.create_dir"] = func(args ...interface{}) (interface{}, error) {
		prefix, suffix, path := "", "", ""
		if len(args) >= 1 {
			prefix, _ = args[0].(string)
		}
		if len(args) >= 2 {
			suffix, _ = args[1].(string)
		}
		if len(args) >= 3 {
			path, _ = args[2].(string)
		}
		return sdktempfile.CreateDir(prefix, suffix, path), nil
	}
	interp.builtins["tempfile.delete"] = func(args ...interface{}) (interface{}, error) {
		path := ""
		if len(args) >= 1 {
			path, _ = args[0].(string)
		}
		return sdktempfile.Delete(path), nil
	}

	// ── fail.* ────────────────────────────────────────────────────────
	interp.builtins["fail.fail"] = func(args ...interface{}) (interface{}, error) {
		msg := ""
		if len(args) >= 1 {
			msg, _ = args[0].(string)
		}
		return sdkfail.Fail(msg), nil
	}

	// ── assert.* ──────────────────────────────────────────────────────
	interp.builtins["assert.assert"] = func(args ...interface{}) (interface{}, error) {
		cond := false
		successMsg, failMsg := "", ""
		if len(args) >= 1 {
			cond, _ = args[0].(bool)
		}
		if len(args) >= 2 {
			successMsg, _ = args[1].(string)
		}
		if len(args) >= 3 {
			failMsg, _ = args[2].(string)
		}
		return sdkassert.Assert(cond, successMsg, failMsg), nil
	}

	// ── debug.* ───────────────────────────────────────────────────────
	interp.builtins["debug.debug"] = func(args ...interface{}) (interface{}, error) {
		msg := ""
		if len(args) >= 1 {
			msg, _ = args[0].(string)
		}
		return sdkdebug.Debug(msg), nil
	}
	interp.builtins["debug.debug_var"] = func(args ...interface{}) (interface{}, error) {
		name, value := "", ""
		if len(args) >= 1 {
			name, _ = args[0].(string)
		}
		if len(args) >= 2 {
			value, _ = args[1].(string)
		}
		return sdkdebug.DebugVar(name, value), nil
	}

	// ── set_fact.* ────────────────────────────────────────────────────
	interp.builtins["set_fact.set"] = func(args ...interface{}) (interface{}, error) {
		kvJSON := "{}"
		if len(args) >= 1 {
			kvJSON, _ = args[0].(string)
		}
		var kv map[string]interface{}
		if err := json.Unmarshal([]byte(kvJSON), &kv); err != nil {
			return sdksetfact.Set(map[string]interface{}{"raw": kvJSON}), nil
		}
		return sdksetfact.Set(kv), nil
	}
	interp.builtins["set_fact.get"] = func(args ...interface{}) (interface{}, error) {
		key := ""
		if len(args) >= 1 {
			key, _ = args[0].(string)
		}
		v, ok := sdksetfact.Get(key)
		return map[string]interface{}{"value": v, "found": ok}, nil
	}
	interp.builtins["set_fact.get_all"] = func(args ...interface{}) (interface{}, error) {
		return sdksetfact.GetAll(), nil
	}
	interp.builtins["set_fact.clear"] = func(args ...interface{}) (interface{}, error) {
		return sdksetfact.Clear(), nil
	}

	// ── unarchive.* ───────────────────────────────────────────────────
	interp.builtins["unarchive.unarchive"] = func(args ...interface{}) (interface{}, error) {
		src, dest, owner, group, mode, creates := "", "", "", "", "", ""
		if len(args) >= 1 {
			src, _ = args[0].(string)
		}
		if len(args) >= 2 {
			dest, _ = args[1].(string)
		}
		if len(args) >= 3 {
			owner, _ = args[2].(string)
		}
		if len(args) >= 4 {
			group, _ = args[3].(string)
		}
		if len(args) >= 5 {
			mode, _ = args[4].(string)
		}
		if len(args) >= 6 {
			creates, _ = args[5].(string)
		}
		return sdkunarchive.Unarchive(src, dest, owner, group, mode, creates), nil
	}

	// ── package_facts.* ───────────────────────────────────────────────
	interp.builtins["package_facts.collect"] = func(args ...interface{}) (interface{}, error) {
		var managers []string
		if len(args) >= 1 {
			if v, ok := args[0].([]interface{}); ok {
				for _, m := range v {
					if s, ok := m.(string); ok {
						managers = append(managers, s)
					}
				}
			}
		}
		return sdkpackagefacts.Collect(managers), nil
	}

	// ── service_facts.* ───────────────────────────────────────────────
	interp.builtins["service_facts.collect"] = func(args ...interface{}) (interface{}, error) {
		return sdkservicefacts.Collect(), nil
	}

	// ── command.* ─────────────────────────────────────────────────────
	interp.builtins["command.run"] = func(args ...interface{}) (interface{}, error) {
		var cmdArgs []string
		if len(args) >= 1 {
			if v, ok := args[0].([]interface{}); ok {
				for _, a := range v {
					if s, ok := a.(string); ok {
						cmdArgs = append(cmdArgs, s)
					}
				}
			}
		}
		chdir, creates, removes := "", "", ""
		if len(args) >= 2 {
			chdir, _ = args[1].(string)
		}
		if len(args) >= 3 {
			creates, _ = args[2].(string)
		}
		if len(args) >= 4 {
			removes, _ = args[3].(string)
		}
		timeoutMs := 0.0
		if len(args) >= 5 {
			timeoutMs, _ = args[4].(float64)
		}
		timeout := time.Duration(timeoutMs) * time.Millisecond
		return sdkcommand.Run(cmdArgs, chdir, creates, removes, timeout), nil
	}
	interp.builtins["command.shell"] = func(args ...interface{}) (interface{}, error) {
		var cmdArgs []string
		if len(args) >= 1 {
			if v, ok := args[0].([]interface{}); ok {
				for _, a := range v {
					if s, ok := a.(string); ok {
						cmdArgs = append(cmdArgs, s)
					}
				}
			}
		}
		chdir, creates, removes := "", "", ""
		if len(args) >= 2 {
			chdir, _ = args[1].(string)
		}
		if len(args) >= 3 {
			creates, _ = args[2].(string)
		}
		if len(args) >= 4 {
			removes, _ = args[3].(string)
		}
		timeoutMs := 0.0
		if len(args) >= 5 {
			timeoutMs, _ = args[4].(float64)
		}
		executable := ""
		if len(args) >= 6 {
			executable, _ = args[5].(string)
		}
		timeout := time.Duration(timeoutMs) * time.Millisecond
		return sdkcommand.Shell(cmdArgs, chdir, creates, removes, timeout, executable), nil
	}
	// ── script ────────────────────────────────────────────────────────
	interp.builtins["script.run"] = func(args ...interface{}) (interface{}, error) {
		sp := getStringArgBridge(args, 0, "")
		var scriptArgs []string
		// Re-parse: script.run(script_path, args_list, chdir, creates, removes, timeout_ms, executable)
		if len(args) >= 2 {
			if v, ok := args[1].([]interface{}); ok {
				for _, a := range v {
					if s, ok := a.(string); ok {
						scriptArgs = append(scriptArgs, s)
					}
				}
			}
		}
		chdir := getStringArgBridge(args, 2, "")
		creates := getStringArgBridge(args, 3, "")
		removes := getStringArgBridge(args, 4, "")
		timeoutMs := opsFloat(args, 5)
		executable := getStringArgBridge(args, 6, "")
		timeout := time.Duration(timeoutMs) * time.Millisecond
		return sdkscript.Run(sp, scriptArgs, chdir, creates, removes, timeout, executable), nil
	}
	// ── copy ──────────────────────────────────────────────────────────
	interp.builtins["copy.file"] = func(args ...interface{}) (interface{}, error) {
		src := getStringArgBridge(args, 0, "")
		dest := getStringArgBridge(args, 1, "")
		mode := getStringArgBridge(args, 2, "")
		owner := getStringArgBridge(args, 3, "")
		group := getStringArgBridge(args, 4, "")
		backup := false
		if len(args) >= 6 {
			backup, _ = args[5].(bool)
		}
		return sdkcopy.File(src, dest, mode, owner, group, backup), nil
	}
	interp.builtins["copy.content"] = func(args ...interface{}) (interface{}, error) {
		content := getStringArgBridge(args, 0, "")
		dest := getStringArgBridge(args, 1, "")
		mode := getStringArgBridge(args, 2, "")
		owner := getStringArgBridge(args, 3, "")
		group := getStringArgBridge(args, 4, "")
		backup := false
		if len(args) >= 6 {
			backup, _ = args[5].(bool)
		}
		return sdkcopy.Content(content, dest, mode, owner, group, backup), nil
	}
	// ── cronvar ───────────────────────────────────────────────────────
	interp.builtins["cronvar.present"] = func(args ...interface{}) (interface{}, error) {
		name := getStringArgBridge(args, 0, "")
		value := getStringArgBridge(args, 1, "")
		user := getStringArgBridge(args, 2, "")
		insertAfter := getStringArgBridge(args, 3, "")
		insertBefore := getStringArgBridge(args, 4, "")
		return sdkcronvar.Present(name, value, user, insertAfter, insertBefore), nil
	}
	interp.builtins["cronvar.absent"] = func(args ...interface{}) (interface{}, error) {
		name := getStringArgBridge(args, 0, "")
		user := getStringArgBridge(args, 1, "")
		return sdkcronvar.Absent(name, user), nil
	}
	interp.builtins["cronvar.get"] = func(args ...interface{}) (interface{}, error) {
		name := getStringArgBridge(args, 0, "")
		user := getStringArgBridge(args, 1, "")
		return sdkcronvar.Get(name, user), nil
	}
	// ── stat ──────────────────────────────────────────────────────────
	interp.builtins["stat.stat"] = func(args ...interface{}) (interface{}, error) {
		path := getStringArgBridge(args, 0, "")
		getChecksum := false
		if len(args) >= 2 {
			getChecksum, _ = args[1].(bool)
		}
		algo := getStringArgBridge(args, 2, "sha256")
		return sdkstat.Stat(path, getChecksum, algo), nil
	}
	// ── add_host ──────────────────────────────────────────────────────
	interp.builtins["add_host.add"] = func(args ...interface{}) (interface{}, error) {
		name := getStringArgBridge(args, 0, "")
		var groups []string
		if len(args) >= 2 {
			if v, ok := args[1].([]interface{}); ok {
				for _, g := range v {
					if s, ok := g.(string); ok {
						groups = append(groups, s)
					}
				}
			}
		}
		vars := map[string]string{}
		if len(args) >= 3 {
			if m, ok := args[2].(map[string]interface{}); ok {
				for k, val := range m {
					if s, ok := val.(string); ok {
						vars[k] = s
					}
				}
			}
		}
		return sdkaddhost.Add(name, groups, vars), nil
	}
	interp.builtins["add_host.get_host"] = func(args ...interface{}) (interface{}, error) {
		name := getStringArgBridge(args, 0, "")
		v, ok := sdkaddhost.GetHost(name)
		return map[string]interface{}{"vars": v, "found": ok}, nil
	}
	interp.builtins["add_host.get_group"] = func(args ...interface{}) (interface{}, error) {
		group := getStringArgBridge(args, 0, "")
		return sdkaddhost.GetGroup(group), nil
	}
	interp.builtins["add_host.list_hosts"] = func(args ...interface{}) (interface{}, error) {
		return sdkaddhost.ListHosts(), nil
	}
	interp.builtins["add_host.list_groups"] = func(args ...interface{}) (interface{}, error) {
		return sdkaddhost.ListGroups(), nil
	}
	// ── set_stats ───────────────────────────────────────────────────────
	interp.builtins["set_stats.set"] = func(args ...interface{}) (interface{}, error) {
		data := map[string]string{}
		if len(args) >= 1 {
			if m, ok := args[0].(map[string]interface{}); ok {
				for k, v := range m {
					if s, ok := v.(string); ok {
						data[k] = s
					} else {
						data[k] = fmt.Sprintf("%v", v)
					}
				}
			}
		}
		return sdksetstats.Set(data), nil
	}
	interp.builtins["set_stats.get"] = func(args ...interface{}) (interface{}, error) {
		key := getStringArgBridge(args, 0, "")
		v, ok := sdksetstats.Get(key)
		return map[string]interface{}{"value": v, "found": ok}, nil
	}
	interp.builtins["set_stats.get_all"] = func(args ...interface{}) (interface{}, error) {
		return sdksetstats.GetAll(), nil
	}
	interp.builtins["set_stats.clear"] = func(args ...interface{}) (interface{}, error) {
		sdksetstats.Clear()
		return map[string]interface{}{"cleared": true}, nil
	}
	// ── include_vars ────────────────────────────────────────────────────
	interp.builtins["include_vars.load"] = func(args ...interface{}) (interface{}, error) {
		file := getStringArgBridge(args, 0, "")
		return sdkincludevars.Load(file), nil
	}
	interp.builtins["include_vars.get"] = func(args ...interface{}) (interface{}, error) {
		key := getStringArgBridge(args, 0, "")
		v, ok := sdkincludevars.Get(key)
		return map[string]interface{}{"value": v, "found": ok}, nil
	}
	interp.builtins["include_vars.get_all"] = func(args ...interface{}) (interface{}, error) {
		return sdkincludevars.GetAll(), nil
	}
	// ── async_status ────────────────────────────────────────────────────
	interp.builtins["async_status.poll"] = func(args ...interface{}) (interface{}, error) {
		jobID := getStringArgBridge(args, 0, "")
		resultsDir := getStringArgBridge(args, 1, "")
		return sdkasyncstatus.Poll(jobID, resultsDir), nil
	}
	interp.builtins["async_status.cleanup"] = func(args ...interface{}) (interface{}, error) {
		jobID := getStringArgBridge(args, 0, "")
		resultsDir := getStringArgBridge(args, 1, "")
		return map[string]interface{}{"removed": sdkasyncstatus.Cleanup(jobID, resultsDir)}, nil
	}
	interp.builtins["async_status.wait_for"] = func(args ...interface{}) (interface{}, error) {
		jobID := getStringArgBridge(args, 0, "")
		resultsDir := getStringArgBridge(args, 1, "")
		timeoutMs := opsFloat(args, 2)
		intervalMs := opsFloat(args, 3)
		return sdkasyncstatus.WaitFor(jobID, resultsDir, time.Duration(timeoutMs)*time.Millisecond, time.Duration(intervalMs)*time.Millisecond), nil
	}
	// ── package ─────────────────────────────────────────────────────────
	interp.builtins["package.install"] = func(args ...interface{}) (interface{}, error) {
		name := getStringArgBridge(args, 0, "")
		return sdkpackagemgr.Install(name), nil
	}
	interp.builtins["package.remove"] = func(args ...interface{}) (interface{}, error) {
		name := getStringArgBridge(args, 0, "")
		return sdkpackagemgr.Remove(name), nil
	}
	interp.builtins["package.update"] = func(args ...interface{}) (interface{}, error) {
		name := getStringArgBridge(args, 0, "")
		return sdkpackagemgr.Update(name), nil
	}
	interp.builtins["package.info"] = func(args ...interface{}) (interface{}, error) {
		name := getStringArgBridge(args, 0, "")
		return sdkpackagemgr.Info(name), nil
	}
	// ── type_debug ──────────────────────────────────────────────────────
	interp.builtins["type_debug.debug"] = func(args ...interface{}) (interface{}, error) {
		var value interface{}
		if len(args) >= 1 {
			value = args[0]
		}
		return sdktypedebug.Debug(value), nil
	}

	// ── group_by ────────────────────────────────────────────────────────
	interp.builtins["group_by.group_by"] = func(args ...interface{}) (interface{}, error) {
		hostsRaw, _ := args[0].([]interface{})
		hosts := make([]string, len(hostsRaw))
		for i, h := range hostsRaw {
			hosts[i] = fmt.Sprintf("%v", h)
		}
		key := getStringArgBridge(args, 1, "")
		return sdkgroupby.GroupBy(hosts, key), nil
	}
	interp.builtins["group_by.get_hosts"] = func(args ...interface{}) (interface{}, error) {
		group := getStringArgBridge(args, 0, "")
		return sdkgroupby.GetHosts(group), nil
	}
	interp.builtins["group_by.list_groups"] = func(args ...interface{}) (interface{}, error) {
		return sdkgroupby.ListGroups(), nil
	}
	interp.builtins["group_by.clear"] = func(args ...interface{}) (interface{}, error) {
		sdkgroupby.Clear()
		return nil, nil
	}

	// ── normalize ───────────────────────────────────────────────────────
	interp.builtins["normalize.lower"] = func(args ...interface{}) (interface{}, error) {
		return sdknormalize.Lower(getStringArgBridge(args, 0, "")), nil
	}
	interp.builtins["normalize.upper"] = func(args ...interface{}) (interface{}, error) {
		return sdknormalize.Upper(getStringArgBridge(args, 0, "")), nil
	}
	interp.builtins["normalize.trim"] = func(args ...interface{}) (interface{}, error) {
		return sdknormalize.Trim(getStringArgBridge(args, 0, "")), nil
	}
	interp.builtins["normalize.slugify"] = func(args ...interface{}) (interface{}, error) {
		return sdknormalize.Slugify(getStringArgBridge(args, 0, "")), nil
	}
	interp.builtins["normalize.title"] = func(args ...interface{}) (interface{}, error) {
		return sdknormalize.Title(getStringArgBridge(args, 0, "")), nil
	}
	interp.builtins["normalize.camel_case"] = func(args ...interface{}) (interface{}, error) {
		return sdknormalize.CamelCase(getStringArgBridge(args, 0, "")), nil
	}
	interp.builtins["normalize.snake_case"] = func(args ...interface{}) (interface{}, error) {
		return sdknormalize.SnakeCase(getStringArgBridge(args, 0, "")), nil
	}
	interp.builtins["normalize.kebab_case"] = func(args ...interface{}) (interface{}, error) {
		return sdknormalize.KebabCase(getStringArgBridge(args, 0, "")), nil
	}

	// ── validate_certs ──────────────────────────────────────────────────
	interp.builtins["validate_certs.validate"] = func(args ...interface{}) (interface{}, error) {
		host := getStringArgBridge(args, 0, "")
		port := 443
		if len(args) > 1 {
			if p, ok := args[1].(float64); ok {
				port = int(p)
			}
		}
		timeoutMs := 5000
		if len(args) > 2 {
			if t, ok := args[2].(float64); ok {
				timeoutMs = int(t)
			}
		}
		return sdkvalidatecerts.Validate(host, port, time.Duration(timeoutMs)*time.Millisecond), nil
	}
	interp.builtins["validate_certs.check_expiry"] = func(args ...interface{}) (interface{}, error) {
		host := getStringArgBridge(args, 0, "")
		port := 443
		if len(args) > 1 {
			if p, ok := args[1].(float64); ok {
				port = int(p)
			}
		}
		days := 30
		if len(args) > 2 {
			if d, ok := args[2].(float64); ok {
				days = int(d)
			}
		}
		timeoutMs := 5000
		if len(args) > 3 {
			if t, ok := args[3].(float64); ok {
				timeoutMs = int(t)
			}
		}
		return sdkvalidatecerts.CheckExpiry(host, port, days, time.Duration(timeoutMs)*time.Millisecond), nil
	}

	// ── mail ──────────────────────────────────────────────────────────────
	interp.builtins["mail.send"] = func(args ...interface{}) (interface{}, error) {
		host := getStringArgBridge(args, 0, "")
		port := 587
		if len(args) > 1 {
			if p, ok := args[1].(float64); ok {
				port = int(p)
			}
		}
		from := getStringArgBridge(args, 2, "")
		var to []string
		if len(args) > 3 {
			if toRaw, ok := args[3].([]interface{}); ok {
				for _, t := range toRaw {
					to = append(to, fmt.Sprintf("%v", t))
				}
			}
		}
		subject := getStringArgBridge(args, 4, "")
		body := getStringArgBridge(args, 5, "")
		return sdkmail.SendSimple(host, port, from, to, subject, body), nil
	}
	interp.builtins["mail.send_html"] = func(args ...interface{}) (interface{}, error) {
		host := getStringArgBridge(args, 0, "")
		port := 587
		if len(args) > 1 {
			if p, ok := args[1].(float64); ok {
				port = int(p)
			}
		}
		from := getStringArgBridge(args, 2, "")
		var to []string
		if len(args) > 3 {
			if toRaw, ok := args[3].([]interface{}); ok {
				for _, t := range toRaw {
					to = append(to, fmt.Sprintf("%v", t))
				}
			}
		}
		subject := getStringArgBridge(args, 4, "")
		body := getStringArgBridge(args, 5, "")
		return sdkmail.Send(sdkmail.MailConfig{
			Host:     host,
			Port:     port,
			From:     from,
			To:       to,
			Subject:  subject,
			Body:     body,
			HTML:     true,
			StartTLS: true,
		}), nil
	}

	// ── webhook ───────────────────────────────────────────────────────────
	interp.builtins["webhook.send"] = func(args ...interface{}) (interface{}, error) {
		url := getStringArgBridge(args, 0, "")
		method := getStringArgBridge(args, 1, "POST")
		var body interface{}
		if len(args) > 2 {
			body = args[2]
		}
		return sdkwebhook.SendGeneric(url, method, nil, body), nil
	}
	interp.builtins["webhook.slack"] = func(args ...interface{}) (interface{}, error) {
		url := getStringArgBridge(args, 0, "")
		text := getStringArgBridge(args, 1, "")
		return sdkwebhook.SendSlack(url, text), nil
	}
	interp.builtins["webhook.discord"] = func(args ...interface{}) (interface{}, error) {
		url := getStringArgBridge(args, 0, "")
		content := getStringArgBridge(args, 1, "")
		return sdkwebhook.SendDiscord(url, content), nil
	}
	interp.builtins["webhook.teams"] = func(args ...interface{}) (interface{}, error) {
		url := getStringArgBridge(args, 0, "")
		title := getStringArgBridge(args, 1, "")
		text := getStringArgBridge(args, 2, "")
		return sdkwebhook.SendTeams(url, title, text), nil
	}

	// ── openssl_privatekey ────────────────────────────────────────────────
	interp.builtins["openssl_privatekey.generate"] = func(args ...interface{}) (interface{}, error) {
		path := getStringArgBridge(args, 0, "")
		keyType := getStringArgBridge(args, 1, "rsa")
		size := 0
		if len(args) > 2 {
			if s, ok := args[2].(float64); ok {
				size = int(s)
			}
		}
		return sdkopensslprivatekey.Generate(sdkopensslprivatekey.GenerateConfig{
			Path: path,
			Type: keyType,
			Size: size,
		}), nil
	}
	interp.builtins["openssl_privatekey.info"] = func(args ...interface{}) (interface{}, error) {
		path := getStringArgBridge(args, 0, "")
		return sdkopensslprivatekey.Info(path), nil
	}
	interp.builtins["openssl_privatekey.delete"] = func(args ...interface{}) (interface{}, error) {
		path := getStringArgBridge(args, 0, "")
		return sdkopensslprivatekey.Delete(path), nil
	}

	// ── ip_route ──────────────────────────────────────────────────────────
	interp.builtins["ip_route.list"] = func(args ...interface{}) (interface{}, error) {
		return sdkiproute.List(), nil
	}
	interp.builtins["ip_route.list_table"] = func(args ...interface{}) (interface{}, error) {
		table := getStringArgBridge(args, 0, "main")
		return sdkiproute.ListTable(table), nil
	}
	interp.builtins["ip_route.add"] = func(args ...interface{}) (interface{}, error) {
		destination := getStringArgBridge(args, 0, "")
		gateway := getStringArgBridge(args, 1, "")
		dev := getStringArgBridge(args, 2, "")
		metric := 0
		if len(args) > 3 {
			if m, ok := args[3].(float64); ok {
				metric = int(m)
			}
		}
		table := getStringArgBridge(args, 4, "")
		return sdkiproute.Add(sdkiproute.AddConfig{
			Destination: destination,
			Gateway:     gateway,
			Dev:         dev,
			Metric:      metric,
			Table:       table,
		}), nil
	}
	interp.builtins["ip_route.delete"] = func(args ...interface{}) (interface{}, error) {
		destination := getStringArgBridge(args, 0, "")
		table := getStringArgBridge(args, 1, "")
		return sdkiproute.Delete(destination, table), nil
	}
	interp.builtins["ip_route.flush"] = func(args ...interface{}) (interface{}, error) {
		dev := getStringArgBridge(args, 0, "")
		table := getStringArgBridge(args, 1, "")
		return sdkiproute.Flush(dev, table), nil
	}
	interp.builtins["ip_route.get"] = func(args ...interface{}) (interface{}, error) {
		destination := getStringArgBridge(args, 0, "")
		return sdkiproute.Get(destination), nil
	}
	// ip_link
	interp.builtins["ip_link.list"] = func(args ...interface{}) (interface{}, error) {
		return sdkiplink.List(), nil
	}
	interp.builtins["ip_link.get"] = func(args ...interface{}) (interface{}, error) {
		name := getStringArgBridge(args, 0, "")
		return sdkiplink.Get(name), nil
	}
	interp.builtins["ip_link.set_up"] = func(args ...interface{}) (interface{}, error) {
		name := getStringArgBridge(args, 0, "")
		return sdkiplink.SetUp(name), nil
	}
	interp.builtins["ip_link.set_down"] = func(args ...interface{}) (interface{}, error) {
		name := getStringArgBridge(args, 0, "")
		return sdkiplink.SetDown(name), nil
	}
	interp.builtins["ip_link.set_mtu"] = func(args ...interface{}) (interface{}, error) {
		name := getStringArgBridge(args, 0, "")
		mtu := 1500
		if len(args) > 1 {
			if v, ok := args[1].(int); ok {
				mtu = v
			}
		}
		return sdkiplink.SetMTU(name, mtu), nil
	}
	interp.builtins["ip_link.set_mac"] = func(args ...interface{}) (interface{}, error) {
		name := getStringArgBridge(args, 0, "")
		mac := getStringArgBridge(args, 1, "")
		return sdkiplink.SetMAC(name, mac), nil
	}
	interp.builtins["ip_link.set_name"] = func(args ...interface{}) (interface{}, error) {
		oldName := getStringArgBridge(args, 0, "")
		newName := getStringArgBridge(args, 1, "")
		return sdkiplink.SetName(oldName, newName), nil
	}
	// ip_netns
	interp.builtins["ip_netns.list"] = func(args ...interface{}) (interface{}, error) {
		return sdkipnetns.List(), nil
	}
	interp.builtins["ip_netns.get"] = func(args ...interface{}) (interface{}, error) {
		name := getStringArgBridge(args, 0, "")
		return sdkipnetns.Get(name), nil
	}
	interp.builtins["ip_netns.add"] = func(args ...interface{}) (interface{}, error) {
		name := getStringArgBridge(args, 0, "")
		return sdkipnetns.Add(name), nil
	}
	interp.builtins["ip_netns.delete"] = func(args ...interface{}) (interface{}, error) {
		name := getStringArgBridge(args, 0, "")
		return sdkipnetns.Delete(name), nil
	}
	interp.builtins["ip_netns.exec"] = func(args ...interface{}) (interface{}, error) {
		ns := getStringArgBridge(args, 0, "")
		cmd := getStringArgBridge(args, 1, "")
		var cmdArgs []string
		if len(args) > 2 {
			if strArgs, ok := args[2].([]interface{}); ok {
				for _, a := range strArgs {
					cmdArgs = append(cmdArgs, fmt.Sprintf("%v", a))
				}
			}
		}
		return sdkipnetns.Exec(ns, cmd, cmdArgs...), nil
	}
	interp.builtins["ip_netns.pids"] = func(args ...interface{}) (interface{}, error) {
		name := getStringArgBridge(args, 0, "")
		return sdkipnetns.Pids(name), nil
	}
	// ip_neighbor
	interp.builtins["ip_neighbor.list"] = func(args ...interface{}) (interface{}, error) {
		return sdkipneighbor.List(), nil
	}
	interp.builtins["ip_neighbor.list_dev"] = func(args ...interface{}) (interface{}, error) {
		dev := getStringArgBridge(args, 0, "")
		return sdkipneighbor.ListDev(dev), nil
	}
	interp.builtins["ip_neighbor.add"] = func(args ...interface{}) (interface{}, error) {
		ip := getStringArgBridge(args, 0, "")
		dev := getStringArgBridge(args, 1, "")
		mac := getStringArgBridge(args, 2, "")
		return sdkipneighbor.Add(ip, dev, mac), nil
	}
	interp.builtins["ip_neighbor.delete"] = func(args ...interface{}) (interface{}, error) {
		ip := getStringArgBridge(args, 0, "")
		dev := getStringArgBridge(args, 1, "")
		return sdkipneighbor.Delete(ip, dev), nil
	}
	interp.builtins["ip_neighbor.flush"] = func(args ...interface{}) (interface{}, error) {
		dev := getStringArgBridge(args, 0, "")
		return sdkipneighbor.Flush(dev), nil
	}
	// openssl_csr
	interp.builtins["openssl_csr.generate"] = func(args ...interface{}) (interface{}, error) {
		commonName := getStringArgBridge(args, 0, "")
		keyFile := getStringArgBridge(args, 1, "")
		outputFile := getStringArgBridge(args, 2, "")
		organization := getStringArgBridge(args, 3, "")
		orgUnit := getStringArgBridge(args, 4, "")
		country := getStringArgBridge(args, 5, "")
		state := getStringArgBridge(args, 6, "")
		locality := getStringArgBridge(args, 7, "")
		email := getStringArgBridge(args, 8, "")
		var dnsNames []string
		if len(args) > 9 {
			if list, ok := args[9].([]interface{}); ok {
				for _, v := range list {
					dnsNames = append(dnsNames, fmt.Sprintf("%v", v))
				}
			}
		}
		force := false
		if len(args) > 10 {
			if f, ok := args[10].(bool); ok {
				force = f
			}
		}
		cfg := sdkopensslcsr.CSRConfig{
			CommonName:         commonName,
			Organization:       organization,
			OrganizationalUnit: orgUnit,
			Country:            country,
			State:              state,
			Locality:           locality,
			Email:              email,
			DNSNames:           dnsNames,
			KeyFile:            keyFile,
			OutputFile:         outputFile,
			Force:              force,
		}
		return sdkopensslcsr.Generate(cfg), nil
	}
	interp.builtins["openssl_csr.info"] = func(args ...interface{}) (interface{}, error) {
		csrFile := getStringArgBridge(args, 0, "")
		return sdkopensslcsr.Info(csrFile), nil
	}
	interp.builtins["openssl_csr.delete"] = func(args ...interface{}) (interface{}, error) {
		csrFile := getStringArgBridge(args, 0, "")
		return sdkopensslcsr.Delete(csrFile), nil
	}
	// openssl_publickey
	interp.builtins["openssl_publickey.extract"] = func(args ...interface{}) (interface{}, error) {
		privKeyFile := getStringArgBridge(args, 0, "")
		outputFile := getStringArgBridge(args, 1, "")
		force := false
		if len(args) > 2 {
			if f, ok := args[2].(bool); ok {
				force = f
			}
		}
		return sdkopensslpublickey.Extract(privKeyFile, outputFile, force), nil
	}
	interp.builtins["openssl_publickey.info"] = func(args ...interface{}) (interface{}, error) {
		pubKeyFile := getStringArgBridge(args, 0, "")
		return sdkopensslpublickey.Info(pubKeyFile), nil
	}
	interp.builtins["openssl_publickey.delete"] = func(args ...interface{}) (interface{}, error) {
		pubKeyFile := getStringArgBridge(args, 0, "")
		return sdkopensslpublickey.Delete(pubKeyFile), nil
	}
	// etcd
	interp.builtins["etcd.get"] = func(args ...interface{}) (interface{}, error) {
		key := getStringArgBridge(args, 0, "")
		var endpoints []string
		if len(args) > 1 {
			if list, ok := args[1].([]interface{}); ok {
				for _, v := range list {
					endpoints = append(endpoints, fmt.Sprintf("%v", v))
				}
			}
		}
		return sdketcd.Get(key, endpoints), nil
	}
	interp.builtins["etcd.set"] = func(args ...interface{}) (interface{}, error) {
		key := getStringArgBridge(args, 0, "")
		value := getStringArgBridge(args, 1, "")
		var endpoints []string
		if len(args) > 2 {
			if list, ok := args[2].([]interface{}); ok {
				for _, v := range list {
					endpoints = append(endpoints, fmt.Sprintf("%v", v))
				}
			}
		}
		return sdketcd.Set(key, value, endpoints), nil
	}
	interp.builtins["etcd.delete"] = func(args ...interface{}) (interface{}, error) {
		key := getStringArgBridge(args, 0, "")
		var endpoints []string
		if len(args) > 1 {
			if list, ok := args[1].([]interface{}); ok {
				for _, v := range list {
					endpoints = append(endpoints, fmt.Sprintf("%v", v))
				}
			}
		}
		return sdketcd.Delete(key, endpoints), nil
	}
	interp.builtins["etcd.list"] = func(args ...interface{}) (interface{}, error) {
		prefix := getStringArgBridge(args, 0, "")
		var endpoints []string
		if len(args) > 1 {
			if list, ok := args[1].([]interface{}); ok {
				for _, v := range list {
					endpoints = append(endpoints, fmt.Sprintf("%v", v))
				}
			}
		}
		return sdketcd.List(prefix, endpoints), nil
	}
	// zookeeper
	interp.builtins["zookeeper.get"] = func(args ...interface{}) (interface{}, error) {
		path := getStringArgBridge(args, 0, "")
		var servers []string
		if len(args) > 1 {
			if list, ok := args[1].([]interface{}); ok {
				for _, v := range list {
					servers = append(servers, fmt.Sprintf("%v", v))
				}
			}
		}
		return sdkzookeeper.Get(path, servers), nil
	}
	interp.builtins["zookeeper.set"] = func(args ...interface{}) (interface{}, error) {
		path := getStringArgBridge(args, 0, "")
		value := getStringArgBridge(args, 1, "")
		var servers []string
		if len(args) > 2 {
			if list, ok := args[2].([]interface{}); ok {
				for _, v := range list {
					servers = append(servers, fmt.Sprintf("%v", v))
				}
			}
		}
		return sdkzookeeper.Set(path, value, servers), nil
	}
	interp.builtins["zookeeper.create"] = func(args ...interface{}) (interface{}, error) {
		path := getStringArgBridge(args, 0, "")
		value := getStringArgBridge(args, 1, "")
		ephemeral := false
		if len(args) > 2 {
			if e, ok := args[2].(bool); ok {
				ephemeral = e
			}
		}
		var servers []string
		if len(args) > 3 {
			if list, ok := args[3].([]interface{}); ok {
				for _, v := range list {
					servers = append(servers, fmt.Sprintf("%v", v))
				}
			}
		}
		return sdkzookeeper.Create(path, value, ephemeral, servers), nil
	}
	interp.builtins["zookeeper.delete"] = func(args ...interface{}) (interface{}, error) {
		path := getStringArgBridge(args, 0, "")
		var servers []string
		if len(args) > 1 {
			if list, ok := args[1].([]interface{}); ok {
				for _, v := range list {
					servers = append(servers, fmt.Sprintf("%v", v))
				}
			}
		}
		return sdkzookeeper.Delete(path, servers), nil
	}
	interp.builtins["zookeeper.list"] = func(args ...interface{}) (interface{}, error) {
		path := getStringArgBridge(args, 0, "")
		var servers []string
		if len(args) > 1 {
			if list, ok := args[1].([]interface{}); ok {
				for _, v := range list {
					servers = append(servers, fmt.Sprintf("%v", v))
				}
			}
		}
		return sdkzookeeper.List(path, servers), nil
	}
	interp.builtins["zookeeper.exists"] = func(args ...interface{}) (interface{}, error) {
		path := getStringArgBridge(args, 0, "")
		var servers []string
		if len(args) > 1 {
			if list, ok := args[1].([]interface{}); ok {
				for _, v := range list {
					servers = append(servers, fmt.Sprintf("%v", v))
				}
			}
		}
		return sdkzookeeper.Exists(path, servers), nil
	}
	// vault
	interp.builtins["vault.read"] = func(args ...interface{}) (interface{}, error) {
		path := getStringArgBridge(args, 0, "")
		token := getStringArgBridge(args, 1, "")
		address := getStringArgBridge(args, 2, "")
		return sdkvault.Read(path, token, address), nil
	}
	interp.builtins["vault.write"] = func(args ...interface{}) (interface{}, error) {
		path := getStringArgBridge(args, 0, "")
		token := getStringArgBridge(args, 1, "")
		address := getStringArgBridge(args, 2, "")
		data := make(map[string]interface{})
		if len(args) > 3 {
			if m, ok := args[3].(map[string]interface{}); ok {
				data = m
			}
		}
		return sdkvault.Write(path, token, address, data), nil
	}
	interp.builtins["vault.delete"] = func(args ...interface{}) (interface{}, error) {
		path := getStringArgBridge(args, 0, "")
		token := getStringArgBridge(args, 1, "")
		address := getStringArgBridge(args, 2, "")
		return sdkvault.Delete(path, token, address), nil
	}
	interp.builtins["vault.list"] = func(args ...interface{}) (interface{}, error) {
		path := getStringArgBridge(args, 0, "")
		token := getStringArgBridge(args, 1, "")
		address := getStringArgBridge(args, 2, "")
		return sdkvault.List(path, token, address), nil
	}

	// git_config
	interp.builtins["git_config.get"] = func(args ...interface{}) (interface{}, error) {
		key := getStringArgBridge(args, 0, "")
		scope := getStringArgBridge(args, 1, "")
		return sdkgitconfig.Get(key, scope)
	}
	interp.builtins["git_config.set"] = func(args ...interface{}) (interface{}, error) {
		key := getStringArgBridge(args, 0, "")
		value := getStringArgBridge(args, 1, "")
		scope := getStringArgBridge(args, 2, "")
		return sdkgitconfig.Set(key, value, scope)
	}
	interp.builtins["git_config.unset"] = func(args ...interface{}) (interface{}, error) {
		key := getStringArgBridge(args, 0, "")
		scope := getStringArgBridge(args, 1, "")
		return sdkgitconfig.Unset(key, scope)
	}
	interp.builtins["git_config.list"] = func(args ...interface{}) (interface{}, error) {
		scope := getStringArgBridge(args, 0, "")
		return sdkgitconfig.List(scope)
	}

	interp.builtins["sshd_config.get"] = func(args ...interface{}) (interface{}, error) {
		key := getStringArgBridge(args, 0, "")
		return sdksshdconfig.Get(key)
	}
	interp.builtins["sshd_config.set"] = func(args ...interface{}) (interface{}, error) {
		key := getStringArgBridge(args, 0, "")
		value := getStringArgBridge(args, 1, "")
		return sdksshdconfig.Set(key, value)
	}
	interp.builtins["sshd_config.absent"] = func(args ...interface{}) (interface{}, error) {
		key := getStringArgBridge(args, 0, "")
		return sdksshdconfig.Absent(key)
	}

	interp.builtins["docker_network.inspect"] = func(args ...interface{}) (interface{}, error) {
		name := getStringArgBridge(args, 0, "")
		return sdkdockernet.Inspect(name)
	}
	interp.builtins["docker_network.create"] = func(args ...interface{}) (interface{}, error) {
		name := getStringArgBridge(args, 0, "")
		driver := getStringArgBridge(args, 1, "bridge")
		return sdkdockernet.Create(name, driver)
	}
	interp.builtins["docker_network.remove"] = func(args ...interface{}) (interface{}, error) {
		name := getStringArgBridge(args, 0, "")
		return sdkdockernet.Remove(name)
	}
	interp.builtins["docker_network.list"] = func(args ...interface{}) (interface{}, error) {
		return sdkdockernet.List()
	}

	interp.builtins["docker_volume.inspect"] = func(args ...interface{}) (interface{}, error) {
		name := getStringArgBridge(args, 0, "")
		return sdkdockervol.Inspect(name)
	}
	interp.builtins["docker_volume.create"] = func(args ...interface{}) (interface{}, error) {
		name := getStringArgBridge(args, 0, "")
		driver := getStringArgBridge(args, 1, "local")
		return sdkdockervol.Create(name, driver)
	}
	interp.builtins["docker_volume.remove"] = func(args ...interface{}) (interface{}, error) {
		name := getStringArgBridge(args, 0, "")
		return sdkdockervol.Remove(name)
	}
	interp.builtins["docker_volume.list"] = func(args ...interface{}) (interface{}, error) {
		return sdkdockervol.List()
	}

	interp.builtins["journald.get"] = func(args ...interface{}) (interface{}, error) {
		key := getStringArgBridge(args, 0, "")
		return sdkjournald.Get(key)
	}
	interp.builtins["journald.set"] = func(args ...interface{}) (interface{}, error) {
		key := getStringArgBridge(args, 0, "")
		value := getStringArgBridge(args, 1, "")
		return sdkjournald.Set(key, value)
	}

	interp.builtins["nfs_exports.present"] = func(args ...interface{}) (interface{}, error) {
		path := getStringArgBridge(args, 0, "")
		hosts := getStringArgBridge(args, 1, "")
		options := getStringArgBridge(args, 2, "")
		return sdknfsexports.Present(path, hosts, options)
	}
	interp.builtins["nfs_exports.absent"] = func(args ...interface{}) (interface{}, error) {
		path := getStringArgBridge(args, 0, "")
		return sdknfsexports.Absent(path)
	}
	interp.builtins["nfs_exports.list"] = func(args ...interface{}) (interface{}, error) {
		return sdknfsexports.List()
	}

	interp.builtins["postfix.get"] = func(args ...interface{}) (interface{}, error) {
		key := getStringArgBridge(args, 0, "")
		return sdkpostfix.Get(key)
	}
	interp.builtins["postfix.set"] = func(args ...interface{}) (interface{}, error) {
		key := getStringArgBridge(args, 0, "")
		value := getStringArgBridge(args, 1, "")
		return sdkpostfix.Set(key, value)
	}
	interp.builtins["postfix.reload"] = func(args ...interface{}) (interface{}, error) {
		return sdkpostfix.Reload()
	}

	interp.builtins["dnsmasq.get"] = func(args ...interface{}) (interface{}, error) {
		key := getStringArgBridge(args, 0, "")
		return sdkdnsmasq.Get(key)
	}
	interp.builtins["dnsmasq.set"] = func(args ...interface{}) (interface{}, error) {
		key := getStringArgBridge(args, 0, "")
		value := getStringArgBridge(args, 1, "")
		return sdkdnsmasq.Set(key, value)
	}
	interp.builtins["dnsmasq.absent"] = func(args ...interface{}) (interface{}, error) {
		key := getStringArgBridge(args, 0, "")
		return sdkdnsmasq.Absent(key)
	}
	interp.builtins["dnsmasq.restart"] = func(args ...interface{}) (interface{}, error) {
		return sdkdnsmasq.Restart()
	}

	interp.builtins["apache2_module.check"] = func(args ...interface{}) (interface{}, error) {
		module := getStringArgBridge(args, 0, "")
		return sdkapache2mod.Check(module)
	}
	interp.builtins["apache2_module.enable"] = func(args ...interface{}) (interface{}, error) {
		module := getStringArgBridge(args, 0, "")
		return sdkapache2mod.Enable(module)
	}
	interp.builtins["apache2_module.disable"] = func(args ...interface{}) (interface{}, error) {
		module := getStringArgBridge(args, 0, "")
		return sdkapache2mod.Disable(module)
	}

	interp.builtins["pipx.install"] = func(args ...interface{}) (interface{}, error) {
		name := getStringArgBridge(args, 0, "")
		return sdkpipx.Install(name)
	}
	interp.builtins["pipx.uninstall"] = func(args ...interface{}) (interface{}, error) {
		name := getStringArgBridge(args, 0, "")
		return sdkpipx.Uninstall(name)
	}
	interp.builtins["pipx.upgrade"] = func(args ...interface{}) (interface{}, error) {
		name := getStringArgBridge(args, 0, "")
		return sdkpipx.Upgrade(name)
	}
	interp.builtins["pipx.list"] = func(args ...interface{}) (interface{}, error) {
		return sdkpipx.List()
	}

	interp.builtins["ssh_config.get"] = func(args ...interface{}) (interface{}, error) {
		host := getStringArgBridge(args, 0, "")
		option := getStringArgBridge(args, 1, "")
		scope := getStringArgBridge(args, 2, "")
		return sdksshconfig.Get(host, option, scope)
	}
	interp.builtins["ssh_config.set"] = func(args ...interface{}) (interface{}, error) {
		host := getStringArgBridge(args, 0, "")
		option := getStringArgBridge(args, 1, "")
		value := getStringArgBridge(args, 2, "")
		scope := getStringArgBridge(args, 3, "")
		return sdksshconfig.Set(host, option, value, scope)
	}
	interp.builtins["ssh_config.absent"] = func(args ...interface{}) (interface{}, error) {
		host := getStringArgBridge(args, 0, "")
		option := getStringArgBridge(args, 1, "")
		scope := getStringArgBridge(args, 2, "")
		return sdksshconfig.Absent(host, option, scope)
	}
	interp.builtins["openvpn.status"] = func(args ...interface{}) (interface{}, error) {
		return sdkopenvpn.Status()
	}
	interp.builtins["openvpn.start"] = func(args ...interface{}) (interface{}, error) {
		return sdkopenvpn.Start()
	}
	interp.builtins["openvpn.stop"] = func(args ...interface{}) (interface{}, error) {
		return sdkopenvpn.Stop()
	}
	interp.builtins["openvpn.restart"] = func(args ...interface{}) (interface{}, error) {
		return sdkopenvpn.Restart()
	}
	interp.builtins["openvpn.enable"] = func(args ...interface{}) (interface{}, error) {
		return sdkopenvpn.Enable()
	}
	interp.builtins["openvpn.disable"] = func(args ...interface{}) (interface{}, error) {
		return sdkopenvpn.Disable()
	}
	interp.builtins["openvpn.genkey"] = func(args ...interface{}) (interface{}, error) {
		outputPath := getStringArgBridge(args, 0, "")
		return sdkopenvpn.GenKey(outputPath)
	}
	interp.builtins["openvpn.gen_tls_auth"] = func(args ...interface{}) (interface{}, error) {
		outputPath := getStringArgBridge(args, 0, "")
		return sdkopenvpn.GenTLSAuth(outputPath)
	}
	interp.builtins["btrfs.filesystem_list"] = func(args ...interface{}) (interface{}, error) {
		return sdkbtrfs.FilesystemList()
	}
	interp.builtins["btrfs.subvolume_list"] = func(args ...interface{}) (interface{}, error) {
		mountPoint := getStringArgBridge(args, 0, "/")
		return sdkbtrfs.SubvolumeList(mountPoint)
	}
	interp.builtins["btrfs.subvolume_create"] = func(args ...interface{}) (interface{}, error) {
		path := getStringArgBridge(args, 0, "")
		return sdkbtrfs.SubvolumeCreate(path)
	}
	interp.builtins["btrfs.subvolume_delete"] = func(args ...interface{}) (interface{}, error) {
		path := getStringArgBridge(args, 0, "")
		return sdkbtrfs.SubvolumeDelete(path)
	}
	interp.builtins["btrfs.snapshot_create"] = func(args ...interface{}) (interface{}, error) {
		source := getStringArgBridge(args, 0, "")
		dest := getStringArgBridge(args, 1, "")
		readOnly := opsBool(args[2])
		return sdkbtrfs.SnapshotCreate(source, dest, readOnly)
	}
	interp.builtins["btrfs.scrub_start"] = func(args ...interface{}) (interface{}, error) {
		mountPoint := getStringArgBridge(args, 0, "/")
		return sdkbtrfs.ScrubStart(mountPoint)
	}
	interp.builtins["btrfs.scrub_status"] = func(args ...interface{}) (interface{}, error) {
		mountPoint := getStringArgBridge(args, 0, "/")
		return sdkbtrfs.ScrubStatus(mountPoint)
	}
	interp.builtins["btrfs.device_add"] = func(args ...interface{}) (interface{}, error) {
		devicePath := getStringArgBridge(args, 0, "")
		mountPoint := getStringArgBridge(args, 1, "")
		return sdkbtrfs.DeviceAdd(devicePath, mountPoint)
	}
	interp.builtins["btrfs.device_remove"] = func(args ...interface{}) (interface{}, error) {
		devicePath := getStringArgBridge(args, 0, "")
		mountPoint := getStringArgBridge(args, 1, "")
		return sdkbtrfs.DeviceRemove(devicePath, mountPoint)
	}
	interp.builtins["btrfs.balance_start"] = func(args ...interface{}) (interface{}, error) {
		mountPoint := getStringArgBridge(args, 0, "/")
		return sdkbtrfs.BalanceStart(mountPoint)
	}
	interp.builtins["btrfs.balance_status"] = func(args ...interface{}) (interface{}, error) {
		mountPoint := getStringArgBridge(args, 0, "/")
		return sdkbtrfs.BalanceStatus(mountPoint)
	}
	interp.builtins["btrfs.quota_enable"] = func(args ...interface{}) (interface{}, error) {
		mountPoint := getStringArgBridge(args, 0, "/")
		return sdkbtrfs.QuotaEnable(mountPoint)
	}
	interp.builtins["btrfs.quota_disable"] = func(args ...interface{}) (interface{}, error) {
		mountPoint := getStringArgBridge(args, 0, "/")
		return sdkbtrfs.QuotaDisable(mountPoint)
	}
	interp.builtins["certbot.certificates"] = func(args ...interface{}) (interface{}, error) {
		return sdkcertbot.Certificates()
	}
	interp.builtins["certbot.obtain"] = func(args ...interface{}) (interface{}, error) {
		domains := getStringSliceArgBridge(args, 0)
		email := getStringArgBridge(args, 1, "")
		webroot := getStringArgBridge(args, 2, "")
		standalone := opsBool(args[3])
		return sdkcertbot.Obtain(domains, email, webroot, standalone)
	}
	interp.builtins["certbot.renew"] = func(args ...interface{}) (interface{}, error) {
		force := opsBool(args[0])
		return sdkcertbot.Renew(force)
	}
	interp.builtins["certbot.delete"] = func(args ...interface{}) (interface{}, error) {
		domain := getStringArgBridge(args, 0, "")
		return sdkcertbot.Delete(domain)
	}
	interp.builtins["gluster.volume_list"] = func(args ...interface{}) (interface{}, error) {
		return sdkgluster.VolumeList()
	}
	interp.builtins["gluster.volume_create"] = func(args ...interface{}) (interface{}, error) {
		name := getStringArgBridge(args, 0, "")
		bricks := getStringSliceArgBridge(args, 1)
		replica := 0
		if len(args) > 2 {
			if n, ok := args[2].(float64); ok {
				replica = int(n)
			}
		}
		stripe := 0
		if len(args) > 3 {
			if n, ok := args[3].(float64); ok {
				stripe = int(n)
			}
		}
		transport := getStringArgBridge(args, 4, "")
		return sdkgluster.VolumeCreate(name, bricks, replica, stripe, transport)
	}
	interp.builtins["gluster.volume_delete"] = func(args ...interface{}) (interface{}, error) {
		name := getStringArgBridge(args, 0, "")
		return sdkgluster.VolumeDelete(name)
	}
	interp.builtins["gluster.volume_start"] = func(args ...interface{}) (interface{}, error) {
		name := getStringArgBridge(args, 0, "")
		return sdkgluster.VolumeStart(name)
	}
	interp.builtins["gluster.volume_stop"] = func(args ...interface{}) (interface{}, error) {
		name := getStringArgBridge(args, 0, "")
		return sdkgluster.VolumeStop(name)
	}
	interp.builtins["gluster.peer_list"] = func(args ...interface{}) (interface{}, error) {
		return sdkgluster.PeerList()
	}
	interp.builtins["gluster.peer_probe"] = func(args ...interface{}) (interface{}, error) {
		host := getStringArgBridge(args, 0, "")
		return sdkgluster.PeerProbe(host)
	}
	interp.builtins["gluster.peer_detach"] = func(args ...interface{}) (interface{}, error) {
		host := getStringArgBridge(args, 0, "")
		return sdkgluster.PeerDetach(host)
	}
	interp.builtins["nomad.job_list"] = func(args ...interface{}) (interface{}, error) {
		return sdknomad.JobList(getStringArgBridge(args, 0, ""))
	}
	interp.builtins["nomad.job_run"] = func(args ...interface{}) (interface{}, error) {
		return sdknomad.JobRun(getStringArgBridge(args, 0, ""), getStringArgBridge(args, 1, ""))
	}
	interp.builtins["nomad.job_stop"] = func(args ...interface{}) (interface{}, error) {
		return sdknomad.JobStop(getStringArgBridge(args, 0, ""), getStringArgBridge(args, 1, ""))
	}
	interp.builtins["nomad.alloc_list"] = func(args ...interface{}) (interface{}, error) {
		return sdknomad.AllocList(getStringArgBridge(args, 0, ""), getStringArgBridge(args, 1, ""))
	}
	interp.builtins["nomad.node_list"] = func(args ...interface{}) (interface{}, error) {
		return sdknomad.NodeList()
	}
	interp.builtins["nomad.node_drain"] = func(args ...interface{}) (interface{}, error) {
		enable := false
		if len(args) > 1 {
			enable = opsBool(args[1])
		}
		return sdknomad.NodeDrain(getStringArgBridge(args, 0, ""), enable)
	}
}
func toStringMap(args []interface{}, idx int) map[string]string {
	result := make(map[string]string)
	if idx >= len(args) {
		return result
	}
	m, ok := args[idx].(map[string]interface{})
	if !ok {
		return result
	}
	for k, v := range m {
		result[k] = fmt.Sprintf("%v", v)
	}
	return result
}

// mapStr extracts a string value from a map with a default fallback.
func mapStr(m map[string]interface{}, key string, def string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
		return fmt.Sprintf("%v", v)
	}
	return def
}

// getStringArgBridge extracts a string argument at the given index with a default.
func getStringArgBridge(args []interface{}, idx int, def string) string {
	if idx >= len(args) {
		return def
	}
	if s, ok := args[idx].(string); ok {
		return s
	}
	return def
}

// getStringSliceArgBridge extracts a []string argument at the given index.
func getStringSliceArgBridge(args []interface{}, idx int) []string {
	if idx >= len(args) {
		return nil
	}
	switch v := args[idx].(type) {
	case []string:
		return v
	case []interface{}:
		result := make([]string, 0, len(v))
		for _, item := range v {
			result = append(result, fmt.Sprintf("%v", item))
		}
		return result
	}
	return nil
}

// verifyBridgeCoverage is a self-check that every function the canonical
// opsspec table promises for the controller (interpreter) is registered.
// It panics at init if the bridge and the spec drift apart — the two used
// to disagree silently, which made docs lie.
func init() {
	registered := make(map[string]bool)
	for _, name := range SDKBuiltinNames() {
		registered[name] = true
	}
	for _, f := range opsspec.Funcs {
		if !registered[f.Name] {
			panic(fmt.Sprintf("opsspec/interpreter mismatch: %s is in the spec but not registered in the interpreter bridge", f.Name))
		}
	}
}

// opsBool extracts a bool from an interface{} value.
func opsBool(v interface{}) bool {
	if b, ok := v.(bool); ok {
		return b
	}
	return false
}

// opsFloat extracts a float64 from an interface{} value at the given index.
func opsFloat(args []interface{}, idx int) float64 {
	if idx >= len(args) {
		return 0
	}
	if f, ok := args[idx].(float64); ok {
		return f
	}
	return 0
}
