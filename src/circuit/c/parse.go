package c

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path"
	"strings"
)

func filterGo(fi os.FileInfo) bool {
	n := fi.Name()
	return len(n) > 0 && strings.HasSuffix(n, ".go") && n[0] != '_'
}

// PkgSet holds a set of parsed packages and the respective file set
type PkgSet struct {
	fset *token.FileSet
	pkgs map[string]*ast.Package
}

const mode = parser.ParseComments

// ParsePkgSet parses a package and returns the result in PkgSet
func ParsePkgSet(gopath, pkg string) (ps *PkgSet, err error) {
	ps = &PkgSet{}
	ps.fset = token.NewFileSet()
	if ps.pkgs, err = parser.ParseDir(ps.fset, path.Join(gopath, pkg), filterGo, mode); err != nil {
		return nil, err
	}
	return ps, nil
}
