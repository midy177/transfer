package separation

import "webhook-transfer/separation/weixin"

type Separation interface {
	ConvertAndSend(payload []byte, id string) error
}

func New() map[string]Separation {
	var separations = make(map[string]Separation)
	separations["weixin"] = weixin.New()
	return separations
}
