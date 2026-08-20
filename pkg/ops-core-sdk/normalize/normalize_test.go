package normalize

import "testing"

func TestLower(t *testing.T) {
	r := Lower("HELLO")
	if r.Result != "hello" {
		t.Errorf("expected hello, got %s", r.Result)
	}
}

func TestUpper(t *testing.T) {
	r := Upper("hello")
	if r.Result != "HELLO" {
		t.Errorf("expected HELLO, got %s", r.Result)
	}
}

func TestTrim(t *testing.T) {
	r := Trim("  hello  ")
	if r.Result != "hello" {
		t.Errorf("expected hello, got %s", r.Result)
	}
}

func TestSlugify(t *testing.T) {
	r := Slugify("Hello World!")
	if r.Result != "hello-world" {
		t.Errorf("expected hello-world, got %s", r.Result)
	}
}

func TestCamelCase(t *testing.T) {
	r := CamelCase("hello-world-test")
	if r.Result != "helloWorldTest" {
		t.Errorf("expected helloWorldTest, got %s", r.Result)
	}
}

func TestSnakeCase(t *testing.T) {
	r := SnakeCase("HelloWorld")
	if r.Result != "hello_world" {
		t.Errorf("expected hello_world, got %s", r.Result)
	}
}

func TestKebabCase(t *testing.T) {
	r := KebabCase("hello_world")
	if r.Result != "hello-world" {
		t.Errorf("expected hello-world, got %s", r.Result)
	}
}
