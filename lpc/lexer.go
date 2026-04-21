package lpc

// Lexer 词法分析器，将源代码分解为 Token 序列
type Lexer struct {
	input  string
	pos    int  // 当前字符位置
	line   int  // 当前行号
	col    int  // 当前列号
	ch     byte // 当前字符
}

// NewLexer 创建一个新的词法分析器
func NewLexer(input string) *Lexer {
	l := &Lexer{
		input: input,
		line:  1,
		col:   0,
	}
	l.advance()
	return l
}

// advance 向前移动一个字符
func (l *Lexer) advance() {
	if l.pos >= len(l.input) {
		l.ch = 0
		return
	}
	l.ch = l.input[l.pos]
	l.pos++
	l.col++
}

// peek 查看下一个字符但不移动
func (l *Lexer) peek() byte {
	if l.pos >= len(l.input) {
		return 0
	}
	return l.input[l.pos]
}

// skipWhitespace 跳过空白字符
func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\r' || l.ch == '\n' {
		if l.ch == '\n' {
			l.line++
			l.col = 0
		}
		l.advance()
	}
}

// skipComment 跳过注释
func (l *Lexer) skipComment() {
	if l.ch == '/' && l.peek() == '/' {
		// 单行注释
		for l.ch != 0 && l.ch != '\n' {
			l.advance()
		}
	} else if l.ch == '/' && l.peek() == '*' {
		// 多行注释
		l.advance() // skip /
		l.advance() // skip *
		for l.ch != 0 {
			if l.ch == '*' && l.peek() == '/' {
				l.advance() // skip *
				l.advance() // skip /
				return
			}
			if l.ch == '\n' {
				l.line++
				l.col = 0
			}
			l.advance()
		}
	}
}

// readString 读取字符串字面量（调用时已跳过开头的 "）
func (l *Lexer) readString() string {
	start := l.pos
	for l.ch != 0 && l.ch != '"' {
		if l.ch == '\\' {
			l.advance() // skip escaped char
		}
		l.advance()
	}
	s := l.input[start:l.pos]
	if l.ch == '"' {
		l.advance() // skip closing "
	}
	return s
}

// readInt 读取整数字面量
func (l *Lexer) readInt() string {
	start := l.pos - 1
	for l.ch >= '0' && l.ch <= '9' {
		l.advance()
	}
	return l.input[start : l.pos-1]
}

// readIdent 读取标识符或关键字
func (l *Lexer) readIdent() string {
	start := l.pos - 1
	for (l.ch >= 'a' && l.ch <= 'z') || (l.ch >= 'A' && l.ch <= 'Z') ||
		(l.ch >= '0' && l.ch <= '9') || l.ch == '_' {
		l.advance()
	}
	return l.input[start : l.pos-1]
}

// NextToken 返回下一个 Token
func (l *Lexer) NextToken() Token {
	for {
		l.skipWhitespace()

		// 检查注释
		if l.ch == '/' && (l.peek() == '/' || l.peek() == '*') {
			l.skipComment()
			continue
		}

		break
	}

	tok := Token{Line: l.line, Col: l.col}

	if l.ch == 0 {
		tok.Type = TokenEOF
		tok.Literal = ""
		return tok
	}

	ch := l.ch

	// 字符串
	if ch == '"' {
		start := l.pos // l.pos 指向 " 后面的第一个字符
		l.advance()    // l.ch = 第一个内容字符
		for l.ch != 0 && l.ch != '"' {
			if l.ch == '\\' {
				l.advance()
			}
			l.advance()
		}
		// l.ch == '"' 或 0, l.pos 已经越过关闭引号
		// 用 l.pos-1 排除关闭引号
		tok.Type = TokenString
		if l.ch == '"' {
			tok.Literal = l.input[start : l.pos-1]
			l.advance() // 跳过关闭 "
		} else {
			tok.Literal = l.input[start:l.pos]
		}
		return tok
	}

	// 整数
	if ch >= '0' && ch <= '9' {
		tok.Type = TokenInt
		tok.Literal = l.readInt()
		return tok
	}

	// 标识符和关键字
	if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_' {
		ident := l.readIdent()
		tok.Type = LookupIdent(ident)
		tok.Literal = ident
		return tok
	}

	// 双字符运算符和分隔符
	l.advance()
	switch ch {
	case '=':
		if l.ch == '=' {
			l.advance()
			tok.Type = TokenEq
			tok.Literal = "=="
		} else {
			tok.Type = TokenAssign
			tok.Literal = "="
		}
	case '!':
		if l.ch == '=' {
			l.advance()
			tok.Type = TokenNeq
			tok.Literal = "!="
		} else {
			tok.Type = TokenNot
			tok.Literal = "!"
		}
	case '<':
		if l.ch == '=' {
			l.advance()
			tok.Type = TokenLe
			tok.Literal = "<="
		} else {
			tok.Type = TokenLt
			tok.Literal = "<"
		}
	case '>':
		if l.ch == '=' {
			l.advance()
			tok.Type = TokenGe
			tok.Literal = ">="
		} else {
			tok.Type = TokenGt
			tok.Literal = ">"
		}
	case '&':
		if l.ch == '&' {
			l.advance()
			tok.Type = TokenAnd
			tok.Literal = "&&"
		} else {
			tok.Type = TokenIllegal
			tok.Literal = string(ch)
		}
	case '|':
		if l.ch == '|' {
			l.advance()
			tok.Type = TokenOr
			tok.Literal = "||"
		} else {
			tok.Type = TokenIllegal
			tok.Literal = string(ch)
		}
	case '+':
		if l.ch == '+' {
			l.advance()
			tok.Type = TokenInc
			tok.Literal = "++"
		} else {
			tok.Type = TokenPlus
			tok.Literal = "+"
		}
	case '-':
		if l.ch == '>' {
			l.advance()
			tok.Type = TokenArrow
			tok.Literal = "->"
		} else if l.ch == '-' {
			l.advance()
			tok.Type = TokenDec
			tok.Literal = "--"
		} else {
			tok.Type = TokenMinus
			tok.Literal = "-"
		}
	case '*':
		tok.Type = TokenStar
		tok.Literal = "*"
	case '/':
		tok.Type = TokenSlash
		tok.Literal = "/"
	case '%':
		tok.Type = TokenPercent
		tok.Literal = "%"
	case '(':
		tok.Type = TokenLParen
		tok.Literal = "("
	case ')':
		tok.Type = TokenRParen
		tok.Literal = ")"
	case '{':
		tok.Type = TokenLBrace
		tok.Literal = "{"
	case '}':
		tok.Type = TokenRBrace
		tok.Literal = "}"
	case '[':
		tok.Type = TokenLBracket
		tok.Literal = "["
	case ']':
		tok.Type = TokenRBracket
		tok.Literal = "]"
	case ';':
		tok.Type = TokenSemicolon
		tok.Literal = ";"
	case ',':
		tok.Type = TokenComma
		tok.Literal = ","
	case '.':
		tok.Type = TokenDot
		tok.Literal = "."
	default:
		tok.Type = TokenIllegal
		tok.Literal = string(ch)
	}

	return tok
}

// Tokenize 将整个输入分解为 Token 切片
func (l *Lexer) Tokenize() []Token {
	var tokens []Token
	for {
		tok := l.NextToken()
		tokens = append(tokens, tok)
		if tok.Type == TokenEOF {
			break
		}
	}
	return tokens
}
