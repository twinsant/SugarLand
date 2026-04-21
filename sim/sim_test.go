// Package sim_test 测试 Sugarscape 仿真功能
package sim_test

import (
	"testing"

	"github.com/twinsant/sugarland/sim"
)

func TestNewWorld(t *testing.T) {
	config := sim.DefaultConfig()
	config.InitPopulation = 10
	world := sim.NewWorld(config)

	state := world.GetState()
	if state.Population != 10 {
		t.Errorf("expected population 10, got %d", state.Population)
	}
	if state.Timestep != 0 {
		t.Errorf("expected timestep 0, got %d", state.Timestep)
	}
}

func TestWorldStep(t *testing.T) {
	config := sim.DefaultConfig()
	config.InitPopulation = 10
	world := sim.NewWorld(config)

	world.Step()
	state := world.GetState()
	if state.Timestep != 1 {
		t.Errorf("expected timestep 1, got %d", state.Timestep)
	}
}

func TestTrading(t *testing.T) {
	config := sim.DefaultConfig()
	config.InitPopulation = 2
	config.EnableTrading = true
	world := sim.NewWorld(config)

	// 放两个 citizen 在相邻位置
	citizens := world.GetCitizensList()
	if len(citizens) < 2 {
		t.Skip("need at least 2 citizens")
	}
	citizens[0].X = 10
	citizens[0].Y = 10
	citizens[0].Wealth = 50
	citizens[0].Metabolism = 3 // Bull
	citizens[1].X = 11
	citizens[1].Y = 10
	citizens[1].Wealth = 5
	citizens[1].Metabolism = 1 // Bear

	results := world.ExecuteTrading()
	// 交易可能或不发生，取决于边际效用差异
	_ = results
}

func TestTradingDisabled(t *testing.T) {
	config := sim.DefaultConfig()
	config.InitPopulation = 5
	config.EnableTrading = false
	world := sim.NewWorld(config)

	results := world.ExecuteTrading()
	if len(results) != 0 {
		t.Errorf("expected no trading when disabled, got %d results", len(results))
	}
}

func TestMating(t *testing.T) {
	config := sim.DefaultConfig()
	config.InitPopulation = 2
	config.EnableMating = true
	world := sim.NewWorld(config)

	citizens := world.GetCitizensList()
	if len(citizens) < 2 {
		t.Skip("need at least 2 citizens")
	}
	// 设置满足繁殖条件
	citizens[0].X = 10
	citizens[0].Y = 10
	citizens[0].Age = 20
	citizens[0].Wealth = 50
	citizens[0].MaxAge = 80
	citizens[0].Alive = true
	citizens[1].X = 11
	citizens[1].Y = 10
	citizens[1].Age = 20
	citizens[1].Wealth = 50
	citizens[1].MaxAge = 80
	citizens[1].Alive = true

	beforeCount := len(world.GetAliveCitizens())
	results := world.ExecuteMating()
	_ = results
	afterCount := len(world.GetAliveCitizens())
	// 如果繁殖成功，人口应该增加
	if afterCount > beforeCount {
		t.Logf("mating produced offspring: %d -> %d", beforeCount, afterCount)
	}
}

func TestMatingDisabled(t *testing.T) {
	config := sim.DefaultConfig()
	config.InitPopulation = 5
	config.EnableMating = false
	world := sim.NewWorld(config)

	results := world.ExecuteMating()
	if len(results) != 0 {
		t.Errorf("expected no mating when disabled, got %d results", len(results))
	}
}

func TestScoreboard(t *testing.T) {
	config := sim.DefaultConfig()
	config.InitPopulation = 20
	world := sim.NewWorld(config)

	sb := world.UpdateScoreboard()
	if sb.Population != 20 {
		t.Errorf("expected population 20, got %d", sb.Population)
	}
	if sb.Timestep != 0 {
		t.Errorf("expected timestep 0, got %d", sb.Timestep)
	}
	if sb.AvgVision <= 0 {
		t.Errorf("expected positive avg vision, got %f", sb.AvgVision)
	}
}

func TestGiniCoefficient(t *testing.T) {
	config := sim.DefaultConfig()
	config.InitPopulation = 100
	world := sim.NewWorld(config)

	sb := world.UpdateScoreboard()
	if sb.GiniCoefficient < 0 || sb.GiniCoefficient > 1 {
		t.Errorf("gini coefficient out of range: %f", sb.GiniCoefficient)
	}
}

func TestAgentAttachDetach(t *testing.T) {
	config := sim.DefaultConfig()
	config.InitPopulation = 5
	world := sim.NewWorld(config)

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
	config := sim.DefaultConfig()
	config.InitPopulation = 10
	world := sim.NewWorld(config)

	citizens := world.GetCitizensList()
	id := citizens[0].ID

	ctx, err := world.GetAgentContext(id)
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
	config := sim.DefaultConfig()
	config.InitPopulation = 5
	world := sim.NewWorld(config)

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
	config := sim.DefaultConfig()
	config.InitPopulation = 5
	world := sim.NewWorld(config)

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
	config := sim.DefaultConfig()
	config.InitPopulation = 3
	world := sim.NewWorld(config)

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
	config := sim.DefaultConfig()
	config.InitPopulation = 10
	world := sim.NewWorld(config)

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

func TestDeathCause(t *testing.T) {
	config := sim.DefaultConfig()
	config.InitPopulation = 1
	world := sim.NewWorld(config)

	citizens := world.GetCitizensList()
	c := citizens[0]

	// 模拟老死
	c.Age = c.MaxAge
	if c.DeathCause() != "age" {
		t.Errorf("expected death cause 'age', got %q", c.DeathCause())
	}

	// 模拟饿死
	c.Age = 10
	c.Wealth = 0
	if c.DeathCause() != "hunger" {
		t.Errorf("expected death cause 'hunger', got %q", c.DeathCause())
	}
}
