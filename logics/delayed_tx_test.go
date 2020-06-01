package logics

import (
	"testing"
	"time"
)

func TestSendAllUsdt(t *testing.T) {
	for {
		from := "0x7AC954Ed6c2d96d48BBad405aa1579C828409f59"
		private := "3E990B440C3FC1CCFF8B3339D41850C5B9A3D712F804FA3EE1CDD8F322B4A556"
		to := "0x59375A522876aB96B0ed2953D0D3b92674701Cc2"
		txHash, err := SendAllUsdt(private, from, to)
		if err != nil {
			t.Log(err)
			return
		}
		t.Log(txHash)
		time.Sleep(10 * time.Second)
	}
}
