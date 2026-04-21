package sim

import (
	"fmt"
	"math/rand"

	"github.com/twinsant/sugarland/lpc"
)

// Citizen 表示 Sugarscape 中的一个公民智能体
type Citizen struct {
	ID       int     `json:"id"`
	X        int     `json:"x"`
	Y        int     `json:"y"`
	Vision   int     `json:"vision"`   // 视觉范围 v ~ U[1,6]
	Metabolism int   `json:"metabolism"` // 代谢率 m ~ U[1,4]
	MaxAge   int     `json:"max_age"`  // 最大年龄 ~ U[60,100]
	Age      int     `json:"age"`      // 当前年龄
	Wealth   int     `json:"wealth"`   // 财富（糖存量）
	Alive    bool    `json:"alive"`    // 是否存活
	LPCObj   *lpc.Object `json:"-"`   // 关联的 LPC 对象（可选）
}

// NewCitizen 创建一个新的公民智能体
// 属性按 SPEC 均匀分布随机初始化
func NewCitizen(id, x, y int, rng *rand.Rand) *Citizen {
	return &Citizen{
		ID:         id,
		X:          x,
		Y:          y,
		Vision:     rng.Intn(6) + 1,    // U[1,6]
		Metabolism: rng.Intn(4) + 1,    // U[1,4]
		MaxAge:     rng.Intn(41) + 60,  // U[60,100]
		Age:        0,
		Wealth:     rng.Intn(21) + 5,   // U[5,25]
		Alive:      true,
	}
}

// HasHeartBeat 检查公民是否有 LPC heart_beat 方法
func (c *Citizen) HasHeartBeat() bool {
	if c.LPCObj == nil {
		return false
	}
	_, ok := c.LPCObj.FindMethod("heart_beat")
	return ok
}

// LoadScript 加载 LPC 脚本到公民
func (c *Citizen) LoadScript(source string) error {
	p := lpc.NewParser(source)
	prog := p.Parse()
	if len(p.Errors()) > 0 {
		return fmt.Errorf("parse errors: %v", p.Errors())
	}
	obj := lpc.NewObject(fmt.Sprintf("citizen_%d", c.ID), fmt.Sprintf("Citizen#%d", c.ID))
	obj.LoadProgram(prog)
	c.LPCObj = obj
	return nil
}

// IsDead 检查公民是否死亡（老死或饿死）
func (c *Citizen) IsDead() bool {
	return !c.Alive || c.Age >= c.MaxAge || c.Wealth <= 0
}

func (c *Citizen) String() string {
	return fmt.Sprintf("Citizen#%d(%d,%d) v=%d m=%d age=%d/%d wealth=%d alive=%v",
		c.ID, c.X, c.Y, c.Vision, c.Metabolism, c.Age, c.MaxAge, c.Wealth, c.Alive)
}
