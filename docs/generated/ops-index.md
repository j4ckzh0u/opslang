# 原子操作总索引（自动生成）

> **本文档由 `make docs` 从 internal/opsspec/spec.go 自动生成，请勿手改。**
> 参数名即调用时的位置参数顺序；`可变` 表示该操作会改变系统状态，需要 admin 及以上权限。

共 1150 个原子操作，209 个包。

## 目录

- [cron](#cron)（3 个）
- [archive](#archive)（2 个）
- [apt](#apt)（13 个）
- [dnf](#dnf)（15 个）
- [apk](#apk)（10 个）
- [sysvinit](#sysvinit)（8 个）
- [runit](#runit)（8 个）
- [fail2ban](#fail2ban)（7 个）
- [lsb_release](#lsb_release)（1 个）
- [docker_compose](#docker_compose)（7 个）
- [cloud_init](#cloud_init)（4 个）
- [sys_persist](#sys_persist)（4 个）
- [wireguard](#wireguard)（8 个）
- [smartctl_notify](#smartctl_notify)（4 个）
- [dpkg_selections](#dpkg_selections)（5 个）
- [homebrew](#homebrew)（13 个）
- [apt_repo](#apt_repo)（5 个）
- [disk](#disk)（2 个）
- [docker](#docker)（8 个）
- [file](#file)（22 个）
- [firewall](#firewall)（1 个）
- [firewalld](#firewalld)（8 个）
- [git](#git)（2 个）
- [group](#group)（7 个）
- [hosts](#hosts)（4 个）
- [json](#json)（2 个）
- [known_hosts](#known_hosts)（4 个）
- [kernel](#kernel)（3 个）
- [limits](#limits)（4 个）
- [locale](#locale)（3 个）
- [logrotate](#logrotate)（4 个）
- [lvg](#lvg)（8 个）
- [net](#net)（10 个）
- [ntp](#ntp)（2 个）
- [pkg](#pkg)（6 个）
- [pip](#pip)（6 个）
- [process](#process)（5 个）
- [resolv](#resolv)（4 个）
- [service](#service)（8 个）
- [snap](#snap)（9 个）
- [flatpak](#flatpak)（8 个）
- [zfs](#zfs)（10 个）
- [nmcli](#nmcli)（10 个）
- [crypttab](#crypttab)（8 个）
- [sysfs](#sysfs)（9 个）
- [pamd](#pamd)（7 个）
- [getent](#getent)（9 个）
- [haproxy](#haproxy)（8 个）
- [openssl_cert](#openssl_cert)（6 个）
- [redis](#redis)（7 个）
- [gem](#gem)（6 个）
- [rabbitmq](#rabbitmq)（18 个）
- [consul](#consul)（10 个）
- [memcached](#memcached)（6 个）
- [selinux](#selinux)（2 个）
- [ssh](#ssh)（3 个）
- [sys](#sys)（32 个）
- [svn](#svn)（7 个）
- [sysctl](#sysctl)（3 个）
- [time](#time)（6 个）
- [user](#user)（8 个）
- [yaml](#yaml)（2 个）
- [yum_repo](#yum_repo)（4 个）
- [zypper](#zypper)（15 个）
- [ufw](#ufw)（9 个）
- [ini_file](#ini_file)（5 个）
- [mount](#mount)（6 个）
- [hostname](#hostname)（3 个）
- [timezone](#timezone)（3 个）
- [iptables](#iptables)（6 个）
- [npm](#npm)（4 个）
- [mysql](#mysql)（7 个）
- [nginx](#nginx)（5 个）
- [modprobe](#modprobe)（5 个）
- [alternatives](#alternatives)（5 个）
- [blockdev](#blockdev)（4 个）
- [at](#at)（3 个）
- [postgresql](#postgresql)（7 个）
- [apache2](#apache2)（8 个）
- [filesystem](#filesystem)（4 个）
- [parted](#parted)（4 个）
- [acl](#acl)（4 个）
- [wait_for](#wait_for)（3 个）
- [lvol](#lvol)（5 个）
- [synchronize](#synchronize)（1 个）
- [fetch](#fetch)（2 个）
- [seboolean](#seboolean)（3 个）
- [uri](#uri)（6 个）
- [lineinfile](#lineinfile)（2 个）
- [replace](#replace)（1 个）
- [xml](#xml)（2 个）
- [systemd](#systemd)（13 个）
- [patch](#patch)（2 个）
- [xattr](#xattr)（4 个）
- [firewalld_zone](#firewalld_zone)（12 个）
- [get_url](#get_url)（1 个）
- [seport](#seport)（4 个）
- [sefcontext](#sefcontext)（5 个）
- [composer](#composer)（7 个）
- [cargo](#cargo)（7 个）
- [rpmkey](#rpmkey)（3 个）
- [aptkey](#aptkey)（4 个）
- [dmidecode](#dmidecode)（5 个）
- [tuned](#tuned)（6 个）
- [supervisor](#supervisor)（8 个）
- [smartctl](#smartctl)（5 个）
- [virsh](#virsh)（9 个）
- [ethtool](#ethtool)（6 个）
- [systemd_analyze](#systemd_analyze)（5 个）
- [nvme](#nvme)（5 个）
- [lshw](#lshw)（7 个）
- [ipaddr](#ipaddr)（8 个）
- [udevadm](#udevadm)（5 个）
- [modinfo](#modinfo)（3 个）
- [dconf](#dconf)（4 个）
- [locale_gen](#locale_gen)（3 个）
- [pam_limits](#pam_limits)（2 个）
- [motd](#motd)（2 个）
- [issue](#issue)（2 个）
- [authorized_key](#authorized_key)（3 个）
- [blockinfile](#blockinfile)（2 个）
- [debconf](#debconf)（3 个）
- [reboot](#reboot)（3 个）
- [swap](#swap)（4 个）
- [raw](#raw)（2 个）
- [expect](#expect)（2 个）
- [slurp](#slurp)（2 个）
- [wait_for_connection](#wait_for_connection)（2 个）
- [firewalld_rich_rule](#firewalld_rich_rule)（4 个）
- [firewalld_ipset](#firewalld_ipset)（6 个）
- [pause](#pause)（3 个）
- [meta](#meta)（10 个）
- [uri_ext](#uri_ext)（4 个）
- [hwclock](#hwclock)（4 个）
- [mdadm](#mdadm)（6 个）
- [open_iscsi](#open_iscsi)（6 个）
- [rfkill](#rfkill)（5 个）
- [multipath](#multipath)（6 个）
- [dmsetup](#dmsetup)（7 个）
- [lvm_enhanced](#lvm_enhanced)（10 个）
- [pacman](#pacman)（11 个）
- [puppet](#puppet)（8 个）
- [yarn](#yarn)（4 个）
- [htpasswd](#htpasswd)（4 个）
- [sudoers](#sudoers)（3 个）
- [monit](#monit)（7 个）
- [kubernetes](#kubernetes)（14 个）
- [portage](#portage)（9 个）
- [pkgng](#pkgng)（9 个）
- [podman](#podman)（13 个）
- [nftables](#nftables)（16 个）
- [mongodb](#mongodb)（14 个）
- [tomcat](#tomcat)（9 个）
- [java_cert](#java_cert)（8 个）
- [maven_artifact](#maven_artifact)（5 个）
- [docker_image](#docker_image)（7 个）
- [docker_container](#docker_container)（9 个）
- [ping](#ping)（2 个）
- [find](#find)（1 个）
- [tempfile](#tempfile)（3 个）
- [fail](#fail)（1 个）
- [assert](#assert)（1 个）
- [debug](#debug)（2 个）
- [set_fact](#set_fact)（4 个）
- [unarchive](#unarchive)（1 个）
- [package_facts](#package_facts)（1 个）
- [service_facts](#service_facts)（1 个）
- [command](#command)（2 个）
- [script](#script)（1 个）
- [copy](#copy)（2 个）
- [cronvar](#cronvar)（3 个）
- [stat](#stat)（1 个）
- [add_host](#add_host)（5 个）
- [set_stats](#set_stats)（4 个）
- [include_vars](#include_vars)（3 个）
- [async_status](#async_status)（3 个）
- [package](#package)（4 个）
- [type_debug](#type_debug)（1 个）
- [group_by](#group_by)（4 个）
- [normalize](#normalize)（8 个）
- [validate_certs](#validate_certs)（2 个）
- [mail](#mail)（2 个）
- [webhook](#webhook)（4 个）
- [openssl_privatekey](#openssl_privatekey)（3 个）
- [ip_route](#ip_route)（6 个）
- [ip_link](#ip_link)（7 个）
- [ip_netns](#ip_netns)（6 个）
- [ip_neighbor](#ip_neighbor)（5 个）
- [openssl_csr](#openssl_csr)（3 个）
- [openssl_publickey](#openssl_publickey)（3 个）
- [etcd](#etcd)（4 个）
- [zookeeper](#zookeeper)（6 个）
- [vault](#vault)（4 个）
- [git_config](#git_config)（4 个）
- [sshd_config](#sshd_config)（3 个）
- [docker_network](#docker_network)（4 个）
- [docker_volume](#docker_volume)（4 个）
- [journald](#journald)（2 个）
- [nfs_exports](#nfs_exports)（3 个）
- [postfix](#postfix)（3 个）
- [dnsmasq](#dnsmasq)（4 个）
- [apache2_module](#apache2_module)（3 个）
- [pipx](#pipx)（4 个）
- [ssh_config](#ssh_config)（3 个）
- [openvpn](#openvpn)（8 个）
- [btrfs](#btrfs)（13 个）
- [certbot](#certbot)（4 个）
- [gluster](#gluster)（8 个）
- [nomad](#nomad)（6 个）

## cron

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `cron.add` | `user`, `entry` | ✓ | 全部引擎 |
| `cron.list` | `user` |  | 全部引擎 |
| `cron.remove` | `user`, `line_match` | ✓ | 全部引擎 |

## archive

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `archive.create` | `dest`, `sources` | ✓ | 全部引擎 |
| `archive.extract` | `src`, `dest` | ✓ | 全部引擎 |

## apt

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `apt.install` | `name`, `version`, `update_cache` | ✓ | 全部引擎 |
| `apt.remove` | `name`, `purge` | ✓ | 全部引擎 |
| `apt.upgrade` | `name` | ✓ | 全部引擎 |
| `apt.update_cache` | - | ✓ | 全部引擎 |
| `apt.full_upgrade` | - | ✓ | 全部引擎 |
| `apt.dist_upgrade` | - | ✓ | 全部引擎 |
| `apt.autoremove` | - | ✓ | 全部引擎 |
| `apt.clean` | - | ✓ | 全部引擎 |
| `apt.info` | `name` |  | 全部引擎 |
| `apt.list` | - |  | 全部引擎 |
| `apt.policy` | `name` |  | 全部引擎 |
| `apt.mark_auto` | `name` | ✓ | 全部引擎 |
| `apt.mark_manual` | `name` | ✓ | 全部引擎 |

## dnf

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `dnf.install` | `name`, `version` | ✓ | 全部引擎 |
| `dnf.remove` | `name` | ✓ | 全部引擎 |
| `dnf.update` | `name` | ✓ | 全部引擎 |
| `dnf.info` | `name` |  | 全部引擎 |
| `dnf.list` | - |  | 全部引擎 |
| `dnf.search` | `name` |  | 全部引擎 |
| `dnf.clean` | - | ✓ | 全部引擎 |
| `dnf.repolist` | - |  | 全部引擎 |
| `dnf.grouplist` | - |  | 全部引擎 |
| `dnf.groupinstall` | `name` | ✓ | 全部引擎 |
| `dnf.groupremove` | `name` | ✓ | 全部引擎 |
| `dnf.history` | `count` |  | 全部引擎 |
| `dnf.check_update` | - |  | 全部引擎 |
| `dnf.modulelist` | - |  | 全部引擎 |
| `dnf.module_enable` | `spec` | ✓ | 全部引擎 |

## apk

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `apk.install` | `name`, `version` | ✓ | 全部引擎 |
| `apk.remove` | `name`, `purge` | ✓ | 全部引擎 |
| `apk.update` | - | ✓ | 全部引擎 |
| `apk.upgrade` | `name` | ✓ | 全部引擎 |
| `apk.info` | `name` |  | 全部引擎 |
| `apk.list` | - |  | 全部引擎 |
| `apk.search` | `name` |  | 全部引擎 |
| `apk.cache` | - | ✓ | 全部引擎 |
| `apk.upgrade_available` | - |  | 全部引擎 |
| `apk.repository` | - |  | 全部引擎 |

## sysvinit

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `sysvinit.status` | `name` |  | 全部引擎 |
| `sysvinit.start` | `name` | ✓ | 全部引擎 |
| `sysvinit.stop` | `name` | ✓ | 全部引擎 |
| `sysvinit.restart` | `name` | ✓ | 全部引擎 |
| `sysvinit.reload` | `name` | ✓ | 全部引擎 |
| `sysvinit.enable` | `name`, `runlevels` | ✓ | 全部引擎 |
| `sysvinit.disable` | `name` | ✓ | 全部引擎 |
| `sysvinit.list` | - |  | 全部引擎 |

## runit

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `runit.status` | `service` |  | 全部引擎 |
| `runit.start` | `service` | ✓ | 全部引擎 |
| `runit.stop` | `service` | ✓ | 全部引擎 |
| `runit.restart` | `service` | ✓ | 全部引擎 |
| `runit.reload` | `service` | ✓ | 全部引擎 |
| `runit.enable` | `service`, `service_dir` | ✓ | 全部引擎 |
| `runit.disable` | `service` | ✓ | 全部引擎 |
| `runit.list` | - |  | 全部引擎 |

## fail2ban

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `fail2ban.get` | - |  | 全部引擎 |
| `fail2ban.jail_status` | `jail` |  | 全部引擎 |
| `fail2ban.start` | - | ✓ | 全部引擎 |
| `fail2ban.stop` | - | ✓ | 全部引擎 |
| `fail2ban.reload` | - | ✓ | 全部引擎 |
| `fail2ban.ban_ip` | `jail`, `ip` | ✓ | 全部引擎 |
| `fail2ban.unban_ip` | `jail`, `ip` | ✓ | 全部引擎 |

## lsb_release

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `lsb_release.get` | - |  | 全部引擎 |

## docker_compose

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `docker_compose.up` | `project_dir` | ✓ | 全部引擎 |
| `docker_compose.down` | `project_dir` | ✓ | 全部引擎 |
| `docker_compose.restart` | `project_dir` | ✓ | 全部引擎 |
| `docker_compose.pull` | `project_dir` | ✓ | 全部引擎 |
| `docker_compose.status` | `project_dir` |  | 全部引擎 |
| `docker_compose.build` | `project_dir` | ✓ | 全部引擎 |
| `docker_compose.logs` | `project_dir`, `tail` |  | 全部引擎 |

## cloud_init

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `cloud_init.status` | - |  | 全部引擎 |
| `cloud_init.modules` | - |  | 全部引擎 |
| `cloud_init.clean` | `remove_logs` | ✓ | 全部引擎 |
| `cloud_init.init` | - | ✓ | 全部引擎 |

## sys_persist

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `sys_persist.set` | `name`, `value` | ✓ | 全部引擎 |
| `sys_persist.get` | `name` |  | 全部引擎 |
| `sys_persist.remove` | `name` | ✓ | 全部引擎 |
| `sys_persist.list` | - |  | 全部引擎 |

## wireguard

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `wireguard.show` | - |  | 全部引擎 |
| `wireguard.up` | `interface`, `config_path` | ✓ | 全部引擎 |
| `wireguard.down` | `interface` | ✓ | 全部引擎 |
| `wireguard.add_peer` | `interface`, `public_key`, `allowed_ips`, `endpoint` | ✓ | 全部引擎 |
| `wireguard.remove_peer` | `interface`, `public_key` | ✓ | 全部引擎 |
| `wireguard.genkey` | - |  | 全部引擎 |
| `wireguard.genpsk` | - |  | 全部引擎 |
| `wireguard.pubkey` | `private_key` |  | 全部引擎 |

## smartctl_notify

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `smartctl_notify.check` | `device` |  | 全部引擎 |
| `smartctl_notify.list_devices` | - |  | 全部引擎 |
| `smartctl_notify.short_test` | `device` | ✓ | 全部引擎 |
| `smartctl_notify.long_test` | `device` | ✓ | 全部引擎 |

## dpkg_selections

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `dpkg_selections.set` | `name`, `state` | ✓ | 全部引擎 |
| `dpkg_selections.get` | `name` |  | 全部引擎 |
| `dpkg_selections.list` | - |  | 全部引擎 |
| `dpkg_selections.hold` | `name` | ✓ | 全部引擎 |
| `dpkg_selections.unhold` | `name` | ✓ | 全部引擎 |

## homebrew

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `homebrew.install` | `name`, `cask` | ✓ | 全部引擎 |
| `homebrew.remove` | `name`, `cask` | ✓ | 全部引擎 |
| `homebrew.upgrade` | `name` | ✓ | 全部引擎 |
| `homebrew.update` | - | ✓ | 全部引擎 |
| `homebrew.info` | `name` |  | 全部引擎 |
| `homebrew.list` | - |  | 全部引擎 |
| `homebrew.list_casks` | - |  | 全部引擎 |
| `homebrew.outdated` | - |  | 全部引擎 |
| `homebrew.clean` | - | ✓ | 全部引擎 |
| `homebrew.tap` | `name` | ✓ | 全部引擎 |
| `homebrew.untap` | `name` | ✓ | 全部引擎 |
| `homebrew.list_taps` | - |  | 全部引擎 |
| `homebrew.doctor` | - |  | 全部引擎 |

## apt_repo

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `apt_repo.list` | - |  | 全部引擎 |
| `apt_repo.exists` | `uri` |  | 全部引擎 |
| `apt_repo.add` | `uri`, `dist`, `components` | ✓ | 全部引擎 |
| `apt_repo.remove` | `uri` | ✓ | 全部引擎 |
| `apt_repo.update` | - | ✓ | 全部引擎 |

## disk

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `disk.filesystem` | `device`, `fs_type` | ✓ | 全部引擎 |
| `disk.part_list` | `device` |  | 全部引擎 |

## docker

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `docker.container_list` | `all` |  | 全部引擎 |
| `docker.container_exists` | `name` |  | 全部引擎 |
| `docker.container_run` | `name`, `image`, `opts` | ✓ | 全部引擎 |
| `docker.container_stop` | `name` | ✓ | 全部引擎 |
| `docker.container_remove` | `name`, `force` | ✓ | 全部引擎 |
| `docker.image_list` | - |  | 全部引擎 |
| `docker.image_pull` | `image` | ✓ | 全部引擎 |
| `docker.image_remove` | `image`, `force` | ✓ | 全部引擎 |

## file

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `file.append` | `path`, `content` | ✓ | 全部引擎 |
| `file.blockinfile` | `path`, `marker`, `content`, `present`, `insert_after`, `insert_before` | ✓ | 全部引擎 |
| `file.checksum` | `path`, `algo` |  | 全部引擎 |
| `file.chmod` | `path`, `mode` | ✓ | 全部引擎 |
| `file.collect` | `source`, `targets`, `options` | ✓ | 仅控制器 |
| `file.copy` | `src`, `dst` | ✓ | 全部引擎 |
| `file.delete` | `path` | ✓ | 全部引擎 |
| `file.distribute` | `source`, `targets`, `options` | ✓ | 仅控制器 |
| `file.exists` | `path` |  | 全部引擎 |
| `file.ensure` | `path`, `state`, `mode` | ✓ | 全部引擎 |
| `file.find` | `paths`, `patterns`, `regex`, `file_type`, `max_depth`, `age`, `size` |  | 全部引擎 |
| `file.ini_get` | `path`, `section`, `key` |  | 全部引擎 |
| `file.ini_set` | `path`, `section`, `key`, `value` | ✓ | 全部引擎 |
| `file.lineinfile` | `path`, `line`, `present`, `regexp` | ✓ | 全部引擎 |
| `file.list` | `dir` |  | 全部引擎 |
| `file.mkdir` | `path` | ✓ | 全部引擎 |
| `file.move` | `src`, `dst` | ✓ | 全部引擎 |
| `file.read` | `path` |  | 全部引擎 |
| `file.replace` | `path`, `pattern`, `replacement`, `after`, `before` | ✓ | 全部引擎 |
| `file.stat` | `path` |  | 全部引擎 |
| `file.template` | `path`, `vars` |  | 全部引擎 |
| `file.write` | `path`, `content` | ✓ | 全部引擎 |

## firewall

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `firewall.rule` | `action`, `protocol`, `port`, `source` | ✓ | 全部引擎 |

## firewalld

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `firewalld.get` | - |  | 全部引擎 |
| `firewalld.start` | - | ✓ | 全部引擎 |
| `firewalld.stop` | - | ✓ | 全部引擎 |
| `firewalld.restart` | - | ✓ | 全部引擎 |
| `firewalld.enable` | - | ✓ | 全部引擎 |
| `firewalld.disable` | - | ✓ | 全部引擎 |
| `firewalld.list_zones` | - |  | 全部引擎 |
| `firewalld.reload` | - | ✓ | 全部引擎 |

## git

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `git.clone` | `url`, `dest`, `opts` | ✓ | 全部引擎 |
| `git.pull` | `repo_path`, `remote`, `branch` | ✓ | 全部引擎 |

## group

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `group.add` | `name`, `opts` | ✓ | 全部引擎 |
| `group.absent` | `name` | ✓ | 全部引擎 |
| `group.ensure` | `name`, `opts` | ✓ | 全部引擎 |
| `group.exists` | `name` |  | 全部引擎 |
| `group.info` | `name` |  | 全部引擎 |
| `group.list` | - |  | 全部引擎 |
| `group.remove` | `name` | ✓ | 全部引擎 |

## hosts

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `hosts.list` | - |  | 全部引擎 |
| `hosts.exists` | `hostname` |  | 全部引擎 |
| `hosts.add` | `ip`, `hostnames` | ✓ | 全部引擎 |
| `hosts.remove` | `hostnames` | ✓ | 全部引擎 |

## json

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `json.decode` | `input` |  | 全部引擎 |
| `json.encode` | `value` |  | 全部引擎 |

## known_hosts

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `known_hosts.list` | - |  | 全部引擎 |
| `known_hosts.check` | `host` |  | 全部引擎 |
| `known_hosts.add` | `host` | ✓ | 全部引擎 |
| `known_hosts.remove` | `host` | ✓ | 全部引擎 |

## kernel

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `kernel.module_list` | - |  | 全部引擎 |
| `kernel.module_load` | `name` | ✓ | 全部引擎 |
| `kernel.module_unload` | `name` | ✓ | 全部引擎 |

## limits

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `limits.list` | - |  | 全部引擎 |
| `limits.get` | `domain` |  | 全部引擎 |
| `limits.set` | `domain`, `type`, `item`, `value` | ✓ | 全部引擎 |
| `limits.remove` | `domain` | ✓ | 全部引擎 |

## locale

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `locale.get` | - |  | 全部引擎 |
| `locale.available` | - |  | 全部引擎 |
| `locale.set` | `locale` | ✓ | 全部引擎 |

## logrotate

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `logrotate.list` | - |  | 全部引擎 |
| `logrotate.get` | `name` |  | 全部引擎 |
| `logrotate.set` | `name`, `pattern`, `frequency`, `rotate`, `compress`, `post_rotate` | ✓ | 全部引擎 |
| `logrotate.remove` | `name` | ✓ | 全部引擎 |

## lvg

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `lvg.create` | `name`, `pvs` | ✓ | 全部引擎 |
| `lvg.remove` | `name` | ✓ | 全部引擎 |
| `lvg.extend` | `name`, `pvs` | ✓ | 全部引擎 |
| `lvg.reduce` | `name`, `pvs` | ✓ | 全部引擎 |
| `lvg.activate` | `name` | ✓ | 全部引擎 |
| `lvg.deactivate` | `name` | ✓ | 全部引擎 |
| `lvg.list` | - |  | 全部引擎 |
| `lvg.get` | `name` |  | 全部引擎 |

## net

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `net.connections` | `kind` |  | 全部引擎 |
| `net.dns_lookup` | `host` |  | 全部引擎 |
| `net.download` | `url`, `dest`, `checksum_algo`, `checksum_expected` | ✓ | 全部引擎 |
| `net.http_get` | `url` |  | 全部引擎 |
| `net.http_post` | `url`, `body` | ✓ | 全部引擎 |
| `net.interfaces` | - |  | 全部引擎 |
| `net.capture` | `iface`, `seconds`, `max_packets`, `pcap_path` |  | 全部引擎 |
| `net.tcp_check` | `host`, `port` |  | 全部引擎 |
| `net.wait_for` | `host`, `port`, `timeout` |  | 全部引擎 |
| `net.wait_for_connection` | `host`, `port`, `timeout` |  | 全部引擎 |

## ntp

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `ntp.get` | - |  | 全部引擎 |
| `ntp.set` | `server` | ✓ | 全部引擎 |

## pkg

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `pkg.ensure` | `name` | ✓ | 全部引擎 |
| `pkg.info` | `name` |  | 全部引擎 |
| `pkg.install` | `name` | ✓ | 全部引擎 |
| `pkg.list` | - |  | 全部引擎 |
| `pkg.owner` | `path` |  | 全部引擎 |
| `pkg.remove` | `name` | ✓ | 全部引擎 |

## pip

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `pip.list` | - |  | 全部引擎 |
| `pip.exists` | `name` |  | 全部引擎 |
| `pip.install` | `name`, `version` | ✓ | 全部引擎 |
| `pip.uninstall` | `name` | ✓ | 全部引擎 |
| `pip.freeze` | `executable` |  | 全部引擎 |
| `pip.install_requirements` | `requirements`, `executable` | ✓ | 全部引擎 |

## process

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `process.exec` | `command`, `args` | ✓ | 全部引擎 |
| `process.find_by_name` | `name` |  | 全部引擎 |
| `process.find_by_port` | `port` |  | 全部引擎 |
| `process.kill` | `pid`, `signal` | ✓ | 全部引擎 |
| `process.list` | - |  | 全部引擎 |

## resolv

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `resolv.get` | - |  | 全部引擎 |
| `resolv.set` | `nameservers`, `search`, `options`, `domain` | ✓ | 全部引擎 |
| `resolv.add_nameserver` | `nameserver` | ✓ | 全部引擎 |
| `resolv.remove_nameserver` | `nameserver` | ✓ | 全部引擎 |

## service

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `service.disable` | `name` | ✓ | 全部引擎 |
| `service.enable` | `name` | ✓ | 全部引擎 |
| `service.ensure` | `name`, `state` | ✓ | 全部引擎 |
| `service.ensure_enabled` | `name`, `enabled` | ✓ | 全部引擎 |
| `service.restart` | `name` | ✓ | 全部引擎 |
| `service.start` | `name` | ✓ | 全部引擎 |
| `service.status` | `name` |  | 全部引擎 |
| `service.stop` | `name` | ✓ | 全部引擎 |

## snap

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `snap.install` | `name`, `channel`, `classic` | ✓ | 全部引擎 |
| `snap.remove` | `name` | ✓ | 全部引擎 |
| `snap.refresh` | `name`, `channel` | ✓ | 全部引擎 |
| `snap.list` | - |  | 全部引擎 |
| `snap.get` | `name` |  | 全部引擎 |
| `snap.enable` | `name` | ✓ | 全部引擎 |
| `snap.disable` | `name` | ✓ | 全部引擎 |
| `snap.switch` | `name`, `channel` | ✓ | 全部引擎 |
| `snap.changes` | - |  | 全部引擎 |

## flatpak

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `flatpak.install` | `name`, `from`, `user` | ✓ | 全部引擎 |
| `flatpak.remove` | `name`, `user` | ✓ | 全部引擎 |
| `flatpak.update` | `name`, `user` | ✓ | 全部引擎 |
| `flatpak.list` | `user` |  | 全部引擎 |
| `flatpak.info` | `name`, `user` |  | 全部引擎 |
| `flatpak.run` | `name`, `args`, `user` | ✓ | 全部引擎 |
| `flatpak.repair` | `user` | ✓ | 全部引擎 |
| `flatpak.add_remote` | `name`, `url` | ✓ | 全部引擎 |

## zfs

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `zfs.create` | `name`, `properties` | ✓ | 全部引擎 |
| `zfs.destroy` | `name`, `recursive` | ✓ | 全部引擎 |
| `zfs.set` | `name`, `property`, `value` | ✓ | 全部引擎 |
| `zfs.get` | `name`, `property` |  | 全部引擎 |
| `zfs.list` | - |  | 全部引擎 |
| `zfs.exists` | `name` |  | 全部引擎 |
| `zfs.list_pools` | - |  | 全部引擎 |
| `zfs.get_pool_status` | `name` |  | 全部引擎 |
| `zfs.snapshot` | `name`, `snapshot_name` | ✓ | 全部引擎 |
| `zfs.destroy_snapshot` | `name`, `snapshot_name` | ✓ | 全部引擎 |

## nmcli

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `nmcli.add` | `name`, `type`, `settings` | ✓ | 全部引擎 |
| `nmcli.modify` | `name`, `settings` | ✓ | 全部引擎 |
| `nmcli.delete` | `name` | ✓ | 全部引擎 |
| `nmcli.up` | `name` | ✓ | 全部引擎 |
| `nmcli.down` | `name` | ✓ | 全部引擎 |
| `nmcli.list` | - |  | 全部引擎 |
| `nmcli.show` | `name` |  | 全部引擎 |
| `nmcli.list_devices` | - |  | 全部引擎 |
| `nmcli.reload` | - | ✓ | 全部引擎 |
| `nmcli.get_general_status` | - |  | 全部引擎 |

## crypttab

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `crypttab.add` | `name`, `device`, `key_file`, `options` | ✓ | 全部引擎 |
| `crypttab.remove` | `name` | ✓ | 全部引擎 |
| `crypttab.modify` | `name`, `device`, `key_file`, `options` | ✓ | 全部引擎 |
| `crypttab.get` | `name` |  | 全部引擎 |
| `crypttab.list` | - |  | 全部引擎 |
| `crypttab.exists` | `name` |  | 全部引擎 |
| `crypttab.validate` | - |  | 全部引擎 |
| `crypttab.backup` | `backup_dir` | ✓ | 全部引擎 |

## sysfs

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `sysfs.read` | `path` |  | 全部引擎 |
| `sysfs.write` | `path`, `value` | ✓ | 全部引擎 |
| `sysfs.exists` | `path` |  | 全部引擎 |
| `sysfs.get` | `path` |  | 全部引擎 |
| `sysfs.list` | `dir_path` |  | 全部引擎 |
| `sysfs.set_device_power` | `device_path`, `state` | ✓ | 全部引擎 |
| `sysfs.get_device_power` | `device_path` |  | 全部引擎 |
| `sysfs.set_kernel_parameter` | `param`, `value` | ✓ | 全部引擎 |
| `sysfs.get_kernel_parameter` | `param` |  | 全部引擎 |

## pamd

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `pamd.get` | `service` |  | 全部引擎 |
| `pamd.list` | - |  | 全部引擎 |
| `pamd.add_rule` | `service`, `type`, `control`, `module`, `args` | ✓ | 全部引擎 |
| `pamd.remove_rule` | `service`, `type`, `module` | ✓ | 全部引擎 |
| `pamd.modify_rule` | `service`, `type`, `module`, `new_control`, `new_args` | ✓ | 全部引擎 |
| `pamd.validate` | `service` |  | 全部引擎 |
| `pamd.backup` | `service`, `backup_dir` |  | 全部引擎 |

## getent

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `getent.passwd` | - |  | 全部引擎 |
| `getent.lookup_user` | `key` |  | 全部引擎 |
| `getent.groups` | - |  | 全部引擎 |
| `getent.lookup_group` | `key` |  | 全部引擎 |
| `getent.services` | - |  | 全部引擎 |
| `getent.lookup_service` | `key` |  | 全部引擎 |
| `getent.protocols` | - |  | 全部引擎 |
| `getent.lookup_protocol` | `key` |  | 全部引擎 |
| `getent.shells` | - |  | 全部引擎 |

## haproxy

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `haproxy.get_status` | - |  | 全部引擎 |
| `haproxy.list_backends` | `socket` |  | 全部引擎 |
| `haproxy.enable_backend` | `backend`, `server`, `socket` | ✓ | 全部引擎 |
| `haproxy.disable_backend` | `backend`, `server`, `socket` | ✓ | 全部引擎 |
| `haproxy.validate_config` | `config_file` |  | 全部引擎 |
| `haproxy.reload` | `config_file` | ✓ | 全部引擎 |
| `haproxy.restart` | - | ✓ | 全部引擎 |
| `haproxy.version` | - |  | 全部引擎 |

## openssl_cert

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `openssl_cert.create_csr` | `key_path`, `csr_path`, `subject`, `key_bits` | ✓ | 全部引擎 |
| `openssl_cert.generate_self_signed` | `cert_path`, `key_path`, `subject`, `days`, `key_bits` | ✓ | 全部引擎 |
| `openssl_cert.inspect` | `cert_path` |  | 全部引擎 |
| `openssl_cert.verify` | `cert_path`, `ca_path` |  | 全部引擎 |
| `openssl_cert.check_expiry` | `cert_path` |  | 全部引擎 |
| `openssl_cert.convert_format` | `input_path`, `output_path`, `output_format` | ✓ | 全部引擎 |

## redis

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `redis.ping` | `host`, `port`, `auth` |  | 全部引擎 |
| `redis.get` | `key`, `host`, `port`, `auth` |  | 全部引擎 |
| `redis.set` | `key`, `value`, `host`, `port`, `auth`, `expiry_sec` | ✓ | 全部引擎 |
| `redis.del` | `keys`, `host`, `port`, `auth` | ✓ | 全部引擎 |
| `redis.keys` | `pattern`, `host`, `port`, `auth` |  | 全部引擎 |
| `redis.info` | `host`, `port`, `auth` |  | 全部引擎 |
| `redis.flush_db` | - | ✓ | 全部引擎 |

## gem

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `gem.install` | `name`, `version`, `user_install` | ✓ | 全部引擎 |
| `gem.uninstall` | `name`, `force` | ✓ | 全部引擎 |
| `gem.update` | `name` | ✓ | 全部引擎 |
| `gem.info` | `name` |  | 全部引擎 |
| `gem.list` | - |  | 全部引擎 |
| `gem.version` | - |  | 全部引擎 |

## rabbitmq

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `rabbitmq.add_vhost` | `name` | ✓ | 全部引擎 |
| `rabbitmq.delete_vhost` | `name` | ✓ | 全部引擎 |
| `rabbitmq.list_vhosts` | - |  | 全部引擎 |
| `rabbitmq.add_user` | `name`, `password`, `tags` | ✓ | 全部引擎 |
| `rabbitmq.delete_user` | `name` | ✓ | 全部引擎 |
| `rabbitmq.set_user_tags` | `name`, `tags` | ✓ | 全部引擎 |
| `rabbitmq.list_users` | - |  | 全部引擎 |
| `rabbitmq.set_permission` | `user`, `vhost`, `configure`, `write`, `read` | ✓ | 全部引擎 |
| `rabbitmq.clear_permission` | `user`, `vhost` | ✓ | 全部引擎 |
| `rabbitmq.set_policy` | `name`, `vhost`, `pattern`, `definition`, `apply_to` | ✓ | 全部引擎 |
| `rabbitmq.delete_policy` | `name`, `vhost` | ✓ | 全部引擎 |
| `rabbitmq.declare_queue` | `name`, `vhost`, `queue_type`, `durable`, `auto_delete` | ✓ | 全部引擎 |
| `rabbitmq.delete_queue` | `name`, `vhost` | ✓ | 全部引擎 |
| `rabbitmq.declare_exchange` | `name`, `vhost`, `type`, `durable`, `auto_delete` | ✓ | 全部引擎 |
| `rabbitmq.delete_exchange` | `name`, `vhost` | ✓ | 全部引擎 |
| `rabbitmq.bind_queue` | `queue`, `exchange`, `vhost`, `routing_key` | ✓ | 全部引擎 |
| `rabbitmq.unbind_queue` | `queue`, `exchange`, `vhost`, `routing_key` | ✓ | 全部引擎 |
| `rabbitmq.get_status` | - |  | 全部引擎 |

## consul

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `consul.kv_get` | `key`, `addr` |  | 全部引擎 |
| `consul.kv_put` | `key`, `value`, `addr` | ✓ | 全部引擎 |
| `consul.kv_delete` | `key`, `addr` | ✓ | 全部引擎 |
| `consul.kv_list` | `prefix`, `addr` |  | 全部引擎 |
| `consul.service_register` | `name`, `id`, `addr`, `port`, `consul_addr` | ✓ | 全部引擎 |
| `consul.service_deregister` | `id`, `consul_addr` | ✓ | 全部引擎 |
| `consul.members` | `addr` |  | 全部引擎 |
| `consul.info` | `addr` |  | 全部引擎 |
| `consul.health_check` | `service`, `addr` |  | 全部引擎 |
| `consul.version` | - |  | 全部引擎 |

## memcached

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `memcached.get` | `key`, `host`, `port` |  | 全部引擎 |
| `memcached.set` | `key`, `value`, `host`, `port`, `expiry` | ✓ | 全部引擎 |
| `memcached.delete` | `key`, `host`, `port` | ✓ | 全部引擎 |
| `memcached.flush_all` | `host`, `port` | ✓ | 全部引擎 |
| `memcached.stats` | `host`, `port` |  | 全部引擎 |
| `memcached.version` | `host`, `port` |  | 全部引擎 |

## selinux

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `selinux.get` | - |  | 全部引擎 |
| `selinux.set` | `mode` | ✓ | 全部引擎 |

## ssh

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `ssh.authorized_key_add` | `user`, `key`, `exclusive` | ✓ | 全部引擎 |
| `ssh.authorized_key_list` | `user` |  | 全部引擎 |
| `ssh.authorized_key_remove` | `user`, `key` | ✓ | 全部引擎 |

## sys

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `sys.cpu.count` | - |  | 全部引擎 |
| `sys.cpu.info` | - |  | 全部引擎 |
| `sys.cpu.usage` | - |  | 全部引擎 |
| `sys.disk.partitions` | - |  | 全部引擎 |
| `sys.disk.usage` | `path` |  | 全部引擎 |
| `sys.ethtool` | `iface` |  | 全部引擎 |
| `sys.hostname` | - |  | 全部引擎 |
| `sys.hostname_set` | `name` | ✓ | 全部引擎 |
| `sys.ip_route` | - |  | 全部引擎 |
| `sys.list_mounts` | - |  | 全部引擎 |
| `sys.load` | - |  | 全部引擎 |
| `sys.lsusb` | - |  | 全部引擎 |
| `sys.memory.info` | - |  | 全部引擎 |
| `sys.mount` | `device`, `mountpoint`, `fs_type`, `opts` | ✓ | 全部引擎 |
| `sys.net.interfaces` | - |  | 全部引擎 |
| `sys.net.all_interfaces` | - |  | 全部引擎 |
| `sys.net.primary_ip` | - |  | 全部引擎 |
| `sys.os` | - |  | 全部引擎 |
| `sys.reboot` | - | ✓ | 全部引擎 |
| `sys.timezone_get` | - |  | 全部引擎 |
| `sys.timezone_set` | `timezone` | ✓ | 全部引擎 |
| `sys.unmount` | `mountpoint` | ✓ | 全部引擎 |
| `sys.uptime` | - |  | 全部引擎 |
| `sys.virt` | - |  | 全部引擎 |
| `sys.users` | - |  | 全部引擎 |
| `sys.uuid` | - |  | 全部引擎 |
| `sys.random_password` | `length`, `use_special`, `use_numbers`, `use_uppercase` |  | 全部引擎 |
| `sys.mac_address` | `interface` |  | 全部引擎 |
| `sys.mac_addresses` | - |  | 全部引擎 |
| `sys.dmidecode` | - |  | 全部引擎 |
| `sys.lspci` | - |  | 全部引擎 |
| `sys.lsblk` | - |  | 全部引擎 |

## svn

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `svn.checkout` | `url`, `dest`, `revision`, `force` | ✓ | 全部引擎 |
| `svn.cleanup` | `dest` | ✓ | 全部引擎 |
| `svn.export` | `url`, `dest`, `revision`, `force` | ✓ | 全部引擎 |
| `svn.info` | `dest` |  | 全部引擎 |
| `svn.revert` | `dest`, `recursive` | ✓ | 全部引擎 |
| `svn.status` | `dest` |  | 全部引擎 |
| `svn.update` | `dest`, `revision` | ✓ | 全部引擎 |

## sysctl

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `sysctl.get` | `name` |  | 全部引擎 |
| `sysctl.list` | - |  | 全部引擎 |
| `sysctl.set` | `name`, `value` | ✓ | 全部引擎 |

## time

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `time.diff` | `t1`, `t2` |  | 全部引擎 |
| `time.format` | `ts`, `layout` |  | 全部引擎 |
| `time.now` | - |  | 全部引擎 |
| `time.parse` | `layout`, `value` |  | 全部引擎 |
| `time.since` | `ts` |  | 全部引擎 |
| `time.sleep` | `ms` |  | 全部引擎 |

## user

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `user.absent` | `username`, `remove_home` | ✓ | 全部引擎 |
| `user.add` | `username`, `opts` | ✓ | 全部引擎 |
| `user.ensure` | `username`, `opts` | ✓ | 全部引擎 |
| `user.exists` | `username` |  | 全部引擎 |
| `user.info` | `username` |  | 全部引擎 |
| `user.list` | - |  | 全部引擎 |
| `user.modify` | `username`, `opts` | ✓ | 全部引擎 |
| `user.remove` | `username`, `remove_home` | ✓ | 全部引擎 |

## yaml

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `yaml.decode` | `input` |  | 全部引擎 |
| `yaml.encode` | `value` |  | 全部引擎 |

## yum_repo

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `yum_repo.list` | - |  | 全部引擎 |
| `yum_repo.exists` | `id` |  | 全部引擎 |
| `yum_repo.add` | `id`, `name`, `base_url`, `gpg_check`, `gpg_key` | ✓ | 全部引擎 |
| `yum_repo.remove` | `id` | ✓ | 全部引擎 |

## zypper

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `zypper.clean` | - | ✓ | 全部引擎 |
| `zypper.dist_upgrade` | - | ✓ | 全部引擎 |
| `zypper.info` | `name` |  | 全部引擎 |
| `zypper.install` | `name`, `version` | ✓ | 全部引擎 |
| `zypper.list` | - |  | 全部引擎 |
| `zypper.patch` | - | ✓ | 全部引擎 |
| `zypper.pattern_install` | `name` | ✓ | 全部引擎 |
| `zypper.pattern_remove` | `name` | ✓ | 全部引擎 |
| `zypper.refresh` | - | ✓ | 全部引擎 |
| `zypper.remove` | `name` | ✓ | 全部引擎 |
| `zypper.repo_add` | `name`, `url` | ✓ | 全部引擎 |
| `zypper.repo_list` | - |  | 全部引擎 |
| `zypper.repo_remove` | `name` | ✓ | 全部引擎 |
| `zypper.search` | `name` |  | 全部引擎 |
| `zypper.update` | `name` | ✓ | 全部引擎 |

## ufw

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `ufw.status` | - |  | 全部引擎 |
| `ufw.list` | - |  | 全部引擎 |
| `ufw.enable` | - | ✓ | 全部引擎 |
| `ufw.disable` | - | ✓ | 全部引擎 |
| `ufw.allow` | `port`, `proto` | ✓ | 全部引擎 |
| `ufw.deny` | `port`, `proto` | ✓ | 全部引擎 |
| `ufw.delete` | `number` | ✓ | 全部引擎 |
| `ufw.reset` | - | ✓ | 全部引擎 |
| `ufw.reload` | - | ✓ | 全部引擎 |

## ini_file

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `ini_file.sections` | `path` |  | 全部引擎 |
| `ini_file.get` | `path`, `section`, `key` |  | 全部引擎 |
| `ini_file.set` | `path`, `section`, `key`, `value` | ✓ | 全部引擎 |
| `ini_file.remove` | `path`, `section`, `key` | ✓ | 全部引擎 |
| `ini_file.remove_section` | `path`, `section` | ✓ | 全部引擎 |

## mount

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `mount.list` | - |  | 全部引擎 |
| `mount.mount` | `device`, `mountpoint`, `fstype`, `options` | ✓ | 全部引擎 |
| `mount.umount` | `mountpoint` | ✓ | 全部引擎 |
| `mount.fstab` | - |  | 全部引擎 |
| `mount.add_fstab` | `device`, `mountpoint`, `fstype`, `options` | ✓ | 全部引擎 |
| `mount.remove_fstab` | `target` | ✓ | 全部引擎 |

## hostname

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `hostname.get` | - |  | 全部引擎 |
| `hostname.set` | `hostname` | ✓ | 全部引擎 |
| `hostname.set_fqdn` | `fqdn` | ✓ | 全部引擎 |

## timezone

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `timezone.get` | - |  | 全部引擎 |
| `timezone.set` | `timezone` | ✓ | 全部引擎 |
| `timezone.list` | - |  | 全部引擎 |

## iptables

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `iptables.list` | `chain` |  | 全部引擎 |
| `iptables.flush` | `table` | ✓ | 全部引擎 |
| `iptables.add_rule` | `chain`, `rule_spec` | ✓ | 全部引擎 |
| `iptables.delete_rule` | `chain`, `number` | ✓ | 全部引擎 |
| `iptables.save` | - |  | 全部引擎 |
| `iptables.list_chains` | - |  | 全部引擎 |

## npm

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `npm.list` | `global` |  | 全部引擎 |
| `npm.install` | `name`, `global` | ✓ | 全部引擎 |
| `npm.uninstall` | `name`, `global` | ✓ | 全部引擎 |
| `npm.outdated` | `global` |  | 全部引擎 |

## mysql

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `mysql.databases` | - |  | 全部引擎 |
| `mysql.create_database` | `name` | ✓ | 全部引擎 |
| `mysql.drop_database` | `name` | ✓ | 全部引擎 |
| `mysql.users` | - |  | 全部引擎 |
| `mysql.create_user` | `user`, `host`, `password` | ✓ | 全部引擎 |
| `mysql.drop_user` | `user`, `host` | ✓ | 全部引擎 |
| `mysql.grant` | `privileges`, `database`, `user`, `host` | ✓ | 全部引擎 |

## nginx

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `nginx.config_test` | - |  | 全部引擎 |
| `nginx.reload` | - | ✓ | 全部引擎 |
| `nginx.sites_list` | - |  | 全部引擎 |
| `nginx.site_enable` | `name` | ✓ | 全部引擎 |
| `nginx.site_disable` | `name` | ✓ | 全部引擎 |

## modprobe

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `modprobe.list` | - |  | 全部引擎 |
| `modprobe.load` | `name` | ✓ | 全部引擎 |
| `modprobe.unload` | `name` | ✓ | 全部引擎 |
| `modprobe.is_loaded` | `name` |  | 全部引擎 |
| `modprobe.set_boot` | `name`, `present` | ✓ | 全部引擎 |

## alternatives

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `alternatives.list` | `name` |  | 全部引擎 |
| `alternatives.display` | `name` |  | 全部引擎 |
| `alternatives.set` | `name`, `path` | ✓ | 全部引擎 |
| `alternatives.install` | `name`, `link`, `path`, `priority` | ✓ | 全部引擎 |
| `alternatives.remove` | `name`, `path` | ✓ | 全部引擎 |

## blockdev

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `blockdev.list` | - |  | 全部引擎 |
| `blockdev.info` | `device` |  | 全部引擎 |
| `blockdev.flush_buffers` | `device` | ✓ | 全部引擎 |
| `blockdev.set_readahead` | `device`, `value` | ✓ | 全部引擎 |

## at

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `at.list` | - |  | 全部引擎 |
| `at.schedule` | `command`, `time_spec` | ✓ | 全部引擎 |
| `at.remove` | `job_id` | ✓ | 全部引擎 |

## postgresql

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `postgresql.databases` | - |  | 全部引擎 |
| `postgresql.create_database` | `name` | ✓ | 全部引擎 |
| `postgresql.drop_database` | `name` | ✓ | 全部引擎 |
| `postgresql.users` | - |  | 全部引擎 |
| `postgresql.create_user` | `user`, `password` | ✓ | 全部引擎 |
| `postgresql.drop_user` | `user` | ✓ | 全部引擎 |
| `postgresql.grant` | `privileges`, `database`, `user` | ✓ | 全部引擎 |

## apache2

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `apache2.config_test` | - |  | 全部引擎 |
| `apache2.reload` | - | ✓ | 全部引擎 |
| `apache2.sites_list` | - |  | 全部引擎 |
| `apache2.site_enable` | `name` | ✓ | 全部引擎 |
| `apache2.site_disable` | `name` | ✓ | 全部引擎 |
| `apache2.modules_list` | - |  | 全部引擎 |
| `apache2.module_enable` | `name` | ✓ | 全部引擎 |
| `apache2.module_disable` | `name` | ✓ | 全部引擎 |

## filesystem

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `filesystem.mkfs` | `device`, `fstype`, `label` | ✓ | 全部引擎 |
| `filesystem.resize_ext4` | `device` | ✓ | 全部引擎 |
| `filesystem.resize_xfs` | `mountpoint` | ✓ | 全部引擎 |
| `filesystem.check` | `device` |  | 全部引擎 |

## parted

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `parted.list` | `device` |  | 全部引擎 |
| `parted.mklabel` | `device`, `label_type` | ✓ | 全部引擎 |
| `parted.mkpart` | `device`, `part_type`, `fstype`, `start`, `end` | ✓ | 全部引擎 |
| `parted.rm` | `device`, `number` | ✓ | 全部引擎 |

## acl

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `acl.get` | `path` |  | 全部引擎 |
| `acl.set` | `path`, `entry`, `recursive` | ✓ | 全部引擎 |
| `acl.remove` | `path`, `entry`, `recursive` | ✓ | 全部引擎 |
| `acl.remove_all` | `path`, `recursive` | ✓ | 全部引擎 |

## wait_for

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `wait_for.port` | `host`, `port`, `timeout_ms` |  | 全部引擎 |
| `wait_for.file` | `path`, `timeout_ms` |  | 全部引擎 |
| `wait_for.url` | `url`, `timeout_ms` |  | 全部引擎 |

## lvol

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `lvol.list` | - |  | 全部引擎 |
| `lvol.vg_list` | - |  | 全部引擎 |
| `lvol.create` | `name`, `vg`, `size` | ✓ | 全部引擎 |
| `lvol.remove` | `name`, `vg` | ✓ | 全部引擎 |
| `lvol.resize` | `name`, `vg`, `size` | ✓ | 全部引擎 |

## synchronize

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `synchronize.sync` | `source`, `dest`, `delete`, `compress` | ✓ | 全部引擎 |

## fetch

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `fetch.file` | `source`, `dest` | ✓ | 全部引擎 |
| `fetch.url` | `url`, `dest` | ✓ | 全部引擎 |

## seboolean

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `seboolean.list` | - |  | 全部引擎 |
| `seboolean.get` | `name` |  | 全部引擎 |
| `seboolean.set` | `name`, `state`, `persistent` | ✓ | 全部引擎 |

## uri

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `uri.do` | `url`, `method`, `headers`, `body`, `timeout_ms` |  | 全部引擎 |
| `uri.get` | `url` |  | 全部引擎 |
| `uri.post` | `url`, `body` | ✓ | 全部引擎 |
| `uri.put` | `url`, `body` | ✓ | 全部引擎 |
| `uri.delete` | `url` | ✓ | 全部引擎 |
| `uri.download` | `url`, `dest` | ✓ | 全部引擎 |

## lineinfile

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `lineinfile.present` | `path`, `line`, `regexp`, `create` | ✓ | 全部引擎 |
| `lineinfile.absent` | `path`, `regexp` | ✓ | 全部引擎 |

## replace

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `replace.replace` | `path`, `pattern`, `replacement`, `regexp_mode` | ✓ | 全部引擎 |

## xml

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `xml.get_element` | `path`, `element` |  | 全部引擎 |
| `xml.set_element` | `path`, `element`, `value` | ✓ | 全部引擎 |

## systemd

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `systemd.is_active` | `unit` |  | 全部引擎 |
| `systemd.is_enabled` | `unit` |  | 全部引擎 |
| `systemd.enable` | `unit` | ✓ | 全部引擎 |
| `systemd.disable` | `unit` | ✓ | 全部引擎 |
| `systemd.start` | `unit` | ✓ | 全部引擎 |
| `systemd.stop` | `unit` | ✓ | 全部引擎 |
| `systemd.restart` | `unit` | ✓ | 全部引擎 |
| `systemd.reload` | `unit` | ✓ | 全部引擎 |
| `systemd.daemon_reload` | - | ✓ | 全部引擎 |
| `systemd.mask` | `unit` | ✓ | 全部引擎 |
| `systemd.unmask` | `unit` | ✓ | 全部引擎 |
| `systemd.show` | `unit` |  | 全部引擎 |
| `systemd.list` | `unit_type` |  | 全部引擎 |

## patch

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `patch.apply` | `patch_content`, `reverse` | ✓ | 全部引擎 |
| `patch.dry_run` | `patch_content` |  | 全部引擎 |

## xattr

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `xattr.get` | `path`, `name` |  | 全部引擎 |
| `xattr.set` | `path`, `name`, `value` | ✓ | 全部引擎 |
| `xattr.remove` | `path`, `name` | ✓ | 全部引擎 |
| `xattr.list` | `path` |  | 全部引擎 |

## firewalld_zone

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `firewalld_zone.get_default` | - |  | 全部引擎 |
| `firewalld_zone.set_default` | `zone` | ✓ | 全部引擎 |
| `firewalld_zone.add_zone` | `zone` | ✓ | 全部引擎 |
| `firewalld_zone.remove_zone` | `zone` | ✓ | 全部引擎 |
| `firewalld_zone.add_service` | `zone`, `service` | ✓ | 全部引擎 |
| `firewalld_zone.remove_service` | `zone`, `service` | ✓ | 全部引擎 |
| `firewalld_zone.add_port` | `zone`, `port_protocol` | ✓ | 全部引擎 |
| `firewalld_zone.remove_port` | `zone`, `port_protocol` | ✓ | 全部引擎 |
| `firewalld_zone.add_rich_rule` | `zone`, `rule` | ✓ | 全部引擎 |
| `firewalld_zone.remove_rich_rule` | `zone`, `rule` | ✓ | 全部引擎 |
| `firewalld_zone.info` | `zone` |  | 全部引擎 |
| `firewalld_zone.list_zones` | - |  | 全部引擎 |

## get_url

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `get_url.download` | `url`, `dest`, `checksum`, `force` | ✓ | 全部引擎 |

## seport

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `seport.add` | `seport_type`, `protocol`, `port` | ✓ | 全部引擎 |
| `seport.remove` | `protocol`, `port` | ✓ | 全部引擎 |
| `seport.list` | - |  | 全部引擎 |
| `seport.get` | `protocol`, `port` |  | 全部引擎 |

## sefcontext

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `sefcontext.add` | `filespec`, `se_type` | ✓ | 全部引擎 |
| `sefcontext.modify` | `filespec`, `se_type` | ✓ | 全部引擎 |
| `sefcontext.remove` | `filespec` | ✓ | 全部引擎 |
| `sefcontext.list` | - |  | 全部引擎 |
| `sefcontext.apply` | `filespec`, `recursive` | ✓ | 全部引擎 |

## composer

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `composer.install` | `dir`, `no_dev` | ✓ | 全部引擎 |
| `composer.update` | `dir`, `no_dev` | ✓ | 全部引擎 |
| `composer.require` | `dir`, `package`, `version` | ✓ | 全部引擎 |
| `composer.remove` | `dir`, `package` | ✓ | 全部引擎 |
| `composer.create_project` | `dir`, `package`, `version` | ✓ | 全部引擎 |
| `composer.global_install` | `package`, `version` | ✓ | 全部引擎 |
| `composer.version` | - |  | 全部引擎 |

## cargo

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `cargo.install` | `package`, `version`, `force` | ✓ | 全部引擎 |
| `cargo.uninstall` | `package` | ✓ | 全部引擎 |
| `cargo.update` | `package` | ✓ | 全部引擎 |
| `cargo.list` | - |  | 全部引擎 |
| `cargo.build` | `dir`, `release` |  | 全部引擎 |
| `cargo.test` | `dir` |  | 全部引擎 |
| `cargo.version` | - |  | 全部引擎 |

## rpmkey

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `rpmkey.import` | `key_path` | ✓ | 全部引擎 |
| `rpmkey.list` | - |  | 全部引擎 |
| `rpmkey.remove` | `key_id` | ✓ | 全部引擎 |

## aptkey

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `aptkey.add` | `url`, `keyring` | ✓ | 全部引擎 |
| `aptkey.add_from_key` | `path`, `keyring` | ✓ | 全部引擎 |
| `aptkey.remove` | `key_id`, `keyring` | ✓ | 全部引擎 |
| `aptkey.list` | - |  | 全部引擎 |

## dmidecode

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `dmidecode.system` | - |  | 全部引擎 |
| `dmidecode.bios` | - |  | 全部引擎 |
| `dmidecode.chassis` | - |  | 全部引擎 |
| `dmidecode.processor` | - |  | 全部引擎 |
| `dmidecode.keyword` | `keyword` |  | 全部引擎 |

## tuned

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `tuned.set` | `profile` | ✓ | 全部引擎 |
| `tuned.status` | - |  | 全部引擎 |
| `tuned.list` | - |  | 全部引擎 |
| `tuned.off` | - | ✓ | 全部引擎 |
| `tuned.profile` | - |  | 全部引擎 |
| `tuned.verify` | - |  | 全部引擎 |

## supervisor

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `supervisor.start` | `name` | ✓ | 全部引擎 |
| `supervisor.stop` | `name` | ✓ | 全部引擎 |
| `supervisor.restart` | `name` | ✓ | 全部引擎 |
| `supervisor.reload` | - | ✓ | 全部引擎 |
| `supervisor.status` | - |  | 全部引擎 |
| `supervisor.clear_log` | `name` | ✓ | 全部引擎 |
| `supervisor.reread` | - | ✓ | 全部引擎 |
| `supervisor.update` | `name` | ✓ | 全部引擎 |

## smartctl

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `smartctl.device` | `device` |  | 全部引擎 |
| `smartctl.health` | `device` |  | 全部引擎 |
| `smartctl.attributes` | `device` |  | 全部引擎 |
| `smartctl.list` | - |  | 全部引擎 |
| `smartctl.json` | `device` |  | 全部引擎 |

## virsh

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `virsh.start` | `domain` | ✓ | 全部引擎 |
| `virsh.stop` | `domain` | ✓ | 全部引擎 |
| `virsh.reboot` | `domain` | ✓ | 全部引擎 |
| `virsh.shutdown` | `domain` | ✓ | 全部引擎 |
| `virsh.suspend` | `domain` | ✓ | 全部引擎 |
| `virsh.resume` | `domain` | ✓ | 全部引擎 |
| `virsh.list` | - |  | 全部引擎 |
| `virsh.info` | `domain` |  | 全部引擎 |
| `virsh.version` | - |  | 全部引擎 |

## ethtool

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `ethtool.show` | `interface` |  | 全部引擎 |
| `ethtool.set_speed` | `interface`, `speed` | ✓ | 全部引擎 |
| `ethtool.set_duplex` | `interface`, `duplex` | ✓ | 全部引擎 |
| `ethtool.set_autoneg` | `interface`, `autoneg` | ✓ | 全部引擎 |
| `ethtool.set_pause` | `interface`, `rx`, `tx` | ✓ | 全部引擎 |
| `ethtool.set_offload` | `interface`, `feature`, `value` | ✓ | 全部引擎 |

## systemd_analyze

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `systemd_analyze.time` | - |  | 全部引擎 |
| `systemd_analyze.blame` | - |  | 全部引擎 |
| `systemd_analyze.critical_chain` | - |  | 全部引擎 |
| `systemd_analyze.security` | `unit` |  | 全部引擎 |
| `systemd_analyze.verify` | `unit` |  | 全部引擎 |

## nvme

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `nvme.list` | - |  | 全部引擎 |
| `nvme.smart_log` | `device` |  | 全部引擎 |
| `nvme.firmware_log` | `device` |  | 全部引擎 |
| `nvme.error_log` | `device` |  | 全部引擎 |
| `nvme.version` | - |  | 全部引擎 |

## lshw

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `lshw.short` | - |  | 全部引擎 |
| `lshw.class` | `class` |  | 全部引擎 |
| `lshw.json` | - |  | 全部引擎 |
| `lshw.system` | - |  | 全部引擎 |
| `lshw.memory` | - |  | 全部引擎 |
| `lshw.disk` | - |  | 全部引擎 |
| `lshw.network` | - |  | 全部引擎 |

## ipaddr

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `ipaddr.list` | - |  | 全部引擎 |
| `ipaddr.list_interface` | `interface` |  | 全部引擎 |
| `ipaddr.add` | `address`, `interface` | ✓ | 全部引擎 |
| `ipaddr.delete` | `address`, `interface` | ✓ | 全部引擎 |
| `ipaddr.flush` | `interface` | ✓ | 全部引擎 |
| `ipaddr.links` | - |  | 全部引擎 |
| `ipaddr.link_up` | `interface` | ✓ | 全部引擎 |
| `ipaddr.link_down` | `interface` | ✓ | 全部引擎 |

## udevadm

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `udevadm.control` | `action` | ✓ | 全部引擎 |
| `udevadm.trigger` | `subsystem` | ✓ | 全部引擎 |
| `udevadm.settle` | `timeout` |  | 全部引擎 |
| `udevadm.info` | `query`, `device` |  | 全部引擎 |
| `udevadm.monitor` | - |  | 全部引擎 |

## modinfo

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `modinfo.info` | `module` |  | 全部引擎 |
| `modinfo.list` | - |  | 全部引擎 |
| `modinfo.version` | - |  | 全部引擎 |

## dconf

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `dconf.read` | `key` |  | 全部引擎 |
| `dconf.write` | `key`, `value` | ✓ | 全部引擎 |
| `dconf.list` | `dir` |  | 全部引擎 |
| `dconf.reset` | `key` | ✓ | 全部引擎 |

## locale_gen

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `locale_gen.generate` | `locale` | ✓ | 全部引擎 |
| `locale_gen.list` | - |  | 全部引擎 |
| `locale_gen.remove` | `locale` | ✓ | 全部引擎 |

## pam_limits

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `pam_limits.set` | `domain`, `type`, `item`, `value` | ✓ | 全部引擎 |
| `pam_limits.list` | - |  | 全部引擎 |

## motd

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `motd.read` | - |  | 全部引擎 |
| `motd.write` | `content` | ✓ | 全部引擎 |

## issue

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `issue.read` | - |  | 全部引擎 |
| `issue.write` | `content` | ✓ | 全部引擎 |

## authorized_key

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `authorized_key.manage` | `username`, `key`, `state`, `path` | ✓ | 全部引擎 |
| `authorized_key.list` | `username`, `path` |  | 全部引擎 |
| `authorized_key.check` | `username`, `key`, `path` |  | 全部引擎 |

## blockinfile

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `blockinfile.manage` | `path`, `block`, `state`, `marker`, `insert_after`, `insert_before` | ✓ | 全部引擎 |
| `blockinfile.read` | `path`, `marker` |  | 全部引擎 |

## debconf

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `debconf.set` | `package`, `name`, `vtype`, `value` | ✓ | 全部引擎 |
| `debconf.get` | `package`, `name` |  | 全部引擎 |
| `debconf.list` | `package` |  | 全部引擎 |

## reboot

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `reboot.request` | `msg`, `delay` | ✓ | 全部引擎 |
| `reboot.dry_run` | `msg`, `delay` |  | 全部引擎 |
| `reboot.check` | - |  | 全部引擎 |

## swap

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `swap.info` | - |  | 全部引擎 |
| `swap.create` | `path`, `size_mb` | ✓ | 全部引擎 |
| `swap.enable` | `device` | ✓ | 全部引擎 |
| `swap.disable` | `device` | ✓ | 全部引擎 |

## raw

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `raw.execute` | `command`, `timeout` | ✓ | 全部引擎 |
| `raw.execute_with_env` | `command`, `timeout`, `env` | ✓ | 全部引擎 |

## expect

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `expect.run` | `command`, `responses`, `timeout` | ✓ | 全部引擎 |
| `expect.run_simple` | `command`, `prompt`, `response`, `timeout` | ✓ | 全部引擎 |

## slurp

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `slurp.encode` | `path` |  | 全部引擎 |
| `slurp.decode` | `encoded`, `dest_path` | ✓ | 全部引擎 |

## wait_for_connection

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `wait_for_connection.wait` | `host`, `port`, `timeout`, `delay` |  | 全部引擎 |
| `wait_for_connection.check_once` | `host`, `port` |  | 全部引擎 |

## firewalld_rich_rule

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `firewalld_rich_rule.add` | `zone`, `rule` | ✓ | 全部引擎 |
| `firewalld_rich_rule.remove` | `zone`, `rule` | ✓ | 全部引擎 |
| `firewalld_rich_rule.list` | `zone` |  | 全部引擎 |
| `firewalld_rich_rule.exists` | `zone`, `rule` |  | 全部引擎 |

## firewalld_ipset

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `firewalld_ipset.create` | `name`, `type` | ✓ | 全部引擎 |
| `firewalld_ipset.delete` | `name` | ✓ | 全部引擎 |
| `firewalld_ipset.add_entry` | `name`, `entry` | ✓ | 全部引擎 |
| `firewalld_ipset.remove_entry` | `name`, `entry` | ✓ | 全部引擎 |
| `firewalld_ipset.list` | - |  | 全部引擎 |
| `firewalld_ipset.info` | `name` |  | 全部引擎 |

## pause

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `pause.seconds` | `duration` |  | 全部引擎 |
| `pause.prompt` | `message` |  | 全部引擎 |
| `pause.prompt_with_default` | `message`, `default` |  | 全部引擎 |

## meta

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `meta.end_host` | - |  | 全部引擎 |
| `meta.end_play` | - |  | 全部引擎 |
| `meta.clear_host_errors` | - |  | 全部引擎 |
| `meta.refresh_inventory` | - |  | 全部引擎 |
| `meta.flush_handlers` | - |  | 全部引擎 |
| `meta.reset_connection` | - |  | 全部引擎 |
| `meta.noop` | - |  | 全部引擎 |
| `meta.fail` | `message` |  | 全部引擎 |
| `meta.assert` | `condition`, `message` |  | 全部引擎 |
| `meta.debug` | `message`, `vars` |  | 全部引擎 |

## uri_ext

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `uri_ext.patch` | `url`, `body`, `headers`, `timeout` | ✓ | 全部引擎 |
| `uri_ext.delete` | `url`, `headers`, `timeout` | ✓ | 全部引擎 |
| `uri_ext.head` | `url`, `headers`, `timeout` |  | 全部引擎 |
| `uri_ext.options` | `url`, `headers`, `timeout` |  | 全部引擎 |

## hwclock

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `hwclock.get` | - |  | 全部引擎 |
| `hwclock.set` | - | ✓ | 全部引擎 |
| `hwclock.hctosys` | - | ✓ | 全部引擎 |
| `hwclock.set_time` | `time` | ✓ | 全部引擎 |

## mdadm

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `mdadm.create` | `device`, `level`, `devices` | ✓ | 全部引擎 |
| `mdadm.destroy` | `device` | ✓ | 全部引擎 |
| `mdadm.detail` | `device` |  | 全部引擎 |
| `mdadm.scan` | - |  | 全部引擎 |
| `mdadm.add` | `device`, `member` | ✓ | 全部引擎 |
| `mdadm.remove` | `device`, `member` | ✓ | 全部引擎 |

## open_iscsi

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `open_iscsi.discover` | `portal`, `port` |  | 全部引擎 |
| `open_iscsi.login` | `target`, `portal` | ✓ | 全部引擎 |
| `open_iscsi.logout` | `target`, `portal` | ✓ | 全部引擎 |
| `open_iscsi.list_sessions` | - |  | 全部引擎 |
| `open_iscsi.list_nodes` | - |  | 全部引擎 |
| `open_iscsi.set_startup` | `target`, `portal`, `startup` | ✓ | 全部引擎 |

## rfkill

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `rfkill.list` | - |  | 全部引擎 |
| `rfkill.block` | `device` | ✓ | 全部引擎 |
| `rfkill.unblock` | `device` | ✓ | 全部引擎 |
| `rfkill.block_all` | `type` | ✓ | 全部引擎 |
| `rfkill.unblock_all` | `type` | ✓ | 全部引擎 |

## multipath

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `multipath.reconfigure` | - | ✓ | 全部引擎 |
| `multipath.list_paths` | - |  | 全部引擎 |
| `multipath.list_maps` | - |  | 全部引擎 |
| `multipath.add_map` | `device` | ✓ | 全部引擎 |
| `multipath.remove_map` | `device` | ✓ | 全部引擎 |
| `multipath.flush` | - | ✓ | 全部引擎 |

## dmsetup

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `dmsetup.create` | `name`, `table` | ✓ | 全部引擎 |
| `dmsetup.remove` | `name` | ✓ | 全部引擎 |
| `dmsetup.remove_all` | - | ✓ | 全部引擎 |
| `dmsetup.list` | - |  | 全部引擎 |
| `dmsetup.info` | `name` |  | 全部引擎 |
| `dmsetup.suspend` | `name` | ✓ | 全部引擎 |
| `dmsetup.resume` | `name` | ✓ | 全部引擎 |

## lvm_enhanced

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `lvm_enhanced.pv_create` | `device` | ✓ | 全部引擎 |
| `lvm_enhanced.pv_remove` | `device`, `force` | ✓ | 全部引擎 |
| `lvm_enhanced.pv_list` | - |  | 全部引擎 |
| `lvm_enhanced.vg_create` | `name`, `devices` | ✓ | 全部引擎 |
| `lvm_enhanced.vg_remove` | `name`, `force` | ✓ | 全部引擎 |
| `lvm_enhanced.vg_extend` | `vg_name`, `device` | ✓ | 全部引擎 |
| `lvm_enhanced.vg_list` | - |  | 全部引擎 |
| `lvm_enhanced.lv_extend` | `lv_path`, `size` | ✓ | 全部引擎 |
| `lvm_enhanced.lv_extend_all` | `lv_path` | ✓ | 全部引擎 |
| `lvm_enhanced.lv_list` | - |  | 全部引擎 |

## pacman

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `pacman.clean` | - | ✓ | 全部引擎 |
| `pacman.info` | `name` |  | 全部引擎 |
| `pacman.install` | `name` | ✓ | 全部引擎 |
| `pacman.install_file` | `path` | ✓ | 全部引擎 |
| `pacman.list` | - |  | 全部引擎 |
| `pacman.remove` | `name`, `cascade` | ✓ | 全部引擎 |
| `pacman.remove_orphans` | - | ✓ | 全部引擎 |
| `pacman.search` | `name` |  | 全部引擎 |
| `pacman.update` | `name` | ✓ | 全部引擎 |
| `pacman.update_database` | - | ✓ | 全部引擎 |
| `pacman.upgrade` | - | ✓ | 全部引擎 |

## puppet

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `puppet.run` | `environment`, `tags` | ✓ | 全部引擎 |
| `puppet.run_noop` | `environment`, `tags` |  | 全部引擎 |
| `puppet.status` | - |  | 全部引擎 |
| `puppet.disable` | `message` | ✓ | 全部引擎 |
| `puppet.enable` | - | ✓ | 全部引擎 |
| `puppet.fact` | `name` |  | 全部引擎 |
| `puppet.module_list` | - |  | 全部引擎 |
| `puppet.module_install` | `name`, `version` | ✓ | 全部引擎 |

## yarn

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `yarn.install` | `name`, `version`, `global` | ✓ | 全部引擎 |
| `yarn.remove` | `name`, `global` | ✓ | 全部引擎 |
| `yarn.global` | `directory` | ✓ | 全部引擎 |
| `yarn.list` | `global` |  | 全部引擎 |

## htpasswd

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `htpasswd.set` | `path`, `username`, `password`, `create` | ✓ | 全部引擎 |
| `htpasswd.remove` | `path`, `username` | ✓ | 全部引擎 |
| `htpasswd.info` | `path` |  | 全部引擎 |
| `htpasswd.hash_sha1` | `password` |  | 全部引擎 |

## sudoers

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `sudoers.set` | `name`, `user`, `commands`, `nopasswd`, `sudoers_dir` | ✓ | 全部引擎 |
| `sudoers.remove` | `name`, `sudoers_dir` | ✓ | 全部引擎 |
| `sudoers.info` | `name`, `sudoers_dir` |  | 全部引擎 |

## monit

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `monit.start` | `service` | ✓ | 全部引擎 |
| `monit.stop` | `service` | ✓ | 全部引擎 |
| `monit.monitor` | `service` | ✓ | 全部引擎 |
| `monit.unmonitor` | `service` | ✓ | 全部引擎 |
| `monit.restart` | `service` | ✓ | 全部引擎 |
| `monit.status` | - |  | 全部引擎 |
| `monit.reload` | - | ✓ | 全部引擎 |

## kubernetes

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `kubernetes.apply` | `manifest`, `namespace`, `dry_run` | ✓ | 全部引擎 |
| `kubernetes.delete` | `manifest`, `namespace` | ✓ | 全部引擎 |
| `kubernetes.get` | `resource_type`, `name`, `namespace` |  | 全部引擎 |
| `kubernetes.list` | `resource_type`, `namespace`, `labels` |  | 全部引擎 |
| `kubernetes.create_namespace` | `name` | ✓ | 全部引擎 |
| `kubernetes.delete_namespace` | `name` | ✓ | 全部引擎 |
| `kubernetes.get_pods` | `namespace`, `labels` |  | 全部引擎 |
| `kubernetes.get_services` | `namespace` |  | 全部引擎 |
| `kubernetes.get_deployments` | `namespace` |  | 全部引擎 |
| `kubernetes.scale` | `deployment`, `replicas`, `namespace` | ✓ | 全部引擎 |
| `kubernetes.rollout_status` | `deployment`, `namespace` |  | 全部引擎 |
| `kubernetes.exec` | `pod`, `command`, `namespace`, `container` | ✓ | 全部引擎 |
| `kubernetes.logs` | `pod`, `namespace`, `container`, `tail` |  | 全部引擎 |
| `kubernetes.wait_ready` | `resource_type`, `name`, `namespace`, `timeout` |  | 全部引擎 |

## portage

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `portage.install` | `name`, `version` | ✓ | 全部引擎 |
| `portage.remove` | `name` | ✓ | 全部引擎 |
| `portage.update` | `name`, `deep` | ✓ | 全部引擎 |
| `portage.sync` | - | ✓ | 全部引擎 |
| `portage.info` | `name` |  | 全部引擎 |
| `portage.list` | - |  | 全部引擎 |
| `portage.search` | `name` |  | 全部引擎 |
| `portage.depclean` | - | ✓ | 全部引擎 |
| `portage.metadata` | `name` |  | 全部引擎 |

## pkgng

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `pkgng.install` | `name`, `version` | ✓ | 全部引擎 |
| `pkgng.remove` | `name` | ✓ | 全部引擎 |
| `pkgng.update` | - | ✓ | 全部引擎 |
| `pkgng.upgrade` | `name` | ✓ | 全部引擎 |
| `pkgng.autoclean` | - | ✓ | 全部引擎 |
| `pkgng.info` | `name` |  | 全部引擎 |
| `pkgng.list` | - |  | 全部引擎 |
| `pkgng.search` | `name` |  | 全部引擎 |
| `pkgng.stats` | - |  | 全部引擎 |

## podman

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `podman.run` | `image`, `name`, `command` | ✓ | 全部引擎 |
| `podman.stop` | `name`, `timeout` | ✓ | 全部引擎 |
| `podman.start` | `name` | ✓ | 全部引擎 |
| `podman.remove` | `name`, `force` | ✓ | 全部引擎 |
| `podman.list_containers` | `all` |  | 全部引擎 |
| `podman.inspect` | `name` |  | 全部引擎 |
| `podman.pull` | `image` | ✓ | 全部引擎 |
| `podman.list_images` | - |  | 全部引擎 |
| `podman.remove_image` | `image_id`, `force` | ✓ | 全部引擎 |
| `podman.create_pod` | `name` | ✓ | 全部引擎 |
| `podman.stop_pod` | `name` | ✓ | 全部引擎 |
| `podman.remove_pod` | `name`, `force` | ✓ | 全部引擎 |
| `podman.list_pods` | - |  | 全部引擎 |

## nftables

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `nftables.add_table` | `family`, `name` | ✓ | 全部引擎 |
| `nftables.delete_table` | `family`, `name` | ✓ | 全部引擎 |
| `nftables.list_tables` | - |  | 全部引擎 |
| `nftables.add_chain` | `family`, `table`, `name`, `type`, `hook`, `priority` | ✓ | 全部引擎 |
| `nftables.delete_chain` | `family`, `table`, `name` | ✓ | 全部引擎 |
| `nftables.add_rule` | `family`, `table`, `chain`, `expression` | ✓ | 全部引擎 |
| `nftables.delete_rule` | `family`, `table`, `chain`, `handle` | ✓ | 全部引擎 |
| `nftables.flush_chain` | `family`, `table`, `chain` | ✓ | 全部引擎 |
| `nftables.flush_table` | `family`, `table` | ✓ | 全部引擎 |
| `nftables.flush_ruleset` | - | ✓ | 全部引擎 |
| `nftables.list_ruleset` | - |  | 全部引擎 |
| `nftables.add_set` | `family`, `table`, `name`, `type`, `flags` | ✓ | 全部引擎 |
| `nftables.delete_set` | `family`, `table`, `name` | ✓ | 全部引擎 |
| `nftables.add_element` | `family`, `table`, `set`, `element` | ✓ | 全部引擎 |
| `nftables.delete_element` | `family`, `table`, `set`, `element` | ✓ | 全部引擎 |
| `nftables.export` | `format` |  | 全部引擎 |

## mongodb

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `mongodb.create_database` | `host`, `port`, `name` | ✓ | 全部引擎 |
| `mongodb.drop_database` | `host`, `port`, `name` | ✓ | 全部引擎 |
| `mongodb.list_databases` | `host`, `port` |  | 全部引擎 |
| `mongodb.create_user` | `host`, `port`, `database`, `user`, `password`, `roles` | ✓ | 全部引擎 |
| `mongodb.drop_user` | `host`, `port`, `database`, `user` | ✓ | 全部引擎 |
| `mongodb.list_users` | `host`, `port`, `database` |  | 全部引擎 |
| `mongodb.create_collection` | `host`, `port`, `database`, `collection` | ✓ | 全部引擎 |
| `mongodb.drop_collection` | `host`, `port`, `database`, `collection` | ✓ | 全部引擎 |
| `mongodb.list_collections` | `host`, `port`, `database` |  | 全部引擎 |
| `mongodb.create_index` | `host`, `port`, `database`, `collection`, `keys`, `unique`, `name` | ✓ | 全部引擎 |
| `mongodb.drop_index` | `host`, `port`, `database`, `collection`, `index_name` | ✓ | 全部引擎 |
| `mongodb.list_indexes` | `host`, `port`, `database`, `collection` |  | 全部引擎 |
| `mongodb.server_status` | `host`, `port` |  | 全部引擎 |
| `mongodb.replica_set_status` | `host`, `port` |  | 全部引擎 |

## tomcat

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `tomcat.start` | `catalina_home` | ✓ | 全部引擎 |
| `tomcat.stop` | `catalina_home` | ✓ | 全部引擎 |
| `tomcat.restart` | `catalina_home` | ✓ | 全部引擎 |
| `tomcat.status` | `catalina_home` |  | 全部引擎 |
| `tomcat.deploy` | `catalina_home`, `war_path`, `context_path` | ✓ | 全部引擎 |
| `tomcat.undeploy` | `catalina_home`, `context_path` | ✓ | 全部引擎 |
| `tomcat.list_apps` | `catalina_home` |  | 全部引擎 |
| `tomcat.reload` | `catalina_home`, `context_path` | ✓ | 全部引擎 |
| `tomcat.version` | `catalina_home` |  | 全部引擎 |

## java_cert

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `java_cert.import` | `keystore_path`, `password`, `alias`, `cert_path`, `cert_type` | ✓ | 全部引擎 |
| `java_cert.remove` | `keystore_path`, `password`, `alias` | ✓ | 全部引擎 |
| `java_cert.list` | `keystore_path`, `password` |  | 全部引擎 |
| `java_cert.exists` | `keystore_path`, `password`, `alias` |  | 全部引擎 |
| `java_cert.export` | `keystore_path`, `password`, `alias`, `output_path`, `cert_type` | ✓ | 全部引擎 |
| `java_cert.info` | `keystore_path`, `password` |  | 全部引擎 |
| `java_cert.import_chain` | `keystore_path`, `password`, `p12_path`, `p12_password` | ✓ | 全部引擎 |
| `java_cert.change_password` | `keystore_path`, `old_password`, `new_password` | ✓ | 全部引擎 |

## maven_artifact

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `maven_artifact.download` | `repo_url`, `group_id`, `artifact_id`, `version`, `dest`, `extension` | ✓ | 全部引擎 |
| `maven_artifact.resolve` | `repo_url`, `group_id`, `artifact_id`, `version`, `extension` |  | 全部引擎 |
| `maven_artifact.deploy` | `repo_url`, `group_id`, `artifact_id`, `version`, `src_path`, `extension` | ✓ | 全部引擎 |
| `maven_artifact.get_latest_version` | `repo_url`, `group_id`, `artifact_id` |  | 全部引擎 |
| `maven_artifact.checksum` | `file_path` |  | 全部引擎 |

## docker_image

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `docker_image.pull` | `name`, `tag`, `force` | ✓ | 全部引擎 |
| `docker_image.build` | `path`, `name`, `tag`, `dockerfile` | ✓ | 全部引擎 |
| `docker_image.remove` | `name`, `tag`, `force` | ✓ | 全部引擎 |
| `docker_image.tag` | `source`, `target` | ✓ | 全部引擎 |
| `docker_image.inspect` | `name` |  | 全部引擎 |
| `docker_image.list` | - |  | 全部引擎 |
| `docker_image.push` | `name`, `tag` |  | 全部引擎 |

## docker_container

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `docker_container.start` | `name` | ✓ | 全部引擎 |
| `docker_container.stop` | `name`, `timeout` | ✓ | 全部引擎 |
| `docker_container.remove` | `name`, `force` | ✓ | 全部引擎 |
| `docker_container.restart` | `name`, `timeout` | ✓ | 全部引擎 |
| `docker_container.pause` | `name` | ✓ | 全部引擎 |
| `docker_container.unpause` | `name` | ✓ | 全部引擎 |
| `docker_container.inspect` | `name` |  | 全部引擎 |
| `docker_container.list` | `all` |  | 全部引擎 |
| `docker_container.logs` | `name`, `tail` |  | 全部引擎 |

## ping

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `ping.ping` | - |  | 全部引擎 |
| `ping.win_ping` | - |  | 全部引擎 |

## find

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `find.find` | `paths`, `patterns`, `file_type`, `recurse`, `depth` |  | 全部引擎 |

## tempfile

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `tempfile.create_file` | `prefix`, `suffix`, `path` | ✓ | 全部引擎 |
| `tempfile.create_dir` | `prefix`, `suffix`, `path` | ✓ | 全部引擎 |
| `tempfile.delete` | `path` | ✓ | 全部引擎 |

## fail

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `fail.fail` | `message` |  | 全部引擎 |

## assert

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `assert.assert` | `condition`, `success_msg`, `fail_msg` |  | 全部引擎 |

## debug

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `debug.debug` | `message` |  | 全部引擎 |
| `debug.debug_var` | `name`, `value` |  | 全部引擎 |

## set_fact

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `set_fact.set` | `key_values` | ✓ | 全部引擎 |
| `set_fact.get` | `key` |  | 全部引擎 |
| `set_fact.get_all` | - |  | 全部引擎 |
| `set_fact.clear` | - | ✓ | 全部引擎 |

## unarchive

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `unarchive.unarchive` | `src`, `dest`, `owner`, `group`, `mode`, `creates` | ✓ | 全部引擎 |

## package_facts

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `package_facts.collect` | `managers` |  | 全部引擎 |

## service_facts

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `service_facts.collect` | - |  | 全部引擎 |

## command

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `command.run` | `command_args`, `chdir`, `creates`, `removes`, `timeout_ms` | ✓ | 全部引擎 |
| `command.shell` | `command_args`, `chdir`, `creates`, `removes`, `timeout_ms`, `executable` | ✓ | 全部引擎 |

## script

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `script.run` | `script_path`, `args`, `chdir`, `creates`, `removes`, `timeout_ms`, `executable` | ✓ | 全部引擎 |

## copy

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `copy.file` | `src`, `dest`, `mode`, `owner`, `group`, `backup` | ✓ | 全部引擎 |
| `copy.content` | `content`, `dest`, `mode`, `owner`, `group`, `backup` | ✓ | 全部引擎 |

## cronvar

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `cronvar.present` | `name`, `value`, `user`, `insertafter`, `insertbefore` | ✓ | 全部引擎 |
| `cronvar.absent` | `name`, `user` | ✓ | 全部引擎 |
| `cronvar.get` | `name`, `user` |  | 全部引擎 |

## stat

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `stat.stat` | `path`, `get_checksum`, `checksum_algo` |  | 全部引擎 |

## add_host

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `add_host.add` | `name`, `groups`, `vars` | ✓ | 全部引擎 |
| `add_host.get_host` | `name` |  | 全部引擎 |
| `add_host.get_group` | `group` |  | 全部引擎 |
| `add_host.list_hosts` | - |  | 全部引擎 |
| `add_host.list_groups` | - |  | 全部引擎 |

## set_stats

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `set_stats.set` | `data` | ✓ | 全部引擎 |
| `set_stats.get` | `key` |  | 全部引擎 |
| `set_stats.get_all` | - |  | 全部引擎 |
| `set_stats.clear` | - | ✓ | 全部引擎 |

## include_vars

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `include_vars.load` | `file` | ✓ | 全部引擎 |
| `include_vars.get` | `key` |  | 全部引擎 |
| `include_vars.get_all` | - |  | 全部引擎 |

## async_status

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `async_status.poll` | `job_id`, `results_dir` |  | 全部引擎 |
| `async_status.cleanup` | `job_id`, `results_dir` | ✓ | 全部引擎 |
| `async_status.wait_for` | `job_id`, `results_dir`, `timeout_ms`, `interval_ms` |  | 全部引擎 |

## package

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `package.install` | `name` | ✓ | 全部引擎 |
| `package.remove` | `name` | ✓ | 全部引擎 |
| `package.update` | `name` | ✓ | 全部引擎 |
| `package.info` | `name` |  | 全部引擎 |

## type_debug

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `type_debug.debug` | `value` |  | 全部引擎 |

## group_by

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `group_by.group_by` | `hosts`, `key` | ✓ | 全部引擎 |
| `group_by.get_hosts` | `group` |  | 全部引擎 |
| `group_by.list_groups` | - |  | 全部引擎 |
| `group_by.clear` | - | ✓ | 全部引擎 |

## normalize

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `normalize.lower` | `value` |  | 全部引擎 |
| `normalize.upper` | `value` |  | 全部引擎 |
| `normalize.trim` | `value` |  | 全部引擎 |
| `normalize.slugify` | `value` |  | 全部引擎 |
| `normalize.title` | `value` |  | 全部引擎 |
| `normalize.camel_case` | `value` |  | 全部引擎 |
| `normalize.snake_case` | `value` |  | 全部引擎 |
| `normalize.kebab_case` | `value` |  | 全部引擎 |

## validate_certs

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `validate_certs.validate` | `host`, `port`, `timeout_ms` |  | 全部引擎 |
| `validate_certs.check_expiry` | `host`, `port`, `days`, `timeout_ms` |  | 全部引擎 |

## mail

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `mail.send` | `host`, `port`, `from`, `to`, `subject`, `body` | ✓ | 全部引擎 |
| `mail.send_html` | `host`, `port`, `from`, `to`, `subject`, `body` | ✓ | 全部引擎 |

## webhook

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `webhook.send` | `url`, `method`, `body` | ✓ | 全部引擎 |
| `webhook.slack` | `url`, `text` | ✓ | 全部引擎 |
| `webhook.discord` | `url`, `content` | ✓ | 全部引擎 |
| `webhook.teams` | `url`, `title`, `text` | ✓ | 全部引擎 |

## openssl_privatekey

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `openssl_privatekey.generate` | `path`, `type`, `size` | ✓ | 全部引擎 |
| `openssl_privatekey.info` | `path` |  | 全部引擎 |
| `openssl_privatekey.delete` | `path` | ✓ | 全部引擎 |

## ip_route

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `ip_route.list` | - |  | 全部引擎 |
| `ip_route.list_table` | `table` |  | 全部引擎 |
| `ip_route.add` | `destination`, `gateway`, `dev`, `metric`, `table` | ✓ | 全部引擎 |
| `ip_route.delete` | `destination`, `table` | ✓ | 全部引擎 |
| `ip_route.flush` | `dev`, `table` | ✓ | 全部引擎 |
| `ip_route.get` | `destination` |  | 全部引擎 |

## ip_link

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `ip_link.list` | - |  | 全部引擎 |
| `ip_link.get` | `name` |  | 全部引擎 |
| `ip_link.set_up` | `name` | ✓ | 全部引擎 |
| `ip_link.set_down` | `name` | ✓ | 全部引擎 |
| `ip_link.set_mtu` | `name`, `mtu` | ✓ | 全部引擎 |
| `ip_link.set_mac` | `name`, `mac` | ✓ | 全部引擎 |
| `ip_link.set_name` | `old_name`, `new_name` | ✓ | 全部引擎 |

## ip_netns

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `ip_netns.list` | - |  | 全部引擎 |
| `ip_netns.get` | `name` |  | 全部引擎 |
| `ip_netns.add` | `name` | ✓ | 全部引擎 |
| `ip_netns.delete` | `name` | ✓ | 全部引擎 |
| `ip_netns.exec` | `namespace`, `command`, `args` |  | 全部引擎 |
| `ip_netns.pids` | `name` |  | 全部引擎 |

## ip_neighbor

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `ip_neighbor.list` | - |  | 全部引擎 |
| `ip_neighbor.list_dev` | `dev` |  | 全部引擎 |
| `ip_neighbor.add` | `ip`, `dev`, `mac` | ✓ | 全部引擎 |
| `ip_neighbor.delete` | `ip`, `dev` | ✓ | 全部引擎 |
| `ip_neighbor.flush` | `dev` | ✓ | 全部引擎 |

## openssl_csr

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `openssl_csr.generate` | `common_name`, `key_file`, `output_file`, `organization`, `organizational_unit`, `country`, `state`, `locality`, `email`, `dns_names`, `force` | ✓ | 全部引擎 |
| `openssl_csr.info` | `csr_file` |  | 全部引擎 |
| `openssl_csr.delete` | `csr_file` | ✓ | 全部引擎 |

## openssl_publickey

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `openssl_publickey.extract` | `private_key_file`, `output_file`, `force` | ✓ | 全部引擎 |
| `openssl_publickey.info` | `public_key_file` |  | 全部引擎 |
| `openssl_publickey.delete` | `public_key_file` | ✓ | 全部引擎 |

## etcd

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `etcd.get` | `key`, `endpoints` |  | 全部引擎 |
| `etcd.set` | `key`, `value`, `endpoints` | ✓ | 全部引擎 |
| `etcd.delete` | `key`, `endpoints` | ✓ | 全部引擎 |
| `etcd.list` | `prefix`, `endpoints` |  | 全部引擎 |

## zookeeper

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `zookeeper.get` | `path`, `servers` |  | 全部引擎 |
| `zookeeper.set` | `path`, `value`, `servers` | ✓ | 全部引擎 |
| `zookeeper.create` | `path`, `value`, `ephemeral`, `servers` | ✓ | 全部引擎 |
| `zookeeper.delete` | `path`, `servers` | ✓ | 全部引擎 |
| `zookeeper.list` | `path`, `servers` |  | 全部引擎 |
| `zookeeper.exists` | `path`, `servers` |  | 全部引擎 |

## vault

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `vault.read` | `path`, `token`, `address` |  | 全部引擎 |
| `vault.write` | `path`, `token`, `address`, `data` | ✓ | 全部引擎 |
| `vault.delete` | `path`, `token`, `address` | ✓ | 全部引擎 |
| `vault.list` | `path`, `token`, `address` |  | 全部引擎 |

## git_config

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `git_config.get` | `key`, `scope` |  | 全部引擎 |
| `git_config.set` | `key`, `value`, `scope` | ✓ | 全部引擎 |
| `git_config.unset` | `key`, `scope` | ✓ | 全部引擎 |
| `git_config.list` | `scope` |  | 全部引擎 |

## sshd_config

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `sshd_config.get` | `key` |  | 全部引擎 |
| `sshd_config.set` | `key`, `value` | ✓ | 全部引擎 |
| `sshd_config.absent` | `key` | ✓ | 全部引擎 |

## docker_network

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `docker_network.inspect` | `name` |  | 全部引擎 |
| `docker_network.create` | `name`, `driver` | ✓ | 全部引擎 |
| `docker_network.remove` | `name` | ✓ | 全部引擎 |
| `docker_network.list` | - |  | 全部引擎 |

## docker_volume

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `docker_volume.inspect` | `name` |  | 全部引擎 |
| `docker_volume.create` | `name`, `driver` | ✓ | 全部引擎 |
| `docker_volume.remove` | `name` | ✓ | 全部引擎 |
| `docker_volume.list` | - |  | 全部引擎 |

## journald

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `journald.get` | `key` |  | 全部引擎 |
| `journald.set` | `key`, `value` | ✓ | 全部引擎 |

## nfs_exports

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `nfs_exports.present` | `path`, `hosts`, `options` | ✓ | 全部引擎 |
| `nfs_exports.absent` | `path` | ✓ | 全部引擎 |
| `nfs_exports.list` | - |  | 全部引擎 |

## postfix

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `postfix.get` | `key` |  | 全部引擎 |
| `postfix.set` | `key`, `value` | ✓ | 全部引擎 |
| `postfix.reload` | - | ✓ | 全部引擎 |

## dnsmasq

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `dnsmasq.get` | `key` |  | 全部引擎 |
| `dnsmasq.set` | `key`, `value` | ✓ | 全部引擎 |
| `dnsmasq.absent` | `key` | ✓ | 全部引擎 |
| `dnsmasq.restart` | - | ✓ | 全部引擎 |

## apache2_module

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `apache2_module.check` | `module` |  | 全部引擎 |
| `apache2_module.enable` | `module` | ✓ | 全部引擎 |
| `apache2_module.disable` | `module` | ✓ | 全部引擎 |

## pipx

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `pipx.install` | `name` | ✓ | 全部引擎 |
| `pipx.uninstall` | `name` | ✓ | 全部引擎 |
| `pipx.upgrade` | `name` | ✓ | 全部引擎 |
| `pipx.list` | - |  | 全部引擎 |

## ssh_config

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `ssh_config.get` | `host`, `option`, `scope` |  | 全部引擎 |
| `ssh_config.set` | `host`, `option`, `value`, `scope` | ✓ | 全部引擎 |
| `ssh_config.absent` | `host`, `option`, `scope` | ✓ | 全部引擎 |

## openvpn

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `openvpn.status` | - |  | 全部引擎 |
| `openvpn.start` | - | ✓ | 全部引擎 |
| `openvpn.stop` | - | ✓ | 全部引擎 |
| `openvpn.restart` | - | ✓ | 全部引擎 |
| `openvpn.enable` | - | ✓ | 全部引擎 |
| `openvpn.disable` | - | ✓ | 全部引擎 |
| `openvpn.genkey` | `output_path` | ✓ | 全部引擎 |
| `openvpn.gen_tls_auth` | `output_path` | ✓ | 全部引擎 |

## btrfs

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `btrfs.filesystem_list` | - |  | 全部引擎 |
| `btrfs.subvolume_list` | `mount_point` |  | 全部引擎 |
| `btrfs.subvolume_create` | `path` | ✓ | 全部引擎 |
| `btrfs.subvolume_delete` | `path` | ✓ | 全部引擎 |
| `btrfs.snapshot_create` | `source`, `dest`, `read_only` | ✓ | 全部引擎 |
| `btrfs.scrub_start` | `mount_point` | ✓ | 全部引擎 |
| `btrfs.scrub_status` | `mount_point` |  | 全部引擎 |
| `btrfs.device_add` | `device_path`, `mount_point` | ✓ | 全部引擎 |
| `btrfs.device_remove` | `device_path`, `mount_point` | ✓ | 全部引擎 |
| `btrfs.balance_start` | `mount_point` | ✓ | 全部引擎 |
| `btrfs.balance_status` | `mount_point` |  | 全部引擎 |
| `btrfs.quota_enable` | `mount_point` | ✓ | 全部引擎 |
| `btrfs.quota_disable` | `mount_point` | ✓ | 全部引擎 |

## certbot

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `certbot.certificates` | - |  | 全部引擎 |
| `certbot.obtain` | `domains`, `email`, `webroot`, `standalone` | ✓ | 全部引擎 |
| `certbot.renew` | `force` | ✓ | 全部引擎 |
| `certbot.delete` | `domain` | ✓ | 全部引擎 |

## gluster

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `gluster.volume_list` | - |  | 全部引擎 |
| `gluster.volume_create` | `name`, `bricks`, `replica`, `stripe`, `transport` | ✓ | 全部引擎 |
| `gluster.volume_delete` | `name` | ✓ | 全部引擎 |
| `gluster.volume_start` | `name` | ✓ | 全部引擎 |
| `gluster.volume_stop` | `name` | ✓ | 全部引擎 |
| `gluster.peer_list` | - |  | 全部引擎 |
| `gluster.peer_probe` | `host` | ✓ | 全部引擎 |
| `gluster.peer_detach` | `host` | ✓ | 全部引擎 |

## nomad

| 操作 | 参数（按位置顺序） | 可变 | 可用范围 |
|---|---|---|---|
| `nomad.job_list` | `namespace` |  | 全部引擎 |
| `nomad.job_run` | `job_file`, `namespace` | ✓ | 全部引擎 |
| `nomad.job_stop` | `job_id`, `namespace` | ✓ | 全部引擎 |
| `nomad.alloc_list` | `job_id`, `namespace` |  | 全部引擎 |
| `nomad.node_list` | - |  | 全部引擎 |
| `nomad.node_drain` | `node_id`, `enable` | ✓ | 全部引擎 |
