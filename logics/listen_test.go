package logics

import (
	"github.com/zyjblockchain/tt_tac/utils"
	"testing"
)

func TestTacProcess_ListenErc20CollectionAddress(t *testing.T) {
	addr := "0x000000000000000000000000000000d3c349b165e64ff01c5d66"
	t.Log(utils.FormatAddressHex(addr) == "0xd3c349b165e64ff01c5d66")
}
