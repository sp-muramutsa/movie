package main

type config struct {
	API              apiConfig              `yaml:"api"`
	ServiceDiscovery serviceDiscoveryConfig `yaml:"serviceDiscovery"`
	Jaeger           jaegerConfig           `yaml:"jaeger"`
}

type apiConfig struct {
	Port int `yaml:"port"`
}

type serviceDiscoveryConfig struct {
	Consul consulConfig `yaml:"consul"`
	Kafka  kafkaConfig  `yaml:"kafka"`
}

type consulConfig struct {
	Address string `yaml:"address"`
}

type kafkaConfig struct {
	Address string `yaml:"address"`
}

type jaegerConfig struct {
	URL string `yaml:"url"`
}
