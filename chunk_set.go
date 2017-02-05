package godless

type ChunkSet struct {
	Chunks []Chunk
}

func (set *ChunkSet) Join(other *ChunkSet) *ChunkSet {
	ret := &ChunkSet{}
	ret.joinSelf(set)
	ret.joinSelf(other)
	return ret
}

func (set *ChunkSet) joinSelf(other *ChunkSet) {
	set.Chunks = uniqChunks(append(set.Chunks, other.Chunks...))
}

func (set *ChunkSet) Validate(checker ChunkValidator) (*ChunkSet, []InvalidChunk) {
	ret := &ChunkSet{}
	invalid := ret.validateSelf(checker)
	return ret, invalid
}


type InvalidChunk struct {
	Ch Chunk
	Error error
}

func (set *ChunkSet) validateSelf(checker ChunkValidator) []InvalidChunk {
	ret := []InvalidChunk{}

	for _, ch := range set.Chunks {
		err := checker.Validate(ch)
		if err != nil {
			ret = append(ret, InvalidChunk{Ch: ch, Error: err})
		}
	}

	return ret
}

type ChunkValidator interface {
	Validate(Chunk) error
}

type Chunk struct {
	Data []byte
	SigSize int
	Sig [MAX_SIG_LEN]byte
}

func uniqChunks(chunks []Chunk) []Chunk {
	ret := []Chunk{}
	uniq := map[[MAX_SIG_LEN]byte]Chunk{}

	for _, ch := range chunks {
		if _, ok := uniq[ch.Sig]; ok {
			continue
		}

		uniq[ch.Sig] = ch
	}

	for _, ch := range uniq {
		ret = append(ret, ch)
	}

	return ret
}

const MAX_SIG_LEN = 512
