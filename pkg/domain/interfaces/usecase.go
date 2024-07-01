package interfaces

import (
	"context"

	"github.com/m-mizutani/nounify/pkg/domain/model"
	"github.com/m-mizutani/nounify/pkg/domain/types"
)

type UseCases interface {
	HandleMessage(ctx context.Context, schema types.Schema, input *model.MessageQueryInput) error
}
