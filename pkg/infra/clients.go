package infra

import "github.com/m-mizutani/nounify/pkg/domain/interfaces"

type Clients struct {
	slack  interfaces.Slack
	policy interfaces.Policy
}

func (x *Clients) Slack() interfaces.Slack   { return x.slack }
func (x *Clients) Policy() interfaces.Policy { return x.policy }

func New(options ...Option) *Clients {
	clients := &Clients{}
	for _, option := range options {
		option(clients)
	}

	return clients
}

type Option func(*Clients)

func WithSlack(slack interfaces.Slack) Option {
	return func(clients *Clients) {
		clients.slack = slack
	}
}

func WithPolicy(policy interfaces.Policy) Option {
	return func(clients *Clients) {
		clients.policy = policy
	}
}
