package godless

type CvrdtChunkSet struct {
	Chunks []Chunk
}

func (set *CvrdtChunkSet) Join(other *CvrdtChunkSet) *CvrdtChunkSet {
	ret := &CvrdtChunkSet{}
	ret.joinSelf(set)
	ret.joinSelf(other)
	return ret
}

func (set *CvrdtChunkSet) joinSelf(other *CvrdtChunkSet) {
	set.Chunks = uniqChunks(append(set.Chunks, other.Chunks...))
}

func (set *CvrdtChunkSet) Validate(checker ChunkValidator) (*CvrdtChunkSet, []InvalidChunk) {
	ret := &CvrdtChunkSet{}
	invalid := ret.validateSelf(checker)
	return ret, invalid
}


type InvalidChunk struct {
	Ch Chunk
	Error error
}

func (set *CvrdtChunkSet) validateSelf(checker ChunkValidator) []InvalidChunk {
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
