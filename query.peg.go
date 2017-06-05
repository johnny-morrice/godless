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
	ruleJoin
	ruleJoinKey
	ruleJoinRow
	ruleKeyJoin
	ruleValueJoin
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
)

var rul3s = [...]string{
	"Unknown",
	"Query",
	"Join",
	"JoinKey",
	"JoinRow",
	"KeyJoin",
	"ValueJoin",
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
	rules  [45]func() bool
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
			p.AddJoinRow()
		case ruleAction4:
			p.SetJoinRowKey(buffer[begin:end])
		case ruleAction5:
			p.SetJoinKey(buffer[begin:end])
		case ruleAction6:
			p.SetJoinValue(buffer[begin:end])
		case ruleAction7:
			p.SetTableName(buffer[begin:end])
		case ruleAction8:
			p.SetLimit(buffer[begin:end])
		case ruleAction9:
			p.PushWhere()
		case ruleAction10:
			p.PopWhere()
		case ruleAction11:
			p.SetWhereCommand("and")
		case ruleAction12:
			p.SetWhereCommand("or")
		case ruleAction13:
			p.InitPredicate()
		case ruleAction14:
			p.SetPredicateCommand(buffer[begin:end])
		case ruleAction15:
			p.UsePredicateRowKey()
		case ruleAction16:
			p.AddPredicateKey(buffer[begin:end])
		case ruleAction17:
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
		/* 0 Query <- <(Spacing ((Select Action0) / (Join Action1)) !.)> */
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
						{
							position5 := position
							{
								position6 := position
								if !_rules[ruleKey]() {
									goto l3
								}
								add(rulePegText, position6)
							}
							{
								add(ruleAction7, position)
							}
							add(ruleSelectKey, position5)
						}
						if !_rules[ruleMustSpacing]() {
							goto l3
						}
						{
							position8, tokenIndex8 := position, tokenIndex
							{
								position10 := position
								if buffer[position] != rune('w') {
									goto l8
								}
								position++
								if buffer[position] != rune('h') {
									goto l8
								}
								position++
								if buffer[position] != rune('e') {
									goto l8
								}
								position++
								if buffer[position] != rune('r') {
									goto l8
								}
								position++
								if buffer[position] != rune('e') {
									goto l8
								}
								position++
								if !_rules[ruleMustSpacing]() {
									goto l8
								}
								if !_rules[ruleWhereClause]() {
									goto l8
								}
								add(ruleWhere, position10)
							}
							if !_rules[ruleMustSpacing]() {
								goto l8
							}
							goto l9
						l8:
							position, tokenIndex = position8, tokenIndex8
						}
					l9:
						{
							position11 := position
							if buffer[position] != rune('l') {
								goto l3
							}
							position++
							if buffer[position] != rune('i') {
								goto l3
							}
							position++
							if buffer[position] != rune('m') {
								goto l3
							}
							position++
							if buffer[position] != rune('i') {
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
							{
								position12 := position
								{
									position13 := position
									if c := buffer[position]; c < rune('1') || c > rune('9') {
										goto l3
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
								add(ruleAction8, position)
							}
							add(ruleLimit, position11)
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
						position18 := position
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
						{
							position19 := position
							{
								position20 := position
								if !_rules[ruleKey]() {
									goto l0
								}
								add(rulePegText, position20)
							}
							{
								add(ruleAction2, position)
							}
							add(ruleJoinKey, position19)
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
					l22:
						{
							position23, tokenIndex23 := position, tokenIndex
							if !_rules[ruleSpacing]() {
								goto l23
							}
							if buffer[position] != rune(',') {
								goto l23
							}
							position++
							if !_rules[ruleSpacing]() {
								goto l23
							}
							if !_rules[ruleJoinRow]() {
								goto l23
							}
							goto l22
						l23:
							position, tokenIndex = position23, tokenIndex23
						}
						if !_rules[ruleSpacing]() {
							goto l0
						}
						add(ruleJoin, position18)
					}
					{
						add(ruleAction1, position)
					}
				}
			l2:
				{
					position25, tokenIndex25 := position, tokenIndex
					if !matchDot() {
						goto l25
					}
					goto l0
				l25:
					position, tokenIndex = position25, tokenIndex25
				}
				add(ruleQuery, position1)
			}
			return true
		l0:
			position, tokenIndex = position0, tokenIndex0
			return false
		},
		/* 1 Join <- <('j' 'o' 'i' 'n' MustSpacing JoinKey MustSpacing ('r' 'o' 'w' 's') MustSpacing JoinRow (Spacing ',' Spacing JoinRow)* Spacing)> */
		nil,
		/* 2 JoinKey <- <(<Key> Action2)> */
		nil,
		/* 3 JoinRow <- <(Action3 '(' Spacing KeyJoin Spacing (',' Spacing ValueJoin Spacing)* ')')> */
		func() bool {
			position28, tokenIndex28 := position, tokenIndex
			{
				position29 := position
				{
					add(ruleAction3, position)
				}
				if buffer[position] != rune('(') {
					goto l28
				}
				position++
				if !_rules[ruleSpacing]() {
					goto l28
				}
				{
					position31 := position
					if buffer[position] != rune('@') {
						goto l28
					}
					position++
					if buffer[position] != rune('k') {
						goto l28
					}
					position++
					if buffer[position] != rune('e') {
						goto l28
					}
					position++
					if buffer[position] != rune('y') {
						goto l28
					}
					position++
					if !_rules[ruleSpacing]() {
						goto l28
					}
					if buffer[position] != rune('=') {
						goto l28
					}
					position++
					if !_rules[ruleSpacing]() {
						goto l28
					}
					{
						position32, tokenIndex32 := position, tokenIndex
						if buffer[position] != rune('@') {
							goto l33
						}
						position++
						if buffer[position] != rune('"') {
							goto l33
						}
						position++
						{
							position34 := position
							if !_rules[ruleLiteral]() {
								goto l33
							}
							add(rulePegText, position34)
						}
						if buffer[position] != rune('"') {
							goto l33
						}
						position++
						goto l32
					l33:
						position, tokenIndex = position32, tokenIndex32
						{
							position35 := position
							if !_rules[ruleKey]() {
								goto l28
							}
							add(rulePegText, position35)
						}
					}
				l32:
					{
						add(ruleAction4, position)
					}
					add(ruleKeyJoin, position31)
				}
				if !_rules[ruleSpacing]() {
					goto l28
				}
			l37:
				{
					position38, tokenIndex38 := position, tokenIndex
					if buffer[position] != rune(',') {
						goto l38
					}
					position++
					if !_rules[ruleSpacing]() {
						goto l38
					}
					{
						position39 := position
						{
							position40, tokenIndex40 := position, tokenIndex
							{
								position42 := position
								if !_rules[ruleKey]() {
									goto l41
								}
								add(rulePegText, position42)
							}
							goto l40
						l41:
							position, tokenIndex = position40, tokenIndex40
							if buffer[position] != rune('@') {
								goto l38
							}
							position++
							if buffer[position] != rune('"') {
								goto l38
							}
							position++
							{
								position43 := position
								if !_rules[ruleLiteral]() {
									goto l38
								}
								add(rulePegText, position43)
							}
							if buffer[position] != rune('"') {
								goto l38
							}
							position++
						}
					l40:
						{
							add(ruleAction5, position)
						}
						if !_rules[ruleSpacing]() {
							goto l38
						}
						if buffer[position] != rune('=') {
							goto l38
						}
						position++
						if !_rules[ruleSpacing]() {
							goto l38
						}
						if buffer[position] != rune('"') {
							goto l38
						}
						position++
						{
							position45 := position
							if !_rules[ruleLiteral]() {
								goto l38
							}
							add(rulePegText, position45)
						}
						if buffer[position] != rune('"') {
							goto l38
						}
						position++
						{
							add(ruleAction6, position)
						}
						add(ruleValueJoin, position39)
					}
					if !_rules[ruleSpacing]() {
						goto l38
					}
					goto l37
				l38:
					position, tokenIndex = position38, tokenIndex38
				}
				if buffer[position] != rune(')') {
					goto l28
				}
				position++
				add(ruleJoinRow, position29)
			}
			return true
		l28:
			position, tokenIndex = position28, tokenIndex28
			return false
		},
		/* 4 KeyJoin <- <('@' 'k' 'e' 'y' Spacing '=' Spacing (('@' '"' <Literal> '"') / <Key>) Action4)> */
		nil,
		/* 5 ValueJoin <- <((<Key> / ('@' '"' <Literal> '"')) Action5 Spacing '=' Spacing '"' <Literal> '"' Action6)> */
		nil,
		/* 6 Select <- <('s' 'e' 'l' 'e' 'c' 't' MustSpacing SelectKey MustSpacing (Where MustSpacing)? Limit)> */
		nil,
		/* 7 SelectKey <- <(<Key> Action7)> */
		nil,
		/* 8 Limit <- <('l' 'i' 'm' 'i' 't' MustSpacing <PositiveInteger> Action8)> */
		nil,
		/* 9 Where <- <('w' 'h' 'e' 'r' 'e' MustSpacing WhereClause)> */
		nil,
		/* 10 WhereClause <- <(Action9 ((&('s') PredicateClause) | (&('o') OrClause) | (&('a') AndClause)) Action10)> */
		func() bool {
			position53, tokenIndex53 := position, tokenIndex
			{
				position54 := position
				{
					add(ruleAction9, position)
				}
				{
					switch buffer[position] {
					case 's':
						{
							position57 := position
							{
								add(ruleAction13, position)
							}
							{
								position59 := position
								{
									position60 := position
									{
										position61, tokenIndex61 := position, tokenIndex
										if buffer[position] != rune('s') {
											goto l62
										}
										position++
										if buffer[position] != rune('t') {
											goto l62
										}
										position++
										if buffer[position] != rune('r') {
											goto l62
										}
										position++
										if buffer[position] != rune('_') {
											goto l62
										}
										position++
										if buffer[position] != rune('e') {
											goto l62
										}
										position++
										if buffer[position] != rune('q') {
											goto l62
										}
										position++
										goto l61
									l62:
										position, tokenIndex = position61, tokenIndex61
										if buffer[position] != rune('s') {
											goto l53
										}
										position++
										if buffer[position] != rune('t') {
											goto l53
										}
										position++
										if buffer[position] != rune('r') {
											goto l53
										}
										position++
										if buffer[position] != rune('_') {
											goto l53
										}
										position++
										if buffer[position] != rune('n') {
											goto l53
										}
										position++
										if buffer[position] != rune('e') {
											goto l53
										}
										position++
										if buffer[position] != rune('q') {
											goto l53
										}
										position++
									}
								l61:
									add(rulePegText, position60)
								}
								{
									add(ruleAction14, position)
								}
								add(rulePredicate, position59)
							}
							if !_rules[ruleSpacing]() {
								goto l53
							}
							if buffer[position] != rune('(') {
								goto l53
							}
							position++
							if !_rules[ruleSpacing]() {
								goto l53
							}
							if !_rules[rulePredicateValue]() {
								goto l53
							}
						l64:
							{
								position65, tokenIndex65 := position, tokenIndex
								if buffer[position] != rune(',') {
									goto l65
								}
								position++
								if !_rules[ruleSpacing]() {
									goto l65
								}
								if !_rules[rulePredicateValue]() {
									goto l65
								}
								if !_rules[ruleSpacing]() {
									goto l65
								}
								goto l64
							l65:
								position, tokenIndex = position65, tokenIndex65
							}
							if buffer[position] != rune(')') {
								goto l53
							}
							position++
							add(rulePredicateClause, position57)
						}
						break
					case 'o':
						{
							position66 := position
							if buffer[position] != rune('o') {
								goto l53
							}
							position++
							if buffer[position] != rune('r') {
								goto l53
							}
							position++
							{
								add(ruleAction12, position)
							}
							if !_rules[ruleSpacing]() {
								goto l53
							}
							if buffer[position] != rune('(') {
								goto l53
							}
							position++
							if !_rules[ruleSpacing]() {
								goto l53
							}
							if !_rules[ruleWhereClause]() {
								goto l53
							}
							if !_rules[ruleSpacing]() {
								goto l53
							}
						l68:
							{
								position69, tokenIndex69 := position, tokenIndex
								if buffer[position] != rune(',') {
									goto l69
								}
								position++
								if !_rules[ruleSpacing]() {
									goto l69
								}
								if !_rules[ruleWhereClause]() {
									goto l69
								}
								if !_rules[ruleSpacing]() {
									goto l69
								}
								goto l68
							l69:
								position, tokenIndex = position69, tokenIndex69
							}
							if buffer[position] != rune(')') {
								goto l53
							}
							position++
							add(ruleOrClause, position66)
						}
						break
					default:
						{
							position70 := position
							if buffer[position] != rune('a') {
								goto l53
							}
							position++
							if buffer[position] != rune('n') {
								goto l53
							}
							position++
							if buffer[position] != rune('d') {
								goto l53
							}
							position++
							{
								add(ruleAction11, position)
							}
							if !_rules[ruleSpacing]() {
								goto l53
							}
							if buffer[position] != rune('(') {
								goto l53
							}
							position++
							if !_rules[ruleSpacing]() {
								goto l53
							}
							if !_rules[ruleWhereClause]() {
								goto l53
							}
							if !_rules[ruleSpacing]() {
								goto l53
							}
						l72:
							{
								position73, tokenIndex73 := position, tokenIndex
								if buffer[position] != rune(',') {
									goto l73
								}
								position++
								if !_rules[ruleSpacing]() {
									goto l73
								}
								if !_rules[ruleWhereClause]() {
									goto l73
								}
								if !_rules[ruleSpacing]() {
									goto l73
								}
								goto l72
							l73:
								position, tokenIndex = position73, tokenIndex73
							}
							if buffer[position] != rune(')') {
								goto l53
							}
							position++
							add(ruleAndClause, position70)
						}
						break
					}
				}

				{
					add(ruleAction10, position)
				}
				add(ruleWhereClause, position54)
			}
			return true
		l53:
			position, tokenIndex = position53, tokenIndex53
			return false
		},
		/* 11 AndClause <- <('a' 'n' 'd' Action11 Spacing '(' Spacing WhereClause Spacing (',' Spacing WhereClause Spacing)* ')')> */
		nil,
		/* 12 OrClause <- <('o' 'r' Action12 Spacing '(' Spacing WhereClause Spacing (',' Spacing WhereClause Spacing)* ')')> */
		nil,
		/* 13 PredicateClause <- <(Action13 Predicate Spacing '(' Spacing PredicateValue (',' Spacing PredicateValue Spacing)* ')')> */
		nil,
		/* 14 Predicate <- <(<(('s' 't' 'r' '_' 'e' 'q') / ('s' 't' 'r' '_' 'n' 'e' 'q'))> Action14)> */
		nil,
		/* 15 PredicateValue <- <(PredicateRowKey / PredicateKey / PredicateLiteralValue)> */
		func() bool {
			position79, tokenIndex79 := position, tokenIndex
			{
				position80 := position
				{
					position81, tokenIndex81 := position, tokenIndex
					{
						position83 := position
						if buffer[position] != rune('@') {
							goto l82
						}
						position++
						if buffer[position] != rune('k') {
							goto l82
						}
						position++
						if buffer[position] != rune('e') {
							goto l82
						}
						position++
						if buffer[position] != rune('y') {
							goto l82
						}
						position++
						{
							add(ruleAction15, position)
						}
						add(rulePredicateRowKey, position83)
					}
					goto l81
				l82:
					position, tokenIndex = position81, tokenIndex81
					{
						position86 := position
						{
							position87, tokenIndex87 := position, tokenIndex
							{
								position89 := position
								if !_rules[ruleKey]() {
									goto l88
								}
								add(rulePegText, position89)
							}
							goto l87
						l88:
							position, tokenIndex = position87, tokenIndex87
							if buffer[position] != rune('@') {
								goto l85
							}
							position++
							if buffer[position] != rune('"') {
								goto l85
							}
							position++
							{
								position90 := position
								if !_rules[ruleLiteral]() {
									goto l85
								}
								add(rulePegText, position90)
							}
							if buffer[position] != rune('"') {
								goto l85
							}
							position++
						}
					l87:
						{
							add(ruleAction16, position)
						}
						add(rulePredicateKey, position86)
					}
					goto l81
				l85:
					position, tokenIndex = position81, tokenIndex81
					{
						position92 := position
						if buffer[position] != rune('"') {
							goto l79
						}
						position++
						{
							position93 := position
							if !_rules[ruleLiteral]() {
								goto l79
							}
							add(rulePegText, position93)
						}
						if buffer[position] != rune('"') {
							goto l79
						}
						position++
						{
							add(ruleAction17, position)
						}
						add(rulePredicateLiteralValue, position92)
					}
				}
			l81:
				add(rulePredicateValue, position80)
			}
			return true
		l79:
			position, tokenIndex = position79, tokenIndex79
			return false
		},
		/* 16 PredicateRowKey <- <('@' 'k' 'e' 'y' Action15)> */
		nil,
		/* 17 PredicateKey <- <((<Key> / ('@' '"' <Literal> '"')) Action16)> */
		nil,
		/* 18 PredicateLiteralValue <- <('"' <Literal> '"' Action17)> */
		nil,
		/* 19 Literal <- <(Escape / (!'"' .))*> */
		func() bool {
			{
				position99 := position
			l100:
				{
					position101, tokenIndex101 := position, tokenIndex
					{
						position102, tokenIndex102 := position, tokenIndex
						{
							position104 := position
							if buffer[position] != rune('\\') {
								goto l103
							}
							position++
							{
								switch buffer[position] {
								case 'v':
									if buffer[position] != rune('v') {
										goto l103
									}
									position++
									break
								case 't':
									if buffer[position] != rune('t') {
										goto l103
									}
									position++
									break
								case 'r':
									if buffer[position] != rune('r') {
										goto l103
									}
									position++
									break
								case 'n':
									if buffer[position] != rune('n') {
										goto l103
									}
									position++
									break
								case 'f':
									if buffer[position] != rune('f') {
										goto l103
									}
									position++
									break
								case 'b':
									if buffer[position] != rune('b') {
										goto l103
									}
									position++
									break
								case 'a':
									if buffer[position] != rune('a') {
										goto l103
									}
									position++
									break
								case '\\':
									if buffer[position] != rune('\\') {
										goto l103
									}
									position++
									break
								default:
									if buffer[position] != rune('"') {
										goto l103
									}
									position++
									break
								}
							}

							add(ruleEscape, position104)
						}
						goto l102
					l103:
						position, tokenIndex = position102, tokenIndex102
						{
							position106, tokenIndex106 := position, tokenIndex
							if buffer[position] != rune('"') {
								goto l106
							}
							position++
							goto l101
						l106:
							position, tokenIndex = position106, tokenIndex106
						}
						if !matchDot() {
							goto l101
						}
					}
				l102:
					goto l100
				l101:
					position, tokenIndex = position101, tokenIndex101
				}
				add(ruleLiteral, position99)
			}
			return true
		},
		/* 20 PositiveInteger <- <([1-9] [0-9]*)> */
		nil,
		/* 21 Key <- <((&('0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9') [0-9]) | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z]))+> */
		func() bool {
			position108, tokenIndex108 := position, tokenIndex
			{
				position109 := position
				{
					switch buffer[position] {
					case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l108
						}
						position++
						break
					case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l108
						}
						position++
						break
					default:
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l108
						}
						position++
						break
					}
				}

			l110:
				{
					position111, tokenIndex111 := position, tokenIndex
					{
						switch buffer[position] {
						case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l111
							}
							position++
							break
						case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l111
							}
							position++
							break
						default:
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l111
							}
							position++
							break
						}
					}

					goto l110
				l111:
					position, tokenIndex = position111, tokenIndex111
				}
				add(ruleKey, position109)
			}
			return true
		l108:
			position, tokenIndex = position108, tokenIndex108
			return false
		},
		/* 22 Escape <- <('\\' ((&('v') 'v') | (&('t') 't') | (&('r') 'r') | (&('n') 'n') | (&('f') 'f') | (&('b') 'b') | (&('a') 'a') | (&('\\') '\\') | (&('"') '"')))> */
		nil,
		/* 23 MustSpacing <- <((&('\n') '\n') | (&('\t') '\t') | (&(' ') ' '))+> */
		func() bool {
			position115, tokenIndex115 := position, tokenIndex
			{
				position116 := position
				{
					switch buffer[position] {
					case '\n':
						if buffer[position] != rune('\n') {
							goto l115
						}
						position++
						break
					case '\t':
						if buffer[position] != rune('\t') {
							goto l115
						}
						position++
						break
					default:
						if buffer[position] != rune(' ') {
							goto l115
						}
						position++
						break
					}
				}

			l117:
				{
					position118, tokenIndex118 := position, tokenIndex
					{
						switch buffer[position] {
						case '\n':
							if buffer[position] != rune('\n') {
								goto l118
							}
							position++
							break
						case '\t':
							if buffer[position] != rune('\t') {
								goto l118
							}
							position++
							break
						default:
							if buffer[position] != rune(' ') {
								goto l118
							}
							position++
							break
						}
					}

					goto l117
				l118:
					position, tokenIndex = position118, tokenIndex118
				}
				add(ruleMustSpacing, position116)
			}
			return true
		l115:
			position, tokenIndex = position115, tokenIndex115
			return false
		},
		/* 24 Spacing <- <((&('\n') '\n') | (&('\t') '\t') | (&(' ') ' '))*> */
		func() bool {
			{
				position122 := position
			l123:
				{
					position124, tokenIndex124 := position, tokenIndex
					{
						switch buffer[position] {
						case '\n':
							if buffer[position] != rune('\n') {
								goto l124
							}
							position++
							break
						case '\t':
							if buffer[position] != rune('\t') {
								goto l124
							}
							position++
							break
						default:
							if buffer[position] != rune(' ') {
								goto l124
							}
							position++
							break
						}
					}

					goto l123
				l124:
					position, tokenIndex = position124, tokenIndex124
				}
				add(ruleSpacing, position122)
			}
			return true
		},
		/* 26 Action0 <- <{ p.AddSelect() }> */
		nil,
		/* 27 Action1 <- <{ p.AddJoin() }> */
		nil,
		nil,
		/* 29 Action2 <- <{ p.SetTableName(buffer[begin:end]) }> */
		nil,
		/* 30 Action3 <- <{ p.AddJoinRow() }> */
		nil,
		/* 31 Action4 <- <{ p.SetJoinRowKey(buffer[begin:end]) }> */
		nil,
		/* 32 Action5 <- <{ p.SetJoinKey(buffer[begin:end]) }> */
		nil,
		/* 33 Action6 <- <{ p.SetJoinValue(buffer[begin:end]) }> */
		nil,
		/* 34 Action7 <- <{ p.SetTableName(buffer[begin:end]) }> */
		nil,
		/* 35 Action8 <- <{ p.SetLimit(buffer[begin:end])}> */
		nil,
		/* 36 Action9 <- <{ p.PushWhere() }> */
		nil,
		/* 37 Action10 <- <{ p.PopWhere() }> */
		nil,
		/* 38 Action11 <- <{ p.SetWhereCommand("and") }> */
		nil,
		/* 39 Action12 <- <{ p.SetWhereCommand("or") }> */
		nil,
		/* 40 Action13 <- <{ p.InitPredicate() }> */
		nil,
		/* 41 Action14 <- <{ p.SetPredicateCommand(buffer[begin:end]) }> */
		nil,
		/* 42 Action15 <- <{ p.UsePredicateRowKey() }> */
		nil,
		/* 43 Action16 <- <{ p.AddPredicateKey(buffer[begin:end]) }> */
		nil,
		/* 44 Action17 <- <{ p.AddPredicateLiteral(buffer[begin:end])}> */
		nil,
	}
	p.rules = _rules
}
