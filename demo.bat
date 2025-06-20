@echo off
echo SVG转PNG工具演示
echo ====================

echo.
echo 1. 基本转换（透明背景）
go run main.go -input mermaid_flowchart.svg -output demo1_transparent.png -width 800

echo.
echo 2. 白色背景转换
go run main.go -input mermaid_flowchart.svg -output demo2_white.png -width 800 -bg white

echo.
echo 3. 高分辨率转换
go run main.go -input mermaid_flowchart.svg -output demo3_hires.png -width 1200 -bg white

echo.
echo 4. 批量处理演示
go run main.go -batch -width 600 -bg white

echo.
echo 演示完成！查看生成的PNG文件。
pause 