package rpg

import "math/rand"

func init() {
	addOpponentPool("intro.training.8.pool", []opponentPoolMember{
		{
			weight:    1,
			primortal: "pumbl",
		},
		{
			weight:    1,
			primortal: "toxmidge",
		},
		{
			weight:    1,
			primortal: "myceli",
		},
	})
}

type opponentPoolMember struct {
	weight    int
	primortal PrimortalType
	archetype string
}

type opponentPool struct {
	members     []opponentPoolMember
	totalWeight int
}

func (op opponentPool) Random() (PrimortalType, string) {
	if len(op.members) == 0 {
		panic("no members")
	}
	n := rand.Intn(op.totalWeight)
	for _, member := range op.members {
		n -= member.weight
		if n <= 0 {
			return member.primortal, member.archetype
		}
	}
	return op.members[0].primortal, op.members[0].archetype
}

var opponentPools = map[string]opponentPool{}

type OpponentPool interface {
	Random() (PrimortalType, string)
}

func GetOpponentPool(name string) OpponentPool {
	pool, ok := opponentPools[name]
	if !ok {
		panic("unknown opponent pool: " + name)
	}
	return pool
}

func addOpponentPool(name string, members []opponentPoolMember) {
	if _, ok := opponentPools[name]; ok {
		panic("duplicate opponent pool: " + name)
	}
	totalWeight := 0
	for _, member := range members {
		totalWeight += member.weight
	}
	opponentPools[name] = opponentPool{members: members, totalWeight: totalWeight}
}
