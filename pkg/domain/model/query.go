package model

type MessageQueryInput struct {
	Method string            `json:"method"`
	Path   string            `json:"path"`
	Header map[string]string `json:"header"`
	Body   any               `json:"body"`
	Auth   AuthContext       `json:"auth"`
}

type MessageQueryOutput struct {
	Messages []Message `json:"msg"`
}

type Message struct {
	Channel string         `json:"channel"`
	Color   string         `json:"color"`
	Title   string         `json:"title"`
	Body    string         `json:"body"`
	Fields  []MessageField `json:"fields"`
	Icon    string         `json:"icon"`
	Emoji   string         `json:"emoji"`
}

type MessageField struct {
	Name  string `json:"name"`
	Value string `json:"value"`
	Link  string `json:"link"`
}

type AuthContext struct {
	GitHub GitHubAuth     `json:"github"`
	Google map[string]any `json:"google"`
	AWS    AwsAuth        `json:"aws"`
}

type AuthQueryInput struct {
	Method string            `json:"method"`
	Path   string            `json:"path"`
	Header map[string]string `json:"header"`
	Auth   AuthContext       `json:"auth"`
}

type AuthQueryOutput struct {
	Allow bool `json:"allow"`
}
