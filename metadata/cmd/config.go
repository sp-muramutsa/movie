package main

type config struct {
	API              apiConfig              `yaml:"api"`
	Postgres         PostgresConfig         `yaml:"postgres"`
	ServiceDiscovery serviceDiscoveryConfig `yaml:"serviceDiscovery"`
}

type PostgresConfig struct {
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Host     string `yaml:"host"`
	DBName   string `yaml:"dbname"`
}

type apiConfig struct {
	Port int `yaml:"port"`
}

type serviceDiscoveryConfig struct {
	Consul consulConfig `yaml:"consul"`
}

type consulConfig struct {
	Address string `yaml:"address"`
}