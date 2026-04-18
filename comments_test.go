package i18n

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestPublicPackagesHavePackageComments(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		dir    string
		prefix string
	}{
		{name: "i18n", dir: ".", prefix: "Package i18n "},
		{name: "middleware", dir: "middleware", prefix: "Package middleware "},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			files := parsePackageFiles(t, tc.dir)
			doc := strings.TrimSpace(packageDoc(files))
			if doc == "" {
				t.Fatalf("%s package is missing a package comment", tc.dir)
			}
			if !strings.HasPrefix(doc, tc.prefix) {
				t.Fatalf("%s package comment = %q, want prefix %q", tc.dir, doc, tc.prefix)
			}
		})
	}
}

func TestExportedDeclarationsHaveDocComments(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		dir  string
	}{
		{name: "i18n", dir: "."},
		{name: "middleware", dir: "middleware"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			checkExportedDocs(t, tc.dir)
		})
	}
}

func checkExportedDocs(t *testing.T, dir string) {
	t.Helper()

	for _, file := range parsePackageFiles(t, dir) {
		for _, decl := range file.Decls {
			switch decl := decl.(type) {
			case *ast.FuncDecl:
				if decl.Name.IsExported() {
					requireDocPrefix(t, dir, decl.Name.Name, decl.Doc)
				}
			case *ast.GenDecl:
				checkGenDeclDocs(t, dir, decl)
			}
		}
	}
}

func checkGenDeclDocs(t *testing.T, dir string, decl *ast.GenDecl) {
	t.Helper()

	if decl.Tok == token.TYPE {
		for _, spec := range decl.Specs {
			ts := spec.(*ast.TypeSpec)
			if !ts.Name.IsExported() {
				continue
			}
			requireDocPrefix(t, dir, ts.Name.Name, firstCommentGroup(ts.Doc, decl.Doc))
			st, ok := ts.Type.(*ast.StructType)
			if !ok {
				continue
			}
			for _, field := range st.Fields.List {
				for _, name := range field.Names {
					if name.IsExported() {
						requireDocPrefix(t, dir, name.Name, firstCommentGroup(field.Doc, field.Comment))
					}
				}
			}
		}
		return
	}

	if decl.Tok != token.CONST && decl.Tok != token.VAR {
		return
	}

	for _, spec := range decl.Specs {
		vs := spec.(*ast.ValueSpec)
		for _, name := range vs.Names {
			if name.IsExported() {
				requireDocPrefix(t, dir, name.Name, firstCommentGroup(vs.Doc, vs.Comment, decl.Doc))
			}
		}
	}
}

func requireDocPrefix(t *testing.T, dir, name string, group *ast.CommentGroup) {
	t.Helper()

	text := strings.TrimSpace(commentText(group))
	if text == "" {
		t.Errorf("%s: %s is missing a doc comment", dir, name)
		return
	}
	if !strings.HasPrefix(text, name+" ") && !strings.HasPrefix(text, name+".") {
		t.Errorf("%s: %s doc comment must start with its name, got %q", dir, name, text)
	}
}

func firstCommentGroup(groups ...*ast.CommentGroup) *ast.CommentGroup {
	for _, group := range groups {
		if group != nil && strings.TrimSpace(commentText(group)) != "" {
			return group
		}
	}
	return nil
}

func commentText(group *ast.CommentGroup) string {
	if group == nil {
		return ""
	}
	return group.Text()
}

func packageDoc(files []*ast.File) string {
	for _, file := range files {
		if text := strings.TrimSpace(commentText(file.Doc)); text != "" {
			return text
		}
	}
	return ""
}

func parsePackageFiles(t *testing.T, dir string) []*ast.File {
	t.Helper()

	matches, err := filepath.Glob(filepath.Join(dir, "*.go"))
	if err != nil {
		t.Fatalf("Glob(%s): %v", dir, err)
	}

	var paths []string
	for _, path := range matches {
		base := filepath.Base(path)
		if strings.HasSuffix(base, "_test.go") || strings.HasPrefix(base, "zz_generated_") || strings.HasSuffix(base, ".pb.go") {
			continue
		}
		paths = append(paths, path)
	}
	slices.Sort(paths)

	fset := token.NewFileSet()
	files := make([]*ast.File, 0, len(paths))
	for _, path := range paths {
		file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			t.Fatalf("ParseFile(%s): %v", path, err)
		}
		files = append(files, file)
	}
	if len(files) == 0 {
		t.Fatalf("no package files found in %s", dir)
	}
	return files
}
