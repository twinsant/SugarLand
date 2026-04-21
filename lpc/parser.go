package lpc

import "fmt"

// Parser 语法分析器，将 Token 序列构建为 AST
type Parser struct {
	lexer *Lexer
	cur   Token
	peek  Token
	errs  []string
}

// NewParser 创建一个新的语法分析器
func NewParser(input string) *Parser {
	p := &Parser{
		lexer: NewLexer(input),
	}
	p.advance() // fill cur
	p.advance() // fill peek
	return p
}

// advance 向前移动一个 Token
func (p *Parser) advance() {
	p.cur = p.peek
	p.peek = p.lexer.NextToken()
}

// expect 期望当前 Token 为指定类型，否则记录错误
func (p *Parser) expect(t TokenType) Token {
	tok := p.cur
	if p.cur.Type != t {
		p.errs = append(p.errs, fmt.Sprintf("L%d:C%d: expected %d, got %d (%q)",
			p.cur.Line, p.cur.Col, t, p.cur.Type, p.cur.Literal))
	}
	p.advance()
	return tok
}

// Errors 返回解析过程中的错误列表
func (p *Parser) Errors() []string {
	return p.errs
}

// isType 判断 Token 是否为类型关键字
func (p *Parser) isType(t TokenType) bool {
	return t == TokenIntType || t == TokenStringType || t == TokenVoid || t == TokenObject
}

// isIdent 判断 Token 是否为标识符
func (p *Parser) isIdent(t TokenType) bool {
	return t == TokenIdent
}

// Parse 解析完整的 LPC 程序
func (p *Parser) Parse() *ProgramAST {
	prog := &ProgramAST{
		Functions: make(map[string]*FunctionDecl),
	}

	for p.cur.Type != TokenEOF {
		if p.cur.Type == TokenIllegal {
			p.advance()
			continue
		}

		node := p.parseTopLevel()
		if node != nil {
			if fn, ok := node.(*FunctionDecl); ok {
				prog.Functions[fn.Name] = fn
			} else {
				prog.Globals = append(prog.Globals, node)
			}
		}
	}

	return prog
}

// parseTopLevel 解析顶层声明：函数定义或变量声明
// 区分方式：type name (...) 是函数，type name = ...; 是变量
func (p *Parser) parseTopLevel() Node {
	if !p.isType(p.cur.Type) {
		p.errs = append(p.errs, fmt.Sprintf("L%d:C%d: expected type, got %q",
			p.cur.Line, p.cur.Col, p.cur.Literal))
		p.advance()
		return nil
	}

	typ := p.cur.Literal
	p.advance() // consume type

	if !p.isIdent(p.cur.Type) {
		p.errs = append(p.errs, fmt.Sprintf("L%d:C%d: expected identifier, got %q",
			p.cur.Line, p.cur.Col, p.cur.Literal))
		p.advance()
		return nil
	}

	name := p.cur.Literal
	p.advance() // consume name

	// 现在 cur 是 name 后面的 token
	// 如果是 (，则是函数定义
	if p.cur.Type == TokenLParen {
		return p.parseFunctionBody(typ, name)
	}

	// 否则是变量声明
	var value Node
	if p.cur.Type == TokenAssign {
		p.advance()
		value = p.parseExpression()
	}
	p.expect(TokenSemicolon)

	return &VarDecl{
		Type:  typ,
		Name:  name,
		Value: value,
	}
}

// parseFunctionBody 解析函数体（cur 已经是 (）
func (p *Parser) parseFunctionBody(retType, name string) *FunctionDecl {
	p.expect(TokenLParen)

	var params []Param
	for p.cur.Type != TokenRParen && p.cur.Type != TokenEOF {
		pt := p.cur.Literal
		p.advance()
		pn := p.cur.Literal
		p.advance()
		params = append(params, Param{Type: pt, Name: pn})
		if p.cur.Type == TokenComma {
			p.advance()
		}
	}
	p.expect(TokenRParen)

	body := p.parseBlock()

	return &FunctionDecl{
		ReturnType: retType,
		Name:       name,
		Params:     params,
		Body:       body,
	}
}

// parseVarDecl 解析变量声明: type name [= expr];
func (p *Parser) parseVarDecl() *VarDecl {
	typ := p.cur.Literal
	p.advance()
	name := p.cur.Literal
	p.advance()

	var value Node
	if p.cur.Type == TokenAssign {
		p.advance()
		value = p.parseExpression()
	}
	p.expect(TokenSemicolon)

	return &VarDecl{
		Type:  typ,
		Name:  name,
		Value: value,
	}
}

// parseBlock 解析代码块 { stmts }
func (p *Parser) parseBlock() *BlockStmt {
	p.expect(TokenLBrace)
	block := &BlockStmt{}
	for p.cur.Type != TokenRBrace && p.cur.Type != TokenEOF {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
	}
	p.expect(TokenRBrace)
	return block
}

// parseStatement 解析一条语句
func (p *Parser) parseStatement() Node {
	switch p.cur.Type {
	case TokenIf:
		return p.parseIf()
	case TokenFor:
		return p.parseFor()
	case TokenWhile:
		return p.parseWhile()
	case TokenReturn:
		return p.parseReturn()
	case TokenLBrace:
		return p.parseBlock()
	case TokenIntType, TokenStringType, TokenVoid, TokenObject:
		if p.peek.Type == TokenIdent {
			return p.parseVarDecl()
		}
		fallthrough
	default:
		return p.parseExprOrAssign()
	}
}

// parseIf 解析 if/else 语句
func (p *Parser) parseIf() *IfStmt {
	p.expect(TokenIf)
	p.expect(TokenLParen)
	cond := p.parseExpression()
	p.expect(TokenRParen)
	then := p.parseBlock()

	var elseBlock *BlockStmt
	if p.cur.Type == TokenElse {
		p.advance()
		elseBlock = p.parseBlock()
	}

	return &IfStmt{
		Cond: cond,
		Then: then,
		Else: elseBlock,
	}
}

// parseFor 解析 for 循环
func (p *Parser) parseFor() *ForStmt {
	p.expect(TokenFor)
	p.expect(TokenLParen)

	// init: var decl 或 assign
	var init Node
	if p.isType(p.cur.Type) && p.isIdent(p.peek.Type) {
		init = p.parseVarDecl() // 自带 ;
	} else if p.cur.Type != TokenSemicolon {
		init = p.parseAssignOrExpr()
	} else {
		p.expect(TokenSemicolon)
	}

	// cond
	var cond Node
	if p.cur.Type != TokenSemicolon {
		cond = p.parseExpression()
	}
	p.expect(TokenSemicolon)

	// post: assign 或 expr
	var post Node
	if p.cur.Type != TokenRParen {
		post = p.parseAssignOrExpr()
	}
	p.expect(TokenRParen)

	body := p.parseBlock()

	return &ForStmt{
		Init: init,
		Cond: cond,
		Post: post,
		Body: body,
	}
}

// parseWhile 解析 while 循环
func (p *Parser) parseWhile() *WhileStmt {
	p.expect(TokenWhile)
	p.expect(TokenLParen)
	cond := p.parseExpression()
	p.expect(TokenRParen)
	body := p.parseBlock()
	return &WhileStmt{Cond: cond, Body: body}
}

// parseReturn 解析 return 语句
func (p *Parser) parseReturn() *ReturnStmt {
	p.expect(TokenReturn)
	var value Node
	if p.cur.Type != TokenSemicolon {
		value = p.parseExpression()
	}
	p.expect(TokenSemicolon)
	return &ReturnStmt{Value: value}
}

// parseAssignOrExpr 解析赋值语句或表达式语句 (for 的 init/post 部分)
func (p *Parser) parseAssignOrExpr() Node {
	// 检查 ident = expr
	if p.cur.Type == TokenIdent && p.peek.Type == TokenAssign {
		name := p.cur.Literal
		p.advance() // ident
		p.advance() // =
		value := p.parseExpression()
		if p.cur.Type == TokenSemicolon {
			p.advance()
		}
		return &AssignStmt{Name: name, Value: value}
	}
	// 检查 ident++ 或 ident--
	if p.cur.Type == TokenIdent && (p.peek.Type == TokenInc || p.peek.Type == TokenDec) {
		name := p.cur.Literal
		op := p.peek.Literal
		p.advance() // ident
		p.advance() // ++/--
		if p.cur.Type == TokenSemicolon {
			p.advance()
		}
		return &AssignStmt{
			Name:  name,
			Value: &BinaryExpr{Op: op, Left: &IdentExpr{Name: name}, Right: &IntLiteral{Value: 1}},
		}
	}

	expr := p.parseExpression()
	if p.cur.Type == TokenSemicolon {
		p.advance()
	}
	return &ExprStmt{Expr: expr}
}

// parseExprOrAssign 解析表达式语句或赋值语句
func (p *Parser) parseExprOrAssign() Node {
	if p.cur.Type == TokenIdent && p.peek.Type == TokenAssign {
		name := p.cur.Literal
		p.advance() // ident
		p.advance() // =
		value := p.parseExpression()
		p.expect(TokenSemicolon)
		return &AssignStmt{Name: name, Value: value}
	}

	expr := p.parseExpression()

	if p.cur.Type == TokenAssign {
		p.advance()
		value := p.parseExpression()
		p.expect(TokenSemicolon)
		if ident, ok := expr.(*IdentExpr); ok {
			return &AssignStmt{Name: ident.Name, Value: value}
		}
		return &ExprStmt{Expr: expr}
	}

	if p.cur.Type == TokenInc || p.cur.Type == TokenDec {
		op := p.cur.Literal
		p.advance()
		p.expect(TokenSemicolon)
		if ident, ok := expr.(*IdentExpr); ok {
			return &AssignStmt{
				Name:  ident.Name,
				Value: &BinaryExpr{Op: op, Left: expr, Right: &IntLiteral{Value: 1}},
			}
		}
	}

	p.expect(TokenSemicolon)
	return &ExprStmt{Expr: expr}
}

// --- 表达式解析 (优先级递归下降) ---

func (p *Parser) parseExpression() Node {
	return p.parseOr()
}

func (p *Parser) parseOr() Node {
	left := p.parseAnd()
	for p.cur.Type == TokenOr {
		op := p.cur.Literal
		p.advance()
		right := p.parseAnd()
		left = &BinaryExpr{Op: op, Left: left, Right: right}
	}
	return left
}

func (p *Parser) parseAnd() Node {
	left := p.parseEquality()
	for p.cur.Type == TokenAnd {
		op := p.cur.Literal
		p.advance()
		right := p.parseEquality()
		left = &BinaryExpr{Op: op, Left: left, Right: right}
	}
	return left
}

func (p *Parser) parseEquality() Node {
	left := p.parseComparison()
	for p.cur.Type == TokenEq || p.cur.Type == TokenNeq {
		op := p.cur.Literal
		p.advance()
		right := p.parseComparison()
		left = &BinaryExpr{Op: op, Left: left, Right: right}
	}
	return left
}

func (p *Parser) parseComparison() Node {
	left := p.parseAddition()
	for p.cur.Type == TokenLt || p.cur.Type == TokenGt ||
		p.cur.Type == TokenLe || p.cur.Type == TokenGe {
		op := p.cur.Literal
		p.advance()
		right := p.parseAddition()
		left = &BinaryExpr{Op: op, Left: left, Right: right}
	}
	return left
}

func (p *Parser) parseAddition() Node {
	left := p.parseMultiplication()
	for p.cur.Type == TokenPlus || p.cur.Type == TokenMinus {
		op := p.cur.Literal
		p.advance()
		right := p.parseMultiplication()
		left = &BinaryExpr{Op: op, Left: left, Right: right}
	}
	return left
}

func (p *Parser) parseMultiplication() Node {
	left := p.parseUnary()
	for p.cur.Type == TokenStar || p.cur.Type == TokenSlash || p.cur.Type == TokenPercent {
		op := p.cur.Literal
		p.advance()
		right := p.parseUnary()
		left = &BinaryExpr{Op: op, Left: left, Right: right}
	}
	return left
}

func (p *Parser) parseUnary() Node {
	if p.cur.Type == TokenNot || p.cur.Type == TokenMinus {
		op := p.cur.Literal
		p.advance()
		expr := p.parseUnary()
		return &UnaryExpr{Op: op, Expr: expr}
	}
	return p.parsePostfix()
}

func (p *Parser) parsePostfix() Node {
	left := p.parsePrimary()

	for {
		switch p.cur.Type {
		case TokenDot:
			p.advance()
			field := p.cur.Literal
			p.advance()
			left = &MemberExpr{Object: left, Field: field, Arrow: false}
		case TokenArrow:
			p.advance()
			field := p.cur.Literal
			p.advance()
			left = &MemberExpr{Object: left, Field: field, Arrow: true}
		case TokenLParen:
			p.advance()
			var args []Node
			for p.cur.Type != TokenRParen && p.cur.Type != TokenEOF {
				args = append(args, p.parseExpression())
				if p.cur.Type == TokenComma {
					p.advance()
				}
			}
			p.expect(TokenRParen)
			left = &CallExpr{Function: left, Args: args}
		case TokenLBracket:
			p.advance()
			index := p.parseExpression()
			p.expect(TokenRBracket)
			left = &IndexExpr{Object: left, Index: index}
		default:
			return left
		}
	}
}

func (p *Parser) parsePrimary() Node {
	switch p.cur.Type {
	case TokenInt:
		val := 0
		fmt.Sscanf(p.cur.Literal, "%d", &val)
		p.advance()
		return &IntLiteral{Value: val}

	case TokenString:
		val := p.cur.Literal
		p.advance()
		return &StringLiteral{Value: val}

	case TokenIdent:
		name := p.cur.Literal
		p.advance()
		return &IdentExpr{Name: name}

	case TokenThis:
		p.advance()
		return &IdentExpr{Name: "this"}

	case TokenLParen:
		p.advance()
		expr := p.parseExpression()
		p.expect(TokenRParen)
		return expr

	default:
		p.errs = append(p.errs, fmt.Sprintf("L%d:C%d: unexpected token %q in expression",
			p.cur.Line, p.cur.Col, p.cur.Literal))
		p.advance()
		return &IntLiteral{Value: 0}
	}
}
