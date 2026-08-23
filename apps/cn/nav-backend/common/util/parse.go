package util

/*
 * @Desc: 格式转换工具类
 * @author: 福狼
 * @version: v1.0.0
 */

import (
	"bytes"
	"github.com/yuin/goldmark"
	"regexp"
	"strings"
)

// BBCode解析规则
var replacements = map[string]string{
	// 文本样式标签
	`(?i)\[b\](.*?)\[/b\]`:               `<strong>$1</strong>`,
	`(?i)\[i\](.*?)\[/i\]`:               `<em>$1</em>`,
	`(?i)\[u\](.*?)\[/u\]`:               `<u>$1</u>`,
	`(?i)\[s\](.*?)\[/s\]`:               `<s>$1</s>`,
	`(?i)\[color=(.*?)\](.*?)\[/color\]`: `<span style="color:$1;">$2</span>`,
	`(?i)\[size=(.*?)\](.*?)\[/size\]`:   `<span style="font-size:$1;">$2</span>`,
	`(?i)\[font=(.*?)\](.*?)\[/font\]`:   `<span style="font-family:$1;">$2</span>`,

	// 结构化标签
	`(?i)\[h1\](.*?)\[/h1\]`:       `<h1>$1</h1>`,
	`(?i)\[h2\](.*?)\[/h2\]`:       `<h2>$1</h2>`,
	`(?i)\[h3\](.*?)\[/h3\]`:       `<h3>$1</h3>`,
	`(?i)\[quote\](.*?)\[/quote\]`: `<blockquote>$1</blockquote>`,
	`(?i)\[code\](.*?)\[/code\]`:   `<pre><code>$1</code></pre>`,

	// 列表标签
	`(?i)\[list\](.*?)\[/list\]`:   `<ul>$1</ul>`,
	`(?i)\[\*\](.*?)\n`:            `<li>$1</li>`,
	`(?i)\[list=1\](.*?)\[/list\]`: `<ol>$1</ol>`,

	// 链接与媒体
	`(?i)\[url\](.*?)\[/url\]`:         `<a href="$1">$1</a>`,
	`(?i)\[url=(.*?)\](.*?)\[/url\]`:   `<a href="$1">$2</a>`,
	`(?i)\[img\](.*?)\[/img\]`:         `<img src="$1" />`,
	`(?i)\[video\](.*?)\[/video\]`:     `<video src="$1" controls></video>`,
	`(?i)\[youtube\](.*?)\[/youtube\]`: `<iframe src="https://www.youtube.com/embed/$1" frameborder="0" allowfullscreen></iframe>`,

	// 表格
	`(?i)\[table\](.*?)\[/table\]`: `<table>$1</table>`,
	`(?i)\[tr\](.*?)\[/tr\]`:       `<tr>$1</tr>`,
	`(?i)\[td\](.*?)\[/td\]`:       `<td>$1</td>`,

	// 对齐
	`(?i)\[center\](.*?)\[/center\]`: `<div style="text-align: center;">$1</div>`,
	`(?i)\[right\](.*?)\[/right\]`:   `<div style="text-align: right;">$1</div>`,
}

// Markdown 转 html
func MarkdownToHTML(markdownContent string) (string, error) {
	// 使用 bytes.Buffer 来接收转换后的 HTML 内容
	var buf bytes.Buffer
	md := goldmark.New()

	// 将 Markdown 转换为 HTML
	err := md.Convert([]byte(markdownContent), &buf)
	if err != nil {
		return "", err
	}

	// 返回转换后的 HTML 字符串
	return buf.String(), nil
}

// BBCode 转 html
func BBCodeToHTML(input string) string {
	// 正则表达式替换 BBCode
	for pattern, replacement := range replacements {
		re := regexp.MustCompile(pattern)
		input = re.ReplaceAllString(input, replacement)
	}
	// 将 \n 替换为 <br>
	input = strings.ReplaceAll(input, "\n", "<br>")
	// 替换换行符周围的 URL
	urlInBrPattern := regexp.MustCompile(`(?i)<br>(https?://[^\s]+?)<br>`)
	input = urlInBrPattern.ReplaceAllString(input, `<a href="$1">$1</a>`)
	urlPattern := regexp.MustCompile(`(?i)<br>(https?://[^\s]+) <br>`)
	input = urlPattern.ReplaceAllString(input, `<a href="$1">$1</a>`)

	return input
}
