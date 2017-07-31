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
	rulePredicateText
	rulePredicatePlaceholder
	rulePredicateValue
	rulePredicateRowKey
	rulePredicateRowKeyText
	rulePredicateRowKeyLiteral
	rulePredicateKey
	rulePredicateKeyText
	rulePredicateKeyLiteral
	rulePredicateLiteral
	rulePredicateLiteralText
	rulePredicateLiteralPlaceholder
	ruleRowKeyPlaceholder
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
	ruleAction25
	ruleAction26
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
	"PredicateText",
	"PredicatePlaceholder",
	"PredicateValue",
	"PredicateRowKey",
	"PredicateRowKeyText",
	"PredicateRowKeyLiteral",
	"PredicateKey",
	"PredicateKeyText",
	"PredicateKeyLiteral",
	"PredicateLiteral",
	"PredicateLiteralText",
	"PredicateLiteralPlaceholder",
	"RowKeyPlaceholder",
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
	"Action25",
	"Action26",
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
	rules  [76]func() bool
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
			p.SetPredicatePlaceholder(begin)
		case ruleAction21:
			p.UsePredicateRowKey()
		case ruleAction22:
			p.UsePredicateRowKeyPlaceholder(begin)
		case ruleAction23:
			p.AddPredicateKey(buffer[begin:end])
		case ruleAction24:
			p.AddPredicateKeyPlaceholder(begin)
		case ruleAction25:
			p.AddPredicateLiteral(buffer[begin:end])
		case ruleAction26:
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
								position115, tokenIndex115 := position, tokenIndex
								{
									position117 := position
									{
										position118 := position
										if !_rules[ruleKey]() {
											goto l116
										}
										add(rulePegText, position118)
									}
									{
										add(ruleAction19, position)
									}
									add(rulePredicateText, position117)
								}
								goto l115
							l116:
								position, tokenIndex = position115, tokenIndex115
								{
									position120 := position
									{
										position121 := position
										if !_rules[ruleKeyPlaceholder]() {
											goto l98
										}
										add(rulePegText, position121)
									}
									{
										add(ruleAction20, position)
									}
									add(rulePredicatePlaceholder, position120)
								}
							}
						l115:
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
					l123:
						{
							position124, tokenIndex124 := position, tokenIndex
							if buffer[position] != rune(',') {
								goto l124
							}
							position++
							if !_rules[ruleSpacing]() {
								goto l124
							}
							if !_rules[rulePredicateValue]() {
								goto l124
							}
							if !_rules[ruleSpacing]() {
								goto l124
							}
							goto l123
						l124:
							position, tokenIndex = position124, tokenIndex124
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
		/* 25 Predicate <- <(PredicateText / PredicatePlaceholder)> */
		nil,
		/* 26 PredicateText <- <(<Key> Action19)> */
		nil,
		/* 27 PredicatePlaceholder <- <(<KeyPlaceholder> Action20)> */
		nil,
		/* 28 PredicateValue <- <(PredicateRowKey / PredicateKey / PredicateLiteral)> */
		func() bool {
			position132, tokenIndex132 := position, tokenIndex
			{
				position133 := position
				{
					position134, tokenIndex134 := position, tokenIndex
					{
						position136 := position
						{
							position137, tokenIndex137 := position, tokenIndex
							{
								position139 := position
								if buffer[position] != rune('@') {
									goto l138
								}
								position++
								if buffer[position] != rune('k') {
									goto l138
								}
								position++
								if buffer[position] != rune('e') {
									goto l138
								}
								position++
								if buffer[position] != rune('y') {
									goto l138
								}
								position++
								{
									add(ruleAction21, position)
								}
								add(rulePredicateRowKeyText, position139)
							}
							goto l137
						l138:
							position, tokenIndex = position137, tokenIndex137
							{
								position141 := position
								{
									position142 := position
									{
										position143 := position
										if buffer[position] != rune('@') {
											goto l135
										}
										position++
										if buffer[position] != rune('?') {
											goto l135
										}
										position++
										if buffer[position] != rune('?') {
											goto l135
										}
										position++
										add(ruleRowKeyPlaceholder, position143)
									}
									add(rulePegText, position142)
								}
								{
									add(ruleAction22, position)
								}
								add(rulePredicateRowKeyLiteral, position141)
							}
						}
					l137:
						add(rulePredicateRowKey, position136)
					}
					goto l134
				l135:
					position, tokenIndex = position134, tokenIndex134
					{
						position146 := position
						{
							position147, tokenIndex147 := position, tokenIndex
							{
								position149 := position
								{
									position150, tokenIndex150 := position, tokenIndex
									{
										position152 := position
										if !_rules[ruleKey]() {
											goto l151
										}
										add(rulePegText, position152)
									}
									goto l150
								l151:
									position, tokenIndex = position150, tokenIndex150
									if buffer[position] != rune('@') {
										goto l148
									}
									position++
									if buffer[position] != rune('"') {
										goto l148
									}
									position++
									{
										position153 := position
										if !_rules[ruleLiteral]() {
											goto l148
										}
										add(rulePegText, position153)
									}
									if buffer[position] != rune('"') {
										goto l148
									}
									position++
								}
							l150:
								{
									add(ruleAction23, position)
								}
								add(rulePredicateKeyText, position149)
							}
							goto l147
						l148:
							position, tokenIndex = position147, tokenIndex147
							{
								position155 := position
								{
									position156 := position
									if !_rules[ruleKeyPlaceholder]() {
										goto l145
									}
									add(rulePegText, position156)
								}
								{
									add(ruleAction24, position)
								}
								add(rulePredicateKeyLiteral, position155)
							}
						}
					l147:
						add(rulePredicateKey, position146)
					}
					goto l134
				l145:
					position, tokenIndex = position134, tokenIndex134
					{
						position158 := position
						{
							position159, tokenIndex159 := position, tokenIndex
							{
								position161 := position
								if buffer[position] != rune('"') {
									goto l160
								}
								position++
								{
									position162 := position
									if !_rules[ruleLiteral]() {
										goto l160
									}
									add(rulePegText, position162)
								}
								if buffer[position] != rune('"') {
									goto l160
								}
								position++
								{
									add(ruleAction25, position)
								}
								add(rulePredicateLiteralText, position161)
							}
							goto l159
						l160:
							position, tokenIndex = position159, tokenIndex159
							{
								position164 := position
								{
									position165 := position
									if !_rules[ruleLiteralPlaceholder]() {
										goto l132
									}
									add(rulePegText, position165)
								}
								{
									add(ruleAction26, position)
								}
								add(rulePredicateLiteralPlaceholder, position164)
							}
						}
					l159:
						add(rulePredicateLiteral, position158)
					}
				}
			l134:
				add(rulePredicateValue, position133)
			}
			return true
		l132:
			position, tokenIndex = position132, tokenIndex132
			return false
		},
		/* 29 PredicateRowKey <- <(PredicateRowKeyText / PredicateRowKeyLiteral)> */
		nil,
		/* 30 PredicateRowKeyText <- <('@' 'k' 'e' 'y' Action21)> */
		nil,
		/* 31 PredicateRowKeyLiteral <- <(<RowKeyPlaceholder> Action22)> */
		nil,
		/* 32 PredicateKey <- <(PredicateKeyText / PredicateKeyLiteral)> */
		nil,
		/* 33 PredicateKeyText <- <((<Key> / ('@' '"' <Literal> '"')) Action23)> */
		nil,
		/* 34 PredicateKeyLiteral <- <(<KeyPlaceholder> Action24)> */
		nil,
		/* 35 PredicateLiteral <- <(PredicateLiteralText / PredicateLiteralPlaceholder)> */
		nil,
		/* 36 PredicateLiteralText <- <('"' <Literal> '"' Action25)> */
		nil,
		/* 37 PredicateLiteralPlaceholder <- <(<LiteralPlaceholder> Action26)> */
		nil,
		/* 38 RowKeyPlaceholder <- <('@' '?' '?')> */
		nil,
		/* 39 KeyPlaceholder <- <('?' '?')> */
		func() bool {
			position177, tokenIndex177 := position, tokenIndex
			{
				position178 := position
				if buffer[position] != rune('?') {
					goto l177
				}
				position++
				if buffer[position] != rune('?') {
					goto l177
				}
				position++
				add(ruleKeyPlaceholder, position178)
			}
			return true
		l177:
			position, tokenIndex = position177, tokenIndex177
			return false
		},
		/* 40 LiteralPlaceholder <- <'?'> */
		func() bool {
			position179, tokenIndex179 := position, tokenIndex
			{
				position180 := position
				if buffer[position] != rune('?') {
					goto l179
				}
				position++
				add(ruleLiteralPlaceholder, position180)
			}
			return true
		l179:
			position, tokenIndex = position179, tokenIndex179
			return false
		},
		/* 41 Literal <- <(Escape / (!'"' .))*> */
		func() bool {
			{
				position182 := position
			l183:
				{
					position184, tokenIndex184 := position, tokenIndex
					{
						position185, tokenIndex185 := position, tokenIndex
						{
							position187 := position
							if buffer[position] != rune('\\') {
								goto l186
							}
							position++
							{
								switch buffer[position] {
								case 'v':
									if buffer[position] != rune('v') {
										goto l186
									}
									position++
									break
								case 't':
									if buffer[position] != rune('t') {
										goto l186
									}
									position++
									break
								case 'r':
									if buffer[position] != rune('r') {
										goto l186
									}
									position++
									break
								case 'n':
									if buffer[position] != rune('n') {
										goto l186
									}
									position++
									break
								case 'f':
									if buffer[position] != rune('f') {
										goto l186
									}
									position++
									break
								case 'b':
									if buffer[position] != rune('b') {
										goto l186
									}
									position++
									break
								case 'a':
									if buffer[position] != rune('a') {
										goto l186
									}
									position++
									break
								case '\\':
									if buffer[position] != rune('\\') {
										goto l186
									}
									position++
									break
								default:
									if buffer[position] != rune('"') {
										goto l186
									}
									position++
									break
								}
							}

							add(ruleEscape, position187)
						}
						goto l185
					l186:
						position, tokenIndex = position185, tokenIndex185
						{
							position189, tokenIndex189 := position, tokenIndex
							if buffer[position] != rune('"') {
								goto l189
							}
							position++
							goto l184
						l189:
							position, tokenIndex = position189, tokenIndex189
						}
						if !matchDot() {
							goto l184
						}
					}
				l185:
					goto l183
				l184:
					position, tokenIndex = position184, tokenIndex184
				}
				add(ruleLiteral, position182)
			}
			return true
		},
		/* 42 PositiveInteger <- <([1-9] [0-9]*)> */
		nil,
		/* 43 Key <- <((&('-') '-') | (&('+') '+') | (&('.') '.') | (&('_') '_') | (&('0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9') [0-9]) | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z]))+> */
		func() bool {
			position191, tokenIndex191 := position, tokenIndex
			{
				position192 := position
				{
					switch buffer[position] {
					case '-':
						if buffer[position] != rune('-') {
							goto l191
						}
						position++
						break
					case '+':
						if buffer[position] != rune('+') {
							goto l191
						}
						position++
						break
					case '.':
						if buffer[position] != rune('.') {
							goto l191
						}
						position++
						break
					case '_':
						if buffer[position] != rune('_') {
							goto l191
						}
						position++
						break
					case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l191
						}
						position++
						break
					case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l191
						}
						position++
						break
					default:
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l191
						}
						position++
						break
					}
				}

			l193:
				{
					position194, tokenIndex194 := position, tokenIndex
					{
						switch buffer[position] {
						case '-':
							if buffer[position] != rune('-') {
								goto l194
							}
							position++
							break
						case '+':
							if buffer[position] != rune('+') {
								goto l194
							}
							position++
							break
						case '.':
							if buffer[position] != rune('.') {
								goto l194
							}
							position++
							break
						case '_':
							if buffer[position] != rune('_') {
								goto l194
							}
							position++
							break
						case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l194
							}
							position++
							break
						case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l194
							}
							position++
							break
						default:
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l194
							}
							position++
							break
						}
					}

					goto l193
				l194:
					position, tokenIndex = position194, tokenIndex194
				}
				add(ruleKey, position192)
			}
			return true
		l191:
			position, tokenIndex = position191, tokenIndex191
			return false
		},
		/* 44 Escape <- <('\\' ((&('v') 'v') | (&('t') 't') | (&('r') 'r') | (&('n') 'n') | (&('f') 'f') | (&('b') 'b') | (&('a') 'a') | (&('\\') '\\') | (&('"') '"')))> */
		nil,
		/* 45 MustSpacing <- <((&('\n') '\n') | (&('\t') '\t') | (&(' ') ' '))+> */
		func() bool {
			position198, tokenIndex198 := position, tokenIndex
			{
				position199 := position
				{
					switch buffer[position] {
					case '\n':
						if buffer[position] != rune('\n') {
							goto l198
						}
						position++
						break
					case '\t':
						if buffer[position] != rune('\t') {
							goto l198
						}
						position++
						break
					default:
						if buffer[position] != rune(' ') {
							goto l198
						}
						position++
						break
					}
				}

			l200:
				{
					position201, tokenIndex201 := position, tokenIndex
					{
						switch buffer[position] {
						case '\n':
							if buffer[position] != rune('\n') {
								goto l201
							}
							position++
							break
						case '\t':
							if buffer[position] != rune('\t') {
								goto l201
							}
							position++
							break
						default:
							if buffer[position] != rune(' ') {
								goto l201
							}
							position++
							break
						}
					}

					goto l200
				l201:
					position, tokenIndex = position201, tokenIndex201
				}
				add(ruleMustSpacing, position199)
			}
			return true
		l198:
			position, tokenIndex = position198, tokenIndex198
			return false
		},
		/* 46 Spacing <- <((&('\n') '\n') | (&('\t') '\t') | (&(' ') ' '))*> */
		func() bool {
			{
				position205 := position
			l206:
				{
					position207, tokenIndex207 := position, tokenIndex
					{
						switch buffer[position] {
						case '\n':
							if buffer[position] != rune('\n') {
								goto l207
							}
							position++
							break
						case '\t':
							if buffer[position] != rune('\t') {
								goto l207
							}
							position++
							break
						default:
							if buffer[position] != rune(' ') {
								goto l207
							}
							position++
							break
						}
					}

					goto l206
				l207:
					position, tokenIndex = position207, tokenIndex207
				}
				add(ruleSpacing, position205)
			}
			return true
		},
		/* 48 Action0 <- <{ p.AddSelect() }> */
		nil,
		/* 49 Action1 <- <{ p.AddJoin() }> */
		nil,
		nil,
		/* 51 Action2 <- <{ p.SetTableName(buffer[begin:end]) }> */
		nil,
		/* 52 Action3 <- <{ p.SetTableNamePlaceholder(begin) }> */
		nil,
		/* 53 Action4 <- <{ p.AddJoinRow() }> */
		nil,
		/* 54 Action5 <- <{ p.SetJoinRowKeyPlaceholder(begin) }> */
		nil,
		/* 55 Action6 <- <{ p.SetJoinRowKey(buffer[begin:end]) }> */
		nil,
		/* 56 Action7 <- <{ p.SetJoinValuePlaceholder(begin) }> */
		nil,
		/* 57 Action8 <- <{ p.SetJoinValue(buffer[begin:end]) }> */
		nil,
		/* 58 Action9 <- <{ p.SetJoinKey(buffer[begin:end]) }> */
		nil,
		/* 59 Action10 <- <{ p.SetJoinKeyPlaceholder(begin) }> */
		nil,
		/* 60 Action11 <- <{ p.SetLimit(buffer[begin:end])}> */
		nil,
		/* 61 Action12 <- <{ p.SetLimitPlaceholder(begin) }> */
		nil,
		/* 62 Action13 <- <{ p.AddCryptoKey(buffer[begin:end]) }> */
		nil,
		/* 63 Action14 <- <{ p.PushWhere() }> */
		nil,
		/* 64 Action15 <- <{ p.PopWhere() }> */
		nil,
		/* 65 Action16 <- <{ p.SetWhereCommand("and") }> */
		nil,
		/* 66 Action17 <- <{ p.SetWhereCommand("or") }> */
		nil,
		/* 67 Action18 <- <{ p.InitPredicate() }> */
		nil,
		/* 68 Action19 <- <{ p.SetPredicateCommand(buffer[begin:end]) }> */
		nil,
		/* 69 Action20 <- <{ p.SetPredicatePlaceholder(begin) }> */
		nil,
		/* 70 Action21 <- <{ p.UsePredicateRowKey() }> */
		nil,
		/* 71 Action22 <- <{ p.UsePredicateRowKeyPlaceholder(begin) }> */
		nil,
		/* 72 Action23 <- <{ p.AddPredicateKey(buffer[begin:end]) }> */
		nil,
		/* 73 Action24 <- <{ p.AddPredicateKeyPlaceholder(begin) }> */
		nil,
		/* 74 Action25 <- <{ p.AddPredicateLiteral(buffer[begin:end])}> */
		nil,
		/* 75 Action26 <- <{ p.AddPredicateLiteralPlaceholder(begin) }> */
		nil,
	}
	p.rules = _rules
}
