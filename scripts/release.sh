#!/usr/bin/env bash
# scripts/release.sh — OpsLang 发布脚本
#
# 用法：
#   ./scripts/release.sh 0.2.0
#
# 该脚本会：
#   1. 验证版本号格式
#   2. 运行全部测试
#   3. 更新版本号
#   4. 创建 Git tag
#   5. 构建所有平台二进制
#   6. 输出发布检查清单

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

VERSION="$1"

if [ -z "$VERSION" ]; then
    echo -e "${RED}错误: 请提供版本号${NC}"
    echo "用法: $0 <version>"
    echo "示例: $0 0.2.0"
    exit 1
fi

# 验证版本号格式
if ! echo "$VERSION" | grep -qE '^[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9.]+)?$'; then
    echo -e "${RED}错误: 版本号格式不正确${NC}"
    echo "期望格式: X.Y.Z 或 X.Y.Z-rc.1"
    exit 1
fi

echo -e "${GREEN}=== OpsLang Release v${VERSION} ===${NC}"
echo ""

# 检查 git 状态
if [ -n "$(git status --porcelain)" ]; then
    echo -e "${RED}错误: 工作目录未提交${NC}"
    git status --short
    exit 1
fi

# 运行全部测试
echo -e "${YELLOW}1/5 运行测试...${NC}"
go test ./... -v
echo -e "${GREEN}✓ 测试通过${NC}"
echo ""

# 更新版本号
echo -e "${YELLOW}2/5 更新版本号...${NC}"
# 更新 cmd/ops/main.go 中的版本常量
sed -i.bak "s/version = \"[^\"]*\"/version = \"${VERSION}\"/" cmd/ops/main.go && rm -f cmd/ops/main.go.bak
# 更新 internal/version/version.go
sed -i.bak "s/Version   = \"[^\"]*\"/Version   = \"${VERSION}\"/" internal/version/version.go && rm -f internal/version/version.go.bak
git add -A
git commit -m "chore: bump version to ${VERSION}"
echo -e "${GREEN}✓ 版本号已更新${NC}"
echo ""

# 构建所有平台二进制
echo -e "${YELLOW}3/5 构建二进制...${NC}"
PLATFORMS=(
    "linux/amd64"
    "linux/arm64"
    "darwin/amd64"
    "darwin/arm64"
    "windows/amd64"
)

mkdir -p dist

for platform in "${PLATFORMS[@]}"; do
    IFS='/' read -r GOOS GOARCH <<< "$platform"
    OUTPUT="dist/ops-${GOOS}-${GOARCH}"
    if [ "$GOOS" = "windows" ]; then
        OUTPUT="${OUTPUT}.exe"
    fi
    echo "  构建 ${GOOS}/${GOARCH}..."
    GOOS=$GOOS GOARCH=$GOARCH go build -trimpath -ldflags="-s -w" -o "$OUTPUT" ./cmd/ops
done

echo -e "${GREEN}✓ 二进制已构建${NC}"
ls -lh dist/
echo ""

# 创建 Git tag
echo -e "${YELLOW}4/5 创建 Git tag...${NC}"
git tag -a "v${VERSION}" -m "Release v${VERSION}"
echo -e "${GREEN}✓ Tag v${VERSION} 已创建${NC}"
echo ""

# 推送
echo -e "${YELLOW}5/5 推送到远程...${NC}"
git push origin main
git push origin "v${VERSION}"
echo -e "${GREEN}✓ 已推送${NC}"
echo ""

# 输出检查清单
echo -e "${GREEN}=== 发布准备完成 ===${NC}"
echo ""
echo -e "${YELLOW}发布检查清单:${NC}"
echo "  ☐ 在 GitHub 创建 Release，上传 dist/ 下的二进制文件"
echo "  ☐ 更新 CHANGELOG.md"
echo "  ☐ 更新官网版本号"
echo "  ☐ 通知社区（Twitter/微信/Discord）"
echo ""
echo "  Release URL: https://github.com/opslang/opslang/releases/new?tag=v${VERSION}"
echo ""
