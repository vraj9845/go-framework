package tools

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
	"github.com/yugabyte/gocql"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Tools struct {
	MyRouter        *mux.Router
	sqlObj          *sql.DB
	cqlObj          cqlObj
	mongoObj        *mongo.Client
	redisObj        *redis.Client
	lastServedIndex int
	numberOfHosts   int
	sliceOfHosts    []string
	//kafka implemented soon :)
}

type cqlObj struct {
	cluster *gocql.ClusterConfig
	session *gocql.Session
}

type loadBalancerConfig struct {
	ListOfHosts string `mapstructure:"LOAD_BALANCER_HOSTS"`
}

type Config struct {
	// SQL - SQL, Postgres, YSQL, MariaDB, MicrosoftSQL ...
	SQL_DRIVER_NAME string
	SQL_HOST        string `mapstructure:"SQL_HOST"`
	SQL_PORT        string `mapstructure:"SQL_PORT"`
	SQL_USER        string `mapstructure:"SQL_USER"`
	SQL_PASSWORD    string `mapstructure:"SQL_PASSWORD"`
	SQL_DB_NAME     string `mapstructure:"SQL_DB_NAME"`

	// CASSANDRA
	CASSANDRA_HOSTS string `mapstructure:"CASSANDRA_HOSTS"`

	// MONGO
	MONGO_DRIVER_NAME string `mapstructure:"MONGO_DRIVER_NAME"`
	MONGO_HOST        string `mapstructure:"MONGO_HOST"`
	MONGO_PORT        string `mapstructure:"MONGO_PORT"`

	// REDIS
	REDIS_HOST       string `mapstructure:"REDIS_HOST"`
	REDIS_PORT       string `mapstructure:"REDIS_PORT"`
	REDIS_DB_INT_VAL string `mapstructure:"REDIS_DB_INT_VAL"`
	REDIS_PASSWORD   string `mapstructure:"REDIS_PASSWORD"`
}

func NewTools() Tools {
	return Tools{
		MyRouter: nil,
		sqlObj:   nil,
		cqlObj: cqlObj{
			cluster: nil,
			session: nil,
		},
		mongoObj:        nil,
		redisObj:        nil,
		lastServedIndex: 0,
		numberOfHosts:   0,
		sliceOfHosts:    nil,
	}
}

func defaultPage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello World!")
	fmt.Println("Endpoint Hit: default Page")
}

func (t *Tools) NewRestRequest(reqType string, path string, f func(http.ResponseWriter, *http.Request)) {

	if path == "" {
		switch reqType {
		case "POST":
			path = "/examplePOST"
		case "DELETE":
			path = "/exampleDELETE"
		case "PUT":
			path = "/examplePUT"
		case "GET":
			path = "/exampleGET"
		default:
			fmt.Println("Invalid reqType :(")
		}
	}

	if f == nil {
		t.MyRouter.HandleFunc(path, defaultPage)
	} else {
		t.MyRouter.HandleFunc(path, f)
	}
}

func NewMyRouter() *mux.Router {
	return mux.NewRouter().StrictSlash(true)
}

func (t *Tools) Start(port string) {
	fmt.Println("Rest API v2.0 - Mux Routers")
	if t.MyRouter == nil {
		t.MyRouter = NewMyRouter()
	}

	if port == "" {
		port = "10000"
	}

	log.Fatal(http.ListenAndServe(":"+port, t.MyRouter))
}

func (t *Tools) ConnectToDB() {

	// get the configs
	config, err := LoadConfigDB()
	if err != nil {
		fmt.Errorf(err.Error())
	}

	// get sql string to conect to DB
	var sqlConfig string
	sqlConfig = fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		config.SQL_HOST, config.SQL_PORT, config.SQL_USER, config.SQL_PASSWORD, config.SQL_DB_NAME)
	fmt.Println(config.SQL_DRIVER_NAME)
	fmt.Println(sqlConfig)

	sqlObj, err := sql.Open(config.SQL_DRIVER_NAME, sqlConfig)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		t.sqlObj = sqlObj
		fmt.Println("Connected to SQL!")
	}

	cluster := gocql.NewCluster(config.CASSANDRA_HOSTS)
	cluster.Timeout = 12 * time.Second
	// Create the session.
	session, err := cluster.CreateSession()
	if err != nil {
		fmt.Println(err.Error())
	} else {
		t.cqlObj.cluster = cluster
		t.cqlObj.session = session
		fmt.Println("Connected to Cassandra!")
	}

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	// MongoDB
	clientOptions := options.Client().ApplyURI(config.MONGO_DRIVER_NAME + "://" + config.MONGO_HOST + ":" + config.MONGO_PORT)
	client, err := mongo.Connect(ctx, clientOptions)

	if err != nil {
		fmt.Println(err.Error())
	} else {
		t.mongoObj = client
		fmt.Println("Connected to MongoDB!")
	}

	// redis
	intVar, err := strconv.Atoi(config.REDIS_DB_INT_VAL)
	if err != nil {
		fmt.Println(err.Error())
	}
	rc := redis.NewClient(&redis.Options{
		Addr:     config.REDIS_HOST + ":" + config.REDIS_PORT,
		Password: config.REDIS_PASSWORD,
		DB:       intVar,
	})
	
	cmd := rc.Ping()
	_, err = cmd.Result()
	if err != nil {
		fmt.Println(err.Error())
	} else {
		t.redisObj = rc
		fmt.Println("Connected to Redis!")
	}

}

func LoadConfigDB() (config Config, err error) {
	viper.SetConfigType("env")
	viper.SetConfigFile("./config/app.env")

	// read in config file and check your errors
	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("Error in reading config file!")
		fmt.Println(err)
	}

	// confirm where the file has been read in from
	fmt.Println("Config file used:" + viper.ConfigFileUsed())

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	return
}

func LoadConfigLoadBalancer() (lbConfig loadBalancerConfig, err error) {
	viper.SetConfigType("env")
	viper.SetConfigFile("./config/loadBalancerApp.env")

	// read in config file and check your errors
	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("Error in reading config file!")
		fmt.Println(err)
	}

	// confirm where the file has been read in from
	fmt.Println("Config file used:" + viper.ConfigFileUsed())

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&lbConfig)
	return
}

func (t *Tools) StartLoandBalancer() {
	lbConfig, err := LoadConfigLoadBalancer()
	if err != nil {
		fmt.Println("Unable to connect to load Balancer")
	}

	t.sliceOfHosts = strings.Split(lbConfig.ListOfHosts, ",")
	t.numberOfHosts = len(t.sliceOfHosts)

	http.HandleFunc("/", t.forwardRequest)
	log.Fatal(http.ListenAndServe(":10080", nil))
}

func (t *Tools) forwardRequest(res http.ResponseWriter, req *http.Request) {
	url := t.getServer()
	rProxy := httputil.NewSingleHostReverseProxy(url)
	fmt.Printf("Routing the request to the URL: %s", url.String())
	rProxy.ServeHTTP(res, req)
}
func (t *Tools) getServer() *url.URL {
	nextIndex := (t.lastServedIndex + 1) % t.numberOfHosts
	url, _ := url.Parse(t.sliceOfHosts[t.lastServedIndex])
	t.lastServedIndex = nextIndex
	return url
}
