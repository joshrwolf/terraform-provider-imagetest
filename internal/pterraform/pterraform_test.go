package pterraform

import (
	"context"
	"os"
	"path"
	"testing"
)

func TestApply(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		content string
	}{
		{
			name: "should do stuff",
			content: `
provider "local" {}

resource "local_file" "example" {
  filename = "${path.module}/example.txt"
  content  = "Hello, Terraform!"
}

output "example" { value = "hi mom"}
`,
		},
	}

	for _, tt := range tests {
		t.Logf("Running test: %v", tt.name)
		dir := t.TempDir()

		cpath := path.Join(dir, "main.tf")
		if err := os.WriteFile(path.Join(dir, "main.tf"), []byte(tt.content), 0644); err != nil {
			t.Fatal(err)
		}
		t.Logf("Writing main.tf to: %s", cpath)

		p := NewPterraform(os.DirFS(dir))

		err := p.Apply(ctx)
		if err != nil {
			t.Fatal(err)
		}
	}
}
