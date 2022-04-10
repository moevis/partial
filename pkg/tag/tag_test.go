package tag_test

import (
	"testing"

	"github.com/moevis/partial/pkg/tag"
	"github.com/stretchr/testify/assert"
)

func TestTagParse(t *testing.T) {
	tagStr := "`json:\"name\" partial:\"-Person,APerson,+BPerson\"`"
	structTag, err := tag.ParseTag(tagStr)
	assert.NoError(t, err)
	assert.Len(t, structTag.Tags, 2)

	assert.Equal(t, structTag.Tags[0].Name, "json")
	assert.Equal(t, structTag.Tags[0].Value, "name")
	assert.Equal(t, structTag.Tags[1].Name, "partial")
	assert.Equal(t, structTag.Tags[1].Value, "-Person,APerson,+BPerson")

	assert.Equal(t, structTag.String(), tagStr)

	assert.Equal(t, structTag.Get("json"), "name")
	assert.Equal(t, structTag.Get("partial"), "-Person,APerson,+BPerson")

	structTag.Remove("json")
	assert.Equal(t, structTag.String(), "`partial:\"-Person,APerson,+BPerson\"`")
}

func TestTagValue(t *testing.T) {
	tagValues, err := tag.ParseTagValue("-Person,APerson,+BPerson:CPerson")
	assert.NoError(t, err)

	assert.Len(t, tagValues.Values, 3)

	tagValue := tagValues.Values[0]
	assert.Equal(t, tagValue.Name, "Person")
	assert.Equal(t, tagValue.Sign, "-")
	assert.True(t, tagValue.Negative())

	tagValue = tagValues.Values[1]
	assert.False(t, tagValue.Negative())
	assert.Equal(t, tagValue.RenameAs, "")

	tagValue = tagValues.Values[2]
	assert.Equal(t, tagValue.Name, "BPerson")
	assert.False(t, tagValue.Negative())
	assert.Equal(t, tagValue.RenameAs, "CPerson")
}

type Person struct {
	Name string `partial:"APerson"`
	Age  int    `partial:"-BPerson"`
	Sex  string `partial:"-BPerson"`
}
