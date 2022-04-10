package tag

import (
	"fmt"
	"go/ast"
	"strings"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

var tagLexer = lexer.MustSimple([]lexer.Rule{
	{`Ident`, `[a-zA-Z][a-zA-Z_\d]*`, nil},
	{`String`, `"(?:\\.|[^"])*"`, nil},
	{`Punct`, `:`, nil},
	{"whitespace", `\s+`, nil},
})

var tagParser = participle.MustBuild(&StructTag{}, participle.Lexer(tagLexer), participle.Unquote(), participle.Elide("whitespace"))

func ParseTag(tagStr string) (StructTag, error) {
	st := StructTag{Original: tagStr}
	tagStr = strings.Trim(tagStr, "`")
	err := tagParser.Parse("", strings.NewReader(tagStr), &st)
	if err != nil {
		return st, fmt.Errorf("failed to parse tag string: %s, err: %w", tagStr, err)
	}
	return st, nil
}

type StructTag struct {
	Original        string
	PartialTagValue []TagValue
	Tags            []*Tag `@@*`
}

func (s *StructTag) String() string {
	tagStr := ""
	for _, tag := range s.Tags {
		tagStr += " " + tag.String()
	}
	return fmt.Sprintf("`%s`", strings.TrimSpace(tagStr))
}

func (s *StructTag) StringWithoutPartialTag() string {
	tagStr := ""
	for _, tag := range s.Tags {
		if tag.Name != PartialTag {
			tagStr += " " + tag.String()
		}
	}
	if tagStr == "" {
		return ""
	}
	return fmt.Sprintf("`%s`", strings.TrimSpace(tagStr))
}

func (s *StructTag) Remove(tagName string) {
	newTags := []*Tag{}
	for _, t := range s.Tags {
		if t.Name == tagName {
			continue
		}
		newTags = append(newTags, t)
	}
	s.Tags = newTags
}

func (s *StructTag) Get(name string) string {
	for _, tag := range s.Tags {
		return tag.Value
	}
	return ""
}

type Tag struct {
	Name  string `@Ident ":"`
	Value string `@String`
}

func (t Tag) String() string {
	return fmt.Sprintf("%s:\"%s\"", t.Name, t.Value)
}

type TagValues struct {
	Values []TagValue `(@@ ("," @@)*)*`
}

type TagValue struct {
	Sign     string `(@("-"|"+"))?`
	Name     string `@Ident`
	RenameAs string `(":" @Ident)?`
}

func (t TagValue) Negative() bool {
	if t.Sign == "-" {
		return true
	}
	return false
}

var tagValueLexer = lexer.MustSimple([]lexer.Rule{
	{"Ident", `[a-zA-Z][a-zA-Z_\d]*`, nil},
	{`Punct`, `[:,\+\-]`, nil},
	{"whitespace", `\s+`, nil},
})

var tagValueParser = participle.MustBuild(&TagValues{}, participle.Lexer(tagValueLexer), participle.Elide("whitespace"))

func ParseTagValue(tagValueStr string) (TagValues, error) {
	tv := TagValues{}
	err := tagValueParser.Parse("", strings.NewReader(tagValueStr), &tv)
	if err != nil {
		return tv, fmt.Errorf("failed to parse tag value: %s, err: %w", tagValueStr, err)
	}
	return tv, err
}

type TagSet struct {
	NegativeSet map[string]struct{}
	PositiveSet map[string]struct{}
	Fields      []*StructTag
}

var PartialTag = "partial"

func NewTagSet(fields []*ast.Field) (TagSet, error) {
	tagSet := TagSet{
		NegativeSet: map[string]struct{}{},
		PositiveSet: map[string]struct{}{},
		Fields:      []*StructTag{},
	}
	for _, field := range fields {
		structTag := StructTag{Tags: []*Tag{}}
		if field.Tag != nil {
			structTag, _ = ParseTag(field.Tag.Value)
		}
		partialTag := structTag.Get(PartialTag)
		if partialTag == "" {
			continue
		}
		structNames, err := ParseTagValue(partialTag)
		if err != nil {
			return tagSet, fmt.Errorf("failed to parse tag: %s, err: %w", structTag.Original, err)
		}
		for _, tag := range structNames.Values {
			if tag.Name == "" {
				continue
			}
			if tag.Negative() {
				if _, exist := tagSet.PositiveSet[tag.Name]; !exist {
					tagSet.NegativeSet[tag.Name] = struct{}{}
				}
			} else {
				tagSet.PositiveSet[tag.Name] = struct{}{}
				delete(tagSet.NegativeSet, tag.Name)
			}
		}
		tagSet.Fields = append(tagSet.Fields, &structTag)
		structTag.PartialTagValue = structNames.Values
	}
	return tagSet, nil
}
