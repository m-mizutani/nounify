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
  "title": "[New Issue] " + input.body.issue.title,
  "body": input.body.issue.body,
  "fields": [
    {
      "name": "Author",
      "value": input.body.issue.user.login,
      "link": input.body.issue.user.html_url,
    },
    {
      "name": "Issue",
      "value": input.body.issue.number,
      "link": input.body.issue.html_url,
    },
  ],
}] {
  input.header["X-GitHub-Event"] == "issues"
  input.body.action == "opened"
}
```

## Usage

## Policy

### Message

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
