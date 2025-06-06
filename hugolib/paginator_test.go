// Copyright 2019 The Hugo Authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package hugolib

import (
	"fmt"
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestPaginator(t *testing.T) {
	configFile := `
baseURL = "https://example.com/foo/"

[pagination]
pagerSize = 3
path = "thepage"

[languages.en]
weight = 1
contentDir = "content/en"

[languages.nn]
weight = 2
contentDir = "content/nn"

`
	b := newTestSitesBuilder(t).WithConfigFile("toml", configFile)
	var content []string
	for i := range 9 {
		for _, contentDir := range []string{"content/en", "content/nn"} {
			content = append(content, fmt.Sprintf(contentDir+"/blog/page%d.md", i), fmt.Sprintf(`---
title: Page %d
---

Content.
`, i))
		}
	}

	b.WithContent(content...)

	pagTemplate := `
{{ $pag := $.Paginator }}
Total: {{ $pag.TotalPages }}
First: {{ $pag.First.URL }}
Page Number: {{ $pag.PageNumber }}
URL: {{ $pag.URL }}
{{ with $pag.Next }}Next: {{ .URL }}{{ end }}
{{ with $pag.Prev }}Prev: {{ .URL }}{{ end }}
{{ range $i, $e := $pag.Pagers }}
{{ printf "%d: %d/%d  %t" $i $pag.PageNumber .PageNumber (eq . $pag) -}}
{{ end }}
`

	b.WithTemplatesAdded("index.html", pagTemplate)
	b.WithTemplatesAdded("index.xml", pagTemplate)

	b.Build(BuildCfg{})

	b.AssertFileContent("public/index.html",
		"Page Number: 1",
		"0: 1/1  true")

	b.AssertFileContent("public/thepage/2/index.html",
		"Total: 3",
		"Page Number: 2",
		"URL: /foo/thepage/2/",
		"Next: /foo/thepage/3/",
		"Prev: /foo/",
		"1: 2/2  true",
	)

	b.AssertFileContent("public/index.xml",
		"Page Number: 1",
		"0: 1/1  true")
	b.AssertFileContent("public/thepage/2/index.xml",
		"Page Number: 2",
		"1: 2/2  true")

	b.AssertFileContent("public/nn/index.html",
		"Page Number: 1",
		"0: 1/1  true")

	b.AssertFileContent("public/nn/index.xml",
		"Page Number: 1",
		"0: 1/1  true")
}

// Issue 6023
func TestPaginateWithSort(t *testing.T) {
	files := `
-- hugo.toml --
-- content/a/a.md --
-- content/z/b.md --
-- content/x/b.md --
-- content/x/a.md --
-- layouts/home.html --
Paginate: {{ range (.Paginate (sort .Site.RegularPages ".File.Filename" "desc")).Pages }}|{{ .Path }}{{ end }}
`
	b := Test(t, files)

	b.AssertFileContent("public/index.html", "Paginate: |/z/b|/x/b|/x/a|/a/a")
}

// https://github.com/gohugoio/hugo/issues/6797
func TestPaginateOutputFormat(t *testing.T) {
	b := newTestSitesBuilder(t).WithSimpleConfigFile()
	b.WithContent("_index.md", `---
title: "Home"
cascade:
  outputs:
    - JSON
---`)

	for i := range 22 {
		b.WithContent(fmt.Sprintf("p%d.md", i+1), fmt.Sprintf(`---
title: "Page"
weight: %d
---`, i+1))
	}

	b.WithTemplatesAdded("index.json", `JSON: {{ .Paginator.TotalNumberOfElements }}: {{ range .Paginator.Pages }}|{{ .RelPermalink }}{{ end }}:DONE`)
	b.Build(BuildCfg{})

	b.AssertFileContent("public/index.json",
		`JSON: 22
|/p1/index.json|/p2/index.json|
`)

	// This looks odd, so are most bugs.
	b.Assert(b.CheckExists("public/page/1/index.json/index.html"), qt.Equals, false)
	b.Assert(b.CheckExists("public/page/1/index.json"), qt.Equals, false)
	b.AssertFileContent("public/page/2/index.json", `JSON: 22: |/p11/index.json|/p12/index.json`)
}

// Issue 10802
func TestPaginatorEmptyPageGroups(t *testing.T) {
	t.Parallel()

	files := `
-- config.toml --
baseURL = "https://example.com/"
-- content/p1.md --
-- content/p2.md --
-- layouts/index.html --
{{ $empty := site.RegularPages | complement site.RegularPages }}
Len: {{ len $empty }}: Type: {{ printf "%T" $empty }}
{{ $pgs := $empty.GroupByPublishDate "January 2006" }}
{{ $pag := .Paginate $pgs }}
Len Pag: {{ len $pag.Pages }}
`
	b := Test(t, files)

	b.AssertFileContent("public/index.html", "Len: 0", "Len Pag: 0")
}

func TestPaginatorNodePagesOnly(t *testing.T) {
	files := `
-- hugo.toml --
[pagination]
pagerSize = 1
-- content/p1.md --
-- layouts/_default/single.html --
Paginator: {{ .Paginator }}
`
	b, err := TestE(t, files)
	b.Assert(err, qt.IsNotNil)
	b.Assert(err.Error(), qt.Contains, `error calling Paginator: pagination not supported for this page: kind: "page"`)
}

func TestNilPointerErrorMessage(t *testing.T) {
	files := `
-- hugo.toml --	
-- content/p1.md --
-- layouts/_default/single.html --
Home Filename: {{ site.Home.File.Filename }}
`
	b, err := TestE(t, files)
	b.Assert(err, qt.IsNotNil)
	b.Assert(err.Error(), qt.Contains, `single.html:1:22: executing "single.html" – File is nil; wrap it in if or with: {{ with site.Home.File }}{{ .Filename }}{{ end }}`)
}
