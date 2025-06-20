package main

import (
	"flag"
	"fmt"
	"image"
	"image/png"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
)

func main() {
	var (
		inputFile  = flag.String("input", "", "输入SVG文件路径")
		outputFile = flag.String("output", "", "输出PNG文件路径")
		width      = flag.Int("width", 0, "输出图片宽度(像素)")
		height     = flag.Int("height", 0, "输出图片高度(像素)")
		scale      = flag.Float64("scale", 1.0, "缩放比例")
	)
	flag.Parse()

	if *inputFile == "" {
		log.Fatal("请指定输入SVG文件路径，使用 -input 参数")
	}

	if *outputFile == "" {
		// 如果没有指定输出文件，使用输入文件名但改为.png扩展名
		ext := filepath.Ext(*inputFile)
		*outputFile = strings.TrimSuffix(*inputFile, ext) + ".png"
	}

	// 读取SVG文件
	svgContent, err := ioutil.ReadFile(*inputFile)
	if err != nil {
		log.Fatalf("读取SVG文件失败: %v", err)
	}

	// 转换SVG到PNG
	err = convertSVGToPNG(string(svgContent), *outputFile, *width, *height, *scale)
	if err != nil {
		log.Fatalf("转换失败: %v", err)
	}

	fmt.Printf("成功转换 %s 到 %s\n", *inputFile, *outputFile)
}

func convertSVGToPNG(svgContent, outputPath string, width, height int, scale float64) error {
	// 解析SVG
	icon, err := oksvg.ReadIconStream(strings.NewReader(svgContent))
	if err != nil {
		return fmt.Errorf("解析SVG失败: %v", err)
	}

	// 获取SVG的原始尺寸
	svgWidth := icon.ViewBox.W
	svgHeight := icon.ViewBox.H

	// 如果SVG没有ViewBox，尝试从width/height属性获取
	if svgWidth == 0 || svgHeight == 0 {
		svgWidth = icon.ViewBox.W
		svgHeight = icon.ViewBox.H
		if svgWidth == 0 {
			svgWidth = 256
		}
		if svgHeight == 0 {
			svgHeight = 256
		}
	}

	// 设置输出尺寸
	outputWidth := width
	outputHeight := height

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
	outputWidth = int(float64(outputWidth) * scale)
	outputHeight = int(float64(outputHeight) * scale)

	fmt.Printf("输出尺寸: %d x %d 像素\n", outputWidth, outputHeight)

	// 设置图标尺寸
	icon.SetTarget(0, 0, float64(outputWidth), float64(outputHeight))

	// 创建图像
	img := image.NewRGBA(image.Rect(0, 0, outputWidth, outputHeight))

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

	err = png.Encode(file, img)
	if err != nil {
		return fmt.Errorf("编码PNG失败: %v", err)
	}

	return nil
}
