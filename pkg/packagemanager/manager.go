// Package packagemanager 实现 OpsLang 的包管理器
package packagemanager

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// PackageManifest 包清单
type PackageManifest struct {
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Description string   `json:"description"`
	Author      string   `json:"author"`
	Files       []string `json:"files"`
	Dependencies []string `json:"dependencies"`
}

// Manager 包管理器
type Manager struct {
	rootDir string // ~/.opslang
}

// NewManager 创建包管理器
func NewManager() (*Manager, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("获取用户目录失败: %w", err)
	}

	rootDir := filepath.Join(home, ".opslang")
	packagesDir := filepath.Join(rootDir, "packages")

	// 创建目录
	if err := os.MkdirAll(packagesDir, 0755); err != nil {
		return nil, fmt.Errorf("创建包目录失败: %w", err)
	}

	return &Manager{rootDir: rootDir}, nil
}

// Install 安装包
func (m *Manager) Install(packageName string) error {
	fmt.Printf("正在安装 %s ...\n", packageName)

	// 解析包名（支持 user/repo 格式）
	parts := strings.Split(packageName, "/")
	if len(parts) != 2 {
		return fmt.Errorf("包名格式错误，应为 user/repo: %s", packageName)
	}
	user, repo := parts[0], parts[1]

	// 从 GitHub 下载 ops.toml
	manifestURL := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/main/ops.json", user, repo)
	resp, err := http.Get(manifestURL)
	if err != nil {
		return fmt.Errorf("下载清单失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("包不存在或无法访问: %s (HTTP %d)", packageName, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取清单失败: %w", err)
	}

	var manifest PackageManifest
	if err := json.Unmarshal(body, &manifest); err != nil {
		return fmt.Errorf("解析清单失败: %w", err)
	}

	// 创建包目录
	pkgDir := filepath.Join(m.rootDir, "packages", manifest.Name)
	if err := os.MkdirAll(pkgDir, 0755); err != nil {
		return fmt.Errorf("创建包目录失败: %w", err)
	}

	// 下载包文件
	for _, file := range manifest.Files {
		fileURL := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/main/%s", user, repo, file)
		fmt.Printf("  下载 %s ...\n", file)

		resp, err := http.Get(fileURL)
		if err != nil {
			return fmt.Errorf("下载文件失败 %s: %w", file, err)
		}

		fileBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return fmt.Errorf("读取文件失败 %s: %w", file, err)
		}

		destPath := filepath.Join(pkgDir, file)
		os.MkdirAll(filepath.Dir(destPath), 0755)
		if err := os.WriteFile(destPath, fileBody, 0644); err != nil {
			return fmt.Errorf("写入文件失败 %s: %w", file, err)
		}
	}

	// 保存清单
	manifestPath := filepath.Join(pkgDir, "ops.json")
	if err := os.WriteFile(manifestPath, body, 0644); err != nil {
		return fmt.Errorf("保存清单失败: %w", err)
	}

	fmt.Printf("✅ %s v%s 安装成功\n", manifest.Name, manifest.Version)
	return nil
}

// InstallLocal 从本地路径安装
func (m *Manager) InstallLocal(path string) error {
	// 读取本地 ops.json
	manifestPath := filepath.Join(path, "ops.json")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return fmt.Errorf("读取清单失败: %w", err)
	}

	var manifest PackageManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return fmt.Errorf("解析清单失败: %w", err)
	}

	// 复制文件到包目录
	pkgDir := filepath.Join(m.rootDir, "packages", manifest.Name)
	if err := os.MkdirAll(pkgDir, 0755); err != nil {
		return fmt.Errorf("创建包目录失败: %w", err)
	}

	for _, file := range manifest.Files {
		src := filepath.Join(path, file)
		dst := filepath.Join(pkgDir, file)
		os.MkdirAll(filepath.Dir(dst), 0755)

		data, err := os.ReadFile(src)
		if err != nil {
			return fmt.Errorf("读取文件失败 %s: %w", file, err)
		}
		if err := os.WriteFile(dst, data, 0644); err != nil {
			return fmt.Errorf("写入文件失败 %s: %w", file, err)
		}
	}

	// 保存清单
	destManifest := filepath.Join(pkgDir, "ops.json")
	if err := os.WriteFile(destManifest, data, 0644); err != nil {
		return fmt.Errorf("保存清单失败: %w", err)
	}

	fmt.Printf("✅ %s v%s 本地安装成功\n", manifest.Name, manifest.Version)
	return nil
}

// List 列出已安装的包
func (m *Manager) List() ([]PackageManifest, error) {
	packagesDir := filepath.Join(m.rootDir, "packages")
	entries, err := os.ReadDir(packagesDir)
	if err != nil {
		return nil, fmt.Errorf("读取包目录失败: %w", err)
	}

	var packages []PackageManifest
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		manifestPath := filepath.Join(packagesDir, entry.Name(), "ops.json")
		data, err := os.ReadFile(manifestPath)
		if err != nil {
			continue // 跳过无效包
		}

		var manifest PackageManifest
		if err := json.Unmarshal(data, &manifest); err != nil {
			continue
		}
		packages = append(packages, manifest)
	}

	return packages, nil
}

// Uninstall 卸载包
func (m *Manager) Uninstall(packageName string) error {
	pkgDir := filepath.Join(m.rootDir, "packages", packageName)
	if _, err := os.Stat(pkgDir); os.IsNotExist(err) {
		return fmt.Errorf("包未安装: %s", packageName)
	}

	if err := os.RemoveAll(pkgDir); err != nil {
		return fmt.Errorf("删除包失败: %w", err)
	}

	fmt.Printf("✅ %s 已卸载\n", packageName)
	return nil
}

// GetPackagePath 获取包路径
func (m *Manager) GetPackagePath(packageName string) (string, error) {
	pkgDir := filepath.Join(m.rootDir, "packages", packageName)
	if _, err := os.Stat(pkgDir); os.IsNotExist(err) {
		return "", fmt.Errorf("包未安装: %s", packageName)
	}
	return pkgDir, nil
}

// InitPackage 初始化一个新包
func (m *Manager) InitPackage(name, version, description string) error {
	manifest := PackageManifest{
		Name:        name,
		Version:     version,
		Description: description,
		Files:       []string{name + ".ops"},
	}

	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("生成清单失败: %w", err)
	}

	if err := os.WriteFile("ops.json", data, 0644); err != nil {
		return fmt.Errorf("写入清单失败: %w", err)
	}

	// 创建主文件
	mainFile := name + ".ops"
	mainContent := fmt.Sprintf("// %s - %s\n// Version: %s\n", name, description, version)
	if err := os.WriteFile(mainFile, []byte(mainContent), 0644); err != nil {
		return fmt.Errorf("创建主文件失败: %w", err)
	}

	fmt.Printf("✅ 包 %s 初始化成功\n", name)
	fmt.Println("  ops.json    — 包清单")
	fmt.Printf("  %s  — 主文件\n", mainFile)
	return nil
}
