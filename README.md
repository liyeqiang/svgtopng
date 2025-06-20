# SVG转PNG工具

这是一个使用Go语言开发的SVG转PNG转换工具，采用纯Go实现，无需任何外部依赖。

## 特性

- 支持任何有效的SVG文件转换为PNG
- 纯Go实现，无需浏览器或其他外部工具
- 支持自定义输出尺寸和缩放比例
- 自动保持SVG原始宽高比
- 简单的命令行界面
- 跨平台支持

## 依赖

- Go 1.21或更高版本

## 安装

1. 克隆项目：
```bash
git clone <repository-url>
cd svgtopng
```

2. 安装依赖：
```bash
go mod tidy
```

3. 编译：
```bash
go build -o svgtopng
```

## 使用方法

### 基本用法

```bash
# 转换SVG文件为PNG（使用原始尺寸）
./svgtopng -input input.svg

# 指定输出文件名
./svgtopng -input input.svg -output output.png
```

### 高级选项

```bash
# 设置输出尺寸
./svgtopng -input input.svg -output output.png -width 1024 -height 768

# 只设置宽度，高度按比例计算
./svgtopng -input input.svg -output output.png -width 800

# 只设置高度，宽度按比例计算
./svgtopng -input input.svg -output output.png -height 600

# 设置缩放比例
./svgtopng -input input.svg -output output.png -scale 2.0

# 组合使用
./svgtopng -input input.svg -output output.png -width 800 -scale 1.5
```

### 参数说明

- `-input`: 输入SVG文件路径（必需）
- `-output`: 输出PNG文件路径（可选，默认与输入文件同名但扩展名为.png）
- `-width`: 输出图片宽度（像素），为0时按比例计算或使用原始尺寸
- `-height`: 输出图片高度（像素），为0时按比例计算或使用原始尺寸
- `-scale`: 缩放比例（默认1.0）

## 示例

假设你有一个名为`logo.svg`的SVG文件：

```bash
# 使用原始尺寸转换
./svgtopng -input logo.svg

# 转换为高分辨率PNG
./svgtopng -input logo.svg -output logo-hd.png -width 2048 -height 2048

# 创建2倍大小的版本
./svgtopng -input logo.svg -output logo@2x.png -scale 2.0

# 转换为固定宽度，保持宽高比
./svgtopng -input logo.svg -output logo-800w.png -width 800
```

## 工作原理

1. 读取输入的SVG文件
2. 使用oksvg库解析SVG内容
3. 使用rasterx库将SVG渲染为位图
4. 将位图编码为PNG格式并保存

## 支持的SVG特性

- 基本图形（矩形、圆形、椭圆、线条、多边形等）
- 路径元素
- 文本元素（需要系统字体支持）
- 渐变和图案
- 变换（旋转、缩放、平移等）
- 样式和CSS属性

## 注意事项

- 不支持某些高级SVG特性（如滤镜、动画等）
- 文本渲染需要系统安装相应字体
- 对于复杂的SVG文件，可能需要调整输出尺寸以获得最佳效果

## 许可证

请查看LICENSE文件了解许可证信息。 