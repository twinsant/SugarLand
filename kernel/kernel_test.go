// Package kernel_test 测试 Sugarscape 仿真功能
package kernel_test

import (
	"testing"

	"github.com/twinsant/sugarland/kernel"
)

func TestNewWorld(t *testing.T) {
	config := kernel.DefaultConfig()
	config["init_population"] = 10
	world := kernel.NewWorld(config)

	state := world.GetState()
	if state.Population != 10 {
		t.Errorf("expected population 10, got %d", state.Population)
	}
	if state.Timestep != 0 {
		t.Errorf("expected timestep 0, got %d", state.Timestep)
	}
}

func TestWorldStep(t *testing.T) {
	config := kernel.DefaultConfig()
	config["init_population"] = 10
	world := kernel.NewWorld(config)

	world.Step()
	state := world.GetState()
	if state.Timestep != 1 {
		t.Errorf("expected timestep 1, got %d", state.Timestep)
	}
}

func TestTrading(t *testing.T) {
	config := kernel.DefaultConfig()
	config["init_population"] = 2
	config["enable_trading"] = true
	world := kernel.NewWorld(config)

	// 放两个 citizen 在相邻位置
	citizens := world.GetCitizensList()
	if len(citizens) < 2 {
		t.Skip("need at least 2 citizens")
	}
	citizens[0].X = 10
	citizens[0].Y = 10
	citizens[0].SetAttr("wealth", 50)
	citizens[0].SetAttr("metabolism", 3) // Bull
	citizens[1].X = 11
	citizens[1].Y = 10
	citizens[1].SetAttr("wealth", 5)
	citizens[1].SetAttr("metabolism", 1) // Bear

	// 贸易规则在 LPC mudlib 中实现，通过世界对象调用
	if _, err := world.CallWorldFunc("do_trading", nil); err != nil {
		t.Fatalf("do_trading error: %v", err)
	}
	// 交易是零和的，且相邻双方财富应发生转移
	total := citizens[0].AttrInt("wealth") + citizens[1].AttrInt("wealth")
	if total != 55 {
		t.Errorf("trading should conserve wealth, got total %d", total)
	}
	if citizens[0].Attr("wealth") == 50 && citizens[1].Attr("wealth") == 5 {
		t.Error("expected a trade to occur between adjacent citizens")
	}
}

func TestTradingDisabled(t *testing.T) {
	config := kernel.DefaultConfig()
	config["init_population"] = 2
	config["enable_trading"] = false
	world := kernel.NewWorld(config)

	citizens := world.GetCitizensList()
	if len(citizens) < 2 {
		t.Skip("need at least 2 citizens")
	}
	citizens[0].X = 10
	citizens[0].Y = 10
	citizens[0].SetAttr("wealth", 50)
	citizens[1].X = 11
	citizens[1].Y = 10
	citizens[1].SetAttr("wealth", 5)
	beforeA, beforeB := citizens[0].Attr("wealth"), citizens[1].Attr("wealth")

	if _, err := world.CallWorldFunc("do_trading", nil); err != nil {
		t.Fatalf("do_trading error: %v", err)
	}
	if citizens[0].Attr("wealth") != beforeA || citizens[1].Attr("wealth") != beforeB {
		t.Error("expected no wealth transfer when trading disabled")
	}
}

func TestBreeding(t *testing.T) {
	config := kernel.DefaultConfig()
	config["init_population"] = 2
	world := kernel.NewWorld(config)
	// do_breeding 以 init_population 为人口上限，扩帽让本轮有繁殖空间
	world.ApplyConfig(map[string]interface{}{"init_population": 10})

	citizens := world.GetCitizensList()
	if len(citizens) < 2 {
		t.Skip("need at least 2 citizens")
	}
	// 满足本地 breed 条件：同行相邻（双向视野）、财富≥代谢×5、年龄 15~65
	citizens[0].X = 10
	citizens[0].Y = 10
	citizens[0].SetAttr("age", 20)
	citizens[0].SetAttr("wealth", 50)
	citizens[0].SetAttr("max_age", 80)
	citizens[0].Alive = true
	citizens[1].X = 11
	citizens[1].Y = 10
	citizens[1].SetAttr("age", 20)
	citizens[1].SetAttr("wealth", 50)
	citizens[1].SetAttr("max_age", 80)
	citizens[1].Alive = true

	beforeCount := len(world.GetAliveCitizens())
	if _, err := world.CallWorldFunc("do_breeding", nil); err != nil {
		t.Fatalf("do_breeding error: %v", err)
	}
	afterCount := len(world.GetAliveCitizens())
	if afterCount != beforeCount+1 {
		t.Errorf("expected one offspring: %d -> %d", beforeCount, afterCount)
	}
}

func TestBreedingDisabledInSystemMode(t *testing.T) {
	config := kernel.DefaultConfig()
	config["init_population"] = 2
	config["repopulate_mode"] = "system"
	world := kernel.NewWorld(config)

	citizens := world.GetCitizensList()
	if len(citizens) < 2 {
		t.Skip("need at least 2 citizens")
	}
	citizens[0].X = 10
	citizens[0].Y = 10
	citizens[0].SetAttr("age", 20)
	citizens[0].SetAttr("wealth", 50)
	citizens[0].SetAttr("max_age", 80)
	citizens[1].X = 11
	citizens[1].Y = 10
	citizens[1].SetAttr("age", 20)
	citizens[1].SetAttr("wealth", 50)
	citizens[1].SetAttr("max_age", 80)

	// system 模式下 heart_beat 不执行繁殖；LastBirths 只记录 emit_birth 事件
	world.Step()
	if len(world.LastBirths) != 0 {
		t.Error("expected no breeding births in system mode")
	}
}

func TestScoreboard(t *testing.T) {
	config := kernel.DefaultConfig()
	config["init_population"] = 20
	world := kernel.NewWorld(config)

	// 游戏指标由 mudlib report_stats() 在 heart_beat 末尾经 emit_stat 上报
	world.Step()
	sb := world.Scoreboard
	if sb["population"] != 20 {
		t.Errorf("expected population 20, got %v", sb["population"])
	}
	if sb["timestep"] != 1 {
		t.Errorf("expected timestep 1, got %v", sb["timestep"])
	}
	if v, _ := sb["avg_vision"].(float64); v <= 0 {
		t.Errorf("expected positive avg vision, got %v", sb["avg_vision"])
	}
	if v, _ := sb["total_sugar"].(float64); v <= 0 {
		t.Errorf("expected positive total sugar, got %v", sb["total_sugar"])
	}
}

func TestGiniCoefficient(t *testing.T) {
	config := kernel.DefaultConfig()
	config["init_population"] = 100
	world := kernel.NewWorld(config)

	world.Step()
	gini, _ := world.Scoreboard["gini_coefficient"].(float64)
	if gini < 0 || gini > 1 {
		t.Errorf("gini coefficient out of range: %f", gini)
	}
}

func TestAgentAttachDetach(t *testing.T) {
	config := kernel.DefaultConfig()
	config["init_population"] = 5
	world := kernel.NewWorld(config)

	citizens := world.GetCitizensList()
	id := citizens[0].ID

	err := world.AttachAIAgent(id, "openclaw")
	if err != nil {
		t.Fatalf("attach error: %v", err)
	}

	c := world.GetCitizenByID(id)
	if !c.IsAgentControlled {
		t.Error("expected citizen to be agent-controlled")
	}
	if c.AgentName != "openclaw" {
		t.Errorf("expected agent name 'openclaw', got %q", c.AgentName)
	}

	err = world.DetachAIAgent(id)
	if err != nil {
		t.Fatalf("detach error: %v", err)
	}
	c = world.GetCitizenByID(id)
	if c.IsAgentControlled {
		t.Error("expected citizen to not be agent-controlled after detach")
	}
}

func TestAgentContext(t *testing.T) {
	config := kernel.DefaultConfig()
	config["init_population"] = 10
	world := kernel.NewWorld(config)

	citizens := world.GetCitizensList()
	id := citizens[0].ID

	ctx, err := world.GetAgentContext(id, -1)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if ctx.CitizenID != id {
		t.Errorf("expected citizen_id %d, got %d", id, ctx.CitizenID)
	}
	if ctx.Timestep != 0 {
		t.Errorf("expected timestep 0, got %d", ctx.Timestep)
	}
	if len(ctx.AvailableActions) == 0 {
		t.Error("expected available actions")
	}
}

func TestStepWithPause(t *testing.T) {
	config := kernel.DefaultConfig()
	config["init_population"] = 5
	world := kernel.NewWorld(config)

	citizens := world.GetCitizensList()
	id := citizens[0].ID

	world.AttachAIAgent(id, "test-agent")

	waiting := world.StepWithPause()
	found := false
	for _, cid := range waiting {
		if cid == id {
			found = true
		}
	}
	if !found {
		t.Error("expected agent-controlled citizen to be in waiting list")
	}
}

func TestExecuteCommand(t *testing.T) {
	config := kernel.DefaultConfig()
	config["init_population"] = 5
	world := kernel.NewWorld(config)

	citizens := world.GetCitizensList()
	id := citizens[0].ID

	world.AttachAIAgent(id, "test-agent")
	world.StepWithPause()

	err := world.ExecuteCommand(id, "move", map[string]interface{}{"x": 10.0, "y": 20.0})
	if err != nil {
		t.Fatalf("command error: %v", err)
	}

	c := world.GetCitizenByID(id)
	if c.X != 10 || c.Y != 20 {
		t.Errorf("expected position (10,20), got (%d,%d)", c.X, c.Y)
	}
}

func TestLPCScriptAttach(t *testing.T) {
	config := kernel.DefaultConfig()
	config["init_population"] = 3
	world := kernel.NewWorld(config)

	citizens := world.GetCitizensList()
	id := citizens[0].ID

	src := `void heart_beat() { int x = query_x(); }`
	err := world.AttachAgent(id, src, 1)
	if err != nil {
		t.Fatalf("attach error: %v", err)
	}

	c := world.GetCitizenByID(id)
	if !c.HasHeartBeat() {
		t.Error("expected citizen to have heart_beat")
	}

	err = world.DetachAgent(id)
	if err != nil {
		t.Fatalf("detach error: %v", err)
	}
}

func TestWorldReset(t *testing.T) {
	config := kernel.DefaultConfig()
	config["init_population"] = 10
	world := kernel.NewWorld(config)

	world.Step()
	world.Step()

	world.Reset()
	state := world.GetState()
	if state.Timestep != 0 {
		t.Errorf("expected timestep 0 after reset, got %d", state.Timestep)
	}
	if state.Population != 10 {
		t.Errorf("expected population 10 after reset, got %d", state.Population)
	}
}

// TestDeathCause 验证 mudlib reap() 的死因判定：
// 寿终记为 "age"，饿死记为 "hunger"
func TestDeathCause(t *testing.T) {
	config := kernel.DefaultConfig()
	config["init_population"] = 2
	world := kernel.NewWorld(config)

	citizens := world.GetCitizensList()
	old := citizens[0]
	old.SetAttr("age", old.Attr("max_age")) // 老死
	hungry := citizens[1]
	hungry.SetAttr("age", 10)
	hungry.SetAttr("wealth", 0) // 饿死

	if _, err := world.CallWorldFunc("reap", nil); err != nil {
		t.Fatalf("reap error: %v", err)
	}

	deaths := world.GetSnapshot().Deaths
	causes := map[int]string{}
	for _, d := range deaths {
		causes[d.ID] = d.Reason
	}
	if causes[old.ID] != "age" {
		t.Errorf("expected death cause 'age' for citizen %d, got %q", old.ID, causes[old.ID])
	}
	if causes[hungry.ID] != "hunger" {
		t.Errorf("expected death cause 'hunger' for citizen %d, got %q", hungry.ID, causes[hungry.ID])
	}
}

// TestMudlibDrivesStep 验证一步仿真完全由 LPC mudlib 驱动：
// 采集、代谢、衰老等记账都通过 efun 发生，Go 内核不再隐式执行规则
func TestMudlibDrivesStep(t *testing.T) {
	config := kernel.DefaultConfig()
	config["init_population"] = 10
	world := kernel.NewWorld(config)

	citizens := world.GetCitizensList()
	c0 := citizens[0]

	world.Step()
	snap := world.GetSnapshot()

	if len(snap.Harvests) == 0 {
		t.Error("expected harvest events from mudlib")
	}
	if snap.Consumed <= 0 {
		t.Error("expected metabolism consumed > 0")
	}
	if c0.Attr("age") != 1 {
		t.Errorf("expected citizen age 1 after mudlib metabolize, got %d", c0.Attr("age"))
	}
	if snap.WorldAgent {
		t.Error("default mudlib should not be marked as custom world agent")
	}
}

// TestMudlibCitizenScript 验证挂接的居民 LPC 由世界脚本通过 call_citizen 驱动
func TestMudlibCitizenScript(t *testing.T) {
	config := kernel.DefaultConfig()
	config["init_population"] = 3
	world := kernel.NewWorld(config)

	citizens := world.GetCitizensList()
	id := citizens[0].ID
	targetX := (citizens[0].X + 3) % config.Int("width")

	src := `void heart_beat() { move(query_x() + 3, query_y()); take_cell(query_x(), query_y(), "sugar"); }`
	if err := world.AttachAgent(id, src, 1); err != nil {
		t.Fatalf("attach error: %v", err)
	}

	world.Step()

	c := world.GetCitizenByID(id)
	if c.X%config.Int("width") != targetX {
		t.Errorf("expected scripted citizen at x=%d, got %d", targetX, c.X)
	}
	// 脚本采集 + mudlib metabolize 都应在账上体现
	if c.Attr("age") != 1 {
		t.Errorf("expected metabolize applied by mudlib, age=%d", c.Attr("age"))
	}
}

// TestWorldScriptReplace 验证自定义世界脚本整体替换规则，卸载后恢复默认 mudlib
func TestWorldScriptReplace(t *testing.T) {
	config := kernel.DefaultConfig()
	config["init_population"] = 10
	world := kernel.NewWorld(config)

	citizens := world.GetCitizensList()
	c0 := citizens[0]

	// 自定义脚本只做生长，居民全部静止
	custom := `
void heart_beat() {
    grow_all();
}
void grow_all() {
    set_cell(0, 0, "sugar", 0.0);
}
`
	if err := world.AttachWorldScript(custom); err != nil {
		t.Fatalf("attach world script error: %v", err)
	}
	world.Step()

	if c0.Attr("age") != 0 {
		t.Errorf("custom world script replaced default rules, age=%d", c0.Attr("age"))
	}
	snap := world.GetSnapshot()
	if !snap.WorldAgent {
		t.Error("expected world_agent=true for custom script")
	}
	if world.GetRoom(0, 0).Get("sugar") != 0 {
		t.Error("expected custom grow_all to set cell (0,0) sugar to 0")
	}

	// 卸载恢复默认 mudlib，规则恢复
	world.DetachWorldScript()
	world.Step()
	if c0.Attr("age") != 1 {
		t.Errorf("default mudlib should resume rules after detach, age=%d", c0.Attr("age"))
	}
	if world.GetSnapshot().WorldAgent {
		t.Error("expected world_agent=false after detach")
	}
}

// TestCellObjects 验证格子对象化：每格天生绑定 LPC room 对象，
// bind_cell 可替换区域脚本，call_cell 驱动对象方法
func TestCellObjects(t *testing.T) {
	config := kernel.DefaultConfig()
	config["init_population"] = 5
	world := kernel.NewWorld(config)

	// 每格天生绑定默认蓝图对象
	cell := world.GetRoom(3, 4)
	if cell.LPCObj == nil {
		t.Fatal("expected cell to have bound LPC object")
	}
	if _, ok := cell.LPCObj.FindMethod("create"); !ok {
		t.Error("default cell blueprint should have create method")
	}

	// 世界脚本用 bind_cell 替换格子脚本、call_cell 驱动（含 \" 转义）
	custom := `
void heart_beat() {
    bind_cell(3, 4, "void mark() { set_cell_here(\"visited\", 1.0); }");
    call_cell(3, 4, "mark");
}
`
	if err := world.AttachWorldScript(custom); err != nil {
		t.Fatalf("attach world script error: %v", err)
	}
	world.Step()

	if got := world.GetRoom(3, 4).Get("visited"); got != 1.0 {
		t.Errorf("expected cell script to set visited=1, got %v", got)
	}
	// 换绑后的对象带有新方法、不再有默认蓝图的 short
	obj := world.GetRoom(3, 4).LPCObj
	if _, ok := obj.FindMethod("mark"); !ok {
		t.Error("rebound cell object should have mark method")
	}
	// 未换绑的格子仍是默认蓝图
	if _, ok := world.GetRoom(0, 0).LPCObj.FindMethod("short"); !ok {
		t.Error("untouched cell should keep default blueprint")
	}
}

// TestRoomTopology 验证完整 room 语义：
// exits 连接表由 mudlib create() 建立，contents 由内核在移动时维护，
// walk 沿命名出口移动，query_contents/query_exits/query_exit 提供查询
func TestRoomTopology(t *testing.T) {
	config := kernel.DefaultConfig()
	config["init_population"] = 0
	world := kernel.NewWorld(config)

	// mudlib create() 已建好四邻环面 exits（含边界包裹）
	if dest := world.GetRoom(3, 4).Exits["east"]; dest != [2]int{4, 4} {
		t.Fatalf("east exit = %v, want [4 4]", dest)
	}
	if dest := world.GetRoom(0, 0).Exits["west"]; dest != [2]int{world.Width() - 1, 0} {
		t.Fatalf("west wrap exit = %v, want [%d 0]", dest, world.Width()-1)
	}

	// efun 链路：spawn 入 contents、move 转移、walk 沿 exit、remove 清 contents
	custom := `
void heart_beat() {
    int id = spawn_citizen(5, 5);
    set_cell(0, 0, "spawn_n", to_float(sizeof(query_contents(5, 5))));
    move_citizen(id, 6, 6);
    set_cell(0, 0, "src_n", to_float(sizeof(query_contents(5, 5))));
    set_cell(0, 0, "dst_n", to_float(sizeof(query_contents(6, 6))));
    set_cell(0, 0, "walked", to_float(walk(id, "east")));
    set_cell(0, 0, "walked_x", to_float(query_citizen(id, "x")));
    set_cell(0, 0, "no_exit", to_float(walk(id, "up")));
    remove_citizen(id, "test");
    set_cell(0, 0, "after_rm", to_float(sizeof(query_contents(7, 6))));
    set_cell(0, 0, "n_exits", to_float(sizeof(query_exits(3, 4))));
}
`
	if err := world.AttachWorldScript(custom); err != nil {
		t.Fatalf("attach world script error: %v", err)
	}
	world.Step()

	c := world.GetRoom(0, 0)
	for name, want := range map[string]float64{
		"spawn_n": 1, "src_n": 0, "dst_n": 1,
		"walked": 1, "walked_x": 7, "no_exit": 0,
		"after_rm": 0, "n_exits": 4,
	} {
		if got := c.Get(name); got != want {
			t.Errorf("%s = %v, want %v", name, got, want)
		}
	}
}

// TestRoomCitizenWalk 验证居民对象侧的 walk/query_env efun
func TestRoomCitizenWalk(t *testing.T) {
	config := kernel.DefaultConfig()
	config["init_population"] = 1
	world := kernel.NewWorld(config)
	c := world.GetCitizenByID(1)
	if c == nil {
		t.Fatal("no citizen spawned")
	}
	x0, y0 := c.X, c.Y

	if err := world.AttachAgent(1, `
void heart_beat() {
    int env = query_env();
    walk("south");
}
`, 1); err != nil {
		t.Fatalf("attach agent error: %v", err)
	}
	world.Step()

	if c.X != x0 || c.Y != y0+1 {
		t.Errorf("after walk south: (%d,%d), want (%d,%d)", c.X, c.Y, x0, y0+1)
	}
	// contents 同步：原 room 移除、新 room 包含
	if world.GetRoom(x0, y0).Contents[1] {
		t.Error("old room should not contain citizen after walk")
	}
	if !world.GetRoom(c.X, c.Y).Contents[1] {
		t.Error("new room should contain citizen after walk")
	}
}
