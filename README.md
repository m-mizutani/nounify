# nounify

The unified notification service for all HTTP requests.

`nounify` can receives any notification from any services via HTTP. For example, you can send a notification from GitHub Webhooks, Google Pub/Sub, and so on. When receiving a notification via HTTP request, `nounify` validates and modifies the notification message based on Rego policies. Rego can not only permit or deny the request but also creating a new message from notification data. So you can customize the notification message for each channel.

![architecture](https://github.com/m-mizutani/nounify/assets/605953/1896df93-d853-45a2-8482-73044c425615)

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

- `NOUNIFY_ADDR` (required): The address to listen to. e.g. `0.0.0.0:8080`
- `NOUNIFY_POLICY_FILE` (required): The path to the Rego policy file. e.g. `policies.rego`
- `NOUNIFY_SLACK_OAUTH_TOKEN` (required): The OAuth token of Slack App. It's recommended to set the token as a secret.
- `NOUNIFY_GITHUB_SECRET` (optional): The secret key for GitHub webhook. If you don't need to receive messages from GitHub, you can skip this.

Run `nounify` with the following command.

```shell
$ nounify serve
```

See [the example release configs](https://github.com/m-mizutani/releases/tree/main/cloud-run/nounify) with Cloud Build and Cloud Run.

## Policy

### Message

package: `msg.{schema}`

`schema` should be set to specify the policy for the message. You can name arbitrarily `schema` (e.g. `msg.github`) within the scope of Rego's package name rules. The `schema` is linked with POST path of `/msg/{schema}`. When the HTTP request is sent to the `/msg/my_policy`, the policy of package `msg.my_policy` is triggered.

#### Input

- `method` (string): The HTTP method.
- `path` (string): The HTTP path.
- `header` (map[string]string): The HTTP headers.
- `body` (any): The HTTP body. If `Content-Type` is `application/json`, the body is parsed as JSON. Otherwise, the body is a string.

#### Output

- `channel` (string, required): The channel to send the message to.
- `color` (string): The color of the message. Specify a hex color code with `#` (e.g. `#2EB67D`) or a color name (`info`, `warning` and `error` are available). Default is `info`.
- `title` (string): The title of the message.
- `body` (string): The body of the message.
- `fields` (array): The fields of the message.
  - `name` (string): The title of the field.
  - `value` (string): The value of the field.
  - `link` (string): Whether the field is short.
- `icon` (string): The icon URL of the message.
- `emoji` (string): The emoji for icon of the message. This is prioritized over `icon`.

### Auth

package: `auth`

#### Input

- `method` (string): The HTTP method.
- `path` (string): The HTTP path.
- `headers` (map[string]string): The HTTP headers.
- `auth`: The authentication information.
  - `github`: The GitHub authentication information.
    - `delivery` (string): The GitHub delivery ID.
    - `event` (string): The GitHub event. See [GitHub Webhooks](https://docs.github.com/en/webhooks/webhook-events-and-payloads) for more information.
    - `hook_id`: The GitHub hook ID.
    - `install_id`: The GitHub App installation ID.
    - `install_type`: The GitHub App installation type.
  - `google`: ID token claims from Google
    - `aud` (string): The audience. If you run `nounify` on Google Cloud Run and `nounify` receives the message via Pub/Sub, the audience is the Cloud Run URL.
    - `email` (string): The email address of the subject.
    - `exp` (int): The expiration time in unix time.
    - `iat` (int): The issued at time in unix time.
    - `iss` (string): The issuer. If the ID token is issued by Google, the issuer must be `https://accounts.google.com`.
    - `sub` (string): The subject. The subject is the unique identifier of the user.

#### Output

- `allow` (bool): Whether the request is allowed.

## License

Apache License 2.0
