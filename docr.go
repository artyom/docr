package main

import (
	"errors"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/artyom/autoflags"
	"github.com/russross/blackfriday"
	"github.com/speedata/gogit"
)

var config = struct {
	Repo      string `flag:"repo,path to repository root (.git directory)"`
	Reference string `flag:"ref,reference (HEAD, refs/heads/develop, etc.)"`
	Listen    string `flag:"bind,address to listen"`
}{
	Repo:      "./.git",
	Reference: "HEAD",
	Listen:    "127.0.0.1:8080",
}

func main() {
	if err := autoflags.Define(&config); err != nil {
		log.Fatal(err)
	}
	flag.Parse()

	repository, err := gogit.OpenRepository(config.Repo)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", PlugHandler(repository, config.Reference))
	log.Print("listening at ", config.Listen)
	log.Fatal(http.ListenAndServe(config.Listen, nil))
}

func PlugHandler(repo *gogit.Repository, ref string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		reference, err := repo.LookupReference(ref)
		if err != nil {
			log.Print(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		commit, err := repo.LookupCommit(reference.Oid)
		if err != nil {
			log.Print(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		path := strings.Trim(r.URL.Path, "/")
		if len(path) == 0 {
			if err := WriteTreeListing(w, commit.Tree); err != nil {
				log.Print(err)
			}
			return
		}
		te, err := EntryByFullName(repo, commit.Tree, path)
		if err != nil {
			log.Print(err)
			switch err {
			case ErrNotFound:
				http.NotFound(w, r)
			default:
				http.Error(w, err.Error(),
					http.StatusInternalServerError)
			}
			return
		}

		switch te.Type {
		case gogit.ObjectBlob:
			// see outside switch
		case gogit.ObjectTree:
			if !strings.HasSuffix(r.URL.Path, "/") {
				http.Redirect(w, r, r.URL.Path+"/", http.StatusFound)
				return
			}
			tree, err := repo.LookupTree(te.Id)
			if err != nil {
				log.Print(err)
				http.Error(w, err.Error(),
					http.StatusInternalServerError)
				return
			}
			if err := WriteTreeListing(w, tree); err != nil {
				log.Print(err)
			}
			return
		default:
			http.Error(w, fmt.Sprintf("unsupported item type: %s", te.Type),
				http.StatusInternalServerError)
			return
		}

		blob, err := repo.LookupBlob(te.Id)

		switch {
		case strings.HasSuffix(te.Name, ".md"), strings.HasSuffix(te.Name, ".markdown"):
			w.Header().Set("Content-Type", "text/html; charset=utf8")
			pageTemplate.Execute(w, template.HTML(Markdown(blob.Contents())))
			//w.Write(Markdown(blob.Contents()))
		default:
			w.Write(blob.Contents())
		}
	}
}

func EntryByFullName(repo *gogit.Repository, tree *gogit.Tree, path string) (*gogit.TreeEntry, error) {
	var item *gogit.TreeEntry
	var err error
	for _, p := range strings.Split(path, "/") {
		item = tree.EntryByName(p)
		if item == nil {
			return nil, ErrNotFound
		}
		switch item.Type {
		case gogit.ObjectBlob:
			return item, nil
		case gogit.ObjectTree:
			if tree, err = repo.LookupTree(item.Id); err != nil {
				return nil, err
			}
		default:
			return nil, ErrInvalidType
		}
	}
	if item == nil {
		return nil, ErrNotFound
	}
	return item, nil
}

var (
	ErrNotFound    = errors.New("item not found")
	ErrInvalidType = errors.New("unsupported object type")
)

func WriteTreeListing(w http.ResponseWriter, tree *gogit.Tree) error {
	return indexTemplate.Execute(w, tree)
}

func Markdown(input []byte) []byte {
	// set up the HTML renderer
	htmlFlags := 0
	htmlFlags |= blackfriday.HTML_TOC
	htmlFlags |= blackfriday.HTML_USE_SMARTYPANTS
	htmlFlags |= blackfriday.HTML_SMARTYPANTS_FRACTIONS
	htmlFlags |= blackfriday.HTML_SMARTYPANTS_LATEX_DASHES
	htmlFlags |= blackfriday.HTML_FOOTNOTE_RETURN_LINKS
	htmlFlags |= blackfriday.HTML_GITHUB_BLOCKCODE
	renderer := blackfriday.HtmlRenderer(htmlFlags, "", "")

	// set up the parser
	extensions := 0
	extensions |= blackfriday.EXTENSION_FOOTNOTES
	extensions |= blackfriday.EXTENSION_FENCED_CODE
	extensions |= blackfriday.EXTENSION_NO_INTRA_EMPHASIS
	extensions |= blackfriday.EXTENSION_TABLES
	extensions |= blackfriday.EXTENSION_AUTOLINK
	extensions |= blackfriday.EXTENSION_STRIKETHROUGH
	extensions |= blackfriday.EXTENSION_SPACE_HEADERS
	extensions |= blackfriday.EXTENSION_HEADER_IDS

	return blackfriday.Markdown(input, renderer, extensions)
}

func init() {
	indexTemplate = template.Must(template.New("index").Parse(indexTemplateBody))
	pageTemplate = template.Must(template.New("page").Parse(pageTemplateBody))
}

var indexTemplate *template.Template
var pageTemplate *template.Template

const indexTemplateBody = `{{range .TreeEntries }}
<a href="{{.Name}}{{if eq .Type 0x20}}/{{end}}">{{.Name}}</a><br>
{{end}}
`

const pageTemplateBody = `<!doctype html>
<meta charset=utf-8>
<style type="text/css">
body {
	font-family: "PT Serif", "Droid Serif", serif;
	font-size: 130%;
	line-height: 170%;
	max-width: 45em;
	margin: auto;
	padding-right: 1em;
	padding-left: 1em;
	color: #333;
	background: white;
	text-rendering: optimizeLegibility;
}

@media only screen and (max-device-width:480px) {
	body {
		font-size:110%;
		text-rendering: auto;
	}
}

img {
	display: block;
	margin: 0 auto;
	max-width: 100%;
}

h1 a, h2 a, h3 a, h4 a, h5 a {
	text-decoration: none;
	color: gray;
}

h1 a:hover, h2 a:hover, h3 a:hover, h4 a:hover, h5 a:hover {
	text-decoration: none;
	color: gray;
}

h1, h2, h3, h4, h5 {
	font-family: Georgia, serif;
	font-weight: bold;
	color: gray;
}

h1 {
	font-size: 150%;
}

h2 {
	font-size: 130%;
}

h3 {
	font-size: 110%;
}

h4, h5 {
	font-size: 100%;
	font-style: italic;
}

pre {
	background-color: rgba(200,200,200,0.2);
	color: #1111111;
	padding: 0.5em;
	overflow: auto;
}

code, pre {
	font-size: 90%;
	font-family: "Consolas", "PT Mono", monospace;
}

hr { border:none; text-align:center; color:gray; }
hr:after {
	content:"\2766";
	display:inline-block;
	font-size:1.5em;
}

dt code {
	font-weight: bold;
}
dd p {
	margin-top: 0;
}

nav {
	padding:.5em;
}
</style>
{{.}}
`
