package usecase_test

import (
	"context"
	_ "embed"
	"testing"

	"github.com/m-mizutani/gt"
	"github.com/m-mizutani/nounify/pkg/domain/mock"
	"github.com/m-mizutani/nounify/pkg/domain/model"
	"github.com/m-mizutani/nounify/pkg/domain/types"
	"github.com/m-mizutani/nounify/pkg/usecase"
	"github.com/m-mizutani/nounify/pkg/utils/testutil"
	"github.com/m-mizutani/opac"
	"github.com/slack-go/slack"
)

func TestSendSlackMessage(t *testing.T) {
	token := testutil.LoadEnv(t, "NOUNIFY_SLACK_OAUTH_TOKEN")

	slackClient := slack.New(token)
	mockPolicy := &mock.PolicyMock{
		QueryFunc: func(ctx context.Context, query string, input, output any, options ...opac.QueryOption) error {
			msg := model.MessageQueryOutput{
				Messages: []model.Message{
					{
						Title:   "Test message",
						Channel: "test2",
						Body:    "This is a test message",
						Fields: []model.MessageField{
							{
								Name:  "wild fire",
								Value: "this is a fire",
								Link:  "https://example.com/hoge.jpg",
							},
						},
						// Icon: "https://example.com/hoge.jpg",
						Emoji: ":fire:",
					},
				},
			}

			testutil.Transcode(t, &output, msg)
			return nil
		},
	}

	uc := usecase.New(
		usecase.WithSlack(slackClient),
		usecase.WithPolicy(mockPolicy),
	)

	ctx := context.Background()
	input := &model.MessageQueryInput{
		Method: "POST",
		Body:   pubsubCloudStorageData,
	}
	gt.NoError(t, uc.HandleMessage(ctx, "cloud_storage", input))
	gt.A(t, mockPolicy.QueryCalls()).Length(1)
}

//go:embed testdata/pubsub/cloud_storage.json
var pubsubCloudStorageData []byte

//go:embed testdata/policy/cloud_storage.rego
var policyCloudStorage string

func TestHandleMessage(t *testing.T) {
	type testCase struct {
		schema   types.Schema
		rawData  []byte
		policy   string
		testMock func(t *testing.T, slackMock *mock.SlackMock)
	}

	runTest := func(tc testCase) func(t *testing.T) {
		return func(t *testing.T) {
			data := testutil.DecodeJSON(t, tc.rawData)
			slackMock := &mock.SlackMock{}

			policyData := map[string]string{"msg": tc.policy}
			policy := gt.R1(opac.New(opac.Data(policyData))).NoError(t)
			uc := usecase.New(
				usecase.WithSlack(slackMock),
				usecase.WithPolicy(policy),
			)
			input := model.MessageQueryInput{
				Method: "POST",
				Body:   data,
			}
			uc.HandleMessage(context.Background(), tc.schema, &input)
			tc.testMock(t, slackMock)
		}
	}

	runTest(testCase{
		schema:  "cloud_storage",
		rawData: pubsubCloudStorageData,
		policy:  policyCloudStorage,
		testMock: func(t *testing.T, mock *mock.SlackMock) {
		},
	})
}
