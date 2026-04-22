// Package lpc 提供 LPC 语言子集的词法分析、语法解析和执行能力
package lpc

import "fmt"

// TokenType 表示词法单元的类型
type TokenType int

const (
	// 特殊 Token
	TokenEOF TokenType = iota
	TokenIllegal

	// 字面量
	TokenInt    // 123
	TokenString // "hello"
	TokenIdent  // foo, bar

	// 关键字
	TokenIntType   // int
	TokenStringType // string
	TokenVoid      // void
	TokenIf        // if
	TokenElse      // else
	TokenFor       // for
	TokenWhile     // while
	TokenReturn    // return
	TokenInherit   // inherit
	TokenObject    // object
	TokenThis      // this

	// 运算符
	TokenPlus     // +
	TokenMinus    // -
	TokenStar     // *
	TokenSlash    // /
	TokenPercent  // %
	TokenAssign   // =
	TokenEq       // ==
	TokenNeq      // !=
	TokenLt       // <
	TokenGt       // >
	TokenLe       // <=
	TokenGe       // >=
	TokenAnd      // &&
	TokenOr       // ||
	TokenNot      // !
	TokenInc      // ++
	TokenDec      // --

	// 分隔符
	TokenLParen    // (
	TokenRParen    // )
	TokenLBrace    // {
	TokenRBrace    // }
	TokenLBracket  // [
	TokenRBracket  // ]
	TokenSemicolon // ;
	TokenComma     // ,
	TokenDot       // .
	TokenArrow     // ->
)

// Token 表示一个词法单元
type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Col     int
}

// String 返回 Token 的可读字符串表示
func (t Token) String() string {
	return fmt.Sprintf("{%d %q L%d:C%d}", t.Type, t.Literal, t.Line, t.Col)
}

// keywords 是关键字到 TokenType 的映射
var keywords = map[string]TokenType{
	"int":     TokenIntType,
	"string":  TokenStringType,
	"void":    TokenVoid,
	"if":      TokenIf,
	"else":    TokenElse,
	"for":     TokenFor,
	"while":   TokenWhile,
	"return":  TokenReturn,
	"inherit": TokenInherit,
	"object":  TokenObject,
	"this":    TokenThis,
}

// LookupIdent 查找标识符是否为关键字
func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return TokenIdent
}
