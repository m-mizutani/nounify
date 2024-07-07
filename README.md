# nounify

The unified notification service for all HTTP requests.

`nounify` can receives any notification from any services via HTTP. For example, you can send a notification from GitHub Webhooks, Google Pub/Sub, and so on. When receiving a notification via HTTP request, `nounify` validates and modifies the notification message based on Rego policies. Rego can not only permit or deny the request but also creating a new message from notification data. So you can customize the notification message for each channel.

![architecture](https://github.com/m-mizutani/nounify/assets/605953/4b8b5460-85ce-42e4-a21b-90106a207134)

For example, here is a policy that converts a GitHub Webhook message to a Slack message. The rule is triggered when a new issue is opened, and the message is sent to the `#github-notify` channel with the octopus emoji.

```rego
package schema.github_webhook

msg[{
  "channel": "github-notify",
  "color": "#2EB67D",
  "emoji": ":octopus:",
  "title": "New issue opened",
  "body": input.body.issue.body,
  "fields": [
    {
      "name": "Author",
      "value": input.body.issue.user.login,
      "link": input.body.issue.user.html_url,
    },
    {
      "name": "Issue",
      "value": sprintf("#%d: %s", [input.body.issue.number, input.body.issue.title]),
      "link": input.body.issue.html_url,
    },
  ],
}] {
  input.header["X-Github-Event"] == "issues"
  input.body.action == "opened"
}
```

When creating a new issue such as [this](https://github.com/m-mizutani/nounify/issues/2), the following message will be emitted.

<img width="503" alt="Screenshot 2024-07-03 at 12 50 24" src="https://github.com/m-mizutani/octovy-deploy/assets/605953/bc6eee31-3348-4169-948b-e51c74656e4a">

## Usage

### Prerequisites

- Create a Slack App and get OAuth token.
  - The app should have `chat:write`, `chat:write.customize` and `chat:write.public` scope.
  - Install the app to your workspace.
- If you need to receive messages from GitHub App, create a GitHub App.
  - Enable permissions for your interest and subscribe them. See [Using webhooks with GitHub Apps](https://docs.github.com/en/apps/creating-github-apps/registering-a-github-app/choosing-permissions-for-a-github-app) for more information.
  - Install the app to your repository.
  - Set random secret key for webhook, and keep it secret.

### Deploy

Set following environment variables to deploy `nounify`.

- Basic settings
  - `NOUNIFY_ADDR` (required): The address to listen to. e.g. `0.0.0.0:8080`
  - `NOUNIFY_POLICY_FILE` (required): The path to the Rego policy file. e.g. `policies.rego`
  - `NOUNIFY_SLACK_OAUTH_TOKEN` (required): The OAuth token of Slack App. It's recommended to set the token as a secret.
- Authentication settings
  - `NOUNIFY_GITHUB_SECRET` (optional): The secret key for GitHub webhook. If you don't need to receive messages from GitHub, you can skip this.
  - `NOUNIFY_GITHUB_ACTION_TOKEN` (optional): If set, nounify validates the token in `Authorization` header as `Bearer` from GitHub Actions OIDC.
  - `NOUNIFY_GOOGLE_ID_TOKEN` (optional): If set, nounify validates the token in `Authorization` header as `Bearer` from Google ID Token.

Run `nounify` with the following command.

```shell
$ nounify serve
```

See [the example release configs](https://github.com/m-mizutani/releases/tree/main/cloud-run/nounify) with Cloud Build and Cloud Run.

## Policy

See [the policy document](docs/policy.md) for more information.

## License

Apache License 2.0
