package query

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
	ruleTableName
	ruleTableNameText
	ruleTableNamePlaceholder
	ruleJoin
	ruleJoinRow
	ruleJoinRowKey
	ruleJoinRowKeyValuePlaceholder
	ruleJoinRowKeyValueText
	ruleJoinPoint
	ruleJoinPointValuePlaceholder
	ruleJoinPointValueText
	ruleJoinPointKeyText
	ruleJoinPointKeyPlaceholder
	ruleSelect
	ruleWherePart
	ruleLimit
	ruleLimitText
	ruleLimitPlaceholder
	ruleCryptoKey
	ruleWhere
	ruleWhereClause
	ruleAndClause
	ruleOrClause
	rulePredicateClause
	rulePredicate
	rulePredicateValue
	rulePredicateRowKey
	rulePredicateKey
	rulePredicateKeyText
	rulePredicateKeyLiteral
	rulePredicateLiteral
	rulePredicateLiteralText
	rulePredicateLiteralPlaceholder
	ruleKeyPlaceholder
	ruleLiteralPlaceholder
	ruleLiteral
	rulePositiveInteger
	ruleKey
	ruleEscape
	ruleMustSpacing
	ruleSpacing
	ruleAction0
	ruleAction1
	rulePegText
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
	ruleAction12
	ruleAction13
	ruleAction14
	ruleAction15
	ruleAction16
	ruleAction17
	ruleAction18
	ruleAction19
	ruleAction20
	ruleAction21
	ruleAction22
	ruleAction23
	ruleAction24
)

var rul3s = [...]string{
	"Unknown",
	"Query",
	"TableName",
	"TableNameText",
	"TableNamePlaceholder",
	"Join",
	"JoinRow",
	"JoinRowKey",
	"JoinRowKeyValuePlaceholder",
	"JoinRowKeyValueText",
	"JoinPoint",
	"JoinPointValuePlaceholder",
	"JoinPointValueText",
	"JoinPointKeyText",
	"JoinPointKeyPlaceholder",
	"Select",
	"WherePart",
	"Limit",
	"LimitText",
	"LimitPlaceholder",
	"CryptoKey",
	"Where",
	"WhereClause",
	"AndClause",
	"OrClause",
	"PredicateClause",
	"Predicate",
	"PredicateValue",
	"PredicateRowKey",
	"PredicateKey",
	"PredicateKeyText",
	"PredicateKeyLiteral",
	"PredicateLiteral",
	"PredicateLiteralText",
	"PredicateLiteralPlaceholder",
	"KeyPlaceholder",
	"LiteralPlaceholder",
	"Literal",
	"PositiveInteger",
	"Key",
	"Escape",
	"MustSpacing",
	"Spacing",
	"Action0",
	"Action1",
	"PegText",
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
	"Action12",
	"Action13",
	"Action14",
	"Action15",
	"Action16",
	"Action17",
	"Action18",
	"Action19",
	"Action20",
	"Action21",
	"Action22",
	"Action23",
	"Action24",
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
	rules  [69]func() bool
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
			p.AddJoin()
		case ruleAction2:
			p.SetTableName(buffer[begin:end])
		case ruleAction3:
			p.SetTableNamePlaceholder(begin)
		case ruleAction4:
			p.AddJoinRow()
		case ruleAction5:
			p.SetJoinRowKeyPlaceholder(begin)
		case ruleAction6:
			p.SetJoinRowKey(buffer[begin:end])
		case ruleAction7:
			p.SetJoinValuePlaceholder(begin)
		case ruleAction8:
			p.SetJoinValue(buffer[begin:end])
		case ruleAction9:
			p.SetJoinKey(buffer[begin:end])
		case ruleAction10:
			p.SetJoinKeyPlaceholder(begin)
		case ruleAction11:
			p.SetLimit(buffer[begin:end])
		case ruleAction12:
			p.SetLimitPlaceholder(begin)
		case ruleAction13:
			p.AddCryptoKey(buffer[begin:end])
		case ruleAction14:
			p.PushWhere()
		case ruleAction15:
			p.PopWhere()
		case ruleAction16:
			p.SetWhereCommand("and")
		case ruleAction17:
			p.SetWhereCommand("or")
		case ruleAction18:
			p.InitPredicate()
		case ruleAction19:
			p.SetPredicateCommand(buffer[begin:end])
		case ruleAction20:
			p.UsePredicateRowKey()
		case ruleAction21:
			p.AddPredicateKey(buffer[begin:end])
		case ruleAction22:
			p.AddPredicateKeyPlaceholder(begin)
		case ruleAction23:
			p.AddPredicateLiteral(buffer[begin:end])
		case ruleAction24:
			p.AddPredicateLiteralPlaceholder(begin)

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
		/* 0 Query <- <(Spacing ((Select Action0) / (Join Action1)) Spacing !.)> */
		func() bool {
			position0, tokenIndex0 := position, tokenIndex
			{
				position1 := position
				if !_rules[ruleSpacing]() {
					goto l0
				}
				{
					position2, tokenIndex2 := position, tokenIndex
					{
						position4 := position
						if buffer[position] != rune('s') {
							goto l3
						}
						position++
						if buffer[position] != rune('e') {
							goto l3
						}
						position++
						if buffer[position] != rune('l') {
							goto l3
						}
						position++
						if buffer[position] != rune('e') {
							goto l3
						}
						position++
						if buffer[position] != rune('c') {
							goto l3
						}
						position++
						if buffer[position] != rune('t') {
							goto l3
						}
						position++
						if !_rules[ruleMustSpacing]() {
							goto l3
						}
						if !_rules[ruleTableName]() {
							goto l3
						}
					l5:
						{
							position6, tokenIndex6 := position, tokenIndex
							if !_rules[ruleMustSpacing]() {
								goto l6
							}
							{
								position7 := position
								{
									switch buffer[position] {
									case 's':
										if !_rules[ruleCryptoKey]() {
											goto l6
										}
										break
									case 'l':
										{
											position9 := position
											if buffer[position] != rune('l') {
												goto l6
											}
											position++
											if buffer[position] != rune('i') {
												goto l6
											}
											position++
											if buffer[position] != rune('m') {
												goto l6
											}
											position++
											if buffer[position] != rune('i') {
												goto l6
											}
											position++
											if buffer[position] != rune('t') {
												goto l6
											}
											position++
											{
												position10, tokenIndex10 := position, tokenIndex
												{
													position12 := position
													if !_rules[ruleMustSpacing]() {
														goto l11
													}
													{
														position13 := position
														{
															position14 := position
															if c := buffer[position]; c < rune('1') || c > rune('9') {
																goto l11
															}
															position++
														l15:
															{
																position16, tokenIndex16 := position, tokenIndex
																if c := buffer[position]; c < rune('0') || c > rune('9') {
																	goto l16
																}
																position++
																goto l15
															l16:
																position, tokenIndex = position16, tokenIndex16
															}
															add(rulePositiveInteger, position14)
														}
														add(rulePegText, position13)
													}
													{
														add(ruleAction11, position)
													}
													add(ruleLimitText, position12)
												}
												goto l10
											l11:
												position, tokenIndex = position10, tokenIndex10
												{
													position18 := position
													{
														position19 := position
														if !_rules[ruleLiteralPlaceholder]() {
															goto l6
														}
														add(rulePegText, position19)
													}
													{
														add(ruleAction12, position)
													}
													add(ruleLimitPlaceholder, position18)
												}
											}
										l10:
											add(ruleLimit, position9)
										}
										break
									default:
										{
											position21 := position
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
											add(ruleWhere, position21)
										}
										break
									}
								}

								add(ruleWherePart, position7)
							}
							goto l5
						l6:
							position, tokenIndex = position6, tokenIndex6
						}
						add(ruleSelect, position4)
					}
					{
						add(ruleAction0, position)
					}
					goto l2
				l3:
					position, tokenIndex = position2, tokenIndex2
					{
						position23 := position
						if buffer[position] != rune('j') {
							goto l0
						}
						position++
						if buffer[position] != rune('o') {
							goto l0
						}
						position++
						if buffer[position] != rune('i') {
							goto l0
						}
						position++
						if buffer[position] != rune('n') {
							goto l0
						}
						position++
						if !_rules[ruleMustSpacing]() {
							goto l0
						}
						if !_rules[ruleTableName]() {
							goto l0
						}
					l24:
						{
							position25, tokenIndex25 := position, tokenIndex
							if !_rules[ruleMustSpacing]() {
								goto l25
							}
							if !_rules[ruleCryptoKey]() {
								goto l25
							}
							goto l24
						l25:
							position, tokenIndex = position25, tokenIndex25
						}
						if !_rules[ruleMustSpacing]() {
							goto l0
						}
						if buffer[position] != rune('r') {
							goto l0
						}
						position++
						if buffer[position] != rune('o') {
							goto l0
						}
						position++
						if buffer[position] != rune('w') {
							goto l0
						}
						position++
						if buffer[position] != rune('s') {
							goto l0
						}
						position++
						if !_rules[ruleMustSpacing]() {
							goto l0
						}
						if !_rules[ruleJoinRow]() {
							goto l0
						}
					l26:
						{
							position27, tokenIndex27 := position, tokenIndex
							if !_rules[ruleSpacing]() {
								goto l27
							}
							if buffer[position] != rune(',') {
								goto l27
							}
							position++
							if !_rules[ruleSpacing]() {
								goto l27
							}
							if !_rules[ruleJoinRow]() {
								goto l27
							}
							goto l26
						l27:
							position, tokenIndex = position27, tokenIndex27
						}
						if !_rules[ruleSpacing]() {
							goto l0
						}
						add(ruleJoin, position23)
					}
					{
						add(ruleAction1, position)
					}
				}
			l2:
				if !_rules[ruleSpacing]() {
					goto l0
				}
				{
					position29, tokenIndex29 := position, tokenIndex
					if !matchDot() {
						goto l29
					}
					goto l0
				l29:
					position, tokenIndex = position29, tokenIndex29
				}
				add(ruleQuery, position1)
			}
			return true
		l0:
			position, tokenIndex = position0, tokenIndex0
			return false
		},
		/* 1 TableName <- <(TableNameText / TableNamePlaceholder)> */
		func() bool {
			position30, tokenIndex30 := position, tokenIndex
			{
				position31 := position
				{
					position32, tokenIndex32 := position, tokenIndex
					{
						position34 := position
						{
							position35 := position
							if !_rules[ruleKey]() {
								goto l33
							}
							add(rulePegText, position35)
						}
						{
							add(ruleAction2, position)
						}
						add(ruleTableNameText, position34)
					}
					goto l32
				l33:
					position, tokenIndex = position32, tokenIndex32
					{
						position37 := position
						{
							position38 := position
							if !_rules[ruleKeyPlaceholder]() {
								goto l30
							}
							add(rulePegText, position38)
						}
						{
							add(ruleAction3, position)
						}
						add(ruleTableNamePlaceholder, position37)
					}
				}
			l32:
				add(ruleTableName, position31)
			}
			return true
		l30:
			position, tokenIndex = position30, tokenIndex30
			return false
		},
		/* 2 TableNameText <- <(<Key> Action2)> */
		nil,
		/* 3 TableNamePlaceholder <- <(<KeyPlaceholder> Action3)> */
		nil,
		/* 4 Join <- <('j' 'o' 'i' 'n' MustSpacing TableName (MustSpacing CryptoKey)* MustSpacing ('r' 'o' 'w' 's') MustSpacing JoinRow (Spacing ',' Spacing JoinRow)* Spacing)> */
		nil,
		/* 5 JoinRow <- <(Action4 '(' Spacing JoinRowKey Spacing (',' Spacing JoinPoint Spacing)* ')')> */
		func() bool {
			position43, tokenIndex43 := position, tokenIndex
			{
				position44 := position
				{
					add(ruleAction4, position)
				}
				if buffer[position] != rune('(') {
					goto l43
				}
				position++
				if !_rules[ruleSpacing]() {
					goto l43
				}
				{
					position46 := position
					if buffer[position] != rune('@') {
						goto l43
					}
					position++
					if buffer[position] != rune('k') {
						goto l43
					}
					position++
					if buffer[position] != rune('e') {
						goto l43
					}
					position++
					if buffer[position] != rune('y') {
						goto l43
					}
					position++
					if !_rules[ruleSpacing]() {
						goto l43
					}
					if buffer[position] != rune('=') {
						goto l43
					}
					position++
					if !_rules[ruleSpacing]() {
						goto l43
					}
					{
						position47, tokenIndex47 := position, tokenIndex
						{
							position49 := position
							{
								position50, tokenIndex50 := position, tokenIndex
								if buffer[position] != rune('@') {
									goto l51
								}
								position++
								if buffer[position] != rune('"') {
									goto l51
								}
								position++
								{
									position52 := position
									if !_rules[ruleLiteral]() {
										goto l51
									}
									add(rulePegText, position52)
								}
								if buffer[position] != rune('"') {
									goto l51
								}
								position++
								goto l50
							l51:
								position, tokenIndex = position50, tokenIndex50
								{
									position53 := position
									if !_rules[ruleKey]() {
										goto l48
									}
									add(rulePegText, position53)
								}
							}
						l50:
							{
								add(ruleAction6, position)
							}
							add(ruleJoinRowKeyValueText, position49)
						}
						goto l47
					l48:
						position, tokenIndex = position47, tokenIndex47
						{
							position55 := position
							{
								position56 := position
								if !_rules[ruleLiteralPlaceholder]() {
									goto l43
								}
								add(rulePegText, position56)
							}
							{
								add(ruleAction5, position)
							}
							add(ruleJoinRowKeyValuePlaceholder, position55)
						}
					}
				l47:
					add(ruleJoinRowKey, position46)
				}
				if !_rules[ruleSpacing]() {
					goto l43
				}
			l58:
				{
					position59, tokenIndex59 := position, tokenIndex
					if buffer[position] != rune(',') {
						goto l59
					}
					position++
					if !_rules[ruleSpacing]() {
						goto l59
					}
					{
						position60 := position
						{
							position61, tokenIndex61 := position, tokenIndex
							{
								position63 := position
								{
									position64, tokenIndex64 := position, tokenIndex
									{
										position66 := position
										if !_rules[ruleKey]() {
											goto l65
										}
										add(rulePegText, position66)
									}
									goto l64
								l65:
									position, tokenIndex = position64, tokenIndex64
									if buffer[position] != rune('@') {
										goto l62
									}
									position++
									if buffer[position] != rune('"') {
										goto l62
									}
									position++
									{
										position67 := position
										if !_rules[ruleLiteral]() {
											goto l62
										}
										add(rulePegText, position67)
									}
									if buffer[position] != rune('"') {
										goto l62
									}
									position++
								}
							l64:
								{
									add(ruleAction9, position)
								}
								add(ruleJoinPointKeyText, position63)
							}
							goto l61
						l62:
							position, tokenIndex = position61, tokenIndex61
							{
								position69 := position
								{
									position70 := position
									if !_rules[ruleKeyPlaceholder]() {
										goto l59
									}
									add(rulePegText, position70)
								}
								{
									add(ruleAction10, position)
								}
								add(ruleJoinPointKeyPlaceholder, position69)
							}
						}
					l61:
						if !_rules[ruleSpacing]() {
							goto l59
						}
						if buffer[position] != rune('=') {
							goto l59
						}
						position++
						if !_rules[ruleSpacing]() {
							goto l59
						}
						{
							position72, tokenIndex72 := position, tokenIndex
							{
								position74 := position
								if buffer[position] != rune('"') {
									goto l73
								}
								position++
								{
									position75 := position
									if !_rules[ruleLiteral]() {
										goto l73
									}
									add(rulePegText, position75)
								}
								if buffer[position] != rune('"') {
									goto l73
								}
								position++
								{
									add(ruleAction8, position)
								}
								add(ruleJoinPointValueText, position74)
							}
							goto l72
						l73:
							position, tokenIndex = position72, tokenIndex72
							{
								position77 := position
								{
									position78 := position
									if !_rules[ruleLiteralPlaceholder]() {
										goto l59
									}
									add(rulePegText, position78)
								}
								{
									add(ruleAction7, position)
								}
								add(ruleJoinPointValuePlaceholder, position77)
							}
						}
					l72:
						add(ruleJoinPoint, position60)
					}
					if !_rules[ruleSpacing]() {
						goto l59
					}
					goto l58
				l59:
					position, tokenIndex = position59, tokenIndex59
				}
				if buffer[position] != rune(')') {
					goto l43
				}
				position++
				add(ruleJoinRow, position44)
			}
			return true
		l43:
			position, tokenIndex = position43, tokenIndex43
			return false
		},
		/* 6 JoinRowKey <- <('@' 'k' 'e' 'y' Spacing '=' Spacing (JoinRowKeyValueText / JoinRowKeyValuePlaceholder))> */
		nil,
		/* 7 JoinRowKeyValuePlaceholder <- <(<LiteralPlaceholder> Action5)> */
		nil,
		/* 8 JoinRowKeyValueText <- <((('@' '"' <Literal> '"') / <Key>) Action6)> */
		nil,
		/* 9 JoinPoint <- <((JoinPointKeyText / JoinPointKeyPlaceholder) Spacing '=' Spacing (JoinPointValueText / JoinPointValuePlaceholder))> */
		nil,
		/* 10 JoinPointValuePlaceholder <- <(<LiteralPlaceholder> Action7)> */
		nil,
		/* 11 JoinPointValueText <- <('"' <Literal> '"' Action8)> */
		nil,
		/* 12 JoinPointKeyText <- <((<Key> / ('@' '"' <Literal> '"')) Action9)> */
		nil,
		/* 13 JoinPointKeyPlaceholder <- <(<KeyPlaceholder> Action10)> */
		nil,
		/* 14 Select <- <('s' 'e' 'l' 'e' 'c' 't' MustSpacing TableName (MustSpacing WherePart)*)> */
		nil,
		/* 15 WherePart <- <((&('s') CryptoKey) | (&('l') Limit) | (&('w') Where))> */
		nil,
		/* 16 Limit <- <('l' 'i' 'm' 'i' 't' (LimitText / LimitPlaceholder))> */
		nil,
		/* 17 LimitText <- <(MustSpacing <PositiveInteger> Action11)> */
		nil,
		/* 18 LimitPlaceholder <- <(<LiteralPlaceholder> Action12)> */
		nil,
		/* 19 CryptoKey <- <('s' 'i' 'g' 'n' 'e' 'd' MustSpacing '"' <Key> '"' Action13)> */
		func() bool {
			position93, tokenIndex93 := position, tokenIndex
			{
				position94 := position
				if buffer[position] != rune('s') {
					goto l93
				}
				position++
				if buffer[position] != rune('i') {
					goto l93
				}
				position++
				if buffer[position] != rune('g') {
					goto l93
				}
				position++
				if buffer[position] != rune('n') {
					goto l93
				}
				position++
				if buffer[position] != rune('e') {
					goto l93
				}
				position++
				if buffer[position] != rune('d') {
					goto l93
				}
				position++
				if !_rules[ruleMustSpacing]() {
					goto l93
				}
				if buffer[position] != rune('"') {
					goto l93
				}
				position++
				{
					position95 := position
					if !_rules[ruleKey]() {
						goto l93
					}
					add(rulePegText, position95)
				}
				if buffer[position] != rune('"') {
					goto l93
				}
				position++
				{
					add(ruleAction13, position)
				}
				add(ruleCryptoKey, position94)
			}
			return true
		l93:
			position, tokenIndex = position93, tokenIndex93
			return false
		},
		/* 20 Where <- <('w' 'h' 'e' 'r' 'e' MustSpacing WhereClause)> */
		nil,
		/* 21 WhereClause <- <(Action14 (AndClause / OrClause / PredicateClause) Action15)> */
		func() bool {
			position98, tokenIndex98 := position, tokenIndex
			{
				position99 := position
				{
					add(ruleAction14, position)
				}
				{
					position101, tokenIndex101 := position, tokenIndex
					{
						position103 := position
						if buffer[position] != rune('a') {
							goto l102
						}
						position++
						if buffer[position] != rune('n') {
							goto l102
						}
						position++
						if buffer[position] != rune('d') {
							goto l102
						}
						position++
						{
							add(ruleAction16, position)
						}
						if !_rules[ruleSpacing]() {
							goto l102
						}
						if buffer[position] != rune('(') {
							goto l102
						}
						position++
						if !_rules[ruleSpacing]() {
							goto l102
						}
						if !_rules[ruleWhereClause]() {
							goto l102
						}
						if !_rules[ruleSpacing]() {
							goto l102
						}
					l105:
						{
							position106, tokenIndex106 := position, tokenIndex
							if buffer[position] != rune(',') {
								goto l106
							}
							position++
							if !_rules[ruleSpacing]() {
								goto l106
							}
							if !_rules[ruleWhereClause]() {
								goto l106
							}
							if !_rules[ruleSpacing]() {
								goto l106
							}
							goto l105
						l106:
							position, tokenIndex = position106, tokenIndex106
						}
						if buffer[position] != rune(')') {
							goto l102
						}
						position++
						add(ruleAndClause, position103)
					}
					goto l101
				l102:
					position, tokenIndex = position101, tokenIndex101
					{
						position108 := position
						if buffer[position] != rune('o') {
							goto l107
						}
						position++
						if buffer[position] != rune('r') {
							goto l107
						}
						position++
						{
							add(ruleAction17, position)
						}
						if !_rules[ruleSpacing]() {
							goto l107
						}
						if buffer[position] != rune('(') {
							goto l107
						}
						position++
						if !_rules[ruleSpacing]() {
							goto l107
						}
						if !_rules[ruleWhereClause]() {
							goto l107
						}
						if !_rules[ruleSpacing]() {
							goto l107
						}
					l110:
						{
							position111, tokenIndex111 := position, tokenIndex
							if buffer[position] != rune(',') {
								goto l111
							}
							position++
							if !_rules[ruleSpacing]() {
								goto l111
							}
							if !_rules[ruleWhereClause]() {
								goto l111
							}
							if !_rules[ruleSpacing]() {
								goto l111
							}
							goto l110
						l111:
							position, tokenIndex = position111, tokenIndex111
						}
						if buffer[position] != rune(')') {
							goto l107
						}
						position++
						add(ruleOrClause, position108)
					}
					goto l101
				l107:
					position, tokenIndex = position101, tokenIndex101
					{
						position112 := position
						{
							add(ruleAction18, position)
						}
						{
							position114 := position
							{
								position115 := position
								if !_rules[ruleKey]() {
									goto l98
								}
								add(rulePegText, position115)
							}
							{
								add(ruleAction19, position)
							}
							add(rulePredicate, position114)
						}
						if !_rules[ruleSpacing]() {
							goto l98
						}
						if buffer[position] != rune('(') {
							goto l98
						}
						position++
						if !_rules[ruleSpacing]() {
							goto l98
						}
						if !_rules[rulePredicateValue]() {
							goto l98
						}
					l117:
						{
							position118, tokenIndex118 := position, tokenIndex
							if buffer[position] != rune(',') {
								goto l118
							}
							position++
							if !_rules[ruleSpacing]() {
								goto l118
							}
							if !_rules[rulePredicateValue]() {
								goto l118
							}
							if !_rules[ruleSpacing]() {
								goto l118
							}
							goto l117
						l118:
							position, tokenIndex = position118, tokenIndex118
						}
						if buffer[position] != rune(')') {
							goto l98
						}
						position++
						add(rulePredicateClause, position112)
					}
				}
			l101:
				{
					add(ruleAction15, position)
				}
				add(ruleWhereClause, position99)
			}
			return true
		l98:
			position, tokenIndex = position98, tokenIndex98
			return false
		},
		/* 22 AndClause <- <('a' 'n' 'd' Action16 Spacing '(' Spacing WhereClause Spacing (',' Spacing WhereClause Spacing)* ')')> */
		nil,
		/* 23 OrClause <- <('o' 'r' Action17 Spacing '(' Spacing WhereClause Spacing (',' Spacing WhereClause Spacing)* ')')> */
		nil,
		/* 24 PredicateClause <- <(Action18 Predicate Spacing '(' Spacing PredicateValue (',' Spacing PredicateValue Spacing)* ')')> */
		nil,
		/* 25 Predicate <- <(<Key> Action19)> */
		nil,
		/* 26 PredicateValue <- <(PredicateRowKey / PredicateKey / PredicateLiteral)> */
		func() bool {
			position124, tokenIndex124 := position, tokenIndex
			{
				position125 := position
				{
					position126, tokenIndex126 := position, tokenIndex
					{
						position128 := position
						if buffer[position] != rune('@') {
							goto l127
						}
						position++
						if buffer[position] != rune('k') {
							goto l127
						}
						position++
						if buffer[position] != rune('e') {
							goto l127
						}
						position++
						if buffer[position] != rune('y') {
							goto l127
						}
						position++
						{
							add(ruleAction20, position)
						}
						add(rulePredicateRowKey, position128)
					}
					goto l126
				l127:
					position, tokenIndex = position126, tokenIndex126
					{
						position131 := position
						{
							position132, tokenIndex132 := position, tokenIndex
							{
								position134 := position
								{
									position135, tokenIndex135 := position, tokenIndex
									{
										position137 := position
										if !_rules[ruleKey]() {
											goto l136
										}
										add(rulePegText, position137)
									}
									goto l135
								l136:
									position, tokenIndex = position135, tokenIndex135
									if buffer[position] != rune('@') {
										goto l133
									}
									position++
									if buffer[position] != rune('"') {
										goto l133
									}
									position++
									{
										position138 := position
										if !_rules[ruleLiteral]() {
											goto l133
										}
										add(rulePegText, position138)
									}
									if buffer[position] != rune('"') {
										goto l133
									}
									position++
								}
							l135:
								{
									add(ruleAction21, position)
								}
								add(rulePredicateKeyText, position134)
							}
							goto l132
						l133:
							position, tokenIndex = position132, tokenIndex132
							{
								position140 := position
								{
									position141 := position
									if !_rules[ruleKeyPlaceholder]() {
										goto l130
									}
									add(rulePegText, position141)
								}
								{
									add(ruleAction22, position)
								}
								add(rulePredicateKeyLiteral, position140)
							}
						}
					l132:
						add(rulePredicateKey, position131)
					}
					goto l126
				l130:
					position, tokenIndex = position126, tokenIndex126
					{
						position143 := position
						{
							position144, tokenIndex144 := position, tokenIndex
							{
								position146 := position
								if buffer[position] != rune('"') {
									goto l145
								}
								position++
								{
									position147 := position
									if !_rules[ruleLiteral]() {
										goto l145
									}
									add(rulePegText, position147)
								}
								if buffer[position] != rune('"') {
									goto l145
								}
								position++
								{
									add(ruleAction23, position)
								}
								add(rulePredicateLiteralText, position146)
							}
							goto l144
						l145:
							position, tokenIndex = position144, tokenIndex144
							{
								position149 := position
								{
									position150 := position
									if !_rules[ruleLiteralPlaceholder]() {
										goto l124
									}
									add(rulePegText, position150)
								}
								{
									add(ruleAction24, position)
								}
								add(rulePredicateLiteralPlaceholder, position149)
							}
						}
					l144:
						add(rulePredicateLiteral, position143)
					}
				}
			l126:
				add(rulePredicateValue, position125)
			}
			return true
		l124:
			position, tokenIndex = position124, tokenIndex124
			return false
		},
		/* 27 PredicateRowKey <- <('@' 'k' 'e' 'y' Action20)> */
		nil,
		/* 28 PredicateKey <- <(PredicateKeyText / PredicateKeyLiteral)> */
		nil,
		/* 29 PredicateKeyText <- <((<Key> / ('@' '"' <Literal> '"')) Action21)> */
		nil,
		/* 30 PredicateKeyLiteral <- <(<KeyPlaceholder> Action22)> */
		nil,
		/* 31 PredicateLiteral <- <(PredicateLiteralText / PredicateLiteralPlaceholder)> */
		nil,
		/* 32 PredicateLiteralText <- <('"' <Literal> '"' Action23)> */
		nil,
		/* 33 PredicateLiteralPlaceholder <- <(<LiteralPlaceholder> Action24)> */
		nil,
		/* 34 KeyPlaceholder <- <('?' '?')> */
		func() bool {
			position159, tokenIndex159 := position, tokenIndex
			{
				position160 := position
				if buffer[position] != rune('?') {
					goto l159
				}
				position++
				if buffer[position] != rune('?') {
					goto l159
				}
				position++
				add(ruleKeyPlaceholder, position160)
			}
			return true
		l159:
			position, tokenIndex = position159, tokenIndex159
			return false
		},
		/* 35 LiteralPlaceholder <- <'?'> */
		func() bool {
			position161, tokenIndex161 := position, tokenIndex
			{
				position162 := position
				if buffer[position] != rune('?') {
					goto l161
				}
				position++
				add(ruleLiteralPlaceholder, position162)
			}
			return true
		l161:
			position, tokenIndex = position161, tokenIndex161
			return false
		},
		/* 36 Literal <- <(Escape / (!'"' .))*> */
		func() bool {
			{
				position164 := position
			l165:
				{
					position166, tokenIndex166 := position, tokenIndex
					{
						position167, tokenIndex167 := position, tokenIndex
						{
							position169 := position
							if buffer[position] != rune('\\') {
								goto l168
							}
							position++
							{
								switch buffer[position] {
								case 'v':
									if buffer[position] != rune('v') {
										goto l168
									}
									position++
									break
								case 't':
									if buffer[position] != rune('t') {
										goto l168
									}
									position++
									break
								case 'r':
									if buffer[position] != rune('r') {
										goto l168
									}
									position++
									break
								case 'n':
									if buffer[position] != rune('n') {
										goto l168
									}
									position++
									break
								case 'f':
									if buffer[position] != rune('f') {
										goto l168
									}
									position++
									break
								case 'b':
									if buffer[position] != rune('b') {
										goto l168
									}
									position++
									break
								case 'a':
									if buffer[position] != rune('a') {
										goto l168
									}
									position++
									break
								case '\\':
									if buffer[position] != rune('\\') {
										goto l168
									}
									position++
									break
								default:
									if buffer[position] != rune('"') {
										goto l168
									}
									position++
									break
								}
							}

							add(ruleEscape, position169)
						}
						goto l167
					l168:
						position, tokenIndex = position167, tokenIndex167
						{
							position171, tokenIndex171 := position, tokenIndex
							if buffer[position] != rune('"') {
								goto l171
							}
							position++
							goto l166
						l171:
							position, tokenIndex = position171, tokenIndex171
						}
						if !matchDot() {
							goto l166
						}
					}
				l167:
					goto l165
				l166:
					position, tokenIndex = position166, tokenIndex166
				}
				add(ruleLiteral, position164)
			}
			return true
		},
		/* 37 PositiveInteger <- <([1-9] [0-9]*)> */
		nil,
		/* 38 Key <- <((&('-') '-') | (&('+') '+') | (&('.') '.') | (&('_') '_') | (&('0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9') [0-9]) | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z]))+> */
		func() bool {
			position173, tokenIndex173 := position, tokenIndex
			{
				position174 := position
				{
					switch buffer[position] {
					case '-':
						if buffer[position] != rune('-') {
							goto l173
						}
						position++
						break
					case '+':
						if buffer[position] != rune('+') {
							goto l173
						}
						position++
						break
					case '.':
						if buffer[position] != rune('.') {
							goto l173
						}
						position++
						break
					case '_':
						if buffer[position] != rune('_') {
							goto l173
						}
						position++
						break
					case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l173
						}
						position++
						break
					case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l173
						}
						position++
						break
					default:
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l173
						}
						position++
						break
					}
				}

			l175:
				{
					position176, tokenIndex176 := position, tokenIndex
					{
						switch buffer[position] {
						case '-':
							if buffer[position] != rune('-') {
								goto l176
							}
							position++
							break
						case '+':
							if buffer[position] != rune('+') {
								goto l176
							}
							position++
							break
						case '.':
							if buffer[position] != rune('.') {
								goto l176
							}
							position++
							break
						case '_':
							if buffer[position] != rune('_') {
								goto l176
							}
							position++
							break
						case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l176
							}
							position++
							break
						case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l176
							}
							position++
							break
						default:
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l176
							}
							position++
							break
						}
					}

					goto l175
				l176:
					position, tokenIndex = position176, tokenIndex176
				}
				add(ruleKey, position174)
			}
			return true
		l173:
			position, tokenIndex = position173, tokenIndex173
			return false
		},
		/* 39 Escape <- <('\\' ((&('v') 'v') | (&('t') 't') | (&('r') 'r') | (&('n') 'n') | (&('f') 'f') | (&('b') 'b') | (&('a') 'a') | (&('\\') '\\') | (&('"') '"')))> */
		nil,
		/* 40 MustSpacing <- <((&('\n') '\n') | (&('\t') '\t') | (&(' ') ' '))+> */
		func() bool {
			position180, tokenIndex180 := position, tokenIndex
			{
				position181 := position
				{
					switch buffer[position] {
					case '\n':
						if buffer[position] != rune('\n') {
							goto l180
						}
						position++
						break
					case '\t':
						if buffer[position] != rune('\t') {
							goto l180
						}
						position++
						break
					default:
						if buffer[position] != rune(' ') {
							goto l180
						}
						position++
						break
					}
				}

			l182:
				{
					position183, tokenIndex183 := position, tokenIndex
					{
						switch buffer[position] {
						case '\n':
							if buffer[position] != rune('\n') {
								goto l183
							}
							position++
							break
						case '\t':
							if buffer[position] != rune('\t') {
								goto l183
							}
							position++
							break
						default:
							if buffer[position] != rune(' ') {
								goto l183
							}
							position++
							break
						}
					}

					goto l182
				l183:
					position, tokenIndex = position183, tokenIndex183
				}
				add(ruleMustSpacing, position181)
			}
			return true
		l180:
			position, tokenIndex = position180, tokenIndex180
			return false
		},
		/* 41 Spacing <- <((&('\n') '\n') | (&('\t') '\t') | (&(' ') ' '))*> */
		func() bool {
			{
				position187 := position
			l188:
				{
					position189, tokenIndex189 := position, tokenIndex
					{
						switch buffer[position] {
						case '\n':
							if buffer[position] != rune('\n') {
								goto l189
							}
							position++
							break
						case '\t':
							if buffer[position] != rune('\t') {
								goto l189
							}
							position++
							break
						default:
							if buffer[position] != rune(' ') {
								goto l189
							}
							position++
							break
						}
					}

					goto l188
				l189:
					position, tokenIndex = position189, tokenIndex189
				}
				add(ruleSpacing, position187)
			}
			return true
		},
		/* 43 Action0 <- <{ p.AddSelect() }> */
		nil,
		/* 44 Action1 <- <{ p.AddJoin() }> */
		nil,
		nil,
		/* 46 Action2 <- <{ p.SetTableName(buffer[begin:end]) }> */
		nil,
		/* 47 Action3 <- <{ p.SetTableNamePlaceholder(begin) }> */
		nil,
		/* 48 Action4 <- <{ p.AddJoinRow() }> */
		nil,
		/* 49 Action5 <- <{ p.SetJoinRowKeyPlaceholder(begin) }> */
		nil,
		/* 50 Action6 <- <{ p.SetJoinRowKey(buffer[begin:end]) }> */
		nil,
		/* 51 Action7 <- <{ p.SetJoinValuePlaceholder(begin) }> */
		nil,
		/* 52 Action8 <- <{ p.SetJoinValue(buffer[begin:end]) }> */
		nil,
		/* 53 Action9 <- <{ p.SetJoinKey(buffer[begin:end]) }> */
		nil,
		/* 54 Action10 <- <{ p.SetJoinKeyPlaceholder(begin) }> */
		nil,
		/* 55 Action11 <- <{ p.SetLimit(buffer[begin:end])}> */
		nil,
		/* 56 Action12 <- <{ p.SetLimitPlaceholder(begin) }> */
		nil,
		/* 57 Action13 <- <{ p.AddCryptoKey(buffer[begin:end]) }> */
		nil,
		/* 58 Action14 <- <{ p.PushWhere() }> */
		nil,
		/* 59 Action15 <- <{ p.PopWhere() }> */
		nil,
		/* 60 Action16 <- <{ p.SetWhereCommand("and") }> */
		nil,
		/* 61 Action17 <- <{ p.SetWhereCommand("or") }> */
		nil,
		/* 62 Action18 <- <{ p.InitPredicate() }> */
		nil,
		/* 63 Action19 <- <{ p.SetPredicateCommand(buffer[begin:end]) }> */
		nil,
		/* 64 Action20 <- <{ p.UsePredicateRowKey() }> */
		nil,
		/* 65 Action21 <- <{ p.AddPredicateKey(buffer[begin:end]) }> */
		nil,
		/* 66 Action22 <- <{ p.AddPredicateKeyPlaceholder(begin) }> */
		nil,
		/* 67 Action23 <- <{ p.AddPredicateLiteral(buffer[begin:end])}> */
		nil,
		/* 68 Action24 <- <{ p.AddPredicateLiteralPlaceholder(begin) }> */
		nil,
	}
	p.rules = _rules
}
