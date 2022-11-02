package middleware

type CreateMiddlewareDto struct {
	Name          string
	Namespace     string
	StripPrefixes []string
}
