package coretestutil

import (
	"bytes"
	"math"
	"math/rand"
	"sort"
	"time"

	"github.com/tendermint/tendermint/pkg/consts"
	"github.com/tendermint/tendermint/types"
)

// Here are test helper functions to generate data used in celestia-core to
// mimic encoded blockdata

// GenerateRandomBlockData returns randomly generated block data for testing purposes
func GenerateRandomBlockData(txCount, evdCount, msgCount, maxSize int) types.Data {
	var out types.Data
	out.Txs = GenerateRandomlySizedContiguousShares(txCount, maxSize)
	out.Evidence = generateIdenticalEvidence(evdCount)
	out.Messages = GenerateRandomlySizedMessages(msgCount, maxSize)
	return out
}

func GenerateRandomlySizedContiguousShares(count, max int) types.Txs {
	txs := make(types.Txs, count)
	for i := 0; i < count; i++ {
		size := rand.Intn(max)
		if size == 0 {
			size = 1
		}
		txs[i] = GenerateRandomContiguousShares(1, size)[0]
	}
	return txs
}

func GenerateRandomContiguousShares(count, size int) types.Txs {
	txs := make(types.Txs, count)
	for i := 0; i < count; i++ {
		tx := make([]byte, size)
		_, err := rand.Read(tx)
		if err != nil {
			panic(err)
		}
		txs[i] = tx
	}
	return txs
}

func generateIdenticalEvidence(count int) types.EvidenceData {
	evidence := make([]types.Evidence, count)
	for i := 0; i < count; i++ {
		ev := types.NewMockDuplicateVoteEvidence(math.MaxInt64, time.Now(), "chainID")
		evidence[i] = ev
	}
	return types.EvidenceData{Evidence: evidence}
}

func GenerateRandomlySizedMessages(count, maxMsgSize int) types.Messages {
	msgs := make([]types.Message, count)
	for i := 0; i < count; i++ {
		msgs[i] = GenerateRandomMessage(rand.Intn(maxMsgSize))
	}

	// this is just to let us use assert.Equal
	if count == 0 {
		msgs = nil
	}

	messages := types.Messages{MessagesList: msgs}
	messages.SortMessages()
	return messages
}

func GenerateRandomMessage(size int) types.Message {
	share := GenerateRandomNamespacedShares(1, size)[0]
	msg := types.Message{
		NamespaceID: share.NamespaceID(),
		Data:        share.Data(),
	}
	return msg
}

func GenerateRandomNamespacedShares(count, msgSize int) types.NamespacedShares {
	shares := GenerateRandNamespacedRawData(uint32(count), consts.NamespaceSize, uint32(msgSize))
	msgs := make([]types.Message, count)
	for i, s := range shares {
		msgs[i] = types.Message{
			Data:        s[consts.NamespaceSize:],
			NamespaceID: s[:consts.NamespaceSize],
		}
	}
	return types.Messages{MessagesList: msgs}.SplitIntoShares()
}

func GenerateRandNamespacedRawData(total, nidSize, leafSize uint32) [][]byte {
	data := make([][]byte, total)
	for i := uint32(0); i < total; i++ {
		nid := make([]byte, nidSize)
		rand.Read(nid)
		data[i] = nid
	}
	sortByteArrays(data)
	for i := uint32(0); i < total; i++ {
		d := make([]byte, leafSize)
		rand.Read(d)
		data[i] = append(data[i], d...)
	}

	return data
}

func sortByteArrays(src [][]byte) {
	sort.Slice(src, func(i, j int) bool { return bytes.Compare(src[i], src[j]) < 0 })
}