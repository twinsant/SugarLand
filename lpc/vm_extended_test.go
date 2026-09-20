package lpc_test

import (
	"testing"

	"github.com/twinsant/sugarland/lpc"
)

func loadObj(t *testing.T, src string) *lpc.Object {
	t.Helper()
	p := lpc.NewParser(src)
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}
	obj := lpc.NewObject("test", "test")
	obj.LoadProgram(prog)
	return obj
}

func TestVMFloatIndexLoop(t *testing.T) {
	src := `
float calc() {
    float x = 1.5;
    float y = x * 2.0 + 0.5;
    return y;
}
int pick() {
    int arr = ({10, 20, 30});
    return arr[1] + arr[2];
}
int loopctl() {
    int sum = 0;
    for (int i = 0; i < 10; i++) {
        if (i == 2) { continue; }
        if (i > 5) { break; }
        sum = sum + i;
    }
    while (1) { break; }
    return sum;
}
string fmtf() {
    return sprintf("v=%f", 2.5);
}
`
	obj := loadObj(t, src)
	if v, err := obj.CallMethod("calc", nil); err != nil || v.FloatVal != 3.5 {
		t.Fatalf("calc=%v err=%v", v, err)
	}
	if v, err := obj.CallMethod("pick", nil); err != nil || v.IntVal != 50 {
		t.Fatalf("pick=%v err=%v", v, err)
	}
	if v, err := obj.CallMethod("loopctl", nil); err != nil || v.IntVal != 13 {
		t.Fatalf("loopctl=%v err=%v", v, err)
	}
	if v, err := obj.CallMethod("fmtf", nil); err != nil || v.StrVal != "v=2.5" {
		t.Fatalf("fmtf=%v err=%v", v, err)
	}
}
