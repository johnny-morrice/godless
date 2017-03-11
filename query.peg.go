package godless

import (
	"fmt"
	"math"
	"sort"
	"strconv"
)

const endSymbol rune = 1114112

/* The rule types inferred from the grammar are below. */
type pegRule uint8

const (
	ruleUnknown pegRule = iota
	ruleQuery
	ruleSelect
	ruleSelectKey
	ruleLimit
	ruleWhere
	ruleWhereClause
	ruleAndClause
	ruleOrClause
	rulePredicateClause
	rulePredicate
	rulePredicateValue
	rulePredicateRowKey
	rulePredicateKey
	rulePredicateLiteralValue
	rulePositiveInteger
	ruleKey
	ruleEscape
	ruleMustSpacing
	ruleSpacing
	ruleAction0
	rulePegText
	ruleAction1
	ruleAction2
	ruleAction3
	ruleAction4
	ruleAction5
	ruleAction6
	ruleAction7
	ruleAction8
	ruleAction9
	ruleAction10
	ruleAction11
)

var rul3s = [...]string{
	"Unknown",
	"Query",
	"Select",
	"SelectKey",
	"Limit",
	"Where",
	"WhereClause",
	"AndClause",
	"OrClause",
	"PredicateClause",
	"Predicate",
	"PredicateValue",
	"PredicateRowKey",
	"PredicateKey",
	"PredicateLiteralValue",
	"PositiveInteger",
	"Key",
	"Escape",
	"MustSpacing",
	"Spacing",
	"Action0",
	"PegText",
	"Action1",
	"Action2",
	"Action3",
	"Action4",
	"Action5",
	"Action6",
	"Action7",
	"Action8",
	"Action9",
	"Action10",
	"Action11",
}

type token32 struct {
	pegRule
	begin, end uint32
}

func (t *token32) String() string {
	return fmt.Sprintf("\x1B[34m%v\x1B[m %v %v", rul3s[t.pegRule], t.begin, t.end)
}

type node32 struct {
	token32
	up, next *node32
}

func (node *node32) print(pretty bool, buffer string) {
	var print func(node *node32, depth int)
	print = func(node *node32, depth int) {
		for node != nil {
			for c := 0; c < depth; c++ {
				fmt.Printf(" ")
			}
			rule := rul3s[node.pegRule]
			quote := strconv.Quote(string(([]rune(buffer)[node.begin:node.end])))
			if !pretty {
				fmt.Printf("%v %v\n", rule, quote)
			} else {
				fmt.Printf("\x1B[34m%v\x1B[m %v\n", rule, quote)
			}
			if node.up != nil {
				print(node.up, depth+1)
			}
			node = node.next
		}
	}
	print(node, 0)
}

func (node *node32) Print(buffer string) {
	node.print(false, buffer)
}

func (node *node32) PrettyPrint(buffer string) {
	node.print(true, buffer)
}

type tokens32 struct {
	tree []token32
}

func (t *tokens32) Trim(length uint32) {
	t.tree = t.tree[:length]
}

func (t *tokens32) Print() {
	for _, token := range t.tree {
		fmt.Println(token.String())
	}
}

func (t *tokens32) AST() *node32 {
	type element struct {
		node *node32
		down *element
	}
	tokens := t.Tokens()
	var stack *element
	for _, token := range tokens {
		if token.begin == token.end {
			continue
		}
		node := &node32{token32: token}
		for stack != nil && stack.node.begin >= token.begin && stack.node.end <= token.end {
			stack.node.next = node.up
			node.up = stack.node
			stack = stack.down
		}
		stack = &element{node: node, down: stack}
	}
	if stack != nil {
		return stack.node
	}
	return nil
}

func (t *tokens32) PrintSyntaxTree(buffer string) {
	t.AST().Print(buffer)
}

func (t *tokens32) PrettyPrintSyntaxTree(buffer string) {
	t.AST().PrettyPrint(buffer)
}

func (t *tokens32) Add(rule pegRule, begin, end, index uint32) {
	if tree := t.tree; int(index) >= len(tree) {
		expanded := make([]token32, 2*len(tree))
		copy(expanded, tree)
		t.tree = expanded
	}
	t.tree[index] = token32{
		pegRule: rule,
		begin:   begin,
		end:     end,
	}
}

func (t *tokens32) Tokens() []token32 {
	return t.tree
}

type QueryParser struct {
	QueryAST

	Buffer string
	buffer []rune
	rules  [33]func() bool
	parse  func(rule ...int) error
	reset  func()
	Pretty bool
	tokens32
}

func (p *QueryParser) Parse(rule ...int) error {
	return p.parse(rule...)
}

func (p *QueryParser) Reset() {
	p.reset()
}

type textPosition struct {
	line, symbol int
}

type textPositionMap map[int]textPosition

func translatePositions(buffer []rune, positions []int) textPositionMap {
	length, translations, j, line, symbol := len(positions), make(textPositionMap, len(positions)), 0, 1, 0
	sort.Ints(positions)

search:
	for i, c := range buffer {
		if c == '\n' {
			line, symbol = line+1, 0
		} else {
			symbol++
		}
		if i == positions[j] {
			translations[positions[j]] = textPosition{line, symbol}
			for j++; j < length; j++ {
				if i != positions[j] {
					continue search
				}
			}
			break search
		}
	}

	return translations
}

type parseError struct {
	p   *QueryParser
	max token32
}

func (e *parseError) Error() string {
	tokens, error := []token32{e.max}, "\n"
	positions, p := make([]int, 2*len(tokens)), 0
	for _, token := range tokens {
		positions[p], p = int(token.begin), p+1
		positions[p], p = int(token.end), p+1
	}
	translations := translatePositions(e.p.buffer, positions)
	format := "parse error near %v (line %v symbol %v - line %v symbol %v):\n%v\n"
	if e.p.Pretty {
		format = "parse error near \x1B[34m%v\x1B[m (line %v symbol %v - line %v symbol %v):\n%v\n"
	}
	for _, token := range tokens {
		begin, end := int(token.begin), int(token.end)
		error += fmt.Sprintf(format,
			rul3s[token.pegRule],
			translations[begin].line, translations[begin].symbol,
			translations[end].line, translations[end].symbol,
			strconv.Quote(string(e.p.buffer[begin:end])))
	}

	return error
}

func (p *QueryParser) PrintSyntaxTree() {
	if p.Pretty {
		p.tokens32.PrettyPrintSyntaxTree(p.Buffer)
	} else {
		p.tokens32.PrintSyntaxTree(p.Buffer)
	}
}

func (p *QueryParser) Execute() {
	buffer, _buffer, text, begin, end := p.Buffer, p.buffer, "", 0, 0
	for _, token := range p.Tokens() {
		switch token.pegRule {

		case rulePegText:
			begin, end = int(token.begin), int(token.end)
			text = string(_buffer[begin:end])

		case ruleAction0:
			p.AddSelect()
		case ruleAction1:
			p.SetTableName(buffer[begin:end])
		case ruleAction2:
			p.SetLimit(buffer[begin:end])
		case ruleAction3:
			p.PushWhere()
		case ruleAction4:
			p.PopWhere()
		case ruleAction5:
			p.SetWhereCommand("and")
		case ruleAction6:
			p.SetWhereCommand("or")
		case ruleAction7:
			p.InitPredicate()
		case ruleAction8:
			p.SetPredicateCommand(buffer[begin:end])
		case ruleAction9:
			p.UsePredicateRowKey()
		case ruleAction10:
			p.AddPredicateKey(buffer[begin:end])
		case ruleAction11:
			p.AddPredicateLiteral(buffer[begin:end])

		}
	}
	_, _, _, _, _ = buffer, _buffer, text, begin, end
}

func (p *QueryParser) Init() {
	var (
		max                  token32
		position, tokenIndex uint32
		buffer               []rune
	)
	p.reset = func() {
		max = token32{}
		position, tokenIndex = 0, 0

		p.buffer = []rune(p.Buffer)
		if len(p.buffer) == 0 || p.buffer[len(p.buffer)-1] != endSymbol {
			p.buffer = append(p.buffer, endSymbol)
		}
		buffer = p.buffer
	}
	p.reset()

	_rules := p.rules
	tree := tokens32{tree: make([]token32, math.MaxInt16)}
	p.parse = func(rule ...int) error {
		r := 1
		if len(rule) > 0 {
			r = rule[0]
		}
		matches := p.rules[r]()
		p.tokens32 = tree
		if matches {
			p.Trim(tokenIndex)
			return nil
		}
		return &parseError{p, max}
	}

	add := func(rule pegRule, begin uint32) {
		tree.Add(rule, begin, position, tokenIndex)
		tokenIndex++
		if begin != position && position > max.end {
			max = token32{rule, begin, position}
		}
	}

	matchDot := func() bool {
		if buffer[position] != endSymbol {
			position++
			return true
		}
		return false
	}

	/*matchChar := func(c byte) bool {
		if buffer[position] == c {
			position++
			return true
		}
		return false
	}*/

	/*matchRange := func(lower byte, upper byte) bool {
		if c := buffer[position]; c >= lower && c <= upper {
			position++
			return true
		}
		return false
	}*/

	_rules = [...]func() bool{
		nil,
		/* 0 Query <- <(Spacing Select Action0 !.)> */
		func() bool {
			position0, tokenIndex0 := position, tokenIndex
			{
				position1 := position
				if !_rules[ruleSpacing]() {
					goto l0
				}
				{
					position2 := position
					if buffer[position] != rune('s') {
						goto l0
					}
					position++
					if buffer[position] != rune('e') {
						goto l0
					}
					position++
					if buffer[position] != rune('l') {
						goto l0
					}
					position++
					if buffer[position] != rune('e') {
						goto l0
					}
					position++
					if buffer[position] != rune('c') {
						goto l0
					}
					position++
					if buffer[position] != rune('t') {
						goto l0
					}
					position++
					if !_rules[ruleMustSpacing]() {
						goto l0
					}
					{
						position3 := position
						{
							position4 := position
							if !_rules[ruleKey]() {
								goto l0
							}
							add(rulePegText, position4)
						}
						{
							add(ruleAction1, position)
						}
						add(ruleSelectKey, position3)
					}
					{
						position6, tokenIndex6 := position, tokenIndex
						if !_rules[ruleMustSpacing]() {
							goto l6
						}
						{
							position8 := position
							if buffer[position] != rune('w') {
								goto l6
							}
							position++
							if buffer[position] != rune('h') {
								goto l6
							}
							position++
							if buffer[position] != rune('e') {
								goto l6
							}
							position++
							if buffer[position] != rune('r') {
								goto l6
							}
							position++
							if buffer[position] != rune('e') {
								goto l6
							}
							position++
							if !_rules[ruleMustSpacing]() {
								goto l6
							}
							if !_rules[ruleWhereClause]() {
								goto l6
							}
							add(ruleWhere, position8)
						}
						goto l7
					l6:
						position, tokenIndex = position6, tokenIndex6
					}
				l7:
					{
						position9, tokenIndex9 := position, tokenIndex
						if !_rules[ruleMustSpacing]() {
							goto l9
						}
						{
							position11 := position
							if buffer[position] != rune('l') {
								goto l9
							}
							position++
							if buffer[position] != rune('i') {
								goto l9
							}
							position++
							if buffer[position] != rune('m') {
								goto l9
							}
							position++
							if buffer[position] != rune('i') {
								goto l9
							}
							position++
							if buffer[position] != rune('t') {
								goto l9
							}
							position++
							if !_rules[ruleMustSpacing]() {
								goto l9
							}
							{
								position12 := position
								{
									position13 := position
									if c := buffer[position]; c < rune('1') || c > rune('9') {
										goto l9
									}
									position++
								l14:
									{
										position15, tokenIndex15 := position, tokenIndex
										if c := buffer[position]; c < rune('0') || c > rune('9') {
											goto l15
										}
										position++
										goto l14
									l15:
										position, tokenIndex = position15, tokenIndex15
									}
									add(rulePositiveInteger, position13)
								}
								add(rulePegText, position12)
							}
							{
								add(ruleAction2, position)
							}
							add(ruleLimit, position11)
						}
						goto l10
					l9:
						position, tokenIndex = position9, tokenIndex9
					}
				l10:
					add(ruleSelect, position2)
				}
				{
					add(ruleAction0, position)
				}
				{
					position18, tokenIndex18 := position, tokenIndex
					if !matchDot() {
						goto l18
					}
					goto l0
				l18:
					position, tokenIndex = position18, tokenIndex18
				}
				add(ruleQuery, position1)
			}
			return true
		l0:
			position, tokenIndex = position0, tokenIndex0
			return false
		},
		/* 1 Select <- <('s' 'e' 'l' 'e' 'c' 't' MustSpacing SelectKey (MustSpacing Where)? (MustSpacing Limit)?)> */
		nil,
		/* 2 SelectKey <- <(<Key> Action1)> */
		nil,
		/* 3 Limit <- <('l' 'i' 'm' 'i' 't' MustSpacing <PositiveInteger> Action2)> */
		nil,
		/* 4 Where <- <('w' 'h' 'e' 'r' 'e' MustSpacing WhereClause)> */
		nil,
		/* 5 WhereClause <- <(Action3 ((&('s') PredicateClause) | (&('o') OrClause) | (&('a') AndClause)) Action4)> */
		func() bool {
			position23, tokenIndex23 := position, tokenIndex
			{
				position24 := position
				{
					add(ruleAction3, position)
				}
				{
					switch buffer[position] {
					case 's':
						{
							position27 := position
							{
								add(ruleAction7, position)
							}
							{
								position29 := position
								{
									position30 := position
									{
										position31, tokenIndex31 := position, tokenIndex
										if buffer[position] != rune('s') {
											goto l32
										}
										position++
										if buffer[position] != rune('t') {
											goto l32
										}
										position++
										if buffer[position] != rune('r') {
											goto l32
										}
										position++
										if buffer[position] != rune('_') {
											goto l32
										}
										position++
										if buffer[position] != rune('e') {
											goto l32
										}
										position++
										if buffer[position] != rune('q') {
											goto l32
										}
										position++
										goto l31
									l32:
										position, tokenIndex = position31, tokenIndex31
										if buffer[position] != rune('s') {
											goto l23
										}
										position++
										if buffer[position] != rune('t') {
											goto l23
										}
										position++
										if buffer[position] != rune('r') {
											goto l23
										}
										position++
										if buffer[position] != rune('_') {
											goto l23
										}
										position++
										if buffer[position] != rune('n') {
											goto l23
										}
										position++
										if buffer[position] != rune('e') {
											goto l23
										}
										position++
										if buffer[position] != rune('q') {
											goto l23
										}
										position++
									}
								l31:
									add(rulePegText, position30)
								}
								{
									add(ruleAction8, position)
								}
								add(rulePredicate, position29)
							}
							if !_rules[ruleSpacing]() {
								goto l23
							}
							if buffer[position] != rune('(') {
								goto l23
							}
							position++
							if !_rules[ruleSpacing]() {
								goto l23
							}
							if !_rules[rulePredicateValue]() {
								goto l23
							}
						l34:
							{
								position35, tokenIndex35 := position, tokenIndex
								if buffer[position] != rune(',') {
									goto l35
								}
								position++
								if !_rules[ruleSpacing]() {
									goto l35
								}
								if !_rules[rulePredicateValue]() {
									goto l35
								}
								if !_rules[ruleSpacing]() {
									goto l35
								}
								goto l34
							l35:
								position, tokenIndex = position35, tokenIndex35
							}
							if buffer[position] != rune(')') {
								goto l23
							}
							position++
							add(rulePredicateClause, position27)
						}
						break
					case 'o':
						{
							position36 := position
							if buffer[position] != rune('o') {
								goto l23
							}
							position++
							if buffer[position] != rune('r') {
								goto l23
							}
							position++
							{
								add(ruleAction6, position)
							}
							if !_rules[ruleSpacing]() {
								goto l23
							}
							if buffer[position] != rune('(') {
								goto l23
							}
							position++
							if !_rules[ruleSpacing]() {
								goto l23
							}
							if !_rules[ruleWhereClause]() {
								goto l23
							}
							if !_rules[ruleSpacing]() {
								goto l23
							}
						l38:
							{
								position39, tokenIndex39 := position, tokenIndex
								if buffer[position] != rune(',') {
									goto l39
								}
								position++
								if !_rules[ruleSpacing]() {
									goto l39
								}
								if !_rules[ruleWhereClause]() {
									goto l39
								}
								if !_rules[ruleSpacing]() {
									goto l39
								}
								goto l38
							l39:
								position, tokenIndex = position39, tokenIndex39
							}
							if buffer[position] != rune(')') {
								goto l23
							}
							position++
							add(ruleOrClause, position36)
						}
						break
					default:
						{
							position40 := position
							if buffer[position] != rune('a') {
								goto l23
							}
							position++
							if buffer[position] != rune('n') {
								goto l23
							}
							position++
							if buffer[position] != rune('d') {
								goto l23
							}
							position++
							{
								add(ruleAction5, position)
							}
							if !_rules[ruleSpacing]() {
								goto l23
							}
							if buffer[position] != rune('(') {
								goto l23
							}
							position++
							if !_rules[ruleSpacing]() {
								goto l23
							}
							if !_rules[ruleWhereClause]() {
								goto l23
							}
							if !_rules[ruleSpacing]() {
								goto l23
							}
						l42:
							{
								position43, tokenIndex43 := position, tokenIndex
								if buffer[position] != rune(',') {
									goto l43
								}
								position++
								if !_rules[ruleSpacing]() {
									goto l43
								}
								if !_rules[ruleWhereClause]() {
									goto l43
								}
								if !_rules[ruleSpacing]() {
									goto l43
								}
								goto l42
							l43:
								position, tokenIndex = position43, tokenIndex43
							}
							if buffer[position] != rune(')') {
								goto l23
							}
							position++
							add(ruleAndClause, position40)
						}
						break
					}
				}

				{
					add(ruleAction4, position)
				}
				add(ruleWhereClause, position24)
			}
			return true
		l23:
			position, tokenIndex = position23, tokenIndex23
			return false
		},
		/* 6 AndClause <- <('a' 'n' 'd' Action5 Spacing '(' Spacing WhereClause Spacing (',' Spacing WhereClause Spacing)* ')')> */
		nil,
		/* 7 OrClause <- <('o' 'r' Action6 Spacing '(' Spacing WhereClause Spacing (',' Spacing WhereClause Spacing)* ')')> */
		nil,
		/* 8 PredicateClause <- <(Action7 Predicate Spacing '(' Spacing PredicateValue (',' Spacing PredicateValue Spacing)* ')')> */
		nil,
		/* 9 Predicate <- <(<(('s' 't' 'r' '_' 'e' 'q') / ('s' 't' 'r' '_' 'n' 'e' 'q'))> Action8)> */
		nil,
		/* 10 PredicateValue <- <((&('\'') PredicateLiteralValue) | (&('@') PredicateRowKey) | (&('0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9' | 'A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z' | '\\' | 'a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') PredicateKey))> */
		func() bool {
			position49, tokenIndex49 := position, tokenIndex
			{
				position50 := position
				{
					switch buffer[position] {
					case '\'':
						{
							position52 := position
							if buffer[position] != rune('\'') {
								goto l49
							}
							position++
							{
								position53 := position
							l54:
								{
									position55, tokenIndex55 := position, tokenIndex
									{
										position56, tokenIndex56 := position, tokenIndex
										if buffer[position] != rune('\'') {
											goto l56
										}
										position++
										goto l55
									l56:
										position, tokenIndex = position56, tokenIndex56
									}
									if !matchDot() {
										goto l55
									}
									goto l54
								l55:
									position, tokenIndex = position55, tokenIndex55
								}
								add(rulePegText, position53)
							}
							if buffer[position] != rune('\'') {
								goto l49
							}
							position++
							{
								add(ruleAction11, position)
							}
							add(rulePredicateLiteralValue, position52)
						}
						break
					case '@':
						{
							position58 := position
							if buffer[position] != rune('@') {
								goto l49
							}
							position++
							if buffer[position] != rune('k') {
								goto l49
							}
							position++
							if buffer[position] != rune('e') {
								goto l49
							}
							position++
							if buffer[position] != rune('y') {
								goto l49
							}
							position++
							{
								add(ruleAction9, position)
							}
							add(rulePredicateRowKey, position58)
						}
						break
					default:
						{
							position60 := position
							{
								position61 := position
								if !_rules[ruleKey]() {
									goto l49
								}
								add(rulePegText, position61)
							}
							{
								add(ruleAction10, position)
							}
							add(rulePredicateKey, position60)
						}
						break
					}
				}

				add(rulePredicateValue, position50)
			}
			return true
		l49:
			position, tokenIndex = position49, tokenIndex49
			return false
		},
		/* 11 PredicateRowKey <- <('@' 'k' 'e' 'y' Action9)> */
		nil,
		/* 12 PredicateKey <- <(<Key> Action10)> */
		nil,
		/* 13 PredicateLiteralValue <- <('\'' <(!'\'' .)*> '\'' Action11)> */
		nil,
		/* 14 PositiveInteger <- <([1-9] [0-9]*)> */
		nil,
		/* 15 Key <- <(Escape / ((&('0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9') [0-9]) | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z])))+> */
		func() bool {
			position67, tokenIndex67 := position, tokenIndex
			{
				position68 := position
				{
					position71, tokenIndex71 := position, tokenIndex
					{
						position73 := position
						if buffer[position] != rune('\\') {
							goto l72
						}
						position++
						{
							switch buffer[position] {
							case 'v':
								if buffer[position] != rune('v') {
									goto l72
								}
								position++
								break
							case 't':
								if buffer[position] != rune('t') {
									goto l72
								}
								position++
								break
							case 'r':
								if buffer[position] != rune('r') {
									goto l72
								}
								position++
								break
							case 'n':
								if buffer[position] != rune('n') {
									goto l72
								}
								position++
								break
							case 'f':
								if buffer[position] != rune('f') {
									goto l72
								}
								position++
								break
							case 'b':
								if buffer[position] != rune('b') {
									goto l72
								}
								position++
								break
							case 'a':
								if buffer[position] != rune('a') {
									goto l72
								}
								position++
								break
							case '\\':
								if buffer[position] != rune('\\') {
									goto l72
								}
								position++
								break
							case '?':
								if buffer[position] != rune('?') {
									goto l72
								}
								position++
								break
							case '"':
								if buffer[position] != rune('"') {
									goto l72
								}
								position++
								break
							default:
								if buffer[position] != rune('\'') {
									goto l72
								}
								position++
								break
							}
						}

						add(ruleEscape, position73)
					}
					goto l71
				l72:
					position, tokenIndex = position71, tokenIndex71
					{
						switch buffer[position] {
						case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l67
							}
							position++
							break
						case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l67
							}
							position++
							break
						default:
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l67
							}
							position++
							break
						}
					}

				}
			l71:
			l69:
				{
					position70, tokenIndex70 := position, tokenIndex
					{
						position76, tokenIndex76 := position, tokenIndex
						{
							position78 := position
							if buffer[position] != rune('\\') {
								goto l77
							}
							position++
							{
								switch buffer[position] {
								case 'v':
									if buffer[position] != rune('v') {
										goto l77
									}
									position++
									break
								case 't':
									if buffer[position] != rune('t') {
										goto l77
									}
									position++
									break
								case 'r':
									if buffer[position] != rune('r') {
										goto l77
									}
									position++
									break
								case 'n':
									if buffer[position] != rune('n') {
										goto l77
									}
									position++
									break
								case 'f':
									if buffer[position] != rune('f') {
										goto l77
									}
									position++
									break
								case 'b':
									if buffer[position] != rune('b') {
										goto l77
									}
									position++
									break
								case 'a':
									if buffer[position] != rune('a') {
										goto l77
									}
									position++
									break
								case '\\':
									if buffer[position] != rune('\\') {
										goto l77
									}
									position++
									break
								case '?':
									if buffer[position] != rune('?') {
										goto l77
									}
									position++
									break
								case '"':
									if buffer[position] != rune('"') {
										goto l77
									}
									position++
									break
								default:
									if buffer[position] != rune('\'') {
										goto l77
									}
									position++
									break
								}
							}

							add(ruleEscape, position78)
						}
						goto l76
					l77:
						position, tokenIndex = position76, tokenIndex76
						{
							switch buffer[position] {
							case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
								if c := buffer[position]; c < rune('0') || c > rune('9') {
									goto l70
								}
								position++
								break
							case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
								if c := buffer[position]; c < rune('A') || c > rune('Z') {
									goto l70
								}
								position++
								break
							default:
								if c := buffer[position]; c < rune('a') || c > rune('z') {
									goto l70
								}
								position++
								break
							}
						}

					}
				l76:
					goto l69
				l70:
					position, tokenIndex = position70, tokenIndex70
				}
				add(ruleKey, position68)
			}
			return true
		l67:
			position, tokenIndex = position67, tokenIndex67
			return false
		},
		/* 16 Escape <- <('\\' ((&('v') 'v') | (&('t') 't') | (&('r') 'r') | (&('n') 'n') | (&('f') 'f') | (&('b') 'b') | (&('a') 'a') | (&('\\') '\\') | (&('?') '?') | (&('"') '"') | (&('\'') '\'')))> */
		nil,
		/* 17 MustSpacing <- <((&('\n') '\n') | (&('\t') '\t') | (&(' ') ' '))+> */
		func() bool {
			position82, tokenIndex82 := position, tokenIndex
			{
				position83 := position
				{
					switch buffer[position] {
					case '\n':
						if buffer[position] != rune('\n') {
							goto l82
						}
						position++
						break
					case '\t':
						if buffer[position] != rune('\t') {
							goto l82
						}
						position++
						break
					default:
						if buffer[position] != rune(' ') {
							goto l82
						}
						position++
						break
					}
				}

			l84:
				{
					position85, tokenIndex85 := position, tokenIndex
					{
						switch buffer[position] {
						case '\n':
							if buffer[position] != rune('\n') {
								goto l85
							}
							position++
							break
						case '\t':
							if buffer[position] != rune('\t') {
								goto l85
							}
							position++
							break
						default:
							if buffer[position] != rune(' ') {
								goto l85
							}
							position++
							break
						}
					}

					goto l84
				l85:
					position, tokenIndex = position85, tokenIndex85
				}
				add(ruleMustSpacing, position83)
			}
			return true
		l82:
			position, tokenIndex = position82, tokenIndex82
			return false
		},
		/* 18 Spacing <- <((&('\n') '\n') | (&('\t') '\t') | (&(' ') ' '))*> */
		func() bool {
			{
				position89 := position
			l90:
				{
					position91, tokenIndex91 := position, tokenIndex
					{
						switch buffer[position] {
						case '\n':
							if buffer[position] != rune('\n') {
								goto l91
							}
							position++
							break
						case '\t':
							if buffer[position] != rune('\t') {
								goto l91
							}
							position++
							break
						default:
							if buffer[position] != rune(' ') {
								goto l91
							}
							position++
							break
						}
					}

					goto l90
				l91:
					position, tokenIndex = position91, tokenIndex91
				}
				add(ruleSpacing, position89)
			}
			return true
		},
		/* 20 Action0 <- <{ p.AddSelect() }> */
		nil,
		nil,
		/* 22 Action1 <- <{ p.SetTableName(buffer[begin:end]) }> */
		nil,
		/* 23 Action2 <- <{ p.SetLimit(buffer[begin:end])}> */
		nil,
		/* 24 Action3 <- <{ p.PushWhere() }> */
		nil,
		/* 25 Action4 <- <{ p.PopWhere() }> */
		nil,
		/* 26 Action5 <- <{ p.SetWhereCommand("and") }> */
		nil,
		/* 27 Action6 <- <{ p.SetWhereCommand("or") }> */
		nil,
		/* 28 Action7 <- <{ p.InitPredicate() }> */
		nil,
		/* 29 Action8 <- <{ p.SetPredicateCommand(buffer[begin:end]) }> */
		nil,
		/* 30 Action9 <- <{ p.UsePredicateRowKey() }> */
		nil,
		/* 31 Action10 <- <{ p.AddPredicateKey(buffer[begin:end]) }> */
		nil,
		/* 32 Action11 <- <{ p.AddPredicateLiteral(buffer[begin:end])}> */
		nil,
	}
	p.rules = _rules
}
