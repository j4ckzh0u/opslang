# OpsLang system summary test

`system_summary.ops` is a read-only Linux host check covering:

- root filesystem usage and an 80% threshold;
- CPU utilization and a 60% threshold, plus 1/5/15 minute load averages;
- aggregate and per-interface network throughput measured over **3 seconds**;
- all packages discovered through `apt`/`rpm` auto-detection, including name and version;
- running Java processes, classpath JAR names/versions, and Docker/containerd cgroup IDs.

The network value is a measured 3-second average (`bits_per_second`), not a five-minute rolling average. Use `sys.net.rate(300)` when a five-minute measurement is required.

Java discovery reads Linux `/proc`. A host can only see Java processes from containers that share its PID namespace; for an isolated container, execute the same script from inside that container.

Result JSON files in this directory are generated test artifacts. No SSH credentials are stored here.
