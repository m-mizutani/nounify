# Rule

The rule of notification and authentication in nounify is written in Rego language. The rule is a set of conditions and actions. The conditions are evaluated with the input of the HTTP request, and the actions are executed when the conditions are satisfied.

## Message Rule

package: `msg.{schema}`

`schema` should be set to specify the rule for the message. You can name arbitrarily `schema` (e.g. `msg.github`) within the scope of Rego's package name rules. The `schema` is linked with POST path of `/msg/{schema}`. When the HTTP request is sent to the `/msg/my_policy`, the policy of package `msg.my_policy` is triggered.

Schema can be nested with `.`. For example, both of `msg.github` and `msg.github.my_repo` are valid schema names. If the schema name with multiple `.`, the path of HTTP request should be `/msg/github/my_repo`.

### Input

- `method` (string): The HTTP method.
- `path` (string): The HTTP path.
- `header` (map[string]string): The HTTP headers.
- `body` (any): The HTTP body. If `Content-Type` is `application/json`, the body is parsed as JSON. Otherwise, the body is a string.
- `auth`: [AuthContext](#authcontext)

### Output

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

## Auth Rule

package: `auth`

The policy for the authentication of the request. The policy is triggered when the HTTP request is sent to the `/msg/*` path. The policy should return `allow` as `true` if the request is allowed.

If you want to allow all requests, you can set the policy as follows.

```rego
package auth

allow := true
```

### Input

- `method` (string): The HTTP method.
- `path` (string): The HTTP path.
- `headers` (map[string]string): The HTTP headers.
- `auth`: [AuthContext](#authcontext)

### Output

- `allow` (bool): Whether the request is allowed.

## Models

### AuthContext

That presents the authentication context. The context has only validated claims and information.

- `github`:
  - `app`: [GitHub App Webhook values](#github-app-webhook-values)
  - `action`: [GitHub Action claims](#github-action-claims)
- `google`: ID token claims from Google


#### Google ID Token clams

- `aud` (string): The audience. If you run `nounify` on Google Cloud Run and `nounify` receives the message via Pub/Sub, the audience is the Cloud Run URL.
- `email` (string): The email address of the subject.
- `exp` (int): The expiration time in unix time.
- `iat` (int): The issued at time in unix time.
- `iss` (string): The issuer. If the ID token is issued by Google, the issuer must be `https://accounts.google.com`.
- `sub` (string): The subject. The subject is the unique identifier of the user.

**Example**
```json
{
  "at_hash": "g6ugwZqiij9FjLWh3GS0pA",
  "aud": "1234567890.apps.googleusercontent.com",
  "azp": "1234567890.apps.googleusercontent.com",
  "email": "mizutani@example.com",
  "email_verified": true,
  "exp": 1720231639,
  "hd": "dr-ubie.com",
  "iat": 1720228039,
  "iss": "https://accounts.google.com",
  "sub": "1234567890987654321"
}
```

#### GitHub App Webhook values

The header values of GitHub App Webhook request that is validated with secret.

- `delivery` (string): The GitHub delivery ID.
- `event` (string): The GitHub event. See [GitHub Webhooks](https://docs.github.com/en/webhooks/webhook-events-and-payloads) for more information.
- `hook_id`: The GitHub hook ID.
- `install_id`: The GitHub App installation ID.
- `install_type`: The GitHub App installation type.

#### GitHub Action claims

- `actor` (string): The GitHub user who triggered the event. e.g. `m-mizutani`
- `actor_id` (string): The ID of the GitHub user who triggered the event. e.g. `12345678`
- `aud` (string): The intended audience for the token.
- `base_ref` (string): The base reference for the event, if applicable.
- `event_name` (string): The name of the event that triggered the workflow.
- `exp` (number): The expiration time of the token in Unix time.
- `head_ref` (string): The head reference for the event, if applicable.
- `iat` (number): The time at which the token was issued in Unix time.
- `iss` (string): The issuer of the token.
- `job_workflow_ref` (string): The reference to the workflow file.
- `job_workflow_sha` (string): The SHA of the workflow file.
- `jti` (string): The unique identifier for the token.
- `nbf` (number): The "not before" time for the token in Unix time.
- `ref` (string): The git reference (branch or tag) that triggered the workflow.
- `ref_protected` (string): Indicates if the reference is protected.
- `ref_type` (string): The type of the reference (e.g., branch or tag).
- `repository` (string): The name of the repository.
- `repository_id` (string): The ID of the repository.
- `repository_owner` (string): The owner of the repository.
- `repository_owner_id` (string): The ID of the repository owner.
- `repository_visibility` (string): The visibility of the repository (e.g., private or public).
- `run_attempt` (string): The attempt number of the workflow run.
- `run_id` (string): The ID of the workflow run.
- `run_number` (string): The number of the workflow run.
- `runner_environment` (string): The environment in which the runner is executing (e.g., GitHub-hosted).
- `sha` (string): The commit SHA that triggered the workflow.
- `sub` (string): The subject of the token, often including the repository and reference.
- `workflow` (string): The name of the workflow.
- `workflow_ref` (string): The reference to the workflow file.
- `workflow_sha` (string): The SHA of the workflow file.

**Example**
```json
{
  "actor": "m-mizutani",
  "actor_id": "605953",
  "aud": "https://github.com/m-mizutani",
  "base_ref": "",
  "event_name": "workflow_dispatch",
  "exp": 1720151757,
  "head_ref": "",
  "iat": 1720151457,
  "iss": "https://token.actions.githubusercontent.com",
  "job_workflow_ref": "m-mizutani/sandbox/.github/workflows/token.yaml@refs/heads/main",
  "job_workflow_sha": "cc0d9fe499612987bb78187fc7d6bffb8766c900",
  "jti": "f2a62b45-e1f9-411c-9cb1-d00d5d42bab7",
  "nbf": 1720150857,
  "ref": "refs/heads/main",
  "ref_protected": "false",
  "ref_type": "branch",
  "repository": "m-mizutani/sandbox",
  "repository_id": "707088626",
  "repository_owner": "m-mizutani",
  "repository_owner_id": "605953",
  "repository_visibility": "private",
  "run_attempt": "1",
  "run_id": "9802816352",
  "run_number": "6",
  "runner_environment": "github-hosted",
  "sha": "cc0d9fe499612987bb78187fc7d6bffb8766c900",
  "sub": "repo:m-mizutani/sandbox:ref:refs/heads/main",
  "workflow": "test_github_oidc",
  "workflow_ref": "m-mizutani/sandbox/.github/workflows/token.yaml@refs/heads/main",
  "workflow_sha": "cc0d9fe499612987bb78187fc7d6bffb8766c900"
}
```