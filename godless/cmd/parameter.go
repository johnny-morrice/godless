package cmd

import (
	"time"
)

type Parameters struct {
	strs      map[string]*string
	strSlices map[string]*[]string
	ints      map[string]*int
	bools     map[string]*bool
	durs      map[string]*time.Duration
}

func (p *Parameters) String(flagName string) *string {
	if p.strs == nil {
		p.strs = map[string]*string{}
	}

	str, ok := p.strs[flagName]

	if !ok {
		str = new(string)
		p.strs[flagName] = str
	}

	return str
}

func (p *Parameters) StringSlice(flagName string) *[]string {
	if p.strSlices == nil {
		p.strSlices = map[string]*[]string{}
	}

	str, ok := p.strSlices[flagName]

	if !ok {
		str = new([]string)
		p.strSlices[flagName] = str
	}

	return str
}

func (p *Parameters) Bool(flagName string) *bool {
	if p.bools == nil {
		p.bools = map[string]*bool{}
	}

	bl, ok := p.bools[flagName]

	if !ok {
		bl = new(bool)
		p.bools[flagName] = bl
	}

	return bl
}

func (p *Parameters) Int(flagName string) *int {
	if p.ints == nil {
		p.ints = map[string]*int{}
	}

	i, ok := p.ints[flagName]

	if !ok {
		i = new(int)
		p.ints[flagName] = i
	}

	return i
}

func (p *Parameters) Duration(flagName string) *time.Duration {
	if p.durs == nil {
		p.durs = map[string]*time.Duration{}
	}

	dur, ok := p.durs[flagName]

	if !ok {
		dur = new(time.Duration)
		p.durs[flagName] = dur
	}

	return dur
}
