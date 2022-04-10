package codegen_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/moevis/partial/pkg/codegen"
	"github.com/stretchr/testify/assert"
)

var code = fmt.Sprintf(`
package testa

type Person struct {
	Age int %s
}
`, "`json:\"age\" partial:\"PersonWithAge\"`")

func TestCodeGen(t *testing.T) {
	cg := codegen.New()
	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.go")
	generatedCodeFile := filepath.Join(tmpDir, "personwithage.go")
	defer func() {
		os.Remove(tmpFile)
	}()
	os.WriteFile(tmpFile, []byte(code), 0755)
	cg.ParseFile(tmpFile)
	cg.Generate("Person")
	defer func() {
		// os.Remove(generatedCodeFile)
	}()
	data, err := os.ReadFile(generatedCodeFile)
	assert.NoError(t, err)
	fmt.Println(string(data))
}
