package helper

import (
	"server/transform/api"
)

type ResponserWrapper struct {
	Resource api.Transformer
}

func (receiver *ResponserWrapper) SetResource(entity api.Transformer) *ResponserWrapper {
	receiver.Resource = entity
	return receiver
}

func (receiver *ResponserWrapper) GetResource() api.Transformer {

	return receiver.Resource
}

func NewResponseWrapper() *ResponserWrapper {
	return &ResponserWrapper{}
}
