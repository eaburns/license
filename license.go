// License adds the MIT LICENSE file, an AUTHORS file, and copyright notices to a code repository.
package main

import (
	"os"
	"text/template"
	"path"
	"time"
	"io/ioutil"
	"bufio"
)

type Copyright struct{
	Year int
	ProjectName string
}

func main() {
	if len(os.Args) < 2 {
		os.Stderr.WriteString("usage: license <project name> <author name and email>*\n")
		os.Exit(1)
	}

	cr := Copyright{
		Year: time.Now().Year(),
		ProjectName: os.Args[1],
	}

	var authors []string
	if len(os.Args) > 2 {
		authors = os.Args[2:]
	}

	writeLicenseFile(cr)
	writeAuthorsFile(authors)
	addCopyrights(cr)
}

func writeLicenseFile(cr Copyright) {
	f, err := os.OpenFile("LICENSE", os.O_WRONLY|os.O_CREATE|os.O_EXCL, os.ModePerm)
	if os.IsExist(err) {
		os.Stdout.WriteString("LICENSE file exists, skipping\n")
		return
	}
	if err != nil {
		panic(err)
	}
	defer f.Close()
	os.Stdout.WriteString("Writing LICENSE file\n")
	err = template.Must(template.New("").Parse(licenseFile)).Execute(f, cr)
	if err != nil {
		panic(err)
	}
}

func writeAuthorsFile(as []string) {
	f, err := os.OpenFile("AUTHORS", os.O_WRONLY|os.O_CREATE|os.O_EXCL, os.ModePerm)
	if os.IsExist(err) {
		os.Stdout.WriteString("AUTHORS file exists, skipping\n")
		return
	}
	if err != nil {
		panic(err)
	}
	defer f.Close()
	os.Stdout.WriteString("Writing AUTHORS file\n")
	err = template.Must(template.New("").Parse(authorsFile)).Execute(f, as)
	if err != nil {
		panic(err)
	}
}

func addCopyrights(cr Copyright) {
	for _, f := range sourceFiles(".") {
		addCopyright(cr, f)
	}
}

func addCopyright(cr Copyright, p string) {
	f, err := os.Open(p)
	if err != nil {
		panic(err)
	}

	if hasCopyright(f) {
		os.Stdout.WriteString(p + " has a copyright, skipping\n")
		f.Close()
		return
	}

	contents, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}
	f.Close()

	tmp, err := ioutil.TempFile("", "license")
	if err != nil {
		panic(err)
	}

	info := struct{
		Copyright
		Comment
	} {
		Copyright: cr,
		Comment: commentStyle[path.Ext(p)],
	}
	err = template.Must(template.New("").Parse(copyrightComment)).Execute(tmp, info)
	if err != nil {
		panic(err)
	}
	if _, err = tmp.Write(contents); err != nil {
		panic(err)
	}
	tname := tmp.Name()
	tmp.Close()

	os.Stdout.WriteString(p + "\n")
	os.Rename(tname, p)
}

func hasCopyright(f *os.File) bool {
	defer func() {
		if _, err := f.Seek(0, 0); err != nil {
			panic(err)
		}
	}()

	s := bufio.NewScanner(f)
	s.Split(bufio.ScanRunes)
	for s.Scan() {
		if s.Text() == "\n" {
			break
		} else if s.Text() == "©" {
			return true
		}
	}
	return false
}

type Comment struct {
	Prefix, Suffix string
}

var commentStyle = map[string]Comment {
	".go": Comment{ "// ", "\n" },
	".cpp": Comment{ "// ", "\n" },
	".sh": Comment{ "# ", "\n" },
	".bash": Comment{ "# ", "\n" },
	".c": Comment{ "/* ", "*/\n" },
}

func sourceFiles(dir string) (ents []string) {
	f, err := os.Open(dir)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	names, err := f.Readdirnames(0)
	if err != nil {
		panic(err)
	}

	for _, ent := range names {
		p := path.Join(dir, ent)

		fi, err := os.Stat(p)
		if err != nil {
			panic(err)
		}

		if fi.IsDir() {
			ents = append(ents, sourceFiles(p)...)
			continue
		}

		if _, ok := commentStyle[path.Ext(ent)]; ok {
			ents = append(ents, p)
		}
	}
	return ents
}

var licenseFile =
`Copyright © {{.Year}} the {{.ProjectName}} Authors

Permission is hereby granted, free of charge, to any person obtaining
a copy of this software and associated documentation files (the
"Software"), to deal in the Software without restriction, including
without limitation the rights to use, copy, modify, merge, publish,
distribute, sublicense, and/or sell copies of the Software, and to
permit persons to whom the Software is furnished to do so, subject to
the following conditions:

The above copyright notice and this permission notice shall be
included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY
CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT,
TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
`

var authorsFile =
`{{range .}}{{.}}
{{end}}`

var copyrightComment = `{{.Prefix}}© {{.Year}} the {{.ProjectName}} Authors under the MIT license. See AUTHORS for the list of authors.{{.Suffix}}
`
