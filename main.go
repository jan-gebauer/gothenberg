package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"

	chromaHtml "github.com/alecthomas/chroma/formatters/html"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
)

func loadText(textPath string) string {
	dat, err := os.ReadFile(textPath)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	text := string(dat)
	return text
}

func extractOrder(orderText string) []string {
	order := []string{}
	for _, line := range strings.Split(orderText, "\n") {
		order = append(order, strings.TrimSpace(line))
	}
	return order
}

func buildToc(organisedText []string) string {

	links := []string{}
	const headingOne = "# "
	const headingTwo = "## "
	rHeadingOne, _ := regexp.Compile(fmt.Sprintf("^%s", headingOne))
	rHeadingTwo, _ := regexp.Compile(fmt.Sprintf("^%s", headingTwo))
	headingTwoCount := 0
	for _, text := range organisedText {
		if !strings.Contains(text, "```") {
			splitText := strings.Split(text, "\n")
			for _, line := range splitText {
				if rHeadingOne.FindString(line) != "" {
					link := createTocLine(line, len(links), headingOne)
					links = append(links, link)
					headingTwoCount = 0
				} else if rHeadingTwo.FindString(line) != "" {
					link := createTocLine(line, headingTwoCount, headingTwo)
					indentedLink := "    " + link
					links = append(links, indentedLink)
					headingTwoCount = headingTwoCount + 1
				}
			}
		}
	}

	joinedLinks := strings.Join(links, "")
	return "# Table of Contents\n" + joinedLinks
}

func createTocLine(line string, index int, headingMd string) string {
	tableAnchor := strings.ToLower(line)
	tableAnchor = strings.ReplaceAll(tableAnchor, headingMd, "#")
	tableAnchor = strings.ReplaceAll(tableAnchor, " ", "-")
	tableAnchor = strings.ReplaceAll(tableAnchor, "'", "")

	chapterName := strings.ReplaceAll(line, headingMd, "")

	link := fmt.Sprintf("%d. [%s](%s)\n", index+1, chapterName, tableAnchor)
	return link
}

func mdToHTML(md []byte) []byte {
	// create markdown parser with extensions
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse(md)

	// create HTML renderer with extensions
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	return markdown.Render(doc, renderer)
}

func organiseText(text string) []string {
	splitText := strings.Split(text, "\n")
	organisedText := []string{}
	currentSegment := ""
	collectingCode := false
	for _, line := range splitText {
		if strings.Contains(line, "```") {
			if collectingCode {
				currentSegment += line + "\n"
				organisedText = append(organisedText, currentSegment)
				collectingCode = !collectingCode
				currentSegment = ""
			} else {
				organisedText = append(organisedText, currentSegment)
				collectingCode = !collectingCode
				currentSegment = line + "\n"
			}
		} else {
			currentSegment += line + "\n"
		}
	}
	organisedText = append(organisedText, currentSegment)
	return organisedText
}

func convertTextToHtml(organisedText []string) string {
	htmlOutput := ""
	for _, segment := range organisedText {
		if strings.Contains(segment, "```") {
			code := strings.ReplaceAll(segment, "```\n", "")
			style := styles.Get("pygments")
			formatter := chromaHtml.New(chromaHtml.WithClasses(true))
			lexer := lexers.Get("python")
			iterator, _ := lexer.Tokenise(nil, code)
			var buf bytes.Buffer
			formatter.Format(&buf, style, iterator)
			htmlOutput = htmlOutput + buf.String()
		} else {
			md := []byte(segment)
			html := mdToHTML(md)
			htmlOutput = htmlOutput + string(html)
		}
	}
	return htmlOutput
}

func generateOutput(convertedHtml string) []byte {
	genericTemplate := template.New("generic.html")
	genericTemplate = template.Must(genericTemplate.ParseFiles("templates/generic.html"))
	var buf bytes.Buffer
	genericTemplate.Execute(&buf, convertedHtml)
	return buf.Bytes()
}

func generateCss(style string) []byte {
	formatter := chromaHtml.New(chromaHtml.WithClasses(true))
	var cssBuf bytes.Buffer
	formatter.WriteCSS(&cssBuf, styles.Get("pygments"))
	return cssBuf.Bytes()
}

func main() {

	cwd, _ := os.Getwd()
	orderPath := filepath.Join(cwd, "text/", "order.txt")
	orderText := loadText(orderPath)
	order := extractOrder(orderText)

	organisedText := []string{}
	for _, section := range order {
		textPath := filepath.Join(cwd, "text/", section+".md")
		text := loadText(textPath)
		organisedText = append(organisedText, organiseText(text)...)
	}

	toc := buildToc(organisedText)

	organisedText = append([]string{toc}, organisedText...)

	htmlOutput := convertTextToHtml(organisedText)
	generatedOutput := generateOutput(htmlOutput)
	os.WriteFile("output/output.html", generatedOutput, 0644)

	css := generateCss("pygments")
	os.WriteFile("output/styles.css", css, 0644)
}
