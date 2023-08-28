package types

type Conflicts []*Conflict

type Conflict struct {
	InitiatedBy []string
	Seqno       string
	Winner      string
	VotePerNode map[string]ConflictVote
}

type ConflictVote struct {
	MD5   string
	Error string
}

func (cs Conflicts) Merge(c Conflict) Conflicts {
	for i := range cs {
		if c.Seqno == cs[i].Seqno {
			for node, vote := range c.VotePerNode {
				cs[i].VotePerNode[node] = vote
			}
			return cs
		}
	}

	return append(cs, &c)
}

func (cs Conflicts) ConflictWithSeqno(seqno string) *Conflict {
	for _, c := range cs {
		if seqno == c.Seqno {
			return c
		}
	}
	return nil
}

func (cs Conflicts) OldestUnresolved() *Conflict {
	for _, c := range cs {
		if c.Winner == "" {
			return c
		}
	}
	return nil
}
func (cs Conflicts) ConflictFromMD5(md5 string) *Conflict {
	for _, c := range cs {
		for _, vote := range c.VotePerNode {
			if vote.MD5 == md5 {
				return c
			}
		}
	}
	return nil
}