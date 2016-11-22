package main

import (
	"fmt"
	"sort"
)

var (
//map guildid to Poll (one poll per channel)

)

type Poll struct {
	description string
	votes       map[string]int
	voters      []string
}

func (p *Poll) vote(vote string) {
	i, ok := p.votes[vote]
	if ok {
		p.votes[vote] = i + 1
	} else {
		p.votes[vote] = 1
	}
}

func (p *Poll) getResult() string {
	out := ""
	pairList := rankByVoteCount(p.votes)
	for _, p := range pairList {
		out = fmt.Sprintf("%s%d  : %s\n", out, p.Value, p.Key)
	}
	return out
}

func rankByVoteCount(voteFrequencies map[string]int) PairList {
	pl := make(PairList, len(voteFrequencies))
	i := 0
	for k, v := range voteFrequencies {
		pl[i] = Pair{k, v}
		i++
	}
	sort.Sort(sort.Reverse(pl))
	return pl
}

type Pair struct {
	Key   string
	Value int
}

type PairList []Pair

func (p PairList) Len() int           { return len(p) }
func (p PairList) Less(i, j int) bool { return p[i].Value < p[j].Value }
func (p PairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
