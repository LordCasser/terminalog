# Terminalog - 测试文档

> 文档版本：v1.3
> 创建日期：2026-04-17
> 最后更新：2026-04-17
> 基于需求文档：requirements.md v1.3

---

## 一、测试策略

### 1.1 测试层次

| 层次 | 测试类型 | 工具/方法 |
|------|---------|-----------|
| 前端单元测试 | 组件测试 | Jest + React Testing Library |
| 前端集成测试 | 页面交互测试 | Chrome DevTools MCP |
| 后端单元测试 | API测试 | Go testing package |
| 后端集成测试 | HTTP API测试 | curl + chrome-devtools |
| 前后端联调测试 | 完整流程测试 | chrome-devtools MCP + 临时测试环境 |

### 1.2 测试环境

**集成测试环境**：
- 目录：`/tmp/terminalog-integration-test`
- 后端端口：18085（debug模式）
- 前端端口：3000（npm run dev）
- 测试数据：真实markdown文件 + Git仓库

---

## 二、前端视觉改进测试（v1.3新增）

### 2.1 导航栏统一性测试

**测试目标**：验证主页面和文章查看页面的导航栏字体、样式、大小统一

**测试步骤**：
1. 启动前后端服务（debug模式）
2. 使用chrome-devtools MCP访问主页面
3. 使用chrome-devtools MCP访问文章查看页面
4. 使用evaluate_script提取导航栏字体样式
5. 验证两个页面的导航栏使用相同的字体（JetBrains Mono）、样式（uppercase、tracking-tight）、大小（text-sm）

**验收标准**：
- ✅ 主页面导航栏使用JetBrains Mono字体
- ✅ 主页面导航栏text-transform为uppercase
- ✅ 主页面导航栏letter-spacing为tracking-tight
- ✅ 主页面导航栏font-size为text-sm（14px）
- ✅ 文章查看页面导航栏与主页导航栏样式完全一致

### 2.2 文章标题字体大小测试

**测试目标**：验证文章查看页面标题字体大小优化为text-4xl

**测试步骤**：
1. 使用chrome-devtools MCP访问文章查看页面
2. 使用evaluate_script提取文章标题字体样式
3. 验证标题使用Space Grotesk字体
4. 验证标题font-size为text-4xl（约36px）

**验收标准**：
- ✅ 文章标题使用Space Grotesk字体
- ✅ 文章标题font-size为text-4xl（约36px）
- ✅ 标题颜色为text-on-surface
- ✅ 标题带渐变下划线装饰

### 2.3 Markdown渲染样式测试

**测试目标**：验证Markdown渲染样式符合原型设计

**测试步骤**：
1. 创建包含多种Markdown元素的测试文章（标题、代码块、引用、列表）
2. 使用chrome-devtools MCP访问文章查看页面
3. 使用evaluate_script提取Markdown渲染元素的样式
4. 验证各元素样式符合设计约束

**验收标准**：
- ✅ 正文文本使用Inter字体，text-lg大小
- ✅ h2标题使用Space Grotesk字体，text-3xl大小
- ✅ 代码块使用JetBrains Mono字体，text-sm大小，bg-surface-container-lowest背景
- ✅ 引用块使用JetBrains Mono字体，text-lg大小，border-l-4装饰
- ✅ 列表使用JetBrains Mono字体，带tertiary颜色箭头符号

### 2.4 终端输入栏功能测试

**测试目标**：验证底部终端输入栏实际具备输入功能

**测试步骤**：
1. 使用chrome-devtools MCP访问主页面
2. 使用take_snapshot查看底部输入栏结构
3. 使用evaluate_script验证输入框存在且可聚焦
4. 使用fill在输入框中输入命令
5. 使用press_key按下Enter执行命令

**验收标准**：
- ✅ 输入栏显示格式为 `guest@blog: ~/path $ `
- ✅ 输入栏包含实际可输入的input元素
- ✅ 输入框可正常输入文本
- ✅ 按Enter键可执行命令
- ✅ 命令执行后输入框清空，保留命令历史

### 2.5 键盘输入默认聚焦测试

**测试目标**：验证页面键盘输入自动聚焦到底部输入栏

**测试步骤**：
1. 使用chrome-devtools MCP访问主页面
2. 使用press_key在页面任意位置按下键盘字符（非输入框焦点状态）
3. 使用evaluate_script验证焦点自动跳转到底部输入栏
4. 使用take_snapshot查看输入栏聚焦状态

**验收标准**：
- ✅ 页面任意位置键盘输入自动聚焦到输入栏
- ✅ 输入栏获得焦点后显示闪烁光标
- ✅ 键盘输入字符出现在输入框中
- ✅ 不影响导航栏交互和其他点击操作

### 2.6 搜索icon交互测试

**测试目标**：验证点击导航栏搜索icon的交互逻辑

**测试步骤**：
1. 使用chrome-devtools MCP访问主页面
2. 使用take_snapshot识别搜索icon元素
3. 使用click点击搜索icon
4. 使用evaluate_script验证输入栏自动填充 `search `
5. 使用evaluate_script验证输入栏获得焦点
6. 使用fill输入搜索关键词
7. 使用press_key按下Enter执行搜索

**验收标准**：
- ✅ 点击搜索icon后输入栏自动填充 `search `
- ✅ 输入栏自动获得焦点
- ✅ 光标位于 `search ` 后面，用户可直接输入关键词
- ✅ 输入关键词并按Enter后执行搜索命令
- ✅ 搜索结果显示在页面中

### 2.7 Tab键自动补全测试（v1.3新增）

**测试目标**：验证Tab键命令自动补全功能

**测试步骤**：
1. 使用chrome-devtools MCP访问主页面
2. 使用click聚焦底部输入栏
3. 使用type_text输入命令前缀（如`se`）
4. 使用press_key按下Tab键
5. 使用take_snapshot验证输入框value自动补全为完整命令（如`search `）
6. 验证浏览器默认Tab键焦点切换行为被禁用

**验收标准**：
- ✅ 输入`se`+Tab键自动补全为`search `
- ✅ 输入`op`+Tab键自动补全为`open `
- ✅ 输入`cd`+Tab键自动补全为`cd `
- ✅ Tab键不触发浏览器默认焦点切换行为
- ✅ 多匹配命令时console.log提示（可选）

### 2.7.1 路径补全测试（v1.3新增）

**测试目标**：验证Tab键支持补全文章/文件夹路径

**测试步骤**：
1. 使用chrome-devtools MCP访问主页面
2. 使用click聚焦底部输入栏
3. 使用type_text输入命令+路径前缀（如`open RE`）
4. 使用press_key按下Tab键
5. 使用take_snapshot验证输入框value自动补全为完整路径（如`open README.md`)
6. 测试文件夹路径补全（如`cd tec`→`cd tech/`)
7. 测试多匹配情况（如`open w`有welcome.md和another-welcome.md时）

**验收标准**：
- ✅ 输入`open RE`+Tab键自动补全为`open README.md`
- ✅ 输入`cd tec`+Tab键自动补全为`cd tech/`
- ✅ 路径补全从后端API获取当前目录文章列表和子目录列表
- ✅ 单匹配时自动补全完整路径
- ✅ 多匹配时console.log提示匹配列表（可选）
- ✅ 文件补全不带斜杠，文件夹补全带斜杠

### 2.8 Placeholder透明度测试（v1.3新增）

**测试目标**：验证底部输入栏placeholder透明度降低，避免干扰视觉焦点

**测试步骤**：
1. 使用chrome-devtools MCP访问主页面
2. 使用take_snapshot查看底部输入栏结构
3. 使用evaluate_script提取placeholder样式（opacity）
4. 验证placeholder透明度为opacity-50（约0.5）

**验收标准**：
- ✅ 输入栏placeholder透明度为opacity-50
- ✅ Placeholder颜色为text-on-surface-variant
- ✅ Placeholder不干扰输入框视觉焦点

### 2.9 命令帮助模态框测试（v1.3新增）

**测试目标**：验证输入`help`或`?`命令弹出模态框展示可用命令

**测试步骤**：
1. 使用chrome-devtools MCP访问主页面
2. 使用click聚焦底部输入栏
3. 使用type_text输入`help`命令
4. 使用press_key按下Enter执行命令
5. 使用take_snapshot验证模态框弹出
6. 验证模态框内容包含所有可用命令说明
7. 验证模态框右上角存在x关闭按钮
8. 等待3秒验证模态框自动关闭
9. 再次输入`?`命令测试相同功能
10. 使用click点击x按钮手动关闭模态框

**验收标准**：
- ✅ 输入`help`命令后弹出模态框
- ✅ 输入`?`命令后弹出模态框（功能相同）
- ✅ 模态框内容包含所有可用命令：cd, open, search, help
- ✅ 模态框右上角存在x关闭按钮
- ✅ 3秒后模态框自动关闭
- ✅ 点击x按钮可手动关闭模态框
- ✅ 模态框样式遵循Dracula Spectrum设计系统（Glass效果）
- ✅ 模态框关闭后不影响后续命令输入

---

## 三、后端API测试

### 3.1 文章列表API测试

**测试端点**：`GET /api/v1/articles`

**测试步骤**：
1. 启动后端服务（debug模式）
2. 使用curl请求 `/api/v1/articles`
3. 验证响应包含文章列表
4. 验证响应包含CORS headers（debug模式）

**验收标准**：
- ✅ HTTP状态码200
- ✅ 响应JSON包含articles数组
- ✅ 每个article包含必要字段（path、name、created、edited等）
- ✅ 响应包含 `Access-Control-Allow-Origin: *`

### 3.2 文章内容API测试

**测试端点**：`GET /api/v1/articles/{path}`

**测试步骤**：
1. 使用curl请求 `/api/v1/articles/welcome.md`
2. 验证响应包含文章内容

**验收标准**：
- ✅ HTTP状态码200
- ✅ 响应JSON包含content字段（Markdown文本）
- ✅ 响应包含文章元数据（created、edited、contributors等）

### 3.3 About Me API测试

**测试端点**：`GET /api/v1/special/aboutme`

**测试步骤**：
1. 创建 `_ABOUTME.md` 文件并提交
2. 使用curl请求 `/api/v1/special/aboutme`
3. 验证响应包含About Me内容

**验收标准**：
- ✅ HTTP状态码200
- ✅ 响应JSON包含content字段
- ✅ 响应JSON包含exists=true字段

---

## 四、前后端联调测试流程

### 4.1 测试环境搭建

**环境准备**：
```bash
# 创建测试目录
mkdir -p /tmp/terminalog-integration-test/articles
cd /tmp/terminalog-integration-test

# 构建后端二进制
cd /Users/lordcasser/workspace/projects/terminalog
go build -o /tmp/terminalog-integration-test/terminalog cmd/terminalog/main.go

# 创建配置文件
cat > /tmp/terminalog-integration-test/config.toml << 'EOF'
[server]
port = 18085
host = "127.0.0.1"

[blog]
content_dir = "/tmp/terminalog-integration-test/articles"

[auth]
debug = true
EOF

# 初始化Git仓库并创建测试文章
cd /tmp/terminalog-integration-test/articles
git init
echo "# Welcome" > welcome.md
echo "Welcome to Terminalog" >> welcome.md
git add welcome.md
git commit -m "Initial commit"

# 启动后端服务（debug模式）
cd /tmp/terminalog-integration-test
./terminalog --debug &

# 启动前端开发服务器
cd /Users/lordcasser/workspace/projects/terminalog/frontend
npm run dev &
```

### 4.2 Chrome DevTools MCP测试

**测试命令示例**：
```typescript
// 访问主页面
chrome-devtools_navigate_page({url: "http://localhost:3000"})

// 查看页面结构
chrome-devtools_take_snapshot()

// 测试底部输入栏
chrome-devtools_evaluate_script({
  function: "() => document.querySelector('footer input')?.focus()"
})

// 测试搜索icon交互
chrome-devtools_click({uid: "搜索icon的uid"})
chrome-devtools_evaluate_script({
  function: "() => document.querySelector('footer input')?.value"
})
```

---

## 五、测试报告模板

### 5.1 测试结果记录

| 测试项 | 测试时间 | 测试结果 | 备注 |
|--------|---------|---------|------|
| 导航栏统一性 | YYYY-MM-DD HH:mm | ✅ PASS | 主页面和文章页导航栏样式一致 |
| 文章标题字体 | YYYY-MM-DD HH:mm | ✅ PASS | 标题使用text-4xl |
| Markdown渲染样式 | YYYY-MM-DD HH:mm | ✅ PASS | 所有元素样式符合设计 |
| 终端输入栏功能 | YYYY-MM-DD HH:mm | ✅ PASS | 输入框可正常输入和执行 |
| 键盘输入聚焦 | YYYY-MM-DD HH:mm | ✅ PASS | 键盘输入自动聚焦到输入栏 |
| 搜索icon交互 | YYYY-MM-DD HH:mm | ✅ PASS | 点击icon自动填充search命令 |

---

## 六、自动化测试脚本

### 6.1 前端组件测试脚本（Jest）

```typescript
// tests/components/Navbar.test.tsx

import { render } from '@testing-library/react';
import { Navbar } from '@/components/brutalist/Navbar';

describe('Navbar Component', () => {
  it('should use JetBrains Mono font', () => {
    const { container } = render(<Navbar />);
    const navElement = container.querySelector('nav');
    expect(navElement).toHaveClass('font-mono');
  });

  it('should have uppercase text', () => {
    const { container } = render(<Navbar />);
    const navElement = container.querySelector('nav');
    expect(navElement).toHaveClass('uppercase');
  });

  it('should have tracking-tight', () => {
    const { container } = render(<Navbar />);
    const navElement = container.querySelector('nav');
    expect(navElement).toHaveClass('tracking-tight');
  });
});
```

---

**文档结束**

> ✅ **测试文档完成**：包含前端视觉改进测试、后端API测试、前后端联调测试流程
> 
> 下一步：编写代码实现前端视觉改进，然后进行前后端联调测试
