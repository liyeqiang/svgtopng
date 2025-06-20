@echo off
echo SVG转PNG工具 - 浏览器渲染功能演示
echo ========================================

echo.
echo 1. 库渲染模式（默认）- 速度快，适合简单SVG
go run main.go -input mermaid_flowchart.svg -output demo_lib_render.png -width 1000 -bg white -verbose

echo.
echo 2. 浏览器渲染模式 - 完美兼容性，适合复杂SVG
go run main.go -input mermaid_flowchart.svg -output demo_browser_render.png -browser -width 1000 -bg white -verbose

echo.
echo 3. 批量浏览器渲染演示
go run main.go -batch -browser -width 800 -bg white -timeout 20

echo.
echo 4. 高分辨率浏览器渲染
go run main.go -input mermaid_flowchart.svg -output demo_hires_browser.png -browser -width 1600 -bg white -scale 1.5

echo.
echo 演示完成！
echo 对比查看以下文件的渲染效果：
echo - demo_lib_render.png （库渲染）
echo - demo_browser_render.png （浏览器渲染）
echo - demo_hires_browser.png （高分辨率浏览器渲染）
echo.
echo 浏览器渲染模式特别适合：
echo - 复杂的Mermaid流程图
echo - 包含Web字体的设计
echo - 使用复杂CSS样式的SVG
echo - 需要最高视觉保真度的场景
pause 