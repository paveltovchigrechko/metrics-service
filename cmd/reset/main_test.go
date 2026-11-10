package main

import (
	"flag"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var update = flag.Bool("update", false, "rewrite the golden files in testdata")

// copyModule copies testdata/module to a temporary directory, so the
// generated files do not end up in testdata.
func copyModule(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	require.NoError(t, os.CopyFS(dir, os.DirFS(filepath.Join("testdata", "module"))))

	return dir
}

// writeModule creates a module with the given files, keyed by relative path.
func writeModule(t *testing.T, files map[string]string) string {
	t.Helper()

	dir := t.TempDir()
	files["go.mod"] = "module example.com/m\n\ngo 1.22\n"
	for name, content := range files {
		path := filepath.Join(dir, name)
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0755))
		require.NoError(t, os.WriteFile(path, []byte(content), 0644))
	}

	return dir
}

func TestRun_Golden(t *testing.T) {
	dir := copyModule(t)
	require.NoError(t, run(dir))

	for _, pkg := range []string{"a", "b"} {
		t.Run(pkg, func(t *testing.T) {
			got, err := os.ReadFile(filepath.Join(dir, pkg, pkg+genFileSuffix))
			require.NoError(t, err)

			golden := filepath.Join("testdata", "module", pkg, pkg+genFileSuffix+".golden")
			if *update {
				require.NoError(t, os.WriteFile(golden, got, 0644))
			}

			want, err := os.ReadFile(golden)
			require.NoError(t, err)
			assert.Equal(t, string(want), string(got))
		})
	}

	_, err := os.Stat(filepath.Join(dir, "x", "b", "b"+genFileSuffix))
	assert.ErrorIs(t, err, os.ErrNotExist, "no file for a package without marked structs")
}

// TestRun_GeneratedCodeWorks compiles the generated code and runs the tests in
// testdata/module/a/reset_test.go against it.
func TestRun_GeneratedCodeWorks(t *testing.T) {
	if testing.Short() {
		t.Skip("runs go test on the generated code")
	}

	dir := copyModule(t)
	require.NoError(t, run(dir))

	cmd := exec.Command("go", "test", "./...")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GOWORK=off", "GOFLAGS=-mod=mod")
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "go test on the generated code:\n%s", out)
}

func TestRun_RewritesStaleFile(t *testing.T) {
	dir := writeModule(t, map[string]string{
		"p/p.go": "package p\n\n// generate:reset\ntype S struct{ labels []string }\n",
		// Generated before the field was renamed: it no longer compiles.
		"p/p.reset.gen.go": "package p\n\nfunc (s *S) Reset() { s.tags = s.tags[:0] }\n",
	})

	require.NoError(t, run(dir))

	got, err := os.ReadFile(filepath.Join(dir, "p", "p"+genFileSuffix))
	require.NoError(t, err)
	assert.Contains(t, string(got), "s.labels = s.labels[:0]")
	assert.NotContains(t, string(got), "tags")
}

func TestRun_FailsOnUnresolvedFieldType(t *testing.T) {
	dir := writeModule(t, map[string]string{
		"p/p.go": "package p\n\n// generate:reset\ntype S struct{ items []Missing }\n",
	})

	err := run(dir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "S.items")

	_, statErr := os.Stat(filepath.Join(dir, "p", "p"+genFileSuffix))
	assert.ErrorIs(t, statErr, os.ErrNotExist, "nothing is written on error")
}

func TestRun_NoMarkedStructs(t *testing.T) {
	dir := writeModule(t, map[string]string{
		"p/p.go": "package p\n\ntype S struct{ n int }\n",
	})

	require.NoError(t, run(dir))

	matches, err := filepath.Glob(filepath.Join(dir, "*", "*"+genFileSuffix))
	require.NoError(t, err)
	assert.Empty(t, matches)
}

func TestHasDirective(t *testing.T) {
	tests := []struct {
		name    string
		comment string
		want    bool
	}{
		{"with space", "// generate:reset", true},
		{"without space", "//generate:reset", true},
		{"trailing spaces", "// generate:reset   ", true},
		{"among other lines", "// S is a struct.\n//\n// generate:reset", true},
		{"other comment", "// generate something", false},
		{"longer word", "// generate:resetter", false},
		{"mentioned in text", "// Run generate:reset on it.", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var group ast.CommentGroup
			for _, line := range strings.Split(tt.comment, "\n") {
				group.List = append(group.List, &ast.Comment{Text: line})
			}

			assert.Equal(t, tt.want, hasDirective(&group))
		})
	}

	t.Run("no comment", func(t *testing.T) {
		assert.False(t, hasDirective(nil))
	})
}

func TestResetSpecs(t *testing.T) {
	const src = `package p

// generate:reset
type Single struct{}

type (
	// generate:reset
	InGroup struct{}

	Unmarked struct{}
)

// generate:reset
type (
	GroupDocDoesNotApply struct{}
)

// generate:reset
type NotAStruct int

// generate:reset

type SeparatedByBlankLine struct{}

// generate:reset
func NotAType() {}
`

	f, err := parser.ParseFile(token.NewFileSet(), "p.go", src, parser.ParseComments)
	require.NoError(t, err)

	var names []string
	for _, spec := range resetSpecs(f) {
		names = append(names, spec.Name.Name)
	}

	assert.Equal(t, []string{"Single", "InGroup"}, names)
}

func TestReceiverName(t *testing.T) {
	tests := []struct {
		typeName string
		reserved []string
		want     string
	}{
		{"ResetableStruct", nil, "rs"},
		{"Person", nil, "p"},
		{"item", nil, "i"},
		{"Ärger", nil, "ä"},
		{"IFace", nil, "i"}, // "if" is a keyword.
		{"Box", []string{"b"}, "recv"},
		{"Box", []string{"b", "recv"}, "recv2"},
		{"Box", []string{"b", "recv", "recv2"}, "recv3"},
	}

	for _, tt := range tests {
		t.Run(tt.typeName+"/"+strings.Join(tt.reserved, ","), func(t *testing.T) {
			reserved := make(map[string]bool)
			for _, name := range tt.reserved {
				reserved[name] = true
			}

			assert.Equal(t, tt.want, receiverName(tt.typeName, reserved))
		})
	}
}

func TestZeroBasic(t *testing.T) {
	tests := map[types.BasicKind]string{
		types.Bool:          "false",
		types.String:        `""`,
		types.Int:           "0",
		types.Uint8:         "0",
		types.Float64:       "0",
		types.Complex128:    "0",
		types.UnsafePointer: "nil",
	}

	for kind, want := range tests {
		basic := types.Typ[kind]
		t.Run(basic.Name(), func(t *testing.T) {
			assert.Equal(t, want, zeroBasic(basic))
		})
	}
}

func TestIsStdlib(t *testing.T) {
	assert.True(t, isStdlib("time"))
	assert.True(t, isStdlib("net/http"))
	assert.False(t, isStdlib("example.com/resettest/b"))
	assert.False(t, isStdlib("github.com/go-chi/chi/v5"))
}

func TestOperand(t *testing.T) {
	assert.Equal(t, "s.list", operand("s.list"))
	assert.Equal(t, "(*s.list)", operand("*s.list"))
}

func TestValidType(t *testing.T) {
	invalid := types.Typ[types.Invalid]
	str := types.Typ[types.String]

	tests := []struct {
		name string
		typ  types.Type
		want bool
	}{
		{"basic", str, true},
		{"invalid", invalid, false},
		{"slice of invalid", types.NewSlice(invalid), false},
		{"pointer to slice of invalid", types.NewPointer(types.NewSlice(invalid)), false},
		{"array of invalid", types.NewArray(invalid, 3), false},
		{"map with invalid key", types.NewMap(invalid, str), false},
		{"map with invalid value", types.NewMap(str, invalid), false},
		{"chan of invalid", types.NewChan(types.SendRecv, invalid), false},
		{"map of slices", types.NewMap(str, types.NewSlice(str)), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, validType(tt.typ))
		})
	}
}
