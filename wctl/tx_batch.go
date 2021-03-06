package wctl

import (
	"github.com/perlin-network/wavelet"
	"github.com/perlin-network/wavelet/sys"
)

func (c *Client) SendBatch(batch wavelet.Batch) (*TxResponse, error) {
	marshaled, err := batch.Marshal()
	if err != nil {
		return nil, err
	}

	return c.SendTransaction(byte(sys.TagBatch), marshaled)
}
