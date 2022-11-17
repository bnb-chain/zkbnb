package test

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/go-redis/redis/v8"
	"github.com/ory/dockertest"
	"github.com/ory/dockertest/docker"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Database is the struct for database docker
type Database struct {
	Name string
	DB   *gorm.DB

	pool     *dockertest.Pool
	resource *dockertest.Resource
	host     string
	port     string
	isGithub bool
}

// Redis is the struct for redis docker
type Redis struct {
	pool     *dockertest.Pool
	resource *dockertest.Resource
	Client   *redis.Client
}

// RunDB run docker of db for unit test
func RunDB(dbName string) (*Database, error) {

	pool, err := dockertest.NewPool("")
	if err != nil {
		return nil, err
	}
	resource := &dockertest.Resource{}

	opt := docker.ListContainersOptions{
		All: true,
		Filters: map[string][]string{
			"ancestor": {"ghcr.io/bnb-chain/zkbnb/zkbnb-ut-postgres:latest"},
			"name":     {"zkbnb_unittest_pg"},
		},
	}
	allContainers, err := pool.Client.ListContainers(opt)
	if err != nil {
		return nil, err
	}

	_, reuse := os.LookupEnv("REUSE_DOCKER")
	if !reuse && len(allContainers) == 0 {
		resource, err = pool.RunWithOptions(
			&dockertest.RunOptions{Repository: "ghcr.io/bnb-chain/zkbnb/zkbnb-ut-postgres", Tag: "latest", Env: []string{"POSTGRES_PASSWORD=ZkBNB@123", "POSTGRES_USER=postgres", "POSTGRES_DB=zkbnb"}, Name: "zkbnb_unittest_pg"},
		)
		if err != nil {
			return nil, err
		}
	} else {
		container := allContainers[0]
		if container.State != "running" {
			fmt.Printf("Try start non-running postgres docker\n")
			err := pool.Client.StartContainer(container.ID, &docker.HostConfig{})
			if err != nil {
				return nil, err
			}
		}

		c, err := pool.Client.InspectContainer(container.ID)
		if err != nil {
			return nil, err
		}
		resource = &dockertest.Resource{
			Container: c,
		}
	}

	host := "127.0.0.1"
	port := resource.GetPort("5432/tcp")

	db, err := getConnection(pool, host, port, "zkbnb")
	if err != nil {
		return nil, err
	}

	_, isGithub := os.LookupEnv("GITHUB_ENV")
	d := &Database{
		Name:     dbName,
		pool:     pool,
		resource: resource,
		host:     host,
		port:     port,
		DB:       db,
		isGithub: isGithub,
	}

	return d, nil
}

func getConnectionPath(host, port, dbName string) string {
	return fmt.Sprintf("user=postgres password=ZkBNB@123 host=%s port=%s dbname=%s sslmode=disable", host, port, dbName)
}

func getConnection(pool *dockertest.Pool, host, port, dbName string) (*gorm.DB, error) {
	var db *gorm.DB
	err := pool.Retry(func() error {
		var err error
		postgresConn := postgres.Open(getConnectionPath(host, port, dbName))
		db, err = gorm.Open(postgresConn, &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
		if err != nil {
			return err
		}
		conn, err := db.DB()
		if err != nil {
			return err
		}
		return conn.Ping()
	})
	if err != nil {
		return nil, err
	}
	// NOTE: Debug() for printing query statements
	return db.Debug(), nil
}

// StopDB stop and remove the docker of db for unit test
func (d *Database) StopDB() error {

	// For github environment, let github deal with it
	if d.isGithub {
		return nil
	}

	_, reuse := os.LookupEnv("REUSE_DOCKER")
	if reuse {
		return nil
	}

	err := d.pool.Purge(d.resource)
	if err != nil {
		return err
	}
	return nil
}

// GetDBName get a db name by using the Suite struct in each test
func GetDBName(s interface{}) string {
	return strings.ReplaceAll(reflect.TypeOf(s).Name(), "/", "_")
}

// InitDB init database schema and data
func (d *Database) InitDB() error {

	return nil
}

// ClearDB truncate the tables in database
func (d *Database) ClearDB(tables []string) error {
	for _, table := range tables {
		dbTx := d.DB.Exec(fmt.Sprintf("TRUNCATE %s;", table))
		if dbTx.Error != nil {
			return dbTx.Error
		}
	}

	return nil
}
