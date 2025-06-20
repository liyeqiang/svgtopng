package main

import (
	"context"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
)

// ConvertOptions 转换选项
type ConvertOptions struct {
	InputFile       string
	OutputFile      string
	Width           int
	Height          int
	Scale           float64
	BackgroundColor string
	Quality         int
	Verbose         bool
	BatchMode       bool
	Directory       string
	OutputFormat    string
	BrowserRender   bool
	BrowserTimeout  int
	BrowserHeadless bool
}

func main() {
	var opts ConvertOptions

	flag.StringVar(&opts.InputFile, "input", "", "输入SVG文件路径")
	flag.StringVar(&opts.OutputFile, "output", "", "输出PNG文件路径")
	flag.IntVar(&opts.Width, "width", 0, "输出图片宽度(像素)")
	flag.IntVar(&opts.Height, "height", 0, "输出图片高度(像素)")
	flag.Float64Var(&opts.Scale, "scale", 1.0, "缩放比例")
	flag.StringVar(&opts.BackgroundColor, "bg", "transparent", "背景色(transparent, white, black, 或十六进制颜色如#ffffff)")
	flag.IntVar(&opts.Quality, "quality", 9, "PNG压缩质量(1-9，9为最佳)")
	flag.BoolVar(&opts.Verbose, "verbose", false, "显示详细信息")
	flag.BoolVar(&opts.BatchMode, "batch", false, "批量处理模式")
	flag.StringVar(&opts.Directory, "dir", "", "批量处理的目录路径")
	flag.StringVar(&opts.OutputFormat, "format", "png", "输出格式(png)")
	flag.BoolVar(&opts.BrowserRender, "browser", false, "使用浏览器渲染模式（提供更好的兼容性）")
	flag.IntVar(&opts.BrowserTimeout, "timeout", 30, "浏览器渲染超时时间（秒）")
	flag.BoolVar(&opts.BrowserHeadless, "headless", true, "无头浏览器模式")

	flag.Parse()

	// 验证参数
	if !opts.BatchMode && opts.InputFile == "" {
		log.Fatal("请指定输入SVG文件路径，使用 -input 参数，或使用 -batch 模式")
	}

	if opts.BatchMode {
		if opts.Directory == "" {
			opts.Directory = "."
		}
		err := batchConvert(&opts)
		if err != nil {
			log.Fatalf("批量转换失败: %v", err)
		}
		return
	}

	// 单文件转换
	err := convertSingleFile(&opts)
	if err != nil {
		log.Fatalf("转换失败: %v", err)
	}
}

func batchConvert(opts *ConvertOptions) error {
	files, err := filepath.Glob(filepath.Join(opts.Directory, "*.svg"))
	if err != nil {
		return fmt.Errorf("搜索SVG文件失败: %v", err)
	}

	if len(files) == 0 {
		return fmt.Errorf("在目录 %s 中未找到SVG文件", opts.Directory)
	}

	mode := "库渲染"
	if opts.BrowserRender {
		mode = "浏览器渲染"
	}
	fmt.Printf("找到 %d 个SVG文件，开始批量转换...（模式：%s）\n", len(files), mode)

	successCount := 0
	startTime := time.Now()

	for i, file := range files {
		fmt.Printf("[%d/%d] 处理: %s\n", i+1, len(files), filepath.Base(file))

		// 为每个文件创建单独的选项
		fileOpts := *opts
		fileOpts.InputFile = file
		fileOpts.OutputFile = ""

		if err := convertSingleFile(&fileOpts); err != nil {
			fmt.Printf("  ❌ 转换失败: %v\n", err)
		} else {
			fmt.Printf("  ✅ 转换成功\n")
			successCount++
		}
	}

	duration := time.Since(startTime)
	fmt.Printf("\n批量转换完成! 成功: %d/%d, 耗时: %v\n", successCount, len(files), duration)

	return nil
}

func convertSingleFile(opts *ConvertOptions) error {
	if opts.OutputFile == "" {
		// 如果没有指定输出文件，使用输入文件名但改为.png扩展名
		ext := filepath.Ext(opts.InputFile)
		opts.OutputFile = strings.TrimSuffix(opts.InputFile, ext) + ".png"
	}

	// 读取SVG文件
	svgContent, err := ioutil.ReadFile(opts.InputFile)
	if err != nil {
		return fmt.Errorf("读取SVG文件失败: %v", err)
	}

	if opts.Verbose {
		mode := "库渲染"
		if opts.BrowserRender {
			mode = "浏览器渲染"
		}
		fmt.Printf("读取SVG文件: %s (大小: %d 字节) - 模式: %s\n", opts.InputFile, len(svgContent), mode)
	}

	// 选择渲染方式
	if opts.BrowserRender {
		err = convertSVGToPNGBrowser(string(svgContent), opts.OutputFile, opts)
	} else {
		err = convertSVGToPNG(string(svgContent), opts.OutputFile, opts)
	}

	if err != nil {
		return fmt.Errorf("转换失败: %v", err)
	}

	if !opts.BatchMode {
		mode := "库渲染"
		if opts.BrowserRender {
			mode = "浏览器渲染"
		}
		fmt.Printf("成功转换 %s 到 %s (模式: %s)\n", opts.InputFile, opts.OutputFile, mode)
	}

	return nil
}

func createSVGHTML(svgContent string, opts *ConvertOptions) string {
	width := opts.Width
	height := opts.Height

	// 如果没有指定尺寸，使用默认值
	if width == 0 {
		width = 1024
	}
	if height == 0 {
		height = 768
	}

	// 应用缩放
	width = int(float64(width) * opts.Scale)
	height = int(float64(height) * opts.Scale)

	bgColor := opts.BackgroundColor
	if bgColor == "transparent" {
		bgColor = "rgba(0,0,0,0)"
	}

	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body {
            margin: 0;
            padding: 0;
            background-color: %s;
            display: flex;
            justify-content: center;
            align-items: center;
            min-height: 100vh;
        }
        .svg-container {
            width: %dpx;
            height: %dpx;
            display: flex;
            justify-content: center;
            align-items: center;
        }
        svg {
            max-width: 100%%;
            max-height: 100%%;
            width: auto;
            height: auto;
        }
    </style>
</head>
<body>
    <div class="svg-container">
        %s
    </div>
</body>
</html>`, bgColor, width, height, svgContent)

	return html
}

func convertSVGToPNGBrowser(svgContent, outputPath string, opts *ConvertOptions) error {
	// 创建HTML内容
	htmlContent := createSVGHTML(svgContent, opts)

	// 设置Chrome选项
	var chromeOpts []chromedp.ExecAllocatorOption
	if opts.BrowserHeadless {
		chromeOpts = append(chromeOpts, chromedp.DefaultExecAllocatorOptions[:]...)
	} else {
		chromeOpts = append(chromeOpts, chromedp.DefaultExecAllocatorOptions[:len(chromedp.DefaultExecAllocatorOptions)-1]...)
		chromeOpts = append(chromeOpts, chromedp.Flag("headless", false))
	}

	// 添加其他选项
	chromeOpts = append(chromeOpts,
		chromedp.Flag("disable-gpu", false),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-extensions", true),
		chromedp.Flag("hide-scrollbars", true),
	)

	// 创建浏览器上下文
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), chromeOpts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	// 设置超时
	ctx, cancel = context.WithTimeout(ctx, time.Duration(opts.BrowserTimeout)*time.Second)
	defer cancel()

	// 计算视口尺寸
	width := opts.Width
	height := opts.Height
	if width == 0 {
		width = 1024
	}
	if height == 0 {
		height = 768
	}
	width = int(float64(width) * opts.Scale)
	height = int(float64(height) * opts.Scale)

	var buf []byte
	err := chromedp.Run(ctx,
		chromedp.EmulateViewport(int64(width), int64(height)),
		chromedp.Navigate("data:text/html,"+htmlContent),
		chromedp.WaitVisible(".svg-container", chromedp.ByQuery),
		chromedp.Sleep(1*time.Second), // 等待渲染完成
		chromedp.FullScreenshot(&buf, 90),
	)

	if err != nil {
		return fmt.Errorf("浏览器渲染失败: %v", err)
	}

	if opts.Verbose {
		fmt.Printf("浏览器渲染完成，截图大小: %.2f KB\n", float64(len(buf))/1024)
	}

	// 保存截图
	err = ioutil.WriteFile(outputPath, buf, 0644)
	if err != nil {
		return fmt.Errorf("保存截图失败: %v", err)
	}

	if opts.Verbose {
		fileInfo, _ := os.Stat(outputPath)
		fmt.Printf("输出文件大小: %.2f KB\n", float64(fileInfo.Size())/1024)
	}

	return nil
}

func parseBackgroundColor(colorStr string) (color.Color, error) {
	switch strings.ToLower(colorStr) {
	case "transparent":
		return color.RGBA{0, 0, 0, 0}, nil
	case "white":
		return color.RGBA{255, 255, 255, 255}, nil
	case "black":
		return color.RGBA{0, 0, 0, 255}, nil
	default:
		// 尝试解析十六进制颜色
		if strings.HasPrefix(colorStr, "#") {
			colorStr = colorStr[1:]
		}
		if len(colorStr) == 6 {
			r, err1 := strconv.ParseUint(colorStr[0:2], 16, 8)
			g, err2 := strconv.ParseUint(colorStr[2:4], 16, 8)
			b, err3 := strconv.ParseUint(colorStr[4:6], 16, 8)
			if err1 != nil || err2 != nil || err3 != nil {
				return nil, fmt.Errorf("无效的颜色格式: %s", colorStr)
			}
			return color.RGBA{uint8(r), uint8(g), uint8(b), 255}, nil
		}
		return nil, fmt.Errorf("不支持的颜色格式: %s", colorStr)
	}
}

func convertSVGToPNG(svgContent, outputPath string, opts *ConvertOptions) error {
	// 解析SVG
	icon, err := oksvg.ReadIconStream(strings.NewReader(svgContent))
	if err != nil {
		return fmt.Errorf("解析SVG失败: %v", err)
	}

	// 获取SVG的原始尺寸
	svgWidth := icon.ViewBox.W
	svgHeight := icon.ViewBox.H

	if opts.Verbose {
		fmt.Printf("SVG原始尺寸: %.1f x %.1f\n", svgWidth, svgHeight)
	}

	// 如果SVG没有ViewBox，尝试从width/height属性获取
	if svgWidth == 0 || svgHeight == 0 {
		if svgWidth == 0 {
			svgWidth = 512
		}
		if svgHeight == 0 {
			svgHeight = 512
		}
		if opts.Verbose {
			fmt.Printf("使用默认尺寸: %.1f x %.1f\n", svgWidth, svgHeight)
		}
	}

	// 设置输出尺寸
	outputWidth := opts.Width
	outputHeight := opts.Height

	if outputWidth == 0 && outputHeight == 0 {
		// 如果没有指定尺寸，使用SVG原始尺寸
		outputWidth = int(svgWidth)
		outputHeight = int(svgHeight)
		// 如果原始尺寸太小，设置最小尺寸
		if outputWidth < 64 {
			outputWidth = 256
		}
		if outputHeight < 64 {
			outputHeight = 256
		}
	} else if outputWidth == 0 {
		// 根据高度按比例计算宽度
		outputWidth = int(float64(outputHeight) * svgWidth / svgHeight)
	} else if outputHeight == 0 {
		// 根据宽度按比例计算高度
		outputHeight = int(float64(outputWidth) * svgHeight / svgWidth)
	}

	// 应用缩放
	outputWidth = int(float64(outputWidth) * opts.Scale)
	outputHeight = int(float64(outputHeight) * opts.Scale)

	if opts.Verbose || !opts.BatchMode {
		fmt.Printf("输出尺寸: %d x %d 像素 (缩放: %.2fx)\n", outputWidth, outputHeight, opts.Scale)
	}

	// 设置图标尺寸
	icon.SetTarget(0, 0, float64(outputWidth), float64(outputHeight))

	// 创建图像
	img := image.NewRGBA(image.Rect(0, 0, outputWidth, outputHeight))

	// 设置背景色
	bgColor, err := parseBackgroundColor(opts.BackgroundColor)
	if err != nil {
		return fmt.Errorf("解析背景色失败: %v", err)
	}

	// 填充背景
	if opts.BackgroundColor != "transparent" {
		draw.Draw(img, img.Bounds(), &image.Uniform{bgColor}, image.Point{}, draw.Src)
	}

	// 创建渲染器
	scanner := rasterx.NewScannerGV(outputWidth, outputHeight, img, img.Bounds())
	raster := rasterx.NewDasher(outputWidth, outputHeight, scanner)

	// 渲染SVG
	icon.Draw(raster, 1.0)

	// 保存为PNG
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("创建输出文件失败: %v", err)
	}
	defer file.Close()

	// 设置PNG编码器选项
	encoder := png.Encoder{
		CompressionLevel: png.CompressionLevel(opts.Quality),
	}

	err = encoder.Encode(file, img)
	if err != nil {
		return fmt.Errorf("编码PNG失败: %v", err)
	}

	if opts.Verbose {
		// 获取文件大小
		fileInfo, _ := file.Stat()
		fmt.Printf("输出文件大小: %.2f KB\n", float64(fileInfo.Size())/1024)
	}

	return nil
}
