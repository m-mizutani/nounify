package usecase

import "github.com/m-mizutani/nounify/pkg/domain/interfaces"

type UseCases struct {
	slack  interfaces.Slack
	policy interfaces.Policy
}

func New(options ...Option) *UseCases {
	uc := &UseCases{}
	for _, option := range options {
		option(uc)
	}

	return uc
}

type Option func(*UseCases)

func WithSlack(slack interfaces.Slack) Option {
	return func(uc *UseCases) {
		uc.slack = slack
	}
}

func WithPolicy(policy interfaces.Policy) Option {
	return func(uc *UseCases) {
		uc.policy = policy
	}
}
