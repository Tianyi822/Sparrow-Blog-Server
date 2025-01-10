package markdown

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"
)

// h2MdRenderer 自定义的Markdown渲染器
type h2MdRenderer struct{}

// render 全局Markdown渲染器实例
var (
	render     goldmark.Markdown
	renderOnce sync.Once
)

// InitRenderer 创建并返回Markdown渲染器实例
func InitRenderer() {
	// 使用sync.Once确保渲染器只初始化一次
	renderOnce.Do(func() {
		mr := &h2MdRenderer{}
		render = goldmark.New(
			goldmark.WithExtensions(
				extension.GFM,           // 启用GitHub Flavored Markdown扩展
				extension.Table,         // 添加表格支持
				extension.Strikethrough, // 添加删除线支持
			),
			goldmark.WithRendererOptions(
				html.WithHardWraps(), // 启用硬换行
				html.WithXHTML(),     // 使用XHTML格式
			),
			goldmark.WithRenderer(
				renderer.NewRenderer(
					renderer.WithNodeRenderers(
						util.Prioritized(mr, 1000), // 注册自定义渲染器，优先级为1000
					),
				),
			),
		)
	})
}

// Parse 解析Markdown文本为HTML
func Parse(source []byte) ([]byte, error) {
	var buf bytes.Buffer
	// 使用渲染器将Markdown转换为HTML
	if err := render.Convert(source, &buf); err != nil {
		return nil, fmt.Errorf("markdown 解析失败: %s", err.Error())
	}
	return buf.Bytes(), nil
}

// RegisterFuncs 注册所有渲染器
func (r *h2MdRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	// 注册所有必要的节点渲染器
	// 注册文本节点渲染器
	reg.Register(ast.KindText, r.renderText)
	// 注册字符串节点渲染器
	reg.Register(ast.KindString, r.renderText)
	// 注册段落节点渲染器
	reg.Register(ast.KindParagraph, r.renderParagraph)
	// 注册标题节点渲染器
	reg.Register(ast.KindHeading, r.renderHeading)
	// 注册图片节点渲染器
	reg.Register(ast.KindImage, r.renderImage)
	// 注册链接节点渲染器
	reg.Register(ast.KindLink, r.renderLink)
	// 注册列表节点渲染器
	reg.Register(ast.KindList, r.renderList)
	// 注册列表项节点渲染器
	reg.Register(ast.KindListItem, r.renderListItem)
	// 注册代码块节点渲染器
	reg.Register(ast.KindCodeBlock, r.renderCodeBlock)
	// 注册带围栏的代码块节点渲染器
	reg.Register(ast.KindFencedCodeBlock, r.renderFencedCodeBlock)
	// 注册块引用节点渲染器
	reg.Register(ast.KindBlockquote, r.renderBlockquote)
	// 注册强调节点渲染器
	reg.Register(ast.KindEmphasis, r.renderEmphasis)
	// 注册主题分隔线节点渲染器
	reg.Register(ast.KindThematicBreak, r.renderThematicBreak)
	// 注册文本块节点渲染器
	reg.Register(ast.KindTextBlock, r.renderTextBlock)
}

// renderTextBlock 新增文本块渲染
func (r *h2MdRenderer) renderTextBlock(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		_, _ = w.WriteString(`<div class="h2-blog-text-block">`)
	} else {
		_, _ = w.WriteString("</div>")
	}
	return ast.WalkContinue, nil
}

// renderThematicBreak 新增分隔线渲染
func (r *h2MdRenderer) renderThematicBreak(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		_, _ = w.WriteString(`<hr class="h2-blog-hr" />`)
	}
	return ast.WalkSkipChildren, nil
}

// renderText 文本节点渲染
func (r *h2MdRenderer) renderText(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	node := n.(*ast.Text)
	_, _ = w.Write(node.Segment.Value(source))
	return ast.WalkContinue, nil
}

// renderParagraph 段落渲染
func (r *h2MdRenderer) renderParagraph(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		_, _ = w.WriteString(`<p class="h2-blog-p">`)
	} else {
		_, _ = w.WriteString("</p>")
	}
	return ast.WalkContinue, nil
}

// renderHeading 标题渲染
func (r *h2MdRenderer) renderHeading(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	node := n.(*ast.Heading)
	if entering {
		_, _ = fmt.Fprintf(w, `<h%d class="h2-blog-h">`, node.Level)
	} else {
		_, _ = fmt.Fprintf(w, "</h%d>", node.Level)
	}
	return ast.WalkContinue, nil
}

// renderImage 图片渲染
func (r *h2MdRenderer) renderImage(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	img := n.(*ast.Image)

	// 获取alt文本
	var alt string
	if firstChild := img.FirstChild(); firstChild != nil {
		if text := firstChild.(*ast.Text); text != nil {
			alt = string(text.Segment.Value(source))
		}
	}

	// 构建img标签
	attrs := fmt.Sprintf(`<img src="%s" alt="%s" class="h2-img"`,
		string(img.Destination),
		alt)

	// 添加title属性(如果存在)
	if len(img.Title) > 0 {
		attrs += fmt.Sprintf(` title="%s"`, string(img.Title))
	}

	_, _ = w.WriteString(attrs + " />")
	return ast.WalkSkipChildren, nil
}

// renderLink 链接渲染
func (r *h2MdRenderer) renderLink(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	node := n.(*ast.Link)
	if entering {
		_, _ = fmt.Fprintf(w, `<a href="%s" class="h2-blog-a" title="%s">`,
			string(node.Destination),
			string(node.Title))
	} else {
		_, _ = w.WriteString("</a>")
	}
	return ast.WalkContinue, nil
}

// renderList 是 h2MdRenderer 结构体的一个方法，用于渲染列表节点
func (r *h2MdRenderer) renderList(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	// 将传入的节点 n 断言为 *ast.List 类型
	node := n.(*ast.List)
	// 初始化标签为 "ul"，表示无序列表
	tag := "ul"
	// 检查列表节点是否为有序列表
	if node.IsOrdered() {
		// 如果是有序列表，则将标签设置为 "ol"
		tag = "ol"
	}
	// 根据进入或离开节点的状态，决定是写入开始标签还是结束标签
	if entering {
		// 如果是进入节点，则写入开始标签，并添加 class 属性 "h2-blog-list"
		_, _ = fmt.Fprintf(w, `<%s class="h2-blog-list">`, tag)
	} else {
		// 如果是离开节点，则写入结束标签
		_, _ = fmt.Fprintf(w, "</%s>", tag)
	}
	// 返回 ast.WalkContinue 表示继续遍历节点
	return ast.WalkContinue, nil
}

// 列表项渲染
func (r *h2MdRenderer) renderListItem(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		_, _ = w.WriteString(`<li class="h2-blog-li">`)
	} else {
		_, _ = w.WriteString("</li>")
	}
	return ast.WalkContinue, nil
}

// renderCodeBlock 是 h2MdRenderer 类型的一个方法，用于渲染代码块。
func (r *h2MdRenderer) renderCodeBlock(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	// 将传入的节点转换为 *ast.CodeBlock 类型
	node := n.(*ast.CodeBlock)
	// 判断是否是进入节点（即开始渲染代码块）
	if entering {
		// 写入代码块开始标签，包括自定义的 CSS 类名 "h2-blog-code"
		_, _ = w.WriteString(`<pre><code class="h2-blog-code">`)
		// 获取代码块中的所有行
		lines := node.Lines()
		// 遍历每一行
		for i := 0; i < lines.Len(); i++ {
			line := lines.At(i)
			// 对代码内容进行HTML转义，防止 XSS 攻击
			_, _ = w.Write(util.EscapeHTML(line.Value(source)))
		}
	} else {
		// 写入代码块结束标签
		_, _ = w.WriteString("</code></pre>")
	}
	// 返回继续遍历状态，表示渲染过程正常进行
	return ast.WalkContinue, nil
}

// renderFencedCodeBlock 是 h2MdRenderer 结构体的一个方法，用于渲染带围栏的代码块。
func (r *h2MdRenderer) renderFencedCodeBlock(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	// 将传入的节点转换为带围栏的代码块节点
	node := n.(*ast.FencedCodeBlock)
	// 判断是否是进入节点
	if entering {
		// 获取代码块的语言
		lang := string(node.Language(source))
		// 写入 HTML 标签，包括代码块的语言类
		_, _ = w.WriteString(`<pre><code class="h2-blog-code language-`)
		_, _ = w.WriteString(lang)
		_, _ = w.WriteString(`">`)
		// 获取代码块的行
		lines := node.Lines()
		for i := 0; i < lines.Len(); i++ {
			line := lines.At(i)
			// 对代码内容进行HTML转义，防止 XSS 攻击
			_, _ = w.Write(util.EscapeHTML(line.Value(source)))
		}
	} else {
		// 写入结束标签
		_, _ = w.WriteString("</code></pre>")
	}
	// 返回继续遍历状态，表示处理成功
	return ast.WalkContinue, nil
}

// 引用块渲染
func (r *h2MdRenderer) renderBlockquote(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		_, _ = w.WriteString(`<blockquote class="h2-blog-quote">`)
	} else {
		_, _ = w.WriteString("</blockquote>")
	}
	return ast.WalkContinue, nil
}

// 强调渲染
func (r *h2MdRenderer) renderEmphasis(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		_, _ = w.WriteString(`<em class="h2-blog-em">`)
	} else {
		_, _ = w.WriteString("</em>")
	}
	return ast.WalkContinue, nil
}
