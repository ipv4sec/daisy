package config

type Config struct {
	Source struct{
		Host []string
		Auth string
		Database int
		Prefix string
	}
	Target struct{
		Host []string
		Auth string
		Database int
		Prefix string
	}
}
