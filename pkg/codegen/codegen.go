package codegen

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/moevis/partial/pkg/tag"
)

type CodeGen struct {
	fset     *token.FileSet
	codePath string
	file     *ast.File
	buf      bytes.Buffer
}

func New() CodeGen {
	return CodeGen{}
}

func (c *CodeGen) ParseFile(file string) error {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, file, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("failed to parse go file: %s, err: %s", file, err)
	}

	c.file = f
	c.fset = fset
	c.codePath = file
	return nil
}

func (c *CodeGen) Generate(typeName string) {
	ast.Inspect(c.file, func(n ast.Node) bool {
		genDecl, ok := n.(*ast.GenDecl)
		if !ok {
			return true
		}
		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}

			if typeName != "" {
				if typeSpec.Name == nil || typeSpec.Name.Name != typeName {
					continue
				}
			}

			tagSet, err := tag.NewTagSet(structType.Fields.List)
			if err != nil {
				return false
			}

			for tag := range tagSet.NegativeSet {
				c.buf.Reset()
				c.buf.WriteString(fmt.Sprintf("// DO NOT EDIT.\npackage %s\n", c.file.Name.Name))
				fields := &ast.FieldList{}
				newAST := newTypeSpec(typeSpec, tag)
				newAST.Type = &ast.StructType{Fields: fields}

				for ix, field := range structType.Fields.List {
					includeField := true
					for _, tagValue := range tagSet.Fields[ix].PartialTagValue {
						if tagValue.Name != tag {
							continue
						}
						if tagValue.Negative() {
							includeField = false
						}
						break
					}
					if includeField {
						var newTag *ast.BasicLit
						if field.Tag != nil {
							newTag = &ast.BasicLit{Value: tagSet.Fields[ix].StringWithoutPartialTag()}
						}
						newField := &ast.Field{Doc: field.Doc, Comment: field.Comment, Names: field.Names, Type: field.Type, Tag: newTag}
						fields.List = append(fields.List, newField)
					}
				}
				fmt.Fprint(&c.buf, "type ")
				format.Node(&c.buf, c.fset, newAST)
				os.WriteFile(filepath.Join(filepath.Dir(c.codePath), fmt.Sprintf("%s.go", strings.ToLower(tag))), c.buf.Bytes(), 0644)
			}

			for tag := range tagSet.PositiveSet {
				c.buf.Reset()
				c.buf.WriteString(fmt.Sprintf("// DO NOT EDIT.\npackage %s\n", c.file.Name.Name))
				fields := &ast.FieldList{}
				newAST := newTypeSpec(typeSpec, tag)
				newAST.Type = &ast.StructType{Fields: fields}

				for ix, field := range structType.Fields.List {
					includeField := false
					for _, tagValue := range tagSet.Fields[ix].PartialTagValue {
						if tagValue.Name != tag {
							continue
						}
						if !tagValue.Negative() {
							includeField = true
						}
						break
					}
					if includeField {
						var newTag *ast.BasicLit
						if field.Tag != nil {
							newTag = &ast.BasicLit{Value: tagSet.Fields[ix].StringWithoutPartialTag()}
						}
						newField := &ast.Field{Doc: field.Doc, Comment: field.Comment, Names: field.Names, Type: field.Type, Tag: newTag}
						fields.List = append(fields.List, newField)
					}
				}
				fmt.Fprint(&c.buf, "type ")
				format.Node(&c.buf, c.fset, newAST)
				os.WriteFile(filepath.Join(filepath.Dir(c.codePath), fmt.Sprintf("%s.go", strings.ToLower(tag))), c.buf.Bytes(), 0644)
			}
		}
		return true
	})
}

func newTypeSpec(origin *ast.TypeSpec, name string) *ast.TypeSpec {
	newAST := ast.TypeSpec{
		Doc:        origin.Doc,
		Name:       ast.NewIdent(name),
		Comment:    origin.Comment,
		TypeParams: origin.TypeParams,
	}
	return &newAST
}

func filterAndReplace(names []*ast.Ident, origin string, dest string) []*ast.Ident {
	newNames := []*ast.Ident{}
	for _, name := range names {
		if origin == name.Name {
			newNames = append(newNames, &ast.Ident{
				Obj: ast.NewObj(name.Obj.Kind, dest),
			})
		} else {
			newNames = append(newNames, name)
		}
	}
	return newNames
}
