package testcontainer

import (
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func GetConfig() (string, []testcontainers.ContainerCustomizer) {
	return "docker.io/postgres:16-alpine", []testcontainers.ContainerCustomizer{
		postgres.WithDatabase("test"),
		postgres.WithUsername("user"),
		postgres.WithPassword("password"),
		postgres.BasicWaitStrategies(),
	}
}