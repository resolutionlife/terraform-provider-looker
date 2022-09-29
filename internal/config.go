package internal

// Config contains env vars
type Config struct {
	RecMode bool `envconfig:"TF_REC" default:"false"`
}
