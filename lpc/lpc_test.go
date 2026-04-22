// Package lpc_test 测试 LPC 词法分析、语法解析和解释执行的完整链路
package lpc_test

import (
	"os"
	"testing"

	"github.com/twinsant/sugarland/lpc"
)

func TestLexerBasic(t *testing.T) {
	src := `int x = 10;`
	lexer := lpc.NewLexer(src)
	tokens := lexer.Tokenize()

	expected := []lpc.TokenType{
		lpc.TokenIntType, lpc.TokenIdent, lpc.TokenAssign, lpc.TokenInt, lpc.TokenSemicolon, lpc.TokenEOF,
	}
	if len(tokens) != len(expected) {
		t.Fatalf("expected %d tokens, got %d", len(expected), len(tokens))
	}
	for i, tok := range tokens {
		if tok.Type != expected[i] {
			t.Errorf("token[%d]: expected type %d, got %d (%q)", i, expected[i], tok.Type, tok.Literal)
		}
	}
}

func TestLexerString(t *testing.T) {
	src := `"hello world"`
	lexer := lpc.NewLexer(src)
	tok := lexer.NextToken()
	if tok.Type != lpc.TokenString {
		t.Errorf("expected TokenString, got %d", tok.Type)
	}
	if tok.Literal != "hello world" {
		t.Errorf("expected 'hello world', got %q", tok.Literal)
	}
}

func TestLexerComment(t *testing.T) {
	src := `// this is a comment
int x;`
	lexer := lpc.NewLexer(src)
	tok := lexer.NextToken()
	if tok.Type != lpc.TokenIntType {
		t.Errorf("expected TokenIntType after comment, got %d", tok.Type)
	}
}

func TestLexerMultiComment(t *testing.T) {
	src := `/* multi
line */ int x;`
	lexer := lpc.NewLexer(src)
	tok := lexer.NextToken()
	if tok.Type != lpc.TokenIntType {
		t.Errorf("expected TokenIntType after multi-line comment, got %d", tok.Type)
	}
}

func TestParserFunctionDecl(t *testing.T) {
	src := `void foo(int a, string b) { int x = a; }`
	p := lpc.NewParser(src)
	prog := p.Parse()

	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}
	fn, ok := prog.Functions["foo"]
	if !ok {
		t.Fatal("function 'foo' not found")
	}
	if fn.ReturnType != "void" {
		t.Errorf("expected return type 'void', got %q", fn.ReturnType)
	}
	if len(fn.Params) != 2 {
		t.Errorf("expected 2 params, got %d", len(fn.Params))
	}
}

func TestParserIfElse(t *testing.T) {
	src := `void test() { if (x == 1) { write("one"); } else { write("other"); } }`
	p := lpc.NewParser(src)
	prog := p.Parse()

	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}
	if _, ok := prog.Functions["test"]; !ok {
		t.Fatal("function 'test' not found")
	}
}

func TestParserElseIf(t *testing.T) {
	src := `void test(int x) { if (x == 1) { write("one"); } else if (x == 2) { write("two"); } else { write("other"); } }`
	p := lpc.NewParser(src)
	prog := p.Parse()

	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}
	if _, ok := prog.Functions["test"]; !ok {
		t.Fatal("function 'test' not found")
	}
}

func TestParserForLoop(t *testing.T) {
	src := `void test() { for (int i = 0; i < 10; i++) { write("x"); } }`
	p := lpc.NewParser(src)
	prog := p.Parse()

	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}
	if _, ok := prog.Functions["test"]; !ok {
		t.Fatal("function 'test' not found")
	}
}

func TestParserWhileLoop(t *testing.T) {
	src := `void test() { int i = 0; while (i < 10) { i = i + 1; } }`
	p := lpc.NewParser(src)
	prog := p.Parse()

	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}
	if _, ok := prog.Functions["test"]; !ok {
		t.Fatal("function 'test' not found")
	}
}

func TestParserExpression(t *testing.T) {
	src := `int calc(int a, int b) { return a + b * 2; }`
	p := lpc.NewParser(src)
	prog := p.Parse()

	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}
	if _, ok := prog.Functions["calc"]; !ok {
		t.Fatal("function 'calc' not found")
	}
}

func TestVMCallFunction(t *testing.T) {
	src := `int add(int a, int b) { return a + b; }`
	p := lpc.NewParser(src)
	prog := p.Parse()

	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	vm := lpc.NewVM()
	vm.LoadProgram(prog)

	result, err := vm.CallFunc("add", []lpc.Value{lpc.IntValue(3), lpc.IntValue(4)})
	if err != nil {
		t.Fatalf("call error: %v", err)
	}
	if result.IntVal != 7 {
		t.Errorf("expected 7, got %d", result.IntVal)
	}
}

func TestVMIfElse(t *testing.T) {
	src := `int check(int x) { if (x > 5) { return 1; } else { return 0; } }`
	p := lpc.NewParser(src)
	prog := p.Parse()

	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	vm := lpc.NewVM()
	vm.LoadProgram(prog)

	result, _ := vm.CallFunc("check", []lpc.Value{lpc.IntValue(10)})
	if result.IntVal != 1 {
		t.Errorf("expected 1, got %d", result.IntVal)
	}

	result, _ = vm.CallFunc("check", []lpc.Value{lpc.IntValue(3)})
	if result.IntVal != 0 {
		t.Errorf("expected 0, got %d", result.IntVal)
	}
}

func TestVMEfun(t *testing.T) {
	src := `int test() { int x = random(10); return x; }`
	p := lpc.NewParser(src)
	prog := p.Parse()

	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	vm := lpc.NewVM()
	vm.RegisterEfun("random", func(args []lpc.Value) lpc.Value {
		return lpc.IntValue(42)
	})
	vm.LoadProgram(prog)

	result, err := vm.CallFunc("test", nil)
	if err != nil {
		t.Fatalf("call error: %v", err)
	}
	if result.IntVal != 42 {
		t.Errorf("expected 42, got %d", result.IntVal)
	}
}

func TestVMForLoop(t *testing.T) {
	src := `int sum(int n) { int total = 0; for (int i = 1; i <= n; i++) { total = total + i; } return total; }`
	p := lpc.NewParser(src)
	prog := p.Parse()

	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	vm := lpc.NewVM()
	vm.LoadProgram(prog)

	result, err := vm.CallFunc("sum", []lpc.Value{lpc.IntValue(10)})
	if err != nil {
		t.Fatalf("call error: %v", err)
	}
	if result.IntVal != 55 {
		t.Errorf("expected 55, got %d", result.IntVal)
	}
}

func TestObjectSystem(t *testing.T) {
	obj := lpc.NewObject("test_obj", "TestObject")
	obj.SetProperty("name", lpc.StringValue("test"))
	obj.SetProperty("value", lpc.IntValue(42))

	v, ok := obj.GetProperty("name")
	if !ok || v.StrVal != "test" {
		t.Errorf("expected name='test', got %q", v.StrVal)
	}
	v, ok = obj.GetProperty("value")
	if !ok || v.IntVal != 42 {
		t.Errorf("expected value=42, got %d", v.IntVal)
	}
}

func TestObjectManager(t *testing.T) {
	mgr := lpc.NewObjectManager()
	mgr.Create("obj1", "Test")
	mgr.Add("obj2", lpc.NewObject("obj2", "Test2"))

	found, ok := mgr.Find("obj1")
	if !ok || found.Name != "Test" {
		t.Errorf("expected to find obj1")
	}

	mgr.Destroy("obj1")
	_, ok = mgr.Find("obj1")
	if ok {
		t.Errorf("expected obj1 to be destroyed")
	}

	list := mgr.List()
	if len(list) != 1 {
		t.Errorf("expected 1 object in list, got %d", len(list))
	}
}

func TestLoadForagerScript(t *testing.T) {
	data, err := os.ReadFile("testdata/forager.lpc")
	if err != nil {
		t.Skipf("skipping: %v", err)
	}

	p := lpc.NewParser(string(data))
	prog := p.Parse()

	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	fn, ok := prog.Functions["heart_beat"]
	if !ok {
		t.Fatal("heart_beat function not found")
	}
	if fn.ReturnType != "void" {
		t.Errorf("expected void, got %q", fn.ReturnType)
	}
}

func TestLoadTraderScript(t *testing.T) {
	data, err := os.ReadFile("testdata/trader.lpc")
	if err != nil {
		t.Skipf("skipping: %v", err)
	}

	p := lpc.NewParser(string(data))
	prog := p.Parse()

	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	fn, ok := prog.Functions["heart_beat"]
	if !ok {
		t.Fatal("heart_beat function not found")
	}
	if fn.ReturnType != "void" {
		t.Errorf("expected void, got %q", fn.ReturnType)
	}
}

func TestLoadBreederScript(t *testing.T) {
	data, err := os.ReadFile("testdata/breeder.lpc")
	if err != nil {
		t.Skipf("skipping: %v", err)
	}

	p := lpc.NewParser(string(data))
	prog := p.Parse()

	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	fn, ok := prog.Functions["heart_beat"]
	if !ok {
		t.Fatal("heart_beat function not found")
	}
	if fn.ReturnType != "void" {
		t.Errorf("expected void, got %q", fn.ReturnType)
	}
}

func TestEndToEndForager(t *testing.T) {
	src := `void heart_beat() {
		int x = query_x();
		int y = query_y();
		int best_sugar = 0;
		int best_x = x;
		int best_y = y;
		int s;
		s = query_cell_sugar(x, y - 1);
		if (s > best_sugar) { best_sugar = s; best_x = x; best_y = y - 1; }
		move(best_x, best_y);
	}`
	p := lpc.NewParser(src)
	prog := p.Parse()

	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	obj := lpc.NewObject("citizen_1", "Citizen#1")
	obj.LoadProgram(prog)

	currentX, currentY := 10, 10
	obj.VM.RegisterEfun("query_x", func(args []lpc.Value) lpc.Value {
		return lpc.IntValue(currentX)
	})
	obj.VM.RegisterEfun("query_y", func(args []lpc.Value) lpc.Value {
		return lpc.IntValue(currentY)
	})
	obj.VM.RegisterEfun("query_cell_sugar", func(args []lpc.Value) lpc.Value {
		if len(args) >= 2 && args[0].IntVal == 10 && args[1].IntVal == 9 {
			return lpc.IntValue(5)
		}
		return lpc.IntValue(0)
	})
	obj.VM.RegisterEfun("move", func(args []lpc.Value) lpc.Value {
		if len(args) >= 2 {
			currentX = args[0].IntVal
			currentY = args[1].IntVal
		}
		return lpc.Null()
	})
	obj.VM.RegisterEfun("random", func(args []lpc.Value) lpc.Value {
		return lpc.IntValue(0)
	})

	_, err := obj.CallMethod("heart_beat", nil)
	if err != nil {
		t.Fatalf("heart_beat error: %v", err)
	}

	if currentX != 10 || currentY != 9 {
		t.Errorf("expected to move to (10,9), got (%d,%d)", currentX, currentY)
	}
}

// --- Phase 3: Efun Tests ---

func TestEfunExplode(t *testing.T) {
	src := `string test() { string result = implode(explode("a,b,c", ","), "-"); return result; }`
	p := lpc.NewParser(src)
	prog := p.Parse()

	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	vm := lpc.NewVM()
	vm.RegisterDefaultEfuns()
	vm.LoadProgram(prog)

	result, err := vm.CallFunc("test", nil)
	if err != nil {
		t.Fatalf("call error: %v", err)
	}
	if result.StrVal != "a-b-c" {
		t.Errorf("expected 'a-b-c', got %q", result.StrVal)
	}
}

func TestEfunStrlen(t *testing.T) {
	src := `int test() { return strlen("hello"); }`
	p := lpc.NewParser(src)
	prog := p.Parse()

	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	vm := lpc.NewVM()
	vm.RegisterDefaultEfuns()
	vm.LoadProgram(prog)

	result, err := vm.CallFunc("test", nil)
	if err != nil {
		t.Fatalf("call error: %v", err)
	}
	if result.IntVal != 5 {
		t.Errorf("expected 5, got %d", result.IntVal)
	}
}

func TestEfunSprintf(t *testing.T) {
	src := `string test() { return sprintf("age=%d", 25); }`
	p := lpc.NewParser(src)
	prog := p.Parse()

	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	vm := lpc.NewVM()
	vm.RegisterDefaultEfuns()
	vm.LoadProgram(prog)

	result, err := vm.CallFunc("test", nil)
	if err != nil {
		t.Fatalf("call error: %v", err)
	}
	if result.StrVal != "age=25" {
		t.Errorf("expected 'age=25', got %q", result.StrVal)
	}
}

func TestEfunSortArray(t *testing.T) {
	src := `int test() {
		int a = sizeof(explode("3,1,2", ","));
		return a;
	}`
	p := lpc.NewParser(src)
	prog := p.Parse()

	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	vm := lpc.NewVM()
	vm.RegisterDefaultEfuns()
	vm.LoadProgram(prog)

	result, err := vm.CallFunc("test", nil)
	if err != nil {
		t.Fatalf("call error: %v", err)
	}
	if result.IntVal != 3 {
		t.Errorf("expected 3, got %d", result.IntVal)
	}
}

func TestEfunAbs(t *testing.T) {
	src := `int test() { return abs(0 - 42); }`
	p := lpc.NewParser(src)
	prog := p.Parse()

	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	vm := lpc.NewVM()
	vm.RegisterDefaultEfuns()
	vm.LoadProgram(prog)

	result, err := vm.CallFunc("test", nil)
	if err != nil {
		t.Fatalf("call error: %v", err)
	}
	if result.IntVal != 42 {
		t.Errorf("expected 42, got %d", result.IntVal)
	}
}

func TestEfunMinMax(t *testing.T) {
	src := `int test_max() { return max(3, 7); }
	int test_min() { return min(3, 7); }`
	p := lpc.NewParser(src)
	prog := p.Parse()

	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	vm := lpc.NewVM()
	vm.RegisterDefaultEfuns()
	vm.LoadProgram(prog)

	result, err := vm.CallFunc("test_max", nil)
	if err != nil {
		t.Fatalf("call error: %v", err)
	}
	if result.IntVal != 7 {
		t.Errorf("expected 7, got %d", result.IntVal)
	}

	result, err = vm.CallFunc("test_min", nil)
	if err != nil {
		t.Fatalf("call error: %v", err)
	}
	if result.IntVal != 3 {
		t.Errorf("expected 3, got %d", result.IntVal)
	}
}

func TestVMElseIf(t *testing.T) {
	src := `int classify(int x) {
		if (x == 1) { return 10; }
		else if (x == 2) { return 20; }
		else if (x == 3) { return 30; }
		else { return 0; }
	}`
	p := lpc.NewParser(src)
	prog := p.Parse()

	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	vm := lpc.NewVM()
	vm.LoadProgram(prog)

	tests := []struct {
		input, expected int
	}{
		{1, 10}, {2, 20}, {3, 30}, {5, 0},
	}
	for _, tt := range tests {
		result, err := vm.CallFunc("classify", []lpc.Value{lpc.IntValue(tt.input)})
		if err != nil {
			t.Fatalf("call error for input %d: %v", tt.input, err)
		}
		if result.IntVal != tt.expected {
			t.Errorf("classify(%d): expected %d, got %d", tt.input, tt.expected, result.IntVal)
		}
	}
}

func TestVMEfunWithObjManager(t *testing.T) {
	src := `int test() {
		all_inventory("dummy");
		return 42;
	}`
	p := lpc.NewParser(src)
	prog := p.Parse()

	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	vm := lpc.NewVM()
	vm.ObjManager = lpc.NewObjectManager()
	vm.RegisterDefaultEfuns()
	vm.LoadProgram(prog)

	result, err := vm.CallFunc("test", nil)
	if err != nil {
		t.Fatalf("call error: %v", err)
	}
	if result.IntVal != 42 {
		t.Errorf("expected 42, got %d", result.IntVal)
	}
}
