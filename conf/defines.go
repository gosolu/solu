package conf

type Service struct {
	Name string `json:"name" toml:"name"`
}

type LogFile struct {
}

type Log struct {
}

type Database struct {
}

type Redis struct {
}

type Kafka struct {
}

type Dependency struct {
}

type Any struct {
	KVs map[string]interface{}
}
